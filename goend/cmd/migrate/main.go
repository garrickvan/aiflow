// migrate 数据迁移工具
// 将旧数据库aiflow_old.db的数据迁移到当前数据库
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"aiflow/internal/config"
	"aiflow/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 旧数据库模型定义

// OldSkill 旧版技能结构
type OldSkill struct {
	ID            uint   `gorm:"primaryKey;autoIncrement"`
	Name          string `gorm:"type:varchar(100);not null"`
	ResourceDir   string `gorm:"type:varchar(100);not null"`
	Description   string `gorm:"type:text"`
	License       string `gorm:"type:varchar(100)"`
	Compatibility string `gorm:"type:text"`
	Metadata      string `gorm:"type:text"`
	AllowedTools  string `gorm:"type:text"`
	Detail        string
	CreatedAt     int64 `gorm:"index"`
	UpdatedAt     int64
	DeletedAt     int64 `gorm:"index"`
}

// TableName 指定表名
func (OldSkill) TableName() string {
	return "skills"
}

// OldTag 旧版标签结构
type OldTag struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"type:varchar(100);not null"`
	CreatedAt int64  `gorm:"index"`
	UpdatedAt int64
	DeletedAt int64 `gorm:"index"`
}

// TableName 指定表名
func (OldTag) TableName() string {
	return "tags"
}

// OldSkillTag 旧版技能标签关联
type OldSkillTag struct {
	SkillID uint `gorm:"primaryKey"`
	TagID   uint `gorm:"primaryKey"`
}

// TableName 指定表名
func (OldSkillTag) TableName() string {
	return "skill_tags"
}

// OldSkillToken 旧版技能分词
type OldSkillToken struct {
	SkillID uint   `gorm:"primaryKey"`
	Term    string `gorm:"type:varchar(100);primaryKey"`
}

// TableName 指定表名
func (OldSkillToken) TableName() string {
	return "skill_tokens"
}

// OldSkillGroup 旧版技能分组
type OldSkillGroup struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"type:varchar(100);not null"`
	CreatedAt int64  `gorm:"index"`
	UpdatedAt int64
	DeletedAt int64 `gorm:"index"`
}

// TableName 指定表名
func (OldSkillGroup) TableName() string {
	return "skill_groups"
}

// OldSkillWithGroup 带分组的旧技能
type OldSkillWithGroup struct {
	OldSkill
	GroupID uint
}

// TableName 指定表名
func (OldSkillWithGroup) TableName() string {
	return "skills"
}

// OldWorkTask 旧版工作任务
type OldWorkTask struct {
	ID               uint   `gorm:"primaryKey;autoIncrement"`
	TaskNo           string `gorm:"type:varchar(50);not null"`
	Project          string `gorm:"type:varchar(100);not null"`
	Type             string `gorm:"type:varchar(20);not null"`
	Description      string `gorm:"type:text;not null"`
	ModulePath       string `gorm:"type:text;not null"`
	AcceptStd        string `gorm:"type:varchar(50);not null"`
	Status           string `gorm:"type:varchar(20);not null;default:'已创建'"`
	ExecutionResult  string `gorm:"type:text"`
	CreatedAt        int64  `gorm:"index"`
	UpdatedAt        int64
	DeletedAt        int64 `gorm:"index"`
}

// TableName 指定表名
func (OldWorkTask) TableName() string {
	return "work_tasks"
}

// OldWorkJob 旧版工作Job
type OldWorkJob struct {
	ID                      uint   `gorm:"primaryKey;autoIncrement"`
	JobNo                   string `gorm:"type:varchar(50);not null"`
	Project                 string `gorm:"type:varchar(100);not null"`
	Type                    string `gorm:"type:varchar(20);not null"`
	Goal                    string `gorm:"type:text;not null"`
	ModulePath              string `gorm:"type:text;not null"`
	AcceptStd               string `gorm:"type:varchar(50);not null"`
	Status                  string `gorm:"type:varchar(20);not null;default:'已创建'"`
	ExecutionRecords        string `gorm:"type:text"`
	ActiveExecutionSequence int    `gorm:"default:0"`
	PassAcceptStd           bool   `gorm:"default:false"`
	CreatedAt               int64  `gorm:"index"`
	UpdatedAt               int64
	DeletedAt               int64  `gorm:"index"`
}

// TableName 指定表名
func (OldWorkJob) TableName() string {
	return "work_jobs"
}

// OldJobTask 旧版JobTask（带module_path字段）
type OldJobTask struct {
	ID                      uint   `gorm:"primaryKey;autoIncrement"`
	JobNo                   string `gorm:"type:varchar(50);not null"`
	Project                 string `gorm:"type:varchar(100);not null"`
	Type                    string `gorm:"type:varchar(20);not null"`
	Goal                    string `gorm:"type:text;not null"`
	ModulePath              string `gorm:"type:text;not null"`
	PassAcceptStd           bool   `gorm:"default:false"`
	Status                  string `gorm:"type:varchar(20);not null;default:'已创建'"`
	ExecutionRecords        string `gorm:"type:text"`
	ActiveExecutionSequence int    `gorm:"default:0"`
	CreatedAt               int64  `gorm:"index"`
	UpdatedAt               int64
	DeletedAt               int64  `gorm:"index"`
}

// TableName 指定表名
func (OldJobTask) TableName() string {
	return "job_tasks"
}

// 迁移统计
var migrateStats = struct {
	Skills       int
	Tags         int
	SkillTags    int
	SkillTokens  int
	SkillGroups  int
	WorkTasks    int
	WorkJobs     int
	JobTasks     int
	TotalRecords int
}{}

func main() {
	oldDBPath := "./db/aiflow_old.db"
	newDBPath := config.DBPath

	// 检查旧数据库是否存在
	if _, err := os.Stat(oldDBPath); os.IsNotExist(err) {
		log.Fatalf("旧数据库不存在: %s", oldDBPath)
	}

	log.Println("开始数据迁移...")
	log.Printf("源数据库: %s", oldDBPath)
	log.Printf("目标数据库: %s", newDBPath)

	// 连接旧数据库
	oldDB, err := gorm.Open(sqlite.Open(oldDBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("连接旧数据库失败: %v", err)
	}

	// 连接新数据库
	newDB, err := gorm.Open(sqlite.Open(newDBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("连接新数据库失败: %v", err)
	}

	// 执行迁移
	migrateSkills(oldDB, newDB)
	migrateTags(oldDB, newDB)
	migrateSkillTags(oldDB, newDB)
	migrateSkillTokens(oldDB, newDB)
	migrateSkillGroups(oldDB, newDB)
	migrateWorkTasks(oldDB, newDB)
	migrateWorkJobs(oldDB, newDB)
	migrateJobTasks(oldDB, newDB)

	// 打印统计
	log.Println("\n========== 迁移统计 ==========")
	log.Printf("Skills:       %d", migrateStats.Skills)
	log.Printf("Tags:         %d", migrateStats.Tags)
	log.Printf("SkillTags:    %d", migrateStats.SkillTags)
	log.Printf("SkillTokens:  %d", migrateStats.SkillTokens)
	log.Printf("SkillGroups:  %d", migrateStats.SkillGroups)
	log.Printf("WorkTasks:    %d", migrateStats.WorkTasks)
	log.Printf("WorkJobs:     %d", migrateStats.WorkJobs)
	log.Printf("JobTasks:     %d", migrateStats.JobTasks)
	log.Printf("总计:         %d", migrateStats.TotalRecords)
	log.Println("==============================")
	log.Println("数据迁移完成!")
}

// migrateSkills 迁移技能数据
func migrateSkills(oldDB, newDB *gorm.DB) {
	var oldSkills []OldSkill
	if err := oldDB.Find(&oldSkills).Error; err != nil {
		log.Printf("查询旧skills失败: %v", err)
		return
	}

	if len(oldSkills) == 0 {
		log.Println("Skills: 无数据需要迁移")
		return
	}

	for _, old := range oldSkills {
		skill := models.Skill{
			ID:            old.ID,
			Name:          old.Name,
			ResourceDir:   old.ResourceDir,
			Description:   old.Description,
			License:       old.License,
			Compatibility: old.Compatibility,
			Metadata:      old.Metadata,
			AllowedTools:  old.AllowedTools,
			Detail:        old.Detail,
			CreatedAt:     old.CreatedAt,
			UpdatedAt:     old.UpdatedAt,
			DeletedAt:     old.DeletedAt,
		}

		if err := newDB.Create(&skill).Error; err != nil {
			log.Printf("迁移Skill %s 失败: %v", old.Name, err)
			continue
		}
		migrateStats.Skills++
	}

	migrateStats.TotalRecords += migrateStats.Skills
	log.Printf("Skills: 迁移 %d 条记录", migrateStats.Skills)
}

// migrateTags 迁移标签数据
func migrateTags(oldDB, newDB *gorm.DB) {
	var oldTags []OldTag
	if err := oldDB.Find(&oldTags).Error; err != nil {
		log.Printf("查询旧tags失败: %v", err)
		return
	}

	if len(oldTags) == 0 {
		log.Println("Tags: 无数据需要迁移")
		return
	}

	for _, old := range oldTags {
		tag := models.Tag{
			ID:        old.ID,
			Name:      old.Name,
			CreatedAt: old.CreatedAt,
			UpdatedAt: old.UpdatedAt,
			DeletedAt: old.DeletedAt,
		}

		if err := newDB.Create(&tag).Error; err != nil {
			log.Printf("迁移Tag %s 失败: %v", old.Name, err)
			continue
		}
		migrateStats.Tags++
	}

	migrateStats.TotalRecords += migrateStats.Tags
	log.Printf("Tags: 迁移 %d 条记录", migrateStats.Tags)
}

// migrateSkillTags 迁移技能标签关联
func migrateSkillTags(oldDB, newDB *gorm.DB) {
	var oldSkillTags []OldSkillTag
	if err := oldDB.Find(&oldSkillTags).Error; err != nil {
		log.Printf("查询旧skill_tags失败: %v", err)
		return
	}

	if len(oldSkillTags) == 0 {
		log.Println("SkillTags: 无数据需要迁移")
		return
	}

	for _, old := range oldSkillTags {
		skillTag := models.SkillTag{
			SkillID: old.SkillID,
			TagID:   old.TagID,
		}

		if err := newDB.Create(&skillTag).Error; err != nil {
			log.Printf("迁移SkillTag (skill_id=%d, tag_id=%d) 失败: %v", old.SkillID, old.TagID, err)
			continue
		}
		migrateStats.SkillTags++
	}

	migrateStats.TotalRecords += migrateStats.SkillTags
	log.Printf("SkillTags: 迁移 %d 条记录", migrateStats.SkillTags)
}

// migrateSkillTokens 迁移技能分词
func migrateSkillTokens(oldDB, newDB *gorm.DB) {
	var oldTokens []OldSkillToken
	if err := oldDB.Find(&oldTokens).Error; err != nil {
		log.Printf("查询旧skill_tokens失败: %v", err)
		return
	}

	if len(oldTokens) == 0 {
		log.Println("SkillTokens: 无数据需要迁移")
		return
	}

	for _, old := range oldTokens {
		token := models.SkillToken{
			SkillID: old.SkillID,
			Term:    old.Term,
		}

		if err := newDB.Create(&token).Error; err != nil {
			log.Printf("迁移SkillToken (skill_id=%d, term=%s) 失败: %v", old.SkillID, old.Term, err)
			continue
		}
		migrateStats.SkillTokens++
	}

	migrateStats.TotalRecords += migrateStats.SkillTokens
	log.Printf("SkillTokens: 迁移 %d 条记录", migrateStats.SkillTokens)
}

// migrateSkillGroups 迁移旧分组到标签
func migrateSkillGroups(oldDB, newDB *gorm.DB) {
	// 查询旧分组
	var oldGroups []OldSkillGroup
	if err := oldDB.Find(&oldGroups).Error; err != nil {
		log.Printf("查询旧skill_groups失败: %v", err)
		return
	}

	if len(oldGroups) == 0 {
		log.Println("SkillGroups: 无数据需要迁移")
		return
	}

	// 迁移分组为标签
	for _, old := range oldGroups {
		// 检查是否已存在同名标签
		var count int64
		newDB.Model(&models.Tag{}).Where("name = ?", old.Name).Count(&count)
		if count > 0 {
			log.Printf("SkillGroup %s 已存在同名标签，跳过", old.Name)
			continue
		}

		tag := models.Tag{
			Name:      old.Name,
			CreatedAt: old.CreatedAt,
			UpdatedAt: old.UpdatedAt,
			DeletedAt: old.DeletedAt,
		}

		if err := newDB.Create(&tag).Error; err != nil {
			log.Printf("迁移SkillGroup %s 到Tag失败: %v", old.Name, err)
			continue
		}
		migrateStats.SkillGroups++
	}

	// 查询带group_id的旧技能并创建关联
	var skillsWithGroup []OldSkillWithGroup
	if err := oldDB.Select("id, group_id").Where("group_id > 0").Find(&skillsWithGroup).Error; err != nil {
		log.Printf("查询带group_id的技能失败: %v", err)
		return
	}

	for _, skill := range skillsWithGroup {
		skillTag := models.SkillTag{
			SkillID: skill.ID,
			TagID:   skill.GroupID,
		}
		// 忽略重复键错误
		_ = newDB.Create(&skillTag)
	}

	migrateStats.TotalRecords += migrateStats.SkillGroups
	log.Printf("SkillGroups: 迁移 %d 条记录到Tags", migrateStats.SkillGroups)
}

// migrateWorkTasks 迁移旧工作任务到job_tasks
func migrateWorkTasks(oldDB, newDB *gorm.DB) {
	var oldTasks []OldWorkTask
	if err := oldDB.Find(&oldTasks).Error; err != nil {
		log.Printf("查询旧work_tasks失败: %v", err)
		return
	}

	if len(oldTasks) == 0 {
		log.Println("WorkTasks: 无数据需要迁移")
		return
	}

	for _, old := range oldTasks {
		// 构造执行记录
		execRecords := []models.ExecutionRecord{}
		if old.ExecutionResult != "" {
			record := models.ExecutionRecord{
				Sequence:     1,
				Status:       old.Status,
				Result:       old.ExecutionResult,
				Solution:     "",
				RelatedFiles: []string{old.ModulePath},
				AcceptStd:    old.AcceptStd,
				Skills:       []string{},
				CreatedAt:    old.CreatedAt,
				UpdatedAt:    old.UpdatedAt,
			}
			execRecords = append(execRecords, record)
		}

		execJSON, _ := json.Marshal(execRecords)

		// 状态映射
		status := mapStatus(old.Status)

		// 类型映射
		taskType := mapType(old.Type)

		jobTask := models.JobTask{
			ID:                      old.ID,
			JobNo:                   old.TaskNo,
			Project:                 old.Project,
			Type:                    taskType,
			Goal:                    old.Description,
			PassAcceptStd:           old.Status == "已完成" || old.Status == "验收通过",
			Status:                  status,
			ExecutionRecords:        string(execJSON),
			ActiveExecutionSequence: 1,
			CreatedAt:               old.CreatedAt,
			UpdatedAt:               old.UpdatedAt,
			DeletedAt:               old.DeletedAt,
		}

		if err := newDB.Create(&jobTask).Error; err != nil {
			log.Printf("迁移WorkTask %s 失败: %v", old.TaskNo, err)
			continue
		}
		migrateStats.WorkTasks++
	}

	migrateStats.TotalRecords += migrateStats.WorkTasks
	log.Printf("WorkTasks: 迁移 %d 条记录到JobTasks", migrateStats.WorkTasks)
}

// migrateWorkJobs 迁移旧WorkJobs到job_tasks
func migrateWorkJobs(oldDB, newDB *gorm.DB) {
	var oldJobs []OldWorkJob
	if err := oldDB.Find(&oldJobs).Error; err != nil {
		log.Printf("查询旧work_jobs失败: %v", err)
		return
	}

	if len(oldJobs) == 0 {
		log.Println("WorkJobs: 无数据需要迁移")
		return
	}

	for _, old := range oldJobs {
		// 状态映射
		status := mapStatus(old.Status)

		// 类型映射
		taskType := mapType(old.Type)

		jobTask := models.JobTask{
			ID:                      old.ID,
			JobNo:                   old.JobNo,
			Project:                 old.Project,
			Type:                    taskType,
			Goal:                    old.Goal,
			PassAcceptStd:           old.PassAcceptStd,
			Status:                  status,
			ExecutionRecords:        old.ExecutionRecords,
			ActiveExecutionSequence: old.ActiveExecutionSequence,
			CreatedAt:               old.CreatedAt,
			UpdatedAt:               old.UpdatedAt,
			DeletedAt:               old.DeletedAt,
		}

		if err := newDB.Create(&jobTask).Error; err != nil {
			log.Printf("迁移WorkJob %s 失败: %v", old.JobNo, err)
			continue
		}
		migrateStats.WorkJobs++
	}

	migrateStats.TotalRecords += migrateStats.WorkJobs
	log.Printf("WorkJobs: 迁移 %d 条记录到JobTasks", migrateStats.WorkJobs)
}

// migrateJobTasks 迁移旧JobTasks（带module_path的版本）
func migrateJobTasks(oldDB, newDB *gorm.DB) {
	var oldTasks []OldJobTask
	if err := oldDB.Find(&oldTasks).Error; err != nil {
		log.Printf("查询旧job_tasks失败: %v", err)
		return
	}

	if len(oldTasks) == 0 {
		log.Println("JobTasks: 无数据需要迁移")
		return
	}

	for _, old := range oldTasks {
		// 检查是否已存在
		var count int64
		newDB.Model(&models.JobTask{}).Where("job_no = ?", old.JobNo).Count(&count)
		if count > 0 {
			log.Printf("JobTask %s 已存在，跳过", old.JobNo)
			continue
		}

		// 状态映射
		status := mapStatus(old.Status)

		// 类型映射
		taskType := mapType(old.Type)

		jobTask := models.JobTask{
			ID:                      old.ID,
			JobNo:                   old.JobNo,
			Project:                 old.Project,
			Type:                    taskType,
			Goal:                    old.Goal,
			PassAcceptStd:           old.PassAcceptStd,
			Status:                  status,
			ExecutionRecords:        old.ExecutionRecords,
			ActiveExecutionSequence: old.ActiveExecutionSequence,
			CreatedAt:               old.CreatedAt,
			UpdatedAt:               old.UpdatedAt,
			DeletedAt:               old.DeletedAt,
		}

		if err := newDB.Create(&jobTask).Error; err != nil {
			log.Printf("迁移JobTask %s 失败: %v", old.JobNo, err)
			continue
		}
		migrateStats.JobTasks++
	}

	migrateStats.TotalRecords += migrateStats.JobTasks
	log.Printf("JobTasks: 迁移 %d 条记录", migrateStats.JobTasks)
}

// mapStatus 映射旧状态到新状态
func mapStatus(oldStatus string) string {
	switch oldStatus {
	case "已创建":
		return "已创建"
	case "处理中":
		return "处理中"
	case "处理失败":
		return "处理失败"
	case "处理完成", "已完成":
		return "处理完成"
	case "验收通过":
		return "验收通过"
	default:
		return oldStatus
	}
}

// mapType 映射旧类型到新类型
func mapType(oldType string) string {
	switch oldType {
	case "新需求":
		return models.JobTaskTypeNewFeature
	case "Bug修复":
		return models.JobTaskTypeBugFix
	case "改进功能":
		return models.JobTaskTypeImprovement
	case "重构代码":
		return models.JobTaskTypeRefactoring
	case "单元测试":
		return models.JobTaskTypeUnitTest
	case "集成测试":
		return models.JobTaskTypeIntegrationTest
	case "数据处理":
		return "数据处理"
	case "版本控制":
		return "版本控制"
	default:
		return oldType
	}
}

// getCurrentTimestamp 获取当前毫秒级时间戳
func getCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}

// generateJobNo 生成任务编号
func generateJobNo(project string) string {
	date := time.Now().Format("20060102")
	return fmt.Sprintf("JT-%s-%s-%d", project, date, getCurrentTimestamp())
}
