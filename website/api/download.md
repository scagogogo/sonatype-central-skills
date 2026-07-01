# 下载 API 参考

`api.Client` 提供的下载相关方法。下载**无需认证**。

## 基础下载

| 方法 | 说明 |
|------|------|
| `Download(ctx, relativePath string) ([]byte, error)` | 按完整相对路径下载 |
| `DownloadFile(ctx, relativePath string) ([]byte, error)` | 同上的别名 |
| `DownloadToWriter(ctx, relativePath string, w io.Writer) error` | 下载到 Writer |
| `DownloadArtifact(ctx, groupId, artifactId, version string) ([]byte, error)` | 按 GAV 下载主制品 |
| `DownloadArtifactWithVersion(ctx, ...) ([]byte, error)` | 带版本的制品 |

## 按类型下载

| 方法 | 说明 |
|------|------|
| `DownloadPom(ctx, groupId, artifactId, version string) ([]byte, error)` | POM 文件 |
| `DownloadJar(ctx, groupId, artifactId, version string) ([]byte, error)` | JAR 文件 |
| `DownloadSources(ctx, groupId, artifactId, version string) ([]byte, error)` | 源码 JAR |
| `DownloadJavadoc(ctx, groupId, artifactId, version string) ([]byte, error)` | Javadoc JAR |

## 校验和

| 方法 | 说明 |
|------|------|
| `DownloadChecksumFile(ctx, relativePath, algorithm string) (string, error)` | 仅下载校验和文件 |
| `DownloadWithChecksum(ctx, relativePath, algorithm string) ([]byte, string, error)` | 下载文件+校验和（不自动验证） |
| `DownloadWithVerifiedChecksum(ctx, relativePath, algorithm string) ([]byte, string, error)` | 下载+自动验证完整性 |

`algorithm` 取值：`"sha1"`、`"sha256"`、`"md5"`。

## SBOM

| 方法 | 说明 |
|------|------|
| `DownloadCycloneDXJSON(ctx, groupId, artifactId, version string) ([]byte, error)` | CycloneDX JSON |
| `DownloadCycloneDXXML(ctx, groupId, artifactId, version string) ([]byte, error)` | CycloneDX XML |
| `DownloadSpdxJSON(ctx, groupId, artifactId, version string) ([]byte, error)` | SPDX JSON |

## 完整 bundle

| 方法 | 说明 |
|------|------|
| `DownloadCompleteBundle(ctx, groupId, artifactId, version string) ([]byte, error)` | 完整制品 bundle |

## 批量与异步

| 方法 | 说明 |
|------|------|
| `BatchDownloadFiles(ctx, paths []string) map[string][]byte` | 批量下载 |
| `DownloadMultipleFiles(ctx, paths []string) (map[string][]byte, error)` | 批量下载（带错误） |
| `BatchDownloadDependencies(ctx, gavs []string) map[string][]byte` | 批量下载依赖 |
| `AsyncDownload(ctx, path string) *Future` | 异步下载 |
| `AsyncBatchDownload(ctx, paths []string) *Future` | 异步批量下载 |
| `AsyncBatchSearch(ctx, reqs []*request.SearchRequest) *Future` | 异步批量搜索 |

## 路径约定

所有 `relativePath` 参数遵循 Maven Central 的路径规则：

```
<groupId 中的 . 替换为 />/<artifactId>/<version>/<artifactId>-<version>[-<classifier>].<ext>
```

例如 `org.apache.commons:commons-lang3:3.12.0` 的 JAR：

```
org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar
```

带分类器的源码包：

```
org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0-sources.jar
```
