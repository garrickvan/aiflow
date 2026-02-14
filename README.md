# 智流MCP AiFlow

> 智流MCP，你的AI工作助手

智流MCP是一个提高AI工作流效率的MCP服务，通过管理技能、处理任务来提升AI工作流的效率。

## 功能特性

### 技能管理
- **技能存储** - 支持创建、查看、更新、删除技能
- **标签分类** - 为技能添加标签，便于分类管理
- **回收站** - 软删除机制，支持恢复误删技能
- **导入导出** - 支持技能导出为Markdown格式
- **关键词搜索** - 支持分词搜索技能

### 任务管理
- **任务跟踪** - 创建任务并跟踪执行过程
- **执行记录** - 每次执行都有独立记录，支持多次重试
- **项目分类** - 按项目组织任务
- **状态流转** - 已创建 → 处理中 → 处理完成/失败 → 验收通过
- **批量导出** - 支持CSV、JSON、Markdown格式导出

### MCP工具
智流MCP提供以下MCP工具供AI调用：

| 工具名 | 功能 |
|--------|------|
| `skill_get` | 查询技能列表（支持标签筛选和关键词搜索） |
| `skill_detail` | 查看技能详情 |
| `skill_save` | 保存/更新技能 |
| `job_new` | 创建新任务 |
| `job_get` | 查询任务详情 |
| `job_report` | 报告任务执行结果 |
| `job_redo` | 重新执行任务（新思路） |

### 支持范围

理论上所有支持MCP协议的AI都可以使用智流MCP，目前在Trae ID、Codebuddy插件中都测试通过。

### 使用方法

1. 先运行aiflow程序，确保MCP服务已启动
2. 确保AI Agent或IDE已连接到智流MCP服务
3. 调用MCP工具，例如：`skill_get`、`job_new`等
4. 浏览器打开http://localhost:9990/web，可管理skill，查看执行结果和任务状态

Agent使用时，配合rules文件，可实现自动处理任务。

### 规则文件

rules文件是一个YAML格式的文件，用于定义任务处理规则。例如：

```markdown
# 项目名：myproject

## 【执行前必做】

1. 检查任务编号：无则调用 `job_new`创建，有则提取
2. 无论任务大小，响应首行必须是任务状态声明
3. 判断是否需要 `job_report`

## 【规则正文】

1. 【任务识别】无"任务编号: JT-{项目名}-{日期}-{序号}"时调用 `job_new`创建，有则提取复用，但绝对不能编造任务编号
2. 【任务报告】代码修改完成/告知完成/任务失败时调用 `job_report`。测试/编译验收先执行再调用，人工验收可直接调用
3. 【验收选择】有测试脚本→测试验收；有编译命令→编译验收；其他→人工验收
4. 【任务重开】重新执行时先 `job_get`查看历史，再调整 `job_redo`
5. 【强制输出】响应首行声明状态：创建后输出"任务编号: JT-XXX"、执行中输出"任务: JT-XXX 执行中"、完成后输出"任务: JT-XXX 已归档"

## 【违规处理】

未遵守规则时立即停止，说明违规点，重新按正确流程执行
```

## 项目结构

```
aiflow/
├── goend/             # Go后端代码（MCP服务 + HTTP API）
│   ├── cmd/
│   │   ├── api/       # HTTP服务入口
│   │   └── migrate/   # 数据库迁移工具
│   ├── internal/
│   │   ├── api/       # HTTP API handlers和路由
│   │   ├── cache/     # 本地缓存
│   │   ├── config/    # 配置管理
│   │   ├── mcp/       # MCP工具实现
│   │   ├── models/    # 数据模型
│   │   ├── repositories/  # 数据访问层
│   │   ├── services/  # 业务逻辑层
│   │   └── utils/     # 工具函数
│   └── scripts/       # 构建脚本
├── frontend/          # 前端代码（React + TypeScript + Vite）
│   ├── src/
│   │   ├── components/    # 组件
│   │   ├── pages/         # 页面
│   │   ├── services/      # API服务
│   │   └── stores/        # 状态管理
│   └── public/
├── vsplugin/          # VSCode插件
├── docs/              # 项目文档
├── .trae/rules/       # 项目规则配置
├── README.md          # 项目说明
├── LICENSE            # 开源许可证
├── CHANGELOG.md       # 版本变更记录
├── go.mod             # Go模块文件
└── package.json       # 前端依赖
```

## 项目开发

### 前置要求

- Go 1.20+
- Node.js 18+
- Yarn

### 后端启动

```bash
# 进入后端目录
cd goend

# 安装依赖
go mod tidy

# 启动服务（默认端口9900）
go run cmd/api/main.go

# 或使用自定义端口
go run cmd/api/main.go -http localhost:9990
```

服务启动后访问：
- Web后台: http://localhost:9900/web
- MCP端点: http://localhost:9900/mcp

### 前端启动

```bash
# 进入前端目录
cd frontend

# 安装依赖
yarn install

# 启动开发服务器
yarn dev
```

前端开发服务器默认运行在 http://localhost:5173

### 构建完整应用

项目提供了构建脚本，用于将前端静态资源打包到Go可执行文件中。

```bash
# 1. 首先构建前端
cd frontend
yarn build

# 2. 执行构建脚本
cd ../goend
python scripts/builder.py
```

构建成功后，会在 `goend/release` 目录下生成 `aiflow.exe` 可执行文件，包含完整的前端静态资源和后端服务。

## 配置说明

配置文件 `config.yml`：

```yaml
server:
  name: "智流MCP"           # 服务名称
  version: "0.5.0"          # 版本号
  addr: "localhost:9900"    # 监听地址
  web_path: "/web"          # Web后台路径
  root_path: "/"            # MCP根路径
  mcp_path: "/mcp"          # MCP协议路径

log:
  level: "info"             # 日志级别: debug/info/warn/error
  output_type: "console"    # 输出方式: console/file
  file_path: ""             # 日志文件路径（output_type为file时生效）
```

## 项目文档

- [API文档](docs/api.md)
- [系统架构](docs/architecture.md)
- [数据库设计](docs/database.md)
- [部署指南](docs/deployment.md)
- [开发环境搭建](docs/development.md)

## VSCode插件

智流MCP提供配套的VSCode插件，可在编辑器内直接访问管理后台。

详见 [vsplugin/README.md](vsplugin/README.md)

## 许可证

本项目采用 MIT 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件。
