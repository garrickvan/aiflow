package mcp

import (
	"aiflow/internal/utils/logx"
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// initDetail 初始化技能详情工具
func initDetail(server *server.MCPServer) {
	server.AddTool(mcp.Tool{
		Name:        "skill_detail",
		Description: "查技能详情",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "技能名称",
				},
			},
			Required: []string{"name"},
		},
	}, detailTool)
}

// detailTool 是一个技能详情工具，用于查看技能的详细信息
func detailTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取技能参数
	skillName := request.GetString("name", "")

	logx.Debug("skill detail: %s", skillName)

	// 构建技能详情文本
	var skillDetail string
	if repo == nil {
		skillDetail = "数据库未初始化，无法获取技能详情"
	} else {
		// 从技能表中获取第一个匹配的技能
		skill, err := repo.GetSkillByName(ctx, skillName)
		if err != nil {
			logx.Error("获取技能详情失败: %v", err)
			skillDetail = "未知技能：" + skillName + "\n"
		} else {
			skillDetail = "技能详情：\n"
			skillDetail += "名称: " + skill.Name + "\n"
			skillDetail += "描述: " + skill.Description + "\n"
			if skill.Detail != "" {
				skillDetail += "详细信息: " + skill.Detail + "\n"
			}
			if skill.Compatibility != "" {
				skillDetail += "兼容性: " + skill.Compatibility + "\n"
			}
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: skillDetail,
			},
		},
	}, nil
}
