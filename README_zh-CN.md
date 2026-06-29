# Sonatype Central SDK for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/scagogogo/sonatype-central-sdk.svg)](https://pkg.go.dev/github.com/scagogogo/sonatype-central-sdk)
[![Build Status](https://github.com/scagogogo/sonatype-central-sdk/actions/workflows/go-test.yml/badge.svg)](https://github.com/scagogogo/sonatype-central-sdk/actions/workflows/go-test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/scagogogo/sonatype-central-sdk)](https://goreportcard.com/report/github.com/scagogogo/sonatype-central-sdk)
[![License: MIT](https://img.shields.io/github/license/scagogogo/sonatype-central-sdk)](https://github.com/scagogogo/sonatype-central-sdk/blob/main/LICENSE)

**[English](README.md)**

一个全面、类型安全的 Go SDK，用于 [Sonatype Central Repository API](https://central.sonatype.org/search/rest-api-guide/)。通过简洁、符合 Go 习惯的接口搜索、下载和发布 Maven 制品。

---

## 功能特性

- **🔍 搜索** — 按 GroupId、ArtifactId、版本、SHA1、类名、全限定类名、标签、打包方式、分类器和全文搜索
- **📦 下载** — 下载 POM、JAR、源码、javadoc、SBOM（CycloneDX/SPDX）及其他制品文件
- **✅ 校验和验证** — 使用 Maven Central 官方 SHA1/MD5/SHA256 校验和验证文件完整性
- **🏷️ 版本** — 列出版本、获取最新版本、比较版本、过滤版本
- **📂 分组** — 搜索分组、获取分组统计、比较分组、搜索子分组
- **🎯 GAV** — 按 GAV 坐标搜索、分页列表、排序查询
- **🔖 标签** — 按标签搜索、多标签搜索、标签前缀搜索
- **🔬 类搜索** — 按类名、全限定类名、包名、类层次结构、接口实现搜索（支持高亮）
- **📄 制品** — 搜索制品、获取制品详情、比较制品、获取制品统计
- **🏷️ 分类器** — 按分类器搜索（sources、javadoc、tests 等）
- **🔤 拼写检查** — 获取搜索查询的拼写建议
- **🚀 发布** — 上传部署包、检查部署状态、带过滤条件的部署列表、发布到 Maven Central
- **⚡ 批量操作** — 批量搜索、批量下载、并发处理
- **🔁 迭代器** — 针对大量结果集的内存高效懒加载迭代器
- **💾 缓存与重试** — 内置缓存支持和指数退避重试机制
- **🎛️ 高级搜索** — 精确匹配、自定义字段列表、查询解析器选择（dismax/edismax）、查询字段权重

## 安装

```bash
go get github.com/scagogogo/sonatype-central-sdk
```

## 快速开始

```go
package main

import (
    "context"
    "fmt"

    "github.com/scagogogo/sonatype-central-sdk/pkg/api"
)

func main() {
    client := api.NewClient()

    // 按 GroupId 搜索制品
    artifacts, err := client.SearchByGroupId(context.Background(), "org.apache.commons", 10)
    if err != nil {
        panic(err)
    }

    for _, artifact := range artifacts {
        fmt.Printf("GroupId: %s, ArtifactId: %s, 最新版本: %s\n",
            artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
    }
}
```

## 发布 API

发布 API 允许你上传和发布制品到 Maven Central，需要 Sonatype API Token 进行认证。

```go
package main

import (
    "context"
    "fmt"

    "github.com/scagogogo/sonatype-central-sdk/pkg/api"
    "github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

func main() {
    client := api.NewPublisherClient(
        api.WithPublisherToken("your-sonatype-api-token"),
    )

    // 检查组件是否已发布
    published, err := client.CheckPublished(context.Background(),
        "com.example", "my-library", "1.0.0")
    if err != nil {
        panic(err)
    }
    fmt.Printf("已发布: %v\n", published.Published)

    // 列出部署，支持过滤和分页
    deployments, err := client.ListDeployments(context.Background(), &response.DeploymentListOptions{
        Namespace: "com.example",
        State:     response.DeploymentStateValidated,
        Paginate:  true,
        Page:      0,
        Size:      20,
    })
    if err != nil {
        panic(err)
    }
    for _, d := range deployments.Deployments {
        fmt.Printf("部署: %s, 状态: %s\n", d.DeploymentName, d.DeploymentState)
    }
}
```

### 发布 API 方法

| 方法 | 说明 |
|------|------|
| `UploadBundle` | 上传部署包（ZIP） |
| `GetDeploymentStatus` | 按 ID 查询部署状态 |
| `CheckPublished` | 检查组件（namespace/name/version）是否已发布 |
| `ListDeployments` | 列出部署，支持过滤和分页 |
| `BrowseDeployment` | 浏览部署文件（便捷方法） |
| `BrowseDeploymentWithOptions` | 浏览部署文件，支持完整选项（sortField、分页、路径过滤） |
| `DownloadDeploymentFile` | 下载部署中的指定文件 |
| `DropDeployment` | 删除部署（仅 FAILED/VALIDATED 状态） |
| `PublishDeployment` | 发布部署（仅 VALIDATED 状态） |

## 高级搜索

### 精确匹配

```go
searchRequest := request.NewSearchRequest().
    SetQuery(request.NewQuery().SetGroupId("org.apache.commons")).
    SetExact(true)  // 启用精确匹配
```

### 拼写检查

```go
// 搜索并获取拼写建议
artifacts, suggestions, err := client.SearchWithSpellcheck(ctx, "commns-lang", 10, 5)
if len(suggestions) > 0 {
    fmt.Printf("你是否要找: %v?\n", suggestions)
}
```

### 自定义字段列表

```go
// 只返回特定字段以减少响应大小
searchRequest := request.NewSearchRequest().
    SetQuery(request.NewQuery().SetText("commons")).
    SetFieldList("id,g,a,latestVersion")
```

### 查询解析器和字段权重

```go
// 使用 edismax 解析器并自定义字段权重
searchRequest := request.NewSearchRequest().
    SetQuery(request.NewQuery().SetText("json parser")).
    SetDefType("edismax").
    SetQueryFields("text^20 g^5 a^10 c^3 fc^2")
```

### 分类器搜索

```go
// 按分类器搜索制品
artifacts, err := client.SearchByClassifier(ctx, "sources", 10)

// 结合 GroupId 搜索
artifacts, err := client.SearchByGroupAndClassifier(ctx, "org.apache.commons", "sources", 10)
```

## 下载 API

### SBOM 下载

```go
// 下载 CycloneDX JSON 格式的 SBOM
sbom, err := client.DownloadCycloneDXJSON(ctx, "com.example", "my-lib", "1.0.0")

// 下载 CycloneDX XML 格式的 SBOM
sbom, err := client.DownloadCycloneDXXML(ctx, "com.example", "my-lib", "1.0.0")

// 下载 SPDX JSON 格式的 SBOM
sbom, err := client.DownloadSpdxJSON(ctx, "com.example", "my-lib", "1.0.0")
```

### 校验和验证

```go
// 下载文件并使用官方校验和验证
data, checksum, err := client.DownloadWithVerifiedChecksum(ctx,
    "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar",
    "sha1")

// 仅下载官方校验和文件
checksum, err := client.DownloadChecksumFile(ctx,
    "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar",
    "sha256")
```

## API 参考

### 基础 URL

| API | 基础 URL | 需要认证 |
|-----|---------|:-------:|
| 搜索 | `https://search.maven.org/solrsearch/select` | 否 |
| 下载 | `https://repo1.maven.org/maven2` | 否 |
| 发布 | `https://central.sonatype.com/api/v1/publisher` | 是 |

### 支持的 Solr 查询字段

| 字段 | 说明 |
|------|------|
| `g:` | GroupId |
| `a:` | ArtifactId |
| `v:` | 版本 |
| `p:` | 打包方式（jar、pom、war 等） |
| `l:` | 分类器（sources、javadoc、tests 等） |
| `c:` | 类名（简单名称） |
| `fc:` | 全限定类名 |
| `1:` | SHA-1 校验和 |
| `tags:` | 标签搜索 |
| `id:` | 制品 ID |
| `text:` | 全文搜索 |

### 高级搜索参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `exact` | 启用精确匹配 | `false` |
| `spellcheck` | 启用拼写检查 | 自动 |
| `spellcheck.count` | 拼写建议数量 | `5` |
| `fl` | 返回字段列表 | Solr 默认 |
| `defType` | 查询解析器类型 | `dismax` |
| `qf` | 查询字段权重 | `text^20 g^5 a^10` |

### 已弃用的 API

> ⚠️ 以下 API 端点在 Sonatype Central 上已不可用：

| API | 状态 | 替代方案 |
|-----|------|---------|
| 安全/漏洞 | `search.maven.org/api/security/*` 返回 403 | 使用 OWASP Dependency-Check |
| 许可证搜索 | Solr `l:` 字段（许可证）返回空结果 | 改为解析 POM 文件 |
| 依赖搜索 | Solr `d:` 字段返回 400 | 下载并解析 POM 文件 |
| 方法搜索 | Solr `m:` 字段返回 400 | — |
| 聚合/统计 | Solr facet 功能已禁用 | — |

## 项目结构

```
sonatype-central-sdk/
├── pkg/
│   ├── api/           # API 客户端和方法实现
│   ├── request/       # HTTP 请求构建器
│   ├── response/      # 响应类型定义
│   └── examples/      # 使用示例
├── .github/           # CI/CD 工作流
├── docs/              # 额外文档
├── go.mod
└── LICENSE
```

## 环境要求

- Go 1.18+

## 许可证

[MIT 许可证](LICENSE)
