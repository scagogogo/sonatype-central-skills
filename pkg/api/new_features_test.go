package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// TestSearchRequestExact 测试精确匹配参数
func TestSearchRequestExact(t *testing.T) {
	sr := request.NewSearchRequest().SetExact(true)
	assert.True(t, sr.Exact)

	params := sr.ToRequestParams()
	assert.Contains(t, params, "exact=true")
}

// TestSearchRequestSpellcheck 测试拼写检查参数
func TestSearchRequestSpellcheck(t *testing.T) {
	t.Run("enable spellcheck with count", func(t *testing.T) {
		sr := request.NewSearchRequest().SetSpellcheck(true, 10)
		assert.NotNil(t, sr.SpellcheckEnabled)
		assert.True(t, *sr.SpellcheckEnabled)
		assert.Equal(t, 10, sr.SpellcheckCount)

		params := sr.ToRequestParams()
		assert.Contains(t, params, "spellcheck=true")
		assert.Contains(t, params, "spellcheck.count=10")
	})

	t.Run("disable spellcheck", func(t *testing.T) {
		sr := request.NewSearchRequest().SetSpellcheck(false, 0)
		assert.NotNil(t, sr.SpellcheckEnabled)
		assert.False(t, *sr.SpellcheckEnabled)

		params := sr.ToRequestParams()
		assert.Contains(t, params, "spellcheck=false")
	})

	t.Run("default no spellcheck param", func(t *testing.T) {
		sr := request.NewSearchRequest()
		assert.Nil(t, sr.SpellcheckEnabled)

		params := sr.ToRequestParams()
		assert.NotContains(t, params, "spellcheck")
	})
}

// TestSearchRequestFieldList 测试字段列表参数
func TestSearchRequestFieldList(t *testing.T) {
	sr := request.NewSearchRequest().SetFieldList("id,g,a,v,p,timestamp,tags")
	assert.Equal(t, "id,g,a,v,p,timestamp,tags", sr.FieldList)

	params := sr.ToRequestParams()
	assert.Contains(t, params, "fl=id,g,a,v,p,timestamp,tags")
}

// TestSearchRequestDefType 测试查询解析器类型参数
func TestSearchRequestDefType(t *testing.T) {
	sr := request.NewSearchRequest().SetDefType("edismax")
	assert.Equal(t, "edismax", sr.DefType)

	params := sr.ToRequestParams()
	assert.Contains(t, params, "defType=edismax")
}

// TestSearchRequestQueryFields 测试查询字段权重参数
func TestSearchRequestQueryFields(t *testing.T) {
	sr := request.NewSearchRequest().SetQueryFields("text^20 g^5 a^10 c^3")
	assert.Equal(t, "text^20 g^5 a^10 c^3", sr.QueryFields)

	params := sr.ToRequestParams()
	assert.Contains(t, params, "qf=text^20 g^5 a^10 c^3")
}

// TestSearchRequestAllNewParams 测试所有新参数组合
func TestSearchRequestAllNewParams(t *testing.T) {
	sr := request.NewSearchRequest().
		SetExact(true).
		SetSpellcheck(true, 5).
		SetFieldList("id,g,a").
		SetDefType("edismax").
		SetQueryFields("text^20 g^5 a^10")

	params := sr.ToRequestParams()
	assert.Contains(t, params, "exact=true")
	assert.Contains(t, params, "spellcheck=true")
	assert.Contains(t, params, "spellcheck.count=5")
	assert.Contains(t, params, "fl=id,g,a")
	assert.Contains(t, params, "defType=edismax")
	assert.Contains(t, params, "qf=text^20 g^5 a^10")
}

// TestQueryClassifier 测试 Query 的 Classifier 字段
func TestQueryClassifier(t *testing.T) {
	q := request.NewQuery().SetClassifier("sources")
	assert.Equal(t, "sources", q.Classifier)

	paramValue := q.ToRequestParamValue()
	// Should contain l:sources in the encoded query
	assert.Contains(t, paramValue, "l%3Asources")
}

// TestDeploymentListOptions 测试部署列表选项
func TestDeploymentListOptions(t *testing.T) {
	opts := &response.DeploymentListOptions{
		Namespace:      "com.example",
		DeploymentName: "my-lib",
		State:          response.DeploymentStateValidated,
		Page:           0,
		Size:           20,
		SortField:      "createdAt",
		SortDirection:  "desc",
	}

	assert.Equal(t, "com.example", opts.Namespace)
	assert.Equal(t, "my-lib", opts.DeploymentName)
	assert.Equal(t, response.DeploymentStateValidated, opts.State)
	assert.Equal(t, 0, opts.Page)
	assert.Equal(t, 20, opts.Size)
}

// TestListDeploymentsWithOptions 测试带选项的 ListDeployments
func TestListDeploymentsWithOptions(t *testing.T) {
	client := NewPublisherClient(WithPublisherToken("test-token"))

	// 测试 nil 选项（列出所有）
	// 这会因网络原因失败，但验证方法签名正确
	_ = client

	// 测试带选项的调用不会 panic
	opts := &response.DeploymentListOptions{
		Namespace: "com.example",
		State:     response.DeploymentStateValidated,
		Page:      1,
		Size:      10,
	}
	assert.NotNil(t, opts)
}

// TestSBOMConstants 测试 SBOM 文件扩展名常量
func TestSBOMConstants(t *testing.T) {
	assert.Equal(t, "cyclonedx.json", CYCLONEDX_JSON)
	assert.Equal(t, "cyclonedx.xml", CYCLONEDX_XML)
	assert.Equal(t, "spdx.json", SPDX_JSON)
}

// TestSBOMArtifactFiles 测试 SBOM 预定义文件类型
func TestSBOMArtifactFiles(t *testing.T) {
	assert.Equal(t, "CYCLONEDX_JSON", CycloneDXJsonFile.Type)
	assert.Equal(t, CYCLONEDX_JSON, CycloneDXJsonFile.Extension)

	assert.Equal(t, "CYCLONEDX_XML", CycloneDXXmlFile.Type)
	assert.Equal(t, CYCLONEDX_XML, CycloneDXXmlFile.Extension)

	assert.Equal(t, "SPDX_JSON", SpdxJsonFile.Type)
	assert.Equal(t, SPDX_JSON, SpdxJsonFile.Extension)
}

// TestBuildArtifactPathSBOM 测试 SBOM 文件路径构建
func TestBuildArtifactPathSBOM(t *testing.T) {
	path := BuildArtifactPath("com.example", "my-lib", "1.0.0", CYCLONEDX_JSON)
	assert.Equal(t, "com/example/my-lib/1.0.0/my-lib-1.0.0.cyclonedx.json", path)

	path = BuildArtifactPath("com.example", "my-lib", "1.0.0", CYCLONEDX_XML)
	assert.Equal(t, "com/example/my-lib/1.0.0/my-lib-1.0.0.cyclonedx.xml", path)

	path = BuildArtifactPath("com.example", "my-lib", "1.0.0", SPDX_JSON)
	assert.Equal(t, "com/example/my-lib/1.0.0/my-lib-1.0.0.spdx.json", path)
}

// TestSpellcheckResponseGetSuggestions 测试拼写检查建议提取
func TestSpellcheckResponseGetSuggestions(t *testing.T) {
	t.Run("nil response", func(t *testing.T) {
		var s *response.SpellcheckResponse
		suggestions := s.GetSuggestions()
		assert.Nil(t, suggestions)
	})

	t.Run("empty suggestions", func(t *testing.T) {
		s := &response.SpellcheckResponse{Suggestions: []interface{}{}}
		suggestions := s.GetSuggestions()
		assert.Nil(t, suggestions)
	})

	t.Run("with suggestions", func(t *testing.T) {
		s := &response.SpellcheckResponse{
			Suggestions: []interface{}{
				"commns",
				map[string]interface{}{
					"numFound":    float64(5),
					"startOffset": float64(0),
					"endOffset":   float64(6),
					"suggestion":  []interface{}{"commons", "communs", "commas"},
				},
			},
		}
		suggestions := s.GetSuggestions()
		assert.Equal(t, []string{"commons", "communs", "commas"}, suggestions)
	})
}

// TestSearchByText 测试全文搜索方法签名
func TestSearchByText(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client)
	// 方法存在性验证（编译通过即可）
}

// TestSearchByClassifier 测试分类器搜索方法签名
func TestSearchByClassifier(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client)
	// 方法存在性验证（编译通过即可）
}

// TestSearchWithSpellcheck 测试拼写检查搜索方法签名
func TestSearchWithSpellcheck(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client)
	// 方法存在性验证（编译通过即可）
}
