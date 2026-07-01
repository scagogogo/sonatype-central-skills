# 缓存与重试

SDK 内置缓存和重试机制，从容应对限流、网络抖动和重复请求。

## 缓存

启用缓存后，相同的查询会命中本地缓存，直接返回，**减少对 Maven Central 的请求**。这对高频调用同一坐标的场景（如轮询、CI 反复检查）尤其有用。

```go
client := api.NewClient()

// 启用缓存
client.EnableCache()

// 自定义 TTL（默认见下方说明）
client.SetCacheTTL(10 * time.Minute)

// 查询缓存状态
client.IsCacheEnabled()   // true
client.GetCacheTTL()      // 10m0s

// 清空缓存
client.ClearCache()

// 关闭缓存
client.DisableCache()
```

## 重试

底层 HTTP 客户端内置**指数退避重试**，自动应对：

- HTTP 429（Too Many Requests / 限流）
- HTTP 5xx（服务端错误）
- 瞬时网络错误（连接超时、EOF）

重试会按指数增长间隔（如 1s → 2s → 4s ...）逐步退避，避免加重服务端压力。重试次数有上限，超过后返回最后一次错误。

::: tip 搜索 API 的限流
`search.maven.org` 对高频请求会返回 429。SDK 的重试机制让你无需手动处理——正常调用即可，被限流时自动等待重试。
:::

## 自定义 HTTP 客户端

如需更精细的控制（代理、自定义超时、TLS 配置），可替换底层 `*http.Client`：

```go
customHTTP := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        Proxy: http.ProxyFromEnvironment,
        // ...其他 transport 配置
    },
}

client := api.NewClient(api.WithHTTPClient(customHTTP))
```

Publisher 客户端同理：

```go
publisher := api.NewPublisherClient(
    api.WithPublisherToken(token),
    api.WithPublisherHTTPClient(customHTTP),
)
```

这也是**注入测试桩**（`httptest.Server` 返回的客户端）的入口，便于单元测试。
