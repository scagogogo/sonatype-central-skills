package api

import (
	"testing"
	"time"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
	"github.com/stretchr/testify/assert"
)

// ---- isGPL ----

func TestIsGPL(t *testing.T) {
	assert.True(t, isGPL("GPL-2.0"))
	assert.True(t, isGPL("GPL-3.0"))
	assert.False(t, isGPL("MIT"))
	assert.False(t, isGPL("Apache-2.0"))
	assert.False(t, isGPL("LGPL-2.0"))
	assert.False(t, isGPL(""))
	assert.False(t, isGPL("BSD-2-Clause"))
}

// ---- isPermissiveLicense ----

func TestIsPermissiveLicense(t *testing.T) {
	assert.True(t, isPermissiveLicense("MIT"))
	assert.True(t, isPermissiveLicense("BSD-2-Clause"))
	assert.True(t, isPermissiveLicense("BSD-3-Clause"))
	assert.True(t, isPermissiveLicense("Apache-2.0"))
	assert.True(t, isPermissiveLicense("Unlicense"))
	assert.False(t, isPermissiveLicense("GPL-2.0"))
	assert.False(t, isPermissiveLicense("GPL-3.0"))
	assert.False(t, isPermissiveLicense("LGPL-2.0"))
	assert.False(t, isPermissiveLicense(""))
}

// ---- isCompatibleWithGPL ----

func TestIsCompatibleWithGPL(t *testing.T) {
	assert.True(t, isCompatibleWithGPL("MIT"))
	assert.True(t, isCompatibleWithGPL("BSD-2-Clause"))
	assert.True(t, isCompatibleWithGPL("BSD-3-Clause"))
	assert.True(t, isCompatibleWithGPL("LGPL-2.0"))
	assert.True(t, isCompatibleWithGPL("LGPL-3.0"))
	assert.True(t, isCompatibleWithGPL("Unlicense"))
	assert.False(t, isCompatibleWithGPL("Apache-2.0"))
	assert.False(t, isCompatibleWithGPL("GPL-2.0"))
	assert.False(t, isCompatibleWithGPL(""))
}

// ---- determineLicenseCategory ----

func TestDetermineLicenseCategory(t *testing.T) {
	assert.Equal(t, LicenseCategoryCopyleft, determineLicenseCategory(LicenseTypeGPLv2))
	assert.Equal(t, LicenseCategoryCopyleft, determineLicenseCategory(LicenseTypeGPLv3))
	// LGPL-2.0 contains "GPL" so strings.Contains("GPL") matches first → copyleft, not weak-copyleft
	assert.Equal(t, LicenseCategoryCopyleft, determineLicenseCategory(LicenseTypeLGPLv2))
	assert.Equal(t, LicenseCategoryCopyleft, determineLicenseCategory(LicenseTypeLGPLv3))
	assert.Equal(t, LicenseCategoryPermissive, determineLicenseCategory(LicenseTypeMIT))
	assert.Equal(t, LicenseCategoryPermissive, determineLicenseCategory(LicenseTypeApache2))
	assert.Equal(t, LicenseCategoryPermissive, determineLicenseCategory(LicenseTypeBSD2))
	assert.Equal(t, LicenseCategoryPermissive, determineLicenseCategory(LicenseTypeBSD3))
	// Unknown license → default permissive
	assert.Equal(t, LicenseCategoryPermissive, determineLicenseCategory("UNKNOWN"))
}

// ---- parseLicense ----

func TestParseLicense(t *testing.T) {
	li := parseLicense("MIT")
	assert.Equal(t, "MIT", li.Name)
	assert.Equal(t, "MIT", li.Type)
	assert.Equal(t, string(LicenseCategoryPermissive), li.Category)
	assert.Contains(t, li.URL, "opensource.org/licenses/")

	li2 := parseLicense("GPL-3.0")
	assert.Equal(t, "GPL-3.0", li2.Name)
	assert.Equal(t, string(LicenseCategoryCopyleft), li2.Category)

	li3 := parseLicense("")
	assert.Equal(t, "", li3.Name)
}

// ---- CheckLicenseCompatibility ----

func TestCheckLicenseCompatibilityPairs(t *testing.T) {
	c := &Client{}

	// Same license → compatible
	ok, reason, err := c.CheckLicenseCompatibility("MIT", "MIT")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Contains(t, reason, "相同")

	// GPL-2.0 vs Apache-2.0 → incompatible
	ok, reason, err = c.CheckLicenseCompatibility("GPL-2.0", "Apache-2.0")
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Contains(t, reason, "不兼容")

	// GPL-3.0 vs CDDL-1.0 → incompatible
	ok, reason, err = c.CheckLicenseCompatibility("GPL-3.0", "CDDL-1.0")
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Contains(t, reason, "不兼容")

	// GPL-3.0 vs MIT (MIT is compatible with GPL)
	ok, _, err = c.CheckLicenseCompatibility("GPL-3.0", "MIT")
	assert.NoError(t, err)
	assert.True(t, ok)

	// MIT vs BSD-3 → permissive
	ok, _, err = c.CheckLicenseCompatibility("MIT", "BSD-3-Clause")
	assert.NoError(t, err)
	assert.True(t, ok)

	// GPL-2.0 vs Unlicense (Unlicense is compatible with GPL)
	ok, _, err = c.CheckLicenseCompatibility("GPL-2.0", "Unlicense")
	assert.NoError(t, err)
	assert.True(t, ok)

	// Apache-2.0 vs GPL-2.0 → reverse order, still incompatible
	ok, reason, err = c.CheckLicenseCompatibility("Apache-2.0", "GPL-2.0")
	assert.NoError(t, err)
	assert.False(t, ok)
}

// ---- BuildArtifactPath ----

func TestBuildArtifactPath(t *testing.T) {
	p := BuildArtifactPath("com.example", "my-lib", "1.0.0", "jar")
	assert.Equal(t, "com/example/my-lib/1.0.0/my-lib-1.0.0.jar", p)

	p2 := BuildArtifactPath("com.example", "my-lib", "1.0.0", "jar", "sources")
	assert.Contains(t, p2, "-sources.jar")

	p3 := BuildArtifactPath("", "", "", "")
	assert.NotEmpty(t, p3)
}

// ---- CommonArtifactFiles ----

func TestCommonArtifactFiles(t *testing.T) {
	files := CommonArtifactFiles()
	assert.NotEmpty(t, files)
}

// ---- contains ----

func TestContains(t *testing.T) {
	assert.True(t, contains([]string{"a", "b", "c"}, "a"))
	assert.True(t, contains([]string{"a", "b", "c"}, "c"))
	assert.False(t, contains([]string{"a", "b", "c"}, "d"))
	assert.False(t, contains([]string{}, "a"))
	assert.False(t, contains(nil, "a"))
}

// ---- hasExtension ----

func TestHasExtensionLocal(t *testing.T) {
	assert.True(t, hasExtension("file.jar"))
	assert.True(t, hasExtension("file.txt"))
	assert.False(t, hasExtension("file"))
	assert.False(t, hasExtension(""))
	assert.False(t, hasExtension("/path/to/file"))
}

// ---- minInt ----

func TestMinInt(t *testing.T) {
	assert.Equal(t, 1, minInt(1, 2))
	assert.Equal(t, 1, minInt(2, 1))
	assert.Equal(t, 5, minInt(5, 5))
}

// ---- isRetriableError ----

func TestIsRetriableError(t *testing.T) {
	assert.True(t, isRetriableError(429))
	assert.True(t, isRetriableError(500))
	assert.True(t, isRetriableError(502))
	assert.True(t, isRetriableError(503))
	assert.True(t, isRetriableError(504))
	assert.False(t, isRetriableError(200))
	assert.False(t, isRetriableError(400))
	assert.False(t, isRetriableError(401))
	assert.False(t, isRetriableError(403))
	assert.False(t, isRetriableError(404))
}

// ---- isRetriableStatusCode ----

func TestIsRetriableStatusCode(t *testing.T) {
	assert.True(t, isRetriableStatusCode(429))
	assert.True(t, isRetriableStatusCode(500))
	assert.True(t, isRetriableStatusCode(502))
	assert.True(t, isRetriableStatusCode(503))
	assert.True(t, isRetriableStatusCode(504))
	assert.False(t, isRetriableStatusCode(200))
	assert.False(t, isRetriableStatusCode(404))
}

// ---- shouldRetryError ----

func TestShouldRetryError(t *testing.T) {
	// nil → false
	assert.False(t, shouldRetryError(nil))

	// HTTPError with retriable status → true
	assert.True(t, shouldRetryError(&response.HTTPError{StatusCode: 429}))
	assert.True(t, shouldRetryError(&response.HTTPError{StatusCode: 503}))
	assert.True(t, shouldRetryError(&response.HTTPError{StatusCode: 500}))

	// HTTPError with non-retriable status → false
	assert.False(t, shouldRetryError(&response.HTTPError{StatusCode: 400}))
	assert.False(t, shouldRetryError(&response.HTTPError{StatusCode: 404}))

	// Generic error → false
	assert.False(t, shouldRetryError(assert.AnError))
}

// ---- handleHttpError ----

func TestHandleHttpError429(t *testing.T) {
	err := handleHttpError(429, []byte(`{"message":"too many requests"}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "限流")
}

func TestHandleHttpError404(t *testing.T) {
	err := handleHttpError(404, []byte(`{"message":"not found"}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不存在")
}

func TestHandleHttpError401(t *testing.T) {
	err := handleHttpError(401, []byte(`{"message":"unauthorized"}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未授权")
}

func TestHandleHttpError403(t *testing.T) {
	err := handleHttpError(403, []byte(`{"message":"forbidden"}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "禁止")
}

func TestHandleHttpError400(t *testing.T) {
	err := handleHttpError(400, []byte(`{"message":"bad request"}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "参数错误")
}

func TestHandleHttpError500(t *testing.T) {
	err := handleHttpError(500, []byte(`{"message":"server error"}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "服务器错误")
}

func TestHandleHttpErrorOther(t *testing.T) {
	// 418 → unknown status code
	err := handleHttpError(418, []byte(`{"error":"teapot"}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "418")
}

func TestHandleHttpErrorNoBody(t *testing.T) {
	// No body, use StatusText
	err := handleHttpError(404, []byte{})
	assert.Error(t, err)
}

func TestHandleHttpErrorInvalidJSON(t *testing.T) {
	// Invalid JSON body → fallback to StatusText
	err := handleHttpError(500, []byte(`not json`))
	assert.Error(t, err)
}

// ---- ExtractHighlightedClasses ----

func TestExtractHighlightedClassesLocalNil(t *testing.T) {
	assert.Nil(t, ExtractHighlightedClasses(nil))
}

func TestExtractHighlightedClassesLocalEmptyHighlighting(t *testing.T) {
	// Response with nil Highlighting
	r := &response.Response[*response.Version]{}
	assert.Nil(t, ExtractHighlightedClasses(r))
}

func TestExtractHighlightedClassesLocalWithData(t *testing.T) {
	r := &response.Response[*response.Version]{
		Highlighting: map[string]map[string][]string{
			"g:a:1.0": {"fch": {"<em>com.example</em>.Foo"}},
		},
	}
	m := ExtractHighlightedClasses(r)
	assert.NotNil(t, m)
	assert.Contains(t, m, "g:a:1.0")
	assert.Equal(t, []string{"<em>com.example</em>.Foo"}, m["g:a:1.0"])
}

// ---- findConflicts ----

func TestFindConflictsEmpty(t *testing.T) {
	conflicts := findConflicts(map[response.ArtifactRef][]response.LicenseInfo{})
	assert.Empty(t, conflicts)
}

func TestFindConflictsSameLicense(t *testing.T) {
	// Same license type → no conflict
	ref := response.ArtifactRef{GroupId: "g", ArtifactId: "a", Version: "1.0"}
	licenses := map[response.ArtifactRef][]response.LicenseInfo{
		ref: {{Name: "MIT", Type: "MIT"}},
	}
	conflicts := findConflicts(licenses)
	assert.Empty(t, conflicts)
}

func TestFindConflictsGPLv2Apache2(t *testing.T) {
	// GPL-2.0 vs Apache-2.0 → conflict
	ref1 := response.ArtifactRef{GroupId: "g1", ArtifactId: "a1", Version: "1.0"}
	ref2 := response.ArtifactRef{GroupId: "g2", ArtifactId: "a2", Version: "1.0"}
	licenses := map[response.ArtifactRef][]response.LicenseInfo{
		ref1: {{Name: "GPL-2.0", Type: "GPL-2.0"}},
		ref2: {{Name: "Apache-2.0", Type: "Apache-2.0"}},
	}
	conflicts := findConflicts(licenses)
	assert.NotEmpty(t, conflicts)
}

// ---- generateRecommendations ----

func TestGenerateRecommendationsNoConflicts(t *testing.T) {
	summary := &response.LicenseSummary{
		PotentialConflicts: []response.LicenseConflict{},
		LicenseDistribution: map[string]int{},
	}
	recs := generateRecommendations(summary)
	assert.NotEmpty(t, recs)
	assert.Contains(t, recs[0], "未发现")
}

func TestGenerateRecommendationsSomeConflicts(t *testing.T) {
	summary := &response.LicenseSummary{
		PotentialConflicts: []response.LicenseConflict{
			{License1: "GPL-2.0", License2: "Apache-2.0", Reason: "test"},
		},
		LicenseDistribution: map[string]int{"GPL-2.0": 1},
	}
	recs := generateRecommendations(summary)
	assert.Contains(t, recs[0], "一些")
}

func TestGenerateRecommendationsManyConflicts(t *testing.T) {
	conflicts := make([]response.LicenseConflict, 6)
	for i := 0; i < 6; i++ {
		conflicts[i] = response.LicenseConflict{License1: "A", License2: "B", Reason: "r"}
	}
	summary := &response.LicenseSummary{
		PotentialConflicts:   conflicts,
		LicenseDistribution:  map[string]int{"GPL-2.0": 1, "LGPL-2.0": 1},
	}
	recs := generateRecommendations(summary)
	assert.Contains(t, recs[0], "大量")
	assert.Contains(t, recs[1], "GPL")
	assert.Contains(t, recs[2], "LGPL")
}

// ---- Client cache methods ----

func TestClientCacheDefaults(t *testing.T) {
	c := NewClient()
	assert.False(t, c.IsCacheEnabled())
	assert.Equal(t, 300, c.GetCacheTTL())
}

func TestClientCacheEnableDisable(t *testing.T) {
	c := NewClient()
	c.EnableCache()
	assert.True(t, c.IsCacheEnabled())
	c.DisableCache()
	assert.False(t, c.IsCacheEnabled())
}

func TestClientCacheSetTTL(t *testing.T) {
	c := NewClient()
	c.SetCacheTTL(600)
	assert.Equal(t, 600, c.GetCacheTTL())
}

func TestClientCacheClear(t *testing.T) {
	c := NewClient()
	// Clear should not panic
	c.ClearCache()
}

// ---- addToCache / getFromCache ----

func TestCacheGetSet(t *testing.T) {
	addToCache("test-key", []byte("test-data"), 60)
	data, ok := getFromCache("test-key")
	assert.True(t, ok)
	assert.Equal(t, "test-data", string(data))
}

func TestCacheMiss(t *testing.T) {
	_, ok := getFromCache("nonexistent-" + time.Now().String())
	assert.False(t, ok)
}

func TestCacheTTLZero(t *testing.T) {
	addToCache("zero-ttl-key", []byte("data"), 0)
	_, ok := getFromCache("zero-ttl-key")
	assert.False(t, ok)
}

