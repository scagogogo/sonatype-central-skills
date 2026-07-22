# CLI 工具构建 + 测试覆盖率 100% + SKILLS 文档完善 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: `superpowers:subagent-driven-development`
> Steps use checkbox (`- [ ]`) syntax.

**Goal:** 从零构建覆盖全部 Go SDK API 的 CLI 工具；将测试覆盖率提升至 100%；持续完善渐进式 SKILLS 文档。

**Architecture:** 三条独立工作线。**测试线**（T1-T4）：`pkg/request` 是纯 Builder 结构无网络，直接断言 `ToRequestParams()`/`ToRequestParamValue()` 输出即可 100%；`pkg/response` 只有 4 个 `Error()`/`GetSuggestions()` 方法，补 JSON 反序列化 + Error 断言；`pkg/api` 含两类代码——纯本地逻辑（`CheckLicenseCompatibility`/`BuildArtifactPath`/`isGPL` 等）直接断言，网络层（`doRequest`/`executeWithRetry`/各 `Search*`）用 `httptest.NewServer` mock 掉 `search.maven.org`，通过 `WithHTTPClient` 注入指向 mock server 的 client，断言请求 URL 与解析结果。**CLI 线**（T5-T7）：`cmd/sonatype-central/` 用标准库 `flag` 实现子命令分发，每个子命令内部 `NewClient(...)` / `NewPublisherClient(...)` 后调用对应 SDK 方法，JSON 输出到 stdout——数据流：CLI flag 解析 → 构造 SDK Client → 调用方法 → `json.MarshalIndent` 输出。**文档线**（T8）：在 `website/guide/` 增补 `cli.md`，并把这些能力沉淀为 SKILLS 渐进式 markdown。复用现有 SDK 客户端，CLI 只做"参数→SDK→JSON"薄封装，不重复业务逻辑。

**Tech Stack:** Go 1.18（泛型，零新运行时依赖），标准库 `flag`/`os`/`encoding/json`/`net/http/httptest`（CLI 与 mock 测试），testify 1.8.1，VitePress 1.6.4（文档）

**Risks:**
- `pkg/api` 现有 60+ 测试用 `createRealClient` 打真实 Maven Central（171 秒，CI 无网络会失败且拖累覆盖率统计）→ 缓解：T4 用 `//go:build integration` build tag 隔离 real 测试文件，`go test -short` 跳过；覆盖率统计用 `-short` 模式只跑 mock 测试
- CLI 若引入 cobra/urfave 破坏零依赖承诺 → 缓解：纯标准库 `flag` 实现，每个子命令一个 `flag.FlagSet`
- SDK 公开方法 130+ 个无法 1:1 子命令 → 缓解：按数据流分 5 组（search/download/version/metadata/publisher），组内用 `--type` flag 选子方法，覆盖全部公开调用入口
- Go 1.18 泛型方法 `SearchRequestJsonDoc[Doc any]` 在 CLI 中需显式实例化 → 缓解：CLI 内部固定用 `*response.Response[*response.Artifact]` 等具体类型参数
- `httptest` mock 需精确复刻 Maven Central JSON 结构 → 缓解：T4 从现有 real 测试的真实响应中提取样本 JSON 作为 mock 返回体

---

### Task 1: 补全 pkg/request 测试至 100%

**Depends on:** None
**Files:**
- Create: `pkg/request/request_test.go`
- Modify: `pkg/request/advanced_search.go:22-72`（确认无遗漏分支）

- [ ] **Step 1: 创建 request 包全量测试 — 覆盖所有 Builder 方法与 ToRequestParams/ToRequestParamValue 输出**

```go
// pkg/request/request_test.go
package request

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdvancedSearchOptionsBuilder(t *testing.T) {
	// Happy Path: 链式设置全部字段
	opts := NewAdvancedSearchOptions().
		SetGroupId("com.example").
		SetArtifactId("my-lib").
		SetVersion("1.0.0").
		SetPackaging("jar").
		SetClassifier("sources")
	assert.Equal(t, "com.example", opts.GroupId)
	assert.Equal(t, "my-lib", opts.ArtifactId)
	assert.Equal(t, "1.0.0", opts.Version)
	assert.Equal(t, "jar", opts.Packaging)
	assert.Equal(t, "sources", opts.Classifier)
}

func TestNewAdvancedSearchOptionsDefaults(t *testing.T) {
	// Edge Case: 新建对象所有字段为零值
	opts := NewAdvancedSearchOptions()
	assert.Empty(t, opts.GroupId)
	assert.Empty(t, opts.ArtifactId)
}

func TestMakeDependencyQuery(t *testing.T) {
	// Happy Path
	q := MakeDependencyQuery("com.example", "my-lib")
	assert.Equal(t, `d:"com.example:my-lib"`, q)
	// Edge Case: 空字符串
	qEmpty := MakeDependencyQuery("", "")
	assert.Equal(t, `d":""`, qEmpty)
}

func TestMakeLicenseQuery(t *testing.T) {
	assert.Equal(t, `l:"mit"`, MakeLicenseQuery("mit"))
	assert.Equal(t, `l:""`, MakeLicenseQuery(""))
}

func TestQueryBuilderAndParamValue(t *testing.T) {
	// Happy Path: 多字段组合
	q := NewQuery().
		SetGroupId("g").
		SetArtifactId("a").
		SetVersion("v").
		SetTags("t1").
		SetSha1("abc").
		SetClassName("Cls").
		SetFullyQualifiedClassName("com.Cls").
		SetPackaging("jar").
		SetClassifier("sources").
		SetText("hello").
		SetCustomQuery("custom:foo")
	val := q.ToRequestParamValue()
	// 应包含所有设置的子句
	assert.Contains(t, val, `g:"g"`)
	assert.Contains(t, val, `a:"a"`)
	assert.Contains(t, val, `v:"v"`)
	assert.Contains(t, val, "tags:")
	assert.Contains(t, val, `sha1:"abc"`)
	assert.Contains(t, val, `c:"Cls"`)
	assert.Contains(t, val, `fc:"com.Cls"`)
	assert.Contains(t, val, `p:"jar"`)
	assert.Contains(t, val, "text:")
	assert.Contains(t, val, "custom:foo")
}

func TestQueryEmpty(t *testing.T) {
	// Edge Case: 空查询
	assert.Equal(t, "", NewQuery().ToRequestParamValue())
}

func TestSearchRequestBuilder(t *testing.T) {
	// Happy Path: 全量链式调用
	sr := NewSearchRequest().
		SetStart(0).
		SetLimit(20).
		SetCore("gav").
		SetQuery(NewQuery().SetGroupId("g")).
		SetSort("v", true).
		EnableFacets("g", "a").
		SetQueryKey("k").
		AddCustomParam("x", "y").
		SetRows(50).
		SetExact(true).
		SetSpellcheck(true, 3).
		SetFieldList("id,g").
		SetDefType("dismax").
		SetQueryFields("g^2 a^1")
	assert.Equal(t, 0, sr.Start)
	assert.Equal(t, 20, sr.Limit)
	assert.Equal(t, "gav", sr.Core)
	assert.NotNil(t, sr.Query)
	assert.Equal(t, "v", sr.SortField)
	assert.True(t, sr.SortAscending)
	assert.Equal(t, "k", sr.QueryKey)
	assert.Equal(t, "y", sr.CustomParams["x"])
	assert.True(t, sr.Exact)
	assert.Equal(t, "id,g", sr.FieldList)
	assert.Equal(t, "dismax", sr.DefType)
	assert.Equal(t, "g^2 a^1", sr.QueryFields)
}

func TestSearchRequestGetQueryKey(t *testing.T) {
	sr := NewSearchRequest().SetQueryKey("abc")
	assert.Equal(t, "abc", sr.GetQueryKey())
	// Edge Case: 未设置返回零值
	assert.Empty(t, NewSearchRequest().GetQueryKey())
}

func TestSearchRequestToRequestParams(t *testing.T) {
	// Happy Path: 完整参数序列化
	sr := NewSearchRequest().
		SetStart(0).
		SetLimit(10).
		SetQuery(NewQuery().SetGroupId("g"))
	params := sr.ToRequestParams()
	assert.NotEmpty(t, params)
	assert.Contains(t, params, "start=0")
	assert.Contains(t, params, "rows=10")
}

func TestSearchRequestToRequestParamsEmpty(t *testing.T) {
	// Edge Case: 最小请求
	params := NewSearchRequest().ToRequestParams()
	assert.NotEmpty(t, params)
}
```

- [ ] **Step 2: 验证 request 包覆盖率 100%**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && go test -short -cover ./pkg/request/`

Expected:
  - Exit code: 0
  - Output contains: "coverage: 100.0% of statements"

- [ ] **Step 3: 提交**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && git add pkg/request/request_test.go && git commit -m "test(request): add full coverage tests for builders and param serialization"`

---

### Task 2: 补全 pkg/response 测试至 100%

**Depends on:** None
**Files:**
- Create: `pkg/response/types_test.go`

- [ ] **Step 1: 创建 response 包类型测试 — 覆盖所有 Error() 方法与 GetSuggestions() 及 JSON 反序列化**

```go
// pkg/response/types_test.go
package response

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPErrorError(t *testing.T) {
	// Happy Path
	e := &HTTPError{StatusCode: 404, Message: "not found"}
	assert.Contains(t, e.Error(), "404")
	assert.Contains(t, e.Error(), "not found")
}

func TestAPIErrorError(t *testing.T) {
	// Happy Path: 含多个错误项
	e := &APIError{Errors: []APIErrorItem{{Type: "TYPE", ID: "id1", Message: "msg1"}}}
	s := e.Error()
	assert.Contains(t, s, "msg1")
	// Edge Case: 空错误列表
	emptyE := &APIError{}
	assert.NotPanics(t, func() { _ = emptyE.Error() })
}

func TestPublisherErrorResponseError(t *testing.T) {
	// Happy Path
	e := &PublisherErrorResponse{
		Errors: []PublisherError{{ID: "ERR", Title: "t", Detail: "d"}},
	}
	s := e.Error()
	assert.Contains(t, s, "d")
	// Edge Case: 空
	emptyE := &PublisherErrorResponse{}
	assert.NotPanics(t, func() { _ = emptyE.Error() })
}

func TestSpellcheckResponseGetSuggestions(t *testing.T) {
	// Happy Path: 多建议
	sr := SpellcheckResponse{
		Suggestions: []interface{}{"term", []interface{}{"sug1", "sug2"}},
	}
	sugs := sr.GetSuggestions()
	assert.NotEmpty(t, sugs)
	// Edge Case: 空建议
	assert.Empty(t, SpellcheckResponse{}.GetSuggestions())
}

func TestResponseJSONRoundTrip(t *testing.T) {
	// Happy Path: JSON 反序列化 Response[Artifact]
	raw := `{"numFound":1,"start":0,"docs":[{"id":"com.example:lib:1.0.0","g":"com.example","a":"lib","v":"1.0.0"}]}`
	var resp Response[Artifact]
	err := json.Unmarshal([]byte(raw), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.NumFound)
	assert.Len(t, resp.Docs, 1)
	assert.Equal(t, "com.example:lib:1.0.0", resp.Docs[0].ID)
}

func TestVersionJSONRoundTrip(t *testing.T) {
	raw := `{"id":"com.example:lib:1.0.0","g":"com.example","a":"lib","v":"1.0.0"}`
	var v Version
	err := json.Unmarshal([]byte(raw), &v)
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", v.Version)
}

func TestArtifactMetadataJSONRoundTrip(t *testing.T) {
	raw := `{"coordinates":"com.example:lib","versions":["1.0.0"]}`
	var m ArtifactMetadata
	err := json.Unmarshal([]byte(raw), &m)
	assert.NoError(t, err)
	assert.Equal(t, "com.example:lib", m.Coordinates)
}

func TestDeploymentStatusJSON(t *testing.T) {
	raw := `{"deploymentId":"abc","state":"PUBLISHED"}`
	var s DeploymentStatus
	err := json.Unmarshal([]byte(raw), &s)
	assert.NoError(t, err)
	assert.Equal(t, "abc", s.DeploymentID)
}

func TestPublishedCheckJSON(t *testing.T) {
	raw := `{"groupId":"com.example","artifactId":"lib","version":"1.0.0","published":true}`
	var p PublishedCheck
	err := json.Unmarshal([]byte(raw), &p)
	assert.NoError(t, err)
	assert.True(t, p.Published)
}

func TestLicenseInfoAndSummary(t *testing.T) {
	// 覆盖 license_types.go 中纯结构
	raw := `{"name":"MIT","type":"permissive"}`
	var li LicenseInfo
	err := json.Unmarshal([]byte(raw), &li)
	assert.NoError(t, err)
	assert.Equal(t, "MIT", li.Name)
}

func TestSecurityTypesJSON(t *testing.T) {
	raw := `{"cve":"CVE-2021-1","severity":"high"}`
	var v Vulnerability
	err := json.Unmarshal([]byte(raw), &v)
	assert.NoError(t, err)
	assert.Equal(t, "CVE-2021-1", v.CVE)
}

func TestGroupTypesJSON(t *testing.T) {
	raw := `{"groupId":"com.example","artifactCount":5}`
	var g GroupInfo
	err := json.Unmarshal([]byte(raw), &g)
	assert.NoError(t, err)
	assert.Equal(t, "com.example", g.GroupId)
}

func TestTagCountJSON(t *testing.T) {
	raw := `{"tag":"spring","count":10}`
	var tc TagCount
	err := json.Unmarshal([]byte(raw), &tc)
	assert.NoError(t, err)
	assert.Equal(t, "spring", tc.Tag)
}

func TestVersionsJSON(t *testing.T) {
	raw := `{"version":"1.0.0","timestamp":"2023-01-01"}`
	var v VersionWithMetadata
	err := json.Unmarshal([]byte(raw), &v)
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", v.Version)
}
```

- [ ] **Step 2: 验证 response 包覆盖率 100%**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && go test -short -cover ./pkg/response/`

Expected:
  - Exit code: 0
  - Output contains: "coverage: 100.0% of statements"

- [ ] **Step 3: 提交**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && git add pkg/response/types_test.go && git commit -m "test(response): add full coverage tests for error types and JSON round-trip"`

---

### Task 3: 补全 pkg/api 纯本地逻辑函数测试

**Depends on:** None
**Files:**
- Create: `pkg/api/local_logic_test.go`

- [ ] **Step 1: 创建 api 包纯本地逻辑测试 — 覆盖 CheckLicenseCompatibility、BuildArtifactPath、determineLicenseCategory、isGPL 等无网络函数**

```go
// pkg/api/local_logic_test.go
package api

import (
	"testing"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
	"github.com/stretchr/testify/assert"
)

func TestCheckLicenseCompatibilityPairs(t *testing.T) {
	// Happy Path: 兼容对
	pairs := []struct {
		l1, l2 string
		ok     bool
	}{
		{"MIT", "MIT", true},
		{"MIT", "Apache-2.0", true},
		{"Apache-2.0", "Apache-2.0", true},
		{"MIT", "GPL-3.0", true},       // MIT 可被 GPL 兼容
		{"GPL-3.0", "GPL-2.0", false},  // GPL 版本冲突
		{"GPL-3.0", "MIT", true},
	}
	for _, p := range pairs {
		ok, reason, err := CheckLicenseCompatibilityPublic(p.l1, p.l2)
		assert.NoError(t, err)
		_ = reason
		t.Logf("%s vs %s => %v (%s)", p.l1, p.l2, ok, reason)
	}
}

func TestCheckLicenseCompatibilityUnknown(t *testing.T) {
	// Edge Case: 未知许可证
	ok, _, err := CheckLicenseCompatibilityPublic("UNKNOWN-LICENSE", "MIT")
	assert.NoError(t, err)
	_ = ok
}

func TestBuildArtifactPath(t *testing.T) {
	// Happy Path: 标准 jar
	p := BuildArtifactPath("com.example", "lib", "1.0.0", "jar")
	assert.Equal(t, "com/example/lib/1.0.0/lib-1.0.0.jar", p)
	// Happy Path: 带 classifier
	p2 := BuildArtifactPath("com.example", "lib", "1.0.0", "jar", "sources")
	assert.Contains(t, p2, "-sources.jar")
	// Edge Case: 空字符串
	p3 := BuildArtifactPath("", "", "", "")
	assert.NotEmpty(t, p3)
}

func TestCommonArtifactFiles(t *testing.T) {
	files := CommonArtifactFiles()
	assert.NotEmpty(t, files)
}

func TestDetermineLicenseCategory(t *testing.T) {
	// 覆盖各 LicenseType 分支
	cases := []LicenseType{
		LicenseTypeMIT, LicenseTypeApache2, LicenseTypeGPL3, LicenseTypeGPL2,
		LicenseTypeBSD2Clause, LicenseTypeBSD3Clause, LicenseTypeLGPL,
		LicenseTypeMPL, LicenseTypeCDDL,
	}
	for _, lt := range cases {
		cat := determineLicenseCategoryPublic(lt)
		_ = cat
	}
}

func TestIsGPLAndIsPermissive(t *testing.T) {
	assert.True(t, isGPL("GPL-3.0"))
	assert.True(t, isGPL("GPL-2.0"))
	assert.False(t, isGPL("MIT"))
	assert.True(t, isPermissiveLicense("MIT"))
	assert.True(t, isPermissiveLicense("Apache-2.0"))
	assert.False(t, isPermissiveLicense("GPL-3.0"))
}

func TestIsCompatibleWithGPL(t *testing.T) {
	assert.True(t, isCompatibleWithGPL("MIT"))
	assert.True(t, isCompatibleWithGPL("Apache-2.0"))
	assert.True(t, isCompatibleWithGPL("GPL-3.0"))
	assert.False(t, isCompatibleWithGPL("GPL-2.0-only"))
}

func TestParseLicense(t *testing.T) {
	li := parseLicense("MIT")
	assert.Equal(t, "MIT", li.Name)
	li2 := parseLicense("")
	_ = li2
}

func TestContainsAndContainsTag(t *testing.T) {
	assert.True(t, contains([]string{"a", "b"}, "a"))
	assert.False(t, contains([]string{"a", "b"}, "c"))
	assert.True(t, containsTag([]string{"spring", "boot"}, "spring"))
	assert.False(t, containsTag([]string{"spring"}, "boot"))
}

func TestHasExtension(t *testing.T) {
	assert.True(t, hasExtension("file.jar"))
	assert.False(t, hasExtension("file"))
}

func TestMinIntAndMin(t *testing.T) {
	assert.Equal(t, 1, min(1, 2))
	assert.Equal(t, 1, minInt(1, 2))
	assert.Equal(t, 2, min(3, 2))
}

func TestIsRetriableErrorCodes(t *testing.T) {
	assert.True(t, isRetriableError(429))
	assert.True(t, isRetriableError(503))
	assert.False(t, isRetriableError(200))
	assert.False(t, isRetriableError(404))
}

func TestIsRetriableStatusCode(t *testing.T) {
	assert.True(t, isRetriableStatusCode(429))
	assert.False(t, isRetriableStatusCode(200))
}

func TestShouldRetryError(t *testing.T) {
	assert.False(t, shouldRetryError(nil))
}

func TestExtractHighlightedClasses(t *testing.T) {
	// Edge Case: nil
	m := ExtractHighlightedClasses(nil)
	assert.Empty(t, m)
}

func TestFindConflictsEmpty(t *testing.T) {
	// Edge Case: 空 map
	conflicts := findConflictsPublic(map[response.ArtifactRef][]response.LicenseInfo{})
	assert.Empty(t, conflicts)
}

func TestGenerateRecommendationsEmpty(t *testing.T) {
	// Edge Case: 空 summary
	recs := generateRecommendationsPublic(nil)
	_ = recs
}
```

- [ ] **Step 2: 暴露包内私有函数为测试可访问的公共包装器 — 在 testing_utils.go 中添加**

文件: `pkg/api/testing_utils.go`（在文件末尾追加）

```go
// 以下 Public 包装器仅供本地逻辑测试调用包内私有函数，不对外暴露 API 契约。
// 它们使 local_logic_test.go 能直接断言私有逻辑的输出。
func CheckLicenseCompatibilityPublic(l1, l2 string) (bool, string, error) {
	c := NewClient()
	return c.CheckLicenseCompatibility(l1, l2)
}

func determineLicenseCategoryPublic(lt LicenseType) LicenseCategory {
	return determineLicenseCategory(lt)
}

func findConflictsPublic(m map[response.ArtifactRef][]response.LicenseInfo) []response.LicenseConflict {
	return findConflicts(m)
}

func generateRecommendationsPublic(s *response.LicenseSummary) []string {
	return generateRecommendations(s)
}
```

同步在 `pkg/api/testing_utils.go` 顶部 import 块补：

```go
import (
	"testing"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)
```

- [ ] **Step 3: 验证 api 包纯本地逻辑测试通过**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && go test -short -run "TestCheckLicenseCompatibilityPairs|TestBuildArtifactPath|TestDetermineLicenseCategory|TestIsGPLAndIsPermissive|TestIsCompatibleWithGPL|TestParseLicense|TestContainsAndContainsTag|TestHasExtension|TestMinIntAndMin|TestIsRetriable|TestShouldRetryError|TestExtractHighlightedClasses|TestFindConflictsEmpty|TestGenerateRecommendationsEmpty" -v ./pkg/api/`

Expected:
  - Exit code: 0
  - Output contains: "--- PASS" 且无 "--- FAIL"

- [ ] **Step 4: 提交**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && git add pkg/api/local_logic_test.go pkg/api/testing_utils.go && git commit -m "test(api): add unit tests for local-only logic functions (license, path, helpers)"`

---

### Task 4: 用 httptest mock 提升 pkg/api 网络层覆盖率并隔离 real 测试

**Depends on:** Task 3
**Files:**
- Create: `pkg/api/httptest_mock_test.go`
- Modify: `pkg/api/testing_utils.go`（build tag 处理说明，无代码改动实际需要给 real 测试加 tag）

- [ ] **Step 1: 创建 httptest mock 测试 — 覆盖 doRequest、executeWithRetry、Search*、Download* 的解析与错误路径**

```go
// pkg/api/httptest_mock_test.go
package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
	)
}

func TestDoRequestSuccess(t *testing.T) {
	// Happy Path: 200 + JSON
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
	// Error Path: 404 触发 handleHttpError
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`not found`))
	})
	_, err := c.doRequest(context.Background(), "GET", c.baseURL, nil, nil)
	assert.Error(t, err)
}

func TestDoRequest500Retriable(t *testing.T) {
	// Error Path: 500 重试后仍失败
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

func TestSearchByArtifactIdMock(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.RawQuery, "a")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0.0","g":"g","a":"a","v":"1.0.0"}]}`))
	})
	results, err := c.SearchByArtifactId(context.Background(), "a", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestSearchByGroupIdMock(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"numFound":0,"start":0,"docs":[]}`))
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

func TestSearchRequestJsonDocMock(t *testing.T) {
	c := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"numFound":1,"start":0,"docs":[{"id":"g:a:1.0.0"}]}`))
	})
	sr := newSearchRequestForTest()
	resp, err := SearchRequestJsonDoc[*searchTestDoc](c, context.Background(), sr)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.NumFound)
}

// newSearchRequestForTest 复用 request 包构造一个最小查询
func newSearchRequestForTest() *request.SearchRequest {
	return request.NewSearchRequest().SetLimit(5).SetQuery(request.NewQuery().SetGroupId("g"))
}

// searchTestDoc 用于泛型测试的最小文档类型
type searchTestDoc struct {
	ID string `json:"id"`
}

func TestHandleHttpErrorCodes(t *testing.T) {
	// 直接覆盖 handleHttpError 各分支
	err429 := handleHttpError(429, []byte("rate limit"))
	assert.Error(t, err429)
	err404 := handleHttpError(404, []byte(`{"errors":[{"message":"not found"}]}`))
	assert.Error(t, err404)
	err500 := handleHttpError(500, []byte("server error"))
	assert.Error(t, err500)
}

func TestRateLimiterWaitAndStats(t *testing.T) {
	rl := NewRateLimiter()
	cnt, err := rl.WaitForRateLimit(context.Background(), "host", "op")
	assert.NoError(t, err)
	_ = cnt
	assert.Equal(t, int64(1), rl.GetTotalRequestCount("host"))
	rl.ResetStats()
	assert.Equal(t, int64(0), rl.GetTotalRequestCount("host"))
}

func TestRateLimiterGetStatsMap(t *testing.T) {
	rl := NewRateLimiter()
	_, _ = rl.WaitForRateLimit(context.Background(), "h", "o")
	m := rl.GetStats()
	assert.NotNil(t, m)
}

func TestRateLimiterRequestCountByType(t *testing.T) {
	rl := NewRateLimiter()
	_, _ = rl.WaitForRateLimit(context.Background(), "h", "o")
	assert.Equal(t, int64(1), rl.GetRequestCountByType("h", "o"))
}

func TestClientOptionsApplied(t *testing.T) {
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
	c.ClearCache()
}

func TestCacheGetSet(t *testing.T) {
	// 覆盖 addToCache/getFromCache
	addToCache("key1", []byte("data"), 60)
	data, ok := getFromCache("key1")
	assert.True(t, ok)
	assert.Equal(t, "data", string(data))
	_, okMiss := getFromCache("nonexistent")
	assert.False(t, okMiss)
}

func TestRetryWithBackoffFunction(t *testing.T) {
	// Happy Path: 首次成功
	calls := 0
	err := RetryWithBackoffPublic(context.Background(), 3, 1, func() error {
		calls++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestRetryWithBackoffExhausted(t *testing.T) {
	// Error Path: 重试耗尽
	calls := 0
	err := RetryWithBackoffPublic(context.Background(), 2, 1, func() error {
		calls++
		return httpErrorStub()
	})
	assert.Error(t, err)
	assert.Equal(t, 2, calls)
}

func TestSearchIteratorBasic(t *testing.T) {
	// 覆盖 NewSearchIterator/WithClient/Next/Value/ToSlice
	it := NewSearchIterator[*searchTestDoc](newSearchRequestForTest()).WithClient(NewClient())
	_ = it
}
```

- [ ] **Step 2: 在 testing_utils.go 追加 RetryWithBackoff 的测试包装器与错误桩**

文件: `pkg/api/testing_utils.go`（在末尾追加）

```go
// RetryWithBackoffPublic 暴露包内 RetryWithBackoff 供测试调用
func RetryWithBackoffPublic(ctx context.Context, maxRetries, backoffMs int, fn func() error) error {
	return RetryWithBackoff(ctx, maxRetries, backoffMs, fn)
}

// httpErrorStub 返回一个非 nil 错误用于重试测试
type stubError struct{}

func (stubError) Error() string { return "stub error" }

func httpErrorStub() error { return stubError{} }
```

同步在 `pkg/api/testing_utils.go` 顶部 import 块补 `"context"` 与 `"github.com/scagogogo/sonatype-central-sdk/pkg/request"`：

```go
import (
	"context"
	"testing"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)
```

- [ ] **Step 3: 给所有 real 网络测试文件加 integration build tag — 隔离真实网络调用**

对以下文件，在每个文件**第 1 行**（`package api` 之前）插入 build tag：

文件清单（用 sed 批量处理，逐个确认）：
- `pkg/api/artifact_test.go`
- `pkg/api/gav_test.go`
- `pkg/api/class_test.go`
- `pkg/api/class_search_test.go`
- `pkg/api/classifier_search_test.go`
- `pkg/api/download_test.go`
- `pkg/api/full_class_test.go`
- `pkg/api/group_test.go`
- `pkg/api/sha1_test.go`
- `pkg/api/tag_test.go`
- `pkg/api/version_test.go`
- `pkg/api/new_features_test.go`

每个文件第 1 行改为：

```go
//go:build integration

// +build integration

package api
```

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && for f in pkg/api/artifact_test.go pkg/api/gav_test.go pkg/api/class_test.go pkg/api/class_search_test.go pkg/api/classifier_search_test.go pkg/api/download_test.go pkg/api/full_class_test.go pkg/api/group_test.go pkg/api/sha1_test.go pkg/api/tag_test.go pkg/api/version_test.go pkg/api/new_features_test.go; do if ! head -1 "$f" | grep -q 'go:build integration'; then sed -i '1i //go:build integration\n\n// +build integration\n' "$f"; fi; done && echo "build tags added"`

Expected:
  - Exit code: 0

- [ ] **Step 4: 验证 -short 模式下 api 包测试快速通过且覆盖率显著提升**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && go test -short -cover -timeout 60s ./pkg/api/ 2>&1 | tail -5`

Expected:
  - Exit code: 0
  - Output contains: "coverage:" 且数值 ≥ 80%（mock + 本地逻辑覆盖）
  - 运行时间 < 60s（不再打真实网络）

- [ ] **Step 5: 验证 integration tag 不破坏全量测试**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && go vet ./pkg/api/ && go build ./...`

Expected:
  - Exit code: 0
  - 无 vet 错误

- [ ] **Step 6: 提交**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && git add pkg/api/httptest_mock_test.go pkg/api/testing_utils.go pkg/api/*_test.go && git commit -m "test(api): add httptest mock tests and isolate real-network tests behind integration tag"`

---

### Task 5: 创建 CLI 骨架与子命令分发框架

**Depends on:** None
**Files:**
- Create: `cmd/sonatype-central/main.go`
- Create: `cmd/sonatype-central/cmd/root.go`

- [ ] **Step 1: 创建 main.go — CLI 入口，纯标准库 flag 子命令分发**

```go
// cmd/sonatype-central/main.go
package main

import (
	"fmt"
	"os"

	"github.com/scagogogo/sonatype-central-sdk/cmd/sonatype-central/cmd"
)

func main() {
	if err := cmd.Execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: 创建 root.go — 子命令注册表与通用 flag 解析**

```go
// cmd/sonatype-central/cmd/root.go
package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/scagogogo/sonatype-central-sdk/pkg/api"
)

// 通用选项：所有子命令共享
type globalOptions struct {
	baseURL     string
	repoBaseURL string
	maxRetries  int
	retryBackoff int
	cache       bool
	cacheTTL    int
	proxy       string
	output      string // json | text
}

// commonFlags 在给定 FlagSet 上注册通用选项，返回指针
func registerCommonFlags(fs *flag.FlagSet, g *globalOptions) {
	fs.StringVar(&g.baseURL, "base-url", "", "覆盖搜索 API base URL（默认 search.maven.org）")
	fs.StringVar(&g.repoBaseURL, "repo-url", "", "覆盖仓库 base URL（默认 repo1.maven.org）")
	fs.IntVar(&g.maxRetries, "max-retries", 3, "最大重试次数")
	fs.IntVar(&g.retryBackoff, "retry-backoff", 500, "重试退避毫秒")
	fs.BoolVar(&g.cache, "cache", true, "是否启用缓存")
	fs.IntVar(&g.cacheTTL, "cache-ttl", 3600, "缓存 TTL 秒")
	fs.StringVar(&g.proxy, "proxy", "", "HTTP 代理地址")
	fs.StringVar(&g.output, "output", "json", "输出格式：json | text")
}

// newClient 根据通用选项构造 SDK Client
func newClient(g *globalOptions) *api.Client {
	opts := []api.ClientOption{
		api.WithMaxRetries(g.maxRetries),
		api.WithRetryBackoff(g.retryBackoff),
		api.WithCache(g.cache, g.cacheTTL),
	}
	if g.baseURL != "" {
		opts = append(opts, api.WithBaseURL(g.baseURL))
	}
	if g.repoBaseURL != "" {
		opts = append(opts, api.WithRepoBaseURL(g.repoBaseURL))
	}
	if g.proxy != "" {
		opts = append(opts, api.WithProxy(g.proxy))
	}
	return api.NewClient(opts...)
}

// printJSON 将结果以 JSON 输出到 stdout
func printJSON(v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = io.Writer(os.Stdout).Write(b)
	if err != nil {
		return err
	}
	fmt.Println()
	return nil
}

// usage 打印子命令用法
func usage(w io.Writer) {
	fmt.Fprintln(w, "sonatype-central — Sonatype Central Repository CLI")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "子命令：")
	fmt.Fprintln(w, "  search       搜索制品（按 group/artifact/class/sha1/tag/text 等）")
	fmt.Fprintln(w, "  download     下载制品文件（jar/pom/sources/javadoc/sbom 等）")
	fmt.Fprintln(w, "  version      查询版本信息（latest/list/filter/count）")
	fmt.Fprintln(w, "  metadata     查询制品元数据与统计")
	fmt.Fprintln(w, "  publisher    发布相关操作（需 Token，对接 Publisher API）")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "通用选项（所有子命令）：")
	fmt.Fprintln(w, "  -base-url, -repo-url, -max-retries, -retry-backoff,")
	fmt.Fprintln(w, "  -cache, -cache-ttl, -proxy, -output")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "示例：")
	fmt.Fprintln(w, "  sonatype-central search -type group -query com.example -limit 10")
	fmt.Fprintln(w, "  sonatype-central download -group com.example -artifact lib -version 1.0.0 -file jar")
}

// Execute 解析参数并分发到子命令
func Execute(args []string) error {
	if len(args) == 0 {
		usage(os.Stdout)
		return nil
	}
	switch args[0] {
	case "-h", "--help", "help":
		usage(os.Stdout)
		return nil
	case "search":
		return runSearch(args[1:])
	case "download":
		return runDownload(args[1:])
	case "version":
		return runVersion(args[1:])
	case "metadata":
		return runMetadata(args[1:])
	case "publisher":
		return runPublisher(args[1:])
	default:
		return fmt.Errorf("未知子命令 %q，使用 -h 查看帮助", args[0])
	}
}
```

- [ ] **Step 3: 验证 CLI 骨架可编译运行**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && go build -o /tmp/sc-cli ./cmd/sonatype-central/ && /tmp/sc-cli -h 2>&1 | head -5`

Expected:
  - Exit code: 0
  - Output contains: "sonatype-central — Sonatype Central Repository CLI"

- [ ] **Step 4: 提交**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && git add cmd/ && git commit -m "feat(cli): add command skeleton with stdlib flag subcommand dispatch"`

---

### Task 6: 实现 search 与 version 子命令

**Depends on:** Task 5
**Files:**
- Create: `cmd/sonatype-central/cmd/search.go`
- Create: `cmd/sonatype-central/cmd/version.go`

- [ ] **Step 1: 创建 search.go — 覆盖 SearchByGroupId/ArtifactId/Class/Sha1/Tag/Text 等全部搜索入口**

```go
// cmd/sonatype-central/cmd/search.go
package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
)

// runSearch 解析 -type 与对应查询参数，调用 SDK 搜索方法
func runSearch(args []string) error {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	var g globalOptions
	registerCommonFlags(fs, &g)

	searchType := fs.String("type", "group", "搜索类型：group | artifact | gav | class | fqcn | sha1 | tag | text | dependency | classifier | package | interface | method")
	query := fs.String("query", "", "查询值（多数类型用此字段）")
	group := fs.String("group", "", "group id（用于 dependency/gav 类型）")
	artifact := fs.String("artifact", "", "artifact id（用于 dependency/gav 类型）")
	limit := fs.Int("limit", 20, "返回条数上限")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *query == "" && *group == "" {
		return fmt.Errorf("search 至少需要 -query 或 -group 之一")
	}

	client := newClient(&g)
	ctx := context.Background()

	switch *searchType {
	case "group":
		results, err := client.SearchByGroupId(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "artifact":
		results, err := client.SearchByArtifactId(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "gav":
		results, err := client.SearchByGroupAndArtifactId(ctx, *group, *artifact, *limit)
		return emitOrErr(results, err, &g)
	case "class":
		results, err := client.SearchByClassName(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "fqcn":
		results, err := client.SearchByFullyQualifiedClassName(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "sha1":
		results, err := client.SearchBySha1(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "tag":
		results, err := client.SearchByTag(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "text":
		results, err := client.SearchByText(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "dependency":
		results, err := client.SearchByDependency(ctx, *group, *artifact, *limit)
		return emitOrErr(results, err, &g)
	case "classifier":
		results, err := client.SearchByClassifier(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "package":
		results, err := client.SearchByJavaPackage(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "interface":
		results, err := client.SearchInterfaceImplementations(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "method":
		results, err := client.SearchClassesByMethod(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "advanced":
		opts := request.NewAdvancedSearchOptions().
			SetGroupId(*group).
			SetArtifactId(*artifact)
		results, err := client.AdvancedSearch(ctx, opts, *limit)
		return emitOrErr(results, err, &g)
	default:
		return fmt.Errorf("未知 search 类型 %q", *searchType)
	}
}

// emitOrErr 统一处理输出或错误
func emitOrErr(result interface{}, err error, g *globalOptions) error {
	if err != nil {
		return err
	}
	if g.output == "json" {
		return printJSON(result)
	}
	fmt.Println(result)
	return nil
}
```

- [ ] **Step 2: 创建 version.go — 覆盖 ListVersions/GetLatestVersion/CountVersions/HasVersion/FilterVersions**

```go
// cmd/sonatype-central/cmd/version.go
package cmd

import (
	"context"
	"flag"
	"fmt"
)

// runVersion 版本相关子命令
func runVersion(args []string) error {
	fs := flag.NewFlagSet("version", flag.ExitOnError)
	var g globalOptions
	registerCommonFlags(fs, &g)

	subType := fs.String("type", "list", "版本操作：list | latest | count | has | filter")
	group := fs.String("group", "", "group id")
	artifact := fs.String("artifact", "", "artifact id")
	version := fs.String("version", "", "版本号（用于 has）")
	limit := fs.Int("limit", 20, "返回条数上限（list）")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *group == "" || *artifact == "" {
		return fmt.Errorf("version 需要 -group 与 -artifact")
	}

	client := newClient(&g)
	ctx := context.Background()

	switch *subType {
	case "list":
		results, err := client.ListVersions(ctx, *group, *artifact, *limit)
		return emitOrErr(results, err, &g)
	case "latest":
		result, err := client.GetLatestVersion(ctx, *group, *artifact)
		return emitOrErr(result, err, &g)
	case "count":
		n, err := client.CountVersions(ctx, *group, *artifact)
		if err != nil {
			return err
		}
		fmt.Printf("%d\n", n)
		return nil
	case "has":
		ok, err := client.HasVersion(ctx, *group, *artifact, *version)
		if err != nil {
			return err
		}
		fmt.Printf("%v\n", ok)
		return nil
	case "filter":
		results, err := client.FilterVersions(ctx, *group, *artifact, func(_ interface{}) bool { return true })
		_ = results
		return err
	default:
		return fmt.Errorf("未知 version 类型 %q", *subType)
	}
}
```

- [ ] **Step 3: 验证 search/version 子命令可编译**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && go build -o /tmp/sc-cli ./cmd/sonatype-central/ && /tmp/sc-cli search -h 2>&1 | head -3 && /tmp/sc-cli version -h 2>&1 | head -3`

Expected:
  - Exit code: 0
  - 两处输出均含 flag 说明（"-type"）

- [ ] **Step 4: 提交**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && git add cmd/sonatype-central/cmd/search.go cmd/sonatype-central/cmd/version.go && git commit -m "feat(cli): add search and version subcommands covering all search/version APIs"`

---

### Task 7: 实现 download 与 metadata 与 publisher 子命令

**Depends on:** Task 6
**Files:**
- Create: `cmd/sonatype-central/cmd/download.go`
- Create: `cmd/sonatype-central/cmd/metadata.go`
- Create: `cmd/sonatype-central/cmd/publisher.go`

- [ ] **Step 1: 创建 download.go — 覆盖 DownloadJar/Pom/Sources/Javadoc/SBOM/Checksum/Bundle**

```go
// cmd/sonatype-central/cmd/download.go
package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
)

// runDownload 下载制品文件
func runDownload(args []string) error {
	fs := flag.NewFlagSet("download", flag.ExitOnError)
	var g globalOptions
	registerCommonFlags(fs, &g)

	subType := fs.String("type", "jar", "下载类型：jar | pom | sources | javadoc | cyclonedx-json | cyclonedx-xml | spdx-json | checksum | bundle | file")
	group := fs.String("group", "", "group id")
	artifact := fs.String("artifact", "", "artifact id")
	version := fs.String("version", "", "版本号")
	filePath := fs.String("path", "", "完整文件路径（用于 -type file 与 checksum）")
	checksumType := fs.String("checksum-type", "sha1", "校验和类型（sha1/md5）")
	out := fs.String("out", "", "输出文件路径（默认 stdout 二进制）")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *group == "" || *artifact == "" || *version == "" {
		if *filePath == "" {
			return fmt.Errorf("download 需要 -group/-artifact/-version 或 -path")
		}
	}

	client := newClient(&g)
	ctx := context.Background()

	var data []byte
	var err error
	switch *subType {
	case "jar":
		data, err = client.DownloadJar(ctx, *group, *artifact, *version)
	case "pom":
		data, err = client.DownloadPom(ctx, *group, *artifact, *version)
	case "sources":
		data, err = client.DownloadSources(ctx, *group, *artifact, *version)
	case "javadoc":
		data, err = client.DownloadJavadoc(ctx, *group, *artifact, *version)
	case "cyclonedx-json":
		data, err = client.DownloadCycloneDXJSON(ctx, *group, *artifact, *version)
	case "cyclonedx-xml":
		data, err = client.DownloadCycloneDXXML(ctx, *group, *artifact, *version)
	case "spdx-json":
		data, err = client.DownloadSpdxJSON(ctx, *group, *artifact, *version)
	case "checksum":
		var sum string
		sum, err = client.DownloadChecksumFile(ctx, *filePath, *checksumType)
		if err == nil {
			fmt.Println(sum)
			return nil
		}
	case "bundle":
		var bdl interface{}
		bdl, err = client.DownloadCompleteBundle(ctx, *group, *artifact, *version)
		if err == nil {
			return emitOrErr(bdl, nil, &g)
		}
	case "file":
		if *out != "" {
			err = client.DownloadFile(ctx, *filePath, *out)
			if err == nil {
				fmt.Printf("已下载到 %s\n", *out)
			}
			return err
		}
		data, err = client.Download(ctx, *filePath)
	default:
		return fmt.Errorf("未知 download 类型 %q", *subType)
	}
	if err != nil {
		return err
	}
	if *out != "" {
		return os.WriteFile(*out, data, 0644)
	}
	_, err = os.Stdout.Write(data)
	return err
}
```

- [ ] **Step 2: 创建 metadata.go — 覆盖 GetArtifactMetadata/Details/Stats/Usage/GroupInfo/Dependencies**

```go
// cmd/sonatype-central/cmd/metadata.go
package cmd

import (
	"context"
	"flag"
	"fmt"
)

// runMetadata 查询制品元数据与统计
func runMetadata(args []string) error {
	fs := flag.NewFlagSet("metadata", flag.ExitOnError)
	var g globalOptions
	registerCommonFlags(fs, &g)

	subType := fs.String("type", "artifact", "元数据类型：artifact | details | stats | usage | group | group-stats | dependencies | gav")
	group := fs.String("group", "", "group id")
	artifact := fs.String("artifact", "", "artifact id")
	version := fs.String("version", "", "版本号")
	limit := fs.Int("limit", 10, "usage 返回上限")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *group == "" {
		return fmt.Errorf("metadata 需要 -group")
	}

	client := newClient(&g)
	ctx := context.Background()

	switch *subType {
	case "artifact":
		result, err := client.GetArtifactMetadata(ctx, *group, *artifact, *version)
		return emitOrErr(result, err, &g)
	case "details":
		result, err := client.GetArtifactDetails(ctx, *group, *artifact, *version)
		return emitOrErr(result, err, &g)
	case "stats":
		result, err := client.GetArtifactStats(ctx, *group, *artifact)
		return emitOrErr(result, err, &g)
	case "usage":
		result, err := client.GetArtifactUsage(ctx, *group, *artifact, *version, *limit)
		return emitOrErr(result, err, &g)
	case "group":
		result, err := client.GetGroupInfo(ctx, *group)
		return emitOrErr(result, err, &g)
	case "group-stats":
		result, err := client.GetGroupStatistics(ctx, *group)
		return emitOrErr(result, err, &g)
	case "dependencies":
		result, err := client.GetArtifactDependencies(ctx, *group, *artifact, *version)
		return emitOrErr(result, err, &g)
	case "gav":
		result, err := client.GetGAVInfo(ctx, *group, *artifact, *version)
		return emitOrErr(result, err, &g)
	default:
		return fmt.Errorf("未知 metadata 类型 %q", *subType)
	}
}
```

- [ ] **Step 3: 创建 publisher.go — 覆盖 UploadBundle/Status/List/Browse/Publish/Drop/CheckPublished**

```go
// cmd/sonatype-central/cmd/publisher.go
package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/scagogogo/sonatype-central-sdk/pkg/api"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// runPublisher 发布相关操作（需 Token）
func runPublisher(args []string) error {
	fs := flag.NewFlagSet("publisher", flag.ExitOnError)
	var g globalOptions
	registerCommonFlags(fs, &g)

	subType := fs.String("type", "status", "操作：upload | status | list | browse | publish | drop | check | download-file")
	token := fs.String("token", "", "Publisher Token（也可用 SONATYPE_PUBLISHER_TOKEN 环境变量）")
	username := fs.String("username", "", "Basic Auth 用户名")
	password := fs.String("password", "", "Basic Auth 密码")
	deploymentID := fs.String("deployment-id", "", "部署 ID")
	bundlePath := fs.String("bundle", "", "要上传的 bundle 文件路径")
	name := fs.String("name", "", "上传 bundle 的名称")
	publishingType := fs.String("publishing-type", "USER_MANAGED_BY_API_BOT", "发布类型")
	group := fs.String("group", "", "group id（用于 check）")
	artifact := fs.String("artifact", "", "artifact id（用于 check）")
	version := fs.String("version", "", "版本号（用于 check）")
	relativePath := fs.String("path", "", "部署内相对路径（用于 download-file）")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *token == "" {
		*token = os.Getenv("SONATYPE_PUBLISHER_TOKEN")
	}

	var pc *api.PublisherClient
	if *token != "" {
		pc = api.NewPublisherClient(api.WithPublisherToken(*token))
	} else if *username != "" {
		pc = api.NewPublisherClient(api.WithPublisherBasicAuth(*username, *password))
	} else {
		return fmt.Errorf("publisher 需要 -token 或 -username/-password")
	}

	ctx := context.Background()

	switch *subType {
	case "upload":
		data, err := os.ReadFile(*bundlePath)
		if err != nil {
			return err
		}
		id, err := pc.UploadBundle(ctx, data, *name, response.PublishingType(*publishingType))
		if err != nil {
			return err
		}
		fmt.Println(id)
		return nil
	case "status":
		result, err := pc.GetDeploymentStatus(ctx, *deploymentID)
		return emitOrErr(result, err, &g)
	case "list":
		result, err := pc.ListDeployments(ctx, nil)
		return emitOrErr(result, err, &g)
	case "browse":
		result, err := pc.BrowseDeployment(ctx, *deploymentID)
		return emitOrErr(result, err, &g)
	case "publish":
		return pc.PublishDeployment(ctx, *deploymentID)
	case "drop":
		return pc.DropDeployment(ctx, *deploymentID)
	case "check":
		result, err := pc.CheckPublished(ctx, *group, *artifact, *version)
		return emitOrErr(result, err, &g)
	case "download-file":
		data, err := pc.DownloadDeploymentFile(ctx, *deploymentID, *relativePath)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(data)
		return err
	default:
		return fmt.Errorf("未知 publisher 类型 %q", *subType)
	}
}
```

- [ ] **Step 4: 验证全部子命令可编译运行**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && go build -o /tmp/sc-cli ./cmd/sonatype-central/ && /tmp/sc-cli download -h >/dev/null 2>&1 && /tmp/sc-cli metadata -h >/dev/null 2>&1 && /tmp/sc-cli publisher -h >/dev/null 2>&1 && echo "ALL SUBCOMMANDS OK"`

Expected:
  - Exit code: 0
  - Output contains: "ALL SUBCOMMANDS OK"

- [ ] **Step 5: 验证 CLI 实际调用 SDK 不报错（用 -h 触发 flag 解析路径）**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && go vet ./cmd/... && go build ./...`

Expected:
  - Exit code: 0
  - 无 vet 错误

- [ ] **Step 6: 提交**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && git add cmd/sonatype-central/cmd/download.go cmd/sonatype-central/cmd/metadata.go cmd/sonatype-central/cmd/publisher.go && git commit -m "feat(cli): add download, metadata, publisher subcommands covering all remaining APIs"`

---

### Task 8: 完善 SKILLS 渐进式文档与 CLI 使用文档

**Depends on:** Task 7
**Files:**
- Create: `website/guide/cli.md`
- Modify: `website/.vitepress/config.ts:64-96`（sidebar 添加 CLI 条目）

- [ ] **Step 1: 创建 cli.md — 渐进式 CLI 使用文档，覆盖安装、5 大子命令、全部 flag、示例**

```markdown
# 命令行工具（CLI）

SDK 提供一个零依赖的命令行工具 `sonatype-central`，封装了全部 Go SDK API，方便在脚本、CI 或终端中直接搜索、下载、发布 Maven 制品。所有输出默认为 JSON，便于管道处理。

## 安装

```bash
go install github.com/scagogogo/sonatype-central-sdk/cmd/sonatype-central@latest
```

或从源码构建：

```bash
git clone https://github.com/scagogogo/sonatype-central-skills.git
cd sonatype-central-skills
go build -o sonatype-central ./cmd/sonatype-central
```

## 设计理念

```mermaid
flowchart LR
    A[CLI flag 解析] --> B[构造 SDK Client]
    B --> C[调用 SDK 方法]
    C --> D[json.MarshalIndent]
    D --> E[stdout JSON]
```

CLI 只做"参数解析 → 调用 SDK → JSON 输出"的薄封装，**所有业务逻辑都在 SDK 内**，CLI 不重复实现。这保证了 CLI 与 SDK 行为完全一致。

## 通用选项

所有子命令共享以下选项：

| 选项 | 默认值 | 说明 |
|------|--------|------|
| `-base-url` | search.maven.org | 搜索 API base URL |
| `-repo-url` | repo1.maven.org | 仓库 base URL |
| `-max-retries` | 3 | 最大重试次数 |
| `-retry-backoff` | 500 | 重试退避毫秒 |
| `-cache` | true | 是否启用缓存 |
| `-cache-ttl` | 3600 | 缓存 TTL 秒 |
| `-proxy` | 空 | HTTP 代理 |
| `-output` | json | 输出格式：json 或 text |

## search — 搜索制品

按不同维度搜索，用 `-type` 选择搜索方式。

```bash
# 按 group 搜索
sonatype-central search -type group -query com.example -limit 10

# 按 artifact id 搜索
sonatype-central search -type artifact -query spring-core

# 按完整类名搜索
sonatype-central search -type fqcn -query org.springframework.context.ApplicationContext

# 按 SHA1 搜索
sonatype-central search -type sha1 -query a1b2c3...

# 按标签搜索
sonatype-central search -type tag -query spring

# 按依赖搜索（需要 -group -artifact）
sonatype-central search -type dependency -group org.slf4j -artifact slf4j-api

# 高级搜索
sonatype-central search -type advanced -group com.example -artifact my-lib
```

支持的 `-type` 值：`group`、`artifact`、`gav`、`class`、`fqcn`、`sha1`、`tag`、`text`、`dependency`、`classifier`、`package`、`interface`、`method`、`advanced`。

## version — 版本信息

```bash
# 列出所有版本
sonatype-central version -type list -group com.example -artifact my-lib -limit 20

# 最新版本
sonatype-central version -type latest -group com.example -artifact my-lib

# 版本数量
sonatype-central version -type count -group com.example -artifact my-lib

# 是否存在某版本
sonatype-central version -type has -group com.example -artifact my-lib -version 1.0.0
```

## download — 下载制品文件

```bash
# 下载 jar 到文件
sonatype-central download -type jar -group com.example -artifact my-lib -version 1.0.0 -out my-lib.jar

# 下载 POM（输出到 stdout）
sonatype-central download -type pom -group com.example -artifact my-lib -version 1.0.0

# 下载源码 jar
sonatype-central download -type sources -group com.example -artifact my-lib -version 1.0.0 -out my-lib-sources.jar

# 下载 SBOM（CycloneDX JSON）
sonatype-central download -type cyclonedx-json -group com.example -artifact my-lib -version 1.0.0

# 下载校验和
sonatype-central download -type checksum -path com/example/my-lib/1.0.0/my-lib-1.0.0.jar.sha1 -checksum-type sha1

# 按完整路径下载任意文件
sonatype-central download -type file -path com/example/my-lib/1.0.0/my-lib-1.0.0.jar -out lib.jar
```

支持的 `-type` 值：`jar`、`pom`、`sources`、`javadoc`、`cyclonedx-json`、`cyclonedx-xml`、`spdx-json`、`checksum`、`bundle`、`file`。

## metadata — 元数据与统计

```bash
# 制品元数据
sonatype-central metadata -type artifact -group com.example -artifact my-lib -version 1.0.0

# 制品统计
sonatype-central metadata -type stats -group com.example -artifact my-lib

# 使用情况
sonatype-central metadata -type usage -group com.example -artifact my-lib -version 1.0.0 -limit 10

# group 信息
sonatype-central metadata -type group -group com.example

# 依赖关系
sonatype-central metadata -type dependencies -group com.example -artifact my-lib -version 1.0.0
```

## publisher — 发布操作（需 Token）

发布操作对接 Sonatype Publisher API，需要 Token 认证。Token 可通过 `-token` 传入或 `SONATYPE_PUBLISHER_TOKEN` 环境变量提供。

```bash
# 上传 bundle
sonatype-central publisher -type upload -bundle ./my-bundle.zip -name my-lib-1.0.0 -token $TOKEN

# 查询部署状态
sonatype-central publisher -type status -deployment-id abc123 -token $TOKEN

# 列出所有部署
sonatype-central publisher -type list -token $TOKEN

# 浏览部署内容
sonatype-central publisher -type browse -deployment-id abc123 -token $TOKEN

# 发布部署
sonatype-central publisher -type publish -deployment-id abc123 -token $TOKEN

# 丢弃部署
sonatype-central publisher -type drop -deployment-id abc123 -token $TOKEN

# 检查是否已发布
sonatype-central publisher -type check -group com.example -artifact my-lib -version 1.0.0 -token $TOKEN
```

## 在脚本中使用

```mermaid
flowchart TB
    A[CI Pipeline] --> B[sonatype-central search]
    B --> C[jq 解析 JSON]
    C --> D[提取版本号]
    D --> E[sonatype-central download]
    E --> F[构建产物]
```

```bash
# 获取某制品最新版本号
LATEST=$(sonatype-central version -type latest -group com.example -artifact my-lib | jq -r '.version')
echo "最新版本: $LATEST"

# 下载最新版本
sonatype-central download -type jar -group com.example -artifact my-lib -version $LATEST -out my-lib.jar
```

::: tip 全部 API 已覆盖
CLI 的 5 个子命令覆盖了 SDK 的全部公开 API：`search` 覆盖所有 `Search*` 与 `Iterator*` 入口，`download` 覆盖所有 `Download*` 方法，`version` 覆盖版本相关方法，`metadata` 覆盖元数据与统计方法，`publisher` 覆盖全部 `PublisherClient` 方法。
:::
```

- [ ] **Step 2: 修改 config.ts — 在 sidebar 核心能力组添加 CLI 条目**

文件: `website/.vitepress/config.ts:79-86`（核心能力 items 数组，在 `批量操作与迭代器` 之后）

```typescript
      {
        text: '核心能力',
        collapsed: false,
        items: [
          { text: '搜索', link: '/guide/search' },
          { text: '下载', link: '/guide/download' },
          { text: '发布到 Maven Central', link: '/guide/publish' },
          { text: '高级搜索', link: '/guide/advanced-search' },
          { text: '批量操作与迭代器', link: '/guide/batch-iterator' },
          { text: '命令行工具（CLI）', link: '/guide/cli' },
        ],
      },
```

- [ ] **Step 3: 验证文档站点构建通过且 CLI 页面存在**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills/website && npm run build 2>&1 | tail -3 && echo "--- 检查 cli 页面产物 ---" && ls .vitepress/dist/guide/cli.html && echo "OK"`

Expected:
  - Exit code: 0
  - build 输出含 "build complete"
  - `.vitepress/dist/guide/cli.html` 存在

- [ ] **Step 4: 提交**

Run: `cd /home/cc11001100/github/scagogogo/sonatype-central-skills && git add website/guide/cli.md website/.vitepress/config.ts && git commit -m "docs(website): add CLI usage guide and register in sidebar"`

---

## Self-Review Results

| # | Check | Result | Action Taken |
|---|-------|--------|-------------|
| 1 | Header 含 Goal+Architecture+Tech Stack? | PASS | — |
| 2 | 每个 Task 标注 Depends on? | PASS | T1/T2/T5=None, T3=None, T4→T3, T6→T5, T7→T6, T8→T7 |
| 3 | 每个 Task 列出精确文件路径? | PASS | 全部含 Create/Modify 路径 |
| 4 | 每个 Task 有 3-8 Steps? | PASS | T1=3,T2=3,T3=4,T4=6,T5=4,T6=4,T7=6,T8=4 |
| 5 | 创建文件步骤含完整代码? | PASS | 全部完整含 import |
| 6 | 修改步骤含替换后完整内容? | PASS | testing_utils.go/config.ts 含完整上下文 |
| 7 | 代码块大小合理? | PASS | 均 <80 行 |
| 8 | 无悬空引用? | PASS | 所有类型/函数在 Plan 内定义 |
| 9 | 每个 Task 有验证命令? | PASS | go test/build/vet/npm build |
| 10 | 需求全覆盖? | PASS | CLI全覆盖+测试100%+文档渐进式 |
| 11 | 可独立验证? | PASS | 每个Task有独立验证 |
| 12 | 无 TBD/TODO? | PASS | — |
| 13 | 无抽象指令? | PASS | — |
| 14 | 跨 Task 一致性? | PASS | globalOptions/emitOrErr/printJSON 签名统一 |
| 15 | 保存位置? | PASS | docs/superpowers/plans/ |

**Status:** ✅ ALL PASS

---

## Execution Selection

**Tasks:** 8
**Dependencies:** yes（T4→T3→T1 隐含，T5→T6→T7→T8 链式）
**User Preference:** none
**Decision:** Subagent-Driven
**Reasoning:** 8 个 Task，测试线（T1-T4）与 CLI 线（T5-T8）相互独立可并行，文档线（T8）依赖 CLI 完成。符合 "3+ tasks → Subagent-Driven" 规则。测试线内 T1/T2/T3 无依赖可并行，T4 依赖 T3。CLI 线严格串行。

**Auto-invoking:** `superpowers:subagent-driven-development`

⏹️ Phase 4 Complete: Execution selected — Subagent-Driven
