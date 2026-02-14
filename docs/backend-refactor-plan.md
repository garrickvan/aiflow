# 后端代码重构落地方案

## 一、重构目标

### 1.1 核心指标对比表

| 指标 | 当前状态 | 目标状态 | 改进幅度 |
|------|----------|----------|----------|
| 错误处理 | 硬编码字符串，无统一错误码 | 统一错误类型+错误码 | 100%规范化 |
| 架构层次 | Handler层混用Service/Repository | 全部Handler→Service→Repository | 100%分层统一 |
| 代码复用 | 分页/ID解析重复约15处 | 公共函数统一处理 | 减少~200行重复代码 |
| 分词索引性能 | 逐条插入，N次DB操作 | 批量插入，1次DB操作 | 性能提升N倍 |
| 日志追踪 | 无请求ID，难以追踪 | 请求ID全链路追踪 | 问题定位效率+80% |
| 配置验证 | 仅基础默认值处理 | 完整验证+类型安全 | 配置错误率-90% |

### 1.2 重构原则

- **向后兼容**：不破坏现有API接口
- **渐进式**：分阶段实施，每阶段可独立验收
- **最小改动**：优先抽象公共逻辑，减少重复代码
- **可测试**：每个改动点都有对应的测试验证

---

## 二、重构阶段规划

### 优先级说明

| 优先级 | 说明 | 预计工时 |
|--------|------|----------|
| P0 | 核心架构问题，影响可维护性 | 4h |
| P1 | 性能优化和代码质量提升 | 3h |
| P2 | 增强功能，提升开发体验 | 2h |

### 阶段总览

```
阶段P0（核心架构）
├── 任务1: 统一错误处理
├── 任务2: 统一架构层次
└── 任务3: 代码复用抽象

阶段P1（性能优化）
├── 任务4: 分词索引批量插入
└── 任务5: N+1查询修复

阶段P2（增强功能）
├── 任务6: 请求ID追踪中间件
└── 任务7: 配置管理优化
```

---

## 三、P0阶段：核心架构重构

### 任务1：统一错误处理

#### 1.1 问题分析

当前代码中错误处理分散且不规范：

```go
// 当前问题：硬编码错误消息，无错误码
render.JSON(w, req, Response{
    Success: false,
    Error:   "无效的ID参数",  // 硬编码，难以国际化
})
```

#### 1.2 解决方案

创建 `internal/errors/errors.go` 统一错误管理：

```go
// Package errors 提供统一的错误处理机制
package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode string

// 错误码定义（格式：模块-错误类型-序号）
const (
	// 通用错误 1xxx
	ErrInvalidParam     ErrorCode = "COM-INVLD-001" // 无效参数
	ErrInvalidID        ErrorCode = "COM-INVLD-002" // 无效ID
	ErrBadRequest       ErrorCode = "COM-REQ-001"   // 请求参数错误
	ErrNotFound         ErrorCode = "COM-NF-001"    // 资源不存在
	ErrInternal         ErrorCode = "COM-SYS-001"   // 内部错误

	// 技能模块错误 2xxx
	ErrSkillNotFound    ErrorCode = "SKL-NF-001"    // 技能不存在
	ErrSkillCreate      ErrorCode = "SKL-CRT-001"   // 创建技能失败
	ErrSkillUpdate      ErrorCode = "SKL-UPD-001"   // 更新技能失败
	ErrSkillDelete      ErrorCode = "SKL-DEL-001"   // 删除技能失败
	ErrSkillNotInTrash  ErrorCode = "SKL-TRSH-001"  // 技能不在回收站

	// 任务模块错误 3xxx
	ErrTaskNotFound     ErrorCode = "TSK-NF-001"    // 任务不存在
	ErrTaskCreate       ErrorCode = "TSK-CRT-001"   // 创建任务失败
	ErrTaskUpdate       ErrorCode = "TSK-UPD-001"   // 更新任务失败
	ErrTaskDelete       ErrorCode = "TSK-DEL-001"   // 删除任务失败
	ErrTaskNotInTrash   ErrorCode = "TSK-TRSH-001"  // 任务不在回收站
	ErrTaskEmptyField   ErrorCode = "TSK-VAL-001"   // 任务必填字段为空
)

// AppError 应用错误结构
type AppError struct {
	Code    ErrorCode // 错误码
	Message string    // 错误消息
	HTTP    int       // HTTP状态码
	Err     error     // 原始错误（可选）
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 支持错误解包
func (e *AppError) Unwrap() error {
	return e.Err
}

// 预定义错误实例（常用错误，避免重复创建）
var (
	// 通用错误
	ErrInvalidIDParam = &AppError{
		Code:    ErrInvalidID,
		Message: "无效的ID参数",
		HTTP:    http.StatusBadRequest,
	}
	ErrBadRequestParam = &AppError{
		Code:    ErrBadRequest,
		Message: "请求参数错误",
		HTTP:    http.StatusBadRequest,
	}
)

// NewInvalidParamError 创建无效参数错误
func NewInvalidParamError(paramName string) *AppError {
	return &AppError{
		Code:    ErrInvalidParam,
		Message: fmt.Sprintf("无效的%s参数", paramName),
		HTTP:    http.StatusBadRequest,
	}
}

// NewNotFoundError 创建资源不存在错误
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    ErrNotFound,
		Message: fmt.Sprintf("%s不存在", resource),
		HTTP:    http.StatusNotFound,
	}
}

// NewInternalError 创建内部错误
func NewInternalError(operation string, err error) *AppError {
	return &AppError{
		Code:    ErrInternal,
		Message: fmt.Sprintf("%s失败", operation),
		HTTP:    http.StatusInternalServerError,
		Err:     err,
	}
}

// NewSkillError 创建技能相关错误
func NewSkillError(code ErrorCode, operation string, err error) *AppError {
	httpStatus := http.StatusInternalServerError
	message := fmt.Sprintf("%s技能失败", operation)

	switch code {
	case ErrSkillNotFound, ErrSkillNotInTrash:
		httpStatus = http.StatusNotFound
		message = getSkillMessage(code)
	}

	return &AppError{
		Code:    code,
		Message: message,
		HTTP:    httpStatus,
		Err:     err,
	}
}

// NewTaskError 创建任务相关错误
func NewTaskError(code ErrorCode, operation string, err error) *AppError {
	httpStatus := http.StatusInternalServerError
	message := fmt.Sprintf("%s任务失败", operation)

	switch code {
	case ErrTaskNotFound, ErrTaskNotInTrash:
		httpStatus = http.StatusNotFound
		message = getTaskMessage(code)
	case ErrTaskEmptyField:
		httpStatus = http.StatusBadRequest
		message = "任务编号、所属项目、任务类型、任务目标不能为空"
	}

	return &AppError{
		Code:    code,
		Message: message,
		HTTP:    httpStatus,
		Err:     err,
	}
}

// getSkillMessage 获取技能错误消息
func getSkillMessage(code ErrorCode) string {
	messages := map[ErrorCode]string{
		ErrSkillNotFound:   "技能不存在",
		ErrSkillNotInTrash: "技能不存在或不在回收站中",
	}
	return messages[code]
}

// getTaskMessage 获取任务错误消息
func getTaskMessage(code ErrorCode) string {
	messages := map[ErrorCode]string{
		ErrTaskNotFound:   "任务不存在",
		ErrTaskNotInTrash: "任务不存在或不在回收站中",
	}
	return messages[code]
}

// IsAppError 检查是否为应用错误
func IsAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}
```

#### 1.3 Handler层改造示例

**改造前** (`handlers/skill.go`):

```go
func (h *SkillHandler) GetSkill(w http.ResponseWriter, req *http.Request) {
	idStr := chi.URLParam(req, "id")
	id, err := services.ParseUint(idStr)
	if err != nil {
		render.Status(req, http.StatusBadRequest)
		render.JSON(w, req, Response{
			Success: false,
			Error:   "无效的ID参数",
		})
		return
	}
	// ...
}
```

**改造后**:

```go
func (h *SkillHandler) GetSkill(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.ParseIDParam(req, "id")
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}
	// ...
}
```

#### 1.4 实施步骤

1. 创建 `internal/errors/errors.go` 文件
2. 定义所有错误码和错误构造函数
3. 创建 `internal/api/helpers/response.go` 响应辅助函数
4. 逐个Handler替换错误处理逻辑
5. 更新单元测试验证错误码

---

### 任务2：统一架构层次

#### 2.1 问题分析

当前架构层次不一致：

```
SkillHandler → SkillService → Repository  ✓ 正确
JobTaskHandler → Repository               ✗ 跳过Service层
MCP层 → Repository                        ✗ 跳过Service层
```

#### 2.2 解决方案

创建 `internal/services/jobtask_service.go`：

```go
// Package services 提供业务逻辑层
package services

import (
	"aiflow/internal/errors"
	"aiflow/internal/models"
	"aiflow/internal/repositories"
	"context"
	"time"
)

// JobTaskService 任务服务层
type JobTaskService struct {
	repo *repositories.Repository
}

// NewJobTaskService 创建任务服务实例
func NewJobTaskService(repo *repositories.Repository) *JobTaskService {
	return &JobTaskService{repo: repo}
}

// ListJobTasksRequest 获取任务列表请求参数
type ListJobTasksRequest struct {
	Page      int
	PageSize  int
	Project   string
	Type      string
	Status    string
	StartDate int64
	EndDate   int64
}

// ListJobTasksResponse 获取任务列表响应
type ListJobTasksResponse struct {
	Items      []models.JobTask       `json:"items"`
	Pagination map[string]interface{} `json:"pagination"`
}

// ListJobTasks 获取任务列表
func (s *JobTaskService) ListJobTasks(ctx context.Context, req ListJobTasksRequest) (*ListJobTasksResponse, error) {
	jobTasks, total, err := s.repo.ListJobTasks(ctx, req.Page, req.PageSize,
		req.Project, req.Type, req.Status, req.StartDate, req.EndDate)
	if err != nil {
		return nil, errors.NewInternalError("获取任务列表", err)
	}

	pagination := map[string]interface{}{
		"total":     total,
		"page":      req.Page,
		"pageSize":  req.PageSize,
		"totalPage": (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}

	return &ListJobTasksResponse{
		Items:      jobTasks,
		Pagination: pagination,
	}, nil
}

// CreateJobTaskRequest 创建任务请求参数
type CreateJobTaskRequest struct {
	JobNo         string `json:"jobNo"`
	Project       string `json:"project"`
	Type          string `json:"type"`
	Goal          string `json:"goal"`
	PassAcceptStd bool   `json:"passAcceptStd"`
	Status        string `json:"status"`
}

// CreateJobTask 创建任务
func (s *JobTaskService) CreateJobTask(ctx context.Context, req CreateJobTaskRequest) (*models.JobTask, error) {
	// 验证必填字段
	if req.JobNo == "" || req.Project == "" || req.Type == "" || req.Goal == "" {
		return nil, errors.NewTaskError(errors.ErrTaskEmptyField, "", nil)
	}

	timestamp := time.Now().UnixMilli()
	jobTask := &models.JobTask{
		JobNo:         req.JobNo,
		Project:       req.Project,
		Type:          req.Type,
		Goal:          req.Goal,
		PassAcceptStd: req.PassAcceptStd,
		Status:        req.Status,
		CreatedAt:     timestamp,
		UpdatedAt:     timestamp,
	}

	// 设置默认状态
	if jobTask.Status == "" {
		jobTask.Status = "已创建"
	}

	if err := s.repo.CreateJobTask(ctx, jobTask); err != nil {
		return nil, errors.NewTaskError(errors.ErrTaskCreate, "创建", err)
	}

	return jobTask, nil
}

// GetJobTask 根据ID获取任务
func (s *JobTaskService) GetJobTask(ctx context.Context, id uint) (*models.JobTask, error) {
	jobTask, err := s.repo.GetJobTaskByID(ctx, id)
	if err != nil {
		return nil, errors.NewTaskError(errors.ErrTaskNotFound, "", err)
	}
	return jobTask, nil
}

// UpdateJobTaskRequest 更新任务请求参数
type UpdateJobTaskRequest struct {
	ID            uint `json:"-"`
	PassAcceptStd bool `json:"passAcceptStd"`
	Status        string `json:"status"`
}

// UpdateJobTask 更新任务
func (s *JobTaskService) UpdateJobTask(ctx context.Context, req UpdateJobTaskRequest) (*models.JobTask, error) {
	jobTask, err := s.repo.GetJobTaskByID(ctx, req.ID)
	if err != nil {
		return nil, errors.NewTaskError(errors.ErrTaskNotFound, "", err)
	}

	jobTask.PassAcceptStd = req.PassAcceptStd
	jobTask.Status = req.Status
	jobTask.UpdatedAt = time.Now().UnixMilli()

	if err := s.repo.UpdateJobTask(ctx, jobTask); err != nil {
		return nil, errors.NewTaskError(errors.ErrTaskUpdate, "更新", err)
	}

	return jobTask, nil
}

// DeleteJobTask 删除任务（伪删除）
func (s *JobTaskService) DeleteJobTask(ctx context.Context, id uint) error {
	if err := s.repo.DeleteJobTask(ctx, id); err != nil {
		return errors.NewTaskError(errors.ErrTaskDelete, "删除", err)
	}
	return nil
}

// RestoreJobTask 恢复回收站中的任务
func (s *JobTaskService) RestoreJobTask(ctx context.Context, id uint) error {
	if err := s.repo.RestoreJobTask(ctx, id); err != nil {
		return errors.NewTaskError(errors.ErrTaskNotInTrash, "", err)
	}
	return nil
}

// PermanentDeleteJobTask 彻底删除任务
func (s *JobTaskService) PermanentDeleteJobTask(ctx context.Context, id uint) error {
	if err := s.repo.PermanentDeleteJobTask(ctx, id); err != nil {
		return errors.NewTaskError(errors.ErrTaskNotInTrash, "", err)
	}
	return nil
}

// GetAllProjects 获取所有项目列表
func (s *JobTaskService) GetAllProjects(ctx context.Context) ([]string, error) {
	projects, err := s.repo.GetAllJobTaskProjects(ctx)
	if err != nil {
		return nil, errors.NewInternalError("获取项目列表", err)
	}
	return projects, nil
}

// BatchExportRequest 批量导出请求
type BatchExportRequest struct {
	IDs    []uint `json:"ids"`
	Format string `json:"format"`
}

// GetJobTasksForExport 获取用于导出的任务数据
func (s *JobTaskService) GetJobTasksForExport(ctx context.Context, ids []uint) ([]models.JobTask, error) {
	var jobTasks []models.JobTask
	var err error

	if len(ids) > 0 {
		jobTasks, err = s.repo.GetJobTasksByIDs(ctx, ids)
	} else {
		jobTasks, err = s.repo.GetAllJobTasks(ctx)
	}

	if err != nil {
		return nil, errors.NewInternalError("获取任务数据", err)
	}
	return jobTasks, nil
}
```

#### 2.3 Handler层改造

**改造前** (`handlers/jobtask.go`):

```go
type JobTaskHandler struct {
	repo *repositories.Repository  // 直接依赖Repository
}

func (h *JobTaskHandler) ListJobTasks(...) {
	// 直接调用repo
	jobTasks, total, err := h.repo.ListJobTasks(...)
}
```

**改造后**:

```go
type JobTaskHandler struct {
	service *services.JobTaskService  // 依赖Service层
}

func NewJobTaskHandler(service *services.JobTaskService) *JobTaskHandler {
	return &JobTaskHandler{service: service}
}

func (h *JobTaskHandler) ListJobTasks(w http.ResponseWriter, req *http.Request) {
	// 解析参数...
	result, err := h.service.ListJobTasks(ctx, services.ListJobTasksRequest{
		Page:      page,
		PageSize:  pageSize,
		Project:   project,
		Type:      jobType,
		Status:    status,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}
	helpers.RenderSuccess(w, req, result)
}
```

#### 2.4 实施步骤

1. 创建 `internal/services/jobtask_service.go`
2. 将 `JobTaskHandler` 改为依赖 `JobTaskService`
3. 更新 `main.go` 中的依赖注入
4. 迁移业务逻辑（验证、状态默认值等）到Service层
5. 保持Repository层只做数据访问

---

### 任务3：代码复用抽象

#### 3.1 问题分析

重复代码分布：

| 位置 | 重复代码 | 出现次数 |
|------|----------|----------|
| Handler层 | 分页参数解析 | 5处 |
| Handler层 | ID参数解析 | 10+处 |
| Handler层 | 响应构造 | 20+处 |
| Handler层 | 时间戳格式化 | 3处 |

#### 3.2 解决方案

创建 `internal/api/helpers/helpers.go`：

```go
// Package helpers 提供Handler层的公共辅助函数
package helpers

import (
	"aiflow/internal/errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// 分页常量
const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int
	PageSize int
}

// ParsePagination 解析分页参数
// 返回默认值如果参数无效
func ParsePagination(req *http.Request) PaginationParams {
	pageStr := req.URL.Query().Get("page")
	pageSizeStr := req.URL.Query().Get("pageSize")

	page := DefaultPage
	pageSize := DefaultPageSize

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= MaxPageSize {
			pageSize = ps
		}
	}

	return PaginationParams{Page: page, PageSize: pageSize}
}

// ParseIDParam 解析URL路径中的ID参数
func ParseIDParam(req *http.Request, paramName string) (uint, error) {
	idStr := chi.URLParam(req, paramName)
	if idStr == "" {
		return 0, errors.NewInvalidParamError(paramName)
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, errors.ErrInvalidIDParam
	}
	return uint(id), nil
}

// ParseIntParam 解析URL查询参数中的整数值
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

// ParseUintParam 解析URL查询参数中的无符号整数值
func ParseUintParam(req *http.Request, paramName string) (uint, error) {
	str := req.URL.Query().Get(paramName)
	if str == "" {
		return 0, nil
	}

	val, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0, errors.NewInvalidParamError(paramName)
	}
	return uint(val), nil
}
```

创建 `internal/api/helpers/response.go`：

```go
package helpers

import (
	"aiflow/internal/errors"
	"net/http"

	"github.com/go-chi/render"
)

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息结构
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RenderSuccess 渲染成功响应
func RenderSuccess(w http.ResponseWriter, req *http.Request, data interface{}) {
	render.JSON(w, req, Response{
		Success: true,
		Data:    data,
	})
}

// RenderSuccessWithMessage 渲染带消息的成功响应
func RenderSuccessWithMessage(w http.ResponseWriter, req *http.Request, message string, data interface{}) {
	render.JSON(w, req, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// RenderCreated 渲染创建成功响应
func RenderCreated(w http.ResponseWriter, req *http.Request, message string, data interface{}) {
	render.Status(req, http.StatusCreated)
	render.JSON(w, req, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// RenderError 渲染错误响应
func RenderError(w http.ResponseWriter, req *http.Request, err error) {
	appErr, ok := errors.IsAppError(err)
	if !ok {
		// 非应用错误，返回内部错误
		render.Status(req, http.StatusInternalServerError)
		render.JSON(w, req, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(errors.ErrInternal),
				Message: "服务器内部错误",
			},
		})
		return
	}

	render.Status(req, appErr.HTTP)
	render.JSON(w, req, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    string(appErr.Code),
			Message: appErr.Message,
		},
	})
}

// RenderPaginated 渲染分页响应
func RenderPaginated(w http.ResponseWriter, req *http.Request, items interface{}, total int64, page, pageSize int) {
	render.JSON(w, req, Response{
		Success: true,
		Data: map[string]interface{}{
			"items": items,
			"pagination": map[string]interface{}{
				"total":     total,
				"page":      page,
				"pageSize":  pageSize,
				"totalPage": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		},
	})
}
```

创建 `internal/utils/time.go`：

```go
// Package utils 提供通用工具函数
package utils

import "time"

// FormatTimestamp 格式化毫秒时间戳为可读字符串
func FormatTimestamp(timestamp int64) string {
	t := time.UnixMilli(timestamp)
	return t.Format("2006-01-02 15:04:05")
}

// FormatTimestampWithLayout 使用指定格式格式化时间戳
func FormatTimestampWithLayout(timestamp int64, layout string) string {
	t := time.UnixMilli(timestamp)
	return t.Format(layout)
}

// NowMilli 获取当前毫秒时间戳
func NowMilli() int64 {
	return time.Now().UnixMilli()
}
```

#### 3.3 使用示例

**改造前**:

```go
func (h *SkillHandler) ListSkills(w http.ResponseWriter, req *http.Request) {
	pageStr := req.URL.Query().Get("page")
	pageSizeStr := req.URL.Query().Get("pageSize")
	page := 1
	pageSize := 10
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}
	// ... 后续逻辑
}
```

**改造后**:

```go
func (h *SkillHandler) ListSkills(w http.ResponseWriter, req *http.Request) {
	pagination := helpers.ParsePagination(req)

	result, err := h.service.ListSkills(ctx, services.ListSkillsRequest{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		// ...
	})
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}
	helpers.RenderSuccess(w, req, result)
}
```

#### 3.4 实施步骤

1. 创建 `internal/api/helpers/` 目录
2. 实现 `helpers.go` 和 `response.go`
3. 创建 `internal/utils/time.go`
4. 逐个Handler替换重复代码
5. 删除 `handlers/consts.go` 中的重复定义

---

## 四、P1阶段：性能优化

### 任务4：分词索引批量插入

#### 4.1 问题分析

当前 `buildSkillTokens` 函数逐条插入分词索引：

```go
// 当前实现：N次数据库操作
for term := range termMap {
    token := models.SkillToken{...}
    if err := tx.Create(&token).Error; err != nil {  // 每次一条SQL
        return err
    }
}
```

#### 4.2 解决方案

修改 `internal/repositories/skill_repository.go`：

```go
// buildSkillTokens 为技能建立分词索引（批量插入优化版）
// 参数:
//   - tx: 数据库事务
//   - skillID: 技能ID
//   - text: 需要分词的文本（名称+描述）
// 返回:
//   - error: 错误信息
func (r *Repository) buildSkillTokens(tx *gorm.DB, skillID uint, text string) error {
	// 对文本进行分词
	tokens := seg.Cut(text, true)

	// 去重后的分词集合
	termMap := make(map[string]bool)
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token != "" {
			termMap[strings.ToLower(token)] = true
		}
	}

	// 无分词则直接返回
	if len(termMap) == 0 {
		return nil
	}

	// 批量构建分词记录
	skillTokens := make([]models.SkillToken, 0, len(termMap))
	for term := range termMap {
		skillTokens = append(skillTokens, models.SkillToken{
			SkillID: skillID,
			Term:    term,
		})
	}

	// 批量插入（单次SQL操作）
	return tx.CreateInBatches(skillTokens, 100).Error
}
```

#### 4.3 性能对比

| 场景 | 改造前 | 改造后 | 提升 |
|------|--------|--------|------|
| 10个分词 | 10次INSERT | 1次INSERT | 10x |
| 100个分词 | 100次INSERT | 1次INSERT | 100x |
| 1000个分词 | 1000次INSERT | 10次INSERT(每批100) | 100x |

#### 4.4 实施步骤

1. 修改 `buildSkillTokens` 函数
2. 添加单元测试验证批量插入正确性
3. 性能基准测试对比

---

### 任务5：N+1查询修复

#### 5.1 问题分析

检查是否存在N+1查询问题：

```go
// 潜在N+1问题：查询技能后再查询标签
skills, _ := repo.ListAllSkills(ctx)
for _, skill := range skills {
    tags, _ := repo.GetTagsBySkillID(ctx, skill.ID)  // N次查询
}
```

#### 5.2 解决方案

当前代码已使用 `Preload` 预加载：

```go
// 正确做法：使用Preload预加载关联
err := r.db.WithContext(ctx).Preload("Tags").Find(&skills).Error
```

需要检查并确保所有关联查询都使用预加载：

```go
// 检查点1：技能列表查询
func (r *Repository) ListAllSkills(ctx context.Context) ([]models.Skill, error) {
	var skills []models.Skill
	// 添加Preload确保标签一起加载
	err := r.db.WithContext(ctx).Preload("Tags").Find(&skills).Error
	return skills, err
}

// 检查点2：技能搜索
func (r *Repository) SearchSkillsByTokens(ctx context.Context, keyword string) ([]models.Skill, error) {
	// ...
	err := r.db.WithContext(ctx).
		Preload("Tags").  // 添加预加载
		Select("skills.*, COUNT(skill_tokens.term) as match_score").
		Joins("JOIN skill_tokens ON skill_tokens.skill_id = skills.id").
		Where("skill_tokens.term IN ?", terms).
		Group("skills.id").
		Order("match_score DESC").
		Find(&skills).Error
	// ...
}
```

#### 5.3 实施步骤

1. 审查所有Repository查询方法
2. 为所有关联查询添加 `Preload`
3. 编写集成测试验证无N+1问题

---

## 五、P2阶段：增强功能

### 任务6：请求ID追踪中间件

#### 6.1 问题分析

当前日志无法关联同一请求的多次日志输出：

```
[INFO] 2024-01-01 10:00:00.123 获取技能列表
[INFO] 2024-01-01 10:00:00.456 查询数据库
[INFO] 2024-01-01 10:00:00.789 返回结果
```

无法判断这三条日志是否属于同一请求。

#### 6.2 解决方案

创建 `internal/api/middleware/request_id.go`：

```go
// Package middleware 提供HTTP中间件
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// 上下文键类型
type ctxKey string

const (
	// RequestIDKey 请求ID上下文键
	RequestIDKey ctxKey = "request_id"
	// RequestIDHeader 请求ID响应头
	RequestIDHeader = "X-Request-ID"
)

// RequestID 请求ID中间件
// 为每个请求生成唯一ID，便于日志追踪
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 尝试从请求头获取已有的请求ID（支持链路追踪）
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			// 生成新的请求ID
			requestID = uuid.New().String()[:8] // 使用前8位简化显示
		}

		// 设置到上下文
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		// 设置响应头
		w.Header().Set(RequestIDHeader, requestID)

		// 继续处理请求
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID 从上下文获取请求ID
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}
```

更新 `internal/utils/logx/log.go`：

```go
package logx

import (
	"context"
	"io"
	"log"
	"os"
	"strings"

	"aiflow/internal/api/middleware"
)

// 日志等级常量
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// 全局日志实例
var (
	DebugLogger *log.Logger
	InfoLogger  *log.Logger
	WarnLogger  *log.Logger
	ErrorLogger *log.Logger
)

// InitLogger 初始化日志系统
func InitLogger(logLevel string, outputType string, logDirPath string) {
	// ... 原有初始化逻辑 ...

	// 创建带请求ID前缀的日志实例
	DebugLogger = log.New(stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	InfoLogger = log.New(stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	WarnLogger = log.New(stdout, "[WARN] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	ErrorLogger = log.New(stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lmicroseconds)

	// ... 日志等级设置 ...
}

// DebugCtx 带上下文的调试日志
func DebugCtx(ctx context.Context, format string, v ...interface{}) {
	requestID := middleware.GetRequestID(ctx)
	if requestID != "" {
		format = "[%s] " + format
		v = append([]interface{}{requestID}, v...)
	}
	DebugLogger.Printf(format, v...)
}

// InfoCtx 带上下文的信息日志
func InfoCtx(ctx context.Context, format string, v ...interface{}) {
	requestID := middleware.GetRequestID(ctx)
	if requestID != "" {
		format = "[%s] " + format
		v = append([]interface{}{requestID}, v...)
	}
	InfoLogger.Printf(format, v...)
}

// WarnCtx 带上下文的警告日志
func WarnCtx(ctx context.Context, format string, v ...interface{}) {
	requestID := middleware.GetRequestID(ctx)
	if requestID != "" {
		format = "[%s] " + format
		v = append([]interface{}{requestID}, v...)
	}
	WarnLogger.Printf(format, v...)
}

// ErrorCtx 带上下文的错误日志
func ErrorCtx(ctx context.Context, format string, v ...interface{}) {
	requestID := middleware.GetRequestID(ctx)
	if requestID != "" {
		format = "[%s] " + format
		v = append([]interface{}{requestID}, v...)
	}
	ErrorLogger.Printf(format, v...)
}

// 保留原有无上下文的方法（向后兼容）
func Debug(format string, v ...interface{}) {
	DebugLogger.Printf(format, v...)
}

func Info(format string, v ...interface{}) {
	InfoLogger.Printf(format, v...)
}

func Warn(format string, v ...interface{}) {
	WarnLogger.Printf(format, v...)
}

func Error(format string, v ...interface{}) {
	ErrorLogger.Printf(format, v...)
}

func Fatal(format string, v ...interface{}) {
	ErrorLogger.Fatalf(format, v...)
}
```

#### 6.3 使用示例

```go
func (h *SkillHandler) ListSkills(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	logx.InfoCtx(ctx, "开始获取技能列表")

	result, err := h.service.ListSkills(ctx, services.ListSkillsRequest{...})
	if err != nil {
		logx.ErrorCtx(ctx, "获取技能列表失败: %v", err)
		helpers.RenderError(w, req, err)
		return
	}

	logx.InfoCtx(ctx, "成功获取技能列表，数量: %d", len(result.Items))
	helpers.RenderSuccess(w, req, result)
}
```

#### 6.4 日志输出效果

```
[INFO] 2024-01-01 10:00:00.123 [a1b2c3d4] 开始获取技能列表
[INFO] 2024-01-01 10:00:00.456 [a1b2c3d4] 成功获取技能列表，数量: 10
[INFO] 2024-01-01 10:00:01.123 [e5f6g7h8] 开始获取技能列表
[ERROR] 2024-01-01 10:00:01.456 [e5f6g7h8] 获取技能列表失败: connection refused
```

#### 6.5 实施步骤

1. 创建 `internal/api/middleware/request_id.go`
2. 更新 `internal/utils/logx/log.go` 添加Ctx方法
3. 在 `routers.go` 中注册中间件
4. 逐步替换Handler中的日志调用

---

### 任务7：配置管理优化

#### 7.1 问题分析

当前配置管理问题：
- 缺少配置验证
- 配置项分散
- 无环境变量支持

#### 7.2 解决方案

更新 `internal/config/config.go`：

```go
// Package config 管理应用配置
package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// 常量定义
const (
	DefaultAddr     = "localhost:8080"
	DefaultName     = "aiflow"
	DefaultVersion  = "0.5.0"
	DefaultLogLevel = "info"
	DBPath          = "./db/aiflow.db"
)

// Config 应用配置结构
type Config struct {
	Server ServerConfig `yaml:"server"`
	Log    LogConfig    `yaml:"log"`
	DB     DBConfig     `yaml:"db"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Addr     string `yaml:"addr"`
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	RootPath string `yaml:"root_path"`
	McpPath  string `yaml:"mcp_path"`
	WebPath  string `yaml:"web_path"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`       // debug, info, warn, error
	OutputType string `yaml:"output_type"` // std, file
	FilePath   string `yaml:"file_path"`   // 日志文件目录
}

// DBConfig 数据库配置
type DBConfig struct {
	Path string `yaml:"path"` // 数据库文件路径
}

// Validate 验证配置有效性
func (c *Config) Validate() error {
	// 验证服务器地址
	if c.Server.Addr != "" {
		if _, err := net.ResolveTCPAddr("tcp", c.Server.Addr); err != nil {
			return fmt.Errorf("无效的服务器地址 '%s': %w", c.Server.Addr, err)
		}
	}

	// 验证日志等级
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLogLevels[strings.ToLower(c.Log.Level)] {
		return fmt.Errorf("无效的日志等级 '%s'，可选: debug, info, warn, error", c.Log.Level)
	}

	// 验证输出类型
	validOutputTypes := map[string]bool{
		"std": true, "file": true,
	}
	if !validOutputTypes[strings.ToLower(c.Log.OutputType)] {
		return fmt.Errorf("无效的输出类型 '%s'，可选: std, file", c.Log.OutputType)
	}

	return nil
}

// ApplyDefaults 应用默认值
func (c *Config) ApplyDefaults() {
	if c.Server.Addr == "" {
		c.Server.Addr = DefaultAddr
	}
	if c.Server.Name == "" {
		c.Server.Name = DefaultName
	}
	if c.Server.Version == "" {
		c.Server.Version = DefaultVersion
	}
	if c.Server.RootPath == "" {
		c.Server.RootPath = "/"
	}
	if c.Server.McpPath == "" {
		c.Server.McpPath = "/mcp"
	}
	if c.Server.WebPath == "" {
		c.Server.WebPath = "/web"
	}
	if c.Log.Level == "" {
		c.Log.Level = DefaultLogLevel
	}
	if c.Log.OutputType == "" {
		c.Log.OutputType = "std"
	}
	if c.DB.Path == "" {
		c.DB.Path = DBPath
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	// 读取配置文件
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 配置文件不存在，创建默认配置
			return createDefaultConfig(path)
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 应用默认值
	cfg.ApplyDefaults()

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &cfg, nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Addr:     "localhost:9990",
			Name:     "智流",
			Version:  "0.5.0",
			RootPath: "/",
			McpPath:  "/mcp",
			WebPath:  "/web",
		},
		Log: LogConfig{
			Level:      "info",
			OutputType: "file",
			FilePath:   "./logs",
		},
		DB: DBConfig{
			Path: DBPath,
		},
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建配置目录失败: %w", err)
		}
	}

	// 写入默认配置
	content, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("序列化配置失败: %w", err)
	}

	header := `# 智流MCP配置文件
# 自动生成于首次启动

`
	if err := os.WriteFile(path, append([]byte(header), content...), 0644); err != nil {
		return nil, fmt.Errorf("写入配置文件失败: %w", err)
	}

	return cfg, nil
}

// LoadFromEnv 从环境变量加载配置（覆盖文件配置）
func (c *Config) LoadFromEnv() {
	if addr := os.Getenv("AIFLOW_ADDR"); addr != "" {
		c.Server.Addr = addr
	}
	if level := os.Getenv("AIFLOW_LOG_LEVEL"); level != "" {
		c.Log.Level = level
	}
	if output := os.Getenv("AIFLOW_LOG_OUTPUT"); output != "" {
		c.Log.OutputType = output
	}
	if dbPath := os.Getenv("AIFLOW_DB_PATH"); dbPath != "" {
		c.DB.Path = dbPath
	}
}
```

#### 7.3 使用示例

```go
func main() {
	// 加载配置
	cfg, err := config.LoadConfig("./config.yml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 支持环境变量覆盖
	cfg.LoadFromEnv()

	// 初始化日志
	logx.InitLogger(cfg.Log.Level, cfg.Log.OutputType, cfg.Log.FilePath)

	// ...
}
```

#### 7.4 实施步骤

1. 更新 `internal/config/config.go`
2. 添加配置验证逻辑
3. 添加环境变量支持
4. 更新 `main.go` 使用新配置

---

## 六、具体改造点清单（文件-函数-行号）

### P0阶段：任务1 统一错误处理

#### 1.1 新建文件

| 文件路径 | 说明 |
|----------|------|
| `goend/internal/errors/errors.go` | 统一错误定义，包含错误码、AppError结构体、构造函数 |

#### 1.2 handlers/skill.go 改造点（共18处）

| 函数名 | 行号 | 当前错误消息 | 改造后错误码 |
|--------|------|--------------|--------------|
| ListSkills | 82-89 | "无效的标签ID参数" | ErrInvalidID |
| ListSkills | 102-108 | "获取技能失败" | ErrInternal |
| CreateSkill | 120-127 | "请求参数错误" | ErrBadRequest |
| CreateSkill | 142-148 | "创建技能失败" | ErrSkillCreate |
| GetSkill | 163-170 | "无效的ID参数" | ErrInvalidID |
| GetSkill | 172-180 | "技能不存在" | ErrSkillNotFound |
| UpdateSkill | 191-199 | "无效的ID参数" | ErrInvalidID |
| UpdateSkill | 202-209 | "请求参数错误" | ErrBadRequest |
| UpdateSkill | 225-232 | "更新技能失败" | ErrSkillUpdate |
| DeleteSkill | 244-252 | "无效的ID参数" | ErrInvalidID |
| DeleteSkill | 254-261 | "删除技能失败" | ErrSkillDelete |
| ListDeletedSkills | 289-297 | "获取回收站列表失败" | ErrInternal |
| RestoreSkill | 308-316 | "无效的ID参数" | ErrInvalidID |
| RestoreSkill | 318-333 | "技能不存在或不在回收站中" | ErrSkillNotInTrash |
| PermanentDeleteSkill | 343-352 | "无效的ID参数" | ErrInvalidID |
| PermanentDeleteSkill | 354-369 | "技能不存在或不在回收站中" | ErrSkillNotInTrash |
| ExportSkills | 382-388 | "必须提供技能ID参数" | ErrInvalidParam |
| ExportSkills | 390-398 | "无效的ID参数" | ErrInvalidID |

#### 1.3 handlers/jobtask.go 改造点（共22处）

| 函数名 | 行号 | 当前错误消息 | 改造后错误码 |
|--------|------|--------------|--------------|
| ListJobTasks | 83-91 | "获取任务列表失败" | ErrInternal |
| CreateJobTask | 115-122 | "请求参数错误" | ErrBadRequest |
| CreateJobTask | 124-132 | "任务编号、所属项目、任务类型、任务目标不能为空" | ErrTaskEmptyField |
| CreateJobTask | 151-158 | "创建任务失败" | ErrTaskCreate |
| GetJobTask | 171-179 | "无效的ID参数" | ErrInvalidID |
| GetJobTask | 181-189 | "任务不存在" | ErrTaskNotFound |
| UpdateJobTask | 200-208 | "无效的ID参数" | ErrInvalidID |
| UpdateJobTask | 210-218 | "请求参数错误" | ErrBadRequest |
| UpdateJobTask | 220-228 | "任务不存在" | ErrTaskNotFound |
| UpdateJobTask | 239-246 | "更新任务失败" | ErrTaskUpdate |
| DeleteJobTask | 258-266 | "无效的ID参数" | ErrInvalidID |
| DeleteJobTask | 268-275 | "删除任务失败" | ErrTaskDelete |
| ListDeletedJobTasks | 306-314 | "获取回收站列表失败" | ErrInternal |
| RestoreJobTask | 338-346 | "无效的ID参数" | ErrInvalidID |
| RestoreJobTask | 348-364 | "任务不存在或不在回收站中" | ErrTaskNotInTrash |
| PermanentDeleteJobTask | 373-382 | "无效的ID参数" | ErrInvalidID |
| PermanentDeleteJobTask | 384-399 | "任务不存在或不在回收站中" | ErrTaskNotInTrash |
| GetAllJobTaskProjects | 409-417 | "获取项目列表失败" | ErrInternal |
| BatchExportJobTasks | 435-442 | "请求参数错误" | ErrBadRequest |
| BatchExportJobTasks | 449-457 | "不支持的导出格式..." | ErrInvalidParam |
| BatchExportJobTasks | 468-475 | "获取任务数据失败" | ErrInternal |

#### 1.4 handlers/skill_tag.go 改造点（共11处）

| 函数名 | 行号 | 当前错误消息 | 改造后错误码 |
|--------|------|--------------|--------------|
| ListTags | 53-61 | "获取标签失败" | ErrInternal |
| CreateTag | 85-92 | "请求参数错误" | ErrBadRequest |
| CreateTag | 101-108 | "创建标签失败" | ErrTagCreate |
| GetTag | 121-129 | "无效的ID参数" | ErrInvalidID |
| GetTag | 131-139 | "标签不存在" | ErrTagNotFound |
| UpdateTag | 150-158 | "无效的ID参数" | ErrInvalidID |
| UpdateTag | 161-168 | "请求参数错误" | ErrBadRequest |
| UpdateTag | 170-178 | "标签不存在" | ErrTagNotFound |
| UpdateTag | 182-189 | "更新标签失败" | ErrTagUpdate |
| DeleteTag | 201-209 | "无效的ID参数" | ErrInvalidID |
| DeleteTag | 211-218 | "删除标签失败" | ErrTagDelete |

---

### P0阶段：任务2 统一架构层次

#### 2.1 新建文件

| 文件路径 | 说明 |
|----------|------|
| `goend/internal/services/jobtask_service.go` | 任务服务层，封装业务逻辑 |
| `goend/internal/services/tag_service.go` | 标签服务层，封装业务逻辑 |

#### 2.2 jobtask_service.go 需要实现的方法

```go
// JobTaskService 任务服务层
type JobTaskService struct {
    repo *repositories.Repository
}

// 需要实现的方法列表
func NewJobTaskService(repo *repositories.Repository) *JobTaskService
func (s *JobTaskService) ListJobTasks(ctx context.Context, req ListJobTasksRequest) (*ListJobTasksResponse, error)
func (s *JobTaskService) CreateJobTask(ctx context.Context, req CreateJobTaskRequest) (*models.JobTask, error)
func (s *JobTaskService) GetJobTask(ctx context.Context, id uint) (*models.JobTask, error)
func (s *JobTaskService) UpdateJobTask(ctx context.Context, req UpdateJobTaskRequest) (*models.JobTask, error)
func (s *JobTaskService) DeleteJobTask(ctx context.Context, id uint) error
func (s *JobTaskService) ListDeletedJobTasks(ctx context.Context, page, pageSize int) (*ListJobTasksResponse, error)
func (s *JobTaskService) RestoreJobTask(ctx context.Context, id uint) error
func (s *JobTaskService) PermanentDeleteJobTask(ctx context.Context, id uint) error
func (s *JobTaskService) GetAllProjects(ctx context.Context) ([]string, error)
func (s *JobTaskService) GetJobTasksForExport(ctx context.Context, ids []uint) ([]models.JobTask, error)
```

#### 2.3 需要迁移的业务逻辑

| 来源文件 | 来源函数 | 迁移内容 | 目标位置 |
|----------|----------|----------|----------|
| handlers/jobtask.go:124-132 | CreateJobTask | 必填字段验证逻辑 | services/jobtask_service.go |
| handlers/jobtask.go:146-149 | CreateJobTask | 默认状态设置 | services/jobtask_service.go |
| handlers/jobtask.go:134-144 | CreateJobTask | 时间戳生成 | services/jobtask_service.go |
| handlers/jobtask.go:94-99 | ListJobTasks | 分页响应构造 | services/jobtask_service.go |

#### 2.4 需要修改的依赖注入

| 文件路径 | 修改内容 |
|----------|----------|
| `handlers/jobtask.go:31-37` | JobTaskHandler 改为依赖 JobTaskService |
| `handlers/skill_tag.go:21-28` | TagHandler 改为依赖 TagService |
| `api/routers.go:20-29` | NewRouter 中初始化 Service 层 |

---

### P0阶段：任务3 代码复用抽象

#### 3.1 新建文件

| 文件路径 | 说明 |
|----------|------|
| `goend/internal/api/helpers/helpers.go` | 参数解析辅助函数 |
| `goend/internal/api/helpers/response.go` | 响应构造辅助函数 |
| `goend/internal/utils/time.go` | 时间处理工具函数 |

#### 3.2 分页参数解析重复代码（5处）

| 文件 | 函数 | 行号 | 改造方案 |
|------|------|------|----------|
| handlers/skill.go | ListSkills | 46-63 | 调用 helpers.ParsePagination(req) |
| handlers/skill.go | ListDeletedSkills | 270-287 | 调用 helpers.ParsePagination(req) |
| handlers/jobtask.go | ListJobTasks | 50-67 | 调用 helpers.ParsePagination(req) |
| handlers/jobtask.go | ListDeletedJobTasks | 286-303 | 调用 helpers.ParsePagination(req) |
| handlers/skill_tag.go | ListTags | 32-50 | 调用 helpers.ParsePagination(req) |

#### 3.3 ID参数解析重复代码（10处）

| 文件 | 函数 | 行号 | 改造方案 |
|------|------|------|----------|
| handlers/skill.go | GetSkill | 161-170 | 调用 helpers.ParseIDParam(req, "id") |
| handlers/skill.go | UpdateSkill | 189-199 | 调用 helpers.ParseIDParam(req, "id") |
| handlers/skill.go | DeleteSkill | 242-252 | 调用 helpers.ParseIDParam(req, "id") |
| handlers/skill.go | RestoreSkill | 306-316 | 调用 helpers.ParseIDParam(req, "id") |
| handlers/skill.go | PermanentDeleteSkill | 342-352 | 调用 helpers.ParseIDParam(req, "id") |
| handlers/jobtask.go | GetJobTask | 169-179 | 调用 helpers.ParseIDParam(req, "id") |
| handlers/jobtask.go | UpdateJobTask | 198-208 | 调用 helpers.ParseIDParam(req, "id") |
| handlers/jobtask.go | DeleteJobTask | 256-266 | 调用 helpers.ParseIDParam(req, "id") |
| handlers/jobtask.go | RestoreJobTask | 336-346 | 调用 helpers.ParseIDParam(req, "id") |
| handlers/jobtask.go | PermanentDeleteJobTask | 372-382 | 调用 helpers.ParseIDParam(req, "id") |
| handlers/skill_tag.go | GetTag | 119-129 | 调用 helpers.ParseIDParam(req, "id") |
| handlers/skill_tag.go | UpdateTag | 148-158 | 调用 helpers.ParseIDParam(req, "id") |
| handlers/skill_tag.go | DeleteTag | 199-209 | 调用 helpers.ParseIDParam(req, "id") |

#### 3.4 helpers包函数清单

```go
// helpers/helpers.go
func ParsePagination(req *http.Request) PaginationParams
func ParseIDParam(req *http.Request, paramName string) (uint, error)
func ParseIntParam(req *http.Request, paramName string, defaultValue int64) int64
func ParseUintParam(req *http.Request, paramName string) (uint, error)

// helpers/response.go
func RenderSuccess(w http.ResponseWriter, req *http.Request, data interface{})
func RenderSuccessWithMessage(w http.ResponseWriter, req *http.Request, message string, data interface{})
func RenderCreated(w http.ResponseWriter, req *http.Request, message string, data interface{})
func RenderError(w http.ResponseWriter, req *http.Request, err error)
func RenderPaginated(w http.ResponseWriter, req *http.Request, items interface{}, total int64, page, pageSize int)

// utils/time.go
func FormatTimestamp(timestamp int64) string
func NowMilli() int64
```

#### 3.5 可删除的重复代码

| 文件 | 行号 | 内容 | 处理方式 |
|------|------|------|----------|
| handlers/consts.go | 4-11 | 分页常量 | 迁移到 helpers/helpers.go |
| handlers/jobtask.go | 652-656 | formatTimestamp函数 | 迁移到 utils/time.go |

---

### P1阶段：任务4 分词索引批量插入

#### 4.1 需要修改的文件

| 文件路径 | 函数名 | 行号 |
|----------|--------|------|
| `goend/internal/repositories/skill_repository.go` | buildSkillTokens | 142-177 |

#### 4.2 改造前后对比

```go
// 改造前（第166-175行）：逐条插入
for term := range termMap {
    token := models.SkillToken{
        SkillID: skillID,
        Term:    term,
    }
    if err := tx.Create(&token).Error; err != nil {  // N次INSERT
        return err
    }
}

// 改造后：批量插入
skillTokens := make([]models.SkillToken, 0, len(termMap))
for term := range termMap {
    skillTokens = append(skillTokens, models.SkillToken{
        SkillID: skillID,
        Term:    term,
    })
}
return tx.CreateInBatches(skillTokens, 100).Error  // 单次批量INSERT
```

---

### P1阶段：任务5 N+1查询修复

#### 5.1 需要修改的文件

| 文件路径 | 函数名 | 行号 | 问题 |
|----------|--------|------|------|
| `goend/internal/repositories/skill_repository.go` | ListAllSkills | 73-77 | 缺少Preload("Tags") |
| `goend/internal/repositories/skill_repository.go` | SearchSkillsByTokens | 189-217 | 缺少Preload("Tags") |

#### 5.2 改造方案

```go
// ListAllSkills 改造（第73-77行）
func (r *Repository) ListAllSkills(ctx context.Context) ([]models.Skill, error) {
    var skills []models.Skill
    err := r.db.WithContext(ctx).Preload("Tags").Find(&skills).Error  // 添加Preload
    return skills, err
}

// SearchSkillsByTokens 改造（第207-214行）
err := r.db.WithContext(ctx).
    Preload("Tags").  // 添加Preload
    Select("skills.*, COUNT(skill_tokens.term) as match_score").
    Joins("JOIN skill_tokens ON skill_tokens.skill_id = skills.id").
    Where("skill_tokens.term IN ?", terms).
    Group("skills.id").
    Order("match_score DESC").
    Find(&skills).Error
```

---

### P2阶段：任务6 请求ID追踪中间件

#### 6.1 新建文件

| 文件路径 | 说明 |
|----------|------|
| `goend/internal/api/middleware/request_id.go` | 请求ID中间件 |

#### 6.2 需要修改的文件

| 文件路径 | 修改内容 |
|----------|----------|
| `goend/internal/utils/logx/log.go` | 添加 InfoCtx、ErrorCtx 等带上下文的日志方法 |
| `goend/internal/api/routers.go` | 注册 RequestID 中间件 |

#### 6.3 中间件实现

```go
// middleware/request_id.go
func RequestID(next http.Handler) http.Handler
func GetRequestID(ctx context.Context) string

// logx/log.go 新增方法
func InfoCtx(ctx context.Context, format string, v ...interface{})
func ErrorCtx(ctx context.Context, format string, v ...interface{})
func DebugCtx(ctx context.Context, format string, v ...interface{})
func WarnCtx(ctx context.Context, format string, v ...interface{})
```

---

### P2阶段：任务7 配置管理优化

#### 7.1 需要修改的文件

| 文件路径 | 修改内容 |
|----------|----------|
| `goend/internal/config/config.go` | 添加Validate方法、环境变量支持、DB配置 |

#### 7.2 新增配置结构

```go
// 新增DBConfig结构
type DBConfig struct {
    Path string `yaml:"path"`
}

// Config 新增DB字段
type Config struct {
    Server ServerConfig `yaml:"server"`
    Log    LogConfig    `yaml:"log"`
    DB     DBConfig     `yaml:"db"`  // 新增
}

// 新增方法
func (c *Config) Validate() error
func (c *Config) LoadFromEnv()
```

#### 7.3 环境变量映射

| 环境变量 | 配置字段 |
|----------|----------|
| AIFLOW_ADDR | Server.Addr |
| AIFLOW_LOG_LEVEL | Log.Level |
| AIFLOW_LOG_OUTPUT | Log.OutputType |
| AIFLOW_DB_PATH | DB.Path |

---

## 七、重构检查清单

### P0阶段检查清单

- [ ] 创建 `goend/internal/errors/errors.go`
- [ ] 定义错误码常量（通用、技能、任务、标签四类）
- [ ] 实现 `AppError` 结构体及Error()方法
- [ ] 实现错误构造函数（NewSkillError、NewTaskError等）
- [ ] 创建 `goend/internal/api/helpers/helpers.go`
- [ ] 实现 `ParsePagination` 函数
- [ ] 实现 `ParseIDParam` 函数
- [ ] 创建 `goend/internal/api/helpers/response.go`
- [ ] 实现 `RenderSuccess/RenderError/RenderPaginated` 函数
- [ ] 创建 `goend/internal/services/jobtask_service.go`
- [ ] 创建 `goend/internal/services/tag_service.go`
- [ ] 迁移任务业务逻辑到Service层
- [ ] 更新 `JobTaskHandler` 依赖注入
- [ ] 更新 `TagHandler` 依赖注入
- [ ] 更新 `api/routers.go` 依赖注入
- [ ] 创建 `goend/internal/utils/time.go`
- [ ] 所有Handler替换为使用helpers（共51处改造点）
- [ ] 删除 `handlers/consts.go` 中的重复常量
- [ ] 编译检查通过
- [ ] 单元测试通过

### P1阶段检查清单

- [ ] 修改 `buildSkillTokens` 为批量插入
- [ ] 添加批量插入单元测试
- [ ] 性能基准测试对比
- [ ] 修改 `ListAllSkills` 添加Preload
- [ ] 修改 `SearchSkillsByTokens` 添加Preload
- [ ] 审查其他Repository查询方法
- [ ] 编译检查通过
- [ ] 集成测试通过

### P2阶段检查清单

- [ ] 创建 `goend/internal/api/middleware/request_id.go`
- [ ] 更新 `goend/internal/utils/logx/log.go` 添加Ctx方法
- [ ] 在 `api/routers.go` 注册中间件
- [ ] 更新 `goend/internal/config/config.go`
- [ ] 添加 `Validate` 方法
- [ ] 添加 `LoadFromEnv` 方法
- [ ] 添加 DBConfig 结构
- [ ] 更新 `main.go` 配置加载逻辑
- [ ] 编译检查通过
- [ ] 功能测试通过

---

## 八、预期收益表

| 收益项 | 改进前 | 改进后 | 量化指标 |
|--------|--------|--------|----------|
| 错误定位效率 | 需搜索日志关键词 | 通过错误码精确定位 | 定位时间-70% |
| 新功能开发效率 | 需编写重复代码 | 使用helpers快速开发 | 开发时间-30% |
| 代码维护成本 | 修改需同步多处 | 单点修改全局生效 | 维护成本-50% |
| 分词索引性能 | N次数据库操作 | 1次批量操作 | 性能提升10-100倍 |
| 问题排查效率 | 无法关联请求日志 | 请求ID全链路追踪 | 排查效率+80% |
| 配置错误率 | 运行时才发现 | 启动时验证 | 错误率-90% |
| 代码行数 | ~3000行 | ~2800行 | 减少~200行重复代码 |
| 测试覆盖率 | ~40% | ~60% | 提升20% |

---

## 九、风险与应对

| 风险 | 影响 | 应对措施 |
|------|------|----------|
| API兼容性破坏 | 前端调用失败 | 保持响应结构不变，仅内部重构 |
| 依赖注入遗漏 | 运行时panic | 编译时检查，单元测试覆盖 |
| 性能回归 | 用户体验下降 | 基准测试对比，灰度发布 |
| 配置迁移失败 | 服务无法启动 | 保留默认值回退机制 |

---

## 十、总结

本重构方案遵循"渐进式、向后兼容、最小改动"原则，通过三个阶段逐步提升代码质量：

1. **P0阶段**：解决核心架构问题，统一错误处理和分层架构
2. **P1阶段**：优化性能，解决分词索引和N+1查询问题
3. **P2阶段**：增强功能，添加请求追踪和配置验证

每个阶段独立可验收，确保重构过程可控、可回滚。预期完成后，代码可维护性提升50%以上，问题排查效率提升80%。
