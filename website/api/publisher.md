# 发布 API 参考

`api.PublisherClient` 提供的发布相关方法。**需要认证**（Bearer Token 或 Basic Auth）。

## 创建客户端

```go
// Bearer Token
client := api.NewPublisherClient(api.WithPublisherToken("token"))

// Basic Auth
client := api.NewPublisherClient(api.WithPublisherBasicAuth("user", "pass"))

// 自定义 base URL（默认 https://central.sonatype.com）
client := api.NewPublisherClient(
    api.WithPublisherToken("token"),
    api.WithPublisherBaseURL("https://custom.example.com"),
)

// 自定义 HTTP 客户端
client := api.NewPublisherClient(
    api.WithPublisherToken("token"),
    api.WithPublisherHTTPClient(customHTTPClient),
)
```

## 方法

### UploadBundle

```go
func (pc *PublisherClient) UploadBundle(ctx context.Context, bundle []byte,
    name string, publishingType response.PublishingType) (string, error)
```

上传部署包到 Maven Central。

- `bundle`：ZIP/JAR 文件内容
- `name`：部署名称（可选，为空用默认）
- `publishingType`：`PublishingTypeAutomatic`（自动发布）或 `PublishingTypeUserManaged`（手动发布）
- 返回部署 ID（`text/plain` 响应，已处理引号和空白）

### GetDeploymentStatus

```go
func (pc *PublisherClient) GetDeploymentStatus(ctx context.Context,
    deploymentID string) (*response.DeploymentStatus, error)
```

查询部署状态。返回 `DeploymentStatus`，包含 `DeploymentState`（PENDING/VALIDATING/VALIDATED/PUBLISHING/PUBLISHED/FAILED）。

### CheckPublished

```go
func (pc *PublisherClient) CheckPublished(ctx context.Context,
    groupID, artifactID, version string) (*response.PublishedCheck, error)
```

检查某坐标是否已在 Maven Central 发布。返回 `PublishedCheck{Published, Namespace, Name, Version}`。

### ListDeployments

```go
func (pc *PublisherClient) ListDeployments(ctx context.Context,
    options *response.DeploymentListOptions) (*response.DeploymentList, error)
```

列出部署，支持过滤和分页。`options` 为 `nil` 时列出所有。

`DeploymentListOptions` 字段：

| 字段 | 说明 |
|------|------|
| `Namespace` | 按 groupId 过滤 |
| `DeploymentName` | 按名称模糊搜索 |
| `State` | 按状态过滤 |
| `Page` / `Size` | 分页（仅 `Paginate=true` 时发送） |
| `Paginate` | 是否启用分页 |
| `SortField` / `SortDirection` | 排序 |

### BrowseDeployment

```go
func (pc *PublisherClient) BrowseDeployment(ctx context.Context,
    deploymentID string) (*response.DeploymentResponseFiles, error)
```

便捷方法，浏览单个部署的文件。

### BrowseDeploymentWithOptions

```go
func (pc *PublisherClient) BrowseDeploymentWithOptions(ctx context.Context,
    req *response.BrowseDeploymentRequest) ([]response.DeploymentResponseFiles, error)
```

完整选项的浏览。`SortField` 必填。

### DownloadDeploymentFile

```go
func (pc *PublisherClient) DownloadDeploymentFile(ctx context.Context,
    deploymentID, relativePath string) ([]byte, error)
```

从部署包下载指定文件（返回原始字节，非 JSON）。

### DropDeployment

```go
func (pc *PublisherClient) DropDeployment(ctx context.Context,
    deploymentID string) error
```

删除部署。仅在 `FAILED` 或 `VALIDATED` 状态下可调用。

### PublishDeployment

```go
func (pc *PublisherClient) PublishDeployment(ctx context.Context,
    deploymentID string) error
```

发布部署。仅在 `VALIDATED` 状态下可调用。

## 错误处理

HTTP ≥ 400 的错误响应解析为 `*response.PublisherErrorResponse`：

```go
type PublisherErrorResponse struct {
    HttpStatus  int         // HTTP 状态码
    ErrorCode   int         // 应用错误码
    Message     string      // 错误消息
    Explanation string      // 附加说明
    Data        interface{} // 附加数据
}

func (e *PublisherErrorResponse) Error() string
```

用法：

```go
err := publisher.PublishDeployment(ctx, id)
var apiErr *response.PublisherErrorResponse
if errors.As(err, &apiErr) {
    // 业务错误
} else {
    // 网络层错误
}
```

## 类型常量

### PublishingType

| 常量 | 值 |
|------|-----|
| `api.PublishingTypeUserManaged` | `"USER_MANAGED"` |
| `api.PublishingTypeAutomatic` | `"AUTOMATIC"` |

### DeploymentState

| 常量 | 值 |
|------|-----|
| `api.DeploymentStatePending` | `"PENDING"` |
| `api.DeploymentStateValidating` | `"VALIDATING"` |
| `api.DeploymentStateValidated` | `"VALIDATED"` |
| `api.DeploymentStatePublishing` | `"PUBLISHING"` |
| `api.DeploymentStatePublished` | `"PUBLISHED"` |
| `api.DeploymentStateFailed` | `"FAILED"` |
