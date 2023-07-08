/*
MIT License

Copyright (c) 2023 Frank Oh

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package echo_http_cache

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	refreshKey = "refreshKey"
)

type (
	// CacheStore is the interface to be implemented by custom stores.
	CacheStore interface {
		// Get retrieves the cached response by a given key. It also
		// returns true or false, whether it exists or not.
		Get(key uint64) ([]byte, bool)

		// Set caches a response for a given key until an expiration date.
		Set(key uint64, response []byte, expiration time.Time)

		// Release frees cache for a given key.
		Release(key uint64)
	}
)

type (
	// CacheConfig data structure for HTTP cache middleware.
	CacheConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		Store CacheStore

		expiration   time.Duration
		includePaths []string
		excludePaths []string
	}

	// CacheResponse is the cached response data structure.
	CacheResponse struct {
		// Value is the cached response value.
		Value []byte `json:"value"`

		// Header is the cached response header.
		Header http.Header `json:"header"`

		// Expiration is the cached response expiration date.
		Expiration time.Time `json:"expiration"`

		// LastAccess is the last date a cached response was accessed.
		// Used by LRU and MRU algorithms.
		LastAccess time.Time `json:"lastAccess"`

		// Frequency is the count of times a cached response is accessed.
		// Used for LFU and MFU algorithms.
		Frequency int `json:"frequency"`
	}
)

var (
	// DefaultCacheConfig defines default values for CacheConfig
	DefaultCacheConfig = CacheConfig{
		Skipper:    middleware.DefaultSkipper,
		expiration: 3 * time.Minute,
	}
)

/*
Cache returns a cache middleware

	e := echo.New()

... 추가 설명
*/
func Cache(store CacheStore) echo.MiddlewareFunc {
	config := DefaultCacheConfig
	config.Store = store

	return CacheWithConfig(config)
}

/*
CacheWithConfig returns a cache middleware
*/
func CacheWithConfig(config CacheConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultCacheConfig.Skipper
	}
	if config.Store == nil {
		panic("Store configuration must be provided")
	}
	if config.expiration < 1 {
		panic("Cache expiration must be provided")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			if config.isExcludePaths(c.Request().URL.String()) {
				return next(c)
			}
			if !config.isIncludePaths(c.Request().URL.String()) {
				return next(c)
			}

			if c.Request().Method == http.MethodGet {
				sortURLParams(c.Request().URL)
				key := generateKey(c.Request().Method, c.Request().URL.String())

				params := c.Request().URL.Query()

				if _, ok := params[refreshKey]; ok {
					refreshKeyValue := params.Get(refreshKey)
					delete(params, refreshKeyValue)
					config.Store.Release(key)
				} else {
					if cachedResponse, ok := config.Store.Get(key); ok {
						response := bytesToResponse(cachedResponse)
						now := time.Now()
						if response.Expiration.After(now) {
							response.LastAccess = now
							response.Frequency++

							config.Store.Set(key, response.bytes(), response.Expiration)

							for k, v := range response.Header {
								c.Response().Header().Set(k, strings.Join(v, ","))
							}
							c.Response().WriteHeader(http.StatusOK)
							c.Response().Write(response.Value)
							return nil
						}

						config.Store.Release(key)
					}
				}

				resBody := new(bytes.Buffer)
				mw := io.MultiWriter(c.Response().Writer, resBody)
				writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: c.Response().Writer}
				c.Response().Writer = writer

				if err := next(c); err != nil {
					c.Error(err)
				}

				statusCode := writer.statusCode
				value := resBody.Bytes()
				if statusCode < http.StatusBadRequest {
					now := time.Now()

					response := CacheResponse{
						Value:      value,
						Header:     writer.Header(),
						Expiration: now.Add(config.expiration),
						LastAccess: now,
						Frequency:  1,
					}
					config.Store.Set(key, response.bytes(), response.Expiration)
				}
				return nil
			}
			if err := next(c); err != nil {
				c.Error(err)
			}
			return nil
		}
	}
}

func (c *CacheConfig) isIncludePaths(URL string) bool {
	for _, p := range c.includePaths {
		if strings.Contains(URL, p) {
			return true
		}
	}
	return false
}

func (c *CacheConfig) isExcludePaths(URL string) bool {
	for _, p := range c.excludePaths {
		if strings.Contains(URL, p) {
			return true
		}
	}
	return false
}

type bodyDumpResponseWriter struct {
	io.Writer
	http.ResponseWriter
	statusCode int
}

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyDumpResponseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *bodyDumpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

// bytes converts CacheResponse data structure into bytes array.
func (r CacheResponse) bytes() []byte {
	data, _ := json.Marshal(r)
	return data
}

// bytesToResponse converts bytes array into CacheResponse data structure.
func bytesToResponse(b []byte) CacheResponse {
	var r CacheResponse
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.Decode(&r)

	return r
}

func sortURLParams(URL *url.URL) {
	params := URL.Query()
	for _, param := range params {
		sort.Slice(param, func(i, j int) bool {
			return param[i] < param[j]
		})
	}
	URL.RawQuery = params.Encode()
}

// keyAsString can be used by store to convert the cache key from uint64 to string.
func keyAsString(key uint64) string {
	return strconv.FormatUint(key, 36)
}

func generateKey(method, URL string) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(fmt.Sprintf("%s:%s", method, URL)))

	return hash.Sum64()
}
