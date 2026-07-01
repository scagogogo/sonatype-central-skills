# 搜索 API 参考

`api.Client` 提供的搜索相关方法。所有方法签名为 `(ctx context.Context, ...) -> (result, error)`。

## 按坐标搜索

| 方法 | 说明 |
|------|------|
| `SearchByGroupId(ctx, groupId string, limit int) ([]*response.Artifact, error)` | 按 GroupId |
| `SearchByArtifactId(ctx, artifactId string, limit int) ([]*response.Artifact, error)` | 按 ArtifactId |
| `SearchByGroupAndArtifactId(ctx, groupId, artifactId string, limit int) ([]*response.Artifact, error)` | G+A 组合 |
| `SearchByGroupPattern(ctx, pattern string, limit int) ([]*response.Artifact, error)` | GroupId 通配 |

## GAV 坐标

| 方法 | 说明 |
|------|------|
| `GetGAVInfo(ctx, groupId, artifactId, version string) (*response.GAV, error)` | 精确 GAV 信息 |
| `ListGAVs(ctx, groupId, artifactId string) ([]*response.GAV, error)` | 列出所有 GAV |
| `ListGAVsPaginated(ctx, groupId, artifactId string, start, rows int) ([]*response.GAV, int, error)` | 分页 GAV |
| `SearchGAVsWithSort(ctx, groupId, sort string, limit int) ([]*response.GAV, error)` | 带排序的 GAV |

## 版本

| 方法 | 说明 |
|------|------|
| `ListVersions(ctx, groupId, artifactId string) ([]*response.Version, error)` | 所有版本 |
| `GetLatestVersion(ctx, groupId, artifactId string) (*response.Version, error)` | 最新版本 |
| `HasVersion(ctx, groupId, artifactId, version string) (bool, error)` | 是否存在该版本 |
| `CountVersions(ctx, groupId, artifactId string) (int, error)` | 版本数 |
| `FilterVersions(ctx, groupId, artifactId string, filter func(string) bool) ([]*response.Version, error)` | 过滤版本 |
| `CompareVersions(v1, v2 string) (int, error)` | 比较版本号 |
| `GetVersionInfo(ctx, groupId, artifactId, version string) (*response.Version, error)` | 版本详情 |
| `IteratorVersions(ctx, groupId, artifactId string) Iterator[*response.Version]` | 版本迭代器 |

## SHA1

| 方法 | 说明 |
|------|------|
| `SearchBySha1(ctx, sha1 string, limit int) ([]*response.Artifact, error)` | 精确 SHA1 |
| `SearchExactSha1(ctx, sha1 string) (*response.Artifact, error)` | 精确匹配单个 |
| `GetFirstBySha1(ctx, sha1 string) (*response.Artifact, error)` | 第一个匹配 |
| `ExistsSha1(ctx, sha1 string) (bool, error)` | 是否存在 |
| `CountBySha1(ctx, sha1 string) (int, error)` | 计数 |
| `SearchBySha1Prefix(ctx, prefix string, limit int) ([]*response.Artifact, error)` | 前缀匹配 |
| `IteratorBySha1(ctx, sha1 string) Iterator[*response.Artifact]` | 迭代器 |
| `IteratorBySha1Prefix(ctx, prefix string) Iterator[*response.Artifact]` | 前缀迭代器 |

## 类搜索

| 方法 | 说明 |
|------|------|
| `SearchByClassName(ctx, name string, limit int) ([]*response.Artifact, error)` | 简单类名 |
| `SearchByFullyQualifiedClassName(ctx, fqcn string, limit int) ([]*response.Artifact, error)` | 全限定类名 |
| `SearchByJavaPackage(ctx, pkg string, limit int) ([]*response.Artifact, error)` | Java 包 |
| `SearchByPackageAndClassName(ctx, pkg, name string, limit int) ([]*response.Artifact, error)` | 包+类名 |
| `SearchByClassSupertype(ctx, supertype string, limit int) ([]*response.Artifact, error)` | 父类查找子类 |
| `SearchInterfaceImplementations(ctx, iface string, limit int) ([]*response.Artifact, error)` | 接口实现 |
| `SearchClassesByMethod(ctx, method string, limit int) ([]*response.Artifact, error)` | 方法名查类 |
| `SearchClassesWithClassHierarchy(...)` | 类层次 |
| `SearchClassesWithHighlighting(...)` | 高亮 |

迭代器系列：`IteratorByClassName`、`IteratorByFullyQualifiedClassName`、`IteratorByJavaPackage`、`IteratorByPackageAndClassName`、`IteratorByClassHierarchy`、`IteratorByClassSupertype`、`IteratorByInterfaceImplementation`、`IteratorByMethod`。

## 标签

| 方法 | 说明 |
|------|------|
| `SearchByTag(ctx, tag string, limit int) ([]*response.Artifact, error)` | 单标签 |
| `SearchByMultipleTags(ctx, tags []string, limit int) ([]*response.Artifact, error)` | 多标签 |
| `SearchByTagPrefix(ctx, prefix string, limit int) ([]*response.Artifact, error)` | 前缀 |
| `SearchArtifactsByTag(...)` / `SearchArtifactsWithAllTags(...)` | 按标签查制品 |
| `GetMostUsedTags(ctx, limit int) ([]string, error)` | 最常用标签 |

## 分组

| 方法 | 说明 |
|------|------|
| `SearchSubgroups(ctx, groupId string) ([]string, error)` | 子分组 |
| `GetGroupInfo(ctx, groupId string) (*response.GroupInfo, error)` | 分组信息 |
| `GetGroupStatistics(ctx, groupId string) (*response.GroupStatistics, error)` | 统计 |
| `CompareTwoGroups(ctx, g1, g2 string) (*response.GroupComparison, error)` | 比较 |
| `GetPopularGroups(ctx, limit int) ([]*response.GroupInfo, error)` | 热门分组 |

## 制品

| 方法 | 说明 |
|------|------|
| `GetArtifactDetails(ctx, groupId, artifactId string) (*response.Artifact, error)` | 详情 |
| `GetArtifactMetadata(ctx, ...) (*response.ArtifactMetadata, error)` | 元数据 |
| `GetArtifactStats(ctx, ...) (*response.ArtifactStats, error)` | 统计 |
| `GetArtifactUsage(ctx, ...) (*response.ArtifactUsage, error)` | 使用情况 |
| `CompareArtifacts(ctx, ...) (*response.ArtifactComparison, error)` | 比较 |
| `SearchPopularArtifacts(ctx, limit int) ([]*response.Artifact, error)` | 热门 |
| `SuggestSimilarArtifacts(ctx, ...) ([]*response.Artifact, error)` | 相似推荐 |

## 全文与其他

| 方法 | 说明 |
|------|------|
| `SearchByText(ctx, text string, limit int) ([]*response.Artifact, error)` | 全文 |
| `SearchByClassifier(ctx, classifier string, limit int) ([]*response.Artifact, error)` | 分类器 |
| `SearchByGroupAndClassifier(ctx, groupId, classifier string, limit int) ([]*response.Artifact, error)` | G+分类器 |
| `SearchWithSpellcheck(ctx, text string, limit, count int) ([]*response.Artifact, []string, error)` | 拼写检查 |
| `SearchWithSort(ctx, query, sort string, limit int) ([]*response.Artifact, error)` | 带排序 |
| `AdvancedSearch(ctx, req *request.SearchRequest) ([]*response.Artifact, error)` | 高级 |

## 自定义查询

```go
sr := request.NewSearchRequest().
    SetQuery(request.NewQuery().SetGroupId("org.apache.commons")).
    SetRows(10)

var result response.Response[*response.Artifact]
err := client.SearchRequest(ctx, sr, &result)
```

详见 [高级搜索](../guide/advanced-search)。
