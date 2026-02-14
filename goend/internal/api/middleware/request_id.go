// middleware 包提供HTTP中间件功能
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// ctxKey 上下文键类型
type ctxKey string

// 上下文键和请求头常量
const (
	// RequestIDKey 请求ID在上下文中的键
	RequestIDKey ctxKey = "request_id"
	// RequestIDHeader 请求ID在HTTP头中的名称
	RequestIDHeader = "X-Request-ID"
)

// RequestID 请求ID中间件
// 从请求头获取现有的请求ID，如果不存在则生成新的UUID（取前8位）
// 将请求ID设置到上下文和响应头中，便于链路追踪
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从请求头获取请求ID
		requestID := r.Header.Get(RequestIDHeader)

		// 如果请求头中没有请求ID，则生成新的UUID（取前8位）
		if requestID == "" {
			requestID = uuid.New().String()[:8]
		}

		// 将请求ID设置到上下文中
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		// 将请求ID设置到响应头中
		w.Header().Set(RequestIDHeader, requestID)

		// 调用下一个处理器
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID 从上下文获取请求ID
// 返回请求ID字符串，如果上下文中不存在则返回空字符串
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}

	return ""
}
