package mcp

import (
	"aiflow/internal/repositories"

	"github.com/mark3labs/mcp-go/server"
)

// repo 全局数据库仓库实例
var repo *repositories.Repository

// InitTools 初始化工具，向MCP服务器添加greet工具
func InitTools(server *server.MCPServer, r *repositories.Repository) {
	repo = r
	initMenu(server)
	initDetail(server)
	initSave(server)
	initJobTask(server)
}
