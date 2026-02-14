package mcp

import (
	"aiflow/internal/models"
	"aiflow/internal/utils/logx"
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// initSave 初始化保存技能工具
func initSave(server *server.MCPServer) {
	server.AddTool(mcp.Tool{
		Name:        "skill_save",
		Description: "存技能",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "技能名称",
				},
				"resource_dir": map[string]any{
					"type":        "string",
					"description": "资源目录，只能包含字母、数字和下划线",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "技能描述",
				},
				"detail": map[string]any{
					"type":        "string",
					"description": "技能详情，Markdown格式，包含使用说明、示例代码等",
				},
			},
			Required: []string{"name", "resource_dir", "description", "detail"},
		},
	}, addTool)
}

// addTool 是一个添加技能的工具，用于向系统中添加新技能
func addTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	description := request.GetString("description", "")
	resourceDir := request.GetString("resource_dir", "")
	name := request.GetString("name", "")
	detail := request.GetString("detail", "")

	logx.Debug("add aiflow: description=%s, resource_dir=%s, name=%s, detail=%s", description, resourceDir, name, detail)

	// 检查技能是否已存在
	skill, err := repo.GetSkillByName(ctx, name)
	if err == nil {
		// 技能已存在，更新
		skill.Description = description
		skill.Detail = detail
		skill.ResourceDir = resourceDir
		err = repo.UpdateSkill(ctx, skill)
		if err != nil {
			logx.Error("failed to update skill: %v", err)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "更新技能失败",
					},
				},
			}, nil
		}
		result := "技能更新成功：\n"
		result += "名称: " + name + "\n"
		result += "描述: " + description
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: result,
				},
			},
		}, nil
	}

	// 技能不存在，创建新技能
	skill = &models.Skill{
		Name:        name,
		Description: description,
		Detail:      detail,
		ResourceDir: resourceDir,
	}
	err = repo.CreateSkill(ctx, skill)
	if err != nil {
		logx.Error("failed to create skill: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "创建技能失败",
				},
			},
		}, nil
	}

	result := "技能添加成功：\n"
	result += "名称: " + name + "\n"
	result += "描述: " + description
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: result,
			},
		},
	}, nil
}
