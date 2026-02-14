// Package errors 提供统一的错误处理机制
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode string

// 通用错误码
const (
	// 参数错误
	ErrCodeInvalidIDParam   ErrorCode = "COM-INVLD-001" // 无效的ID参数
	ErrCodeBadRequestParam  ErrorCode = "COM-INVLD-002" // 无效的请求参数
	ErrCodeBadRequest       ErrorCode = "COM-REQ-001"   // 请求格式错误
	ErrCodeNotFound         ErrorCode = "COM-NF-001"    // 资源不存在
	ErrCodeInternalError    ErrorCode = "COM-SYS-001"   // 系统内部错误
)

// 技能模块错误码
const (
	ErrCodeSkillNotFound   ErrorCode = "SKL-NF-001"   // 技能不存在
	ErrCodeSkillCreate     ErrorCode = "SKL-CRT-001"  // 技能创建失败
	ErrCodeSkillUpdate     ErrorCode = "SKL-UPD-001"  // 技能更新失败
	ErrCodeSkillDelete     ErrorCode = "SKL-DEL-001"  // 技能删除失败
	ErrCodeSkillTrash      ErrorCode = "SKL-TRSH-001" // 技能回收失败
)

// 任务模块错误码
const (
	ErrCodeTaskNotFound   ErrorCode = "TSK-NF-001"   // 任务不存在
	ErrCodeTaskCreate     ErrorCode = "TSK-CRT-001"  // 任务创建失败
	ErrCodeTaskUpdate     ErrorCode = "TSK-UPD-001"  // 任务更新失败
	ErrCodeTaskDelete     ErrorCode = "TSK-DEL-001"  // 任务删除失败
	ErrCodeTaskTrash      ErrorCode = "TSK-TRSH-001" // 任务回收失败
	ErrCodeTaskValidate   ErrorCode = "TSK-VAL-001"  // 任务验证失败
)

// 标签模块错误码
const (
	ErrCodeTagNotFound ErrorCode = "TAG-NF-001"  // 标签不存在
	ErrCodeTagCreate   ErrorCode = "TAG-CRT-001" // 标签创建失败
	ErrCodeTagUpdate   ErrorCode = "TAG-UPD-001" // 标签更新失败
	ErrCodeTagDelete   ErrorCode = "TAG-DEL-001" // 标签删除失败
)

// 错误消息映射
var errorCodeMessages = map[ErrorCode]string{
	ErrCodeInvalidIDParam:  "无效的ID参数",
	ErrCodeBadRequestParam: "无效的请求参数",
	ErrCodeBadRequest:      "请求格式错误",
	ErrCodeNotFound:        "资源不存在",
	ErrCodeInternalError:   "系统内部错误",

	ErrCodeSkillNotFound: "技能不存在",
	ErrCodeSkillCreate:   "技能创建失败",
	ErrCodeSkillUpdate:   "技能更新失败",
	ErrCodeSkillDelete:   "技能删除失败",
	ErrCodeSkillTrash:    "技能回收失败",

	ErrCodeTaskNotFound:  "任务不存在",
	ErrCodeTaskCreate:    "任务创建失败",
	ErrCodeTaskUpdate:    "任务更新失败",
	ErrCodeTaskDelete:    "任务删除失败",
	ErrCodeTaskTrash:     "任务回收失败",
	ErrCodeTaskValidate:  "任务验证失败",

	ErrCodeTagNotFound: "标签不存在",
	ErrCodeTagCreate:   "标签创建失败",
	ErrCodeTagUpdate:   "标签更新失败",
	ErrCodeTagDelete:   "标签删除失败",
}

// 错误码对应的HTTP状态码映射
var errorCodeHTTPStatus = map[ErrorCode]int{
	ErrCodeInvalidIDParam:  http.StatusBadRequest,
	ErrCodeBadRequestParam: http.StatusBadRequest,
	ErrCodeBadRequest:      http.StatusBadRequest,
	ErrCodeNotFound:        http.StatusNotFound,
	ErrCodeInternalError:   http.StatusInternalServerError,

	ErrCodeSkillNotFound: http.StatusNotFound,
	ErrCodeSkillCreate:   http.StatusInternalServerError,
	ErrCodeSkillUpdate:   http.StatusInternalServerError,
	ErrCodeSkillDelete:   http.StatusInternalServerError,
	ErrCodeSkillTrash:    http.StatusInternalServerError,

	ErrCodeTaskNotFound:  http.StatusNotFound,
	ErrCodeTaskCreate:    http.StatusInternalServerError,
	ErrCodeTaskUpdate:    http.StatusInternalServerError,
	ErrCodeTaskDelete:    http.StatusInternalServerError,
	ErrCodeTaskTrash:     http.StatusInternalServerError,
	ErrCodeTaskValidate:  http.StatusBadRequest,

	ErrCodeTagNotFound: http.StatusNotFound,
	ErrCodeTagCreate:   http.StatusInternalServerError,
	ErrCodeTagUpdate:   http.StatusInternalServerError,
	ErrCodeTagDelete:   http.StatusInternalServerError,
}

// AppError 应用错误结构体
type AppError struct {
	Code    ErrorCode // 错误码
	Message string    // 错误消息
	HTTP    int       // HTTP状态码
	Err     error     // 原始错误
}

// Error 实现error接口，返回错误消息
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误，支持errors.Is和errors.As
func (e *AppError) Unwrap() error {
	return e.Err
}

// 预定义错误实例
var (
	// ErrInvalidIDParam 无效ID参数错误
	ErrInvalidIDParam = &AppError{
		Code:    ErrCodeInvalidIDParam,
		Message: errorCodeMessages[ErrCodeInvalidIDParam],
		HTTP:    errorCodeHTTPStatus[ErrCodeInvalidIDParam],
	}
	// ErrBadRequestParam 无效请求参数错误
	ErrBadRequestParam = &AppError{
		Code:    ErrCodeBadRequestParam,
		Message: errorCodeMessages[ErrCodeBadRequestParam],
		HTTP:    errorCodeHTTPStatus[ErrCodeBadRequestParam],
	}
)

// getMessage 获取错误码对应的默认消息
func getMessage(code ErrorCode) string {
	if msg, ok := errorCodeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// getHTTPStatus 获取错误码对应的HTTP状态码
func getHTTPStatus(code ErrorCode) int {
	if status, ok := errorCodeHTTPStatus[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// NewInvalidParamError 创建无效参数错误
func NewInvalidParamError(code ErrorCode, message string, err error) *AppError {
	if message == "" {
		message = getMessage(code)
	}
	return &AppError{
		Code:    code,
		Message: message,
		HTTP:    getHTTPStatus(code),
		Err:     err,
	}
}

// NewNotFoundError 创建资源不存在错误
func NewNotFoundError(code ErrorCode, message string, err error) *AppError {
	if message == "" {
		message = getMessage(code)
	}
	return &AppError{
		Code:    code,
		Message: message,
		HTTP:    http.StatusNotFound,
		Err:     err,
	}
}

// NewInternalError 创建系统内部错误
func NewInternalError(code ErrorCode, message string, err error) *AppError {
	if message == "" {
		message = getMessage(code)
	}
	return &AppError{
		Code:    code,
		Message: message,
		HTTP:    http.StatusInternalServerError,
		Err:     err,
	}
}

// NewSkillError 创建技能模块错误
func NewSkillError(code ErrorCode, message string, err error) *AppError {
	if message == "" {
		message = getMessage(code)
	}
	return &AppError{
		Code:    code,
		Message: message,
		HTTP:    getHTTPStatus(code),
		Err:     err,
	}
}

// NewTaskError 创建任务模块错误
func NewTaskError(code ErrorCode, message string, err error) *AppError {
	if message == "" {
		message = getMessage(code)
	}
	return &AppError{
		Code:    code,
		Message: message,
		HTTP:    getHTTPStatus(code),
		Err:     err,
	}
}

// NewTagError 创建标签模块错误
func NewTagError(code ErrorCode, message string, err error) *AppError {
	if message == "" {
		message = getMessage(code)
	}
	return &AppError{
		Code:    code,
		Message: message,
		HTTP:    getHTTPStatus(code),
		Err:     err,
	}
}

// IsAppError 检查错误是否为AppError类型
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
