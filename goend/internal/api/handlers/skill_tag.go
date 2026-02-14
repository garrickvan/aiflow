package handlers

import (
	"aiflow/internal/api/helpers"
	"aiflow/internal/errors"
	"aiflow/internal/services"
	"context"
	"net/http"

	"github.com/go-chi/render"
)

// TagRequest 标签请求结构
type TagRequest struct {
	Name string `json:"name"`
}

// TagHandler 标签处理器
type TagHandler struct {
	service *services.TagService
}

// NewTagHandler 创建标签处理器
func NewTagHandler(service *services.TagService) *TagHandler {
	return &TagHandler{service: service}
}

// ListTags 获取所有标签（支持分页）
func (h *TagHandler) ListTags(w http.ResponseWriter, req *http.Request) {
	// 解析分页参数
	pagination := helpers.ParsePagination(req)

	// 调用service层
	result, err := h.service.ListTags(context.Background(), services.ListTagsRequest{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	})
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	helpers.RenderSuccess(w, req, result)
}

// CreateTag 创建标签
func (h *TagHandler) CreateTag(w http.ResponseWriter, req *http.Request) {
	var reqBody TagRequest
	if err := render.DecodeJSON(req.Body, &reqBody); err != nil {
		helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeBadRequest, "请求参数错误", err))
		return
	}

	// 调用service层
	result, err := h.service.CreateTag(context.Background(), services.CreateTagRequest{
		Name: reqBody.Name,
	})
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	helpers.RenderCreated(w, req, "标签创建成功", result)
}

// GetTag 根据ID获取标签
func (h *TagHandler) GetTag(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.ParseIDParam(req, "id")
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	result, err := h.service.GetTag(context.Background(), id)
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	helpers.RenderSuccess(w, req, result)
}

// UpdateTag 更新标签
func (h *TagHandler) UpdateTag(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.ParseIDParam(req, "id")
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	var reqBody TagRequest
	if err = render.DecodeJSON(req.Body, &reqBody); err != nil {
		helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeBadRequest, "请求参数错误", err))
		return
	}

	// 调用service层
	result, err := h.service.UpdateTag(context.Background(), services.UpdateTagRequest{
		ID:   id,
		Name: reqBody.Name,
	})
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	helpers.RenderSuccessWithMessage(w, req, "标签更新成功", result)
}

// DeleteTag 删除标签
func (h *TagHandler) DeleteTag(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.ParseIDParam(req, "id")
	if err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	if err := h.service.DeleteTag(context.Background(), id); err != nil {
		helpers.RenderError(w, req, err)
		return
	}

	helpers.RenderSuccessWithMessage(w, req, "标签删除成功", nil)
}
