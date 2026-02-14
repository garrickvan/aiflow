package services

import (
	"context"

	"aiflow/internal/errors"
	"aiflow/internal/models"
	"aiflow/internal/repositories"
)

// TagService 标签服务层
// 处理标签相关的业务逻辑，将业务逻辑从handler中分离
type TagService struct {
	repo *repositories.Repository
}

// TagResponse 标签响应结构
type TagResponse struct {
	ID        uint         `json:"id"`
	Name      string       `json:"name"`
	Skills    []models.Skill `json:"skills,omitempty"`
	CreatedAt int64        `json:"createdAt"`
	UpdatedAt int64        `json:"updatedAt"`
}

// NewTagService 创建标签服务实例
func NewTagService(repo *repositories.Repository) *TagService {
	return &TagService{repo: repo}
}

// ListTagsRequest 获取标签列表请求参数
type ListTagsRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// ListTagsResponse 获取标签列表响应
type ListTagsResponse struct {
	Items      []TagResponse          `json:"items"`
	Pagination map[string]interface{} `json:"pagination"`
}

// ListTags 获取标签列表（支持分页）
func (s *TagService) ListTags(ctx context.Context, req ListTagsRequest) (*ListTagsResponse, error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	tags, total, err := s.repo.ListTagsWithPagination(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, errors.NewInternalError(errors.ErrCodeInternalError, "获取标签列表失败", err)
	}

	// 转换响应格式
	items := make([]TagResponse, 0, len(tags))
	for _, tag := range tags {
		items = append(items, convertToTagResponse(&tag))
	}

	pagination := map[string]interface{}{
		"total":     total,
		"page":      req.Page,
		"pageSize":  req.PageSize,
		"totalPage": (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}

	return &ListTagsResponse{
		Items:      items,
		Pagination: pagination,
	}, nil
}

// CreateTagRequest 创建标签请求参数
type CreateTagRequest struct {
	Name string `json:"name"`
}

// CreateTag 创建标签
func (s *TagService) CreateTag(ctx context.Context, req CreateTagRequest) (*TagResponse, error) {
	// 检查标签名是否已存在
	existingTag, _ := s.repo.GetTagByName(ctx, req.Name)
	if existingTag != nil {
		return nil, errors.NewTagError(errors.ErrCodeTagCreate, "标签名已存在", nil)
	}

	tag := &models.Tag{
		Name: req.Name,
	}

	if err := s.repo.CreateTag(ctx, tag); err != nil {
		return nil, errors.NewTagError(errors.ErrCodeTagCreate, "创建标签失败", err)
	}

	response := convertToTagResponse(tag)
	return &response, nil
}

// GetTag 根据ID获取标签
func (s *TagService) GetTag(ctx context.Context, id uint) (*TagResponse, error) {
	tag, err := s.repo.GetTagByID(ctx, id)
	if err != nil {
		return nil, errors.NewTagError(errors.ErrCodeTagNotFound, "标签不存在", err)
	}

	response := convertToTagResponse(tag)
	return &response, nil
}

// UpdateTagRequest 更新标签请求参数
type UpdateTagRequest struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// UpdateTag 更新标签
func (s *TagService) UpdateTag(ctx context.Context, req UpdateTagRequest) (*TagResponse, error) {
	// 获取现有标签
	tag, err := s.repo.GetTagByID(ctx, req.ID)
	if err != nil {
		return nil, errors.NewTagError(errors.ErrCodeTagNotFound, "标签不存在", err)
	}

	// 检查新名称是否已被其他标签使用
	if req.Name != tag.Name {
		existingTag, _ := s.repo.GetTagByName(ctx, req.Name)
		if existingTag != nil && existingTag.ID != req.ID {
			return nil, errors.NewTagError(errors.ErrCodeTagUpdate, "标签名已存在", nil)
		}
	}

	// 更新标签信息
	tag.Name = req.Name

	if err := s.repo.UpdateTag(ctx, tag); err != nil {
		return nil, errors.NewTagError(errors.ErrCodeTagUpdate, "更新标签失败", err)
	}

	// 获取更新后的标签详情
	updatedTag, err := s.repo.GetTagByID(ctx, tag.ID)
	if err != nil {
		return nil, errors.NewTagError(errors.ErrCodeTagUpdate, "获取更新后的标签失败", err)
	}

	response := convertToTagResponse(updatedTag)
	return &response, nil
}

// DeleteTag 删除标签
func (s *TagService) DeleteTag(ctx context.Context, id uint) error {
	// 检查标签是否存在
	_, err := s.repo.GetTagByID(ctx, id)
	if err != nil {
		return errors.NewTagError(errors.ErrCodeTagNotFound, "标签不存在", err)
	}

	if err := s.repo.DeleteTag(ctx, id); err != nil {
		return errors.NewTagError(errors.ErrCodeTagDelete, "删除标签失败", err)
	}

	return nil
}

// convertToTagResponse 将模型转换为响应结构
func convertToTagResponse(tag *models.Tag) TagResponse {
	return TagResponse{
		ID:        tag.ID,
		Name:      tag.Name,
		Skills:    tag.Skills,
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	}
}
