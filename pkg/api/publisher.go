package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// PublishingType 发布类型（从 response 包重导出以方便使用）
type PublishingType = response.PublishingType

const (
	PublishingTypeUserManaged = response.PublishingTypeUserManaged
	PublishingTypeAutomatic   = response.PublishingTypeAutomatic
)

// DeploymentState 部署状态（从 response 包重导出以方便使用）
type DeploymentState = response.DeploymentState

const (
	DeploymentStatePending    = response.DeploymentStatePending
	DeploymentStateValidating = response.DeploymentStateValidating
	DeploymentStateValidated  = response.DeploymentStateValidated
	DeploymentStatePublishing = response.DeploymentStatePublishing
	DeploymentStatePublished  = response.DeploymentStatePublished
	DeploymentStateFailed     = response.DeploymentStateFailed
)

// PublisherClient Sonatype Central Publisher API 客户端
//
// PublisherClient 封装了 Sonatype Central Publisher API（https://central.sonatype.com/api/v1/publisher/*），
// 提供向 Maven Central 发布制品的完整功能。该 API 需要认证（Bearer Token 或 Basic Auth）。
//
// 主要功能:
//   - 上传发布包（UploadBundle）
//   - 查询部署状态（GetDeploymentStatus）
//   - 检查组件是否已发布（CheckPublished）
//   - 列出部署（ListDeployments）
//   - 浏览部署内容（BrowseDeployment）
//   - 下载部署中的文件（DownloadDeploymentFile）
//   - 删除部署（DropDeployment）
//   - 发布部署（PublishDeployment）
//
// 使用示例:
//
//	// 创建 Publisher 客户端
//	client := api.NewPublisherClient(
//	    api.WithPublisherToken("your-bearer-token"),
//	)
//
//	// 上传发布包
//	deploymentID, err := client.UploadBundle(ctx, bundleData, "my-component", api.PublishingTypeAutomatic)
//
//	// 查询部署状态
//	status, err := client.GetDeploymentStatus(ctx, deploymentID)
type PublisherClient struct {
	baseURL    string
	httpClient *http.Client
	authToken  string // Bearer token
	authUser   string // Basic auth username
	authPass   string // Basic auth password
}

// PublisherClientOption Publisher 客户端配置选项
type PublisherClientOption func(*PublisherClient)

// WithPublisherToken 设置 Bearer Token 认证
//
// 参数:
//   - token: Sonatype Central 的 API Bearer Token
//
// 使用示例:
//
//	client := api.NewPublisherClient(
//	    api.WithPublisherToken("your-bearer-token"),
//	)
func WithPublisherToken(token string) PublisherClientOption {
	return func(c *PublisherClient) {
		c.authToken = token
	}
}

// WithPublisherBasicAuth 设置 Basic Auth 认证
//
// 参数:
//   - username: Sonatype Central 用户名
//   - password: Sonatype Central 密码或 API Key
//
// 使用示例:
//
//	client := api.NewPublisherClient(
//	    api.WithPublisherBasicAuth("username", "password"),
//	)
func WithPublisherBasicAuth(username, password string) PublisherClientOption {
	return func(c *PublisherClient) {
		c.authUser = username
		c.authPass = password
	}
}

// WithPublisherBaseURL 设置自定义 Publisher API 基础 URL
//
// 参数:
//   - baseURL: Publisher API 的基础 URL，默认为 "https://central.sonatype.com"
func WithPublisherBaseURL(baseURL string) PublisherClientOption {
	return func(c *PublisherClient) {
		c.baseURL = baseURL
	}
}

// WithPublisherHTTPClient 设置自定义 HTTP 客户端
func WithPublisherHTTPClient(httpClient *http.Client) PublisherClientOption {
	return func(c *PublisherClient) {
		c.httpClient = httpClient
	}
}

// NewPublisherClient 创建新的 Publisher API 客户端
//
// 参数:
//   - options: 可变数量的 PublisherClientOption 配置函数
//
// 返回:
//   - *PublisherClient: 配置完成的 Publisher 客户端实例
//
// 使用示例:
//
//	// 使用 Bearer Token
//	client := api.NewPublisherClient(
//	    api.WithPublisherToken("your-token"),
//	)
//
//	// 使用 Basic Auth
//	client := api.NewPublisherClient(
//	    api.WithPublisherBasicAuth("user", "pass"),
//	)
func NewPublisherClient(options ...PublisherClientOption) *PublisherClient {
	client := &PublisherClient{
		baseURL:    "https://central.sonatype.com",
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}

	for _, option := range options {
		option(client)
	}

	return client
}

// hasExtension 检查文件名是否包含扩展名
func hasExtension(fileName string) bool {
	for i := len(fileName) - 1; i >= 0; i-- {
		if fileName[i] == '.' {
			return true
		}
		if fileName[i] == '/' || fileName[i] == '\\' {
			return false
		}
	}
	return false
}

// doRequest 执行 HTTP 请求，将 JSON 响应解析到 result 中。
//
// 如果 result 为 nil，则不解析响应体（仅用于无返回值的操作，如 DELETE）。
// 对于非 JSON 响应（如 text/plain），请使用 doRequestRaw。
func (pc *PublisherClient) doRequest(ctx context.Context, method, path string, body io.Reader, contentType string, result interface{}) error {
	respBody, err := pc.doRequestRaw(ctx, method, path, body, contentType)
	if err != nil {
		return err
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}

	return nil
}

// doRequestRaw 执行 HTTP 请求并返回原始响应体字节。
//
// 用于处理非 JSON 响应（如 /upload 返回的 text/plain 部署 ID）。
// 错误响应（HTTP >= 400）会被解析为 PublisherErrorResponse 或 APIError。
func (pc *PublisherClient) doRequestRaw(ctx context.Context, method, path string, body io.Reader, contentType string) ([]byte, error) {
	// 注意：不能使用 url.JoinPath，因为它会对 path 中的 '?' 进行编码，
	// 导致查询字符串被当作路径的一部分。这里手动拼接 base URL 和 path。
	targetURL := pc.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置认证
	if pc.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+pc.authToken)
	} else if pc.authUser != "" {
		req.SetBasicAuth(pc.authUser, pc.authPass)
	}

	req.Header.Set("User-Agent", "sonatype-central-sdk/1.0")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := pc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		// 尝试解析 Publisher API 的标准错误响应
		var errResp response.PublisherErrorResponse
		if jsonErr := json.Unmarshal(respBody, &errResp); jsonErr == nil && errResp.Message != "" {
			return nil, &errResp
		}
		return nil, &response.APIError{
			Code:    fmt.Sprintf("%d", resp.StatusCode),
			Message: fmt.Sprintf("Publisher API 错误 [%d]: %s", resp.StatusCode, string(respBody)),
		}
	}

	return respBody, nil
}

// UploadBundle 上传发布包到 Maven Central
//
// 该方法用于上传一个部署包（bundle）到 Sonatype Central，准备发布到 Maven Central。
// 部署包通常是一个包含 POM、JAR、签名文件等的 ZIP 或 JAR 文件。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - bundle: 部署包的二进制内容
//   - name: 部署包名称（可选，为空则使用文件名）
//   - publishingType: 发布类型（USER_MANAGED 或 AUTOMATIC）
//
// 返回:
//   - string: 部署 ID，用于后续查询状态和发布操作
//   - error: 上传过程中的错误
//
// 使用示例:
//
//	client := api.NewPublisherClient(api.WithPublisherToken("your-token"))
//	ctx := context.Background()
//
//	bundleData, _ := os.ReadFile("my-component-bundle.zip")
//	deploymentID, err := client.UploadBundle(ctx, bundleData, "my-component", api.PublishingTypeAutomatic)
//	if err != nil {
//	    log.Fatalf("上传失败: %v", err)
//	}
//	fmt.Printf("部署 ID: %s\n", deploymentID)
func (pc *PublisherClient) UploadBundle(ctx context.Context, bundle []byte, name string, publishingType response.PublishingType) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加 bundle 文件
	fileName := name
	if fileName == "" {
		fileName = "bundle"
	}
	if !hasExtension(fileName) {
		fileName += ".zip"
	}
	part, err := writer.CreateFormFile("bundle", fileName)
	if err != nil {
		return "", fmt.Errorf("创建 multipart 表单失败: %w", err)
	}
	if _, err := part.Write(bundle); err != nil {
		return "", fmt.Errorf("写入 bundle 数据失败: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("关闭 multipart writer 失败: %w", err)
	}

	// 构建 URL 查询参数
	params := url.Values{}
	if name != "" {
		params.Set("name", name)
	}
	if publishingType != "" {
		params.Set("publishingType", string(publishingType))
	}
	path := "/api/v1/publisher/upload"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	// 根据 API 规范，/upload 返回 text/plain，内容为部署 ID 字符串
	// （可能带引号包围），因此使用 doRequestRaw 读取原始响应体
	respBody, err := pc.doRequestRaw(ctx, "POST", path, &buf, writer.FormDataContentType())
	if err != nil {
		return "", err
	}

	// 去除首尾的空白和可能的引号
	deploymentID := strings.TrimSpace(string(respBody))
	deploymentID = strings.Trim(deploymentID, `"`)

	return deploymentID, nil
}

// GetDeploymentStatus 查询部署状态
//
// 轮询此端点可以确定部署何时更改状态。
// 根据 API 规范，部署 ID 通过 id 查询参数传递。
//
// 参数:
//   - ctx: 上下文对象
//   - deploymentID: 部署 ID（从 UploadBundle 获取）
//
// 返回:
//   - *response.DeploymentStatus: 部署状态信息
//   - error: 查询过程中的错误
func (pc *PublisherClient) GetDeploymentStatus(ctx context.Context, deploymentID string) (*response.DeploymentStatus, error) {
	params := url.Values{}
	params.Set("id", deploymentID)
	path := "/api/v1/publisher/status?" + params.Encode()

	var result response.DeploymentStatus
	if err := pc.doRequest(ctx, "POST", path, nil, "", &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CheckPublished 检查组件是否已在 Maven Central 发布
//
// 根据 API 规范，使用 namespace、name、version 查询参数。
// 为保持调用方友好，参数名沿用 groupId/artifactId 的概念
// （namespace 对应 groupId，name 对应 artifactId）。
//
// 参数:
//   - ctx: 上下文对象
//   - groupID: 组 ID（对应 API 的 namespace 参数）
//   - artifactID: 制品 ID（对应 API 的 name 参数）
//   - version: 版本号
//
// 返回:
//   - *response.PublishedCheck: 发布状态
//   - error: 查询过程中的错误
func (pc *PublisherClient) CheckPublished(ctx context.Context, groupID, artifactID, version string) (*response.PublishedCheck, error) {
	params := url.Values{}
	params.Set("namespace", groupID)
	params.Set("name", artifactID)
	params.Set("version", version)
	path := "/api/v1/publisher/published?" + params.Encode()

	var result response.PublishedCheck
	if err := pc.doRequest(ctx, "GET", path, nil, "", &result); err != nil {
		return nil, err
	}

	// 填充回请求的坐标信息（API 响应只包含 published 字段）
	result.Namespace = groupID
	result.Name = artifactID
	result.Version = version

	return &result, nil
}

// ListDeployments 列出当前用户的部署
//
// 支持按命名空间、名称、状态等条件过滤，以及分页和排序。
//
// 参数:
//   - ctx: 上下文对象
//   - options: 查询选项，可以为 nil（列出所有部署）
//
// 返回:
//   - *response.DeploymentList: 部署列表
//   - error: 查询过程中的错误
//
// 使用示例:
//
//	// 列出所有部署
//	deployments, err := client.ListDeployments(ctx, nil)
//
//	// 按状态过滤
//	deployments, err := client.ListDeployments(ctx, &response.DeploymentListOptions{
//	    State: response.DeploymentStateValidated,
//	})
//
//	// 按命名空间和名称过滤，并分页
//	deployments, err := client.ListDeployments(ctx, &response.DeploymentListOptions{
//	    Namespace:     "com.example",
//	    DeploymentName: "my-lib",
//	    Page:          0,
//	    Size:          20,
//	})
func (pc *PublisherClient) ListDeployments(ctx context.Context, options *response.DeploymentListOptions) (*response.DeploymentList, error) {
	path := "/api/v1/publisher/deployments"

	if options != nil {
		params := url.Values{}
		if options.Namespace != "" {
			params.Set("namespace", options.Namespace)
		}
		if options.DeploymentName != "" {
			params.Set("deploymentName", options.DeploymentName)
		}
		if options.State != "" {
			params.Set("deploymentState", string(options.State))
		}
		if options.Paginate {
			params.Set("page", fmt.Sprintf("%d", options.Page))
			params.Set("size", fmt.Sprintf("%d", options.Size))
		}
		if options.SortField != "" {
			params.Set("sortField", options.SortField)
		}
		if options.SortDirection != "" {
			params.Set("sortDirection", options.SortDirection)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	var result response.DeploymentList
	if err := pc.doRequest(ctx, "GET", path, nil, "", &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BrowseDeployment 浏览部署包中的文件内容（便捷方法）
//
// 这是 BrowseDeploymentWithOptions 的便捷封装，使用默认参数浏览单个部署。
//
// 参数:
//   - ctx: 上下文对象
//   - deploymentID: 部署 ID
//
// 返回:
//   - *response.DeploymentResponseFiles: 部署包的文件内容信息
//   - error: 查询过程中的错误
func (pc *PublisherClient) BrowseDeployment(ctx context.Context, deploymentID string) (*response.DeploymentResponseFiles, error) {
	results, err := pc.BrowseDeploymentWithOptions(ctx, &response.BrowseDeploymentRequest{
		DeploymentIds: []string{deploymentID},
		Page:          0,
		Size:          500,
		SortField:     "createdTimestamp",
		SortDirection: "desc",
	})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("未找到部署 %s 的文件信息", deploymentID)
	}
	return &results[0], nil
}

// BrowseDeploymentWithOptions 浏览部署包文件内容（完整选项）
//
// 根据 API 规范，此端点需要发送 JSON 请求体，包含分页、排序和过滤参数。
// sortField 为必填字段。
//
// 参数:
//   - ctx: 上下文对象
//   - req: 浏览请求选项
//
// 返回:
//   - []response.DeploymentResponseFiles: 部署文件信息列表
//   - error: 查询过程中的错误
//
// 使用示例:
//
//	results, err := client.BrowseDeploymentWithOptions(ctx, &response.BrowseDeploymentRequest{
//	    DeploymentIds: []string{"dep-id-1", "dep-id-2"},
//	    Page:          0,
//	    Size:          100,
//	    SortField:     "createdTimestamp",
//	    SortDirection: "desc",
//	    PathStarting:  "org/sonatype/",
//	})
func (pc *PublisherClient) BrowseDeploymentWithOptions(ctx context.Context, req *response.BrowseDeploymentRequest) ([]response.DeploymentResponseFiles, error) {
	if req == nil {
		return nil, fmt.Errorf("请求选项不能为空")
	}
	if req.SortField == "" {
		return nil, fmt.Errorf("sortField 为必填字段")
	}

	path := "/api/v1/publisher/deployments/files"

	bodyData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	var result struct {
		Deployments []response.DeploymentResponseFiles `json:"deployments"`
		Page        int                                 `json:"page,omitempty"`
		PageSize    int                                 `json:"pageSize,omitempty"`
	}
	if err := pc.doRequest(ctx, "POST", path, bytes.NewReader(bodyData), "application/json", &result); err != nil {
		return nil, err
	}

	return result.Deployments, nil
}

// DownloadDeploymentFile 从部署包中下载文件
//
// 参数:
//   - ctx: 上下文对象
//   - deploymentID: 部署 ID
//   - relativePath: 文件在部署包中的相对路径
//
// 返回:
//   - []byte: 文件内容
//   - error: 下载过程中的错误
func (pc *PublisherClient) DownloadDeploymentFile(ctx context.Context, deploymentID, relativePath string) ([]byte, error) {
	path := fmt.Sprintf("/api/v1/publisher/deployment/%s/download/%s",
		url.PathEscape(deploymentID), url.PathEscape(relativePath))

	// 直接下载原始文件内容，不解析 JSON
	return pc.doRequestRaw(ctx, "GET", path, nil, "")
}

// DropDeployment 删除部署
//
// 部署只能在 FAILED 或 VALIDATED 状态下被删除。
//
// 参数:
//   - ctx: 上下文对象
//   - deploymentID: 部署 ID
//
// 返回:
//   - error: 删除过程中的错误
func (pc *PublisherClient) DropDeployment(ctx context.Context, deploymentID string) error {
	path := fmt.Sprintf("/api/v1/publisher/deployment/%s", url.PathEscape(deploymentID))
	return pc.doRequest(ctx, "DELETE", path, nil, "", nil)
}

// PublishDeployment 发布部署到 Maven Central
//
// 部署只能在 VALIDATED 状态下被发布。
//
// 参数:
//   - ctx: 上下文对象
//   - deploymentID: 部署 ID
//
// 返回:
//   - error: 发布过程中的错误
func (pc *PublisherClient) PublishDeployment(ctx context.Context, deploymentID string) error {
	path := fmt.Sprintf("/api/v1/publisher/deployment/%s", url.PathEscape(deploymentID))
	return pc.doRequest(ctx, "POST", path, nil, "", nil)
}
