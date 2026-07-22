package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
	"github.com/stretchr/testify/assert"
)

// newMockClient 构造一个指向 httptest server 的 Client
func newMockClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewClient(
		WithBaseURL(srv.URL),
		WithRepoBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
		WithMaxRetries(1),
		WithRetryBackoff(1),
		WithCache(false, 0),
	)
}

func TestDoRequestSuccess(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0.0","g":"g","a":"a","v":"1.0.0"}]}`))
	})
	body, err := c.doRequest(context.Background(), "GET", c.baseURL, nil, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)
}

func TestDoRequest404(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`not found`))
	})
	_, err := c.doRequest(context.Background(), "GET", c.baseURL, nil, nil)
	assert.Error(t, err)
}

func TestDoRequest500(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	_, err := c.doRequest(context.Background(), "GET", c.baseURL, nil, nil)
	assert.Error(t, err)
}

func TestExecuteWithRetrySuccess(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	req, _ := http.NewRequest("GET", c.baseURL, nil)
	resp, err := c.executeWithRetry(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestExecuteWithRetry429(t *testing.T) {
	attempts := 0
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(429)
	})
	req, _ := http.NewRequest("GET", c.baseURL, nil)
	_, err := c.executeWithRetry(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, ErrRateLimited, err)
}

func TestSearchRequestJsonDocMock(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0.0","g":"g","a":"a","v":"1.0.0"}]}}`))
	})
	sr := request.NewSearchRequest().SetLimit(5).SetQuery(request.NewQuery().SetGroupId("g"))
	resp, err := SearchRequestJsonDoc[*response.Artifact](c, context.Background(), sr)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.ResponseBody.NumFound)
	assert.Len(t, resp.ResponseBody.Docs, 1)
	assert.Equal(t, "g:a:1.0.0", resp.ResponseBody.Docs[0].ID)
}

func TestSearchByArtifactIdMock(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0.0","g":"g","a":"a","v":"1.0.0"}]}}`))
	})
	results, err := c.SearchByArtifactId(context.Background(), "a", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestSearchByGroupIdMock(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"numFound":0,"start":0,"docs":[]}}`))
	})
	results, err := c.SearchByGroupId(context.Background(), "com.example", 5)
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestDownloadMock(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("JAR-CONTENT"))
	})
	data, err := c.Download(context.Background(), "com/example/lib/1.0.0/lib-1.0.0.jar")
	assert.NoError(t, err)
	assert.Equal(t, "JAR-CONTENT", string(data))
}

func TestDownload404(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	_, err := c.Download(context.Background(), "com/example/lib/1.0.0/lib-1.0.0.jar")
	assert.Error(t, err)
}

func TestParseJsonResponseSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":"ok"}`))
	}))
	defer srv.Close()
	resp, _ := http.Get(srv.URL)
	var result struct{ Value string }
	err := parseJsonResponse(resp, &result)
	assert.NoError(t, err)
	assert.Equal(t, "ok", result.Value)
}

func TestParseJsonResponseNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer srv.Close()
	resp, _ := http.Get(srv.URL)
	err := parseJsonResponse(resp, nil)
	assert.Error(t, err)
}

func TestRateLimiterNewAndStats(t *testing.T) {
	rl := NewRateLimiter()
	_, err := rl.WaitForRateLimit(context.Background(), "host", "search")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), rl.GetTotalRequestCount("host"))
	assert.Equal(t, int64(1), rl.GetRequestCountByType("host", "search"))
	rl.ResetStats()
	assert.Equal(t, int64(0), rl.GetTotalRequestCount("host"))
}

func TestRateLimiterGetStats(t *testing.T) {
	rl := NewRateLimiter()
	_, _ = rl.WaitForRateLimit(context.Background(), "h", "search")
	m := rl.GetStats()
	assert.NotNil(t, m)
	assert.True(t, m["stats_enabled"].(bool))
}

func TestRateLimiterDisabledStats(t *testing.T) {
	config := DefaultRateLimitConfig
	config.EnableStats = false
	rl := NewRateLimiterWithConfig(config)
	_, _ = rl.WaitForRateLimit(context.Background(), "h", "search")
	assert.Equal(t, int64(0), rl.GetTotalRequestCount("h"))
	assert.Equal(t, int64(0), rl.GetRequestCountByType("h", "search"))
	m := rl.GetStats()
	assert.False(t, m["stats_enabled"].(bool))
}

func TestClientOptionsMock(t *testing.T) {
	c := NewClient(
		WithBaseURL("https://example.com"),
		WithRepoBaseURL("https://repo.example.com"),
		WithMaxRetries(5),
		WithRetryBackoff(100),
		WithCache(true, 60),
		WithProxy("http://proxy:8080"),
	)
	assert.Equal(t, "https://example.com", c.GetBaseURL())
	assert.Equal(t, "https://repo.example.com", c.GetRepoBaseURL())
	assert.Equal(t, 60, c.GetCacheTTL())
	assert.True(t, c.IsCacheEnabled())
	c.SetCacheTTL(120)
	assert.Equal(t, 120, c.GetCacheTTL())
	c.DisableCache()
	assert.False(t, c.IsCacheEnabled())
	c.EnableCache()
	assert.True(t, c.IsCacheEnabled())
	c.ClearCache()
}

func TestRetryWithBackoffOnce(t *testing.T) {
	calls := 0
	err := RetryWithBackoff(context.Background(), 3, 1, 2.0, 1000, func() error {
		calls++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestRetryWithBackoffAllFail(t *testing.T) {
	calls := 0
	err := RetryWithBackoff(context.Background(), 2, 1, 2.0, 1000, func() error {
		calls++
		return &response.HTTPError{StatusCode: 429}
	})
	assert.Error(t, err)
	assert.Equal(t, 3, calls)
}

func TestSearchIteratorNew(t *testing.T) {
	sr := request.NewSearchRequest().SetLimit(5)
	it := NewSearchIterator[*response.Artifact](sr)
	assert.NotNil(t, it)
	it2 := it.WithClient(NewClient())
	assert.NotNil(t, it2)
}

func TestWithHTTPClient(t *testing.T) {
	customTransport := &http.Transport{}
	customClient := &http.Client{Transport: customTransport}
	c := NewClient(WithHTTPClient(customClient))
	assert.Equal(t, customClient, c.httpClient)
}

func TestAsyncSearchRequestJsonDoc(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0.0"}]}}`))
	})
	sr := request.NewSearchRequest().SetLimit(5).SetQuery(request.NewQuery().SetGroupId("g"))
	ch := AsyncSearchRequestDoc[*response.Artifact](c, context.Background(), sr)
	result := <-ch
	assert.NoError(t, result.Error)
	assert.NotNil(t, result.Result)
	assert.Equal(t, 1, result.Result.ResponseBody.NumFound)
}

func TestDownloadWithCache(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("CACHED-DATA"))
	})
	// Enable cache for this test
	c.cacheEnabled = true
	c.cacheTTLSeconds = 60
	data, err := c.downloadWithCache(context.Background(), "com/example/lib/1.0.0/lib-1.0.0.jar")
	assert.NoError(t, err)
	assert.Equal(t, "CACHED-DATA", string(data))
}

func TestDownloadWithChecksum(t *testing.T) {
	// DownloadWithChecksum 需要两次请求（.sha1 和实际文件），
	// 单 httptest handler 难以区分
	t.Skip("need two mock endpoints for checksum test")
}

func TestDownloadToWriter(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("writer-content"))
	})
	var buf bytes.Buffer
	err := c.DownloadToWriter(context.Background(), "com/example/lib/1.0.0/lib-1.0.0.jar", &buf)
	assert.NoError(t, err)
	assert.Equal(t, "writer-content", buf.String())
}

func TestDownloadFile(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("file-content"))
	})
	tmpDir := t.TempDir()
	localPath := tmpDir + "/test.jar"
	err := c.DownloadFile(context.Background(), "com/example/lib/1.0.0/lib-1.0.0.jar", localPath)
	assert.NoError(t, err)
}

func TestDownloadJar(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/java-archive")
		_, _ = w.Write([]byte("jar-content"))
	})
	data, err := c.DownloadJar(context.Background(), "com.example", "lib", "1.0.0")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestDownloadPom(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("<project>...</project>"))
	})
	data, err := c.DownloadPom(context.Background(), "com.example", "lib", "1.0.0")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestDownloadSources(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("sources-content"))
	})
	data, err := c.DownloadSources(context.Background(), "com.example", "lib", "1.0.0")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestDownloadJavadoc(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("javadoc-content"))
	})
	data, err := c.DownloadJavadoc(context.Background(), "com.example", "lib", "1.0.0")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestSearchByText(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"numFound":2,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0"}]}}`))
	})
	results, err := c.SearchByText(context.Background(), "test", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestSearchByTag(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0"}]}}`))
	})
	results, err := c.SearchByTag(context.Background(), "spring", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestSearchBySha1(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0"}]}}`))
	})
	results, err := c.SearchBySha1(context.Background(), "abc123", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestSearchByClassName(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0"}]}}`))
	})
	results, err := c.SearchByClassName(context.Background(), "TestClass", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestSearchByGroupAndArtifactId(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0"}]}}`))
	})
	results, err := c.SearchByGroupAndArtifactId(context.Background(), "g", "a", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}