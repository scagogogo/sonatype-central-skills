# 搜索

搜索是 SDK 最核心的能力。Maven Central 的搜索基于 Solr，SDK 把它封装成一组语义化方法。

## 基础客户端

```go
client := api.NewClient()
ctx := context.Background()
```

搜索与下载**无需认证**。

## 按坐标搜索

最常见的搜索方式——按 GroupId / ArtifactId 组合：

```go
// 按 GroupId
arts, _ := client.SearchByGroupId(ctx, "org.apache.commons", 10)

// 按 GroupId + ArtifactId
arts, _ := client.SearchByGroupAndArtifactId(ctx, "org.apache.commons", "commons-lang3", 10)

// 按 ArtifactId
arts, _ := client.SearchByArtifactId(ctx, "commons-lang3", 10)
```

## 按 GAV 坐标

GAV = GroupId + ArtifactId + Version。精确锁定某个版本：

```go
gav, _ := client.GetGAVInfo(ctx, "org.apache.commons", "commons-lang3", "3.12.0")
```

列出某制品的所有版本、判断是否存在某个版本、获取最新版本：

```go
versions, _ := client.ListVersions(ctx, "org.apache.commons", "commons-lang3")
has, _ := client.HasVersion(ctx, "org.apache.commons", "commons-lang3", "3.12.0")
latest, _ := client.GetLatestVersion(ctx, "org.apache.commons", "commons-lang3")
```

## 按 SHA1 搜索

当你只有一个文件的 SHA1 哈希时，可以用它反查制品坐标：

```go
art, _ := client.SearchBySha1(ctx, "a5b5...")      // 精确匹配
exists, _ := client.ExistsSha1(ctx, "a5b5...")     // 是否存在
arts, _ := client.SearchBySha1Prefix(ctx, "a5b5")   // 前缀匹配
```

## 按类名搜索

适用于"我有一个类名，想知道它来自哪个 JAR"的场景：

```go
// 简单类名
arts, _ := client.SearchByClassName(ctx, "StringUtils", 10)

// 全限定类名
arts, _ := client.SearchByFullyQualifiedClassName(ctx,
    "org.apache.commons.lang3.StringUtils", 10)

// 按 Java 包名
arts, _ := client.SearchByJavaPackage(ctx, "org.apache.commons.lang3", 10)
```

更高级的类搜索：

```go
// 类层次结构（查找子类）
arts, _ := client.SearchByClassSupertype(ctx, "java.util.AbstractList", 10)

// 接口实现
arts, _ := client.SearchInterfaceImplementations(ctx, "java.util.List", 10)

// 按方法名查类
arts, _ := client.SearchClassesByMethod(ctx, "isEmpty", 10)
```

## 按标签搜索

```go
arts, _ := client.SearchByTag(ctx, "logging", 10)              // 单标签
arts, _ := client.SearchByMultipleTags(ctx,                     // 多标签（AND）
    []string{"logging", "async"}, 10)
arts, _ := client.SearchByTagPrefix(ctx, "log", 10)            // 前缀
```

## 按分类器搜索

分类器（classifier）区分同一制品的不同附件：`sources`、`javadoc`、`tests` 等：

```go
arts, _ := client.SearchByClassifier(ctx, "sources", 10)
arts, _ := client.SearchByGroupAndClassifier(ctx, "org.apache.commons", "javadoc", 10)
```

## 全文搜索

不指定字段，对所有内容做模糊搜索：

```go
arts, _ := client.SearchByText(ctx, "json parser", 10)
```

## 分组搜索

```go
groups, _ := client.SearchSubgroups(ctx, "org.apache")   // 子分组
info, _ := client.GetGroupStatistics(ctx, "org.apache.commons")  // 统计
```

## 迭代器：处理海量结果

当结果可能很多时，用迭代器避免一次性加载：

```go
iter := client.IteratorByGroupId(ctx, "org.apache.commons")
for iter.HasNext() {
    a, err := iter.Next()
    if err != nil { break }
    // 逐条处理
}
```

支持的迭代器覆盖绝大多数搜索维度：`IteratorByArtifactId`、`IteratorByClassName`、`IteratorBySha1`、`IteratorByText`、`IteratorVersions` 等。

## 自定义 Solr 查询

如果预置方法不满足需求，可以用 `SearchRequest` 构建器完全自定义：

```go
sr := request.NewSearchRequest().
    SetQuery(request.NewQuery().
        SetGroupId("org.apache.commons").
        SetClassifier("sources")).
    SetRows(20).
    SetStart(0)

var result response.Response[*response.Artifact]
err := client.SearchRequest(ctx, sr, &result)
```

更多高级用法（精确匹配、字段列表、查询解析器、字段权重、拼写检查）见 [高级搜索](./advanced-search)。
