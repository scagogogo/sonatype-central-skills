# 下载

SDK 提供丰富的下载能力：POM、JAR、源码、javadoc、SBOM，并支持官方校验和验证。

下载能力通过 `api.Client` 提供，**无需认证**。

## 基础下载

所有下载方法都基于 Maven Central 的路径约定：`<groupId 转斜杠>/<artifactId>/<version>/<artifactId>-<version>[-<classifier>].<ext>`

SDK 内部自动处理 GroupId 的 `.` → `/` 转换和路径拼接：

```go
client := api.NewClient()
ctx := context.Background()

// 下载任意路径的文件（传完整相对路径）
data, _ := client.Download(ctx,
    "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar")

// 按 GAV + 类型快捷下载
pom, _ := client.DownloadPom(ctx, "org.apache.commons", "commons-lang3", "3.12.0")
jar, _ := client.DownloadJar(ctx, "org.apache.commons", "commons-lang3", "3.12.0")
sources, _ := client.DownloadSources(ctx, "org.apache.commons", "commons-lang3", "3.12.0")
javadoc, _ := client.DownloadJavadoc(ctx, "org.apache.commons", "commons-lang3", "3.12.0")
```

## 下载到 io.Writer

直接写入文件或缓冲区，避免大文件占用内存：

```go
f, _ := os.Create("commons-lang3.jar")
defer f.Close()
_ = client.DownloadToWriter(ctx,
    "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar", f)
```

## 校验和验证

Maven Central 为每个文件提供官方的 SHA1/MD5/SHA256 校验和（`.sha1` / `.md5` / `.sha256` 文件）。下载后验证可保证文件完整性、防篡改。

### 仅下载校验和文件

```go
checksum, _ := client.DownloadChecksumFile(ctx,
    "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar",
    "sha256")
```

### 下载文件并自动验证

一行完成"下载文件 + 下载校验和 + 本地计算哈希 + 比对"四步：

```go
data, checksum, err := client.DownloadWithVerifiedChecksum(ctx,
    "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar",
    "sha1")
if err != nil {
    // err 会告诉你校验失败的原因
    log.Fatal(err)
}
// data 已通过校验，可安全使用
fmt.Printf("下载 %d 字节，校验通过，SHA1=%s\n", len(data), checksum)
```

::: tip 选哪种哈希？
- `sha1` —— 最常用，Maven Central 默认提供
- `sha256` —— 更安全，推荐用于安全敏感场景
- `md5` —— 速度最快，但安全性最低
:::

## SBOM 下载

[SBOM](https://en.wikipedia.org/wiki/Software_Bill_of_Materials)（软件物料清单）描述制品的依赖关系。Maven Central 部分制品提供 CycloneDX 和 SPDX 格式：

```go
// CycloneDX JSON
sbom, _ := client.DownloadCycloneDXJSON(ctx, "com.example", "my-lib", "1.0.0")
// CycloneDX XML
sbom, _ := client.DownloadCycloneDXXML(ctx, "com.example", "my-lib", "1.0.0")
// SPDX JSON
sbom, _ := client.DownloadSpdxJSON(ctx, "com.example", "my-lib", "1.0.0")
```

## 批量下载

需要一次下载多个文件时，用批量方法（内部并发）：

```go
paths := []string{
    "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar",
    "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.pom",
}
results := client.BatchDownloadFiles(ctx, paths)
for path, data := range results {
    fmt.Printf("%s: %d 字节\n", path, len(data))
}
```

异步版本：

```go
future := client.AsyncDownload(ctx, path)
// ... 做其他事 ...
data, _ := future.Await()
```

更多批量与异步能力见 [批量操作与迭代器](./batch-iterator)。

## 从部署包下载文件

如果你已经通过 Publisher API 上传了部署包，在它发布前也可以浏览和下载其中的文件：

```go
publisher := api.NewPublisherClient(api.WithPublisherToken(token))
data, _ := publisher.DownloadDeploymentFile(ctx, deploymentID, "com/example/lib/1.0/lib-1.0.jar")
```

详见 [发布到 Maven Central](./publish)。
