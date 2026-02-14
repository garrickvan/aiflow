# 数据库设计说明文档

## 1. 数据库概述

本项目使用 **SQLite** 作为数据库，通过 **GORM** 框架进行 ORM 操作。SQLite 是一款轻量级的文件数据库，适合小型应用和开发环境使用，无需单独的数据库服务器。

数据库主要存储以下核心数据：
- 技能 (Skill)
- 标签 (Tag)
- 技能标签关联 (SkillTag)
- 技能分词索引 (SkillToken)
- 任务 (JobTask)

## 2. 表结构设计

### 2.1 技能表 (skills)

| 字段名 | 数据类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `id` | `INTEGER` | `PRIMARY KEY, AUTOINCREMENT` | 技能ID |
| `name` | `VARCHAR(100)` | `NOT NULL, UNIQUE` | 技能名称（1-64字符，小写字母、数字、连字符） |
| `resource_dir` | `VARCHAR(100)` | `NOT NULL, UNIQUE` | 资源目录名（只能包含字母、数字和下划线） |
| `description` | `TEXT` | | 技能描述（1-1024字符） |
| `license` | `VARCHAR(100)` | | 许可证名称或绑定文件路径 |
| `compatibility` | `TEXT` | | 兼容性信息（最大500字符） |
| `metadata` | `TEXT` | | 元数据（JSON格式） |
| `allowed_tools` | `TEXT` | | 允许的工具列表（空格分隔） |
| `created_at` | `BIGINT` | `INDEX` | 创建时间戳（毫秒级） |
| `updated_at` | `BIGINT` | | 更新时间戳（毫秒级） |
| `deleted_at` | `BIGINT` | `INDEX` | 删除时间戳（软删除） |

### 2.2 标签表 (tags)

| 字段名 | 数据类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `id` | `INTEGER` | `PRIMARY KEY, AUTOINCREMENT` | 标签ID |
| `name` | `VARCHAR(100)` | `NOT NULL, UNIQUE` | 标签名称 |
| `created_at` | `BIGINT` | `INDEX` | 创建时间戳（毫秒级） |
| `updated_at` | `BIGINT` | | 更新时间戳（毫秒级） |
| `deleted_at` | `BIGINT` | `INDEX` | 删除时间戳（软删除） |

### 2.3 技能标签关联表 (skill_tags)

| 字段名 | 数据类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `skill_id` | `INTEGER` | `PRIMARY KEY, INDEX` | 技能ID |
| `tag_id` | `INTEGER` | `PRIMARY KEY, INDEX` | 标签ID |

### 2.4 技能分词索引表 (skill_tokens)

| 字段名 | 数据类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `skill_id` | `INTEGER` | `PRIMARY KEY, INDEX` | 技能ID |
| `term` | `VARCHAR(100)` | `PRIMARY KEY, INDEX` | 分词词条 |

### 2.5 任务表 (job_tasks)

| 字段名 | 数据类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `id` | `INTEGER` | `PRIMARY KEY, AUTOINCREMENT` | 任务ID |
| `job_no` | `VARCHAR(50)` | `NOT NULL, UNIQUE` | 任务编号（格式：JT-项目-日期-序号） |
| `project` | `VARCHAR(100)` | `NOT NULL` | 所属项目 |
| `type` | `VARCHAR(20)` | `NOT NULL` | 任务类型 |
| `goal` | `TEXT` | `NOT NULL` | 任务目标 |
| `pass_accept_std` | `BOOLEAN` | `DEFAULT false` | 是否通过验收 |
| `status` | `VARCHAR(20)` | `NOT NULL` | 任务状态 |
| `execution_records` | `TEXT` | | 执行记录（JSON数组） |
| `active_execution_sequence` | `INTEGER` | `DEFAULT 0` | 当前活跃执行序号 |
| `created_at` | `BIGINT` | `INDEX` | 创建时间戳（毫秒级） |
| `updated_at` | `BIGINT` | | 更新时间戳（毫秒级） |
| `deleted_at` | `BIGINT` | `INDEX` | 删除时间戳（软删除） |

## 3. 字段详细说明

### 3.1 Skill 模型字段说明

| 字段 | 类型 | 约束 | 说明 |
| :--- | :--- | :--- | :--- |
| `ID` | `uint` | `primaryKey, autoIncrement` | 技能唯一标识符，自增主键 |
| `Name` | `string` | `type:varchar(100), not null, uniqueIndex` | 技能名称，唯一且非空 |
| `ResourceDir` | `string` | `type:varchar(100), not null, uniqueIndex` | 资源目录名，唯一且非空 |
| `Description` | `string` | `type:text` | 技能描述，详细说明技能功能 |
| `License` | `string` | `type:varchar(100)` | 许可证名称或绑定文件路径 |
| `Compatibility` | `string` | `type:text` | 兼容性信息，说明环境需求 |
| `Metadata` | `string` | `type:text` | 元数据，JSON格式存储附加信息 |
| `AllowedTools` | `string` | `type:text` | 允许的工具列表，空格分隔 |
| `Detail` | `string` | | 技能详情，运行时填充 |
| `Tags` | `[]Tag` | `gorm:"many2many:skill_tags;"` | 关联的标签列表，多对多关系 |
| `CreatedAt` | `int64` | `index` | 创建时间戳（毫秒级），用于索引和排序 |
| `UpdatedAt` | `int64` | | 更新时间戳（毫秒级） |
| `DeletedAt` | `int64` | `index` | 删除时间戳，用于软删除 |

### 3.2 Tag 模型字段说明

| 字段 | 类型 | 约束 | 说明 |
| :--- | :--- | :--- | :--- |
| `ID` | `uint` | `primaryKey, autoIncrement` | 标签唯一标识符，自增主键 |
| `Name` | `string` | `type:varchar(100), not null, uniqueIndex` | 标签名称，唯一且非空 |
| `CreatedAt` | `int64` | `index` | 创建时间戳（毫秒级） |
| `UpdatedAt` | `int64` | | 更新时间戳（毫秒级） |
| `DeletedAt` | `int64` | `index` | 删除时间戳，用于软删除 |
| `Skills` | `[]Skill` | `gorm:"many2many:skill_tags;"` | 关联的技能列表，多对多关系 |

### 3.3 JobTask 模型字段说明

| 字段 | 类型 | 约束 | 说明 |
| :--- | :--- | :--- | :--- |
| `ID` | `uint` | `primaryKey, autoIncrement` | 任务唯一标识符，自增主键 |
| `JobNo` | `string` | `type:varchar(50), not null, unique` | 任务编号，唯一且非空 |
| `Project` | `string` | `type:varchar(100), not null` | 所属项目名称 |
| `Type` | `string` | `type:varchar(20), not null` | 任务类型 |
| `Goal` | `string` | `type:text, not null` | 任务目标描述 |
| `PassAcceptStd` | `bool` | `default:false` | 是否通过验收标准 |
| `Status` | `string` | `type:varchar(20), not null` | 任务状态 |
| `ExecutionRecords` | `string` | `type:text` | 执行记录，JSON格式数组 |
| `ActiveExecutionSequence` | `int` | `default:0` | 当前活跃执行序号 |
| `CreatedAt` | `int64` | `index` | 创建时间戳（毫秒级） |
| `UpdatedAt` | `int64` | | 更新时间戳（毫秒级） |
| `DeletedAt` | `int64` | `index` | 删除时间戳，用于软删除 |

### 3.4 ExecutionRecord 结构说明

```go
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
```

## 4. 关系设计

### 4.1 实体关系图 (ERD)

```
+---------------+       M:N       +---------------+
|    Skill      |<--------------->|     Tag       |
+---------------+                 +---------------+
| - id          |                 | - id          |
| - name        |                 | - name        |
| - resource_dir|                 | - created_at  |
| - description |                 | - updated_at  |
| - license     |                 | - deleted_at  |
| - compatibility|                +---------------+
| - metadata    |
| - allowed_tools|
| - created_at  |
| - updated_at  |
| - deleted_at  |
+---------------+
        |
        | 1:N
        v
+---------------+
|  SkillToken   |
+---------------+
| - skill_id    |
| - term        |
+---------------+

+---------------+
|   JobTask     |
+---------------+
| - id          |
| - job_no      |
| - project     |
| - type        |
| - goal        |
| - status      |
| - pass_accept_std |
| - execution_records |
| - active_execution_sequence |
| - created_at  |
| - updated_at  |
| - deleted_at  |
+---------------+
```

### 4.2 关系说明

- **Skill 与 Tag**：多对多关系
  - 通过 `skill_tags` 关联表实现
  - 一个技能可以有多个标签
  - 一个标签可以关联多个技能

- **Skill 与 SkillToken**：一对多关系
  - 一个技能可以有多个分词词条
  - 用于实现技能名称和描述的全文搜索

## 5. 索引设计

| 表名 | 字段 | 索引类型 | 说明 |
| :--- | :--- | :--- | :--- |
| `skills` | `name` | `UNIQUE` | 确保技能名称唯一 |
| `skills` | `resource_dir` | `UNIQUE` | 确保资源目录名唯一 |
| `skills` | `created_at` | `INDEX` | 加速按创建时间排序和查询 |
| `skills` | `deleted_at` | `INDEX` | 加速软删除相关查询 |
| `tags` | `name` | `UNIQUE` | 确保标签名称唯一 |
| `tags` | `created_at` | `INDEX` | 加速按创建时间排序和查询 |
| `tags` | `deleted_at` | `INDEX` | 加速软删除相关查询 |
| `skill_tags` | `skill_id` | `INDEX` | 加速按技能查询标签 |
| `skill_tags` | `tag_id` | `INDEX` | 加速按标签查询技能 |
| `skill_tokens` | `skill_id` | `INDEX` | 加速按技能查询词条 |
| `skill_tokens` | `term` | `INDEX` | 加速按词条搜索技能 |
| `job_tasks` | `job_no` | `UNIQUE` | 确保任务编号唯一 |
| `job_tasks` | `status` | `INDEX` | 加速按状态查询任务 |
| `job_tasks` | `project` | `INDEX` | 加速按项目查询任务 |
| `job_tasks` | `created_at` | `INDEX` | 加速按创建时间排序和查询 |
| `job_tasks` | `deleted_at` | `INDEX` | 加速软删除相关查询 |

## 6. 数据库操作

### 6.1 数据库初始化

数据库初始化代码位于 `goend/internal/repositories/repository.go`：

```go
// NewRepository 创建新的数据库仓库实例
func NewRepository(dbPath string) (*Repository, error) {
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // 自动迁移数据库表结构
    err = db.AutoMigrate(&Skill{}, &Tag{}, &SkillTag{}, &SkillToken{}, &JobTask{})
    if err != nil {
        return nil, fmt.Errorf("failed to migrate database: %w", err)
    }

    // 创建索引优化查询性能
    err = models.CreateIndexes(db)
    if err != nil {
        return nil, fmt.Errorf("failed to create indexes: %w", err)
    }

    return &Repository{db: db}, nil
}
```

### 6.2 主要操作方法

#### 技能操作
- `CreateSkill`: 创建技能
- `GetSkillByID`: 根据ID获取技能
- `GetSkillByName`: 根据名称获取技能
- `ListAllSkills`: 获取所有技能
- `UpdateSkill`: 更新技能
- `DeleteSkill`: 删除技能（软删除）
- `RestoreSkill`: 恢复技能
- `PermanentDeleteSkill`: 彻底删除技能
- `SearchSkillsByTokens`: 分词搜索技能

#### 标签操作
- `CreateTag`: 创建标签
- `GetTagByID`: 根据ID获取标签
- `GetTagByName`: 根据名称获取标签
- `ListTags`: 获取所有标签
- `UpdateTag`: 更新标签
- `DeleteTag`: 删除标签

#### 任务操作
- `CreateJobTask`: 创建任务
- `GetJobTaskByID`: 根据ID获取任务
- `GetJobTaskByJobNo`: 根据任务编号获取任务
- `ListJobTasks`: 获取任务列表
- `UpdateJobTask`: 更新任务
- `DeleteJobTask`: 删除任务（软删除）
- `RestoreJobTask`: 恢复任务
- `PermanentDeleteJobTask`: 彻底删除任务

## 7. 配置和使用

### 7.1 数据库配置

数据库文件路径在 `goend/internal/config/config.go` 中定义：

```go
const (
    // DBPath 默认数据库文件路径
    DBPath = "./db/aiflow.db"
)
```

### 7.2 使用示例

```go
// 初始化数据库连接
repo, err := repositories.NewRepository("./db/aiflow.db")
if err != nil {
    log.Fatalf("Failed to initialize database: %v", err)
}

// 创建标签
tag := &models.Tag{
    Name: "文本处理",
}
err = repo.CreateTag(context.Background(), tag)

// 创建技能
skill := &models.Skill{
    Name:        "text-analysis",
    ResourceDir: "text_analysis",
    Description: "分析文本内容，提取关键信息",
}
err = repo.CreateSkill(context.Background(), skill)

// 创建任务
jobTask := &models.JobTask{
    JobNo:   "JT-智流MCP-20250211-001",
    Project: "智流MCP",
    Type:    "新需求",
    Goal:    "实现用户登录功能",
    Status:  "已创建",
}
err = repo.CreateJobTask(context.Background(), jobTask)
```

## 8. 数据安全

### 8.1 注意事项

1. **SQLite文件权限**：确保数据库文件具有适当的读写权限
2. **输入验证**：所有用户输入在保存到数据库前应进行验证
3. **软删除**：使用 `deleted_at` 字段实现软删除，避免数据丢失
4. **唯一约束**：通过唯一索引确保关键字段的唯一性

### 8.2 最佳实践

- 定期备份数据库文件
- 避免在生产环境中使用SQLite处理大量并发操作
- 对于大型应用，考虑迁移到更强大的数据库系统如PostgreSQL或MySQL

## 9. 性能优化

1. **索引使用**：合理使用索引加速查询
2. **分词索引**：使用 `skill_tokens` 表实现高效的全文搜索
3. **批量操作**：对于大量数据操作，使用批量插入和更新
4. **预加载**：使用GORM的Preload功能减少N+1查询问题

## 10. 总结

本数据库设计针对智流MCP系统的需求，提供了简洁而完整的数据存储方案。使用SQLite作为存储引擎，结合GORM的便捷操作，既满足了功能需求，又保持了系统的轻量性和可移植性。

数据库结构清晰，关系简单，适合当前项目的规模和需求。随着项目的发展，可以根据实际使用情况进行适当的调整和优化。
