# 批量操作与迭代器

面对成百上千条制品记录时，逐条调用既慢又耗内存。SDK 提供两类机制应对：**批量操作**和**迭代器**。

## 批量搜索

一次提交多个查询，内部并发执行：

```go
queries := []*request.SearchRequest{ /* ... */ }
results := client.BatchSearch(ctx, queries)
```

针对具体维度的批量方法：

```go
// 批量按 GroupId 搜索
results := client.BatchSearchArtifacts(ctx, groupIds, 10)
```

## 批量下载

并发下载多个文件：

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

批量下载依赖（自动解析 POM 拉取依赖 JAR）：

```go
deps := client.BatchDownloadDependencies(ctx, gavCoordinates)
```

## 异步操作

非阻塞调用，返回 Future，稍后 `Await` 取结果：

```go
future := client.AsyncDownload(ctx, path)
// ... 同时做其他事 ...
data, err := future.Await()
```

异步搜索系列：

```go
f1 := client.AsyncSearchByGroupId(ctx, "org.apache.commons", 10)
f2 := client.AsyncSearchByArtifactId(ctx, "commons-lang3", 10)
// 并发执行，分别 Await
a, _ := f1.Await()
b, _ := f2.Await()
```

## 迭代器

当结果集可能很大（如遍历某个大 Group 下所有制品），用迭代器**懒加载**，避免一次性把全部结果载入内存：

```go
iter := client.IteratorByGroupId(ctx, "org.apache.commons")
for iter.HasNext() {
    a, err := iter.Next()
    if err != nil {
        break
    }
    // 逐条处理，内存中只保留当前一条
}
```

### 可用的迭代器

| 方法 | 说明 |
|------|------|
| `IteratorByGroupId` | 遍历某 Group 下的制品 |
| `IteratorByArtifactId` | 遍历某 artifactId |
| `IteratorByGroupAndArtifactId` | G+A 组合遍历 |
| `IteratorByText` | 全文搜索遍历 |
| `IteratorBySha1` / `IteratorBySha1Prefix` | SHA1 遍历 |
| `IteratorByClassName` | 类名遍历 |
| `IteratorByFullyQualifiedClassName` | 全限定类名遍历 |
| `IteratorByJavaPackage` | Java 包遍历 |
| `IteratorByClassHierarchy` | 类层次结构 |
| `IteratorByInterfaceImplementation` | 接口实现 |
| `IteratorByMethod` | 方法名 |
| `IteratorByClassifier` | 分类器 |
| `IteratorByTag` | 标签 |
| `IteratorVersions` | 某制品的所有版本 |
| `IteratorGAVs` | GAV 坐标遍历 |

迭代器内部自动处理分页游标，你只需要写 `for` 循环。

## 何时用批量 vs 迭代器

- **批量**：你已经知道所有要操作的项（一组已知的 GAV/路径），想并发加速。
- **迭代器**：你不知道总量有多少，需要逐条处理直到结束，且关心内存占用。

两者也可以组合：用迭代器流式遍历，每凑够 N 条就批量下载。
