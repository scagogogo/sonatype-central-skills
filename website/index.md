---
# https://vitepress.dev/reference/default-theme-home-frontmatter
layout: home

hero:
  name: Sonatype Central SDK
  text: Maven Central 的 Go SDK
  tagline: 搜索、下载、发布 Maven 制品 —— 全面、类型安全、为 AI Agent 而生
  image:
    src: /logo.svg
    alt: Sonatype Central SDK
  actions:
    - theme: brand
      text: 快速开始
      link: /guide/quick-start
    - theme: alt
      text: AI Agent 接入
      link: /ai-agent/
    - theme: alt
      text: GitHub
      link: https://github.com/scagogogo/sonatype-central-sdk

features:
  - icon: 🔍
    title: 全维度搜索
    details: 按 GroupId、ArtifactId、版本、SHA1、类名、全限定类名、标签、打包方式、分类器和全文搜索 Maven 制品。
  - icon: 📦
    title: 下载与校验
    details: 下载 POM、JAR、源码、javadoc、SBOM（CycloneDX/SPDX），并使用官方 SHA1/MD5/SHA256 校验和验证完整性。
  - icon: 🚀
    title: 发布到 Maven Central
    details: 上传部署包、轮询部署状态、检查发布情况、浏览部署文件、一键发布到 Maven Central。
  - icon: ⚡
    title: 批量与迭代器
    details: 批量搜索、批量下载、并发处理，以及针对海量结果集的内存高效懒加载迭代器。
  - icon: 🤖
    title: 原生 AI Agent 支持
    details: 为 Claude Code、Codex 等 AI 编程助手提供一键接入提示词，复制即用，AI 自动完成配置与调用。
  - icon: 💾
    title: 缓存与重试
    details: 内置缓存支持和指数退避重试机制，从容应对限流与网络抖动。
---

<div class="ai-prompt-section">

## 🤖 AI Agent 一键接入

把这个提示词**复制粘贴**到你的 AI 编程助手（[Claude Code](https://claude.com/claude-code) / Codex），它就会自动帮你安装、配置并使用本 SDK —— 你只需要描述你想做什么。

::: code-group

```markdown [Claude Code / 通用提示词]
你是一个 Go 开发专家。我们现在需要在项目中使用 Sonatype Central Repository 的能力（搜索、下载、发布 Maven 制品）。

请使用 GitHub 上的开源 Go SDK `github.com/scagogogo/sonatype-central-sdk` 来完成。

请按以下步骤引导我：

1. 阅读官方文档了解能力：https://scagogogo.github.io/sonatype-central-sdk/  和  https://github.com/scagogogo/sonatype-central-sdk
2. 在当前 Go 项目中执行 `go get github.com/scagogogo/sonatype-central-sdk` 安装 SDK
3. 根据我要实现的功能，选择合适的包导入：
   - 搜索/下载：`github.com/scagogogo/sonatype-central-sdk/pkg/api`（`api.NewClient()`）
   - 发布到 Maven Central：同包下的 `api.NewPublisherClient(api.WithPublisherToken("..."))`
4. 编写类型安全、带错误处理和 context 超时的调用代码
5. 如果涉及发布操作，先确认我是否已有 Sonatype Central 的 API Token；若没有，引导我去 https://central.sonatype.com 申请

我的具体需求是：<在这里用一句话描述你要做的事，比如“搜索 org.apache.commons 下所有制品的最新版本”>

请直接开始，遇到需要我决策的地方再问我。
```

:::

> 💡 把上面代码块最后那行 `<...>` 替换成你的真实需求即可。AI 会读懂 SDK 的 API、自动补全调用代码、处理错误和分页。更多接入方式见 [AI Agent 接入指南](./ai-agent/)。



::: details 📌 为什么这个提示词能work？
本 SDK 的所有方法都遵循 Go 习惯的 `(ctx, 参数) -> (结果, error)` 签名，类型定义集中在 `pkg/response` 包，AI 可以通过阅读 `pkg/api` 下的方法签名和 godoc 快速理解能力边界，无需你逐个教它。提示词只是告诉 AI **用哪个库**和**去哪看文档**，剩下的它自己能搞定。
:::

</div>

<style>
.ai-prompt-section {
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 1px solid var(--vp-c-divider);
}
.ai-prompt-section h2 {
  text-align: center;
  margin-bottom: 1.5rem;
}
.VPHero .VPImage.image-src {
  max-width: 220px;
}
</style>

---

## 这个 SDK 解决什么问题？

直接调用 [Sonatype Central](https://central.sonatype.com) 和 [Maven Central](https://repo1.maven.org) 的 HTTP API 需要自己处理：Solr 查询语法、URL 拼接、分页、JSON 解析、错误码、限流重试、校验和验证、multipart 上传……繁琐且容易出错。

本 SDK 把这些全部封装成**类型安全**的 Go 方法，一行调用即可完成原本几十行 HTTP 代码的工作：

```go
client := api.NewClient()

// 搜索 org.apache.commons 下的制品
artifacts, _ := client.SearchByGroupId(ctx, "org.apache.commons", 10)

// 下载并验证校验和
data, _, _ := client.DownloadWithVerifiedChecksum(ctx,
    "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar", "sha1")
```

👉 详细的设计与原理见 [它解决了什么问题](./guide/what-it-solves) 和 [工作原理](./guide/how-it-works)。

## 马上开始

- **想直接用？** → [快速开始](./guide/quick-start)
- **想让 AI 帮你写？** → [AI Agent 接入](./ai-agent/)
- **想了解全貌？** → [介绍](./guide/introduction)
