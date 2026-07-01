# 介绍

**Sonatype Central SDK** 是一个全面、类型安全的 [Go](https://go.dev) SDK，用于 [Sonatype Central Repository](https://central.sonatype.com) 和 [Maven Central](https://repo1.maven.org) 的 API。

它让你能够在 Go 程序中**搜索**、**下载**、**发布** Maven 制品，而不必直接面对底层 HTTP 接口的复杂性。

## 为什么需要它

Maven Central 是 JVM 生态最大的制品仓库，几乎所有 Java/Kotlin/Scala 项目都依赖它。但它对外提供的 API 分散、风格不一：

| 能力 | 实际接口 | 风格 |
|------|---------|------|
| 搜索 | `https://search.maven.org/solrsearch/select` | Solr 查询参数 |
| 下载 | `https://repo1.maven.org/maven2/<path>` | 纯路径拼接 |
| 发布 | `https://central.sonatype.com/api/v1/publisher/*` | REST + multipart + Bearer 认证 |

直接用 `net/http` 调用它们，你需要自己处理：Solr 查询语法、URL 路径拼接与转义、分页与游标、JSON 结构解析、HTTP 错误码映射、限流与重试、SHA 校验和验证、multipart 表单上传……

本 SDK 把这一切封装成清晰、类型安全、可测试的 Go 方法。**一行调用，替代几十行模板代码。**

## 核心特性

- **🔍 全维度搜索** — GroupId、ArtifactId、版本、SHA1、类名、全限定类名、标签、打包方式、分类器、全文
- **📦 下载与校验** — POM、JAR、源码、javadoc、SBOM（CycloneDX/SPDX），可附带官方校验和验证
- **🚀 发布到 Maven Central** — 上传部署包、轮询状态、检查发布、浏览文件、一键发布
- **⚡ 批量与迭代器** — 批量搜索/下载、并发处理、海量结果的懒加载迭代器
- **💾 缓存与重试** — 内置缓存、指数退避重试，从容应对限流
- **🎛️ 高级搜索** — 精确匹配、自定义字段列表、查询解析器（dismax/edismax）、字段权重

## 它适合谁

- **Go 后端开发者** —— 需要在服务中查询或拉取 Maven 制品（如依赖分析、SBOM 生成）
- **DevOps / 平台工程师** —— 构建内部制品镜像、依赖巡检、发布自动化流水线
- **安全研究员** —— 批量扫描制品漏洞、追踪版本变更
- **AI 编程助手用户** —— 让 Claude Code / Codex 自动调用本 SDK 完成 Maven 相关任务

## 下一步

- [快速开始](./quick-start) — 5 分钟跑通第一个搜索
- [它解决了什么问题](./what-it-solves) — 逐条对比手写 HTTP 的痛点
- [AI Agent 接入](../ai-agent/) — 让 AI 帮你写代码
