package models

import (
	"gorm.io/gorm"
)

// SkillToken 技能分词索引表
// 用于实现技能名称和描述的全文分词搜索
type SkillToken struct {
	SkillID uint   `gorm:"primaryKey;index:idx_skill_tokens_skill_id"`
	Term    string `gorm:"type:varchar(100);primaryKey;index:idx_skill_tokens_term"`
}

// Tag 标签模型
type Tag struct {
	ID        uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string  `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	CreatedAt int64   `gorm:"index" json:"createdAt"`
	UpdatedAt int64   `json:"updatedAt"`
	DeletedAt int64   `gorm:"index" json:"-"`
	Skills    []Skill `gorm:"many2many:skill_tags;" json:"skills,omitempty"`
}

// SkillTag 技能标签关联表
type SkillTag struct {
	SkillID uint `gorm:"primaryKey;index:idx_skill_tags_skill_id"`
	TagID   uint `gorm:"primaryKey;index:idx_skill_tags_tag_id"`
}

// Skill 技能模型
// name	✅	1-64 字符；仅小写字母、数字、连字符；不首尾连字符、无连续连字符；必须与目录名一致	name: pdf-processing
// description	✅	1-1024 字符；非空；描述功能 + 触发时机；第三人称；含关键词便于技能发现	description: Extract text/tables from PDFs, fill forms, merge docs. Use when handling PDF, forms, document extraction.
// license	❌	许可证名称或绑定文件路径	license: Apache-2.0
// compatibility	❌	最大 500 字符；说明环境需求（产品、系统包、网络等）	compatibility: Requires Claude Code, Python 3.11, pdfplumber
// metadata	❌	任意键值对；用于版本、作者等附加信息	metadata: {author: example-org, version: "1.0"}
// allowed-tools	❌	空格分隔的预批准工具列表（实验性）	allowed-tools: Bash(git:*) Bash(jq:*) Read
type Skill struct {
	ID            uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	ResourceDir   string `gorm:"type:varchar(100);not null;uniqueIndex" json:"resourceDir"`
	Description   string `gorm:"type:text" json:"description"`
	License       string `gorm:"type:varchar(100)" json:"license"`
	Version       string `gorm:"type:varchar(50)" json:"version"`
	Compatibility string `gorm:"type:text" json:"compatibility"`
	Metadata      string `gorm:"type:text" json:"metadata"`
	AllowedTools  string `gorm:"type:text;column:allowed_tools" json:"allowedTools"`
	Detail        string `json:"detail,omitempty"`
	Tags          []Tag  `gorm:"many2many:skill_tags;" json:"tags,omitempty"`

	CreatedAt int64 `gorm:"index" json:"createdAt"`
	UpdatedAt int64 `json:"updatedAt"`
	DeletedAt int64 `gorm:"index" json:"-"`
}

// JobTaskType 任务类型常量定义
const (
	JobTaskTypeNewFeature      = "新需求"   // 新功能开发
	JobTaskTypeBugFix          = "Bug修复" // 缺陷修复
	JobTaskTypeImprovement     = "改进功能"  // 功能优化改进
	JobTaskTypeRefactoring     = "重构代码"  // 代码重构
	JobTaskTypeUnitTest        = "单元测试"  // 单元测试编写
	JobTaskTypeIntegrationTest = "集成测试"  // 集成测试编写
)

// JobTask 任务模型
// 对应AI敏捷工作流规范手册中的标准需求任务
// 任务编号: 按"项目代号-日期-序号"规则生成
// 所属项目: 任务关联的项目名称
// 任务类型: 新需求/Bug修复/改进功能/重构代码/单元测试/集成测试
// 任务目标: 简洁明确的任务目标描述，尽量不超过50字
// 关联模块: 需要处理的相关模块路径
// 验收状态: 是否通过验收
// 完成阶段: 已创建、处理中、处理失败、处理完成、验收通过
// 执行记录: JSON格式数组，记录每次执行的状态和结果
// 统一用伪删除，避免删除数据后导致的问题
type JobTask struct {
	ID                      uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	JobNo                   string `gorm:"type:varchar(50);not null;unique" json:"jobNo"`     // 任务编号
	Project                 string `gorm:"type:varchar(100);not null" json:"project"`         // 所属项目
	Type                    string `gorm:"type:varchar(20);not null" json:"type"`             // 任务类型
	Goal                    string `gorm:"type:text;not null" json:"goal"`                    // 任务目标
	PassAcceptStd           bool   `gorm:"type:boolean;default:false" json:"passAcceptStd"`   // 验收状态
	Status                  string `gorm:"type:varchar(20);not null" json:"status"`           // 完成阶段
	ExecutionRecords        string `gorm:"type:text" json:"executionRecords"`                 // 执行记录(JSON数组)
	ActiveExecutionSequence int    `gorm:"type:int;default:0" json:"activeExecutionSequence"` // 当前活跃执行序号

	CreatedAt int64 `gorm:"index" json:"createdAt"`
	UpdatedAt int64 `json:"updatedAt"`
	DeletedAt int64 `gorm:"index" json:"-"`
}

// ExecutionRecord 单次执行结果记录
type ExecutionRecord struct {
	Sequence     int      `json:"sequence"`     // 执行序号
	Status       string   `json:"status"`       // 执行状态
	Result       string   `json:"result"`       // 执行结果描述
	Solution     string   `json:"solution"`     // 解决方案
	RelatedFiles []string `json:"relatedFiles"` // 关联文件列表
	AcceptStd    string   `json:"acceptStd"`    // 验收标准
	Skills       []string `json:"skills"`       // 使用的技能列表
	CreatedAt    int64    `json:"createdAt"`    // 创建时间（毫秒级时间戳）
	UpdatedAt    int64    `json:"updatedAt"`    // 更新时间（毫秒级时间戳）
}

// CreateIndexes 创建数据库索引优化查询性能
// 参数:
//   - db: GORM数据库连接
//
// 返回:
//   - error: 错误信息
func CreateIndexes(db *gorm.DB) error {
	// 添加复合索引优化常用查询
	_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_job_tasks_status ON job_tasks(status)")
	_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_job_tasks_project ON job_tasks(project)")
	_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_skill_tags_skill_id ON skill_tags(skill_id)")
	_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_skill_tags_tag_id ON skill_tags(tag_id)")
	_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_skill_tokens_skill_id ON skill_tokens(skill_id)")
	_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_skill_tokens_term ON skill_tokens(term)")

	return nil
}
