# 高级搜索

当预置的 `SearchBy*` 方法无法满足需求时，可以用 `SearchRequest` 构建器做更精细的控制。

## SearchRequest 构建器

```go
import (
    "github.com/scagogogo/sonatype-central-sdk/pkg/request"
    "github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

sr := request.NewSearchRequest().
    SetQuery(request.NewQuery().
        SetGroupId("org.apache.commons").
        SetArtifactId("commons-lang3")).
    SetRows(20).
    SetStart(0)

var result response.Response[*response.Artifact]
err := client.SearchRequest(ctx, sr, &result)
```

## 精确匹配

Solr 默认对 GroupId 等字段做分词匹配（`org.apache.commons` 会被当成三个词）。启用精确匹配可以避免：

```go
sr := request.NewSearchRequest().
    SetQuery(request.NewQuery().SetGroupId("org.apache.commons")).
    SetExact(true)   // 精确匹配整个字符串
```

## 自定义返回字段

默认返回所有字段。如果只需要几个字段（比如只关心坐标），用 `SetFieldList` 减小响应体积：

```go
sr := request.NewSearchRequest().
    SetQuery(request.NewQuery().SetText("commons")).
    SetFieldList("id,g,a,latestVersion")   // fl 参数
```

## 查询解析器与字段权重

Solr 支持 `dismax` / `edismax` 查询解析器，可以为不同字段设置权重（boost），影响排序相关性：

```go
sr := request.NewSearchRequest().
    SetQuery(request.NewQuery().SetText("json parser")).
    SetDefType("edismax").                          // defType 参数
    SetQueryFields("text^20 g^5 a^10 c^3 fc^2")     // qf 参数
```

上面表示：在全文（`text`）命中权重最高（×20），其次是 artifactId（×10）、groupId（×5）。

## 拼写检查

获取搜索查询的拼写建议（"你是不是想找……"）：

```go
// 单独的便捷方法：搜索同时返回拼写建议
arts, suggestions, err := client.SearchWithSpellcheck(ctx, "commns-lang", 10, 5)
if len(suggestions) > 0 {
    fmt.Printf("你是否想找: %v?\n", suggestions)
}
```

也可通过 `SearchRequest` 配置：

```go
sr := request.NewSearchRequest().
    SetQuery(request.NewQuery().SetText("commns-lang")).
    SetSpellcheck(true).
    SetSpellcheckCount(5)
```

## 排序

```go
sr := request.NewSearchRequest().
    SetQuery(request.NewQuery().SetGroupId("org.apache.commons")).
    SetSort("timestamp desc")   // 按时间倒序
```

## 高级查询方法

SDK 也提供了一些直接的高级方法：

```go
// 带排序的搜索
arts, _ := client.SearchWithSort(ctx, query, "popularity desc", 10)

// 按日期范围搜索
arts, _ := client.SearchArtifactsByDateRange(ctx, start, end, 10)

// GAV 列表带排序
gavs, _ := client.SearchGAVsWithSort(ctx, "org.apache.commons", "timestamp desc", 20)

// 分页的 GAV 列表
gavs, total, _ := client.ListGAVsPaginated(ctx, "org.apache.commons", 0, 20)
```

## 高级搜索 API 参考

| 参数 | 对应 Solr 参数 | 说明 |
|------|---------------|------|
| `SetExact` | `core_impl=...` | 启用精确匹配 |
| `SetSpellcheck` / `SetSpellcheckCount` | `spellcheck` / `spellcheck.count` | 拼写检查 |
| `SetFieldList` | `fl` | 返回字段列表 |
| `SetDefType` | `defType` | 查询解析器 |
| `SetQueryFields` | `qf` | 查询字段权重 |
| `SetRows` / `SetStart` | `rows` / `start` | 分页 |
| `SetSort` | `sort` | 排序 |
