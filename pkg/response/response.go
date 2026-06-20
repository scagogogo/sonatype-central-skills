package response

type Response[Doc any] struct {
	ResponseHeader *ResponseHeader    `json:"responseHeader"`
	ResponseBody   *ResponseBody[Doc] `json:"response"`
	FacetCounts    *FacetCounts       `json:"facet_counts,omitempty"`

	// Highlighting 包含搜索结果的高亮信息
	// 数据结构为三层嵌套映射：
	// 1. 第一层key: 文档ID (如 "org.apache:commons-io:1.2.3")
	// 2. 第二层key: 高亮字段名 (如 "fch"表示fully qualified class name)
	// 3. 值: 高亮片段数组，其中匹配部分用<em>标签包围
	//
	// 示例:
	// {
	//   "org.apache:commons-io:1.2.3": {
	//     "fch": ["<em>org.apache</em>.commons.io.FileUtils"]
	//   }
	// }
	Highlighting map[string]map[string][]string `json:"highlighting,omitempty"`

	// Spellcheck 包含拼写检查的建议信息
	Spellcheck *SpellcheckResponse `json:"spellcheck,omitempty"`
}

type ResponseHeader struct {
	Status int     `json:"status"`
	QTime  int     `json:"QTime"`
	Params *Params `json:"params"`
}

type Params struct {
	Q       string `json:"q"`
	Core    string `json:"core"`
	Indent  string `json:"indent"`
	Fl      string `json:"fl"`
	Start   string `json:"start"`
	Sort    string `json:"sort"`
	Rows    string `json:"rows"`
	Wt      string `json:"wt"`
	Version string `json:"version"`
}

type ResponseBody[Doc any] struct {
	NumFound int   `json:"numFound"`
	Start    int   `json:"start"`
	Docs     []Doc `json:"docs"`
}

// FacetCounts 表示聚合查询结果
type FacetCounts struct {
	FacetFields  map[string][]interface{} `json:"facet_fields,omitempty"`
	FacetQueries map[string]int           `json:"facet_queries,omitempty"`
	FacetDates   map[string]interface{}   `json:"facet_dates,omitempty"`
}

// SpellcheckResponse 表示拼写检查的响应
//
// Sonatype Central 的搜索 API 默认启用拼写检查功能。
// 当搜索关键词可能存在拼写错误时，API 会返回建议的正确拼写。
//
// 响应结构为 Solr 的 spellcheck 格式：
// "spellcheck": {
//   "suggestions": ["commns", {"numFound": 5, "startOffset": 0, "endOffset": 6, "suggestion": ["commons", "communs", ...]}]
// }
type SpellcheckResponse struct {
	Suggestions []interface{} `json:"suggestions"`
}

// GetSuggestions 从拼写检查响应中提取所有建议词
//
// 解析 Solr 的交替数组格式 [word, {suggestion: [...]}, word2, {suggestion: [...]}]，
// 返回所有建议词的合并列表。
func (s *SpellcheckResponse) GetSuggestions() []string {
	if s == nil || len(s.Suggestions) == 0 {
		return nil
	}

	var suggestions []string
	for i := 0; i < len(s.Suggestions); i++ {
		// Solr spellcheck suggestions 是交替格式：[word, suggestionObject, word2, suggestionObject2, ...]
		// suggestionObject 是 map[string]interface{}，其中 "suggestion" 字段包含建议列表
		if suggestionMap, ok := s.Suggestions[i].(map[string]interface{}); ok {
			if suggestionList, ok := suggestionMap["suggestion"].([]interface{}); ok {
				for _, s := range suggestionList {
					if str, ok := s.(string); ok {
						suggestions = append(suggestions, str)
					}
				}
			}
		}
	}

	return suggestions
}
