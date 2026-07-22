package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/scagogogo/sonatype-central-sdk/pkg/api"
)

// globalOptions 通用选项：所有子命令共享
type globalOptions struct {
	baseURL      string
	repoBaseURL  string
	maxRetries   int
	retryBackoff int
	cache        bool
	cacheTTL     int
	proxy        string
	output       string // json | text
}

// registerCommonFlags 在给定 FlagSet 上注册通用选项，返回指针
func registerCommonFlags(fs *flag.FlagSet, g *globalOptions) {
	fs.StringVar(&g.baseURL, "base-url", "", "覆盖搜索 API base URL（默认 search.maven.org）")
	fs.StringVar(&g.repoBaseURL, "repo-url", "", "覆盖仓库 base URL（默认 repo1.maven.org）")
	fs.IntVar(&g.maxRetries, "max-retries", 3, "最大重试次数")
	fs.IntVar(&g.retryBackoff, "retry-backoff", 500, "重试退避毫秒")
	fs.BoolVar(&g.cache, "cache", true, "是否启用缓存")
	fs.IntVar(&g.cacheTTL, "cache-ttl", 3600, "缓存 TTL 秒")
	fs.StringVar(&g.proxy, "proxy", "", "HTTP 代理地址")
	fs.StringVar(&g.output, "output", "json", "输出格式：json | text")
}

// newClient 根据通用选项构造 SDK Client
func newClient(g *globalOptions) *api.Client {
	opts := []api.ClientOption{
		api.WithMaxRetries(g.maxRetries),
		api.WithRetryBackoff(g.retryBackoff),
		api.WithCache(g.cache, g.cacheTTL),
	}
	if g.baseURL != "" {
		opts = append(opts, api.WithBaseURL(g.baseURL))
	}
	if g.repoBaseURL != "" {
		opts = append(opts, api.WithRepoBaseURL(g.repoBaseURL))
	}
	if g.proxy != "" {
		opts = append(opts, api.WithProxy(g.proxy))
	}
	return api.NewClient(opts...)
}

// printJSON 将结果以 JSON 输出到 stdout
func printJSON(v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = io.Writer(os.Stdout).Write(b)
	if err != nil {
		return err
	}
	fmt.Println()
	return nil
}

// usage 打印子命令用法
func usage(w io.Writer) {
	fmt.Fprintln(w, "sonatype-central — Sonatype Central Repository CLI")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "子命令：")
	fmt.Fprintln(w, "  search       搜索制品（按 group/artifact/class/sha1/tag/text 等）")
	fmt.Fprintln(w, "  download     下载制品文件（jar/pom/sources/javadoc/sbom 等）")
	fmt.Fprintln(w, "  version      查询版本信息（latest/list/filter/count）")
	fmt.Fprintln(w, "  metadata     查询制品元数据与统计")
	fmt.Fprintln(w, "  publisher    发布相关操作（需 Token，对接 Publisher API）")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "通用选项（所有子命令）：")
	fmt.Fprintln(w, "  -base-url, -repo-url, -max-retries, -retry-backoff,")
	fmt.Fprintln(w, "  -cache, -cache-ttl, -proxy, -output")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "示例：")
	fmt.Fprintln(w, "  sonatype-central search -type group -query com.example -limit 10")
	fmt.Fprintln(w, "  sonatype-central download -group com.example -artifact lib -version 1.0.0 -file jar")
}

// Execute 解析参数并分发到子命令
func Execute(args []string) error {
	if len(args) == 0 {
		usage(os.Stdout)
		return nil
	}
	switch args[0] {
	case "-h", "--help", "help":
		usage(os.Stdout)
		return nil
	case "search":
		return runSearch(args[1:])
	case "download":
		return runDownload(args[1:])
	case "version":
		return runVersion(args[1:])
	case "metadata":
		return runMetadata(args[1:])
	case "publisher":
		return runPublisher(args[1:])
	default:
		return fmt.Errorf("未知子命令 %q，使用 -h 查看帮助", args[0])
	}
}