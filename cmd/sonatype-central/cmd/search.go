package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
)

// runSearch 解析 -type 与对应查询参数，调用 SDK 搜索方法
func runSearch(args []string) error {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	var g globalOptions
	registerCommonFlags(fs, &g)

	searchType := fs.String("type", "group", "搜索类型：group | artifact | gav | class | fqcn | sha1 | tag | text | dependency | classifier | package | interface | method | advanced")
	query := fs.String("query", "", "查询值（多数类型用此字段）")
	group := fs.String("group", "", "group id（用于 dependency/gav/advanced 类型）")
	artifact := fs.String("artifact", "", "artifact id（用于 dependency/gav 类型）")
	limit := fs.Int("limit", 20, "返回条数上限")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *query == "" && *group == "" {
		return fmt.Errorf("search 至少需要 -query 或 -group 之一")
	}

	client := newClient(&g)
	ctx := context.Background()

	switch *searchType {
	case "group":
		results, err := client.SearchByGroupId(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "artifact":
		results, err := client.SearchByArtifactId(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "gav":
		results, err := client.SearchByGroupAndArtifactId(ctx, *group, *artifact, *limit)
		return emitOrErr(results, err, &g)
	case "class":
		results, err := client.SearchByClassName(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "fqcn":
		results, err := client.SearchByFullyQualifiedClassName(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "sha1":
		results, err := client.SearchBySha1(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "tag":
		results, err := client.SearchByTag(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "text":
		results, err := client.SearchByText(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "dependency":
		results, err := client.SearchByDependency(ctx, *group, *artifact, *limit)
		return emitOrErr(results, err, &g)
	case "classifier":
		results, err := client.SearchByClassifier(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "package":
		results, err := client.SearchByJavaPackage(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "interface":
		results, err := client.SearchInterfaceImplementations(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "method":
		results, err := client.SearchClassesByMethod(ctx, *query, *limit)
		return emitOrErr(results, err, &g)
	case "advanced":
		opts := request.NewAdvancedSearchOptions().
			SetGroupId(*group).
			SetArtifactId(*artifact)
		results, err := client.AdvancedSearch(ctx, opts, *limit)
		return emitOrErr(results, err, &g)
	default:
		return fmt.Errorf("未知 search 类型 %q", *searchType)
	}
}

// emitOrErr 统一处理输出或错误
func emitOrErr(result interface{}, err error, g *globalOptions) error {
	if err != nil {
		return err
	}
	if g.output == "json" {
		return printJSON(result)
	}
	fmt.Println(result)
	return nil
}