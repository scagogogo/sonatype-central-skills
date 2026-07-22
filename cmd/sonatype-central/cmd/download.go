package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
)

// runDownload 下载制品文件
func runDownload(args []string) error {
	fs := flag.NewFlagSet("download", flag.ExitOnError)
	var g globalOptions
	registerCommonFlags(fs, &g)

	subType := fs.String("type", "jar", "下载类型：jar | pom | sources | javadoc | cyclonedx-json | cyclonedx-xml | spdx-json | checksum | bundle | file")
	group := fs.String("group", "", "group id")
	artifact := fs.String("artifact", "", "artifact id")
	version := fs.String("version", "", "版本号")
	filePath := fs.String("path", "", "完整文件路径（用于 -type file 与 checksum）")
	checksumType := fs.String("checksum-type", "sha1", "校验和类型（sha1/md5）")
	out := fs.String("out", "", "输出文件路径（默认 stdout 二进制）")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *group == "" || *artifact == "" || *version == "" {
		if *filePath == "" {
			return fmt.Errorf("download 需要 -group/-artifact/-version 或 -path")
		}
	}

	client := newClient(&g)
	ctx := context.Background()

	var data []byte
	var err error
	switch *subType {
	case "jar":
		data, err = client.DownloadJar(ctx, *group, *artifact, *version)
	case "pom":
		data, err = client.DownloadPom(ctx, *group, *artifact, *version)
	case "sources":
		data, err = client.DownloadSources(ctx, *group, *artifact, *version)
	case "javadoc":
		data, err = client.DownloadJavadoc(ctx, *group, *artifact, *version)
	case "cyclonedx-json":
		data, err = client.DownloadCycloneDXJSON(ctx, *group, *artifact, *version)
	case "cyclonedx-xml":
		data, err = client.DownloadCycloneDXXML(ctx, *group, *artifact, *version)
	case "spdx-json":
		data, err = client.DownloadSpdxJSON(ctx, *group, *artifact, *version)
	case "checksum":
		var sum string
		sum, err = client.DownloadChecksumFile(ctx, *filePath, *checksumType)
		if err == nil {
			fmt.Println(sum)
			return nil
		}
	case "bundle":
		bdl, bErr := client.DownloadCompleteBundle(ctx, *group, *artifact, *version)
		if bErr == nil {
			return emitOrErr(bdl, nil, &g)
		}
		err = bErr
	case "file":
		if *out != "" {
			err = client.DownloadFile(ctx, *filePath, *out)
			if err == nil {
				fmt.Printf("已下载到 %s\n", *out)
			}
			return err
		}
		data, err = client.Download(ctx, *filePath)
	default:
		return fmt.Errorf("未知 download 类型 %q", *subType)
	}
	if err != nil {
		return err
	}
	if *out != "" {
		return os.WriteFile(*out, data, 0644)
	}
	_, err = os.Stdout.Write(data)
	return err
}