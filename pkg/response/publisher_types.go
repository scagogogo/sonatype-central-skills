package response

import "fmt"

// Publisher API 响应类型
// 基于 https://central.sonatype.com/swagger.json (Central Publisher API) 定义

// PublisherErrorResponse 发布 API 的标准错误响应
//
// 当 Publisher API 请求失败时（HTTP 状态码 >= 400），返回此结构。
// 注意：这与 http_error_types.go 中的通用 ErrorResponse 不同，
// 后者是 Search API 使用的错误格式。
type PublisherErrorResponse struct {
	// HttpStatus HTTP 状态码
	HttpStatus int `json:"httpStatus"`

	// ErrorCode 应用特定的错误代码
	ErrorCode int `json:"errorCode"`

	// Message 人类可读的错误消息
	Message string `json:"message"`

	// Explanation 错误的附加说明（可能为空）
	Explanation string `json:"explanation,omitempty"`

	// Data 附加的错误数据或上下文
	Data interface{} `json:"data,omitempty"`
}

// Error 实现 error 接口
func (e *PublisherErrorResponse) Error() string {
	return fmt.Sprintf("Publisher API 错误 [%d/%d]: %s", e.HttpStatus, e.ErrorCode, e.Message)
}

// DeploymentStatus 表示发布部署的状态
//
// 由 /api/v1/publisher/status 端点返回。
type DeploymentStatus struct {
	// DeploymentID 部署 ID
	DeploymentID string `json:"deploymentId"`

	// DeploymentName 部署名称
	DeploymentName string `json:"deploymentName"`

	// DeploymentState 部署状态
	// PENDING, VALIDATING, VALIDATED, PUBLISHING, PUBLISHED, FAILED
	DeploymentState DeploymentState `json:"deploymentState"`

	// PublishingType 发布类型（USER_MANAGED 或 AUTOMATIC）
	PublishingType PublishingType `json:"publishingType,omitempty"`

	// Purls 部署包含的组件 PURL 列表
	Purls []string `json:"purls,omitempty"`

	// Errors 验证错误信息
	// API 可能返回为对象或数组，这里用 interface{} 兼容
	Errors interface{} `json:"errors,omitempty"`

	// CreateTimestamp 创建时间戳
	CreateTimestamp interface{} `json:"createTimestamp,omitempty"`

	// UpdateTimestamp 更新时间戳
	UpdateTimestamp interface{} `json:"updateTimestamp,omitempty"`
}

// PublishedCheck 检查组件是否已发布
//
// 由 /api/v1/publisher/published 端点返回。
type PublishedCheck struct {
	// Published 是否已发布
	Published bool `json:"published"`

	// Namespace 命名空间（即 groupId）
	Namespace string `json:"namespace,omitempty"`

	// Name 组件名称（即 artifactId）
	Name string `json:"name,omitempty"`

	// Version 版本号
	Version string `json:"version,omitempty"`
}

// DeploymentComponent 部署中的组件信息
type DeploymentComponent struct {
	// Purl 组件的 Package URL
	Purl string `json:"purl"`

	// Name 组件名称
	Name string `json:"name"`

	// Path 组件路径
	Path string `json:"path"`

	// Errors 组件错误列表
	Errors []string `json:"errors,omitempty"`

	// Warnings 组件警告列表
	Warnings []string `json:"warnings,omitempty"`
}

// DeploymentListItem 部署列表项
//
// 由 /api/v1/publisher/deployments 端点返回。
type DeploymentListItem struct {
	// DeploymentID 部署 ID
	DeploymentID string `json:"deploymentId"`

	// DeploymentName 部署名称
	DeploymentName string `json:"deploymentName"`

	// Namespace 命名空间（即 groupId）
	Namespace string `json:"namespace"`

	// DeploymentState 部署状态
	DeploymentState DeploymentState `json:"deploymentState"`

	// CreateTimestamp 创建时间戳
	CreateTimestamp string `json:"createTimestamp"`

	// UpdateTimestamp 更新时间戳
	UpdateTimestamp string `json:"updateTimestamp"`

	// DeploymentComponents 部署包含的组件列表
	DeploymentComponents []DeploymentComponent `json:"deploymentComponents,omitempty"`
}

// DeploymentList 部署列表
//
// 由 /api/v1/publisher/deployments 端点返回。
// 包含部署数组以及分页信息。
type DeploymentList struct {
	// Deployments 部署列表
	Deployments []DeploymentListItem `json:"deployments"`

	// Page 当前页码
	Page int `json:"page,omitempty"`

	// PageSize 每页数量
	PageSize int `json:"pageSize,omitempty"`

	// PageCount 总页数
	PageCount int `json:"pageCount,omitempty"`

	// TotalResultCount 总结果数
	TotalResultCount int `json:"totalResultCount,omitempty"`
}

// DeployedComponentVersion 已部署的组件版本
type DeployedComponentVersion struct {
	// Name 组件名称
	Name string `json:"name"`

	// Path 组件路径
	Path string `json:"path"`

	// Errors 错误列表
	Errors []string `json:"errors,omitempty"`

	// Warnings 警告列表
	Warnings []string `json:"warnings,omitempty"`
}

// DeploymentResponseFiles 部署文件浏览响应
//
// 由 /api/v1/publisher/deployments/files 端点返回。
type DeploymentResponseFiles struct {
	// DeploymentID 部署 ID
	DeploymentID string `json:"deploymentId"`

	// DeploymentName 部署名称
	DeploymentName string `json:"deploymentName"`

	// DeploymentState 部署状态
	DeploymentState DeploymentState `json:"deploymentState"`

	// DeploymentType 部署类型（BUNDLE 或 SINGLE）
	DeploymentType string `json:"deploymentType"`

	// CreateTimestamp 创建时间戳
	CreateTimestamp interface{} `json:"createTimestamp"`

	// Purls 部署包含的组件 PURL 列表
	Purls []string `json:"purls,omitempty"`

	// DeployedComponentVersions 已部署的组件版本列表
	DeployedComponentVersions []DeployedComponentVersion `json:"deployedComponentVersions,omitempty"`

	// DeploymentFiles 部署包含的文件列表（仅浏览时返回）
	DeploymentFiles []DeploymentFile `json:"deploymentFiles,omitempty"`
}

// DeploymentFile 部署包中的文件信息
type DeploymentFile struct {
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
}

// BrowseDeploymentRequest 浏览部署文件的请求体
//
// 用于 /api/v1/publisher/deployments/files 端点。
// sortField 为必填字段。
type BrowseDeploymentRequest struct {
	// Page 页码（从 0 开始）
	Page int `json:"page"`

	// Size 每页数量
	Size int `json:"size"`

	// SortField 排序字段（必填）
	SortField string `json:"sortField"`

	// SortDirection 排序方向（asc 或 desc）
	SortDirection string `json:"sortDirection,omitempty"`

	// DeploymentIds 部署 ID 列表（可选）
	DeploymentIds []string `json:"deploymentIds,omitempty"`

	// PathStarting 起始路径前缀（可选）
	PathStarting string `json:"pathStarting,omitempty"`
}

// PublisherUploadResponse 上传部署包的响应
type PublisherUploadResponse struct {
	DeploymentID string `json:"deploymentId"`
	Message      string `json:"message,omitempty"`
}

// PublisherError 发布 API 的错误响应（旧版，保留向后兼容）
//
// 推荐使用 ErrorResponse 类型。
type PublisherError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// PublishingType 发布类型
type PublishingType string

const (
	PublishingTypeUserManaged PublishingType = "USER_MANAGED"
	PublishingTypeAutomatic   PublishingType = "AUTOMATIC"
)

// DeploymentState 部署状态常量
type DeploymentState string

const (
	DeploymentStatePending    DeploymentState = "PENDING"
	DeploymentStateValidating DeploymentState = "VALIDATING"
	DeploymentStateValidated  DeploymentState = "VALIDATED"
	DeploymentStatePublishing DeploymentState = "PUBLISHING"
	DeploymentStatePublished  DeploymentState = "PUBLISHED"
	DeploymentStateFailed     DeploymentState = "FAILED"
)

// DeploymentListOptions 部署列表查询选项
//
// 用于 ListDeployments 方法的过滤和分页参数。
// 所有字段均为可选，为零值时表示不进行对应过滤。
type DeploymentListOptions struct {
	// Namespace 按 groupId (namespace) 精确匹配过滤
	Namespace string

	// DeploymentName 按部署名称模糊搜索（不区分大小写的子串匹配）
	DeploymentName string

	// State 按部署状态过滤（VALIDATING, VALIDATED, PUBLISHING, PUBLISHED, FAILED）
	State DeploymentState

	// Page 页码（从 0 开始）
	// 注意：当 Paginate 为 true 时才会发送此参数
	Page int

	// Size 每页数量
	// 注意：当 Paginate 为 true 时才会发送此参数
	Size int

	// Paginate 是否启用分页
	// 为 true 时发送 page 和 size 参数；为 false 时不发送（API 使用默认值）
	Paginate bool

	// SortField 排序字段
	SortField string

	// SortDirection 排序方向（asc 或 desc）
	SortDirection string
}
