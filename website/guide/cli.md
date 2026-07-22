# 命令行工具（CLI）

SDK 提供一个零依赖的命令行工具 `sonatype-central`，封装了全部 Go SDK API，方便在脚本、CI 或终端中直接搜索、下载、发布 Maven 制品。所有输出默认为 JSON，便于管道处理。

## 安装

```bash
go install github.com/scagogogo/sonatype-central-sdk/cmd/sonatype-central@latest
```

或从源码构建：

```bash
git clone https://github.com/scagogogo/sonatype-central-skills.git
cd sonatype-central-skills
go build -o sonatype-central ./cmd/sonatype-central
```

## 设计理念

```mermaid
flowchart LR
    A[CLI flag 解析] --> B[构造 SDK Client]
    B --> C[调用 SDK 方法]
    C --> D[json.MarshalIndent]
    D --> E[stdout JSON]
```

CLI 只做"参数解析 → 调用 SDK → JSON 输出"的薄封装，**所有业务逻辑都在 SDK 内**，CLI 不重复实现。这保证了 CLI 与 SDK 行为完全一致。

## 通用选项

所有子命令共享以下选项：

| 选项 | 默认值 | 说明 |
|------|--------|------|
| `-base-url` | search.maven.org | 搜索 API base URL |
| `-repo-url` | repo1.maven.org | 仓库 base URL |
| `-max-retries` | 3 | 最大重试次数 |
| `-retry-backoff` | 500 | 重试退避毫秒 |
| `-cache` | true | 是否启用缓存 |
| `-cache-ttl` | 3600 | 缓存 TTL 秒 |
| `-proxy` | 空 | HTTP 代理 |
| `-output` | json | 输出格式：json 或 text |

## search — 搜索制品

按不同维度搜索，用 `-type` 选择搜索方式。

```bash
# 按 group 搜索
sonatype-central search -type group -query com.example -limit 10

# 按 artifact id 搜索
sonatype-central search -type artifact -query spring-core

# 按完整类名搜索
sonatype-central search -type fqcn -query org.springframework.context.ApplicationContext

# 按 SHA1 搜索
sonatype-central search -type sha1 -query a1b2c3...

# 按标签搜索
sonatype-central search -type tag -query spring

# 按依赖搜索（需要 -group -artifact）
sonatype-central search -type dependency -group org.slf4j -artifact slf4j-api

# 高级搜索
sonatype-central search -type advanced -group com.example -artifact my-lib
```

支持的 `-type` 值：`group`、`artifact`、`gav`、`class`、`fqcn`、`sha1`、`tag`、`text`、`dependency`、`classifier`、`package`、`interface`、`method`、`advanced`。

## version — 版本信息

```bash
# 列出所有版本
sonatype-central version -type list -group com.example -artifact my-lib -limit 20

# 最新版本
sonatype-central version -type latest -group com.example -artifact my-lib

# 版本数量
sonatype-central version -type count -group com.example -artifact my-lib

# 是否存在某版本
sonatype-central version -type has -group com.example -artifact my-lib -version 1.0.0
```

## download — 下载制品文件

```bash
# 下载 jar 到文件
sonatype-central download -type jar -group com.example -artifact my-lib -version 1.0.0 -out my-lib.jar

# 下载 POM（输出到 stdout）
sonatype-central download -type pom -group com.example -artifact my-lib -version 1.0.0

# 下载源码 jar
sonatype-central download -type sources -group com.example -artifact my-lib -version 1.0.0 -out my-lib-sources.jar

# 下载 SBOM（CycloneDX JSON）
sonatype-central download -type cyclonedx-json -group com.example -artifact my-lib -version 1.0.0

# 下载校验和
sonatype-central download -type checksum -path com/example/my-lib/1.0.0/my-lib-1.0.0.jar.sha1 -checksum-type sha1

# 按完整路径下载任意文件
sonatype-central download -type file -path com/example/my-lib/1.0.0/my-lib-1.0.0.jar -out lib.jar
```

支持的 `-type` 值：`jar`、`pom`、`sources`、`javadoc`、`cyclonedx-json`、`cyclonedx-xml`、`spdx-json`、`checksum`、`bundle`、`file`。

## metadata — 元数据与统计

```bash
# 制品元数据
sonatype-central metadata -type artifact -group com.example -artifact my-lib -version 1.0.0

# 制品统计
sonatype-central metadata -type stats -group com.example -artifact my-lib

# 使用情况
sonatype-central metadata -type usage -group com.example -artifact my-lib -version 1.0.0 -limit 10

# group 信息
sonatype-central metadata -type group -group com.example

# 依赖关系
sonatype-central metadata -type dependencies -group com.example -artifact my-lib -version 1.0.0
```

## publisher — 发布操作（需 Token）

发布操作对接 Sonatype Publisher API，需要 Token 认证。Token 可通过 `-token` 传入或 `SONATYPE_PUBLISHER_TOKEN` 环境变量提供。

```bash
# 上传 bundle
sonatype-central publisher -type upload -bundle ./my-bundle.zip -name my-lib-1.0.0 -token $TOKEN

# 查询部署状态
sonatype-central publisher -type status -deployment-id abc123 -token $TOKEN

# 列出所有部署
sonatype-central publisher -type list -token $TOKEN

# 浏览部署内容
sonatype-central publisher -type browse -deployment-id abc123 -token $TOKEN

# 发布部署
sonatype-central publisher -type publish -deployment-id abc123 -token $TOKEN

# 丢弃部署
sonatype-central publisher -type drop -deployment-id abc123 -token $TOKEN

# 检查是否已发布
sonatype-central publisher -type check -group com.example -artifact my-lib -version 1.0.0 -token $TOKEN
```

## 在脚本中使用

```bash
# 获取某制品最新版本号
LATEST=$(sonatype-central version -type latest -group com.example -artifact my-lib | jq -r '.version')
echo "最新版本: $LATEST"

# 下载最新版本
sonatype-central download -type jar -group com.example -artifact my-lib -version $LATEST -out my-lib.jar
```

::: tip 全部 API 已覆盖
CLI 的 5 个子命令覆盖了 SDK 的全部公开 API：`search` 覆盖所有 `Search*` 与 `Iterator*` 入口，`download` 覆盖所有 `Download*` 方法，`version` 覆盖版本相关方法，`metadata` 覆盖元数据与统计方法，`publisher` 覆盖全部 `PublisherClient` 方法。
:::