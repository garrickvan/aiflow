package services

import (
	"aiflow/internal/errors"
	"aiflow/internal/models"
	"aiflow/internal/repositories"
	"context"
	"time"

	"gorm.io/gorm"
)

// 任务状态常量定义
const (
	JobTaskStatusCreated   = "已创建"
	JobTaskStatusRunning   = "处理中"
	JobTaskStatusFailed    = "处理失败"
	JobTaskStatusCompleted = "处理完成"
	JobTaskStatusPassed    = "验收通过"
)

// JobTaskService 任务服务层
// 处理任务相关的业务逻辑，将业务逻辑从handler中分离
type JobTaskService struct {
	repo *repositories.Repository
}

// NewJobTaskService 创建任务服务实例
func NewJobTaskService(repo *repositories.Repository) *JobTaskService {
	return &JobTaskService{repo: repo}
}

// ListJobTasksRequest 获取任务列表请求参数
type ListJobTasksRequest struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
	Project   string `json:"project"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	StartDate int64  `json:"startDate"`
	EndDate   int64  `json:"endDate"`
}

// JobTaskResponse 任务响应结构
type JobTaskResponse struct {
	ID                      uint   `json:"id"`
	JobNo                   string `json:"jobNo"`
	Project                 string `json:"project"`
	Type                    string `json:"type"`
	Goal                    string `json:"goal"`
	PassAcceptStd           bool   `json:"passAcceptStd"`
	Status                  string `json:"status"`
	ExecutionRecords        string `json:"executionRecords"`
	ActiveExecutionSequence int    `json:"activeExecutionSequence"`
	CreatedAt               int64  `json:"createdAt"`
	UpdatedAt               int64  `json:"updatedAt"`
}

// ListJobTasksResponse 获取任务列表响应
type ListJobTasksResponse struct {
	Items      []JobTaskResponse       `json:"items"`
	Pagination map[string]interface{} `json:"pagination"`
}

// CreateJobTaskRequest 创建任务请求参数
type CreateJobTaskRequest struct {
	JobNo                   string `json:"jobNo"`
	Project                 string `json:"project"`
	Type                    string `json:"type"`
	Goal                    string `json:"goal"`
	PassAcceptStd           bool   `json:"passAcceptStd"`
	ExecutionRecords        string `json:"executionRecords"`
	ActiveExecutionSequence int    `json:"activeExecutionSequence"`
}

// UpdateJobTaskRequest 更新任务请求参数
type UpdateJobTaskRequest struct {
	ID                      uint   `json:"id"`
	Status                  string `json:"status"`
	PassAcceptStd           bool   `json:"passAcceptStd"`
	ExecutionRecords        string `json:"executionRecords"`
	ActiveExecutionSequence int    `json:"activeExecutionSequence"`
}

// BatchExportRequest 批量导出请求参数
type BatchExportRequest struct {
	IDs []uint `json:"ids"`
}

// ListJobTasks 获取任务列表（支持分页、项目、类型、状态筛选和日期范围筛选）
func (s *JobTaskService) ListJobTasks(ctx context.Context, req ListJobTasksRequest) (*ListJobTasksResponse, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	jobTasks, total, err := s.repo.ListJobTasks(ctx, req.Page, req.PageSize, req.Project, req.Type, req.Status, req.StartDate, req.EndDate)
	if err != nil {
		return nil, errors.NewInternalError(errors.ErrCodeInternalError, "获取任务列表失败", err)
	}

	// 转换响应格式
	items := make([]JobTaskResponse, 0, len(jobTasks))
	for _, task := range jobTasks {
		items = append(items, convertToJobTaskResponse(&task))
	}

	return &ListJobTasksResponse{
		Items:      items,
		Pagination: buildPagination(total, req.Page, req.PageSize),
	}, nil
}

// CreateJobTask 创建任务
func (s *JobTaskService) CreateJobTask(ctx context.Context, req CreateJobTaskRequest) (*models.JobTask, error) {
	// 验证必填字段
	if err := validateJobTaskRequired(req.JobNo, req.Project, req.Type, req.Goal); err != nil {
		return nil, err
	}

	// 检查任务编号是否已存在
	existing, err := s.repo.GetJobTaskByJobNo(ctx, req.JobNo)
	if err == nil && existing != nil {
		return nil, errors.NewTaskError(errors.ErrCodeTaskCreate, "任务编号已存在", nil)
	}

	timestamp := time.Now().UnixMilli()
	jobTask := &models.JobTask{
		JobNo:                   req.JobNo,
		Project:                 req.Project,
		Type:                    req.Type,
		Goal:                    req.Goal,
		PassAcceptStd:           req.PassAcceptStd,
		Status:                  JobTaskStatusCreated,
		ExecutionRecords:        req.ExecutionRecords,
		ActiveExecutionSequence: req.ActiveExecutionSequence,
		CreatedAt:               timestamp,
		UpdatedAt:               timestamp,
	}

	if err := s.repo.CreateJobTask(ctx, jobTask); err != nil {
		return nil, errors.NewTaskError(errors.ErrCodeTaskCreate, "创建任务失败", err)
	}

	return jobTask, nil
}

// GetJobTask 根据ID获取任务
func (s *JobTaskService) GetJobTask(ctx context.Context, id uint) (*models.JobTask, error) {
	jobTask, err := s.repo.GetJobTaskByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError(errors.ErrCodeTaskNotFound, "任务不存在", err)
		}
		return nil, errors.NewInternalError(errors.ErrCodeInternalError, "获取任务失败", err)
	}
	return jobTask, nil
}

// UpdateJobTask 更新任务
func (s *JobTaskService) UpdateJobTask(ctx context.Context, req UpdateJobTaskRequest) (*models.JobTask, error) {
	jobTask, err := s.repo.GetJobTaskByID(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError(errors.ErrCodeTaskNotFound, "任务不存在", err)
		}
		return nil, errors.NewInternalError(errors.ErrCodeInternalError, "获取任务失败", err)
	}

	// 更新允许修改的字段
	jobTask.Status = req.Status
	jobTask.PassAcceptStd = req.PassAcceptStd
	jobTask.ExecutionRecords = req.ExecutionRecords
	jobTask.ActiveExecutionSequence = req.ActiveExecutionSequence
	jobTask.UpdatedAt = time.Now().UnixMilli()

	if err := s.repo.UpdateJobTask(ctx, jobTask); err != nil {
		return nil, errors.NewTaskError(errors.ErrCodeTaskUpdate, "更新任务失败", err)
	}

	return jobTask, nil
}

// DeleteJobTask 删除任务（伪删除）
func (s *JobTaskService) DeleteJobTask(ctx context.Context, id uint) error {
	// 检查任务是否存在
	_, err := s.repo.GetJobTaskByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError(errors.ErrCodeTaskNotFound, "任务不存在", err)
		}
		return errors.NewInternalError(errors.ErrCodeInternalError, "获取任务失败", err)
	}

	if err := s.repo.DeleteJobTask(ctx, id); err != nil {
		return errors.NewTaskError(errors.ErrCodeTaskDelete, "删除任务失败", err)
	}
	return nil
}

// ListDeletedJobTasks 获取回收站任务列表
func (s *JobTaskService) ListDeletedJobTasks(ctx context.Context, page, pageSize int) (*ListJobTasksResponse, error) {
	// 设置默认分页参数
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	jobTasks, total, err := s.repo.ListDeletedJobTasks(ctx, page, pageSize)
	if err != nil {
		return nil, errors.NewInternalError(errors.ErrCodeInternalError, "获取回收站列表失败", err)
	}

	// 转换响应格式
	items := make([]JobTaskResponse, 0, len(jobTasks))
	for _, task := range jobTasks {
		items = append(items, convertToJobTaskResponse(&task))
	}

	return &ListJobTasksResponse{
		Items:      items,
		Pagination: buildPagination(total, page, pageSize),
	}, nil
}

// RestoreJobTask 恢复回收站中的任务
func (s *JobTaskService) RestoreJobTask(ctx context.Context, id uint) error {
	if err := s.repo.RestoreJobTask(ctx, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError(errors.ErrCodeTaskNotFound, "任务不存在或未删除", err)
		}
		return errors.NewTaskError(errors.ErrCodeTaskTrash, "恢复任务失败", err)
	}
	return nil
}

// PermanentDeleteJobTask 彻底删除任务
func (s *JobTaskService) PermanentDeleteJobTask(ctx context.Context, id uint) error {
	if err := s.repo.PermanentDeleteJobTask(ctx, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError(errors.ErrCodeTaskNotFound, "任务不存在或未删除", err)
		}
		return errors.NewTaskError(errors.ErrCodeTaskDelete, "彻底删除任务失败", err)
	}
	return nil
}

// GetAllProjects 获取所有项目列表
func (s *JobTaskService) GetAllProjects(ctx context.Context) ([]string, error) {
	projects, err := s.repo.GetAllJobTaskProjects(ctx)
	if err != nil {
		return nil, errors.NewInternalError(errors.ErrCodeInternalError, "获取项目列表失败", err)
	}
	return projects, nil
}

// GetJobTasksForExport 根据ID列表批量获取任务用于导出
// 如果IDs为空，则获取所有未删除的任务
func (s *JobTaskService) GetJobTasksForExport(ctx context.Context, ids []uint) ([]models.JobTask, error) {
	var jobTasks []models.JobTask
	var err error

	if len(ids) > 0 {
		jobTasks, err = s.repo.GetJobTasksByIDs(ctx, ids)
	} else {
		jobTasks, err = s.repo.GetAllJobTasks(ctx)
	}

	if err != nil {
		return nil, errors.NewInternalError(errors.ErrCodeInternalError, "获取任务失败", err)
	}
	return jobTasks, nil
}

// validateJobTaskRequired 验证任务必填字段
func validateJobTaskRequired(jobNo, project, jobType, goal string) error {
	if jobNo == "" {
		return errors.NewTaskError(errors.ErrCodeTaskValidate, "任务编号不能为空", nil)
	}
	if project == "" {
		return errors.NewTaskError(errors.ErrCodeTaskValidate, "所属项目不能为空", nil)
	}
	if jobType == "" {
		return errors.NewTaskError(errors.ErrCodeTaskValidate, "任务类型不能为空", nil)
	}
	if goal == "" {
		return errors.NewTaskError(errors.ErrCodeTaskValidate, "任务目标不能为空", nil)
	}
	return nil
}

// convertToJobTaskResponse 将模型转换为响应结构
func convertToJobTaskResponse(task *models.JobTask) JobTaskResponse {
	return JobTaskResponse{
		ID:                      task.ID,
		JobNo:                   task.JobNo,
		Project:                 task.Project,
		Type:                    task.Type,
		Goal:                    task.Goal,
		PassAcceptStd:           task.PassAcceptStd,
		Status:                  task.Status,
		ExecutionRecords:        task.ExecutionRecords,
		ActiveExecutionSequence: task.ActiveExecutionSequence,
		CreatedAt:               task.CreatedAt,
		UpdatedAt:               task.UpdatedAt,
	}
}

// buildPagination 构建分页响应
func buildPagination(total int64, page, pageSize int) map[string]interface{} {
	return map[string]interface{}{
		"total":     total,
		"page":      page,
		"pageSize":  pageSize,
		"totalPage": (total + int64(pageSize) - 1) / int64(pageSize),
	}
}
