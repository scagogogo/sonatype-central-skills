# 工作原理

本章介绍 SDK 的架构设计与关键技术决策，帮助你理解它"为什么这样用"。

## 整体架构

SDK 采用经典的**三层架构**：

```
┌─────────────────────────────────────────────┐
│  你的应用代码                                  │
├─────────────────────────────────────────────┤
│  pkg/api        ← 业务方法（搜索/下载/发布）   │
├─────────────────────────────────────────────┤
│  pkg/request    ← 请求构建器（Solr 参数封装）   │
│  pkg/response   ← 响应类型（强类型结构体）      │
├─────────────────────────────────────────────┤
│  底层 HTTP 客户端（带重试/缓存/限流）           │
├─────────────────────────────────────────────┤
│  Sonatype Central / Maven Central REST API    │
└─────────────────────────────────────────────┘
```

- **`pkg/api`**：暴露给用户的高层方法。两类客户端：`Client`（搜索/下载，无需认证）和 `PublisherClient`（发布，需认证）。
- **`pkg/request`**：把 Solr 查询参数封装成链式构建器（`SearchRequest`），避免字符串拼接。
- **`pkg/response`**：所有 API 响应的强类型定义，集中管理。

## 两个客户端的分工

### `api.Client` —— 搜索与下载

面向公开的 Maven Central 接口，**无需认证**：

```go
client := api.NewClient()  // 默认指向 search.maven.org / repo1.maven.org
```

它封装了 `solrsearch/select` 和 `maven2/<path>` 两类端点。

### `api.PublisherClient` —— 发布

面向 Sonatype Central Publisher API，**需要认证**：

```go
client := api.NewPublisherClient(
    api.WithPublisherToken("your-bearer-token"),
    // 或 Basic Auth:
    // api.WithPublisherBasicAuth("user", "pass"),
)
```

认证通过**选项函数**（functional options）配置，这是 Go 社区惯用的可扩展配置模式。

## 关键技术决策

### 1. URL 拼接：字符串拼接而非 `url.JoinPath`

你可能注意到 `PublisherClient` 内部用的是 `baseURL + path` 而非 `url.JoinPath`。这是因为 Go 的 `url.JoinPath` 会对 `?` 进行百分号编码（变成 `%3F`），导致查询字符串被当作路径的一部分——所有带查询参数的请求都会失败。

这是一个真实踩过的坑，因此 SDK 在内部统一用字符串拼接 + `url.Values.Encode()` 构造查询串，保证行为正确。

### 2. 响应解析：JSON 与 text/plain 分流

Publisher API 的响应类型不统一：
- 大多数端点返回 JSON
- `/upload` 返回 **`text/plain`**（仅一个部署 ID 字符串，可能带引号）

SDK 内部分成两个底层方法：
- `doRequest` —— 解析 JSON 到结构体
- `doRequestRaw` —— 返回原始字节

`UploadBundle` 用 `doRequestRaw` 拿到原始文本，再 `TrimSpace` + 去引号得到部署 ID。错误响应（HTTP ≥ 400）统一解析为 `PublisherErrorResponse`。

### 3. 泛型响应包装

搜索 API 的 Solr 响应有统一的分页外壳（`docs`/`numFound`/`start`），但 `docs` 里的元素类型因查询而异。SDK 用 Go 泛型定义：

```go
type SearchRequestJsonDoc[Doc any] struct { ... }
// 返回 *response.Response[Doc]，Doc 是具体制品类型
```

这样同一套分页逻辑可以复用于不同制品类型，编译期类型安全。

### 4. 迭代器模式处理海量结果

对于可能返回成千上万条结果的查询，SDK 提供 `Iterator*` 系列方法。内部封装了分页游标，对外暴露 `HasNext()`/`Next()` 接口，**按需加载**，避免一次性把所有结果载入内存。

### 5. 缓存与重试的内置集成

- **缓存**：可启用，命中时直接返回，减少对 Maven Central 的请求。可自定义 TTL。
- **重试**：内置指数退避，应对 429 限流和瞬时网络错误。

两者都在底层 HTTP 客户端实现，对上层业务方法透明。

## 包依赖关系图

```
pkg/api
  ├─ pkg/request   (构建查询参数)
  └─ pkg/response  (解析响应)
        └─ 内置 Go 标准库 (net/http, encoding/json)
```

SDK **零外部运行时依赖**（仅测试用 `testify`），适合对依赖体积敏感的生产环境。

## 可测试性

所有方法都接受 `context.Context` 作为首个参数，便于控制超时与取消。底层 HTTP 客户端可通过 `WithHTTPClient` / `WithPublisherHTTPClient` 替换，便于注入 `httptest` 服务器做单元测试——SDK 自身的测试套件正是这样做的。

---

理解了原理后，去看 [快速开始](./quick-start) 跑通第一个例子，或直接进入 [AI Agent 接入](../ai-agent/) 让 AI 帮你写代码。
