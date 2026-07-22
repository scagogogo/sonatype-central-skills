package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/scagogogo/sonatype-central-sdk/pkg/api"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// runPublisher 发布相关操作（需 Token）
func runPublisher(args []string) error {
	fs := flag.NewFlagSet("publisher", flag.ExitOnError)
	var g globalOptions
	registerCommonFlags(fs, &g)

	subType := fs.String("type", "status", "操作：upload | status | list | browse | publish | drop | check | download-file")
	token := fs.String("token", "", "Publisher Token（也可用 SONATYPE_PUBLISHER_TOKEN 环境变量）")
	username := fs.String("username", "", "Basic Auth 用户名")
	password := fs.String("password", "", "Basic Auth 密码")
	deploymentID := fs.String("deployment-id", "", "部署 ID")
	bundlePath := fs.String("bundle", "", "要上传的 bundle 文件路径")
	name := fs.String("name", "", "上传 bundle 的名称")
	publishingType := fs.String("publishing-type", "USER_MANAGED_BY_API_BOT", "发布类型")
	group := fs.String("group", "", "group id（用于 check）")
	artifact := fs.String("artifact", "", "artifact id（用于 check）")
	version := fs.String("version", "", "版本号（用于 check）")
	relativePath := fs.String("path", "", "部署内相对路径（用于 download-file）")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *token == "" {
		*token = os.Getenv("SONATYPE_PUBLISHER_TOKEN")
	}

	var pc *api.PublisherClient
	if *token != "" {
		pc = api.NewPublisherClient(api.WithPublisherToken(*token))
	} else if *username != "" {
		pc = api.NewPublisherClient(api.WithPublisherBasicAuth(*username, *password))
	} else {
		return fmt.Errorf("publisher 需要 -token 或 -username/-password")
	}

	ctx := context.Background()

	switch *subType {
	case "upload":
		data, err := os.ReadFile(*bundlePath)
		if err != nil {
			return err
		}
		id, err := pc.UploadBundle(ctx, data, *name, response.PublishingType(*publishingType))
		if err != nil {
			return err
		}
		fmt.Println(id)
		return nil
	case "status":
		result, err := pc.GetDeploymentStatus(ctx, *deploymentID)
		return emitOrErr(result, err, &g)
	case "list":
		result, err := pc.ListDeployments(ctx, nil)
		return emitOrErr(result, err, &g)
	case "browse":
		result, err := pc.BrowseDeployment(ctx, *deploymentID)
		return emitOrErr(result, err, &g)
	case "publish":
		return pc.PublishDeployment(ctx, *deploymentID)
	case "drop":
		return pc.DropDeployment(ctx, *deploymentID)
	case "check":
		result, err := pc.CheckPublished(ctx, *group, *artifact, *version)
		return emitOrErr(result, err, &g)
	case "download-file":
		data, err := pc.DownloadDeploymentFile(ctx, *deploymentID, *relativePath)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(data)
		return err
	default:
		return fmt.Errorf("未知 publisher 类型 %q", *subType)
	}
}