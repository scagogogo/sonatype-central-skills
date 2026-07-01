# Claude Code 接入

[Claude Code](https://claude.com/claude-code) 是 Anthropic 的命令行 AI 编程助手。本页提供一段**可直接复制粘贴**的提示词，粘贴到 Claude Code 后，它会自动帮你接入本 SDK。

## 一键复制提示词

点击右上角复制按钮，把下面整段贴进 Claude Code：

::: code-group

```markdown
你是一个资深 Go 开发专家。我现在要在这个项目里使用 Sonatype Central Repository 的能力（搜索、下载、发布 Maven 制品）。

请使用开源 Go SDK `github.com/scagogogo/sonatype-central-sdk` 来完成，按以下步骤引导我：

1. 【了解能力】先访问并阅读官方文档，理解这个 SDK 提供了哪些能力：
   - 官网：https://scagogogo.github.io/sonatype-central-sdk/
   - GitHub：https://github.com/scagogogo/sonatype-central-sdk
   - 重点看 pkg/api 目录下 Client（搜索/下载，无需认证）和 PublisherClient（发布，需 Token）的方法签名

2. 【安装】在当前项目目录执行 `go get github.com/scagogogo/sonatype-central-sdk`

3. 【选包】根据我的需求导入正确的包：
   - 搜索 / 下载 / 版本 / 分组 / 类搜索 / 批量 / 迭代器：`github.com/scagogogo/sonatype-central-sdk/pkg/api`，用 `api.NewClient()`
   - 响应类型（如需要类型断言）：`github.com/scagogogo/sonatype-central-sdk/pkg/response`
   - 发布到 Maven Central：同 api 包，用 `api.NewPublisherClient(api.WithPublisherToken("..."))`

4. 【写代码】生成符合以下要求的调用代码：
   - 使用 context.Context 控制超时
   - 完整的 error 处理（发布相关的错误可用 errors.As 判断 *response.PublisherErrorResponse）
   - 遵循 Go 习惯，类型安全，避免 panic
   - 包含必要的注释说明

5. 【发布场景特例】如果我的需求涉及"发布"，先停下来确认我是否已持有 Sonatype Central 的 API Token：
   - 如果有，让我提供 Token（或通过环境变量读取，不要硬编码到提交的代码里）
   - 如果没有，引导我去 https://central.sonatype.com 注册账号、验证命名空间、生成 API Token，然后再继续
   - 发布流程是状态机：UploadBundle → 轮询 GetDeploymentStatus 等待 VALIDATED → PublishDeployment；请按这个流程实现

我的具体需求是：<在这里用一句话描述你要做的事，例如：搜索 org.apache.commons 下所有制品并打印最新版本；或下载 commons-lang3 3.12.0 的 jar 并验证 sha1；或把 ./bundle.zip 发布到 Maven Central>

请直接开始执行第 1 步，遇到需要我决策或提供凭据的地方再停下来问我。
```

:::

## 使用方法

1. **复制**上面整段提示词
2. **替换**末尾 `<在这里用一句话描述你要做的事...>` 那一行，写成你的真实需求
3. **粘贴**到 Claude Code 的输入框，回车
4. Claude Code 会自动开始：读文档 → 装 SDK → 写代码 → 跑测试

## 几个常见需求的填法

<details>
<summary>🔍 搜索场景</summary>

```
我的具体需求是：搜索 GroupId 为 org.apache.commons 的所有制品，打印每个制品的 ArtifactId 和最新版本号，最多取 20 条。
```
</details>

<details>
<summary>📦 下载场景</summary>

```
我的具体需求是：下载 commons-lang3 3.12.0 的 jar 文件，并用官方 sha1 校验和验证完整性，保存到当前目录的 commons-lang3-3.12.0.jar。
```
</details>

<details>
<summary>🚀 发布场景</summary>

```
我的具体需求是：把当前目录下的 my-lib-bundle.zip 发布到 Maven Central，groupId 是 com.example，artifactId 是 my-lib，使用自动发布模式，并轮询直到发布完成或失败。
```
</details>

<details>
<summary>🔬 类搜索场景</summary>

```
我的具体需求是：我有一个类名 StringUtils，帮我搜索 Maven Central 上哪些制品包含这个类，返回制品坐标列表。
```
</details>

## 提示词为什么这样设计

- **第 1 步先读文档**：让 Claude Code 访问官网和 GitHub，确保它用的是**当前版本**的 API，而不是记忆里可能过时的版本
- **明确两类客户端**：`Client`（无认证）和 `PublisherClient`（需 Token），避免 AI 在搜索场景误加认证
- **发布场景特例**：发布涉及凭据和状态机，提示词专门提醒 AI 先确认 Token、按状态机流程实现
- **要求 context 和错误处理**：保证生成的代码是生产可用的，不是玩具代码

## 进阶：加到 CLAUDE.md

如果你经常用，可以把提示词核心部分加到项目的 `CLAUDE.md`，这样每次开新会话 Claude Code 都自动知道用这个 SDK：

```markdown
## Maven Central 依赖操作

本项目使用 `github.com/scagogogo/sonatype-central-sdk` 处理 Maven 制品的搜索、下载、发布。
- 文档：https://scagogogo.github.io/sonatype-central-sdk/
- 搜索/下载用 `api.NewClient()`（无需认证）
- 发布用 `api.NewPublisherClient(api.WithPublisherToken(os.Getenv("SONATYPE_TOKEN")))`
- 发布遵循状态机：UploadBundle → 轮询 GetDeploymentStatus → PublishDeployment
```

Token 通过环境变量传入，不要硬编码。
