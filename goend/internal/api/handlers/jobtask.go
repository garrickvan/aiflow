package handlers

import (
	"aiflow/internal/api/helpers"
	"aiflow/internal/errors"
	"aiflow/internal/models"
	"aiflow/internal/services"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/render"
	"gorm.io/gorm"
)

// JobTaskRequest 任务请求结构
type JobTaskRequest struct {
	JobNo         string `json:"jobNo"`
	Project       string `json:"project"`
	Type          string `json:"type"`
	Goal          string `json:"goal"`
	PassAcceptStd bool   `json:"passAcceptStd"`
	Status        string `json:"status"`
}

// JobTaskHandler 任务处理器
type JobTaskHandler struct {
	service *services.JobTaskService
}

// NewJobTaskHandler 创建任务处理器
func NewJobTaskHandler(service *services.JobTaskService) *JobTaskHandler {
	return &JobTaskHandler{service: service}
}

// ListJobTasks 获取任务列表（支持分页、多条件筛选和日期范围筛选）
func (h *JobTaskHandler) ListJobTasks(w http.ResponseWriter, req *http.Request) {
	// 获取筛选参数
	project := req.URL.Query().Get("project")
	jobType := req.URL.Query().Get("type")
	status := req.URL.Query().Get("status")
	// 解析日期范围参数（毫秒级时间戳）
	startDate := helpers.ParseIntParam(req, "startDate", 0)
	endDate := helpers.ParseIntParam(req, "endDate", 0)
	// 解析分页参数
	pagination := helpers.ParsePagination(req)

	// 调用service层
	result, err := h.service.ListJobTasks(context.Background(), services.ListJobTasksRequest{
		Page:      pagination.Page,
		PageSize:  pagination.PageSize,
		Project:   project,
		Type:      jobType,
		Status:    status,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		helpers.RenderError(w, req, errors.NewTaskError(errors.ErrCodeInternalError, "获取任务列表失败", err))
		return
	}

	helpers.RenderSuccess(w, req, result)
}

// CreateJobTask 创建任务
func (h *JobTaskHandler) CreateJobTask(w http.ResponseWriter, req *http.Request) {
	var reqBody JobTaskRequest
	if err := render.DecodeJSON(req.Body, &reqBody); err != nil {
		helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeBadRequest, "请求参数错误", err))
		return
	}

	// 调用service层
	result, err := h.service.CreateJobTask(context.Background(), services.CreateJobTaskRequest{
		JobNo:         reqBody.JobNo,
		Project:       reqBody.Project,
		Type:          reqBody.Type,
		Goal:          reqBody.Goal,
		PassAcceptStd: reqBody.PassAcceptStd,
	})
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	helpers.RenderCreated(w, req, "任务创建成功", result)
}

// GetJobTask 根据ID获取任务
func (h *JobTaskHandler) GetJobTask(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.ParseIDParam(req, "id")
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	result, err := h.service.GetJobTask(context.Background(), id)
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	helpers.RenderSuccess(w, req, result)
}

// UpdateJobTask 更新任务
func (h *JobTaskHandler) UpdateJobTask(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.ParseIDParam(req, "id")
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	var reqBody JobTaskRequest
	if err = render.DecodeJSON(req.Body, &reqBody); err != nil {
		helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeBadRequest, "请求参数错误", err))
		return
	}

	// 调用service层
	result, err := h.service.UpdateJobTask(context.Background(), services.UpdateJobTaskRequest{
		ID:            id,
		Status:        reqBody.Status,
		PassAcceptStd: reqBody.PassAcceptStd,
	})
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	helpers.RenderSuccessWithMessage(w, req, "任务更新成功", result)
}

// DeleteJobTask 删除任务（伪删除，进入回收站）
func (h *JobTaskHandler) DeleteJobTask(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.ParseIDParam(req, "id")
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	if err := h.service.DeleteJobTask(context.Background(), id); err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	helpers.RenderSuccessWithMessage(w, req, "任务已移至回收站", nil)
}

// ListDeletedJobTasks 获取回收站任务列表
func (h *JobTaskHandler) ListDeletedJobTasks(w http.ResponseWriter, req *http.Request) {
	// 解析分页参数
	pagination := helpers.ParsePagination(req)

	// 调用service层
	result, err := h.service.ListDeletedJobTasks(context.Background(), pagination.Page, pagination.PageSize)
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	helpers.RenderSuccess(w, req, result)
}

// RestoreJobTask 恢复回收站中的任务
func (h *JobTaskHandler) RestoreJobTask(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.ParseIDParam(req, "id")
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	if err := h.service.RestoreJobTask(context.Background(), id); err != nil {
		if err == gorm.ErrRecordNotFound {
			helpers.RenderError(w, req, errors.NewNotFoundError(errors.ErrCodeTaskNotFound, "任务不存在或不在回收站中", err))
			return
		}
		helpers.RenderError(w, req, err)
		return
	}

	helpers.RenderSuccessWithMessage(w, req, "任务恢复成功", nil)
}

// PermanentDeleteJobTask 彻底删除任务（从回收站中永久删除）
func (h *JobTaskHandler) PermanentDeleteJobTask(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.ParseIDParam(req, "id")
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	if err := h.service.PermanentDeleteJobTask(context.Background(), id); err != nil {
		if err == gorm.ErrRecordNotFound {
			helpers.RenderError(w, req, errors.NewNotFoundError(errors.ErrCodeTaskNotFound, "任务不存在或不在回收站中", err))
			return
		}
		helpers.RenderError(w, req, err)
		return
	}

	helpers.RenderSuccessWithMessage(w, req, "任务已彻底删除", nil)
}

// GetAllJobTaskProjects 获取所有项目列表（去重）
func (h *JobTaskHandler) GetAllJobTaskProjects(w http.ResponseWriter, req *http.Request) {
	projects, err := h.service.GetAllProjects(context.Background())
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	helpers.RenderSuccess(w, req, projects)
}

// BatchExportJobTasksRequest 批量导出任务请求结构
type BatchExportJobTasksRequest struct {
	IDs    []uint `json:"ids"`    // 指定导出的任务ID列表，为空则导出所有
	Format string `json:"format"` // 导出格式：csv、json、md
}

// BatchExportJobTasks 批量导出任务
// 支持按ID列表导出或导出全部，支持CSV/JSON/MD格式
func (h *JobTaskHandler) BatchExportJobTasks(w http.ResponseWriter, req *http.Request) {
	var reqBody BatchExportJobTasksRequest
	if err := render.DecodeJSON(req.Body, &reqBody); err != nil {
		helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeBadRequest, "请求参数错误", err))
		return
	}

	// 默认格式为CSV
	if reqBody.Format == "" {
		reqBody.Format = "csv"
	}

	// 验证格式
	if reqBody.Format != "csv" && reqBody.Format != "json" && reqBody.Format != "md" {
		helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeBadRequestParam, "不支持的导出格式，仅支持 csv、json、md", nil))
		return
	}

	// 获取任务数据
	jobTasks, err := h.service.GetJobTasksForExport(context.Background(), reqBody.IDs)
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	// 根据格式导出
	switch reqBody.Format {
	case "csv":
		exportJobTasksAsCSV(w, req, jobTasks)
	case "json":
		exportJobTasksAsJSON(w, req, jobTasks)
	case "md":
		exportJobTasksAsMD(w, req, jobTasks)
	}
}

// exportJobTasksAsCSV 导出任务为CSV格式
func exportJobTasksAsCSV(w http.ResponseWriter, req *http.Request, jobTasks []models.JobTask) {
	// 设置响应头
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=jobtasks.csv")
	w.Header().Set("Content-Transfer-Encoding", "binary")

	// 写入UTF-8 BOM，确保Excel正确识别中文
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// 写入表头
	headers := []string{"任务编号", "所属项目", "任务类型", "任务目标", "验收状态", "完成阶段", "使用技能", "创建时间", "更新时间"}
	writer.Write(headers)

	// 写入数据
	for _, task := range jobTasks {
		passAcceptStd := "未通过"
		if task.PassAcceptStd {
			passAcceptStd = "已通过"
		}

		// 解析执行记录获取技能列表
		var skillsStr string
		if task.ExecutionRecords != "" {
			var records []models.ExecutionRecord
			if err := json.Unmarshal([]byte(task.ExecutionRecords), &records); err == nil && len(records) > 0 {
				// 获取最新执行记录的技能
				latestRecord := records[len(records)-1]
				skillsStr = strings.Join(latestRecord.Skills, ", ")
			}
		}

		record := []string{
			task.JobNo,
			task.Project,
			task.Type,
			task.Goal,
			passAcceptStd,
			task.Status,
			skillsStr,
			formatTimestamp(task.CreatedAt),
			formatTimestamp(task.UpdatedAt),
		}
		writer.Write(record)
	}
}

// exportJobTasksAsJSON 导出任务为JSON格式
func exportJobTasksAsJSON(w http.ResponseWriter, req *http.Request, jobTasks []models.JobTask) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=jobtasks.json")

	// 构造导出数据结构
	type ExportJobTask struct {
		JobNo         string   `json:"jobNo"`
		Project       string   `json:"project"`
		Type          string   `json:"type"`
		Goal          string   `json:"goal"`
		PassAcceptStd bool     `json:"passAcceptStd"`
		Status        string   `json:"status"`
		Skills        []string `json:"skills"`
		CreatedAt     string   `json:"createdAt"`
		UpdatedAt     string   `json:"updatedAt"`
	}

	exportData := make([]ExportJobTask, 0, len(jobTasks))
	for _, task := range jobTasks {
		// 解析执行记录获取技能列表
		var skills []string
		if task.ExecutionRecords != "" {
			var records []models.ExecutionRecord
			if err := json.Unmarshal([]byte(task.ExecutionRecords), &records); err == nil && len(records) > 0 {
				// 获取最新执行记录的技能
				skills = records[len(records)-1].Skills
			}
		}

		exportData = append(exportData, ExportJobTask{
			JobNo:         task.JobNo,
			Project:       task.Project,
			Type:          task.Type,
			Goal:          task.Goal,
			PassAcceptStd: task.PassAcceptStd,
			Status:        task.Status,
			Skills:        skills,
			CreatedAt:     formatTimestamp(task.CreatedAt),
			UpdatedAt:     formatTimestamp(task.UpdatedAt),
		})
	}

	jsonData, _ := json.MarshalIndent(exportData, "", "  ")
	w.Write(jsonData)
}

// exportJobTasksAsMD 导出任务为Markdown格式
func exportJobTasksAsMD(w http.ResponseWriter, req *http.Request, jobTasks []models.JobTask) {
	// 设置响应头
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=jobtasks.md")

	var mdContent strings.Builder

	mdContent.WriteString("# 任务导出列表\n\n")
	mdContent.WriteString(fmt.Sprintf("导出时间: %s\n\n", formatTimestamp(time.Now().UnixMilli())))
	mdContent.WriteString(fmt.Sprintf("任务数量: %d\n\n", len(jobTasks)))
	mdContent.WriteString("---\n\n")

	for i, task := range jobTasks {
		mdContent.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, task.JobNo))
		mdContent.WriteString(fmt.Sprintf("- **所属项目**: %s\n", task.Project))
		mdContent.WriteString(fmt.Sprintf("- **任务类型**: %s\n", task.Type))
		mdContent.WriteString(fmt.Sprintf("- **任务目标**: %s\n", task.Goal))

		passAcceptStd := "未通过"
		if task.PassAcceptStd {
			passAcceptStd = "已通过"
		}
		mdContent.WriteString(fmt.Sprintf("- **验收状态**: %s\n", passAcceptStd))
		mdContent.WriteString(fmt.Sprintf("- **完成阶段**: %s\n", task.Status))

		// 解析执行记录获取技能列表（收集所有记录中的技能并去重）
		if task.ExecutionRecords != "" {
			var records []models.ExecutionRecord
			if err := json.Unmarshal([]byte(task.ExecutionRecords), &records); err == nil && len(records) > 0 {
				skillSet := make(map[string]struct{})
				for _, record := range records {
					for _, skill := range record.Skills {
						skillSet[skill] = struct{}{}
					}
				}
				if len(skillSet) > 0 {
					skills := make([]string, 0, len(skillSet))
					for skill := range skillSet {
						skills = append(skills, skill)
					}
					mdContent.WriteString(fmt.Sprintf("- **使用技能**: %s\n", strings.Join(skills, ", ")))
				}
			}
		}

		mdContent.WriteString(fmt.Sprintf("- **创建时间**: %s\n", formatTimestamp(task.CreatedAt)))
		mdContent.WriteString(fmt.Sprintf("- **更新时间**: %s\n", formatTimestamp(task.UpdatedAt)))

		// 解析执行记录
		if task.ExecutionRecords != "" {
			mdContent.WriteString("- **执行记录**:\n")
			var records []models.ExecutionRecord
			if err := json.Unmarshal([]byte(task.ExecutionRecords), &records); err == nil {
				for _, record := range records {
					mdContent.WriteString(fmt.Sprintf("  - 序号 %d: %s (%s)\n", record.Sequence, record.Status, record.Result))
				}
			}
		}

		mdContent.WriteString("\n---\n\n")
	}

	w.Write([]byte(mdContent.String()))
}

// formatTimestamp 格式化时间戳为可读字符串
func formatTimestamp(timestamp int64) string {
	t := time.UnixMilli(timestamp)
	return t.Format("2006-01-02 15:04:05")
}
