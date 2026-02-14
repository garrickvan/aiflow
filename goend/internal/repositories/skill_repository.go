package repositories

import (
	"context"
	"strings"
	"time"

	"aiflow/internal/models"
	"aiflow/internal/utils"

	"github.com/go-ego/gse"
	"gorm.io/gorm"
)

// seg 是全局分词器实例
var seg gse.Segmenter

// init 初始化分词器
func init() {
	seg.LoadDict()
}

// Skill CRUD 操作

// CreateSkill 创建技能
func (r *Repository) CreateSkill(ctx context.Context, skill *models.Skill) error {
	// 检查 skill的 ResourceDir 是否存在, 如果不存在，随机4个字母 + 时间戳 作为目录名
	if skill.ResourceDir == "" {
		skill.ResourceDir = utils.GenerateRandomDirName()
	}
	// 设置时间戳，毫秒级精度
	timestamp := time.Now().UnixMilli()
	skill.CreatedAt = timestamp
	skill.UpdatedAt = timestamp

	// 使用事务创建技能并建立分词索引
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Create(skill).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 建立分词索引
	if err := r.buildSkillTokens(tx, skill.ID, skill.Name+" "+skill.Description); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetSkillByID 根据ID获取技能
func (r *Repository) GetSkillByID(ctx context.Context, id uint) (*models.Skill, error) {
	var skill models.Skill
	err := r.db.WithContext(ctx).Preload("Tags").First(&skill, id).Error
	if err != nil {
		return nil, err
	}
	return &skill, nil
}

// GetSkillByName 根据名称获取技能
func (r *Repository) GetSkillByName(ctx context.Context, name string) (*models.Skill, error) {
	var skill models.Skill
	err := r.db.WithContext(ctx).Preload("Tags").Where("name = ?", name).First(&skill).Error
	if err != nil {
		return nil, err
	}
	return &skill, nil
}

// ListAllSkills 获取所有技能
func (r *Repository) ListAllSkills(ctx context.Context) ([]models.Skill, error) {
	var skills []models.Skill
	err := r.db.WithContext(ctx).Find(&skills).Error
	return skills, err
}

// UpdateSkill 更新技能
func (r *Repository) UpdateSkill(ctx context.Context, skill *models.Skill) error {
	// 检查 skill的 ResourceDir 是否存在, 如果不存在，随机4个字母 + 时间戳 作为目录名
	if skill.ResourceDir == "" {
		skill.ResourceDir = utils.GenerateRandomDirName()
	}
	// 更新时间戳，毫秒级精度
	skill.UpdatedAt = time.Now().UnixMilli()

	// 使用事务更新技能并重建分词索引
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Save(skill).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除旧的分词索引
	if err := tx.Where("skill_id = ?", skill.ID).Delete(&models.SkillToken{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 重建分词索引
	if err := r.buildSkillTokens(tx, skill.ID, skill.Name+" "+skill.Description); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// DeleteSkill 删除技能（伪删除，进入回收站）
func (r *Repository) DeleteSkill(ctx context.Context, id uint) error {
	// 伪删除：设置 deleted_at 时间戳
	return r.db.WithContext(ctx).Model(&models.Skill{}).Where("id = ?", id).Update("deleted_at", time.Now().UnixMilli()).Error
}

// RestoreSkill 恢复回收站中的技能
func (r *Repository) RestoreSkill(ctx context.Context, id uint) error {
	// 恢复：清空 deleted_at 时间戳
	return r.db.WithContext(ctx).Model(&models.Skill{}).Where("id = ?", id).Update("deleted_at", 0).Error
}

// PermanentDeleteSkill 彻底删除技能
func (r *Repository) PermanentDeleteSkill(ctx context.Context, id uint) error {
	// 使用事务彻底删除技能及其分词索引
	tx := r.db.WithContext(ctx).Begin()

	// 删除分词索引
	if err := tx.Where("skill_id = ?", id).Delete(&models.SkillToken{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 彻底删除技能
	if err := tx.Unscoped().Delete(&models.Skill{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// buildSkillTokens 为技能建立分词索引
// 参数:
//
//	tx: 数据库事务
//	skillID: 技能ID
//	text: 需要分词的文本（名称+描述）
//
// 返回:
//
//	error: 错误信息
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

	// 批量插入分词索引
	for term := range termMap {
		token := models.SkillToken{
			SkillID: skillID,
			Term:    term,
		}
		if err := tx.Create(&token).Error; err != nil {
			return err
		}
	}

	return nil
}

// SearchSkillsByTokens 根据关键词分词搜索技能
// 参数:
//
//	ctx: 上下文
//	keyword: 搜索关键词
//
// 返回:
//
//	[]models.Skill: 匹配的技能列表，按匹配度降序排列
//	error: 错误信息
func (r *Repository) SearchSkillsByTokens(ctx context.Context, keyword string) ([]models.Skill, error) {
	// 先转小写再分词，确保大小写不敏感
	queryTokens := seg.Cut(strings.ToLower(keyword), true)

	// 过滤空白分词
	var terms []string
	for _, token := range queryTokens {
		token = strings.TrimSpace(token)
		if token != "" {
			terms = append(terms, strings.ToLower(token))
		}
	}

	if len(terms) == 0 {
		return r.ListAllSkills(ctx)
	}

	// 使用 SQL 查询匹配的技能，按匹配分词数量降序排列
	var skills []models.Skill
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Select("skills.*, COUNT(skill_tokens.term) as match_score").
		Joins("JOIN skill_tokens ON skill_tokens.skill_id = skills.id").
		Where("skill_tokens.term IN ?", terms).
		Group("skills.id").
		Order("match_score DESC").
		Find(&skills).Error

	return skills, err
}
