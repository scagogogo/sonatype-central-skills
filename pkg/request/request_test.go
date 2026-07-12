package request

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// -----------------------------------------------------------------------------
// advanced_search.go
// -----------------------------------------------------------------------------

func TestNewAdvancedSearchOptions(t *testing.T) {
	got := NewAdvancedSearchOptions()
	assert.NotNil(t, got)
	assert.Equal(t, "", got.GroupId)
	assert.Equal(t, "", got.ArtifactId)
	assert.Equal(t, "", got.Version)
	assert.Equal(t, "", got.Packaging)
	assert.Equal(t, "", got.Classifier)
}

func TestAdvancedSearchOptions_Setters_Chain(t *testing.T) {
	// 验证链式调用与字段赋值
	opts := NewAdvancedSearchOptions().
		SetGroupId("com.example").
		SetArtifactId("my-lib").
		SetVersion("1.2.3").
		SetPackaging("jar").
		SetClassifier("sources")

	assert.Equal(t, "com.example", opts.GroupId)
	assert.Equal(t, "my-lib", opts.ArtifactId)
	assert.Equal(t, "1.2.3", opts.Version)
	assert.Equal(t, "jar", opts.Packaging)
	assert.Equal(t, "sources", opts.Classifier)

	// 链式 setter 返回值即自身
	assert.Same(t, opts, opts.SetGroupId("x"))
}

func TestMakeDependencyQuery_BothNonEmpty(t *testing.T) {
	got := MakeDependencyQuery("com.example", "my-lib")
	assert.Equal(t, "d:com.example:my-lib", got)
}

func TestMakeDependencyQuery_OnlyGroupId(t *testing.T) {
	got := MakeDependencyQuery("com.example", "")
	assert.Equal(t, "d:com.example", got)
}

func TestMakeDependencyQuery_OnlyArtifactId(t *testing.T) {
	got := MakeDependencyQuery("", "my-lib")
	assert.Equal(t, "d:*:my-lib", got)
}

func TestMakeDependencyQuery_BothEmpty(t *testing.T) {
	got := MakeDependencyQuery("", "")
	assert.Equal(t, "", got)
}

func TestMakeLicenseQuery(t *testing.T) {
	assert.Equal(t, "l:mit", MakeLicenseQuery("mit"))
	assert.Equal(t, "l:", MakeLicenseQuery(""))
	assert.Equal(t, "l:apache-2.0", MakeLicenseQuery("apache-2.0"))
}

// -----------------------------------------------------------------------------
// query.go
// -----------------------------------------------------------------------------

func TestNewQuery(t *testing.T) {
	q := NewQuery()
	assert.NotNil(t, q)
	assert.Equal(t, "", q.CustomQuery)
}

func TestQuery_Setters_Chain(t *testing.T) {
	q := NewQuery().
		SetGroupId("g").
		SetArtifactId("a").
		SetVersion("v").
		SetTags("t").
		SetSha1("s").
		SetClassName("c").
		SetFullyQualifiedClassName("fc").
		SetPackaging("p").
		SetClassifier("l")

	assert.Equal(t, "g", q.GroupId)
	assert.Equal(t, "a", q.ArtifactId)
	assert.Equal(t, "v", q.Version)
	assert.Equal(t, "t", q.Tags)
	assert.Equal(t, "s", q.Sha1)
	assert.Equal(t, "c", q.ClassName)
	assert.Equal(t, "fc", q.FullyQualifiedClassName)
	assert.Equal(t, "p", q.Packaging)
	assert.Equal(t, "l", q.Classifier)

	// SetText 与 SetCustomQuery 写入同一字段 CustomQuery
	q2 := NewQuery().SetText("some text")
	assert.Equal(t, "some text", q2.CustomQuery)
	q3 := NewQuery().SetCustomQuery("custom query")
	assert.Equal(t, "custom query", q3.CustomQuery)
}

func TestQuery_ToRequestParamValue_AllEmpty(t *testing.T) {
	// 全空：返回 url.QueryEscape("") == ""
	got := NewQuery().ToRequestParamValue()
	assert.Equal(t, "", got)
}

func TestQuery_ToRequestParamValue_CustomQuery_Branch(t *testing.T) {
	// CustomQuery 非空时直接返回 escape(CustomQuery)，跳过其他字段
	q := NewQuery().
		SetGroupId("should-be-ignored").
		SetArtifactId("also-ignored").
		SetCustomQuery("d:com.example:my-lib")
	got := q.ToRequestParamValue()
	assert.Equal(t, url.QueryEscape("d:com.example:my-lib"), got)
	// 确认没有拼接其他字段
	assert.NotContains(t, got, "should-be-ignored")
}

func TestQuery_ToRequestParamValue_SetText_Branch(t *testing.T) {
	// SetText 走的也是 CustomQuery 分支
	q := NewQuery().SetText("l:mit")
	assert.Equal(t, url.QueryEscape("l:mit"), q.ToRequestParamValue())
}

func TestQuery_ToRequestParamValue_SingleField(t *testing.T) {
	// 单字段：完整断言，避免多字段拼接干扰
	q := NewQuery().SetGroupId("g")
	assert.Equal(t, url.QueryEscape("g:g"), q.ToRequestParamValue())
}

func TestQuery_ToRequestParamValue_MultiField_Concat(t *testing.T) {
	// 多字段用 " AND " 连接后整体 QueryEscape
	q := NewQuery().
		SetGroupId("g").
		SetArtifactId("a")
	got := q.ToRequestParamValue()
	// "g:g AND a:a" 经 escape：空格->+，: 不编码
	assert.Equal(t, "g%3Ag+AND+a%3Aa", got)
}

func TestQuery_ToRequestParamValue_AllFields_EachPrefix(t *testing.T) {
	// 验证每个字段的前缀与拼接顺序
	q := NewQuery().
		SetGroupId("g").
		SetArtifactId("a").
		SetVersion("v").
		SetTags("t").
		SetSha1("s").
		SetClassName("c").
		SetFullyQualifiedClassName("fc").
		SetPackaging("p").
		SetClassifier("l")

	got := q.ToRequestParamValue()
	// 先解码，得到原始拼接串
	decoded, err := url.QueryUnescape(got)
	assert.NoError(t, err)
	assert.Equal(t, "g:g AND a:a AND v:v AND tags:t AND 1:s AND c:c AND fc:fc AND p:p AND l:l", decoded)

	// 同时逐项验证 escape 后子串存在
	assert.Contains(t, got, url.QueryEscape("g:g"))
	assert.Contains(t, got, url.QueryEscape("a:a"))
	assert.Contains(t, got, url.QueryEscape("v:v"))
	assert.Contains(t, got, url.QueryEscape("tags:t"))
	assert.Contains(t, got, url.QueryEscape("1:s"))
	assert.Contains(t, got, url.QueryEscape("c:c"))
	assert.Contains(t, got, url.QueryEscape("fc:fc"))
	assert.Contains(t, got, url.QueryEscape("p:p"))
	assert.Contains(t, got, url.QueryEscape("l:l"))
	// AND 连接符
	assert.True(t, strings.Contains(got, "+AND+"))
}

// -----------------------------------------------------------------------------
// search_request.go
// -----------------------------------------------------------------------------

func TestNewSearchRequest_Defaults(t *testing.T) {
	sr := NewSearchRequest()
	assert.NotNil(t, sr)
	assert.Equal(t, 0, sr.Start)
	assert.Equal(t, SearchRequestLimitMax, sr.Limit)
	assert.Equal(t, 200, sr.Limit)
	assert.NotNil(t, sr.Query)
	assert.NotNil(t, sr.CustomParams)
	assert.Equal(t, 0, len(sr.CustomParams))
}

func TestSearchRequest_Setters_Chain(t *testing.T) {
	q := NewQuery().SetGroupId("g")
	sr := NewSearchRequest().
		SetStart(10).
		SetLimit(50).
		SetCore("core1").
		SetQuery(q).
		SetQueryKey("batch-1")

	assert.Equal(t, 10, sr.Start)
	assert.Equal(t, 50, sr.Limit)
	assert.Equal(t, "core1", sr.Core)
	assert.Same(t, q, sr.Query)
	assert.Equal(t, "batch-1", sr.QueryKey)

	// SetRows 等价于 SetLimit
	sr.SetRows(77)
	assert.Equal(t, 77, sr.Limit)
}

func TestSearchRequest_GetQueryKey(t *testing.T) {
	sr := NewSearchRequest()
	assert.Equal(t, "", sr.GetQueryKey())
	sr.SetQueryKey("k1")
	assert.Equal(t, "k1", sr.GetQueryKey())
}

func TestSearchRequest_AddCustomParam(t *testing.T) {
	sr := NewSearchRequest().
		AddCustomParam("k1", "v1").
		AddCustomParam("k2", "v2")
	assert.Equal(t, "v1", sr.CustomParams["k1"])
	assert.Equal(t, "v2", sr.CustomParams["k2"])
	assert.Equal(t, 2, len(sr.CustomParams))
}

func TestSearchRequest_SetSort(t *testing.T) {
	sr := NewSearchRequest()
	sr.SetSort("timestamp", true)
	assert.Equal(t, "timestamp", sr.SortField)
	assert.True(t, sr.SortAscending)

	sr.SetSort("timestamp", false)
	assert.False(t, sr.SortAscending)
}

func TestSearchRequest_EnableFacet(t *testing.T) {
	sr := NewSearchRequest().EnableFacet("f1", "f2")
	assert.True(t, sr.FacetEnabled)
	assert.Equal(t, []string{"f1", "f2"}, sr.FacetFields)

	// 无参数：变参传 0 个得到 nil slice，长度为 0
	sr2 := NewSearchRequest().EnableFacet()
	assert.True(t, sr2.FacetEnabled)
	assert.Equal(t, 0, len(sr2.FacetFields))
}

func TestSearchRequest_SetExact(t *testing.T) {
	sr := NewSearchRequest()
	assert.False(t, sr.Exact)
	sr.SetExact(true)
	assert.True(t, sr.Exact)
	sr.SetExact(false)
	assert.False(t, sr.Exact)
}

func TestSearchRequest_SetSpellcheck_TrueWithCount(t *testing.T) {
	sr := NewSearchRequest().SetSpellcheck(true, 5)
	assert.NotNil(t, sr.SpellcheckEnabled)
	assert.True(t, *sr.SpellcheckEnabled)
	assert.Equal(t, 5, sr.SpellcheckCount)
}

func TestSearchRequest_SetSpellcheck_TrueZeroCount(t *testing.T) {
	// count<=0 时不设 SpellcheckCount
	sr := NewSearchRequest().SetSpellcheck(true, 0)
	assert.NotNil(t, sr.SpellcheckEnabled)
	assert.True(t, *sr.SpellcheckEnabled)
	assert.Equal(t, 0, sr.SpellcheckCount)
}

func TestSearchRequest_SetSpellcheck_False(t *testing.T) {
	// false 分支：SpellcheckEnabled 指向 false，count<=0 不设
	sr := NewSearchRequest().SetSpellcheck(false, 0)
	assert.NotNil(t, sr.SpellcheckEnabled)
	assert.False(t, *sr.SpellcheckEnabled)
	assert.Equal(t, 0, sr.SpellcheckCount)
}

func TestSearchRequest_SetFieldList_SetDefType_SetQueryFields(t *testing.T) {
	sr := NewSearchRequest().
		SetFieldList("id,g,a,v").
		SetDefType("dismax").
		SetQueryFields("text^20 g^5 a^10")
	assert.Equal(t, "id,g,a,v", sr.FieldList)
	assert.Equal(t, "dismax", sr.DefType)
	assert.Equal(t, "text^20 g^5 a^10", sr.QueryFields)
}

// ---- ToRequestParams 各分支 ----

func TestSearchRequest_ToRequestParams_Basic(t *testing.T) {
	// 最小请求：仅 q/rows/wt/start，Query 为空 -> q 为空
	sr := NewSearchRequest() // Start=0, Limit=200, Query 空
	got := sr.ToRequestParams()
	assert.Equal(t, "q=&rows=200&wt=json&start=0", got)
}

func TestSearchRequest_ToRequestParams_WithQuery(t *testing.T) {
	sr := NewSearchRequest().SetLimit(50).SetStart(10)
	sr.Query.SetGroupId("g")
	got := sr.ToRequestParams()
	assert.Equal(t, "q="+url.QueryEscape("g:g")+"&rows=50&wt=json&start=10", got)
}

func TestSearchRequest_ToRequestParams_Core(t *testing.T) {
	sr := NewSearchRequest().SetCore("my-core")
	got := sr.ToRequestParams()
	assert.Contains(t, got, "&core=my-core")
}

func TestSearchRequest_ToRequestParams_Sort_Asc(t *testing.T) {
	sr := NewSearchRequest().SetSort("timestamp", true)
	got := sr.ToRequestParams()
	assert.Contains(t, got, "&sort=timestamp+asc")
}

func TestSearchRequest_ToRequestParams_Sort_Desc(t *testing.T) {
	sr := NewSearchRequest().SetSort("timestamp", false)
	got := sr.ToRequestParams()
	assert.Contains(t, got, "&sort=timestamp+desc")
}

func TestSearchRequest_ToRequestParams_Facet_WithFields(t *testing.T) {
	sr := NewSearchRequest().EnableFacet("f1", "f2")
	got := sr.ToRequestParams()
	assert.Contains(t, got, "&facet=true")
	assert.Contains(t, got, "&facet.field=f1")
	assert.Contains(t, got, "&facet.field=f2")
}

func TestSearchRequest_ToRequestParams_Facet_NoFields(t *testing.T) {
	// FacetEnabled=true 但无字段：只追加 &facet=true
	sr := NewSearchRequest().EnableFacet()
	got := sr.ToRequestParams()
	assert.Contains(t, got, "&facet=true")
	assert.NotContains(t, got, "&facet.field=")
}

func TestSearchRequest_ToRequestParams_Exact(t *testing.T) {
	sr := NewSearchRequest().SetExact(true)
	got := sr.ToRequestParams()
	assert.Contains(t, got, "&exact=true")
}

func TestSearchRequest_ToRequestParams_Spellcheck_True_WithCount(t *testing.T) {
	sr := NewSearchRequest().SetSpellcheck(true, 5)
	got := sr.ToRequestParams()
	assert.Contains(t, got, "&spellcheck=true")
	assert.Contains(t, got, "&spellcheck.count=5")
}

func TestSearchRequest_ToRequestParams_Spellcheck_True_NoCount(t *testing.T) {
	// count<=0 时不追加 spellcheck.count
	sr := NewSearchRequest().SetSpellcheck(true, 0)
	got := sr.ToRequestParams()
	assert.Contains(t, got, "&spellcheck=true")
	assert.NotContains(t, got, "&spellcheck.count=")
}

func TestSearchRequest_ToRequestParams_Spellcheck_False(t *testing.T) {
	sr := NewSearchRequest().SetSpellcheck(false, 0)
	got := sr.ToRequestParams()
	assert.Contains(t, got, "&spellcheck=false")
	assert.NotContains(t, got, "&spellcheck=true")
}

func TestSearchRequest_ToRequestParams_FieldList(t *testing.T) {
	sr := NewSearchRequest().SetFieldList("id,g,a,v")
	got := sr.ToRequestParams()
	assert.Contains(t, got, "&fl=id,g,a,v")
}

func TestSearchRequest_ToRequestParams_DefType(t *testing.T) {
	sr := NewSearchRequest().SetDefType("dismax")
	got := sr.ToRequestParams()
	assert.Contains(t, got, "&defType=dismax")
}

func TestSearchRequest_ToRequestParams_QueryFields(t *testing.T) {
	sr := NewSearchRequest().SetQueryFields("text^20 g^5 a^10")
	got := sr.ToRequestParams()
	assert.Contains(t, got, "&qf=text^20 g^5 a^10")
}

func TestSearchRequest_ToRequestParams_CustomParams(t *testing.T) {
	sr := NewSearchRequest().
		AddCustomParam("ck1", "cv1").
		AddCustomParam("ck2", "cv2")
	got := sr.ToRequestParams()
	assert.Contains(t, got, "&ck1=cv1")
	assert.Contains(t, got, "&ck2=cv2")
}

func TestSearchRequest_ToRequestParams_AllConditions(t *testing.T) {
	// 一次性打开所有条件分支，确保全部追加
	sr := NewSearchRequest().
		SetCore("c").
		SetSort("timestamp", false).
		EnableFacet("f1", "f2").
		SetExact(true).
		SetSpellcheck(true, 5).
		SetFieldList("id,g").
		SetDefType("edismax").
		SetQueryFields("text^20").
		AddCustomParam("ck", "cv")
	sr.Query.SetGroupId("g")

	got := sr.ToRequestParams()
	// 基础段
	assert.Contains(t, got, "q="+url.QueryEscape("g:g"))
	assert.Contains(t, got, "&rows=200")
	assert.Contains(t, got, "&wt=json")
	assert.Contains(t, got, "&start=0")
	// 各追加段
	assert.Contains(t, got, "&core=c")
	assert.Contains(t, got, "&sort=timestamp+desc")
	assert.Contains(t, got, "&facet=true")
	assert.Contains(t, got, "&facet.field=f1")
	assert.Contains(t, got, "&facet.field=f2")
	assert.Contains(t, got, "&exact=true")
	assert.Contains(t, got, "&spellcheck=true")
	assert.Contains(t, got, "&spellcheck.count=5")
	assert.Contains(t, got, "&fl=id,g")
	assert.Contains(t, got, "&defType=edismax")
	assert.Contains(t, got, "&qf=text^20")
	assert.Contains(t, got, "&ck=cv")
}
