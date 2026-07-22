package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// runVersion 版本相关子命令
func runVersion(args []string) error {
	fs := flag.NewFlagSet("version", flag.ExitOnError)
	var g globalOptions
	registerCommonFlags(fs, &g)

	subType := fs.String("type", "list", "版本操作：list | latest | count | has | filter")
	group := fs.String("group", "", "group id")
	artifact := fs.String("artifact", "", "artifact id")
	version := fs.String("version", "", "版本号（用于 has）")
	limit := fs.Int("limit", 20, "返回条数上限（list）")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *group == "" || *artifact == "" {
		return fmt.Errorf("version 需要 -group 与 -artifact")
	}

	client := newClient(&g)
	ctx := context.Background()

	switch *subType {
	case "list":
		results, err := client.ListVersions(ctx, *group, *artifact, *limit)
		return emitOrErr(results, err, &g)
	case "latest":
		result, err := client.GetLatestVersion(ctx, *group, *artifact)
		return emitOrErr(result, err, &g)
	case "count":
		n, err := client.CountVersions(ctx, *group, *artifact)
		if err != nil {
			return err
		}
		fmt.Printf("%d\n", n)
		return nil
	case "has":
		ok, err := client.HasVersion(ctx, *group, *artifact, *version)
		if err != nil {
			return err
		}
		fmt.Printf("%v\n", ok)
		return nil
	case "filter":
		results, err := client.FilterVersions(ctx, *group, *artifact, func(v *response.Version) bool { return true })
		return emitOrErr(results, err, &g)
	default:
		return fmt.Errorf("未知 version 类型 %q", *subType)
	}
}
