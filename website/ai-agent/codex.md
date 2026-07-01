# Codex 接入

[Codex](https://openai.com/index/codex/) 是 OpenAI 的 AI 编程助手。本页提供一段**可直接复制粘贴**的提示词，让 Codex 自动接入本 SDK。

::: tip 同样适用
本提示词也适用于其他理解自然语言的 AI 编程助手（如 Cursor、Windsurf 等）。核心思路一致：告诉 AI 用哪个库、去哪看文档、写什么需求。
:::

## 一键复制提示词

::: code-group

```markdown
You are a senior Go developer. I need to use the Sonatype Central Repository capabilities (search, download, and publish Maven artifacts) in this project.

Please use the open-source Go SDK `github.com/scagogogo/sonatype-central-sdk` and guide me through these steps:

1. [Understand capabilities] Read the official docs to learn what this SDK offers:
   - Website: https://scagogogo.github.io/sonatype-central-sdk/
   - GitHub: https://github.com/scagogogo/sonatype-central-sdk
   - Focus on the method signatures of `Client` (search/download, no auth) and `PublisherClient` (publishing, needs token) in the `pkg/api` directory.

2. [Install] Run `go get github.com/scagogogo/sonatype-central-sdk` in the current project directory.

3. [Choose package] Import the correct package based on my need:
   - Search / download / version / group / class search / batch / iterator: `github.com/scagogogo/sonatype-central-sdk/pkg/api`, use `api.NewClient()`
   - Response types (if type assertions needed): `github.com/scagogogo/sonatype-central-sdk/pkg/response`
   - Publish to Maven Central: same `api` package, use `api.NewPublisherClient(api.WithPublisherToken("..."))`

4. [Write code] Generate call code that:
   - Uses context.Context for timeout control
   - Has complete error handling (for publishing, use errors.As to check *response.PublisherErrorResponse)
   - Follows Go idioms, is type-safe, avoids panic
   - Includes helpful comments

5. [Publish special case] If my need involves "publishing", first confirm whether I have a Sonatype Central API Token:
   - If yes, ask me to provide it (or read from env var, don't hardcode into committed code)
   - If no, guide me to register at https://central.sonatype.com, verify namespace, generate API Token, then continue
   - Publishing is a state machine: UploadBundle → poll GetDeploymentStatus until VALIDATED → PublishDeployment. Implement accordingly.

My specific requirement is: <describe what you want to do in one sentence, e.g.: search all artifacts under GroupId org.apache.commons and print latest versions; or download commons-lang3 3.12.0 jar and verify sha1; or publish ./bundle.zip to Maven Central>

Start with step 1 directly. Only pause to ask me when you need a decision or credentials.
```

:::

::: tip 中英皆可
Codex 对中文提示词同样支持良好。如果你的需求描述是中文，直接用中文写在 `<...>` 处即可。提示词主体用英文是因为 Codex 对英文指令的遵循度略高，但你也可以整体改成中文。
:::

## 使用方法

1. **复制**上面整段提示词
2. **替换**末尾 `<describe what you want to do...>` 那一行，写成你的真实需求
3. **粘贴**到 Codex，回车
4. Codex 会自动读文档、装 SDK、写代码

## 常见需求填法（中文示例）

<details>
<summary>🔍 搜索</summary>

```
My specific requirement is: 搜索 GroupId 为 org.apache.commons 的所有制品，打印 ArtifactId 和最新版本，最多 20 条。
```
</details>

<details>
<summary>📦 下载</summary>

```
My specific requirement is: 下载 commons-lang3 3.12.0 的 jar，用官方 sha1 校验和验证完整性，保存到当前目录。
```
</details>

<details>
<summary>🚀 发布</summary>

```
My specific requirement is: 把 ./my-lib-bundle.zip 发布到 Maven Central，groupId=com.example，artifactId=my-lib，自动发布模式，轮询直到完成或失败。
```
</details>

## 提示词要点

- **先读文档再写码**：确保 Codex 用当前版本 API，不靠可能过时的训练记忆
- **区分两类客户端**：搜索无认证、发布需 Token，避免混淆
- **发布状态机**：明确告诉 Codex 发布是 `Upload → 轮询 → Publish` 流程
- **凭据安全**：要求从环境变量读 Token，不硬编码

## 与 Claude Code 提示词的区别

两者结构相同，主要差异：

| 维度 | Claude Code | Codex |
|------|-------------|-------|
| 语言 | 简体中文 | 英文（中文亦可） |
| 文档交互 | 可直接用 WebFetch 读官网 | 同样可读网页 |
| 凭据环境变量 | 同样推荐 `os.Getenv` | 同样推荐 |

如果你同时用多个 AI 助手，可以都配上这段提示词，生成的代码风格会一致。
