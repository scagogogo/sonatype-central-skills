package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// TestCheckPublishedParams 测试 CheckPublished 发送正确的查询参数（namespace/name/version）
func TestCheckPublishedParams(t *testing.T) {
	var capturedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		assert.Equal(t, "/api/v1/publisher/published", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"published": true}`))
	}))
	defer server.Close()

	client := NewPublisherClient(
		WithPublisherToken("test-token"),
		WithPublisherBaseURL(server.URL),
	)

	result, err := client.CheckPublished(context.Background(), "com.example", "my-lib", "1.0.0")
	require.NoError(t, err)
	assert.True(t, result.Published)
	// 验证使用 namespace/name/version 参数，而非 groupId/artifactId
	assert.Contains(t, capturedQuery, "namespace=com.example")
	assert.Contains(t, capturedQuery, "name=my-lib")
	assert.Contains(t, capturedQuery, "version=1.0.0")
	assert.NotContains(t, capturedQuery, "groupId=")
	assert.NotContains(t, capturedQuery, "artifactId=")

	// 验证回填坐标信息
	assert.Equal(t, "com.example", result.Namespace)
	assert.Equal(t, "my-lib", result.Name)
	assert.Equal(t, "1.0.0", result.Version)
}

// TestGetDeploymentStatusParams 测试 GetDeploymentStatus 使用 id 查询参数
func TestGetDeploymentStatusParams(t *testing.T) {
	var capturedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		assert.Equal(t, "/api/v1/publisher/status", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"deploymentId": "dep-123",
			"deploymentName": "my-lib",
			"deploymentState": "VALIDATED"
		}`))
	}))
	defer server.Close()

	client := NewPublisherClient(
		WithPublisherToken("test-token"),
		WithPublisherBaseURL(server.URL),
	)

	result, err := client.GetDeploymentStatus(context.Background(), "dep-123")
	require.NoError(t, err)
	assert.Equal(t, "dep-123", result.DeploymentID)
	assert.Equal(t, "my-lib", result.DeploymentName)
	assert.Equal(t, response.DeploymentStateValidated, result.DeploymentState)
	// 验证 id 作为查询参数
	assert.Contains(t, capturedQuery, "id=dep-123")
}

// TestListDeploymentsResponse 测试 ListDeployments 解析正确的响应结构
func TestListDeploymentsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/publisher/deployments", r.URL.Path)
		// 验证过滤参数
		q := r.URL.Query()
		assert.Equal(t, "com.example", q.Get("namespace"))
		assert.Equal(t, "VALIDATED", q.Get("deploymentState"))
		assert.Equal(t, "0", q.Get("page"))
		assert.Equal(t, "20", q.Get("size"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"deployments": [
				{
					"deploymentId": "dep-1",
					"deploymentName": "my-lib-1.0.0",
					"namespace": "com.example",
					"deploymentState": "PUBLISHED",
					"createTimestamp": "2025-04-18T18:13:20.000Z",
					"updateTimestamp": "2025-04-18T18:33:54.567Z",
					"deploymentComponents": [
						{
							"purl": "pkg:maven/com.example/my-lib@1.0.0",
							"name": "my-lib-1.0.0.jar",
							"path": "com/example/my-lib/1.0.0/"
						}
					]
				}
			],
			"page": 0,
			"pageSize": 20,
			"pageCount": 1,
			"totalResultCount": 1
		}`))
	}))
	defer server.Close()

	client := NewPublisherClient(
		WithPublisherToken("test-token"),
		WithPublisherBaseURL(server.URL),
	)

	result, err := client.ListDeployments(context.Background(), &response.DeploymentListOptions{
		Namespace: "com.example",
		State:     response.DeploymentStateValidated,
		Paginate:  true,
		Page:      0,
		Size:      20,
	})
	require.NoError(t, err)

	require.Len(t, result.Deployments, 1)
	dep := result.Deployments[0]
	assert.Equal(t, "dep-1", dep.DeploymentID)
	assert.Equal(t, "my-lib-1.0.0", dep.DeploymentName)
	assert.Equal(t, "com.example", dep.Namespace)
	assert.Equal(t, response.DeploymentStatePublished, dep.DeploymentState)
	require.Len(t, dep.DeploymentComponents, 1)
	assert.Equal(t, "pkg:maven/com.example/my-lib@1.0.0", dep.DeploymentComponents[0].Purl)

	// 验证分页字段
	assert.Equal(t, 0, result.Page)
	assert.Equal(t, 20, result.PageSize)
	assert.Equal(t, 1, result.PageCount)
	assert.Equal(t, 1, result.TotalResultCount)
}

// TestBrowseDeploymentWithOptions 测试 BrowseDeploymentWithOptions 发送正确的 JSON 请求体
func TestBrowseDeploymentWithOptions(t *testing.T) {
	var capturedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/publisher/deployments/files", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		capturedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"deployments": [
				{
					"deploymentId": "dep-1",
					"deploymentName": "test",
					"deploymentState": "FAILED",
					"deploymentType": "BUNDLE",
					"createTimestamp": 1679067205978,
					"purls": [],
					"deployedComponentVersions": [
						{
							"name": "manifest.json",
							"path": "org/sonatype/nexus/nexus-api/2.3.0",
							"errors": ["Some error"]
						}
					]
				}
			]
		}`))
	}))
	defer server.Close()

	client := NewPublisherClient(
		WithPublisherToken("test-token"),
		WithPublisherBaseURL(server.URL),
	)

	results, err := client.BrowseDeploymentWithOptions(context.Background(), &response.BrowseDeploymentRequest{
		DeploymentIds: []string{"dep-1"},
		Page:          0,
		Size:          100,
		SortField:     "createdTimestamp",
		SortDirection: "desc",
	})
	require.NoError(t, err)

	require.Len(t, results, 1)
	assert.Equal(t, "dep-1", results[0].DeploymentID)
	assert.Equal(t, "test", results[0].DeploymentName)
	assert.Equal(t, response.DeploymentStateFailed, results[0].DeploymentState)
	assert.Equal(t, "BUNDLE", results[0].DeploymentType)
	require.Len(t, results[0].DeployedComponentVersions, 1)
	assert.Equal(t, "manifest.json", results[0].DeployedComponentVersions[0].Name)
	require.Len(t, results[0].DeployedComponentVersions[0].Errors, 1)

	// 验证请求体
	var reqBody response.BrowseDeploymentRequest
	require.NoError(t, json.Unmarshal(capturedBody, &reqBody))
	assert.Equal(t, "createdTimestamp", reqBody.SortField)
	assert.Equal(t, "desc", reqBody.SortDirection)
	assert.Equal(t, []string{"dep-1"}, reqBody.DeploymentIds)
	assert.Equal(t, 0, reqBody.Page)
	assert.Equal(t, 100, reqBody.Size)
}

// TestBrowseDeploymentRequiresSortField 测试 sortField 必填校验
func TestBrowseDeploymentRequiresSortField(t *testing.T) {
	client := NewPublisherClient(WithPublisherToken("test-token"))
	_, err := client.BrowseDeploymentWithOptions(context.Background(), &response.BrowseDeploymentRequest{
		DeploymentIds: []string{"dep-1"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sortField")
}

// TestBrowseDeploymentNilRequest 测试 nil 请求校验
func TestBrowseDeploymentNilRequest(t *testing.T) {
	client := NewPublisherClient(WithPublisherToken("test-token"))
	_, err := client.BrowseDeploymentWithOptions(context.Background(), nil)
	assert.Error(t, err)
}

// TestBrowseDeploymentConvenience 测试 BrowseDeployment 便捷方法
func TestBrowseDeploymentConvenience(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"deployments": [
				{
					"deploymentId": "dep-1",
					"deploymentName": "test",
					"deploymentState": "VALIDATED",
					"deploymentType": "BUNDLE"
				}
			]
		}`))
	}))
	defer server.Close()

	client := NewPublisherClient(
		WithPublisherToken("test-token"),
		WithPublisherBaseURL(server.URL),
	)

	result, err := client.BrowseDeployment(context.Background(), "dep-1")
	require.NoError(t, err)
	assert.Equal(t, "dep-1", result.DeploymentID)
	assert.Equal(t, response.DeploymentStateValidated, result.DeploymentState)
}

// TestBrowseDeploymentConvenienceEmpty 测试便捷方法在无结果时返回错误
func TestBrowseDeploymentConvenienceEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"deployments": []}`))
	}))
	defer server.Close()

	client := NewPublisherClient(
		WithPublisherToken("test-token"),
		WithPublisherBaseURL(server.URL),
	)

	_, err := client.BrowseDeployment(context.Background(), "dep-1")
	assert.Error(t, err)
}

// TestPublisherErrorResponse 测试 Publisher API 错误响应解析
func TestPublisherErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"httpStatus": 401,
			"errorCode": 10401,
			"message": "Wrong authorization data",
			"explanation": null,
			"data": null
		}`))
	}))
	defer server.Close()

	client := NewPublisherClient(
		WithPublisherToken("invalid-token"),
		WithPublisherBaseURL(server.URL),
	)

	_, err := client.CheckPublished(context.Background(), "com.example", "my-lib", "1.0.0")
	require.Error(t, err)

	var errResp *response.PublisherErrorResponse
	require.ErrorAs(t, err, &errResp)
	assert.Equal(t, 401, errResp.HttpStatus)
	assert.Equal(t, 10401, errResp.ErrorCode)
	assert.Equal(t, "Wrong authorization data", errResp.Message)
}

// TestPublisherErrorResponseError 测试 PublisherErrorResponse 的 Error() 方法
func TestPublisherErrorResponseError(t *testing.T) {
	errResp := &response.PublisherErrorResponse{
		HttpStatus: 400,
		ErrorCode:  10400,
		Message:    "Bad request",
	}
	errStr := errResp.Error()
	assert.Contains(t, errStr, "400")
	assert.Contains(t, errStr, "Bad request")
}

// TestListDeploymentsNilOptions 测试 nil 选项列出所有部署
func TestListDeploymentsNilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/publisher/deployments", r.URL.Path)
		assert.Empty(t, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"deployments": [], "totalResultCount": 0}`))
	}))
	defer server.Close()

	client := NewPublisherClient(
		WithPublisherToken("test-token"),
		WithPublisherBaseURL(server.URL),
	)

	result, err := client.ListDeployments(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, result.Deployments)
}

// TestUploadBundleRequest 测试上传部署包的 multipart 和查询参数
func TestUploadBundleRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/publisher/upload", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		// 验证查询参数
		q := r.URL.Query()
		assert.Equal(t, "my-component", q.Get("name"))
		assert.Equal(t, "AUTOMATIC", q.Get("publishingType"))

		// 验证 multipart 内容
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
		err := r.ParseMultipartForm(10 << 20)
		require.NoError(t, err)
		file, header, err := r.FormFile("bundle")
		require.NoError(t, err)
		defer file.Close()
		assert.Equal(t, "my-component.zip", header.Filename)
		data, _ := io.ReadAll(file)
		assert.Equal(t, []byte("test-bundle-content"), data)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"deploymentId": "dep-123"}`))
	}))
	defer server.Close()

	client := NewPublisherClient(
		WithPublisherToken("test-token"),
		WithPublisherBaseURL(server.URL),
	)

	deploymentID, err := client.UploadBundle(context.Background(), []byte("test-bundle-content"), "my-component", PublishingTypeAutomatic)
	require.NoError(t, err)
	assert.Equal(t, "dep-123", deploymentID)
}

// TestUploadBundleDefaultNameRequest 测试上传时 name 为空使用默认文件名
func TestUploadBundleDefaultNameRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		require.NoError(t, err)
		_, header, err := r.FormFile("bundle")
		require.NoError(t, err)
		// name 为空时应使用默认文件名 bundle.zip
		assert.Equal(t, "bundle.zip", header.Filename)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"deploymentId": "dep-456"}`))
	}))
	defer server.Close()

	client := NewPublisherClient(
		WithPublisherToken("test-token"),
		WithPublisherBaseURL(server.URL),
	)

	deploymentID, err := client.UploadBundle(context.Background(), []byte("data"), "", PublishingTypeAutomatic)
	require.NoError(t, err)
	assert.Equal(t, "dep-456", deploymentID)
}

// TestHasExtension 测试 hasExtension 辅助函数
func TestHasExtension(t *testing.T) {
	assert.True(t, hasExtension("file.zip"))
	assert.True(t, hasExtension("bundle.jar"))
	assert.False(t, hasExtension("bundle"))
	assert.False(t, hasExtension("path/to/bundle"))
}

// TestDropDeployment 测试删除部署
func TestDropDeployment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/publisher/deployment/dep-123", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewPublisherClient(
		WithPublisherToken("test-token"),
		WithPublisherBaseURL(server.URL),
	)

	err := client.DropDeployment(context.Background(), "dep-123")
	require.NoError(t, err)
}

// TestPublishDeployment 测试发布部署
func TestPublishDeployment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/publisher/deployment/dep-123", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewPublisherClient(
		WithPublisherToken("test-token"),
		WithPublisherBaseURL(server.URL),
	)

	err := client.PublishDeployment(context.Background(), "dep-123")
	require.NoError(t, err)
}
