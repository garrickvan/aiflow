package helpers

import (
	"aiflow/internal/errors"
	"net/http"

	"github.com/go-chi/render"
)

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`           // 是否成功
	Data    interface{} `json:"data,omitempty"`    // 响应数据
	Message string      `json:"message,omitempty"` // 响应消息
	Error   *ErrorInfo  `json:"error,omitempty"`   // 错误信息
}

// ErrorInfo 错误信息结构
type ErrorInfo struct {
	Code    string `json:"code"`    // 错误码
	Message string `json:"message"` // 错误消息
}

// RenderSuccess 渲染成功响应
// 返回标准成功响应，包含数据
func RenderSuccess(w http.ResponseWriter, req *http.Request, data interface{}) {
	render.JSON(w, req, Response{
		Success: true,
		Data:    data,
	})
}

// RenderSuccessWithMessage 渲染带消息的成功响应
// 返回成功响应，包含消息和数据
func RenderSuccessWithMessage(w http.ResponseWriter, req *http.Request, message string, data interface{}) {
	render.JSON(w, req, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// RenderCreated 渲染创建成功响应
// 返回201状态码和创建成功消息
func RenderCreated(w http.ResponseWriter, req *http.Request, message string, data interface{}) {
	render.Status(req, http.StatusCreated)
	render.JSON(w, req, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// RenderError 渲染错误响应
// 支持AppError类型，自动提取错误码和消息
func RenderError(w http.ResponseWriter, req *http.Request, err error) {
	var statusCode int
	var errorInfo *ErrorInfo

	// 判断是否为AppError类型
	if appErr, ok := errors.IsAppError(err); ok {
		statusCode = appErr.HTTP
		errorInfo = &ErrorInfo{
			Code:    string(appErr.Code),
			Message: appErr.Message,
		}
	} else {
		// 默认内部服务器错误
		statusCode = http.StatusInternalServerError
		errorInfo = &ErrorInfo{
			Code:    string(errors.ErrCodeInternalError),
			Message: err.Error(),
		}
	}

	render.Status(req, statusCode)
	render.JSON(w, req, Response{
		Success: false,
		Error:   errorInfo,
	})
}

// PaginatedResponse 分页响应结构
type PaginatedResponse struct {
	Items      interface{} `json:"items"`      // 数据列表
	Pagination Pagination  `json:"pagination"` // 分页信息
}

// Pagination 分页信息结构
type Pagination struct {
	Total     int64 `json:"total"`     // 总记录数
	Page      int   `json:"page"`      // 当前页码
	PageSize  int   `json:"pageSize"`  // 每页条数
	TotalPage int64 `json:"totalPage"` // 总页数
}

// RenderPaginated 渲染分页响应
// 自动计算总页数，返回标准分页数据结构
func RenderPaginated(w http.ResponseWriter, req *http.Request, items interface{}, total int64, page int, pageSize int) {
	// 计算总页数
	var totalPage int64
	if pageSize > 0 {
		totalPage = (total + int64(pageSize) - 1) / int64(pageSize)
	}

	render.JSON(w, req, Response{
		Success: true,
		Data: PaginatedResponse{
			Items: items,
			Pagination: Pagination{
				Total:     total,
				Page:      page,
				PageSize:  pageSize,
				TotalPage: totalPage,
			},
		},
	})
}
