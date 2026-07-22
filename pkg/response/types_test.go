package response

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Error() 方法测试
// =============================================================================

func TestHTTPError_Error(t *testing.T) {
	e := &HTTPError{
		StatusCode: 404,
		Message:    "Not Found",
		URL:        "https://central.sonatype.com/api/v1/search",
	}
	got := e.Error()
	assert.Contains(t, got, "404")
	assert.Contains(t, got, "Not Found")
	assert.Contains(t, got, "https://central.sonatype.com/api/v1/search")
	// 空值不应 panic
	empty := &HTTPError{}
	_ = empty.Error()
}

func TestAPIError_Error(t *testing.T) {
	// Happy path
	e := &APIError{
		Code:    "INVALID_REQUEST",
		Message: "the query parameter 'q' is required",
	}
	got := e.Error()
	assert.Contains(t, got, "INVALID_REQUEST")
	assert.Contains(t, got, "the query parameter 'q' is required")
	// Edge case: 空值不应 panic
	empty := &APIError{}
	_ = empty.Error()
}

func TestPublisherErrorResponse_Error(t *testing.T) {
	// Happy path
	e := &PublisherErrorResponse{
		HttpStatus: 400,
		ErrorCode:  1001,
		Message:    "deployment not found",
	}
	got := e.Error()
	assert.Contains(t, got, "400")
	assert.Contains(t, got, "1001")
	assert.Contains(t, got, "deployment not found")
	// Edge case: 空值不应 panic
	empty := &PublisherErrorResponse{}
	_ = empty.Error()
}

// =============================================================================
// SpellcheckResponse.GetSuggestions 测试（覆盖所有分支）
// =============================================================================

func TestSpellcheckResponse_GetSuggestions_NilReceiver(t *testing.T) {
	// nil receiver 分支
	var s *SpellcheckResponse
	assert.Nil(t, s.GetSuggestions())
}

func TestSpellcheckResponse_GetSuggestions_EmptySuggestions(t *testing.T) {
	// 空 Suggestions 分支
	s := &SpellcheckResponse{Suggestions: nil}
	assert.Nil(t, s.GetSuggestions())

	s2 := &SpellcheckResponse{Suggestions: []interface{}{}}
	assert.Nil(t, s2.GetSuggestions())
}

func TestSpellcheckResponse_GetSuggestions_NormalSolrFormat(t *testing.T) {
	// 正常解析 Solr 交替格式：[word, {suggestion: [...]}, ...]
	// map 元素的 "suggestion" key 是 []interface{}，提取其中的 string
	s := &SpellcheckResponse{
		Suggestions: []interface{}{
			"commns",
			map[string]interface{}{
				"suggestion": []interface{}{"commons", "communs"},
			},
		},
	}
	got := s.GetSuggestions()
	assert.Equal(t, []string{"commons", "communs"}, got)
}

func TestSpellcheckResponse_GetSuggestions_NonStringElementsSkipped(t *testing.T) {
	// 覆盖：suggestion 列表中含非 string 元素（如数字）被跳过的分支
	s := &SpellcheckResponse{
		Suggestions: []interface{}{
			map[string]interface{}{
				"suggestion": []interface{}{"commons", 123, "communs", true},
			},
		},
	}
	got := s.GetSuggestions()
	assert.Equal(t, []string{"commons", "communs"}, got)
}

func TestSpellcheckResponse_GetSuggestions_MultipleSuggestionObjects(t *testing.T) {
	// 多组 word/suggestionObject 对，验证合并逻辑
	s := &SpellcheckResponse{
		Suggestions: []interface{}{
			"commns",
			map[string]interface{}{
				"numFound":    2,
				"startOffset": 0,
				"endOffset":   6,
				"suggestion":  []interface{}{"commons"},
			},
			"guvava",
			map[string]interface{}{
				"suggestion": []interface{}{"guava"},
			},
		},
	}
	got := s.GetSuggestions()
	assert.Equal(t, []string{"commons", "guava"}, got)
}

func TestSpellcheckResponse_GetSuggestions_NonMapElementIgnored(t *testing.T) {
	// 顶层元素是 string（非 map）应被忽略，不 panic
	s := &SpellcheckResponse{
		Suggestions: []interface{}{
			"justastring", // 非 map，被忽略
		},
	}
	got := s.GetSuggestions()
	assert.Nil(t, got)
}

func TestSpellcheckResponse_GetSuggestions_MapWithoutSuggestionKey(t *testing.T) {
	// map 元素没有 "suggestion" key 的分支（类型断言失败）
	s := &SpellcheckResponse{
		Suggestions: []interface{}{
			map[string]interface{}{
				"numFound":    2,
				"startOffset": 0,
			},
		},
	}
	got := s.GetSuggestions()
	assert.Nil(t, got)
}

func TestSpellcheckResponse_GetSuggestions_SuggestionNotSlice(t *testing.T) {
	// "suggestion" 字段值非 []interface{} 的分支
	s := &SpellcheckResponse{
		Suggestions: []interface{}{
			map[string]interface{}{
				"suggestion": "commons", // string 而非 []interface{}
			},
		},
	}
	got := s.GetSuggestions()
	assert.Nil(t, got)
}

// =============================================================================
// JSON 反序列化 round-trip 测试
// =============================================================================

func TestArtifact_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"id": "org.apache:commons-io:2.11.0",
		"g": "org.apache",
		"a": "commons-io",
		"latestVersion": "2.11.0",
		"repositoryId": "central",
		"p": "jar",
		"timestamp": 1633036800000,
		"versionCount": 25,
		"text": ["commons-io", "IO", "utilities"],
		"ec": ["-sources.jar", ".jar", ".pom"],
		"tags": ["io", "utilities"]
	}`
	var a Artifact
	err := json.Unmarshal([]byte(jsonData), &a)
	assert.NoError(t, err)
	assert.Equal(t, "org.apache:commons-io:2.11.0", a.ID)
	assert.Equal(t, "org.apache", a.GroupId)
	assert.Equal(t, "commons-io", a.ArtifactId)
	assert.Equal(t, "2.11.0", a.LatestVersion)
	assert.Equal(t, "central", a.RepositoryID)
	assert.Equal(t, "jar", a.Packaging)
	assert.Equal(t, int64(1633036800000), a.Timestamp)
	assert.Equal(t, 25, a.VersionCount)
	assert.Equal(t, []string{"commons-io", "IO", "utilities"}, a.Text)
	assert.Equal(t, []string{"-sources.jar", ".jar", ".pom"}, a.Ec)
	assert.Equal(t, []string{"io", "utilities"}, a.Tags)
}

func TestVersion_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"id": "org.apache:commons-io:2.11.0",
		"g": "org.apache",
		"a": "commons-io",
		"v": "2.11.0",
		"p": "jar",
		"timestamp": 1633036800000,
		"ec": ["-sources.jar", ".jar"],
		"tags": ["io"]
	}`
	var v Version
	err := json.Unmarshal([]byte(jsonData), &v)
	assert.NoError(t, err)
	assert.Equal(t, "org.apache:commons-io:2.11.0", v.ID)
	assert.Equal(t, "org.apache", v.GroupId)
	assert.Equal(t, "commons-io", v.ArtifactId)
	assert.Equal(t, "2.11.0", v.Version)
	assert.Equal(t, "jar", v.Packaging)
	assert.Equal(t, int64(1633036800000), v.Timestamp)
	assert.Equal(t, []string{"-sources.jar", ".jar"}, v.Ec)
	assert.Equal(t, []string{"io"}, v.Tags)
}

func TestArtifactMetadata_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"groupId": "org.apache",
		"artifactId": "commons-io",
		"latestVersion": "2.11.0",
		"packaging": "jar",
		"lastUpdated": 1633036800000,
		"pomContent": "<project></project>",
		"dependencies": [
			{"groupId": "junit", "artifactId": "junit", "version": "4.13.2", "scope": "test"}
		],
		"securityRating": {"score": 8.5, "vulnCount": 1, "severity": "LOW"},
		"licenses": ["Apache-2.0"],
		"developers": [{"name": "dev", "email": "dev@apache.org"}],
		"projectInfo": {"name": "Commons IO", "url": "https://commons.apache.org/io"}
	}`
	var m ArtifactMetadata
	err := json.Unmarshal([]byte(jsonData), &m)
	assert.NoError(t, err)
	assert.Equal(t, "org.apache", m.GroupId)
	assert.Equal(t, "commons-io", m.ArtifactId)
	assert.Equal(t, "2.11.0", m.LatestVersion)
	assert.Equal(t, "jar", m.Packaging)
	assert.Equal(t, int64(1633036800000), m.LastUpdated)
	assert.Equal(t, "<project></project>", m.PomContent)
	assert.Len(t, m.Dependencies, 1)
	assert.Equal(t, "junit", m.Dependencies[0].GroupId)
	assert.Equal(t, "4.13.2", m.Dependencies[0].Version)
	assert.Equal(t, "test", m.Dependencies[0].Scope)
	assert.NotNil(t, m.SecurityRating)
	assert.Equal(t, 8.5, m.SecurityRating.Score)
	assert.Equal(t, 1, m.SecurityRating.VulnCount)
	assert.Equal(t, "LOW", m.SecurityRating.Severity)
	assert.Equal(t, []string{"Apache-2.0"}, m.Licenses)
	assert.Len(t, m.Developers, 1)
	assert.Equal(t, "dev", m.Developers[0].Name)
	assert.NotNil(t, m.ProjectInfo)
	assert.Equal(t, "Commons IO", m.ProjectInfo.Name)
}

func TestDeploymentStatus_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"deploymentId": "abc-123",
		"deploymentName": "my-deployment",
		"deploymentState": "PUBLISHED",
		"publishingType": "AUTOMATIC",
		"purls": ["pkg:maven/org.apache/commons-io@2.11.0"],
		"errors": null,
		"createTimestamp": "2023-01-01T00:00:00Z",
		"updateTimestamp": "2023-01-02T00:00:00Z"
	}`
	var d DeploymentStatus
	err := json.Unmarshal([]byte(jsonData), &d)
	assert.NoError(t, err)
	assert.Equal(t, "abc-123", d.DeploymentID)
	assert.Equal(t, "my-deployment", d.DeploymentName)
	assert.Equal(t, DeploymentStatePublished, d.DeploymentState)
	assert.Equal(t, PublishingTypeAutomatic, d.PublishingType)
	assert.Equal(t, []string{"pkg:maven/org.apache/commons-io@2.11.0"}, d.Purls)
	assert.Equal(t, "2023-01-01T00:00:00Z", d.CreateTimestamp)
	assert.Equal(t, "2023-01-02T00:00:00Z", d.UpdateTimestamp)
}

func TestPublishedCheck_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"published": true,
		"namespace": "org.apache",
		"name": "commons-io",
		"version": "2.11.0"
	}`
	var p PublishedCheck
	err := json.Unmarshal([]byte(jsonData), &p)
	assert.NoError(t, err)
	assert.True(t, p.Published)
	assert.Equal(t, "org.apache", p.Namespace)
	assert.Equal(t, "commons-io", p.Name)
	assert.Equal(t, "2.11.0", p.Version)
}

func TestVulnerability_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"id": "VULN-001",
		"cve": "CVE-2021-44228",
		"title": "Log4Shell",
		"description": "Remote code execution",
		"severity": "CRITICAL",
		"cvssScore": 10.0,
		"cvssVector": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H",
		"advisoryLink": "https://nvd.nist.gov/vuln/detail/CVE-2021-44228"
	}`
	var v Vulnerability
	err := json.Unmarshal([]byte(jsonData), &v)
	assert.NoError(t, err)
	assert.Equal(t, "VULN-001", v.ID)
	assert.Equal(t, "CVE-2021-44228", v.CVE)
	assert.Equal(t, "Log4Shell", v.Title)
	assert.Equal(t, "Remote code execution", v.Description)
	assert.Equal(t, "CRITICAL", v.Severity)
	assert.Equal(t, 10.0, v.CvssScore)
	assert.Equal(t, "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H", v.CvssVector)
	assert.Equal(t, "https://nvd.nist.gov/vuln/detail/CVE-2021-44228", v.Advisory)
}

func TestGroupInfo_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"groupId": "org.apache",
		"artifactCount": 42,
		"lastUpdated": 1633036800000,
		"lastUpdatedDate": "2021-10-01",
		"description": "Apache Commons",
		"website": "https://commons.apache.org"
	}`
	var g GroupInfo
	err := json.Unmarshal([]byte(jsonData), &g)
	assert.NoError(t, err)
	assert.Equal(t, "org.apache", g.GroupId)
	assert.Equal(t, 42, g.ArtifactCount)
	assert.Equal(t, int64(1633036800000), g.LastUpdated)
	assert.Equal(t, "2021-10-01", g.LastUpdatedDate)
	assert.Equal(t, "Apache Commons", g.Description)
	assert.Equal(t, "https://commons.apache.org", g.Website)
}

func TestTagCount_JSONRoundTrip(t *testing.T) {
	jsonData := `{"tag": "io", "count": 123}`
	var tc TagCount
	err := json.Unmarshal([]byte(jsonData), &tc)
	assert.NoError(t, err)
	assert.Equal(t, "io", tc.Tag)
	assert.Equal(t, 123, tc.Count)
}

func TestVersionWithMetadata_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"version": {
			"id": "org.apache:commons-io:2.11.0",
			"g": "org.apache",
			"a": "commons-io",
			"v": "2.11.0",
			"p": "jar",
			"timestamp": 1633036800000,
			"ec": [".jar"],
			"tags": ["io"]
		},
		"versionInfo": {
			"groupId": "org.apache",
			"artifactId": "commons-io",
			"version": "2.11.0",
			"lastUpdated": "2021-10-01",
			"packaging": "jar"
		}
	}`
	var vwm VersionWithMetadata
	err := json.Unmarshal([]byte(jsonData), &vwm)
	assert.NoError(t, err)
	assert.NotNil(t, vwm.Version)
	assert.Equal(t, "org.apache:commons-io:2.11.0", vwm.Version.ID)
	assert.Equal(t, "2.11.0", vwm.Version.Version)
	assert.NotNil(t, vwm.VersionInfo)
	assert.Equal(t, "org.apache", vwm.VersionInfo.GroupId)
	assert.Equal(t, "commons-io", vwm.VersionInfo.ArtifactId)
	assert.Equal(t, "2.11.0", vwm.VersionInfo.Version)
	assert.Equal(t, "2021-10-01", vwm.VersionInfo.LastUpdated)
	assert.Equal(t, "jar", vwm.VersionInfo.Packaging)
}

func TestLicenseInfo_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"name": "Apache-2.0",
		"type": "open source",
		"category": "permissive",
		"url": "https://www.apache.org/licenses/LICENSE-2.0",
		"description": "Apache License 2.0"
	}`
	var l LicenseInfo
	err := json.Unmarshal([]byte(jsonData), &l)
	assert.NoError(t, err)
	assert.Equal(t, "Apache-2.0", l.Name)
	assert.Equal(t, "open source", l.Type)
	assert.Equal(t, "permissive", l.Category)
	assert.Equal(t, "https://www.apache.org/licenses/LICENSE-2.0", l.URL)
	assert.Equal(t, "Apache License 2.0", l.Description)
}

func TestDeploymentList_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"deployments": [
			{
				"deploymentId": "dep-1",
				"deploymentName": "first",
				"namespace": "org.apache",
				"deploymentState": "PUBLISHED",
				"createTimestamp": "2023-01-01T00:00:00Z",
				"updateTimestamp": "2023-01-02T00:00:00Z"
			}
		],
		"page": 0,
		"pageSize": 10,
		"pageCount": 1,
		"totalResultCount": 1
	}`
	var dl DeploymentList
	err := json.Unmarshal([]byte(jsonData), &dl)
	assert.NoError(t, err)
	assert.Len(t, dl.Deployments, 1)
	assert.Equal(t, "dep-1", dl.Deployments[0].DeploymentID)
	assert.Equal(t, "first", dl.Deployments[0].DeploymentName)
	assert.Equal(t, "org.apache", dl.Deployments[0].Namespace)
	assert.Equal(t, DeploymentStatePublished, dl.Deployments[0].DeploymentState)
	assert.Equal(t, "2023-01-01T00:00:00Z", dl.Deployments[0].CreateTimestamp)
	assert.Equal(t, "2023-01-02T00:00:00Z", dl.Deployments[0].UpdateTimestamp)
	assert.Equal(t, 0, dl.Page)
	assert.Equal(t, 10, dl.PageSize)
	assert.Equal(t, 1, dl.PageCount)
	assert.Equal(t, 1, dl.TotalResultCount)
}

func TestPublisherError_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"code": 500,
		"message": "internal error",
		"details": "stack trace"
	}`
	var pe PublisherError
	err := json.Unmarshal([]byte(jsonData), &pe)
	assert.NoError(t, err)
	assert.Equal(t, 500, pe.Code)
	assert.Equal(t, "internal error", pe.Message)
	assert.Equal(t, "stack trace", pe.Details)
}

func TestBrowseDeploymentRequest_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"page": 0,
		"size": 50,
		"sortField": "name",
		"sortDirection": "asc",
		"deploymentIds": ["dep-1", "dep-2"],
		"pathStarting": "org/apache"
	}`
	var b BrowseDeploymentRequest
	err := json.Unmarshal([]byte(jsonData), &b)
	assert.NoError(t, err)
	assert.Equal(t, 0, b.Page)
	assert.Equal(t, 50, b.Size)
	assert.Equal(t, "name", b.SortField)
	assert.Equal(t, "asc", b.SortDirection)
	assert.Equal(t, []string{"dep-1", "dep-2"}, b.DeploymentIds)
	assert.Equal(t, "org/apache", b.PathStarting)
}

func TestAPIError_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"code": "BAD_REQUEST",
		"message": "invalid query"
	}`
	var ae APIError
	err := json.Unmarshal([]byte(jsonData), &ae)
	assert.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", ae.Code)
	assert.Equal(t, "invalid query", ae.Message)
}

func TestHTTPError_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"statusCode": 500,
		"message": "internal server error",
		"url": "https://central.sonatype.com/api"
	}`
	var he HTTPError
	err := json.Unmarshal([]byte(jsonData), &he)
	assert.NoError(t, err)
	assert.Equal(t, 500, he.StatusCode)
	assert.Equal(t, "internal server error", he.Message)
	assert.Equal(t, "https://central.sonatype.com/api", he.URL)
}

// =============================================================================
// 嵌套 Response[Artifact] 反序列化测试
// =============================================================================

func TestResponseGenericArtifact_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"responseHeader": {"status": 0, "QTime": 5},
		"response": {
			"numFound": 1,
			"start": 0,
			"docs": [
				{
					"id": "org.apache:commons-io:2.11.0",
					"g": "org.apache",
					"a": "commons-io",
					"latestVersion": "2.11.0",
					"repositoryId": "central",
					"p": "jar",
					"timestamp": 1633036800000,
					"versionCount": 25,
					"text": ["commons-io"],
					"ec": [".jar"]
				}
			]
		},
		"facet_counts": {
			"facet_queries": {"org.apache": 30}
		},
		"spellcheck": {
			"suggestions": ["commns", {"suggestion": ["commons"]}]
		}
	}`
	var resp Response[Artifact]
	err := json.Unmarshal([]byte(jsonData), &resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp.ResponseHeader)
	assert.Equal(t, 0, resp.ResponseHeader.Status)
	assert.Equal(t, 5, resp.ResponseHeader.QTime)
	assert.NotNil(t, resp.ResponseBody)
	assert.Equal(t, 1, resp.ResponseBody.NumFound)
	assert.Equal(t, 0, resp.ResponseBody.Start)
	assert.Len(t, resp.ResponseBody.Docs, 1)
	assert.Equal(t, "org.apache:commons-io:2.11.0", resp.ResponseBody.Docs[0].ID)
	assert.Equal(t, "org.apache", resp.ResponseBody.Docs[0].GroupId)
	assert.NotNil(t, resp.FacetCounts)
	assert.Equal(t, 30, resp.FacetCounts.FacetQueries["org.apache"])
	assert.NotNil(t, resp.Spellcheck)
	assert.Equal(t, []string{"commons"}, resp.Spellcheck.GetSuggestions())
}

// =============================================================================
// PublisherErrorResponse JSON round-trip（含 Explanation/Data）
// =============================================================================

func TestPublisherErrorResponse_JSONRoundTrip(t *testing.T) {
	jsonData := `{
		"httpStatus": 400,
		"errorCode": 1001,
		"message": "invalid deployment",
		"explanation": "deploymentId is required",
		"data": {"field": "deploymentId"}
	}`
	var e PublisherErrorResponse
	err := json.Unmarshal([]byte(jsonData), &e)
	assert.NoError(t, err)
	assert.Equal(t, 400, e.HttpStatus)
	assert.Equal(t, 1001, e.ErrorCode)
	assert.Equal(t, "invalid deployment", e.Message)
	assert.Equal(t, "deploymentId is required", e.Explanation)
	assert.NotNil(t, e.Data)
}
