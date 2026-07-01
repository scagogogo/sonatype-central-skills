# 它解决了什么问题

本章逐条对比**手写 HTTP 调用**与**使用本 SDK**，让你清楚看到 SDK 抹平了哪些复杂性。

## 1. Solr 查询语法的门槛

Maven Central 的搜索 API 底层是 [Solr](https://solr.apache.org/)。要按 GroupId 搜索，你得构造这样的查询字符串：

```
https://search.maven.org/solrsearch/select?q=g:"org.apache.commons"&rows=10&wt=json
```

字段名是缩写（`g`、`a`、`v`、`fc`...），新人很难记住。而且不同字段的语义差异大——比如 `c:` 是类名、`fc:` 是全限定类名、`1:` 是 SHA1。

**使用 SDK：**

```go
artifacts, err := client.SearchByGroupId(ctx, "org.apache.commons", 10)
```

方法名即语义，参数即业务概念。SDK 内部负责构造 Solr 查询、拼 URL、解析 JSON。

## 2. URL 路径拼接的转义陷阱

下载制品时，路径是 GroupId 的点号转斜杠 + ArtifactId + 版本 + 文件名：

```
https://repo1.maven.org/maven2/org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar
```

手写时容易在 `strings.Replace(g, ".", "/", -1)` 上出错，或在版本号含特殊字符时漏掉转义。SDK 的 `Download` 系列方法接受 `groupId/artifactId/version` 三个参数，内部正确处理。

## 3. 分页与海量结果

Solr 一次最多返回几百条。要遍历一个大型 Group 下的所有版本，得自己维护 `start`/`rows` 游标、处理总数、防止越界。

**使用 SDK 的迭代器：**

```go
iter := client.IteratorByGroupId(ctx, "org.apache.commons")
for iter.HasNext() {
    artifact, err := iter.Next()
    if err != nil { break }
    // 处理每一条
}
```

内存高效、自动分页、无需关心游标。

## 4. 校验和验证

下载文件后验证完整性是安全最佳实践，但需要先下载文件、再下载 `.sha1` 文件、再计算本地哈希、再比对——四步。

**使用 SDK 一行完成：**

```go
data, checksum, err := client.DownloadWithVerifiedChecksum(ctx, path, "sha1")
// data 已通过校验，可直接使用
```

## 5. 发布流程的状态机

发布到 Maven Central 不是一次 HTTP 调用，而是一个**异步状态机**：

```
上传 bundle → PENDING → VALIDATING → VALIDATED → PUBLISHING → PUBLISHED
                                        ↓ (失败)
                                      FAILED
```

你需要：上传 → 轮询状态 → 校验通过后发布 → 继续轮询直到 PUBLISHED。每一步的 HTTP 方法、参数、响应体都不同。

**使用 SDK：**

```go
publisher := api.NewPublisherClient(api.WithPublisherToken(token))

id, _ := publisher.UploadBundle(ctx, bundle, "my-lib", api.PublishingTypeAutomatic)
// 轮询...
status, _ := publisher.GetDeploymentStatus(ctx, id)
// 校验通过后发布
_ = publisher.PublishDeployment(ctx, id)
```

每个状态转移都是一个方法调用，错误响应被解析为结构化的 `PublisherErrorResponse`。

## 6. 认证与限流

发布 API 需要 Bearer Token 或 Basic Auth；搜索 API 有速率限制，高频调用会被限流。手写时你得自己加 `Authorization` 头、自己实现退避重试。

**使用 SDK：**

```go
// 认证：通过选项函数配置
client := api.NewPublisherClient(api.WithPublisherToken("your-token"))

// 重试：内置指数退避
client.EnableCache()      // 命中缓存，减少请求
client.SetCacheTTL(...)   // 自定义 TTL
```

## 7. 类型安全与可维护性

手写 HTTP + `json.Unmarshal` 到 `map[string]interface{}`，没有编译期检查，字段名拼错只能运行时发现。SDK 把所有响应定义成强类型结构体（集中在 `pkg/response` 包），IDE 自动补全、编译期校验、重构友好。

---

## 总结

| 痛点 | 手写 HTTP | 本 SDK |
|------|----------|--------|
| Solr 查询语法 | 自己拼字符串 | 语义化方法 |
| URL 转义 | 手动处理 | 自动处理 |
| 分页 | 手动游标 | 迭代器封装 |
| 校验和 | 四步手动 | 一行调用 |
| 发布状态机 | 多步手动编排 | 每步一个方法 |
| 认证/重试 | 自己实现 | 内置支持 |
| 类型安全 | 弱（map） | 强（结构体） |

👉 想了解 SDK 内部是如何做到这些的？继续看 [工作原理](./how-it-works)。
