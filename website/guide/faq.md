# 常见问题

## 通用

### 这个 SDK 和直接调 HTTP 有什么区别？

参见 [它解决了什么问题](./what-it-solves)。简而言之：把 Solr 查询、URL 拼接、分页、校验和、发布状态机、认证重试等全部封装成类型安全的方法，一行调用替代几十行模板代码。

### 搜索和下载需要认证吗？

**不需要。** Maven Central 的搜索（`search.maven.org`）和下载（`repo1.maven.org`）都是公开接口。只有**发布**需要 Sonatype Central 的 API Token。

### 支持哪些 Go 版本？

Go 1.18+。因为用到了泛型（`SearchRequestJsonDoc[Doc any]`），低于 1.18 无法编译。

### SDK 有运行时依赖吗？

没有。运行时零外部依赖，仅标准库。测试用到了 `testify`，但不会进入你的二进制。

## 发布

### 怎么申请 API Token？

1. 注册 [central.sonatype.com](https://central.sonatype.com) 账号
2. 验证你的命名空间（groupId）
3. 在账户设置中生成 API Token

Token 用于 Bearer 认证（`WithPublisherToken`）或 Basic Auth（`WithPublisherBasicAuth`，用户名 + Token 作为密码）。

### 发布一直卡在 VALIDATING 怎么办？

可能是 Sonatype 校验队列拥堵，通常几分钟内会完成。如果长时间不推进：

- 检查 bundle 是否符合 [Maven Central 要求](https://central.sonatype.org/publish/requirements/)（POM 必填字段、签名、校验和等）
- 用 `BrowseDeployment` 查看文件列表，用 `GetDeploymentStatus` 的 `Errors` 字段看具体校验错误
- 校验失败的常见原因：groupId 未验证、缺 POM、缺签名、版本号已存在

### 上传后部署 ID 是空的？

`/upload` 端点返回 `text/plain`，SDK 已处理可能的引号包裹和空白。如果还拿到空值，检查认证是否正确（401 会返回 `PublisherErrorResponse`），以及 bundle 是否为空。

### 可以取消发布吗？

部署在 `FAILED` 或 `VALIDATED` 状态下可以用 `DropDeployment` 删除。一旦进入 `PUBLISHING` 就无法取消。

## AI Agent

### Claude Code / Codex 怎么用这个 SDK？

参见 [AI Agent 接入](../ai-agent/)。我们提供了一键复制的提示词，粘贴到 AI 助手后，它会自动安装 SDK、查阅文档、生成调用代码。

### AI 会不会用错 API？

本 SDK 的方法签名遵循 Go 惯例（`ctx` 首参、`(result, error)` 返回），类型集中在 `pkg/response`，godoc 完整。AI 阅读后能准确调用。提示词里也引导 AI 先看官方文档和 GitHub，再写代码。

## 其他

### 如何贡献代码？

欢迎 PR。流程：

1. Fork [仓库](https://github.com/scagogogo/sonatype-central-sdk)
2. 创建特性分支
3. 确保通过 `go test ./...`
4. 提交 PR，描述清楚改动

### 发现 Bug 或缺能力？

在 [GitHub Issues](https://github.com/scagogogo/sonatype-central-sdk/issues) 提交，附上复现步骤和期望行为。
