# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.0] - 2026-02-14

### Added
- MCP服务端实现，支持技能动态加载和执行
- 任务管理功能，支持任务的创建、执行、状态跟踪和报告
- 技能持久化存储，支持SQLite数据库
- 前端管理界面重构，优化技能和任务管理体验
- VSCode插件扩展，支持侧边栏集成和快捷操作
- 系统托盘功能，支持后台运行和快速访问

### Changed
- 统一项目版本号为0.5.0
- 优化前端组件结构，提升代码可维护性
- 改进API响应格式，统一错误处理

### Fixed
- 修复技能执行时的并发问题
- 修复前端路由状态同步问题

## [0.2.0] - 2026-01-30

### Added
- 前端项目初始化，包含React + TypeScript + Vite配置
- 后端项目初始化，包含Go + Chi + GORM配置
- 技能管理功能，支持技能的创建、编辑、删除和查询
- 技能分组功能，支持分组的创建、编辑、删除和查询
- 技能上传功能，支持通过文件上传技能
- 静态文件服务，支持前端资源的访问
- 项目文档结构，包含API、架构、数据库等文档

### Changed
- 统一项目版本号为0.2.0

### Fixed
- 修复了后端路由配置问题
- 修复了前端API请求路径问题

### Security
- 初始化项目安全配置
