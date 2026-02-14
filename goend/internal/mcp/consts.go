package mcp

// 任务编号生成相关常量
const (
	// JobNoSequenceMod 任务编号序号取模值，用于生成5位序号
	JobNoSequenceMod = 100000
)

// 任务类型常量
const (
	// JobTypeNewFeature 新需求
	JobTypeNewFeature = "新需求"
	// JobTypeBugFix Bug修复
	JobTypeBugFix = "Bug修复"
	// JobTypeImprovement 改进功能
	JobTypeImprovement = "改进功能"
	// JobTypeRefactoring 重构代码
	JobTypeRefactoring = "重构代码"
	// JobTypeUnitTest 单元测试
	JobTypeUnitTest = "单元测试"
	// JobTypeIntegrationTest 集成测试
	JobTypeIntegrationTest = "集成测试"
)

// 任务类型选项字符串（用于MCP工具描述）
const JobTypeOptions = "新需求、Bug修复、改进功能、重构代码、单元测试、集成测试、数据处理、版本控制"

// 验收标准选项字符串（用于MCP工具描述）
const AcceptStdOptions = "人工验收、测试验收、编译验收"

// 任务状态常量
const (
	// JobStatusCreated 已创建
	JobStatusCreated = "已创建"
	// JobStatusProcessing 处理中
	JobStatusProcessing = "处理中"
	// JobStatusFailed 处理失败
	JobStatusFailed = "处理失败"
	// JobStatusCompleted 处理完成
	JobStatusCompleted = "处理完成"
	// JobStatusAccepted 验收通过
	JobStatusAccepted = "验收通过"
)

// 任务状态选项字符串（用于MCP工具描述）
const JobStatusOptions = "已创建、处理中、处理失败、处理完成、验收通过"
