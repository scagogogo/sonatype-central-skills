package api

import (
	"context"
	"errors"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// SearchByClassifier 按分类器搜索制品
//
// 使用 Solr 的 l: 字段搜索指定分类器的制品。
// 常见的分类器包括 sources、javadoc、tests 等。
//
// 参数:
//   - ctx: 上下文对象
//   - classifier: 分类器名称，如 "sources"、"javadoc"、"tests"
//   - limit: 返回结果的最大数量
//
// 返回:
//   - []*response.Artifact: 搜索结果列表
//   - error: 搜索过程中的错误
func (c *Client) SearchByClassifier(ctx context.Context, classifier string, limit int) ([]*response.Artifact, error) {
	searchRequest := request.NewSearchRequest().
		SetLimit(limit).
		SetQuery(request.NewQuery().SetClassifier(classifier))

	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, searchRequest)
	if err != nil {
		return nil, err
	}
	if result == nil || result.ResponseBody == nil {
		return nil, errors.New("empty response body")
	}
	return result.ResponseBody.Docs, nil
}

// IteratorByClassifier 按分类器搜索制品的迭代器
//
// 返回一个内存高效的迭代器，适用于大量结果的遍历。
//
// 参数:
//   - ctx: 上下文对象
//   - classifier: 分类器名称
func (c *Client) IteratorByClassifier(ctx context.Context, classifier string) *SearchIterator[*response.Artifact] {
	searchRequest := request.NewSearchRequest().
		SetQuery(request.NewQuery().SetClassifier(classifier))

	return NewSearchIterator[*response.Artifact](searchRequest).WithClient(c)
}

// SearchByGroupAndClassifier 按 GroupId 和分类器搜索制品
//
// 组合使用 g: 和 l: 字段进行搜索。
//
// 参数:
//   - ctx: 上下文对象
//   - groupId: Maven 组 ID
//   - classifier: 分类器名称
//   - limit: 返回结果的最大数量
//
// 返回:
//   - []*response.Artifact: 搜索结果列表
//   - error: 搜索过程中的错误
func (c *Client) SearchByGroupAndClassifier(ctx context.Context, groupId, classifier string, limit int) ([]*response.Artifact, error) {
	searchRequest := request.NewSearchRequest().
		SetLimit(limit).
		SetQuery(request.NewQuery().SetGroupId(groupId).SetClassifier(classifier))

	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, searchRequest)
	if err != nil {
		return nil, err
	}
	if result == nil || result.ResponseBody == nil {
		return nil, errors.New("empty response body")
	}
	return result.ResponseBody.Docs, nil
}

// SearchByText 全文搜索
//
// 使用 Solr 的 dismax 查询解析器进行全文搜索，
// 支持跨多个字段（text、g、a）的加权搜索。
//
// 参数:
//   - ctx: 上下文对象
//   - text: 搜索文本
//   - limit: 返回结果的最大数量
//
// 返回:
//   - []*response.Artifact: 搜索结果列表
//   - error: 搜索过程中的错误
func (c *Client) SearchByText(ctx context.Context, text string, limit int) ([]*response.Artifact, error) {
	searchRequest := request.NewSearchRequest().
		SetLimit(limit).
		SetQuery(request.NewQuery().SetText(text))

	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, searchRequest)
	if err != nil {
		return nil, err
	}
	if result == nil || result.ResponseBody == nil {
		return nil, errors.New("empty response body")
	}
	return result.ResponseBody.Docs, nil
}

// IteratorByText 全文搜索的迭代器
//
// 参数:
//   - ctx: 上下文对象
//   - text: 搜索文本
func (c *Client) IteratorByText(ctx context.Context, text string) *SearchIterator[*response.Artifact] {
	searchRequest := request.NewSearchRequest().
		SetQuery(request.NewQuery().SetText(text))

	return NewSearchIterator[*response.Artifact](searchRequest).WithClient(c)
}

// SearchWithSpellcheck 执行搜索并返回拼写建议
//
// 执行搜索请求的同时启用拼写检查，如果搜索关键词可能有拼写错误，
// 响应中会包含建议的正确拼写。
//
// 参数:
//   - ctx: 上下文对象
//   - text: 搜索文本
//   - limit: 返回结果的最大数量
//   - spellcheckCount: 拼写建议数量（通常为 5）
//
// 返回:
//   - []*response.Artifact: 搜索结果列表
//   - []string: 拼写建议列表
//   - error: 搜索过程中的错误
func (c *Client) SearchWithSpellcheck(ctx context.Context, text string, limit, spellcheckCount int) ([]*response.Artifact, []string, error) {
	searchRequest := request.NewSearchRequest().
		SetLimit(limit).
		SetQuery(request.NewQuery().SetText(text)).
		SetSpellcheck(true, spellcheckCount)

	result, err := SearchRequestJsonDoc[*response.Artifact](c, ctx, searchRequest)
	if err != nil {
		return nil, nil, err
	}

	var suggestions []string
	if result.Spellcheck != nil {
		suggestions = result.Spellcheck.GetSuggestions()
	}

	var artifacts []*response.Artifact
	if result.ResponseBody != nil {
		artifacts = result.ResponseBody.Docs
	}

	return artifacts, suggestions, nil
}
