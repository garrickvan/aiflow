# 智流 MCP (AiFlow) - Agent Development Guide

> 本文件面向 AI 编程助手，提供项目背景、架构说明和开发指南。

## 项目概述

**智流 MCP (AiFlow)** 是一个用于管理 AI 工作流技能的任务处理服务。它通过管理技能（Skills）和处理任务（JobTasks）来提升 AI 工作流的效率。

### 核心功能

- **技能管理**: 创建、编辑、删除、查询技能，支持标签分类和全文搜索
- **任务管理**: 按照 AI 敏捷工作流规范管理任务（新需求、Bug修复、功能改进等）
- **MCP 服务**: 提供 Model Context Protocol 接口供 AI 客户端调用
- **回收站机制**: 技能和任务支持伪删除，可恢复或彻底删除

## 技术栈

### 前端 (`frontend/`)

| 技术 | 版本 | 用途 |
|------|------|------|
| React | ^19.2.0 | UI 框架 |
| TypeScript | ~5.9.3 | 类型系统 |
| Vite | ^5.4.0 | 构建工具 |
| Ant Design | ^6.2.2 | UI 组件库 |
| Zustand | ^5.0.11 | 状态管理 |
| React Router | ^7.13.0 | 路由管理 |
| React Markdown | ^10.1.0 | Markdown 渲染 |

### 后端 (`goend/`)

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.24.5+ | 后端语言 |
| Chi Router | v5.2.4 | HTTP 路由 |
| GORM | v1.25.6 | ORM 框架 |
| SQLite | - | 数据库 |
| MCP-Go | v0.43.2 | MCP 协议实现 |
| Systray | v1.2.2 | 系统托盘 |

### VSCode 插件 (`vsplugin/`)

| 技术 | 版本 | 用途 |
|------|------|------|
| TypeScript | ^4.9.4 | 开发语言 |
| VSCode API | ^1.74.0 | 扩展 API |

## 项目结构

```
aiflow/
├── frontend/              # 前端应用 (React + TypeScript)
│   ├── src/
│   │   ├── components/    # 通用组件
│   │   ├── pages/         # 页面组件
│   │   ├── services/      # API 服务层
│   │   ├── stores/        # Zustand 状态管理
│   │   ├── types/         # TypeScript 类型定义
│   │   └── utils/         # 工具函数
│   ├── package.json       # Yarn 包管理
│   └── vite.config.ts     # Vite 配置
│
├── goend/                 # Go 后端应用
│   ├── cmd/
│   │   ├── api/           # API 服务入口
│   │   │   ├── main.go    # 程序入口
│   │   │   └── static/    # 嵌入的前端静态资源
│   │   └── migrate/       # 数据库迁移工具
│   ├── internal/
│   │   ├── api/           # HTTP API 层
│   │   │   ├── handlers/  # 请求处理器
│   │   │   └── routers.go # 路由注册
│   │   ├── cache/         # 本地缓存
│   │   ├── config/        # 配置管理
│   │   ├── mcp/           # MCP 工具实现
│   │   ├── models/        # 数据模型
│   │   ├── repositories/  # 数据访问层
│   │   ├── services/      # 业务逻辑层
│   │   └── utils/         # 工具函数
│   ├── db/                # SQLite 数据库文件
│   ├── scripts/
│   │   └── build.py       # 构建脚本
│   ├── config.yml         # 配置文件
│   └── go.mod             # Go 模块定义
│
├── vsplugin/              # VSCode 扩展
│   ├── src/
│   │   └── extension.ts   # 扩展主文件
│   └── package.json       # 扩展配置
│
└── docs/                  # 项目文档
    ├── architecture.md    # 架构说明
    ├── api.md             # API 文档
    └── database.md        # 数据库设计
```

## 构建命令

### 前端开发

```bash
cd frontend
yarn install      # 安装依赖
yarn dev          # 启动开发服务器 (Vite)
yarn build        # 生产构建
yarn lint         # ESLint 检查
```

### 后端开发

```bash
cd goend
go mod tidy       # 安装依赖
go run ./cmd/api  # 启动服务
go test ./...     # 运行测试
```

### 完整构建（前端 + 后端）

```bash
# 需要 Python 3.7+
cd goend
python scripts/build.py
```

构建流程：
1. 构建前端应用 (`yarn build`)
2. 复制前端产物到 `cmd/api/static/`
3. 生成 Windows 资源文件（图标）
4. 构建 Go 可执行文件（无控制台窗口）
5. 输出: `goend/aiflow.exe`

### VSCode 插件构建

```bash
cd vsplugin
npm install       # 安装依赖
npm run compile   # 编译 TypeScript
vsce package      # 打包扩展 (.vsix)
```

## 代码风格指南

### Go 代码规范

- **命名规范**: 使用驼峰命名法，导出成员首字母大写
- **包注释**: 每个包应该有包级别的注释说明用途
- **函数注释**: 公共函数必须包含注释，格式：
  ```go
  // FunctionName 函数描述
  // 参数:
  //   - param1: 参数说明
  // 返回:
  //   - 返回值说明
  ```
- **错误处理**: 显式检查错误，使用 `logx` 包记录日志
- **代码组织**: 按功能分层（api → service → repository → model）

### TypeScript/React 代码规范

- **类型定义**: 优先使用 TypeScript 类型，避免 `any`
- **组件命名**: 大写驼峰法（PascalCase）
- ** hooks 命名**: 以 `use` 开头
- **样式**: 优先使用 Ant Design 组件和 CSS 变量
- **状态管理**: 使用 Zustand 管理全局状态

### 注释语言

项目主要使用 **简体中文** 进行注释和文档编写。

## 数据库模型

### 核心实体

**Skill（技能）**
- `id`: 主键
- `name`: 技能名称（唯一）
- `resourceDir`: 资源目录（唯一）
- `description`: 描述
- `detail`: 详细内容（Markdown）
- `license`: 许可证
- `compatibility`: 兼容性说明
- `metadata`: 元数据（JSON 字符串）
- `allowedTools`: 允许的工具列表
- `tags`: 多对多关联标签
- `createdAt/updatedAt/deletedAt`: 时间戳（软删除）

**Tag（标签）**
- `id`: 主键
- `name`: 标签名称（唯一）
- `skills`: 关联的技能列表

**JobTask（任务）**
- `id`: 主键
- `jobNo`: 任务编号（唯一，格式：项目代号-日期-序号）
- `project`: 所属项目
- `type`: 任务类型（新需求/Bug修复/改进功能/重构代码/单元测试/集成测试）
- `goal`: 任务目标
- `passAcceptStd`: 验收状态
- `status`: 完成阶段（已创建/处理中/处理失败/处理完成/验收通过）
- `executionRecords`: 执行记录（JSON 数组）
- `activeExecutionSequence`: 当前活跃执行序号

### 数据库索引

- `idx_job_tasks_status`: 任务状态索引
- `idx_job_tasks_project`: 项目索引
- `idx_skill_tags_skill_id/tag_id`: 技能标签关联索引
- `idx_skill_tokens_skill_id/term`: 分词搜索索引

## API 设计

### REST API 端点

**标签管理**
- `GET /api/tags` - 获取标签列表
- `POST /api/tags` - 创建标签
- `GET/PUT/DELETE /api/tags/{id}` - 标签 CRUD

**技能管理**
- `GET /api/skills` - 获取技能列表（支持分页、标签筛选、日期筛选）
- `POST /api/skills` - 创建技能
- `GET/PUT/DELETE /api/skills/{id}` - 技能 CRUD
- `GET /api/skills/trash` - 回收站列表
- `POST /api/skills/{id}/restore` - 恢复技能
- `DELETE /api/skills/{id}/permanent` - 彻底删除
- `GET /api/skills/{id}/export` - 导出为 Markdown

**任务管理**
- `GET /api/jobtasks` - 获取任务列表（支持多条件筛选）
- `POST /api/jobtasks` - 创建任务
- `GET/PUT/DELETE /api/jobtasks/{id}` - 任务 CRUD
- `GET /api/jobtasks/projects` - 获取项目列表
- `GET /api/jobtasks/trash` - 回收站列表
- `POST /api/jobtasks/{id}/restore` - 恢复任务
- `DELETE /api/jobtasks/{id}/permanent` - 彻底删除
- `POST /api/jobtasks/export` - 批量导出

**文件上传**
- `POST /api/upload_data` - 上传技能文件

### MCP 工具

- `skill_get`: 查询技能列表（支持标签和关键词搜索）
- `skill_detail`: 获取技能详情
- `skill_save`: 保存技能
- `job_task_*`: 任务相关操作

## 配置说明

配置文件位置: `goend/config.yml`

```yaml
server:
  addr: "localhost:9900"     # HTTP 监听地址
  name: "智流"               # MCP 服务器名称
  version: "0.2.0"           # 版本号
  root_path: "/"             # 根路径
  mcp_path: "/mcp"           # MCP 端点路径
  web_path: "/web"           # Web 后台路径

log:
  level: "debug"             # 日志级别: debug/info/warn/error
  output_type: "std"         # 输出方式: std/file
  file_path: "./logs"        # 日志文件目录（当 output_type 为 file 时）
```

## 开发注意事项

### 前端开发

1. **API 基础路径**: 使用 `API_BASE_URL` 常量（定义在 `types/constant.ts`）
2. **状态管理**: 
   - `appStore`: 应用级状态（侧边栏折叠、移动端菜单等）
   - `modalStore`: 模态框状态管理
3. **响应式设计**: 支持移动端，断点 768px
4. **路由**: 使用查询参数进行页面导航 (`?page=skill|job`)

### 后端开发

1. **数据库**: 使用 SQLite，文件位于 `goend/db/aiflow.db`
2. **日志**: 使用 `internal/utils/logx` 包，支持分级和文件输出
3. **静态资源**: 前端构建后嵌入 Go 二进制（使用 `//go:embed static`）
4. **CORS**: 已配置允许跨域访问
5. **系统托盘**: 启动后显示在系统托盘，可通过托盘菜单退出

### 软删除机制

技能和任务都实现了软删除：
- 删除操作设置 `deleted_at` 字段为当前时间戳
- 查询默认排除 `deleted_at > 0` 的记录
- 回收站查询只返回 `deleted_at > 0` 的记录
- 恢复操作将 `deleted_at` 设为 0

### 分词搜索

技能搜索使用 `go-ego/gse` 进行中文分词：
- 技能创建/更新时自动分词并写入 `skill_tokens` 表
- 搜索时先分词再查询索引

## 测试策略

- **单元测试**: Go 代码使用标准 `testing` 包
- **测试文件**: 以 `_test.go` 结尾，与被测文件同目录
- **示例**: `goend/internal/mcp/skill_test.go`

## 部署流程

1. 执行完整构建: `python scripts/build.py`
2. 获取可执行文件: `goend/aiflow.exe`
3. 部署时需携带:
   - `aiflow.exe`（包含前端资源）
   - `config.yml`（配置文件）
   - `db/` 目录（数据库文件，可选）

## 版本管理

项目使用语义化版本 (Semantic Versioning):
- 当前版本: 0.2.0
- 版本格式: MAJOR.MINOR.PATCH

## 许可证

MIT License - 详见 LICENSE 文件

## 相关文档

- [架构说明](docs/architecture.md)
- [API 文档](docs/api.md)
- [数据库设计](docs/database.md)
- [部署指南](docs/deployment.md)
- [开发环境搭建](docs/development.md)
