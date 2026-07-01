# 快速开始

本页带你 5 分钟跑通第一个搜索调用。

## 环境要求

- **Go 1.18+**（SDK 用到了泛型，低于 1.18 无法编译）

::: tip 检查 Go 版本
```bash
go version
# go version go1.21.x ...
```
:::

## 安装

在你的 Go 项目目录下执行：

```bash
go get github.com/scagogogo/sonatype-central-sdk
```

## 第一个搜索

创建 `main.go`：

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/scagogogo/sonatype-central-sdk/pkg/api"
)

func main() {
    client := api.NewClient()
    ctx := context.Background()

    // 搜索 org.apache.commons 下的制品，最多 10 条
    artifacts, err := client.SearchByGroupId(ctx, "org.apache.commons", 10)
    if err != nil {
        log.Fatalf("搜索失败: %v", err)
    }

    for _, a := range artifacts {
        fmt.Printf("GroupId: %s, ArtifactId: %s, 最新版本: %s\n",
            a.GroupId, a.ArtifactId, a.LatestVersion)
    }
}
```

运行：

```bash
go run main.go
```

你会看到类似输出：

```
GroupId: org.apache.commons, ArtifactId: commons-lang3, 最新版本: 3.13.0
GroupId: org.apache.commons, ArtifactId: commons-io, 最新版本: 2.15.1
...
```

**就这么简单——无需认证、无需 API Key。** Maven Central 的搜索和下载是公开的。

## 下载制品

下载一个 JAR 并用官方校验和验证完整性：

```go
data, checksum, err := client.DownloadWithVerifiedChecksum(ctx,
    "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar",
    "sha1",
)
if err != nil {
    log.Fatalf("下载或校验失败: %v", err)
}
fmt.Printf("下载 %d 字节，SHA1=%s ✓\n", len(data), checksum)
```

## 发布到 Maven Central（可选）

发布需要认证。先去 [central.sonatype.com](https://central.sonatype.com) 注册并申请 API Token，然后：

```go
publisher := api.NewPublisherClient(
    api.WithPublisherToken("your-sonatype-api-token"),
)

bundle, _ := os.ReadFile("my-lib-bundle.zip")
id, err := publisher.UploadBundle(ctx, bundle, "my-lib", api.PublishingTypeAutomatic)
if err != nil {
    log.Fatalf("上传失败: %v", err)
}
fmt.Printf("部署 ID: %s，等待校验...\n", id)

// 轮询状态（实际使用时建议带间隔的循环）
status, _ := publisher.GetDeploymentStatus(ctx, id)
fmt.Printf("当前状态: %s\n", status.DeploymentState)

// 校验通过后发布
if status.DeploymentState == api.DeploymentStateValidated {
    _ = publisher.PublishDeployment(ctx, id)
}
```

::: warning 发布前置条件
- `PublishDeployment` 只能在 `VALIDATED` 状态下调用
- `DropDeployment` 只能在 `FAILED` 或 `VALIDATED` 状态下调用
- bundle 必须是符合 Maven Central 要求的 ZIP/JAR（含 POM、签名等）
:::

## 下一步

- [搜索能力详解](./search) — 各种搜索方法
- [下载能力详解](./download) — POM/JAR/SBOM/校验和
- [发布流程详解](./publish) — 完整的发布状态机
- [AI Agent 接入](../ai-agent/) — 让 AI 帮你把上面这些写出来
