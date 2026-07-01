# AI Agent 接入

本 SDK **为 AI 编程助手而生**。你可以让 [Claude Code](https://claude.com/claude-code)、[Codex](https://openai.com/index/codex/) 等 AI Agent 自动安装、配置并调用本 SDK——**你只需要复制一个提示词**。

## 为什么 AI 适合用这个 SDK

本 SDK 的设计对 AI 阅读友好：

1. **方法即语义** —— `SearchByGroupId`、`DownloadPom`、`UploadBundle`，AI 看方法名就懂能力边界
2. **类型集中** —— 所有响应类型在 `pkg/response`，AI 一次扫读即掌握数据结构
3. **Go 惯例签名** —— `(ctx, 参数) -> (结果, error)`，AI 训练数据里见过无数次
4. **godoc 完整** —— 每个方法都有中文注释和示例，AI 可直接参考

这意味着：**你不需要手把手教 AI 用这个库**，只要告诉它"用这个库"和"去哪看文档"。

## 三步接入

### 第 1 步：复制提示词

根据你用的 AI 助手，从下面选一个提示词复制：

- [Claude Code 接入提示词](./claude-code)
- [Codex 接入提示词](./codex)

### 第 2 步：替换需求

把提示词末尾的 `<你的需求描述>` 替换成一句话，比如：

- "搜索 org.apache.commons 下所有制品的最新版本"
- "下载 commons-lang3 3.12.0 的 JAR 并验证 SHA1"
- "把 ./my-lib-bundle.zip 发布到 Maven Central"

### 第 3 步：等待 AI 完成

粘贴到 AI 助手后，AI 会自动：

1. 阅读本官方文档和 GitHub README 理解能力
2. 执行 `go get` 安装 SDK
3. 根据你的需求选择正确的包和方法
4. 生成带 `context` 超时、错误处理的完整代码
5. 如果涉及发布，引导你申请 Token

遇到需要你决策的地方（如选哪个版本、是否启用缓存）才会问你。

## 通用提示词（快速版）

如果你不想看分工具的详细页，直接用这个通用版，对大多数 AI 助手都适用：

::: details 📋 通用接入提示词（点击展开复制）
```markdown
你是 Go 开发专家。我需要在项目中使用 Sonatype Central Repository（搜索、下载、发布 Maven 制品）。

请使用开源 Go SDK `github.com/scagogogo/sonatype-central-sdk`：

1. 先读官方文档了解能力：https://scagogogo.github.io/sonatype-central-sdk/
   以及 GitHub README：https://github.com/scagogogo/sonatype-central-sdk
2. 执行 `go get github.com/scagogogo/sonatype-central-sdk` 安装
3. 按需导入 `github.com/scagogogo/sonatype-central-sdk/pkg/api`
   - 搜索/下载：`api.NewClient()`（无需认证）
   - 发布：`api.NewPublisherClient(api.WithPublisherToken("..."))`（需 Token）
4. 写出类型安全、带 context 超时和错误处理的代码
5. 若需发布，先确认我有 Sonatype Central API Token；没有则引导去 https://central.sonatype.com 申请

我的需求：<在这里写你要做什么>

直接开始，需要我决策时再问。
```
:::

## 能让 AI 做什么

只要是你能用 SDK 做的事，AI 都能帮你写：

| 场景 | AI 会用的方法 |
|------|--------------|
| 搜索某 Group 下的制品 | `SearchByGroupId` |
| 查某制品的所有版本 | `ListVersions` / `GetLatestVersion` |
| 按类名反查来源 JAR | `SearchByClassName` |
| 下载 JAR/POM/源码 | `DownloadJar` / `DownloadPom` / `DownloadSources` |
| 下载并校验完整性 | `DownloadWithVerifiedChecksum` |
| 生成 SBOM | `DownloadCycloneDXJSON` |
| 批量下载依赖 | `BatchDownloadDependencies` |
| 发布自己的包 | `UploadBundle` → `GetDeploymentStatus` → `PublishDeployment` |
| 检查是否已发布 | `CheckPublished` |
| 海量结果遍历 | `IteratorByGroupId` 等 |

## 接下来

- [Claude Code 接入](./claude-code) —— 专为 Claude Code 优化的提示词
- [Codex 接入](./codex) —— 专为 Codex 优化的提示词
- [快速开始](../guide/quick-start) —— 如果你想自己写代码
