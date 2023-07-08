package echo_http_cache

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	inMemoryStore := NewCacheMemoryStoreWithConfig(CacheMemoryStoreConfig{
		Capacity:  5,
		Algorithm: LFU,
	})

	mw := Cache(inMemoryStore)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = mw(handler)(c)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func Test_CacheWithConfig(t *testing.T) {
	e := echo.New()

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	type args struct {
		method      string
		url         string
		cacheConfig CacheConfig
	}

	type wants struct {
		code     int
		isCached bool
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "test includePaths",
			args: args{
				method: http.MethodGet,
				url:    "http://foo.bar/test-1",
				cacheConfig: CacheConfig{
					Store: NewCacheMemoryStoreWithConfig(CacheMemoryStoreConfig{
						Capacity:  5,
						Algorithm: LFU,
					}),
					expiration:   5 * time.Second,
					includePaths: []string{"foo.bar"},
				},
			},
			wants: wants{
				code:     http.StatusOK,
				isCached: true,
			},
		},
		{
			name: "test excludePaths",
			args: args{
				method: http.MethodGet,
				url:    "http://foo.bar/test-1",
				cacheConfig: CacheConfig{
					Store: NewCacheMemoryStoreWithConfig(CacheMemoryStoreConfig{
						Capacity:  5,
						Algorithm: LFU,
					}),
					expiration:   5 * time.Second,
					excludePaths: []string{"foo.bar"},
				},
			},
			wants: wants{
				code:     http.StatusOK,
				isCached: false,
			},
		},
		{
			name: "test post method",
			args: args{
				method: http.MethodPost,
				url:    "http://foo.bar/test-3",
				cacheConfig: CacheConfig{
					Store: NewCacheMemoryStoreWithConfig(CacheMemoryStoreConfig{
						Capacity:  5,
						Algorithm: LFU,
					}),
					expiration:   5 * time.Second,
					includePaths: []string{"foo.bar"},
				},
			},
			wants: wants{
				code:     http.StatusOK,
				isCached: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.args.method, tt.args.url, nil)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mw := CacheWithConfig(tt.args.cacheConfig)
			_ = mw(handler)(c)

			assert.Equal(t, tt.wants.code, rec.Code)
			assert.Equal(t, "test", rec.Body.String())

			cacheResp, ok := tt.args.cacheConfig.Store.Get(generateKey(tt.args.method, tt.args.url))
			assert.Equal(t, tt.wants.isCached, ok)

			if tt.wants.isCached {
				var cacheResponse CacheResponse
				err := json.Unmarshal(cacheResp, &cacheResponse)
				assert.NoError(t, err)
				assert.Equal(t, "test", string(cacheResponse.Value))
			}
		})
	}
}

func TestCache_panicBehavior(t *testing.T) {
	inMemoryStore := NewCacheMemoryStoreWithConfig(CacheMemoryStoreConfig{
		Capacity:  5,
		Algorithm: LFU,
	})

	assert.Panics(t, func() {
		Cache(nil)
	})

	assert.NotPanics(t, func() {
		Cache(inMemoryStore)
	})
}

func Test_bytesToResponse(t *testing.T) {
	r := CacheResponse{
		Value:      []byte("value 1"),
		Expiration: time.Time{},
		Frequency:  1,
		LastAccess: time.Time{},
	}

	tests := []struct {
		name      string
		b         []byte
		wantValue string
	}{

		{
			"convert bytes array to response",
			r.bytes(),
			"value 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bytesToResponse(tt.b)
			assert.Equal(t, tt.wantValue, string(got.Value))
		})
	}
}

func Test_nytes(t *testing.T) {
	r := CacheResponse{
		Value:      []byte("test"),
		Expiration: time.Time{},
		Frequency:  1,
		LastAccess: time.Time{},
	}

	bytes := r.bytes()
	assert.Equal(t, `{"value":"dGVzdA==","header":null,"expiration":"0001-01-01T00:00:00Z","lastAccess":"0001-01-01T00:00:00Z","frequency":1}`, string(bytes))
}

func Test_keyAsString(t *testing.T) {
	tests := []struct {
		name string
		URL  string
		want string
	}{
		{
			name: "keyAsString 1",
			URL:  "http://localhost:8080/category",
			want: "1auf9gt7r09l5",
		},
		{
			name: "keyAsString 2",
			URL:  "http://localhost:8080/category/morisco",
			want: "503atd5m9ojy",
		},
		{
			name: "keyAsString 3",
			URL:  "http://localhost:8080/category/mourisquinho",
			want: "110cga4fxnxb5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyAsNum := generateKey(http.MethodGet, tt.URL)
			keyAsStr := keyAsString(keyAsNum)
			assert.Equal(t, tt.want, keyAsStr)
		})
	}
}

func Test_generateKey(t *testing.T) {
	tests := []struct {
		name string
		URL  string
		want uint64
	}{
		{
			"get url checksum",
			"http://foo.bar/test-1",
			0x3e18cc11d24701d7,
		},
		{
			"get url 2 checksum",
			"http://foo.bar/test-2",
			0x3e18cd11d247038a,
		},
		{
			"get url 3 checksum",
			"http://foo.bar/test-3",
			0x3e18ce11d247053d,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := generateKey(http.MethodGet, tt.URL)
			assert.Equal(t, tt.want, key)
		})
	}
}

func Test_sortURLParams(t *testing.T) {
	u, _ := url.Parse("http://test.com?zaz=bar&foo=zaz&boo=foo&boo=baz")
	tests := []struct {
		name string
		URL  *url.URL
		want string
	}{
		{
			"returns url with ordered querystring params",
			u,
			"http://test.com?boo=baz&boo=foo&foo=zaz&zaz=bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortURLParams(tt.URL)
			got := tt.URL.String()
			if got != tt.want {
				t.Errorf("sortURLParams() = %v, want %v", got, tt.want)
			}
		})
	}
}
