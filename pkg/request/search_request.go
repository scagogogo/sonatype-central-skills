package request

import "fmt"

const SearchRequestLimitMax = 200

// SearchRequest 表示一次搜索请求
type SearchRequest struct {

	// 从第几条开始返回
	Start int

	// 最多返回200条，默认设置为200
	Limit int

	// 查询参数是啥
	Query *Query

	// Core参数
	Core string

	// 排序字段
	SortField string

	// 排序方向（升序/降序）
	SortAscending bool

	// 是否启用聚合
	FacetEnabled bool

	// 聚合字段
	FacetFields []string

	// 查询键，用于标识批量查询
	QueryKey string

	// 其他自定义参数
	CustomParams map[string]string

	// Exact 是否启用精确匹配模式
	// 设置为 true 时，搜索将进行精确匹配而非模糊匹配
	Exact bool

	// SpellcheckEnabled 是否启用拼写检查
	// 默认情况下 Solr 会自动启用拼写检查，可通过设置为 false 显式禁用
	SpellcheckEnabled *bool

	// SpellcheckCount 拼写检查返回的建议数量，默认为 5
	SpellcheckCount int

	// FieldList 指定返回的字段列表
	// 为空时使用 Solr 默认字段列表
	// 例如: "id,g,a,v,p,timestamp,tags"
	FieldList string

	// DefType 查询解析器类型
	// 默认为 "dismax"，可选值包括 "lucene"、"dismax"、"edismax" 等
	DefType string

	// QueryFields 查询字段及其权重
	// 默认为 "text^20 g^5 a^10"
	// 例如: "text^20 g^5 a^10 c^3 fc^2"
	QueryFields string
}

func NewSearchRequest() *SearchRequest {
	return &SearchRequest{
		Start:        0,
		Limit:        SearchRequestLimitMax,
		Query:        NewQuery(),
		CustomParams: make(map[string]string),
	}
}

func (x *SearchRequest) SetStart(start int) *SearchRequest {
	x.Start = start
	return x
}

func (x *SearchRequest) SetLimit(limit int) *SearchRequest {
	x.Limit = limit
	return x
}

func (x *SearchRequest) SetCore(core string) *SearchRequest {
	x.Core = core
	return x
}

func (x *SearchRequest) SetQuery(query *Query) *SearchRequest {
	x.Query = query
	return x
}

// SetSort 设置排序字段和方向
func (x *SearchRequest) SetSort(field string, ascending bool) *SearchRequest {
	x.SortField = field
	x.SortAscending = ascending
	return x
}

// EnableFacet 启用聚合查询
func (x *SearchRequest) EnableFacet(fields ...string) *SearchRequest {
	x.FacetEnabled = true
	x.FacetFields = fields
	return x
}

// SetQueryKey 设置查询键，用于标识批量查询
func (x *SearchRequest) SetQueryKey(key string) *SearchRequest {
	x.QueryKey = key
	return x
}

// GetQueryKey 获取查询键
func (x *SearchRequest) GetQueryKey() string {
	return x.QueryKey
}

// AddCustomParam 添加自定义参数
func (x *SearchRequest) AddCustomParam(key, value string) *SearchRequest {
	x.CustomParams[key] = value
	return x
}

// SetRows 设置返回结果的最大数量（与SetLimit相同，为兼容旧代码保留）
func (sr *SearchRequest) SetRows(rows int) *SearchRequest {
	return sr.SetLimit(rows)
}

// SetExact 设置是否启用精确匹配模式
//
// 启用后搜索将进行精确匹配而非模糊匹配。
// 适用于需要精确查找特定 groupId/artifactId 的场景。
//
// 参数:
//   - exact: true 启用精确匹配，false 禁用
func (x *SearchRequest) SetExact(exact bool) *SearchRequest {
	x.Exact = exact
	return x
}

// SetSpellcheck 设置拼写检查
//
// 参数:
//   - enabled: true 启用拼写检查，false 禁用
//   - count: 拼写建议数量，默认为 5
func (x *SearchRequest) SetSpellcheck(enabled bool, count int) *SearchRequest {
	x.SpellcheckEnabled = &enabled
	if count > 0 {
		x.SpellcheckCount = count
	}
	return x
}

// SetFieldList 设置返回的字段列表
//
// 为空时使用 Solr 默认字段列表。
// 可用于优化响应大小，只返回需要的字段。
//
// 参数:
//   - fields: 逗号分隔的字段列表，例如 "id,g,a,v,p,timestamp,tags"
func (x *SearchRequest) SetFieldList(fields string) *SearchRequest {
	x.FieldList = fields
	return x
}

// SetDefType 设置查询解析器类型
//
// 默认为 "dismax"。可选值包括 "lucene"、"dismax"、"edismax" 等。
// 不同的解析器支持不同的查询语法和功能。
//
// 参数:
//   - defType: 查询解析器类型
func (x *SearchRequest) SetDefType(defType string) *SearchRequest {
	x.DefType = defType
	return x
}

// SetQueryFields 设置查询字段及其权重
//
// 默认为 "text^20 g^5 a^10"。
// 可用于调整搜索结果的相关性排序。
//
// 参数:
//   - fields: 查询字段权重字符串，例如 "text^20 g^5 a^10 c^3 fc^2"
func (x *SearchRequest) SetQueryFields(fields string) *SearchRequest {
	x.QueryFields = fields
	return x
}

func (x *SearchRequest) ToRequestParams() string {
	params := fmt.Sprintf("q=%s&rows=%d&wt=json&start=%d", x.Query.ToRequestParamValue(), x.Limit, x.Start)

	// 添加Core参数
	if x.Core != "" {
		params += fmt.Sprintf("&core=%s", x.Core)
	}

	// 添加排序参数
	if x.SortField != "" {
		sortOrder := "asc"
		if !x.SortAscending {
			sortOrder = "desc"
		}
		params += fmt.Sprintf("&sort=%s+%s", x.SortField, sortOrder)
	}

	// 添加聚合参数
	if x.FacetEnabled {
		params += "&facet=true"

		if len(x.FacetFields) > 0 {
			for _, field := range x.FacetFields {
				params += fmt.Sprintf("&facet.field=%s", field)
			}
		}
	}

	// 添加精确匹配参数
	if x.Exact {
		params += "&exact=true"
	}

	// 添加拼写检查参数
	if x.SpellcheckEnabled != nil {
		if *x.SpellcheckEnabled {
			params += "&spellcheck=true"
			if x.SpellcheckCount > 0 {
				params += fmt.Sprintf("&spellcheck.count=%d", x.SpellcheckCount)
			}
		} else {
			params += "&spellcheck=false"
		}
	}

	// 添加字段列表参数
	if x.FieldList != "" {
		params += fmt.Sprintf("&fl=%s", x.FieldList)
	}

	// 添加查询解析器类型
	if x.DefType != "" {
		params += fmt.Sprintf("&defType=%s", x.DefType)
	}

	// 添加查询字段权重
	if x.QueryFields != "" {
		params += fmt.Sprintf("&qf=%s", x.QueryFields)
	}

	// 添加自定义参数
	for key, value := range x.CustomParams {
		params += fmt.Sprintf("&%s=%s", key, value)
	}

	return params
}
