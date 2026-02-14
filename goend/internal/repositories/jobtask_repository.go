package repositories

import (
	"aiflow/internal/cache"
	"aiflow/internal/models"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// JobTask缓存相关常量定义
const (
	// jobTaskProjectCacheTTL 项目列表缓存过期时间
	jobTaskProjectCacheTTL = 5 * time.Minute
	// jobTaskProjectCacheMaxSize 项目缓存最大条目数
	jobTaskProjectCacheMaxSize = 100
)

// jobTaskProjectListCacheKey 生成项目列表缓存Key
func jobTaskProjectListCacheKey() string {
	return "jobtasks:projects:all"
}

// jobTaskProjectCache 项目缓存实例（JobTask专用）
var jobTaskProjectCache = cache.NewLocalCache(jobTaskProjectCacheMaxSize)

// clearJobTaskProjectCache 清除JobTask项目相关缓存
func clearJobTaskProjectCache() {
	jobTaskProjectCache.Delete(jobTaskProjectListCacheKey())
}

// JobTask CRUD 操作

// CreateJobTask 创建任务
func (r *Repository) CreateJobTask(ctx context.Context, jobTask *models.JobTask) error {
	// 设置时间戳，毫秒级精度
	timestamp := time.Now().UnixMilli()
	jobTask.CreatedAt = timestamp
	jobTask.UpdatedAt = timestamp

	err := r.db.WithContext(ctx).Create(jobTask).Error
	if err != nil {
		return err
	}

	// 清除项目列表缓存（新增任务可能引入新项目）
	clearJobTaskProjectCache()
	return nil
}

// GetJobTaskByID 根据ID获取任务（不包含已删除的）
func (r *Repository) GetJobTaskByID(ctx context.Context, id uint) (*models.JobTask, error) {
	var jobTask models.JobTask
	err := r.db.WithContext(ctx).Where("deleted_at = ?", 0).First(&jobTask, id).Error
	if err != nil {
		return nil, err
	}
	return &jobTask, nil
}

// GetJobTaskByJobNo 根据任务编号获取任务（不包含已删除的）
func (r *Repository) GetJobTaskByJobNo(ctx context.Context, jobNo string) (*models.JobTask, error) {
	// 检查数据库连接是否初始化
	if r.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var jobTask models.JobTask
	err := r.db.WithContext(ctx).Where("job_no = ? AND deleted_at = ?", jobNo, 0).First(&jobTask).Error
	if err != nil {
		return nil, err
	}
	return &jobTask, nil
}

// ListJobTasks 分页获取任务列表（不包含已删除的），支持项目、类型、状态多条件筛选和日期范围筛选
func (r *Repository) ListJobTasks(ctx context.Context, page, pageSize int, project, jobType, status string, startDate, endDate int64) ([]models.JobTask, int64, error) {
	var jobTasks []models.JobTask
	var total int64

	query := r.db.WithContext(ctx).Model(&models.JobTask{}).Where("deleted_at = ?", 0)

	// 如果指定了项目筛选条件
	if project != "" {
		query = query.Where("project = ?", project)
	}

	// 如果指定了类型筛选条件
	if jobType != "" {
		query = query.Where("`type` = ?", jobType)
	}

	// 如果指定了状态筛选条件
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 如果指定了开始日期筛选条件（毫秒级时间戳）
	if startDate > 0 {
		query = query.Where("created_at >= ?", startDate)
	}

	// 如果指定了结束日期筛选条件（毫秒级时间戳）
	if endDate > 0 {
		query = query.Where("created_at <= ?", endDate)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&jobTasks).Error
	if err != nil {
		return nil, 0, err
	}

	return jobTasks, total, nil
}

// UpdateJobTask 更新任务
func (r *Repository) UpdateJobTask(ctx context.Context, jobTask *models.JobTask) error {
	// 更新时间戳，毫秒级精度
	jobTask.UpdatedAt = time.Now().UnixMilli()

	// 防止更新项目字段、类型、目标字段
	// 这些字段在创建后不允许修改，使用Select指定只更新允许的字段
	err := r.db.WithContext(ctx).Model(jobTask).Select(
		"updated_at",
		"status",
		"pass_accept_std",
		"execution_records",
		"active_execution_sequence",
	).Updates(jobTask).Error
	if err != nil {
		return err
	}

	// 清除项目列表缓存（更新可能修改项目字段）
	clearJobTaskProjectCache()
	return nil
}

// DeleteJobTask 删除任务（伪删除）
func (r *Repository) DeleteJobTask(ctx context.Context, id uint) error {
	timestamp := time.Now().UnixMilli()
	err := r.db.WithContext(ctx).Model(&models.JobTask{}).Where("id = ?", id).Update("deleted_at", timestamp).Error
	if err != nil {
		return err
	}

	// 清除项目列表缓存（删除可能影响项目列表）
	clearJobTaskProjectCache()
	return nil
}

// ListDeletedJobTasks 分页获取已删除任务列表（回收站）
func (r *Repository) ListDeletedJobTasks(ctx context.Context, page, pageSize int) ([]models.JobTask, int64, error) {
	var jobTasks []models.JobTask
	var total int64

	query := r.db.WithContext(ctx).Model(&models.JobTask{}).Where("deleted_at > ?", 0)

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，按删除时间倒序
	offset := (page - 1) * pageSize
	err := query.Order("deleted_at DESC").Offset(offset).Limit(pageSize).Find(&jobTasks).Error
	if err != nil {
		return nil, 0, err
	}

	return jobTasks, total, nil
}

// RestoreJobTask 恢复已删除的任务
func (r *Repository) RestoreJobTask(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Model(&models.JobTask{}).Where("id = ? AND deleted_at > ?", id, 0).Update("deleted_at", 0)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	// 清除项目列表缓存（恢复可能影响项目列表）
	clearJobTaskProjectCache()
	return nil
}

// PermanentDeleteJobTask 彻底删除任务（真删除）
func (r *Repository) PermanentDeleteJobTask(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Where("id = ? AND deleted_at > ?", id, 0).Delete(&models.JobTask{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	// 清除项目列表缓存（删除可能影响项目列表）
	clearJobTaskProjectCache()
	return nil
}

// GetAllJobTaskProjects 获取所有项目列表（去重）
// 从job_task表中查询所有项目字段，汇总并去重返回
// 使用缓存机制提升查询性能
func (r *Repository) GetAllJobTaskProjects(ctx context.Context) ([]string, error) {
	// 检查数据库连接是否初始化
	if r.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	// 先查缓存
	cacheKey := jobTaskProjectListCacheKey()
	if cached, ok := jobTaskProjectCache.Get(cacheKey); ok {
		if projects, ok := cached.([]string); ok {
			return projects, nil
		}
	}

	// 缓存未命中，查数据库
	var projects []string
	err := r.db.WithContext(ctx).
		Model(&models.JobTask{}).
		Select("DISTINCT project").
		Where("deleted_at = 0").
		Pluck("project", &projects).Error
	if err != nil {
		return nil, err
	}

	// 写入缓存
	jobTaskProjectCache.Set(cacheKey, projects, jobTaskProjectCacheTTL)
	return projects, nil
}

// GetJobTasksByIDs 根据ID列表批量获取任务
// 参数:
//   - ctx: 上下文
//   - ids: 任务ID列表
//
// 返回:
//   - []models.JobTask: 任务列表
//   - error: 错误信息
func (r *Repository) GetJobTasksByIDs(ctx context.Context, ids []uint) ([]models.JobTask, error) {
	var jobTasks []models.JobTask
	err := r.db.WithContext(ctx).
		Where("id IN ? AND deleted_at = ?", ids, 0).
		Find(&jobTasks).Error
	if err != nil {
		return nil, err
	}
	return jobTasks, nil
}

// GetAllJobTasks 获取所有未删除的任务
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - []models.JobTask: 任务列表
//   - error: 错误信息
func (r *Repository) GetAllJobTasks(ctx context.Context) ([]models.JobTask, error) {
	var jobTasks []models.JobTask
	err := r.db.WithContext(ctx).
		Where("deleted_at = ?", 0).
		Order("created_at DESC").
		Find(&jobTasks).Error
	if err != nil {
		return nil, err
	}
	return jobTasks, nil
}
