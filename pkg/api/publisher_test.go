package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewPublisherClient 测试创建 Publisher 客户端
func TestNewPublisherClient(t *testing.T) {
	// 测试默认配置
	client := NewPublisherClient()
	assert.NotNil(t, client)
	assert.Equal(t, "https://central.sonatype.com", client.baseURL)
	assert.NotNil(t, client.httpClient)

	// 测试自定义选项
	customClient := NewPublisherClient(
		WithPublisherToken("test-token"),
		WithPublisherBasicAuth("user", "pass"),
		WithPublisherBaseURL("https://custom.example.com"),
	)
	assert.NotNil(t, customClient)
	assert.Equal(t, "test-token", customClient.authToken)
	assert.Equal(t, "user", customClient.authUser)
	assert.Equal(t, "pass", customClient.authPass)
	assert.Equal(t, "https://custom.example.com", customClient.baseURL)
}

// TestPublisherApiAuth 测试 Publisher API 认证（需要有效令牌，在 CI 中跳过）
func TestPublisherApiAuth(t *testing.T) {
	// 这是一个集成测试，需要有效的认证令牌
	// 在 CI 环境中跳过
	t.Skip("需要有效的 Sonatype Central Publisher API 令牌")

	client := NewPublisherClient(
		WithPublisherToken("your-token-here"),
	)

	// 尝试列出部署来验证认证
	deployments, err := client.ListDeployments(t.Context(), nil)
	if err != nil {
		t.Logf("认证测试结果: %v", err)
		// 如果 token 无效，收到 401 是预期的
	}
	_ = deployments
}

// TestUploadBundle 测试上传发布包（需要有效令牌，在 CI 中跳过）
func TestUploadBundle(t *testing.T) {
	t.Skip("需要有效的 Sonatype Central Publisher API 令牌和发布包")

	client := NewPublisherClient(
		WithPublisherToken("your-token-here"),
	)

	bundle := []byte("test-bundle-content")
	deploymentID, err := client.UploadBundle(t.Context(), bundle, "test-component", PublishingTypeAutomatic)
	if err != nil {
		t.Fatalf("上传失败: %v", err)
	}
	t.Logf("部署 ID: %s", deploymentID)
}
