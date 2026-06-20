package response

// Publisher API 响应类型
// 基于 https://central.sonatype.com/swagger.json 定义

// DeploymentStatus 表示发布部署的状态
type DeploymentStatus struct {
	DeploymentID   string `json:"deploymentId"`
	Name           string `json:"name"`
	State          string `json:"state"`    // PENDING, VALIDATING, VALIDATED, PUBLISHING, PUBLISHED, FAILED
	GroupID        string `json:"groupId"`
	ArtifactID     string `json:"artifactId"`
	Version        string `json:"version"`
	PublishingType string `json:"publishingType"` // USER_MANAGED, AUTOMATIC
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
	Errors         []string `json:"errors,omitempty"`
}

// PublishedCheck 检查组件是否已发布
type PublishedCheck struct {
	Published  bool   `json:"published"`
	GroupID    string `json:"groupId"`
	ArtifactID string `json:"artifactId"`
	Version    string `json:"version,omitempty"`
}

// DeploymentInfo 部署信息摘要
type DeploymentInfo struct {
	DeploymentID   string `json:"deploymentId"`
	Name           string `json:"name"`
	State          string `json:"state"`
	PublishingType string `json:"publishingType"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

// DeploymentList 部署列表
type DeploymentList struct {
	Deployments []DeploymentInfo `json:"deployments"`
	Total       int              `json:"total"`
}

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
	Page int

	// Size 每页数量
	Size int

	// SortField 排序字段
	SortField string

	// SortDirection 排序方向（asc 或 desc）
	SortDirection string
}

// DeploymentFile 部署包中的文件信息
type DeploymentFile struct {
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
}

// DeploymentFilesList 部署包中文件列表
type DeploymentFilesList struct {
	Files []DeploymentFile `json:"files"`
}

// PublisherUploadResponse 上传部署包的响应
type PublisherUploadResponse struct {
	DeploymentID string `json:"deploymentId"`
	Message      string `json:"message,omitempty"`
}

// PublisherError 发布 API 的错误响应
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
