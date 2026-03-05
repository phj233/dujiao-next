package router

import (
	"io"
	"net/http"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/upstream"

	"github.com/gin-gonic/gin"
)

const upstreamUserIDKey = "upstream_user_id"
const upstreamCredentialIDKey = "upstream_credential_id"

// UpstreamAPIAuthMiddleware 上游 API 签名鉴权中间件
func UpstreamAPIAuthMiddleware(credRepo repository.ApiCredentialRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(upstream.HeaderApiKey)
		timestampStr := c.GetHeader(upstream.HeaderTimestamp)
		signature := c.GetHeader(upstream.HeaderSignature)

		if apiKey == "" || timestampStr == "" || signature == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error_code": "missing_auth_headers", "error_message": "missing authentication headers"})
			return
		}

		timestamp, err := upstream.ParseTimestamp(timestampStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error_code": "invalid_timestamp", "error_message": "invalid timestamp"})
			return
		}

		if !upstream.IsTimestampValid(timestamp) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error_code": "timestamp_expired", "error_message": "timestamp expired"})
			return
		}

		cred, err := credRepo.GetByApiKey(apiKey)
		if err != nil {
			logger.Errorw("upstream_auth_db_error", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "internal_error", "error_message": "internal error"})
			return
		}
		if cred == nil || cred.Status != constants.ApiCredentialStatusApproved || !cred.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"ok": false, "error_code": "invalid_api_key", "error_message": "api key is invalid or disabled"})
			return
		}

		// 读取 body 用于签名验证
		var body []byte
		if c.Request.Body != nil {
			body, err = io.ReadAll(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "bad_request", "error_message": "failed to read request body"})
				return
			}
			// 重置 body 供后续 handler 读取
			c.Request.Body = io.NopCloser(
				&bodyReader{data: body},
			)
		}

		method := c.Request.Method
		path := c.Request.URL.Path

		if !upstream.Verify(cred.ApiSecret, method, path, signature, timestamp, body) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error_code": "invalid_signature", "error_message": "signature verification failed"})
			return
		}

		// 更新最后使用时间（异步，不阻塞请求）
		now := time.Now()
		cred.LastUsedAt = &now
		go func() {
			if updateErr := credRepo.Update(cred); updateErr != nil {
				logger.Warnw("upstream_auth_update_last_used_failed", "error", updateErr)
			}
		}()

		// 将凭证信息存入 context
		c.Set(upstreamUserIDKey, cred.UserID)
		c.Set(upstreamCredentialIDKey, cred.ID)
		c.Set("upstream_api_key", cred.ApiKey)

		c.Next()
	}
}

// GetUpstreamUserID 从 context 获取上游 API 认证的用户 ID
func GetUpstreamUserID(c *gin.Context) uint {
	v, _ := c.Get(upstreamUserIDKey)
	if id, ok := v.(uint); ok {
		return id
	}
	return 0
}

// GetUpstreamCredentialID 从 context 获取上游 API 凭证 ID
func GetUpstreamCredentialID(c *gin.Context) uint {
	v, _ := c.Get(upstreamCredentialIDKey)
	if id, ok := v.(uint); ok {
		return id
	}
	return 0
}

// bodyReader 实现 io.Reader，用于重置 body
type bodyReader struct {
	data   []byte
	offset int
}

func (r *bodyReader) Read(p []byte) (n int, err error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

// UpstreamCallbackResponse 上游回调响应
func UpstreamCallbackResponse(c *gin.Context, ok bool, msg string) {
	c.JSON(http.StatusOK, gin.H{"ok": ok, "message": msg})
}

// UpstreamSuccessResponse 上游 API 成功响应
func UpstreamSuccessResponse(c *gin.Context, data interface{}) {
	if data == nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	c.JSON(http.StatusOK, data)
}

// UpstreamErrorResponse 上游 API 错误响应
func UpstreamErrorResponse(c *gin.Context, httpStatus int, errorCode, errorMessage string) {
	c.JSON(httpStatus, gin.H{"ok": false, "error_code": errorCode, "error_message": errorMessage})
}
