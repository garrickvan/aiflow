package handlers

import (
	"aiflow/internal/api/helpers"
	"aiflow/internal/errors"
	"aiflow/internal/services"
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"gorm.io/gorm"
)

// SkillRequest 技能请求结构
type SkillRequest struct {
	Name          string `json:"name"`
	ResourceDir   string `json:"resourceDir"`
	Description   string `json:"description"`
	Version       string `json:"version"`
	Detail        string `json:"detail"`
	License       string `json:"license"`
	Compatibility string `json:"compatibility"`
	Metadata      string `json:"metadata"`
	AllowedTools  string `json:"allowedTools"`
	Tags          []uint `json:"tags"`
}

// SkillHandler 技能处理器
type SkillHandler struct {
	service *services.SkillService
}

// NewSkillHandler 创建技能处理器
func NewSkillHandler(service *services.SkillService) *SkillHandler {
	return &SkillHandler{service: service}
}

// ListSkills 获取所有技能（支持分页、标签筛选和日期范围筛选）
func (h *SkillHandler) ListSkills(w http.ResponseWriter, req *http.Request) {
	// 获取标签筛选参数
	tagIDStr := req.URL.Query().Get("tagId")
	// 解析分页参数
	pagination := helpers.ParsePagination(req)

	// 解析日期范围参数
	startDate := helpers.ParseIntParam(req, "startDate", 0)
	endDate := helpers.ParseIntParam(req, "endDate", 0)

	// 解析标签ID
	var tagID uint
	if tagIDStr != "" {
		id, err := services.ParseUint(tagIDStr)
		if err != nil {
			helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeInvalidIDParam, "无效的标签ID参数", err))
			return
		}
		tagID = id
	}

	// 调用service层
	result, err := h.service.ListSkills(context.Background(), services.ListSkillsRequest{
		TagID:     tagID,
		Page:      pagination.Page,
		PageSize:  pagination.PageSize,
		StartDate: startDate,
		EndDate:   endDate,
	})

	if err != nil {
		helpers.RenderError(w, req, errors.NewSkillError(errors.ErrCodeInternalError, "获取技能失败", err))
		return
	}

	helpers.RenderSuccess(w, req, result)
}

// CreateSkill 创建技能
func (h *SkillHandler) CreateSkill(w http.ResponseWriter, req *http.Request) {
	var reqBody SkillRequest
	if err := render.DecodeJSON(req.Body, &reqBody); err != nil {
		helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeBadRequest, "请求参数错误", err))
		return
	}

	result, err := h.service.CreateSkill(context.Background(), services.CreateSkillRequest{
		Name:          reqBody.Name,
		ResourceDir:   reqBody.ResourceDir,
		Description:   reqBody.Description,
		Version:       reqBody.Version,
		Detail:        reqBody.Detail,
		License:       reqBody.License,
		Compatibility: reqBody.Compatibility,
		Metadata:      reqBody.Metadata,
		AllowedTools:  reqBody.AllowedTools,
		Tags:          reqBody.Tags,
	})

	if err != nil {
		helpers.RenderError(w, req, errors.NewSkillError(errors.ErrCodeSkillCreate, "创建技能失败", err))
		return
	}

	helpers.RenderCreated(w, req, "技能创建成功", result)
}

// GetSkill 根据ID获取技能
func (h *SkillHandler) GetSkill(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.ParseIDParam(req, "id")
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	result, err := h.service.GetSkill(context.Background(), id)
	if err != nil {
		helpers.RenderError(w, req, errors.NewNotFoundError(errors.ErrCodeSkillNotFound, "技能不存在", err))
		return
	}

	helpers.RenderSuccess(w, req, result)
}

// UpdateSkill 更新技能
func (h *SkillHandler) UpdateSkill(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.ParseIDParam(req, "id")
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	var reqBody SkillRequest
	if err = render.DecodeJSON(req.Body, &reqBody); err != nil {
		helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeBadRequest, "请求参数错误", err))
		return
	}

	result, err := h.service.UpdateSkill(context.Background(), services.UpdateSkillRequest{
		ID:            id,
		Name:          reqBody.Name,
		ResourceDir:   reqBody.ResourceDir,
		Description:   reqBody.Description,
		Version:       reqBody.Version,
		Detail:        reqBody.Detail,
		License:       reqBody.License,
		Compatibility: reqBody.Compatibility,
		Metadata:      reqBody.Metadata,
		AllowedTools:  reqBody.AllowedTools,
		Tags:          reqBody.Tags,
	})

	if err != nil {
		helpers.RenderError(w, req, errors.NewSkillError(errors.ErrCodeSkillUpdate, "更新技能失败", err))
		return
	}

	helpers.RenderSuccessWithMessage(w, req, "技能更新成功", result)
}

// DeleteSkill 删除技能（伪删除，进入回收站）
func (h *SkillHandler) DeleteSkill(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.ParseIDParam(req, "id")
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	if err := h.service.DeleteSkill(context.Background(), id); err != nil {
		helpers.RenderError(w, req, errors.NewSkillError(errors.ErrCodeSkillDelete, "删除技能失败", err))
		return
	}

	helpers.RenderSuccessWithMessage(w, req, "技能已移至回收站", nil)
}

// ListDeletedSkills 获取回收站技能列表
func (h *SkillHandler) ListDeletedSkills(w http.ResponseWriter, req *http.Request) {
	pagination := helpers.ParsePagination(req)

	result, err := h.service.ListDeletedSkills(context.Background(), pagination.Page, pagination.PageSize)
	if err != nil {
		helpers.RenderError(w, req, errors.NewSkillError(errors.ErrCodeInternalError, "获取回收站列表失败", err))
		return
	}

	helpers.RenderSuccess(w, req, result)
}

// RestoreSkill 恢复回收站中的技能
func (h *SkillHandler) RestoreSkill(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.ParseIDParam(req, "id")
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	if err := h.service.RestoreSkill(context.Background(), id); err != nil {
		if err == gorm.ErrRecordNotFound {
			helpers.RenderError(w, req, errors.NewNotFoundError(errors.ErrCodeSkillNotFound, "技能不存在或不在回收站中", err))
			return
		}
		helpers.RenderError(w, req, errors.NewSkillError(errors.ErrCodeSkillTrash, "恢复技能失败", err))
		return
	}

	helpers.RenderSuccessWithMessage(w, req, "技能恢复成功", nil)
}

// PermanentDeleteSkill 彻底删除技能（从回收站中永久删除）
func (h *SkillHandler) PermanentDeleteSkill(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.ParseIDParam(req, "id")
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	if err := h.service.PermanentDeleteSkill(context.Background(), id); err != nil {
		if err == gorm.ErrRecordNotFound {
			helpers.RenderError(w, req, errors.NewNotFoundError(errors.ErrCodeSkillNotFound, "技能不存在或不在回收站中", err))
			return
		}
		helpers.RenderError(w, req, errors.NewSkillError(errors.ErrCodeSkillDelete, "彻底删除技能失败", err))
		return
	}

	helpers.RenderSuccessWithMessage(w, req, "技能已彻底删除", nil)
}

// ExportSkills 导出技能为MD格式
func (h *SkillHandler) ExportSkills(w http.ResponseWriter, req *http.Request) {
	idStr := req.URL.Query().Get("id")

	if idStr == "" {
		helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeBadRequestParam, "必须提供技能ID参数", nil))
		return
	}

	id, err := services.ParseUint(idStr)
	if err != nil {
		helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeInvalidIDParam, "无效的ID参数", err))
		return
	}

	content, filename, err := h.service.ExportSkill(context.Background(), id)
	if err != nil {
		helpers.RenderError(w, req, errors.NewNotFoundError(errors.ErrCodeSkillNotFound, "技能不存在", err))
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/markdown")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))

	w.Write([]byte(content))
}
