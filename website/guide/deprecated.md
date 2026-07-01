# 已弃用的 API

Sonatype Central 上有些曾经存在、但现在已不可用的 API。SDK 仍保留对应方法以维持向后兼容，但调用时不会返回有效数据。本章帮你识别并规避它们。

## 不可用的端点

| API | 现状 | SDK 中的对应方法 | 替代方案 |
|-----|------|-----------------|---------|
| 安全 / 漏洞 | `search.maven.org/api/security/*` 返回 **403** | `SearchBySecurity*` / `BatchSecurityScan` 等 | 用 [OWASP Dependency-Check](https://owasp.org/www-project-dependency-check/) |
| 许可证搜索 | Solr `l:` 字段（许可证）返回**空结果** | `SearchByLicense` / `SearchByLicenseType` | 解析 POM 文件中的 `<licenses>` |
| 依赖搜索 | Solr `d:` 字段返回 **400** | `SearchByDependency` | 下载并解析 POM 文件 |
| 方法搜索 | Solr `m:` 字段返回 **400** | `SearchClassesByMethod`（部分场景） | — |
| 聚合 / 统计 | Solr facet 功能**已禁用** | `SearchArtifactsWithFacets` 等 | — |

## 为什么保留这些方法

SDK 遵循**向后兼容**承诺：移除公开方法会破坏现有使用方的代码。因此这些方法保留，但：

- 调用安全类方法会返回错误或空结果
- 调用许可证/依赖类方法可能返回空数据

::: warning 不要依赖它们
新代码**不应**使用上述方法。它们在 Maven Central 上已无数据支撑。如果你的用例确实需要这些能力，请使用替代方案。
:::

## 替代方案详解

### 漏洞扫描

Maven Central 不再直接提供漏洞数据。推荐：

1. **OWASP Dependency-Check** —— 开源工具，含 NVD 数据库
2. **GitHub Advisory Database** —— `https://github.com/advisories`
3. **Snyk / Dependabot** —— 商业 SaaS

### 许可证信息

解析 POM 文件中的 `<licenses>` 节点：

```go
pom, _ := client.DownloadPom(ctx, "com.example", "my-lib", "1.0.0")
// 用 encoding/xml 解析 pom 字节，提取 licenses
```

### 依赖关系

同样解析 POM 中的 `<dependencies>`：

```go
pom, _ := client.DownloadPom(ctx, "com.example", "my-lib", "1.0.0")
// 解析 <dependencies> 节点得到依赖列表
```

SDK 也有 `GetArtifactDependencies` 等方法，但其底层数据源同样依赖 POM 解析（而非已失效的 Solr `d:` 字段），可作为便捷封装使用。

## 检测代码中的弃用用法

可以用 `grep` 快速检查项目是否还在用这些方法：

```bash
grep -rn "SearchByLicense\|SearchByDependency\|SearchBySecurity\|WithFacets" --include="*.go" .
```

如有命中，参考上表迁移。
