package helpers

import (
	"aiflow/internal/errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// 分页相关常量
const (
	// DefaultPage 默认页码
	DefaultPage = 1
	// DefaultPageSize 默认每页条数
	DefaultPageSize = 10
	// MaxPageSize 最大每页条数
	MaxPageSize = 100
)

// PaginationParams 分页参数结构体
type PaginationParams struct {
	Page     int // 当前页码
	PageSize int // 每页条数
}

// ParsePagination 解析分页参数
// 从HTTP请求中提取page和pageSize参数，未提供时使用默认值
func ParsePagination(req *http.Request) PaginationParams {
	page := DefaultPage
	pageSize := DefaultPageSize

	pageStr := req.URL.Query().Get("page")
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSizeStr := req.URL.Query().Get("pageSize")
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= MaxPageSize {
			pageSize = ps
		}
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}

// ParseIDParam 解析URL路径中的ID参数
// 返回解析后的uint类型ID，解析失败时返回错误
func ParseIDParam(req *http.Request, paramName string) (uint, error) {
	idStr := chi.URLParam(req, paramName)
	if idStr == "" {
		return 0, errors.NewInvalidParamError(errors.ErrCodeInvalidIDParam, "缺少ID参数", nil)
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, errors.NewInvalidParamError(errors.ErrCodeInvalidIDParam, "无效的ID参数", err)
	}

	return uint(id), nil
}

// ParseIntParam 解析整数查询参数
// 参数不存在或解析失败时返回默认值
func ParseIntParam(req *http.Request, paramName string, defaultValue int64) int64 {
	str := req.URL.Query().Get(paramName)
	if str == "" {
		return defaultValue
	}

	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return defaultValue
	}

	return val
}

// ParseUintParam 解析无符号整数查询参数
// 参数不存在或解析失败时返回错误
func ParseUintParam(req *http.Request, paramName string) (uint, error) {
	str := req.URL.Query().Get(paramName)
	if str == "" {
		return 0, errors.NewInvalidParamError(errors.ErrCodeBadRequestParam, "缺少参数: "+paramName, nil)
	}

	val, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0, errors.NewInvalidParamError(errors.ErrCodeBadRequestParam, "无效的参数: "+paramName, err)
	}

	return uint(val), nil
}
