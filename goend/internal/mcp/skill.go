package mcp

import (
	"aiflow/internal/models"
	"aiflow/internal/utils/logx"
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// 检查MCP的必填参数
func initMenu(server *server.MCPServer) {
	server.AddTool(mcp.Tool{
		Name:        "skill_get",
		Description: "查技能",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"keyword": map[string]any{
					"type":        "string",
					"description": "传关键词获取技能，不传参则返回全部技能，最多20个",
				},
			},
			Required: []string{},
		},
	}, skillMenuTool)

	server.AddTool(mcp.Tool{
		Name:        "skill_by_tag",
		Description: "根据标签查技能",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"tag": map[string]any{
					"type":        "string",
					"description": "要查看的技能标签，最多返回1000条",
				},
			},
			Required: []string{"tag"},
		},
	}, skillByTagMenuTool)
}

// skillMenuTool 是一个简单的技能菜单工具，返回文本响应
// 最多返回20条技能记录
func skillMenuTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取关键词参数
	keyword := request.GetString("keyword", "")

	logx.Debug("keyword: %s", keyword)

	// 构建技能列表文本
	var skillList string
	if repo == nil {
		skillList = "数据库未初始化，无法获取技能列表"
	} else if keyword != "" {
		// 根据关键词进行分词搜索（使用数据库索引）
		skills, err := repo.SearchSkillsByTokens(ctx, keyword)
		if err != nil {
			logx.Error("关键词搜索技能失败: %v", err)
			skillList = "搜索技能失败: " + err.Error()
		} else {
			skillList = formatSkillList(skills, "关键词搜索结果：", 20)
		}
	} else {
		// 无需获取所有技能标签，直接查询所有技能
		skills, err := repo.ListAllSkills(ctx)
		if err != nil {
			logx.Error("获取技能列表失败: %v", err)
			skillList = "获取技能列表失败: " + err.Error()
		} else {
			skillList = formatSkillList(skills, "技能列表：", 20)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: skillList,
			},
		},
	}, nil
}

// skillByTagMenuTool 根据标签查询技能
func skillByTagMenuTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取标签参数
	tag := request.GetString("tag", "")

	logx.Debug("tag: %s", tag)

	// 构建技能列表文本
	var skillList string
	if repo == nil {
		skillList = "数据库未初始化，无法获取技能列表"
	} else if tag == "" {
		skillList = "标签参数不能为空"
	} else {
		// 获取指定标签的技能
		tagModel, err := repo.GetTagByName(ctx, tag)
		if err != nil {
			logx.Error("获取标签失败: %v", err)
			skillList = "获取技能列表失败: " + err.Error()
		} else {
			skillList = formatSkillList(tagModel.Skills, "标签「"+tag+"」的技能列表：", 1000)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: skillList,
			},
		},
	}, nil
}

// formatSkillList 格式化技能列表为字符串
// 参数:
//
//	skills: 技能列表
//	title: 列表标题
//	maxCount: 最多返回的记录数
//
// 返回:
//
//	string: 格式化后的字符串
func formatSkillList(skills []models.Skill, title string, maxCount int) string {
	if len(skills) == 0 {
		return "未找到匹配的技能"
	}

	result := title + "\n"

	// 限制最多返回maxCount条
	count := len(skills)
	if count > maxCount {
		count = maxCount
	}

	for i := 0; i < count; i++ {
		result += "name: " + skills[i].Name + " description: " + skills[i].Description + "\n"
	}

	// 如果还有更多的技能，提示用户
	if len(skills) > maxCount {
		result += fmt.Sprintf("... 还有 %d 个技能未显示\n", len(skills)-maxCount)
	}

	result += "请调用 skill_detail 查技能详情"

	return result
}
