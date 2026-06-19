# Sonatype Central SDK for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/scagogogo/sonatype-central-sdk.svg)](https://pkg.go.dev/github.com/scagogogo/sonatype-central-sdk)
[![Build Status](https://github.com/scagogogo/sonatype-central-sdk/actions/workflows/go-test.yml/badge.svg)](https://github.com/scagogogo/sonatype-central-sdk/actions/workflows/go-test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/scagogogo/sonatype-central-sdk)](https://goreportcard.com/report/github.com/scagogogo/sonatype-central-sdk)
[![License: MIT](https://img.shields.io/github/license/scagogogo/sonatype-central-sdk)](https://github.com/scagogogo/sonatype-central-sdk/blob/main/LICENSE)

**[简体中文](README_zh-CN.md)**

A comprehensive, type-safe Go SDK for the [Sonatype Central Repository API](https://central.sonatype.org/search/rest-api-guide/). Search, download, and publish Maven artifacts with a clean and idiomatic Go interface.

---

## Features

- **🔍 Search** — Search by GroupId, ArtifactId, version, SHA1, class name, fully qualified class name, tags, packaging, and full text
- **📦 Download** — Download POM, JAR, sources, javadoc, and other artifact files
- **🏷️ Version** — List versions, get latest version, compare versions, filter versions
- **📂 Group** — Search groups, get group statistics, compare groups, search subgroups
- **🎯 GAV** — Search by GAV coordinates, paginated listings, sorted queries
- **🔖 Tag** — Search by tags, multiple tag search, tag prefix search
- **🔬 Class Search** — Search by class name, fully qualified class name, package, class hierarchy, interface implementations (with highlighting support)
- **📄 Artifact** — Search artifacts, get artifact details, compare artifacts, get artifact statistics
- **🚀 Publisher** — Upload deployment bundles, check deployment status, publish to Maven Central
- **⚡ Batch Operations** — Batch search, batch download, concurrent processing
- **🔁 Iterators** — Memory-efficient lazy-loading iterators for large result sets
- **💾 Caching & Retry** — Built-in cache support and exponential backoff retry mechanism

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
}
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
| `c:` | Class name (simple) |
| `fc:` | Fully qualified class name |
| `1:` | SHA-1 checksum |
| `tags:` | Tag search |
| `id:` | Artifact ID |
| `text:` | Full text search |

### Deprecated APIs

> ⚠️ The following API endpoints are no longer available on Sonatype Central:

| API | Status | Alternative |
|-----|--------|-------------|
| Security / Vulnerability | `search.maven.org/api/security/*` returns 403 | Use OWASP Dependency-Check |
| License Search | Solr `l:` field returns empty results | Parse POM files instead |
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
