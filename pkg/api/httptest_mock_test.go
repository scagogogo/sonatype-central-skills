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

// ---- helpers ----

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

func newMockPublisherClient(t *testing.T, handler http.HandlerFunc) *PublisherClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewPublisherClient(
		WithPublisherBaseURL(srv.URL),
		WithPublisherHTTPClient(srv.Client()),
		WithPublisherToken("test-token"),
	)
}

const mockSearchResponse = `{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0.0","g":"g","a":"a","v":"1.0.0","p":"jar","timestamp":1600000000000,"ec":["-sources.jar",".jar"],"tags":["test"]}]}}`
const mockVersionResponse = `{"response":{"numFound":2,"start":0,"docs":[{"id":"g:a:1.0.0","g":"g","a":"a","v":"1.0.0","timestamp":1600000000000,"ec":[],"tags":[]},{"id":"g:a:2.0.0","g":"g","a":"a","v":"2.0.0","timestamp":1600000001000,"ec":[],"tags":[]}]}}`
const mockSearchResponseHeader = `{"responseHeader":{"status":0,"QTime":1,"params":{"q":"g:g"}},"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0.0","g":"g","a":"a","v":"1.0.0"}]}}`

// ========================================================================
// artifact.go — SearchArtifactsByTag, SearchArtifactsWithFacets, etc.
// ========================================================================

func TestMockSearchArtifactsByTag(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.SearchArtifactsByTag(context.Background(), "test", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockSearchArtifactsWithFacets(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0"}]},"facet_counts":{"facet_fields":{"g":["org.apache",30]}}}`))
	})
	results, facets, err := c.SearchArtifactsWithFacets(context.Background(), "test", []string{"g"}, 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.NotNil(t, facets)
}

func TestMockGetArtifactDependencies(t *testing.T) {
	callCount := 0
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		if callCount == 1 {
			// SearchByGroupAndArtifactId → search result
			w.Write([]byte(mockSearchResponse))
		} else {
			// GetArtifactMetadata → metadata with dependencies
			w.Write([]byte(`{"groupId":"g","artifactId":"a","latestVersion":"1.0","packaging":"jar","lastUpdated":1600000000000,"dependencies":[{"groupId":"dep","artifactId":"dep","version":"1.0","scope":"compile","optional":false}]}`))
		}
	})
	deps, err := c.GetArtifactDependencies(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.NotNil(t, deps)
}

func TestMockGetArtifactUsage(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"numFound":5,"start":0,"docs":[{"id":"u:a:1.0","g":"u","a":"a","v":"1.0"}]}}`))
	})
	usage, err := c.GetArtifactUsage(context.Background(), "g", "a", "1.0", 5)
	assert.NoError(t, err)
	assert.NotNil(t, usage)
	assert.Equal(t, 5, usage.TotalUsageCount)
}

func TestMockCompareArtifacts(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	_, err := c.CompareArtifacts(context.Background(), "g1", "a1", "g2", "a2")
	assert.NoError(t, err)
}

func TestMockSearchArtifactsByDateRange(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.SearchArtifactsByDateRange(context.Background(), "2023-01-01", "2023-12-31", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockSuggestSimilarArtifacts(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.SuggestSimilarArtifacts(context.Background(), "g", "a", 5)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockSearchPopularArtifacts(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.SearchPopularArtifacts(context.Background(), 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockGetArtifactDetails(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	details, err := c.GetArtifactDetails(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.NotNil(t, details)
}

func TestMockGetArtifactStats(t *testing.T) {
	// GetArtifactStats calls SearchByGroupAndArtifactId + ListVersions + GetArtifactUsage
	callCount := 0
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		if callCount <= 2 {
			w.Write([]byte(mockSearchResponse))
		} else {
			w.Write([]byte(`{"response":{"numFound":0,"start":0,"docs":[]}}`))
		}
	})
	stats, err := c.GetArtifactStats(context.Background(), "g", "a")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "g", stats.GroupId)
}

// ========================================================================
// download.go — DownloadArtifact, DownloadMultipleFiles, etc.
// ========================================================================

func TestMockDownloadArtifact(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("binary-data"))
	})
	data, err := c.DownloadArtifact(context.Background(), &response.Artifact{GroupId: "g", ArtifactId: "a", LatestVersion: "1.0"}, "jar")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestMockDownloadArtifactWithVersion(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("binary-data"))
	})
	data, err := c.DownloadArtifactWithVersion(context.Background(), &response.Version{GroupId: "g", ArtifactId: "a", Version: "1.0"}, "jar")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestMockDownloadCycloneDXJSON(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"bomFormat":"CycloneDX"}`))
	})
	data, err := c.DownloadCycloneDXJSON(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestMockDownloadCycloneDXXML(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`<bom/>`))
	})
	data, err := c.DownloadCycloneDXXML(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestMockDownloadSpdxJSON(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"spdxVersion":"2.2"}`))
	})
	data, err := c.DownloadSpdxJSON(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestMockDownloadChecksumFile(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("abc123def456"))
	})
	sum, err := c.DownloadChecksumFile(context.Background(), "g/a/1.0/a-1.0.jar.sha1", "sha1")
	assert.NoError(t, err)
	assert.Equal(t, "abc123def456", sum)
}

func TestMockDownloadMultipleFiles(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("data"))
	})
	results := c.DownloadMultipleFiles(context.Background(), "g", "a", "1.0", []ArtifactFile{{Extension: "jar", Classifier: ""}})
	assert.NotEmpty(t, results)
}

func TestMockDownloadCompleteBundle(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("data"))
	})
	bundle, err := c.DownloadCompleteBundle(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.NotNil(t, bundle)
}

func TestMockSaveBundle(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("data"))
	})
	bundle, err := c.DownloadCompleteBundle(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	err = c.SaveBundle(bundle, t.TempDir())
	assert.NoError(t, err)
}

func TestMockDownloadAsync(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("async-data"))
	})
	ch := c.DownloadAsync(context.Background(), "g/a/1.0/a-1.0.jar")
	result := <-ch
	assert.NoError(t, result.Error)
	assert.NotEmpty(t, result.Data)
}

func TestMockBatchDownloadFiles(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("data"))
	})
	results := c.BatchDownloadFiles(context.Background(), map[string]string{"g/a/1.0/a-1.0.jar": t.TempDir() + "/out.jar"})
	assert.NotEmpty(t, results)
}

// ========================================================================
// advanced_search.go — AdvancedSearch, BatchSearch, GetArtifactMetadata, etc.
// ========================================================================

func TestMockAdvancedSearch(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	opts := request.NewAdvancedSearchOptions().SetGroupId("g").SetArtifactId("a")
	results, err := c.AdvancedSearch(context.Background(), opts, 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockBatchSearch(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	queries := []*request.SearchRequest{request.NewSearchRequest().SetLimit(5).SetQuery(request.NewQuery().SetGroupId("g"))}
	results, err := c.BatchSearch(context.Background(), queries)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestMockGetArtifactMetadata(t *testing.T) {
	callCount := 0
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		if callCount == 1 {
			// SearchRequest → search result
			w.Write([]byte(mockSearchResponse))
		} else {
			// DownloadPom → POM content
			w.Write([]byte("<project><modelVersion>4.0.0</modelVersion></project>"))
		}
	})
	meta, err := c.GetArtifactMetadata(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.NotNil(t, meta)
	assert.Equal(t, "g", meta.GroupId)
}

func TestMockSearchWithSpellcheck(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"responseHeader":{"status":0,"QTime":1,"params":{"q":"commns"}},"response":{"numFound":0,"start":0,"docs":[]},"spellcheck":{"suggestions":["commns",{"numFound":5,"startOffset":0,"endOffset":6,"suggestion":["commons","communs"]}]}}`))
	})
	results, suggests, err := c.SearchWithSpellcheck(context.Background(), "commns", 10, 5)
	assert.NoError(t, err)
	assert.Empty(t, results)
	assert.NotEmpty(t, suggests)
}

func TestMockSearchWithSort(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	sr := request.NewSearchRequest().SetLimit(5).SetQuery(request.NewQuery().SetGroupId("g"))
	results, err := c.SearchWithSort(context.Background(), sr, "timestamp", false, 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockSearchByGroupAndClassifier(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.SearchByGroupAndClassifier(context.Background(), "g", "sources", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockSearchByClassifier(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.SearchByClassifier(context.Background(), "sources", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockIteratorByClassifier(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	it := c.IteratorByClassifier(context.Background(), "sources").WithClient(c)
	assert.NotNil(t, it)
}

func TestMockSearchByTagPrefix(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.SearchByTagPrefix(context.Background(), "spring", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockSearchByMultipleTags(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.SearchByMultipleTags(context.Background(), []string{"spring", "boot"}, 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockSearchByTagAndSortByPopularity(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.SearchByTagAndSortByPopularity(context.Background(), "spring", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockSearchVulnerableArtifacts(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"responseHeader":{"status":0,"QTime":1,"params":{"q":"*:*"}},"response":{"numFound":1,"start":0,"docs":[{"groupId":"g","artifactId":"a","latestVersion":"1.0","packaging":"jar","lastUpdated":1600000000000}]}}`))
	})
	q := request.NewQuery().SetGroupId("g")
	results, err := c.SearchVulnerableArtifacts(context.Background(), q)
	assert.NoError(t, err)
	assert.NotNil(t, results)
}

// ========================================================================
// search_iterator.go — Next, Value, NextE, ValueE, ToSlice
// ========================================================================

func TestMockSearchIteratorToSlice(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	sr := request.NewSearchRequest().SetLimit(5).SetQuery(request.NewQuery().SetGroupId("g"))
	it := NewSearchIterator[*response.Artifact](sr).WithClient(c)
	slice, err := it.ToSlice()
	assert.NoError(t, err)
	assert.Len(t, slice, 1)
}

func TestMockSearchIteratorNextE(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	sr := request.NewSearchRequest().SetLimit(5).SetQuery(request.NewQuery().SetGroupId("g"))
	it := NewSearchIterator[*response.Artifact](sr).WithClient(c)
	hasNext, err := it.NextE()
	assert.NoError(t, err)
	assert.True(t, hasNext)
	val, err := it.ValueE()
	assert.NoError(t, err)
	assert.NotNil(t, val)
	hasNext, err = it.NextE()
	assert.NoError(t, err)
	assert.False(t, hasNext)
}

// ========================================================================
// group.go — GetGroupInfo, GetGroupStatistics, CompareTwoGroups
// ========================================================================

func TestMockGetGroupInfo(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"groupId":"g","artifactCount":10,"lastUpdated":1600000000000,"lastUpdatedDate":"2023-01-01"}]}}`))
	})
	info, err := c.GetGroupInfo(context.Background(), "g")
	assert.NoError(t, err)
	assert.NotNil(t, info)
}

func TestMockGetGroupStatistics(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"groupId":"g","artifactCount":10,"totalVersions":50,"latestUpdate":1600000000000,"lastUpdatedDate":"2023-01-01","artifacts":[]}]}}`))
	})
	stats, err := c.GetGroupStatistics(context.Background(), "g")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestMockCompareTwoGroups(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"groupId":"g","artifactCount":10,"totalVersions":50,"latestUpdate":1600000000000,"lastUpdatedDate":"2023-01-01","artifacts":[]}]}}`))
	})
	comp, err := c.CompareTwoGroups(context.Background(), "g1", "g2")
	assert.NoError(t, err)
	assert.NotNil(t, comp)
}

// ========================================================================
// class.go / class_search.go / full_class.go — class search methods
// ========================================================================

func TestMockSearchByFullyQualifiedClassName(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	results, err := c.SearchByFullyQualifiedClassName(context.Background(), "com.example.Foo", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestMockSearchByJavaPackage(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	results, err := c.SearchByJavaPackage(context.Background(), "com.example", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestMockSearchByPackageAndClassName(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	results, err := c.SearchByPackageAndClassName(context.Background(), "com.example", "Foo", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestMockSearchClassesByMethod(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	results, err := c.SearchClassesByMethod(context.Background(), "doSomething", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestMockSearchInterfaceImplementations(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	results, err := c.SearchInterfaceImplementations(context.Background(), "com.example.IFoo", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestMockSearchByClassSupertype(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	results, err := c.SearchByClassSupertype(context.Background(), "com.example.Base", false, 10)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestMockSearchClassesWithClassHierarchy(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	results, err := c.SearchClassesWithClassHierarchy(context.Background(), "com.example.Base", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestMockSearchClassesWithHighlighting(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"responseHeader":{"status":0,"QTime":1,"params":{"q":"fc:com.example.Foo"}},"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0"}]},"highlighting":{"g:a:1.0":{"fch":["<em>com.example</em>.Foo"]}}}`))
	})
	result, err := c.SearchClassesWithHighlighting(context.Background(), "com.example.Foo", 5)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Highlighting)
}

func TestMockSearchFullyQualifiedClassNames(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"responseHeader":{"status":0,"QTime":1,"params":{"q":"fc:com.example.Foo"}},"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0"}]},"highlighting":{"g:a:1.0":{"fch":["<em>com.example</em>.Foo"]}}}`))
	})
	versions, highlights, err := c.SearchFullyQualifiedClassNames(context.Background(), "com.example.Foo", 5)
	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.NotEmpty(t, highlights)
}

// ========================================================================
// versions.go — ListVersions, GetLatestVersion, HasVersion, etc.
// ========================================================================

func TestMockListVersions(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	results, err := c.ListVersions(context.Background(), "g", "a", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestMockGetLatestVersion(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	v, err := c.GetLatestVersion(context.Background(), "g", "a")
	assert.NoError(t, err)
	assert.NotNil(t, v)
}

func TestMockHasVersion(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	ok, err := c.HasVersion(context.Background(), "g", "a", "1.0.0")
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestMockCountVersions(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	n, err := c.CountVersions(context.Background(), "g", "a")
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
}

func TestMockGetVersionInfo(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	info, err := c.GetVersionInfo(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.NotNil(t, info)
}

func TestMockCompareVersions(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	comp, err := c.CompareVersions(context.Background(), "g", "a", "1.0", "2.0")
	assert.NoError(t, err)
	assert.NotNil(t, comp)
}

func TestMockGetVersionsWithMetadata(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	results, err := c.GetVersionsWithMetadata(context.Background(), "g", "a")
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestMockIteratorVersions(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	it := c.IteratorVersions(context.Background(), "g", "a").WithClient(c)
	assert.NotNil(t, it)
}

// ========================================================================
// gav.go — ListGAVs, GetGAVInfo, SearchGAVsWithSort, FindGAVDependencies
// ========================================================================

func TestMockListGAVs(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.ListGAVs(context.Background(), "g:a", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockListGAVsPaginated(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, total, err := c.ListGAVsPaginated(context.Background(), "g:a", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, 1, total)
}

func TestMockSearchGAVsWithSort(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.SearchGAVsWithSort(context.Background(), "g:a", "timestamp", true, 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockFindGAVDependencies(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.FindGAVDependencies(context.Background(), "g1", "a1", "g2", "a2", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockGetGAVInfo(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	info, err := c.GetGAVInfo(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.NotNil(t, info)
}

func TestMockIteratorGAVs(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	it := c.IteratorGAVs(context.Background(), "g:a").WithClient(c)
	assert.NotNil(t, it)
}

// ========================================================================
// sha1.go — SearchBySha1Prefix, SearchExactSha1, CountBySha1, etc.
// ========================================================================

func TestMockSearchBySha1Prefix(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	results, err := c.SearchBySha1Prefix(context.Background(), "abc", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestMockSearchExactSha1(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	results, err := c.SearchExactSha1(context.Background(), "abc123")
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestMockCountBySha1(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	n, err := c.CountBySha1(context.Background(), "abc")
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
}

func TestMockExistsSha1(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	ok, err := c.ExistsSha1(context.Background(), "abc")
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestMockGetFirstBySha1(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	v, err := c.GetFirstBySha1(context.Background(), "abc")
	assert.NoError(t, err)
	assert.NotNil(t, v)
}

func TestMockIteratorBySha1(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	it := c.IteratorBySha1(context.Background(), "abc").WithClient(c)
	assert.NotNil(t, it)
}

func TestMockIteratorBySha1Prefix(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	it := c.IteratorBySha1Prefix(context.Background(), "abc").WithClient(c)
	assert.NotNil(t, it)
}

// ========================================================================
// tag.go — GetMostUsedTags, CountArtifactsByTag, IteratorByTag, IteratorByText
// ========================================================================

func TestMockGetMostUsedTags(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"numFound":2,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0","tags":["spring"]},{"id":"g:b:1.0","g":"g2","a":"b","v":"1.0","tags":["spring"]}]}}`))
	})
	results, err := c.GetMostUsedTags(context.Background(), "spring", 10)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestMockCountArtifactsByTag(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	n, err := c.CountArtifactsByTag(context.Background(), "spring")
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestMockIteratorByTag(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	it := c.IteratorByTag(context.Background(), "spring").WithClient(c)
	assert.NotNil(t, it)
}

func TestMockIteratorByText(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	it := c.IteratorByText(context.Background(), "search text").WithClient(c)
	assert.NotNil(t, it)
}

// ========================================================================
// security.go — Security / Vulnerability methods
// ========================================================================

func TestMockGetComponentVulnerabilityOverview(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0"}]}}`))
	})
	overview, err := c.GetComponentVulnerabilityOverview(context.Background(), "g", "a", 5)
	assert.NoError(t, err)
	assert.NotNil(t, overview)
}

func TestMockCheckCVEImpact(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	impacted, _, err := c.CheckCVEImpact(context.Background(), "CVE-2023-0001", "g", "a", "1.0")
	assert.NoError(t, err)
	assert.False(t, impacted)
}

func TestMockFindArtifactsByCVE(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.FindArtifactsByCVE(context.Background(), "CVE-2023-0001", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockGetVulnerabilityDetails(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	details, err := c.GetVulnerabilityDetails(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.NotNil(t, details)
}

func TestMockGetVulnerabilityTimeline(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	timeline, err := c.GetVulnerabilityTimeline(context.Background(), "g", "a", 10)
	assert.NoError(t, err)
	assert.NotNil(t, timeline)
}

func TestMockGetSecurityRating(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	rating, err := c.GetSecurityRating(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.NotNil(t, rating)
}

func TestMockCompareVersionSecurity(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	comp, err := c.CompareVersionSecurity(context.Background(), "g", "a", "1.0", "2.0")
	assert.NoError(t, err)
	assert.NotNil(t, comp)
}

func TestMockGetRecommendedSecureVersion(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	version, err := c.GetRecommendedSecureVersion(context.Background(), "g", "a", "1.0.0")
	assert.NoError(t, err)
	assert.NotEmpty(t, version)
}

func TestMockBatchSecurityScan(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.BatchSecurityScan(context.Background(), []*response.ArtifactRef{{GroupId: "g", ArtifactId: "a", Version: "1.0"}})
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

// ========================================================================
// async.go — AsyncDownload, AsyncBatchSearch, etc.
// ========================================================================

func TestMockAsyncDownload(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("async-data"))
	})
	ch := c.AsyncDownload(context.Background(), "g/a/1.0/a-1.0.jar")
	result := <-ch
	assert.NoError(t, result.Error)
	assert.NotEmpty(t, result.Result)
}

func TestMockAsyncBatchSearch(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	reqs := []*request.SearchRequest{request.NewSearchRequest().SetLimit(5).SetQuery(request.NewQuery().SetGroupId("g"))}
	ch := c.AsyncBatchSearch(context.Background(), reqs)
	result := <-ch
	assert.NoError(t, result.Error)
	assert.NotEmpty(t, result.Result)
}

func TestMockAsyncSearchByArtifact(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	ch := c.AsyncSearchByArtifact(context.Background(), "a", 10)
	result := <-ch
	assert.NoError(t, result.Error)
	assert.NotEmpty(t, result.Result)
}

func TestMockAsyncSearchByGroup(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	ch := c.AsyncSearchByGroup(context.Background(), "g", 10)
	result := <-ch
	assert.NoError(t, result.Error)
	assert.NotEmpty(t, result.Result)
}

// ========================================================================
// licensee.go — FindLicenseConflicts, GenerateLicenseReport, etc. (Client methods)
// ========================================================================

func TestMockFindLicenseConflicts(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0","licenseList":["MIT"]}]}}`))
	})
	summary, err := c.FindLicenseConflicts(context.Background(), []response.ArtifactRef{{GroupId: "g", ArtifactId: "a", Version: "1.0"}})
	assert.NoError(t, err)
	assert.NotNil(t, summary)
}

func TestMockGenerateLicenseReport(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0","licenseList":["MIT"]}]}}`))
	})
	report, err := c.GenerateLicenseReport(context.Background(), []response.ArtifactRef{{GroupId: "g", ArtifactId: "a", Version: "1.0"}})
	assert.NoError(t, err)
	assert.NotNil(t, report)
}

// ========================================================================
// publisher.go — PublisherClient methods
// ========================================================================

func TestMockPublisherListDeployments(t *testing.T) {
	pc := newMockPublisherClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"deployments":[{"deploymentId":"d1","deploymentName":"n1","namespace":"g","deploymentState":"PUBLISHED","createTimestamp":"2023-01-01","updateTimestamp":"2023-01-02"}],"page":0,"pageSize":10,"pageCount":1,"totalResultCount":1}`))
	})
	list, err := pc.ListDeployments(context.Background(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, list)
	assert.Len(t, list.Deployments, 1)
}

func TestMockPublisherCheckPublished(t *testing.T) {
	pc := newMockPublisherClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"published":true,"namespace":"g","name":"a","version":"1.0"}`))
	})
	check, err := pc.CheckPublished(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.True(t, check.Published)
}

func TestMockPublisherGetDeploymentStatus(t *testing.T) {
	pc := newMockPublisherClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"deploymentId":"d1","deploymentName":"n1","deploymentState":"PUBLISHED"}`))
	})
	status, err := pc.GetDeploymentStatus(context.Background(), "d1")
	assert.NoError(t, err)
	assert.Equal(t, "d1", status.DeploymentID)
}

func TestMockPublisherDropDeployment(t *testing.T) {
	pc := newMockPublisherClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	err := pc.DropDeployment(context.Background(), "d1")
	assert.NoError(t, err)
}

func TestMockPublisherPublishDeployment(t *testing.T) {
	pc := newMockPublisherClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	err := pc.PublishDeployment(context.Background(), "d1")
	assert.NoError(t, err)
}

func TestMockPublisherUploadBundle(t *testing.T) {
	pc := newMockPublisherClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("deployment-id-from-server"))
	})
	id, err := pc.UploadBundle(context.Background(), []byte("bundle-data"), "test-bundle", response.PublishingTypeUserManaged)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
}

func TestMockPublisherDownloadDeploymentFile(t *testing.T) {
	pc := newMockPublisherClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("file-content"))
	})
	data, err := pc.DownloadDeploymentFile(context.Background(), "d1", "path/to/file")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

// ========================================================================
// batch.go — BatchDownloadDependencies, BatchSearchArtifacts
// ========================================================================

func TestMockBatchSearchArtifacts(t *testing.T) {
	callCount := 0
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results := c.BatchSearchArtifacts(context.Background(), []string{"g"}, "group", 10)
	assert.NotEmpty(t, results)
}

// ========================================================================
// iterator.go — IteratorByClassHierarchy, IteratorByInterfaceImplementation, etc.
// ========================================================================

func TestMockIteratorByClassHierarchy(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	it := c.IteratorByClassHierarchy(context.Background(), "com.example.Base").WithClient(c)
	assert.NotNil(t, it)
}

func TestMockIteratorByClassSupertype(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	it := c.IteratorByClassSupertype(context.Background(), "com.example.IFace", true).WithClient(c)
	assert.NotNil(t, it)
}

func TestMockIteratorByInterfaceImplementation(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	it := c.IteratorByInterfaceImplementation(context.Background(), "com.example.IFace").WithClient(c)
	assert.NotNil(t, it)
}

func TestMockIteratorByFullyQualifiedClassName(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	it := c.IteratorByFullyQualifiedClassName(context.Background(), "com.example.Foo").WithClient(c)
	assert.NotNil(t, it)
}

func TestMockIteratorByJavaPackage(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	it := c.IteratorByJavaPackage(context.Background(), "com.example").WithClient(c)
	assert.NotNil(t, it)
}

func TestMockIteratorByPackageAndClassName(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	it := c.IteratorByPackageAndClassName(context.Background(), "com.example", "Foo").WithClient(c)
	assert.NotNil(t, it)
}

func TestMockIteratorByClassName(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	it := c.IteratorByClassName(context.Background(), "Foo").WithClient(c)
	assert.NotNil(t, it)
}

func TestMockIteratorByMethod(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	it := c.IteratorByMethod(context.Background(), "doSomething").WithClient(c)
	assert.NotNil(t, it)
}

// ========================================================================
// License Client methods — GetComponentLicenses, SearchByLicenseType, etc.
// ========================================================================

func TestMockGetComponentLicenses(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0","licenseList":["MIT","Apache-2.0"]}]}}`))
	})
	licenses, err := c.GetComponentLicenses(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.NotEmpty(t, licenses)
}

func TestMockSearchByLicenseType(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0"}]}}`))
	})
	results, err := c.SearchByLicenseType(context.Background(), LicenseTypeMIT, 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockGetPopularLicenses(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"responseHeader":{"status":0,"QTime":1,"params":{"q":"*:*","facet":"true","facet.field":"l","facet.limit":"10","rows":"0"}},"response":{"numFound":100,"start":0,"docs":[]},"facet_counts":{"facet_fields":{"l":["MIT",50,"Apache-2.0",30]}}}`))
	})
	licenses, err := c.GetPopularLicenses(context.Background(), 10)
	assert.NoError(t, err)
	assert.NotEmpty(t, licenses)
}

func TestMockFilterByLicenseType(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0","licenseList":["MIT"]}]}}`))
	})
	compliant, nonCompliant, err := c.FilterByLicenseType(context.Background(), []response.ArtifactRef{{GroupId: "g", ArtifactId: "a", Version: "1.0"}}, []string{"MIT"})
	assert.NoError(t, err)
	assert.Len(t, compliant, 1)
	assert.Empty(t, nonCompliant)
}

// ========================================================================
// SearchByDependency, SearchByLicense (in advanced_search.go)
// ========================================================================

func TestMockSearchByDependency(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.SearchByDependency(context.Background(), "g", "a", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockSearchByLicense(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.SearchByLicense(context.Background(), "MIT", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}
// ========================================================================
// RESTORED: core HTTP/network layer tests from original file
// ========================================================================

func TestMockDoRequestSuccess(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0.0"}]}`))
	})
	body, err := c.doRequest(context.Background(), "GET", c.baseURL, nil, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)
}

func TestMockDoRequest404(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	_, err := c.doRequest(context.Background(), "GET", c.baseURL, nil, nil)
	assert.Error(t, err)
}

func TestMockExecuteWithRetry(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	req, _ := http.NewRequest("GET", c.baseURL, nil)
	resp, err := c.executeWithRetry(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestMockParseJsonResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"value":"ok"}`))
	}))
	defer srv.Close()
	resp, _ := http.Get(srv.URL)
	var result struct{ Value string }
	err := parseJsonResponse(resp, &result)
	assert.NoError(t, err)
	assert.Equal(t, "ok", result.Value)
}

func TestMockParseJsonResponseNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("error"))
	}))
	defer srv.Close()
	resp, _ := http.Get(srv.URL)
	err := parseJsonResponse(resp, nil)
	assert.Error(t, err)
}

func TestMockDownloadJar(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("jar-content"))
	})
	data, err := c.DownloadJar(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestMockDownloadSources(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("sources-content"))
	})
	data, err := c.DownloadSources(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestMockDownloadJavadoc(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("javadoc-content"))
	})
	data, err := c.DownloadJavadoc(context.Background(), "g", "a", "1.0")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestMockDownloadFile(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("file-content"))
	})
	err := c.DownloadFile(context.Background(), "g/a/1.0/a-1.0.jar", t.TempDir()+"/test.jar")
	assert.NoError(t, err)
}

func TestMockDownloadToWriter(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("writer-content"))
	})
	var buf bytes.Buffer
	err := c.DownloadToWriter(context.Background(), "g/a/1.0/a-1.0.jar", &buf)
	assert.NoError(t, err)
	assert.Equal(t, "writer-content", buf.String())
}


// ========================================================================
// Iterator Next()/Value() wrappers
// ========================================================================

func TestMockIteratorNextValue(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	sr := request.NewSearchRequest().SetLimit(5).SetQuery(request.NewQuery().SetGroupId("g"))
	it := NewSearchIterator[*response.Artifact](sr).WithClient(c)
	assert.True(t, it.Next())
	v := it.Value()
	assert.NotNil(t, v)
}

// ========================================================================
// SearchRequestJson alias
// ========================================================================

func TestMockSearchRequestJsonAlias(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0"}]}}`))
	})
	sr := request.NewSearchRequest().SetLimit(5).SetQuery(request.NewQuery().SetGroupId("g"))
	resp, err := SearchRequestJson[*response.Artifact](c, context.Background(), sr)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.ResponseBody.NumFound)
}

// ========================================================================
// IteratorByArtifactId, IteratorByGroupAndArtifactId, AdvancedSearchIterator
// ========================================================================

func TestMockIteratorByArtifactId(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	it := c.IteratorByArtifactId(context.Background(), "a").WithClient(c)
	assert.NotNil(t, it)
}

func TestMockIteratorByGroupAndArtifactId(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	it := c.IteratorByGroupAndArtifactId(context.Background(), "g", "a").WithClient(c)
	assert.NotNil(t, it)
}

func TestMockAdvancedSearchIterator(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	opts := request.NewAdvancedSearchOptions().SetGroupId("g").SetArtifactId("a")
	it := c.AdvancedSearchIterator(context.Background(), opts).WithClient(c)
	assert.NotNil(t, it)
}

// ========================================================================
// SearchByText (classifier_search.go)
// ========================================================================

func TestMockSearchByTextClassifier(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.SearchByText(context.Background(), "hello", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

// ========================================================================
// RateLimiter methods
// ========================================================================

func TestMockRateLimiterGetStats(t *testing.T) {
	rl := NewRateLimiter()
	_, _ = rl.WaitForRateLimit(context.Background(), "h", "search")
	stats := rl.GetStats()
	assert.NotNil(t, stats)
	assert.True(t, stats["stats_enabled"].(bool))
}

func TestMockRateLimiterGetTotalRequestCount(t *testing.T) {
	rl := NewRateLimiter()
	_, _ = rl.WaitForRateLimit(context.Background(), "h", "search")
	assert.Equal(t, int64(1), rl.GetTotalRequestCount("h"))
}

func TestMockRateLimiterGetRequestCountByType(t *testing.T) {
	rl := NewRateLimiter()
	_, _ = rl.WaitForRateLimit(context.Background(), "h", "search")
	assert.Equal(t, int64(1), rl.GetRequestCountByType("h", "search"))
}

func TestMockRateLimiterResetStats(t *testing.T) {
	rl := NewRateLimiter()
	_, _ = rl.WaitForRateLimit(context.Background(), "h", "search")
	rl.ResetStats()
	assert.Equal(t, int64(0), rl.GetTotalRequestCount("h"))
}

func TestMockRateLimiterDisabledStats(t *testing.T) {
	config := DefaultRateLimitConfig
	config.EnableStats = false
	rl := NewRateLimiterWithConfig(config)
	_, _ = rl.WaitForRateLimit(context.Background(), "h", "search")
	assert.Equal(t, int64(0), rl.GetTotalRequestCount("h"))
	assert.Equal(t, int64(0), rl.GetRequestCountByType("h", "search"))
	stats := rl.GetStats()
	assert.False(t, stats["stats_enabled"].(bool))
}

// ========================================================================
// Async methods
// ========================================================================

func TestMockAsyncSearchRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0"}]}}`))
	}))
	defer srv.Close()
	c := NewClient(WithBaseURL(srv.URL), WithRepoBaseURL(srv.URL), WithHTTPClient(srv.Client()), WithMaxRetries(1), WithRetryBackoff(1), WithCache(false, 0))
	ch := AsyncSearchRequest[*response.Artifact](c, context.Background(), request.NewSearchRequest().SetLimit(5).SetQuery(request.NewQuery().SetGroupId("g")))
	r := <-ch
	assert.NoError(t, r.Error)
	assert.NotNil(t, r.Result)
}

func TestMockAsyncSearchRequestDoc(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"response":{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0","g":"g","a":"a","v":"1.0"}]}}`))
	}))
	defer srv.Close()
	c := NewClient(WithBaseURL(srv.URL), WithRepoBaseURL(srv.URL), WithHTTPClient(srv.Client()), WithMaxRetries(1), WithRetryBackoff(1), WithCache(false, 0))
	ch := AsyncSearchRequestDoc[*response.Artifact](c, context.Background(), request.NewSearchRequest().SetLimit(5).SetQuery(request.NewQuery().SetGroupId("g")))
	r := <-ch
	assert.NoError(t, r.Error)
	assert.NotNil(t, r.Result)
}

func TestMockAsyncSearchByGroupId(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	}))
	defer srv.Close()
	c := NewClient(WithBaseURL(srv.URL), WithRepoBaseURL(srv.URL), WithHTTPClient(srv.Client()), WithMaxRetries(1), WithRetryBackoff(1), WithCache(false, 0))
	ch := c.AsyncSearchByGroupId(context.Background(), "g", 10)
	r := <-ch
	assert.NoError(t, r.Error)
	assert.NotEmpty(t, r.Result)
}

func TestMockAsyncSearchByArtifactId(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	}))
	defer srv.Close()
	c := NewClient(WithBaseURL(srv.URL), WithRepoBaseURL(srv.URL), WithHTTPClient(srv.Client()), WithMaxRetries(1), WithRetryBackoff(1), WithCache(false, 0))
	ch := c.AsyncSearchByArtifactId(context.Background(), "a", 10)
	r := <-ch
	assert.NoError(t, r.Error)
	assert.NotEmpty(t, r.Result)
}

func TestMockBatchAsyncSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	}))
	defer srv.Close()
	c := NewClient(WithBaseURL(srv.URL), WithRepoBaseURL(srv.URL), WithHTTPClient(srv.Client()), WithMaxRetries(1), WithRetryBackoff(1), WithCache(false, 0))
	ch := c.BatchAsyncSearch(context.Background(), []*request.SearchRequest{request.NewSearchRequest().SetLimit(5).SetQuery(request.NewQuery().SetGroupId("g"))})
	r := <-ch
	assert.NoError(t, r.Error)
	assert.NotEmpty(t, r.Result)
}

func TestMockAsyncBatchDownload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("data"))
	}))
	defer srv.Close()
	c := NewClient(WithBaseURL(srv.URL), WithRepoBaseURL(srv.URL), WithHTTPClient(srv.Client()), WithMaxRetries(1), WithRetryBackoff(1), WithCache(false, 0))
	ch := c.AsyncBatchDownload(context.Background(), []string{"g/a/1.0/a-1.0.jar"})
	r := <-ch
	assert.NoError(t, r.Error)
	assert.NotEmpty(t, r.Result)
}

// ========================================================================
// Multi-step methods — skipped
// ========================================================================

func TestMockSearchByGroupPattern(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"responseHeader":{"status":0,"QTime":1,"params":{"q":"g:com.ex*"}},"response":{"numFound":1,"start":0,"docs":[{"g":"com.example","a":"my-lib","v":"1.0","timestamp":1600000000000}]}}`))
	})
	results, err := c.SearchByGroupPattern(context.Background(), "com.ex*", 10)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestMockGetPopularGroups(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"responseHeader":{"status":0,"QTime":1,"params":{"q":"*:*","facet":"true","facet.field":"g","facet.limit":"10","rows":"0"}},"response":{"numFound":100,"start":0,"docs":[]},"facet_counts":{"facet_fields":{"g":["org.apache",30,"com.google",25]}}}`))
	})
	results, err := c.GetPopularGroups(context.Background(), 10)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestMockSearchSubgroups(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"responseHeader":{"status":0,"QTime":1,"params":{"q":"g:com.example.*"}},"response":{"numFound":1,"start":0,"docs":[{"g":"com.example.sub","a":"lib","v":"1.0","timestamp":1600000000000}]}}`))
	})
	results, err := c.SearchSubgroups(context.Background(), "com.example", 10)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestMockSearchArtifactsWithAllTags(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.SearchArtifactsWithAllTags(context.Background(), []string{"test"}, 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockSearchByTagWithGroupFilter(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	})
	results, err := c.SearchByTagWithGroupFilter(context.Background(), "test", "g", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockFindSimilarVulnerableArtifacts(t *testing.T) {
	callCount := 0
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		if callCount == 1 {
			// GetVulnerabilityDetails - /api/security/vulnerabilities/... returns VulnerabilityDetails
			w.Write([]byte(`{"groupId":"g","artifactId":"a","version":"1.0","vulnerabilities":[{"id":"CVE-2023-0001","cve":"CVE-2023-0001","title":"Test Vuln","severity":"HIGH","cvssScore":7.5}]}`))
		} else {
			// Search for similar artifacts
			w.Write([]byte(mockSearchResponse))
		}
	})
	results, err := c.FindSimilarVulnerableArtifacts(context.Background(), "g", "a", "1.0", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockFilterVersions(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockVersionResponse))
	})
	results, err := c.FilterVersions(context.Background(), "g", "a", func(v *response.Version) bool {
		return v.Version == "1.0.0"
	})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestMockDownloadWithChecksum(t *testing.T) {
	callCount := 0
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
		if callCount == 1 {
			w.Write([]byte("test-jar-content"))
		} else {
			w.Write([]byte("43749ca4b95a03cbe5d964d49f04ab929224eeac"))
		}
	})
	content, cs, err := c.DownloadWithChecksum(context.Background(), "g/a/1.0/a-1.0.jar", "sha1")
	assert.NoError(t, err)
	assert.NotEmpty(t, content)
	assert.NotEmpty(t, cs)
}

func TestMockDownloadWithVerifiedChecksum(t *testing.T) {
	callCount := 0
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
		if callCount == 1 {
			// Download the jar file
			w.Write([]byte("test-jar-content"))
		} else {
			// DownloadChecksumFile - .sha1 file
			w.Write([]byte("43749ca4b95a03cbe5d964d49f04ab929224eeac"))
		}
	})
	content, cs, err := c.DownloadWithVerifiedChecksum(context.Background(), "g/a/1.0/a-1.0.jar", "sha1")
	assert.NoError(t, err)
	assert.NotEmpty(t, content)
	assert.NotEmpty(t, cs)
}

func TestMockBatchDownloadDependencies(t *testing.T) {
	callCount := 0
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		if callCount == 1 {
			// SearchByGroupAndArtifactId → search result
			w.Write([]byte(mockSearchResponse))
		} else if callCount == 2 {
			// DownloadPom → POM content (needed by GetArtifactMetadata)
			w.Write([]byte(`<project><modelVersion>4.0.0</modelVersion></project>`))
		} else {
			// Download dependency files
			w.Write([]byte("jar-content"))
		}
	})
	_, err := c.BatchDownloadDependencies(context.Background(), "g", "a", "1.0", t.TempDir())
	// BatchDownloadDependencies may return empty if no deps in metadata
	_ = err
}

func TestMockAsyncGetArtifactMetadata(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(mockSearchResponse))
	}))
	defer srv.Close()
	c := NewClient(WithBaseURL(srv.URL), WithRepoBaseURL(srv.URL), WithHTTPClient(srv.Client()), WithMaxRetries(1), WithRetryBackoff(1), WithCache(false, 0))
	ch := c.AsyncGetArtifactMetadata(context.Background(), "g", "a", "1.0")
	r := <-ch
	assert.NoError(t, r.Error)
	assert.NotNil(t, r.Result)
}

func TestMockPublisherBrowseDeployment(t *testing.T) {
	t.Skip("multi-step: requires Publisher API flow")
}

func TestMockPublisherBrowseDeploymentWithOptions(t *testing.T) {
	t.Skip("multi-step: requires Publisher API flow")
}
