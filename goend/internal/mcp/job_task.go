package mcp

import (
	"aiflow/internal/models"
	"aiflow/internal/utils/logx"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// splitString 将逗号分隔的字符串分割为字符串数组
// 空字符串返回空数组而非包含空字符串的数组
func splitString(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

// initJobTask 初始化任务相关MCP工具
func initJobTask(server *server.MCPServer) {
	// 注册创建新任务工具
	server.AddTool(mcp.Tool{
		Name:        "job_new",
		Description: "创建新任务，用于跟踪和管理项目中的具体工作任务",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"project": map[string]any{
					"type":        "string",
					"description": "所属项目",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "任务类型，可选值：" + JobTypeOptions,
				},
				"goal": map[string]any{
					"type":        "string",
					"description": "当前任务核心目标的简要描述，用于复盘和管理",
				},
				"relatedFiles": map[string]any{
					"type":        "string",
					"description": "任务涉及的相关文件或文件夹路径，多个文件或文件夹就用逗号分隔",
				},
				"solution": map[string]any{
					"type":        "string",
					"description": "达成目标的具体解决思路，包括使用的技能、工具和步骤",
				},
				"acceptStd": map[string]any{
					"type":        "string",
					"description": "验收标准，包括：" + AcceptStdOptions,
				},
				"skills": map[string]any{
					"type":        "string",
					"description": "使用的技能列表，多个技能用逗号分隔",
				},
			},
			Required: []string{"project", "type", "goal", "relatedFiles", "solution", "acceptStd", "skills"},
		},
	}, newJobTool)
	// 注册报告任务执行结果工具
	server.AddTool(mcp.Tool{
		Name:        "job_report",
		Description: "报告任务执行结果",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"jobNo": map[string]any{
					"type":        "string",
					"description": "任务编号",
				},
				"status": map[string]any{
					"type":        "string",
					"description": "任务状态，可选值：" + JobStatusOptions,
				},
				"result": map[string]any{
					"type":        "string",
					"description": "任务执行结果",
				},
				"passAcceptStd": map[string]any{
					"type":        "boolean",
					"description": "是否通过验收标准",
				},
			},
			Required: []string{"jobNo", "status", "result", "passAcceptStd"},
		},
	}, reportJobTool)
	// 注册重做任务工具
	server.AddTool(mcp.Tool{
		Name:        "job_redo",
		Description: "用新的解决思路执行任务，以达到目标，保持任务编号不变，以便跟踪管理执行情况",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"jobNo": map[string]any{
					"type":        "string",
					"description": "任务编号",
				},
				"solution": map[string]any{
					"type":        "string",
					"description": "达成目标的具体解决思路，包括使用的技能、工具和步骤",
				},
				"relatedFiles": map[string]any{
					"type":        "string",
					"description": "任务涉及的相关文件或文件夹路径，多个文件或文件夹就用逗号分隔",
				},
				"skills": map[string]any{
					"type":        "string",
					"description": "使用的技能列表，多个技能用逗号分隔",
				},
			},
			Required: []string{"jobNo", "solution", "relatedFiles"},
		},
	}, redoJobTool)
	// 注册查询任务工具
	server.AddTool(mcp.Tool{
		Name:        "job_get",
		Description: "查询任务详情",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"jobNo": map[string]any{
					"type":        "string",
					"description": "任务编号",
				},
			},
			Required: []string{"jobNo"},
		},
	}, queryJobTool)
}

// queryJobTool 查询任务详情工具函数
// 根据任务编号查询并返回任务详细信息
func queryJobTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	jobNo := request.GetString("jobNo", "")

	logx.Debug("job_get - jobNo: %s", jobNo)

	// 检查数据库是否初始化
	if repo == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "数据库未初始化，无法查询任务",
				},
			},
		}, nil
	}

	// 根据任务编号查询任务
	jobTask, err := repo.GetJobTaskByJobNo(ctx, jobNo)
	if err != nil {
		logx.Error("查询任务失败: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "查询任务失败: " + err.Error(),
				},
			},
		}, nil
	}

	// 解析执行记录
	var executionRecords []models.ExecutionRecord
	if jobTask.ExecutionRecords != "" {
		if err = json.Unmarshal([]byte(jobTask.ExecutionRecords), &executionRecords); err != nil {
			logx.Error("解析执行结果失败: %v", err)
			executionRecords = []models.ExecutionRecord{}
		}
	}

	// 构建执行记录详情
	var executionDetails strings.Builder
	for i, record := range executionRecords {
		if i > 0 {
			executionDetails.WriteString("\n---\n")
		}
		executionDetails.WriteString(fmt.Sprintf("执行序号: %d\n", record.Sequence))
		executionDetails.WriteString(fmt.Sprintf("状态: %s\n", record.Status))
		executionDetails.WriteString(fmt.Sprintf("解决思路: %s\n", record.Solution))
		executionDetails.WriteString(fmt.Sprintf("涉及文件: %s\n", strings.Join(record.RelatedFiles, ",")))
		if record.Result != "" {
			executionDetails.WriteString(fmt.Sprintf("执行结果: %s\n", record.Result))
		}
		executionDetails.WriteString(fmt.Sprintf("验收标准: %s", record.AcceptStd))
	}

	// 返回任务详情
	passStatus := "未通过"
	if jobTask.PassAcceptStd {
		passStatus = "已通过"
	}
	resultText := fmt.Sprintf("任务详情:\n任务编号: %s\n所属项目: %s\n任务类型: %s\n任务目标: %s\n验收状态: %s\n当前状态: %s\n当前执行序号: %d\n\n执行记录:\n%s",
		jobTask.JobNo,
		jobTask.Project,
		jobTask.Type,
		jobTask.Goal,
		passStatus,
		jobTask.Status,
		jobTask.ActiveExecutionSequence,
		executionDetails.String(),
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// newJobTool 创建任务工具函数
// 根据输入参数创建新任务，任务编号按"项目代号-日期-序号"规则生成
func newJobTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	project := request.GetString("project", "")
	jobType := request.GetString("type", "")
	goal := request.GetString("goal", "")
	relatedFiles := request.GetString("relatedFiles", "")
	solution := request.GetString("solution", "")
	acceptStd := request.GetString("acceptStd", "")
	skills := request.GetString("skills", "")

	logx.Debug("job_new - project: %s, type: %s, goal: %s, relatedFiles: %s, solution: %s, acceptStd: %s, skills: %s", project, jobType, goal, relatedFiles, solution, acceptStd, skills)

	// 检查数据库是否初始化
	if repo == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "数据库未初始化，无法创建任务",
				},
			},
		}, nil
	}

	// 生成任务编号: 项目代号-日期-序号
	jobNo := generateJobNo(project)

	now := time.Now().UnixMilli()
	executionRecords := []models.ExecutionRecord{
		{
			Sequence:     1,
			Status:       JobStatusProcessing,
			Result:       "",
			RelatedFiles: splitString(relatedFiles),
			Solution:     solution,
			AcceptStd:    acceptStd,
			Skills:       splitString(skills),
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	exes, err := json.Marshal(executionRecords)
	if err != nil {
		logx.Error("序列化执行记录失败: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "序列化执行记录失败: " + err.Error(),
				},
			},
		}, nil
	}

	// 创建任务对象
	jobTask := &models.JobTask{
		JobNo:                   jobNo,
		Project:                 project,
		Type:                    jobType,
		Goal:                    goal,
		PassAcceptStd:           false, // 默认未通过验收
		Status:                  JobStatusCreated,
		ActiveExecutionSequence: 1,
		ExecutionRecords:        string(exes),
	}

	// 保存到数据库
	if err := repo.CreateJobTask(ctx, jobTask); err != nil {
		logx.Error("创建任务失败: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "创建任务失败: " + err.Error(),
				},
			},
		}, nil
	}

	// 返回成功结果
	resultText := fmt.Sprintf("任务创建成功\n任务编号: %s	", jobNo)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// reportJobTool 报告任务执行结果工具函数
// 更新任务状态并将执行结果追加到executionRecords数组末尾
func reportJobTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	jobNo := request.GetString("jobNo", "")
	status := request.GetString("status", "")
	result := request.GetString("result", "")
	passAcceptStd := request.GetBool("passAcceptStd", false)

	logx.Debug("job_report - jobNo: %s, status: %s, result: %s, passAcceptStd: %v", jobNo, status, result, passAcceptStd)

	// 检查数据库是否初始化
	if repo == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "数据库未初始化，无法报告任务",
				},
			},
		}, nil
	}

	// 根据任务编号查询任务
	jobTask, err := repo.GetJobTaskByJobNo(ctx, jobNo)
	if err != nil {
		logx.Error("查询任务失败: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "查询任务失败: " + err.Error(),
				},
			},
		}, nil
	}

	// 更新任务状态
	jobTask.Status = status

	// 解析现有的executionRecords
	var executionRecords []models.ExecutionRecord
	if jobTask.ExecutionRecords != "" {
		if err = json.Unmarshal([]byte(jobTask.ExecutionRecords), &executionRecords); err != nil {
			logx.Error("解析执行结果失败: %v", err)
			// 如果解析失败，重置为空数组
			executionRecords = []models.ExecutionRecord{}
		}
	}

	// 找出当前执行记录
	var executionRecord *models.ExecutionRecord = nil
	for i := range executionRecords {
		if executionRecords[i].Sequence == jobTask.ActiveExecutionSequence {
			executionRecord = &executionRecords[i]
			break
		}
	}

	if executionRecord == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "报告任务失败: 未找到当前执行记录",
				},
			},
		}, nil
	}

	// 更新执行记录
	executionRecord.Status = status
	executionRecord.Result = result
	executionRecord.UpdatedAt = time.Now().UnixMilli()

	// 更新JobTask的验收状态
	jobTask.PassAcceptStd = passAcceptStd

	// 序列化回JSON
	executionJSON, err := json.Marshal(executionRecords)
	if err != nil {
		logx.Error("序列化执行结果失败: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "报告任务失败: " + err.Error(),
				},
			},
		}, nil
	}
	jobTask.ExecutionRecords = string(executionJSON)

	// 保存到数据库
	if err := repo.UpdateJobTask(ctx, jobTask); err != nil {
		logx.Error("报告任务失败: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "报告任务失败: " + err.Error(),
				},
			},
		}, nil
	}

	// 返回成功结果
	resultText := fmt.Sprintf("任务报告成功\n任务编号: %s\n当前状态: %s\n历史记录数: %d",
		jobNo, status, len(executionRecords))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// generateJobNo 生成任务编号
// 格式: JT-项目代号-日期-序号 (如: JT-ZL-20250207-001)
func generateJobNo(project string) string {
	// 获取当前日期
	now := time.Now()
	dateStr := now.Format("20060102")

	// 生成序号: 使用当前时间戳的后5位作为序号
	// 这样可以保证同一秒内创建的任务也有不同的编号
	seq := now.UnixMilli() % JobNoSequenceMod

	return fmt.Sprintf("JT-%s-%s-%05d", project, dateStr, seq)
}

// redoJobTool 重复执行任务工具函数
// 根据任务编号查询并返回任务详细信息
func redoJobTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	jobNo := request.GetString("jobNo", "")
	solution := request.GetString("solution", "")
	relatedFiles := request.GetString("relatedFiles", "")
	skills := request.GetString("skills", "")

	logx.Debug("job_redo - jobNo: %s, solution: %s, relatedFiles: %s, skills: %s", jobNo, solution, relatedFiles, skills)

	// 检查数据库是否初始化
	if repo == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "数据库未初始化，无法获取任务",
				},
			},
		}, nil
	}

	// 根据任务编号查询任务
	jobTask, err := repo.GetJobTaskByJobNo(ctx, jobNo)
	if err != nil {
		logx.Error("查询任务失败: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "查询任务失败: " + err.Error(),
				},
			},
		}, nil
	}

	// 解析执行记录
	var executionRecords []models.ExecutionRecord
	if jobTask.ExecutionRecords != "" {
		if err = json.Unmarshal([]byte(jobTask.ExecutionRecords), &executionRecords); err != nil {
			logx.Error("解析执行结果失败: %v", err)
			executionRecords = []models.ExecutionRecord{}
		}
	}

	jobTask.ActiveExecutionSequence = len(executionRecords) + 1
	// 继承上一次的验收标准
	var inheritAcceptStd string
	if len(executionRecords) > 0 {
		inheritAcceptStd = executionRecords[len(executionRecords)-1].AcceptStd
	}
	now := time.Now().UnixMilli()
	newExecutionRecord := models.ExecutionRecord{
		Sequence:     jobTask.ActiveExecutionSequence,
		Solution:     solution,
		RelatedFiles: splitString(relatedFiles),
		Status:       JobStatusProcessing,
		AcceptStd:    inheritAcceptStd,
		Skills:       splitString(skills),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	executionRecords = append(executionRecords, newExecutionRecord)

	// 序列化回JSON
	executionJSON, err := json.Marshal(executionRecords)
	if err != nil {
		logx.Error("序列化执行结果失败: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "重复执行任务失败: " + err.Error(),
				},
			},
		}, nil
	}
	jobTask.ExecutionRecords = string(executionJSON)
	// 保存到数据库
	if err := repo.UpdateJobTask(ctx, jobTask); err != nil {
		logx.Error("重复执行任务失败: %v", err)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "重复执行任务失败: " + err.Error(),
				},
			},
		}, nil
	}

	resultText := fmt.Sprintf("任务内容:\n任务类型: %s\n任务目标: %s\n",
		jobTask.Type,
		jobTask.Goal,
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}
