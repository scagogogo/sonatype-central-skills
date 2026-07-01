# 响应类型参考

所有 API 响应类型定义在 `pkg/response` 包。本页列出核心类型。

## Artifact

搜索结果的主要类型：

```go
type Artifact struct {
    GroupId        string
    ArtifactId     string
    Version        string
    LatestVersion  string
    Timestamp      string
    // ...其他字段
}
```

## Version

```go
type Version struct {
    Version   string
    UpdateTime string
    // ...
}
```

## GAV

```go
type GAV struct {
    GroupId    string
    ArtifactId string
    Version    string
    // ...
}
```

## Publisher 相关

### DeploymentStatus

由 `GetDeploymentStatus` 返回：

```go
type DeploymentStatus struct {
    DeploymentID    string
    DeploymentName  string
    DeploymentState DeploymentState
    PublishingType  PublishingType
    Purls           []string
    Errors          interface{}   // API 可能返回对象或数组
    CreateTimestamp interface{}
    UpdateTimestamp interface{}
}
```

### PublishedCheck

由 `CheckPublished` 返回：

```go
type PublishedCheck struct {
    Published bool
    Namespace string   // 回填自请求
    Name      string   // 回填自请求
    Version   string   // 回填自请求
}
```

### DeploymentList / DeploymentListItem

由 `ListDeployments` 返回：

```go
type DeploymentList struct {
    Deployments      []DeploymentListItem
    Page             int
    PageSize         int
    PageCount        int
    TotalResultCount int
}

type DeploymentListItem struct {
    DeploymentID        string
    DeploymentName      string
    Namespace           string
    DeploymentState     DeploymentState
    CreateTimestamp     string
    UpdateTimestamp     string
    DeploymentComponents []DeploymentComponent
}
```

### DeploymentResponseFiles / DeploymentFile

由 `BrowseDeployment` / `BrowseDeploymentWithOptions` 返回：

```go
type DeploymentResponseFiles struct {
    DeploymentID             string
    DeploymentName           string
    DeploymentState          DeploymentState
    DeploymentType           string   // BUNDLE 或 SINGLE
    CreateTimestamp          interface{}
    Purls                    []string
    DeployedComponentVersions []DeployedComponentVersion
    DeploymentFiles          []DeploymentFile
}

type DeploymentFile struct {
    RelativePath  string
    FileName      string
    FileSize      int64
    FileTimestamp int64
}
```

::: tip 字段名对齐官方 schema
`DeploymentFile` 的字段名（`relativePath`/`fileName`/`fileSize`/`fileTimestamp`）严格对齐 Sonatype Central 的 OpenAPI 规范，早期版本的 `path`/`size`/`contentType` 已修正。
:::

### 请求类型

#### BrowseDeploymentRequest

```go
type BrowseDeploymentRequest struct {
    Page          int
    Size          int
    SortField     string         // 必填
    SortDirection string         // asc 或 desc
    DeploymentIds []string
    PathStarting  string
}
```

#### DeploymentListOptions

```go
type DeploymentListOptions struct {
    Namespace      string
    DeploymentName string
    State          DeploymentState
    Page           int
    Size           int
    Paginate       bool
    SortField      string
    SortDirection  string
}
```

### 错误类型

#### PublisherErrorResponse

```go
type PublisherErrorResponse struct {
    HttpStatus  int
    ErrorCode   int
    Message     string
    Explanation string
    Data        interface{}
}

func (e *PublisherErrorResponse) Error() string
// 输出: Publisher API 错误 [<httpStatus>/<errorCode>]: <message>
```

::: warning 注意名称
`PublisherErrorResponse` 与 Search API 使用的 `ErrorResponse`（在 `http_error_types.go`）是**不同**的类型。Publisher 场景请用 `PublisherErrorResponse`。
:::

## 已弃用类型

以下类型保留向后兼容，新代码不应使用：

| 类型 | 原因 |
|------|------|
| `PublisherUploadResponse` | `/upload` 实际返回 text/plain，非 JSON |
| `PublisherError` | 旧版错误类型，推荐用 `PublisherErrorResponse` |

详见 [已弃用的 API](../guide/deprecated)。
