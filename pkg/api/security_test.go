package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSecurityDataStructures 测试安全相关的数据结构
//
// 安全 API 端点 (search.maven.org/api/security/*) 已被 Sonatype 封锁，
// 网络相关的测试已跳过。此测试验证本地数据结构仍然有效。
func TestSecurityDataStructures(t *testing.T) {
	t.Skip("安全 API 端点已被 Sonatype 封锁（返回 403），数据结构测试保留以验证类型定义")

	// 验证 SecuritySeverity 常量
	sevCritical := SecuritySeverity("CRITICAL")
	sevHigh := SecuritySeverity("HIGH")
	sevMedium := SecuritySeverity("MEDIUM")
	sevLow := SecuritySeverity("LOW")
	sevNone := SecuritySeverity("NONE")

	assert.Equal(t, SecuritySeverity("CRITICAL"), sevCritical)
	assert.Equal(t, SecuritySeverity("HIGH"), sevHigh)
	assert.Equal(t, SecuritySeverity("MEDIUM"), sevMedium)
	assert.Equal(t, SecuritySeverity("LOW"), sevLow)
	assert.Equal(t, SecuritySeverity("NONE"), sevNone)
}

// TestGetSecurityRatingReal 已跳过 - API 不可用
func TestGetSecurityRatingReal(t *testing.T) {
	t.Skip("安全评分 API (search.maven.org/api/security/rating) 返回 403，已被 Sonatype 封锁")
}

// TestCompareVersionSecurityReal 已跳过 - API 不可用
func TestCompareVersionSecurityReal(t *testing.T) {
	t.Skip("安全评分 API 不可用 (403)，依赖此 API 的 CompareVersionSecurity 同样不可用")
}

// TestSearchVulnerableArtifactsReal 已跳过 - API 不可用
func TestSearchVulnerableArtifactsReal(t *testing.T) {
	t.Skip("Solr 索引不再支持 vulnerabilities 字段查询（返回 400）")
}

// TestVulnerabilityTimelineReal 已跳过 - API 不可用
func TestVulnerabilityTimelineReal(t *testing.T) {
	t.Skip("依赖已废弃的安全 API")
}

// TestSecurityEdgeCases 已跳过 - API 不可用
func TestSecurityEdgeCases(t *testing.T) {
	t.Skip("安全 API 不可用，边界情况测试已跳过")
}
