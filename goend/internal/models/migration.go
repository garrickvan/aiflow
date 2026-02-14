package models

import (
	"context"
	"log"

	"gorm.io/gorm"
)

// MigrateData 执行所有数据迁移
// 所有迁移错误都静默处理，不中断流程，确保应用正常启动
func MigrateData(db *gorm.DB) error {
	ctx := context.Background()

	// 1. 从SkillGroup迁移到Tag
	_ = migrateSkillGroupToTag(db, ctx)

	// 2. 删除job_tasks表的废弃module_path字段
	_ = migrateDropModulePathColumn(db, ctx)

	return nil
}

// migrateDropModulePathColumn 删除job_tasks表的module_path废弃字段
// 该字段在早期版本中存在，现已移除，需要清理数据库结构
func migrateDropModulePathColumn(db *gorm.DB, ctx context.Context) error {
	// 检查job_tasks表是否存在module_path字段
	var columnCount int64
	err := db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM pragma_table_info('job_tasks') WHERE name = 'module_path'",
	).Scan(&columnCount).Error
	if err != nil {
		log.Printf("检查module_path字段失败: %v", err)
		return nil // 非致命错误，继续执行
	}

	if columnCount > 0 {
		log.Println("开始数据库迁移: 删除job_tasks表的module_path字段")
		// SQLite不支持直接删除列，需要重建表
		if err := db.WithContext(ctx).Exec(`
			CREATE TABLE job_tasks_new (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				job_no VARCHAR(50) NOT NULL UNIQUE,
				project VARCHAR(100) NOT NULL,
				type VARCHAR(20) NOT NULL,
				goal TEXT NOT NULL,
				pass_accept_std BOOLEAN DEFAULT false,
				status VARCHAR(20) NOT NULL DEFAULT 'created',
				execution_records TEXT,
				active_execution_sequence INTEGER DEFAULT 0,
				created_at INTEGER,
				updated_at INTEGER,
				deleted_at INTEGER
			)
		`).Error; err != nil {
			log.Printf("创建新表失败: %v", err)
			return nil
		}

		// 复制数据
		if err := db.WithContext(ctx).Exec(`
			INSERT INTO job_tasks_new (
				id, job_no, project, type, goal, pass_accept_std, status,
				execution_records, active_execution_sequence, created_at, updated_at, deleted_at
			)
			SELECT
				id, job_no, project, type, goal, pass_accept_std, status,
				execution_records, active_execution_sequence, created_at, updated_at, deleted_at
			FROM job_tasks
		`).Error; err != nil {
			log.Printf("复制数据失败: %v", err)
			return nil
		}

		// 删除旧表，重命名新表
		if err := db.WithContext(ctx).Exec(`
			DROP TABLE job_tasks;
			ALTER TABLE job_tasks_new RENAME TO job_tasks
		`).Error; err != nil {
			log.Printf("替换表失败: %v", err)
			return nil
		}

		// 重建索引
		_ = db.WithContext(ctx).Exec("CREATE INDEX idx_job_tasks_status ON job_tasks(status)")
		_ = db.WithContext(ctx).Exec("CREATE INDEX idx_job_tasks_project ON job_tasks(project)")

		log.Println("数据库迁移完成: 已删除module_path字段")
	} else {
		log.Println("无需数据库迁移: job_tasks表没有module_path字段")
	}

	return nil
}

// migrateSkillGroupToTag 从旧的SkillGroup结构迁移数据到新的Tag结构
// 如果旧表不存在或数据已删除，静默忽略不报错
func migrateSkillGroupToTag(db *gorm.DB, ctx context.Context) error {
	// 检查是否存在旧的SkillGroup表
	var skillGroupCount int64
	if err := db.WithContext(ctx).Model(&SkillGroup{}).Count(&skillGroupCount).Error; err != nil {
		log.Printf("检查SkillGroup表失败(可能表不存在): %v", err)
		return nil
	}

	if skillGroupCount == 0 {
		log.Println("无需数据迁移: 未发现SkillGroup表数据")
		return nil
	}

	log.Println("开始数据迁移: 从SkillGroup到Tag")

	// 迁移SkillGroup数据到Tag表
	var skillGroups []SkillGroup
	if err := db.WithContext(ctx).Find(&skillGroups).Error; err != nil {
		log.Printf("查询SkillGroup数据失败: %v", err)
		return nil
	}

	for _, group := range skillGroups {
		tag := Tag{
			ID:        group.ID,
			Name:      group.Name,
			CreatedAt: group.CreatedAt,
			UpdatedAt: group.UpdatedAt,
			DeletedAt: group.DeletedAt,
		}
		if err := db.WithContext(ctx).Create(&tag).Error; err != nil {
			log.Printf("迁移分组 %s 失败: %v", group.Name, err)
		}
	}

	// 迁移Skill数据，创建SkillTag关联
	var skills []OldSkill
	if err := db.WithContext(ctx).Find(&skills).Error; err != nil {
		log.Printf("查询OldSkill数据失败: %v", err)
		return nil
	}

	for _, skill := range skills {
		if skill.GroupID > 0 {
			skillTag := SkillTag{
				SkillID: skill.ID,
				TagID:   skill.GroupID,
			}
			if err := db.WithContext(ctx).Create(&skillTag).Error; err != nil {
				log.Printf("为技能 %s 创建标签关联失败: %v", skill.Name, err)
			}
		}
	}

	log.Println("数据迁移完成: 从SkillGroup到Tag")
	return nil
}

// 旧的SkillGroup结构，用于数据迁移
type SkillGroup struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"type:varchar(100);not null;uniqueIndex"`
	CreatedAt int64  `gorm:"index"`
	UpdatedAt int64
	DeletedAt int64      `gorm:"index"`
	Skills    []OldSkill `gorm:"foreignKey:GroupID"`
}

// 旧的Skill结构，用于数据迁移
type OldSkill struct {
	ID            uint   `gorm:"primaryKey;autoIncrement"`
	Name          string `gorm:"type:varchar(100);not null;uniqueIndex"`
	ResourceDir   string `gorm:"type:varchar(100);not null;uniqueIndex"`
	Description   string `gorm:"type:text"`
	Detail        string
	License       string `gorm:"type:varchar(100)"`
	Compatibility string `gorm:"type:text"`
	Metadata      string `gorm:"type:text"`
	AllowedTools  string `gorm:"type:text"`
	GroupID       uint
	CreatedAt     int64 `gorm:"index"`
	UpdatedAt     int64
	DeletedAt     int64 `gorm:"index"`
}
