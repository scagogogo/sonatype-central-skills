# 发布到 Maven Central

本章介绍如何用 SDK 把自己的制品发布到 Maven Central。

::: warning 前置准备
发布是**需要认证**的操作。开始前请确保：
1. 在 [central.sonatype.com](https://central.sonatype.com) 拥有账号
2. 你的命名空间（namespace，即 groupId）已通过验证
3. 已生成 **API Token**（在账户设置里创建）

获取 Token 后，把它当作密码使用，配合 Bearer 或 Basic 认证。
:::

## 创建发布客户端

```go
publisher := api.NewPublisherClient(
    api.WithPublisherToken("your-bearer-token"),
)
```

或使用 Basic Auth：

```go
publisher := api.NewPublisherClient(
    api.WithPublisherBasicAuth("username", "token-or-password"),
)
```

## 发布流程

Maven Central 的发布是一个**异步状态机**：

```
        UploadBundle
            │
            ▼
        ┌────────┐  ┌───────────┐  ┌───────────┐  ┌────────────┐  ┌──────────┐
        │PENDING │→│VALIDATING │→│VALIDATED  │→│PUBLISHING  │→│PUBLISHED │
        └────────┘  └───────────┘  └───────────┘  └────────────┘  └──────────┘
                          │             │
                          └─────────────┴──→ FAILED
```

### 1. 上传部署包

部署包（bundle）是一个 ZIP/JAR 文件，包含 POM、主制品、签名、校验和等，符合 [Maven Central 的要求](https://central.sonatype.org/publish/publish-portal-guide/)。

```go
bundle, _ := os.ReadFile("my-lib-bundle.zip")
deploymentID, err := publisher.UploadBundle(ctx, bundle, "my-lib",
    api.PublishingTypeAutomatic)
if err != nil {
    log.Fatalf("上传失败: %v", err)
}
fmt.Printf("部署 ID: %s\n", deploymentID)
```

`publishingType` 参数：
- `api.PublishingTypeAutomatic` —— 校验通过后自动发布
- `api.PublishingTypeUserManaged` —— 需要手动调用 `PublishDeployment` 发布

::: tip 关于响应
`/upload` 端点返回 `text/plain`（仅部署 ID 字符串），SDK 内部已处理可能的引号包裹和空白，直接返回干净的 ID。
:::

### 2. 轮询部署状态

上传后，Sonatype 会异步校验你的包。轮询状态直到变为 `VALIDATED` 或 `FAILED`：

```go
var status *response.DeploymentStatus
for {
    status, err = publisher.GetDeploymentStatus(ctx, deploymentID)
    if err != nil {
        log.Fatal(err)
    }
    if status.DeploymentState == api.DeploymentStateValidated ||
        status.DeploymentState == api.DeploymentStateFailed {
        break
    }
    time.Sleep(5 * time.Second)
}

if status.DeploymentState == api.DeploymentStateFailed {
    log.Fatalf("校验失败: %v", status.Errors)
}
```

状态常量：

| 常量 | 含义 |
|------|------|
| `DeploymentStatePending` | 已接收，等待校验 |
| `DeploymentStateValidating` | 校验中 |
| `DeploymentStateValidated` | 校验通过，可发布 |
| `DeploymentStatePublishing` | 发布中 |
| `DeploymentStatePublished` | 已发布 ✓ |
| `DeploymentStateFailed` | 校验失败 |

### 3. 发布（仅 USER_MANAGED 需要）

如果上传时选了 `USER_MANAGED`，校验通过后需手动触发发布：

```go
if status.DeploymentState == api.DeploymentStateValidated {
    if err := publisher.PublishDeployment(ctx, deploymentID); err != nil {
        log.Fatal(err)
    }
}
```

### 4. 删除部署（可选）

如果校验失败或你想放弃，可以删除部署（仅在 `FAILED` 或 `VALIDATED` 状态下）：

```go
_ = publisher.DropDeployment(ctx, deploymentID)
```

## 其他实用方法

### 检查是否已发布

不经过部署流程，直接查询某个坐标是否已在 Maven Central 发布：

```go
check, _ := publisher.CheckPublished(ctx, "com.example", "my-lib", "1.0.0")
fmt.Printf("已发布: %v\n", check.Published)
```

### 列出部署

列出你的所有部署，支持过滤和分页：

```go
list, _ := publisher.ListDeployments(ctx, &response.DeploymentListOptions{
    Namespace: "com.example",
    State:     response.DeploymentStateValidated,
    Paginate:  true,
    Page:      0,
    Size:      20,
})
for _, d := range list.Deployments {
    fmt.Printf("%s [%s]\n", d.DeploymentName, d.DeploymentState)
}
```

::: tip Paginate 标志
`Page` 和 `Size` 字段只有当 `Paginate: true` 时才会发送给 API。这是为了区分"不传分页参数（API 用默认值）"和"传 page=0"两种语义——否则零值无法区分。
:::

### 浏览部署文件

查看部署包里包含哪些文件：

```go
// 便捷方法：单个部署
files, _ := publisher.BrowseDeployment(ctx, deploymentID)
for _, f := range files.DeploymentFiles {
    fmt.Printf("%s (%d 字节)\n", f.RelativePath, f.FileSize)
}

// 完整选项：多部署、路径过滤、排序
results, _ := publisher.BrowseDeploymentWithOptions(ctx,
    &response.BrowseDeploymentRequest{
        DeploymentIds: []string{deploymentID},
        Page:          0,
        Size:          100,
        SortField:     "createdTimestamp",   // 必填
        SortDirection: "desc",
        PathStarting:  "com/example/",
    })
```

::: warning sortField 必填
`BrowseDeploymentWithOptions` 的 `SortField` 是必填字段，为空会返回错误。
:::

## 错误处理

Publisher API 的错误响应遵循统一 schema，SDK 解析为 `*response.PublisherErrorResponse`：

```go
_, err := publisher.CheckPublished(ctx, "com.example", "my-lib", "1.0.0")
if err != nil {
    var apiErr *response.PublisherErrorResponse
    if errors.As(err, &apiErr) {
        fmt.Printf("HTTP %d, 错误码 %d: %s\n",
            apiErr.HttpStatus, apiErr.ErrorCode, apiErr.Message)
    }
}
```

可以用 `errors.As` 判断是否是 API 业务错误，还是网络层错误。
