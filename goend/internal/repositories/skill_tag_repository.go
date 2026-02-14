package repositories

import (
	"context"
	"fmt"
	"time"

	"aiflow/internal/cache"
	"aiflow/internal/models"
)

// 缓存相关常量定义
const (
	// tagCacheTTL 标签缓存过期时间
	tagCacheTTL = 5 * time.Minute
	// tagCacheMaxSize 标签缓存最大条目数
	tagCacheMaxSize = 1000
)

// tagCacheKey 生成标签缓存Key
func tagCacheKey(id uint) string {
	return fmt.Sprintf("tag:%d", id)
}

// tagListCacheKey 生成标签列表缓存Key
func tagListCacheKey(page, pageSize int) string {
	return fmt.Sprintf("tag:list:%d:%d", page, pageSize)
}

// tagTotalCacheKey 生成标签总数缓存Key
func tagTotalCacheKey() string {
	return "tag:total"
}

// 标签缓存实例
var tagCache = cache.NewLocalCache(tagCacheMaxSize)

// clearTagCache 清除所有标签相关缓存
func clearTagCache() {
	tagCache.DeleteByPrefix("tag:")
}

// Tag CRUD 操作

// CreateTag 创建标签
func (r *Repository) CreateTag(ctx context.Context, tag *models.Tag) error {
	// 设置时间戳，毫秒级精度
	timestamp := time.Now().UnixMilli()
	tag.CreatedAt = timestamp
	tag.UpdatedAt = timestamp

	err := r.db.WithContext(ctx).Create(tag).Error
	if err != nil {
		return err
	}

	// 清除标签列表缓存
	clearTagCache()
	return nil
}

// GetTagByID 根据ID获取标签
func (r *Repository) GetTagByID(ctx context.Context, id uint) (*models.Tag, error) {
	// 先查缓存
	cacheKey := tagCacheKey(id)
	if cached, ok := tagCache.Get(cacheKey); ok {
		if tag, ok := cached.(*models.Tag); ok {
			return tag, nil
		}
	}

	// 缓存未命中，查数据库
	var tag models.Tag
	err := r.db.WithContext(ctx).Preload("Skills").First(&tag, id).Error
	if err != nil {
		return nil, err
	}

	// 写入缓存
	tagCache.Set(cacheKey, &tag, tagCacheTTL)
	return &tag, nil
}

// GetTagByName 根据名称获取标签
func (r *Repository) GetTagByName(ctx context.Context, name string) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.WithContext(ctx).Where("name = ?", name).Preload("Skills").First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// ListTags 获取所有标签
func (r *Repository) ListTags(ctx context.Context) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.db.WithContext(ctx).Find(&tags).Error
	return tags, err
}

// ListTagsWithPagination 分页获取标签列表
// 返回值: 标签列表, 总数, 错误
func (r *Repository) ListTagsWithPagination(ctx context.Context, page, pageSize int) ([]models.Tag, int64, error) {
	// 尝试从缓存获取
	listCacheKey := tagListCacheKey(page, pageSize)
	totalCacheKey := tagTotalCacheKey()

	// 检查列表缓存
	if cachedList, ok := tagCache.Get(listCacheKey); ok {
		if cachedTotal, ok := tagCache.Get(totalCacheKey); ok {
			tags, listOk := cachedList.([]models.Tag)
			total, totalOk := cachedTotal.(int64)
			if listOk && totalOk {
				return tags, total, nil
			}
		}
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询总数
	var total int64
	err := r.db.WithContext(ctx).Model(&models.Tag{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 查询标签列表
	var tags []models.Tag
	err = r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Find(&tags).Error
	if err != nil {
		return nil, 0, err
	}

	// 写入缓存
	tagCache.Set(listCacheKey, tags, tagCacheTTL)
	tagCache.Set(totalCacheKey, total, tagCacheTTL)

	return tags, total, nil
}

// UpdateTag 更新标签
func (r *Repository) UpdateTag(ctx context.Context, tag *models.Tag) error {
	// 更新时间戳，毫秒级精度
	tag.UpdatedAt = time.Now().UnixMilli()

	err := r.db.WithContext(ctx).Save(tag).Error
	if err != nil {
		return err
	}

	// 清除相关缓存
	tagCache.Delete(tagCacheKey(tag.ID))
	clearTagCache()
	return nil
}

// DeleteTag 删除标签
func (r *Repository) DeleteTag(ctx context.Context, id uint) error {
	// 先删除关联关系
	if err := r.db.WithContext(ctx).Where("tag_id = ?", id).Delete(&models.SkillTag{}).Error; err != nil {
		return err
	}

	// 再删除标签
	if err := r.db.WithContext(ctx).Delete(&models.Tag{}, id).Error; err != nil {
		return err
	}

	// 清除相关缓存
	tagCache.Delete(tagCacheKey(id))
	clearTagCache()
	return nil
}

// AddTagToSkill 为技能添加标签
func (r *Repository) AddTagToSkill(ctx context.Context, skillID, tagID uint) error {
	skillTag := &models.SkillTag{
		SkillID: skillID,
		TagID:   tagID,
	}
	return r.db.WithContext(ctx).Create(skillTag).Error
}

// RemoveTagFromSkill 从技能中移除标签
func (r *Repository) RemoveTagFromSkill(ctx context.Context, skillID, tagID uint) error {
	return r.db.WithContext(ctx).Where("skill_id = ? AND tag_id = ?", skillID, tagID).Delete(&models.SkillTag{}).Error
}

// GetTagsBySkillID 获取技能的所有标签
func (r *Repository) GetTagsBySkillID(ctx context.Context, skillID uint) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.db.WithContext(ctx).Joins("JOIN skill_tags ON skill_tags.tag_id = tags.id").Where("skill_tags.skill_id = ?", skillID).Find(&tags).Error
	return tags, err
}
