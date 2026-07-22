package cmd

import (
	"context"
	"flag"
	"fmt"
)

// runMetadata 查询制品元数据与统计
func runMetadata(args []string) error {
	fs := flag.NewFlagSet("metadata", flag.ExitOnError)
	var g globalOptions
	registerCommonFlags(fs, &g)

	subType := fs.String("type", "artifact", "元数据类型：artifact | details | stats | usage | group | group-stats | dependencies | gav")
	group := fs.String("group", "", "group id")
	artifact := fs.String("artifact", "", "artifact id")
	version := fs.String("version", "", "版本号")
	limit := fs.Int("limit", 10, "usage 返回上限")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *group == "" {
		return fmt.Errorf("metadata 需要 -group")
	}

	client := newClient(&g)
	ctx := context.Background()

	switch *subType {
	case "artifact":
		result, err := client.GetArtifactMetadata(ctx, *group, *artifact, *version)
		return emitOrErr(result, err, &g)
	case "details":
		result, err := client.GetArtifactDetails(ctx, *group, *artifact, *version)
		return emitOrErr(result, err, &g)
	case "stats":
		result, err := client.GetArtifactStats(ctx, *group, *artifact)
		return emitOrErr(result, err, &g)
	case "usage":
		result, err := client.GetArtifactUsage(ctx, *group, *artifact, *version, *limit)
		return emitOrErr(result, err, &g)
	case "group":
		result, err := client.GetGroupInfo(ctx, *group)
		return emitOrErr(result, err, &g)
	case "group-stats":
		result, err := client.GetGroupStatistics(ctx, *group)
		return emitOrErr(result, err, &g)
	case "dependencies":
		result, err := client.GetArtifactDependencies(ctx, *group, *artifact, *version)
		return emitOrErr(result, err, &g)
	case "gav":
		result, err := client.GetGAVInfo(ctx, *group, *artifact, *version)
		return emitOrErr(result, err, &g)
	default:
		return fmt.Errorf("未知 metadata 类型 %q", *subType)
	}
}