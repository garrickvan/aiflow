package services

import (
	"aiflow/internal/models"
	"aiflow/internal/repositories"
	"context"
	"strconv"
	"strings"
	"time"
)

// SkillService 技能服务层
// 处理技能相关的业务逻辑，将业务逻辑从handler中分离
type SkillService struct {
	repo *repositories.Repository
}

// SkillResponse 技能响应结构
type SkillResponse struct {
	ID            uint         `json:"id"`
	Name          string       `json:"name"`
	ResourceDir   string       `json:"resourceDir"`
	Description   string       `json:"description"`
	Version       string       `json:"version"`
	Detail        string       `json:"detail"`
	License       string       `json:"license"`
	Compatibility string       `json:"compatibility"`
	Metadata      string       `json:"metadata"`
	AllowedTools  string       `json:"allowedTools"`
	Tags          []models.Tag `json:"tags"`
	CreatedAt     int64        `json:"createdAt"`
	UpdatedAt     int64        `json:"updatedAt"`
}

// NewSkillService 创建技能服务实例
func NewSkillService(repo *repositories.Repository) *SkillService {
	return &SkillService{repo: repo}
}

// ListSkillsRequest 获取技能列表请求参数
type ListSkillsRequest struct {
	TagID     uint
	Page      int
	PageSize  int
	StartDate int64
	EndDate   int64
}

// ListSkillsResponse 获取技能列表响应
type ListSkillsResponse struct {
	Items      []SkillResponse        `json:"items"`
	Pagination map[string]interface{} `json:"pagination"`
}

// ListSkills 获取技能列表（支持分页、标签筛选和日期范围筛选）
func (s *SkillService) ListSkills(ctx context.Context, req ListSkillsRequest) (*ListSkillsResponse, error) {
	offset := (req.Page - 1) * req.PageSize

	var skills []models.Skill
	var total int64

	// 构建基础查询
	baseQuery := s.repo.GetDB().WithContext(ctx).Model(&models.Skill{}).Where("deleted_at = ?", 0)

	// 添加日期范围筛选条件
	if req.StartDate > 0 {
		baseQuery = baseQuery.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate > 0 {
		baseQuery = baseQuery.Where("created_at <= ?", req.EndDate)
	}

	if req.TagID > 0 {
		// 按标签筛选 - 先计算总数
		countQuery := s.repo.GetDB().WithContext(ctx).Model(&models.Skill{}).
			Joins("JOIN skill_tags ON skill_tags.skill_id = skills.id").
			Where("skills.deleted_at = ?", 0).
			Where("skill_tags.tag_id = ?", req.TagID)
		countQuery.Count(&total)

		// 再获取数据列表
		err := s.repo.GetDB().WithContext(ctx).
			Joins("JOIN skill_tags ON skill_tags.skill_id = skills.id").
			Where("skills.deleted_at = ?", 0).
			Where("skill_tags.tag_id = ?", req.TagID).
			Offset(offset).Limit(req.PageSize).
			Preload("Tags").
			Find(&skills).Error
		if err != nil {
			return nil, err
		}
	} else {
		// 计算总数
		baseQuery.Count(&total)
		// 获取所有技能
		err := baseQuery.
			Offset(offset).Limit(req.PageSize).
			Preload("Tags").
			Find(&skills).Error
		if err != nil {
			return nil, err
		}
	}

	// 转换响应格式
	responseSkills := make([]SkillResponse, 0, len(skills))
	for _, skill := range skills {
		responseSkills = append(responseSkills, convertToSkillResponse(&skill))
	}

	pagination := map[string]interface{}{
		"total":     total,
		"page":      req.Page,
		"pageSize":  req.PageSize,
		"totalPage": (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}

	return &ListSkillsResponse{
		Items:      responseSkills,
		Pagination: pagination,
	}, nil
}

// CreateSkillRequest 创建技能请求参数
type CreateSkillRequest struct {
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

// CreateSkill 创建技能
func (s *SkillService) CreateSkill(ctx context.Context, req CreateSkillRequest) (*SkillResponse, error) {
	timestamp := time.Now().UnixMilli()
	skill := &models.Skill{
		Name:          req.Name,
		ResourceDir:   req.ResourceDir,
		Description:   req.Description,
		Version:       req.Version,
		Detail:        req.Detail,
		License:       req.License,
		Compatibility: req.Compatibility,
		Metadata:      req.Metadata,
		AllowedTools:  req.AllowedTools,
		CreatedAt:     timestamp,
		UpdatedAt:     timestamp,
	}

	if err := s.repo.CreateSkill(ctx, skill); err != nil {
		return nil, err
	}

	// 处理标签关联关系
	for _, tagID := range req.Tags {
		if err := s.repo.AddTagToSkill(ctx, skill.ID, tagID); err != nil {
			continue
		}
	}

	// 获取技能详情
	createdSkill, err := s.repo.GetSkillByID(ctx, skill.ID)
	if err != nil {
		return nil, err
	}

	response := convertToSkillResponse(createdSkill)
	return &response, nil
}

// GetSkill 根据ID获取技能
func (s *SkillService) GetSkill(ctx context.Context, id uint) (*SkillResponse, error) {
	skill, err := s.repo.GetSkillByID(ctx, id)
	if err != nil {
		return nil, err
	}
	response := convertToSkillResponse(skill)
	return &response, nil
}

// UpdateSkillRequest 更新技能请求参数
type UpdateSkillRequest struct {
	ID            uint
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

// UpdateSkill 更新技能
func (s *SkillService) UpdateSkill(ctx context.Context, req UpdateSkillRequest) (*SkillResponse, error) {
	skill, err := s.repo.GetSkillByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 更新技能信息
	skill.Name = req.Name
	skill.ResourceDir = req.ResourceDir
	skill.Description = req.Description
	skill.Version = req.Version
	skill.Detail = req.Detail
	skill.License = req.License
	skill.Compatibility = req.Compatibility
	skill.Metadata = req.Metadata
	skill.AllowedTools = req.AllowedTools
	skill.UpdatedAt = time.Now().UnixMilli()

	if err := s.repo.UpdateSkill(ctx, skill); err != nil {
		return nil, err
	}

	// 先删除所有旧的标签关联
	s.repo.GetDB().WithContext(ctx).Where("skill_id = ?", skill.ID).Delete(&models.SkillTag{})

	// 再添加新的标签关联
	for _, tagID := range req.Tags {
		if err := s.repo.AddTagToSkill(ctx, skill.ID, tagID); err != nil {
			continue
		}
	}

	// 获取更新后的技能详情
	updatedSkill, err := s.repo.GetSkillByID(ctx, skill.ID)
	if err != nil {
		return nil, err
	}

	response := convertToSkillResponse(updatedSkill)
	return &response, nil
}

// DeleteSkill 删除技能（伪删除）
func (s *SkillService) DeleteSkill(ctx context.Context, id uint) error {
	return s.repo.DeleteSkill(ctx, id)
}

// ListDeletedSkills 获取回收站技能列表
func (s *SkillService) ListDeletedSkills(ctx context.Context, page, pageSize int) (*ListSkillsResponse, error) {
	offset := (page - 1) * pageSize

	var skills []models.Skill
	var total int64

	baseQuery := s.repo.GetDB().WithContext(ctx).Model(&models.Skill{}).Where("deleted_at > ?", 0)

	baseQuery.Count(&total)
	err := baseQuery.
		Offset(offset).Limit(pageSize).
		Preload("Tags").
		Find(&skills).Error

	if err != nil {
		return nil, err
	}

	responseSkills := make([]SkillResponse, 0, len(skills))
	for _, skill := range skills {
		responseSkills = append(responseSkills, convertToSkillResponse(&skill))
	}

	pagination := map[string]interface{}{
		"total":     total,
		"page":      page,
		"pageSize":  pageSize,
		"totalPage": (total + int64(pageSize) - 1) / int64(pageSize),
	}

	return &ListSkillsResponse{
		Items:      responseSkills,
		Pagination: pagination,
	}, nil
}

// RestoreSkill 恢复回收站中的技能
func (s *SkillService) RestoreSkill(ctx context.Context, id uint) error {
	return s.repo.RestoreSkill(ctx, id)
}

// PermanentDeleteSkill 彻底删除技能
func (s *SkillService) PermanentDeleteSkill(ctx context.Context, id uint) error {
	// 先删除标签关联关系
	s.repo.GetDB().WithContext(ctx).Where("skill_id = ?", id).Delete(&models.SkillTag{})
	return s.repo.PermanentDeleteSkill(ctx, id)
}

// ExportSkill 导出技能为MD格式
func (s *SkillService) ExportSkill(ctx context.Context, id uint) (string, string, error) {
	skill, err := s.repo.GetSkillByID(ctx, id)
	if err != nil {
		return "", "", err
	}

	// 获取标签信息
	tags, _ := s.repo.GetTagsBySkillID(ctx, skill.ID)
	var tagNames []string
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}

	// 生成MD格式内容
	var mdContent strings.Builder

	mdContent.WriteString("---\n")
	mdContent.WriteString("name: " + skill.Name + "\n")
	mdContent.WriteString("resource_dir: " + skill.ResourceDir + "\n")
	mdContent.WriteString("description: " + skill.Description + "\n")

	if len(tagNames) > 0 {
		mdContent.WriteString("tags:\n")
		for _, tagName := range tagNames {
			mdContent.WriteString("  - " + tagName + "\n")
		}
	}

	if skill.License != "" {
		mdContent.WriteString("license: " + skill.License + "\n")
	}

	if skill.Compatibility != "" {
		mdContent.WriteString("compatibility: " + skill.Compatibility + "\n")
	}

	if skill.Metadata != "" {
		mdContent.WriteString("metadata: " + skill.Metadata + "\n")
	}

	if skill.AllowedTools != "" {
		mdContent.WriteString("allowed_tools: " + skill.AllowedTools + "\n")
	}

	mdContent.WriteString("\n---\n\n")
	if skill.Detail != "" {
		mdContent.WriteString(skill.Detail + "\n")
	} else {
		mdContent.WriteString("暂无详细说明\n")
	}

	filename := "SKILL.md"
	return mdContent.String(), filename, nil
}

// convertToSkillResponse 将模型转换为响应结构
func convertToSkillResponse(skill *models.Skill) SkillResponse {
	return SkillResponse{
		ID:            skill.ID,
		Name:          skill.Name,
		ResourceDir:   skill.ResourceDir,
		Description:   skill.Description,
		Version:       skill.Version,
		Detail:        skill.Detail,
		License:       skill.License,
		Compatibility: skill.Compatibility,
		Metadata:      skill.Metadata,
		AllowedTools:  skill.AllowedTools,
		Tags:          skill.Tags,
		CreatedAt:     skill.CreatedAt,
		UpdatedAt:     skill.UpdatedAt,
	}
}

// ParseUint 将字符串解析为uint
func ParseUint(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
