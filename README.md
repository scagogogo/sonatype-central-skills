# Sonatype Central SDK for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/scagogogo/sonatype-central-sdk.svg)](https://pkg.go.dev/github.com/scagogogo/sonatype-central-sdk)
[![Build Status](https://github.com/scagogogo/sonatype-central-sdk/actions/workflows/go-test.yml/badge.svg)](https://github.com/scagogogo/sonatype-central-sdk/actions/workflows/go-test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/scagogogo/sonatype-central-sdk)](https://goreportcard.com/report/github.com/scagogogo/sonatype-central-sdk)
[![License: MIT](https://img.shields.io/github/license/scagogogo/sonatype-central-sdk)](https://github.com/scagogogo/sonatype-central-sdk/blob/main/LICENSE)

**[简体中文](README_zh-CN.md)**

A comprehensive, type-safe Go SDK for the [Sonatype Central Repository API](https://central.sonatype.org/search/rest-api-guide/). Search, download, and publish Maven artifacts with a clean and idiomatic Go interface.

---

## Features

- **🔍 Search** — Search by GroupId, ArtifactId, version, SHA1, class name, fully qualified class name, tags, packaging, classifier, and full text
- **📦 Download** — Download POM, JAR, sources, javadoc, SBOM (CycloneDX/SPDX), and other artifact files
- **✅ Checksum Verification** — Download with official SHA1/MD5/SHA256 checksum verification from Maven Central
- **🏷️ Version** — List versions, get latest version, compare versions, filter versions
- **📂 Group** — Search groups, get group statistics, compare groups, search subgroups
- **🎯 GAV** — Search by GAV coordinates, paginated listings, sorted queries
- **🔖 Tag** — Search by tags, multiple tag search, tag prefix search
- **🔬 Class Search** — Search by class name, fully qualified class name, package, class hierarchy, interface implementations (with highlighting support)
- **📄 Artifact** — Search artifacts, get artifact details, compare artifacts, get artifact statistics
- **🏷️ Classifier** — Search by classifier (sources, javadoc, tests, etc.)
- **🔤 Spellcheck** — Get spelling suggestions for search queries
- **🚀 Publisher** — Upload deployment bundles, check deployment status, list deployments with filtering, publish to Maven Central
- **⚡ Batch Operations** — Batch search, batch download, concurrent processing
- **🔁 Iterators** — Memory-efficient lazy-loading iterators for large result sets
- **💾 Caching & Retry** — Built-in cache support and exponential backoff retry mechanism
- **🎛️ Advanced Search** — Exact match, custom field lists, query parser selection (dismax/edismax), query field boosting

## Installation

```bash
go get github.com/scagogogo/sonatype-central-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    "github.com/scagogogo/sonatype-central-sdk/pkg/api"
)

func main() {
    client := api.NewClient()

    // Search for artifacts by GroupId
    artifacts, err := client.SearchByGroupId(context.Background(), "org.apache.commons", 10)
    if err != nil {
        panic(err)
    }

    for _, artifact := range artifacts {
        fmt.Printf("GroupId: %s, ArtifactId: %s, Latest Version: %s\n",
            artifact.GroupId, artifact.ArtifactId, artifact.LatestVersion)
    }
}
```

## Publisher API

The Publisher API allows you to upload and publish artifacts to Maven Central. It requires a Sonatype API token for authentication.

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

    // Check if a component is already published
    published, err := client.CheckPublished(context.Background(),
        "com.example", "my-library", "1.0.0")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Published: %v\n", published.Published)

    // List deployments with filtering and pagination
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
        fmt.Printf("Deployment: %s, State: %s\n", d.DeploymentName, d.DeploymentState)
    }
}
```

### Publisher API Methods

| Method | Description |
|--------|-------------|
| `UploadBundle` | Upload a deployment bundle (ZIP) |
| `GetDeploymentStatus` | Retrieve deployment status by ID |
| `CheckPublished` | Check if a component (namespace/name/version) is published |
| `ListDeployments` | List deployments with filtering and pagination |
| `BrowseDeployment` | Browse files in a deployment (convenience method) |
| `BrowseDeploymentWithOptions` | Browse files with full options (sortField, pagination, path filter) |
| `DownloadDeploymentFile` | Download a specific file from a deployment |
| `DropDeployment` | Drop a deployment (FAILED/VALIDATED only) |
| `PublishDeployment` | Publish a deployment (VALIDATED only) |

## Advanced Search

### Exact Match

```go
searchRequest := request.NewSearchRequest().
    SetQuery(request.NewQuery().SetGroupId("org.apache.commons")).
    SetExact(true)  // Enable exact matching
```

### Spellcheck

```go
// Search with spelling suggestions
artifacts, suggestions, err := client.SearchWithSpellcheck(ctx, "commns-lang", 10, 5)
if len(suggestions) > 0 {
    fmt.Printf("Did you mean: %v?\n", suggestions)
}
```

### Custom Field List

```go
// Only return specific fields to reduce response size
searchRequest := request.NewSearchRequest().
    SetQuery(request.NewQuery().SetText("commons")).
    SetFieldList("id,g,a,latestVersion")
```

### Query Parser and Field Boosting

```go
// Use edismax parser with custom field weights
searchRequest := request.NewSearchRequest().
    SetQuery(request.NewQuery().SetText("json parser")).
    SetDefType("edismax").
    SetQueryFields("text^20 g^5 a^10 c^3 fc^2")
```

### Classifier Search

```go
// Search for artifacts with a specific classifier
artifacts, err := client.SearchByClassifier(ctx, "sources", 10)

// Combine with GroupId
artifacts, err := client.SearchByGroupAndClassifier(ctx, "org.apache.commons", "sources", 10)
```

## Download API

### SBOM Download

```go
// Download CycloneDX JSON SBOM
sbom, err := client.DownloadCycloneDXJSON(ctx, "com.example", "my-lib", "1.0.0")

// Download CycloneDX XML SBOM
sbom, err := client.DownloadCycloneDXXML(ctx, "com.example", "my-lib", "1.0.0")

// Download SPDX JSON SBOM
sbom, err := client.DownloadSpdxJSON(ctx, "com.example", "my-lib", "1.0.0")
```

### Checksum Verification

```go
// Download with official checksum verification
data, checksum, err := client.DownloadWithVerifiedChecksum(ctx,
    "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar",
    "sha1")

// Download only the official checksum file
checksum, err := client.DownloadChecksumFile(ctx,
    "org/apache/commons/commons-lang3/3.12.0/commons-lang3-3.12.0.jar",
    "sha256")
```

## API Reference

### Base URLs

| API | Base URL | Auth Required |
|-----|----------|:------------:|
| Search | `https://search.maven.org/solrsearch/select` | No |
| Download | `https://repo1.maven.org/maven2` | No |
| Publisher | `https://central.sonatype.com/api/v1/publisher` | Yes |

### Supported Solr Query Fields

| Field | Description |
|-------|-------------|
| `g:` | GroupId |
| `a:` | ArtifactId |
| `v:` | Version |
| `p:` | Packaging (jar, pom, war, etc.) |
| `l:` | Classifier (sources, javadoc, tests, etc.) |
| `c:` | Class name (simple) |
| `fc:` | Fully qualified class name |
| `1:` | SHA-1 checksum |
| `tags:` | Tag search |
| `id:` | Artifact ID |
| `text:` | Full text search |

### Advanced Search Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `exact` | Enable exact matching | `false` |
| `spellcheck` | Enable spell checking | auto |
| `spellcheck.count` | Number of spelling suggestions | `5` |
| `fl` | Field list to return | Solr default |
| `defType` | Query parser type | `dismax` |
| `qf` | Query fields with boost weights | `text^20 g^5 a^10` |

### Deprecated APIs

> ⚠️ The following API endpoints are no longer available on Sonatype Central:

| API | Status | Alternative |
|-----|--------|-------------|
| Security / Vulnerability | `search.maven.org/api/security/*` returns 403 | Use OWASP Dependency-Check |
| License Search | Solr `l:` field (license) returns empty results | Parse POM files instead |
| Dependency Search | Solr `d:` field returns 400 | Download and parse POM files |
| Method Search | Solr `m:` field returns 400 | — |
| Facet / Aggregation | Solr facet functionality is disabled | — |

## Project Structure

```
sonatype-central-sdk/
├── pkg/
│   ├── api/           # API client and method implementations
│   ├── request/       # HTTP request builders
│   ├── response/      # Response type definitions
│   └── examples/      # Usage examples
├── .github/           # CI/CD workflows
├── docs/              # Additional documentation
├── go.mod
└── LICENSE
```

## Requirements

- Go 1.18+

## License

[MIT License](LICENSE)
