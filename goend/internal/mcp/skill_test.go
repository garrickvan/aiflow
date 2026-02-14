package mcp

import (
	"aiflow/internal/models"
	"aiflow/internal/repositories"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// setRepoForTest 设置全局repo变量（仅用于测试）
func setRepoForTest(r *repositories.Repository) {
	repo = r
}

// setupTestRepo 创建测试用的数据库仓库
// 参数:
//   - t: 测试实例
//
// 返回: 仓库实例和清理函数
func setupTestRepo(t *testing.T) (*repositories.Repository, func()) {
	// 创建唯一的临时数据库文件，避免测试间冲突
	dbPath := fmt.Sprintf("test_skill_%s.db", t.Name())

	// 清理已存在的测试数据库
	_ = os.Remove(dbPath)

	repo, err := repositories.NewRepository(dbPath)
	if err != nil {
		t.Fatalf("创建测试仓库失败: %v", err)
	}

	// 返回清理函数
	cleanup := func() {
		_ = os.Remove(dbPath)
	}

	return repo, cleanup
}

// createTestSkill 创建测试用的技能数据
// 参数:
//   - t: 测试实例
//   - repo: 数据库仓库
//   - name: 技能名称
//   - description: 技能描述
//
// 返回: 创建的技能实例
func createTestSkill(t *testing.T, repo *repositories.Repository, name, description string) *models.Skill {
	ctx := context.Background()
	skill := &models.Skill{
		Name:        name,
		Description: description,
		ResourceDir: name + "-dir",
	}

	err := repo.CreateSkill(ctx, skill)
	if err != nil {
		t.Fatalf("创建测试技能失败: %v", err)
	}

	return skill
}

// TestSearchSkillsByTokens_Basic 测试基础分词查询功能
func TestSearchSkillsByTokens_Basic(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试技能
	skill1 := createTestSkill(t, repo, "pdf-processing", "Extract text from PDF documents")
	skill2 := createTestSkill(t, repo, "image-resize", "Resize and optimize images")
	skill3 := createTestSkill(t, repo, "pdf-merge", "Merge multiple PDF files into one")

	// 测试用例1: 搜索"pdf"应该返回两个PDF相关技能
	t.Run("搜索关键词pdf", func(t *testing.T) {
		results, err := repo.SearchSkillsByTokens(ctx, "pdf")
		if err != nil {
			t.Fatalf("搜索失败: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("期望返回2个结果，实际返回%d个", len(results))
		}

		// 验证返回的技能ID
		resultIDs := make(map[uint]bool)
		for _, s := range results {
			resultIDs[s.ID] = true
		}

		if !resultIDs[skill1.ID] {
			t.Errorf("期望包含技能%d (pdf-processing)，但未找到", skill1.ID)
		}
		if !resultIDs[skill3.ID] {
			t.Errorf("期望包含技能%d (pdf-merge)，但未找到", skill3.ID)
		}
	})

	// 测试用例2: 搜索"image"应该返回一个技能
	t.Run("搜索关键词image", func(t *testing.T) {
		results, err := repo.SearchSkillsByTokens(ctx, "image")
		if err != nil {
			t.Fatalf("搜索失败: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("期望返回1个结果，实际返回%d个", len(results))
		}

		if len(results) > 0 && results[0].ID != skill2.ID {
			t.Errorf("期望返回技能%d (image-resize)，实际返回%d", skill2.ID, results[0].ID)
		}
	})

	// 测试用例3: 搜索不存在的关键词应该返回空结果
	t.Run("搜索不存在的关键词", func(t *testing.T) {
		results, err := repo.SearchSkillsByTokens(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("搜索失败: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("期望返回0个结果，实际返回%d个", len(results))
		}
	})
}

// TestSearchSkillsByTokens_Chinese 测试中文分词查询功能
func TestSearchSkillsByTokens_Chinese(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// 创建中文测试技能
	skill1 := createTestSkill(t, repo, "pdf-parser", "从PDF文件中提取文本和表格数据")
	skill2 := createTestSkill(t, repo, "excel-parser", "解析Excel文件并提取数据")
	_ = createTestSkill(t, repo, "pdf-converter", "将PDF文件转换为Word文档格式")

	// 测试用例1: 搜索"提取"应该返回包含该词的技能
	t.Run("搜索中文关键词提取", func(t *testing.T) {
		results, err := repo.SearchSkillsByTokens(ctx, "提取")
		if err != nil {
			t.Fatalf("搜索失败: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("期望返回2个结果，实际返回%d个", len(results))
		}

		resultIDs := make(map[uint]bool)
		for _, s := range results {
			resultIDs[s.ID] = true
		}

		if !resultIDs[skill1.ID] {
			t.Errorf("期望包含技能%d，但未找到", skill1.ID)
		}
		if !resultIDs[skill2.ID] {
			t.Errorf("期望包含技能%d，但未找到", skill2.ID)
		}
	})

	// 测试用例2: 搜索"PDF"应该返回两个PDF相关技能
	t.Run("搜索中文关键词PDF", func(t *testing.T) {
		results, err := repo.SearchSkillsByTokens(ctx, "PDF")
		if err != nil {
			t.Fatalf("搜索失败: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("期望返回2个结果，实际返回%d个", len(results))
		}
	})

	// 测试用例3: 搜索"文件"应该返回所有包含该词的技能
	t.Run("搜索中文关键词文件", func(t *testing.T) {
		results, err := repo.SearchSkillsByTokens(ctx, "文件")
		if err != nil {
			t.Fatalf("搜索失败: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("期望返回3个结果，实际返回%d个", len(results))
		}
	})
}

// TestSearchSkillsByTokens_MatchScore 测试分词匹配度排序
func TestSearchSkillsByTokens_MatchScore(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试技能，用于测试匹配度排序
	// skill1: 包含"pdf"和"processing"
	skill1 := createTestSkill(t, repo, "pdf-processing", "Process PDF files with advanced features")
	// skill2: 只包含"pdf"
	skill2 := createTestSkill(t, repo, "pdf-reader", "Read PDF files")
	// skill3: 包含"pdf"、"processing"和"advanced"
	skill3 := createTestSkill(t, repo, "pdf-advanced", "Advanced PDF processing with AI features")

	// 测试: 搜索"pdf processing"应该按匹配度排序
	t.Run("多关键词匹配度排序", func(t *testing.T) {
		results, err := repo.SearchSkillsByTokens(ctx, "pdf processing")
		if err != nil {
			t.Fatalf("搜索失败: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("期望返回3个结果，实际返回%d个", len(results))
			return
		}

		// skill3应该排在第一位，因为它匹配了"pdf"和"processing"
		// skill1应该排在第二位，也匹配了"pdf"和"processing"
		// skill2应该排在最后，只匹配了"pdf"

		// 验证skill3和skill1都匹配了两个词
		firstTwoIDs := map[uint]bool{results[0].ID: true, results[1].ID: true}
		if !firstTwoIDs[skill1.ID] || !firstTwoIDs[skill3.ID] {
			t.Errorf("前两个结果应该包含skill1和skill3")
		}

		// skill2应该排在最后
		if results[2].ID != skill2.ID {
			t.Errorf("最后一个结果应该是skill2，实际是%d", results[2].ID)
		}
	})
}

// TestSearchSkillsByTokens_EmptyKeyword 测试空关键词处理
func TestSearchSkillsByTokens_EmptyKeyword(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试技能
	createTestSkill(t, repo, "skill-1", "Description 1")
	createTestSkill(t, repo, "skill-2", "Description 2")

	// 测试空关键词应该返回所有技能
	t.Run("空关键词返回所有技能", func(t *testing.T) {
		results, err := repo.SearchSkillsByTokens(ctx, "")
		if err != nil {
			t.Fatalf("搜索失败: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("期望返回2个结果，实际返回%d个", len(results))
		}
	})

	// 测试纯空格关键词
	t.Run("纯空格关键词", func(t *testing.T) {
		results, err := repo.SearchSkillsByTokens(ctx, "   ")
		if err != nil {
			t.Fatalf("搜索失败: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("期望返回2个结果，实际返回%d个", len(results))
		}
	})
}

// TestSkillMenuTool_WithKeyword 测试技能菜单工具的分词搜索功能
func TestSkillMenuTool_WithKeyword(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	// 设置全局repo变量
	originalRepo := repo
	setRepoForTest(repo)
	defer func() {
		setRepoForTest(originalRepo)
	}()

	ctx := context.Background()

	// 创建测试技能
	createTestSkill(t, repo, "pdf-extractor", "Extract text from PDF documents")
	createTestSkill(t, repo, "image-processor", "Process and optimize images")

	// 测试带关键词的搜索
	t.Run("工具关键词搜索", func(t *testing.T) {
		request := mcp.CallToolRequest{}
		request.Params.Arguments = map[string]interface{}{
			"tag":     "",
			"keyword": "pdf",
		}

		result, err := skillMenuTool(ctx, request)
		if err != nil {
			t.Fatalf("工具调用失败: %v", err)
		}

		if len(result.Content) == 0 {
			t.Error("期望返回内容，但实际为空")
			return
		}

		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Error("期望返回文本内容")
			return
		}

		// 验证返回内容包含关键词搜索结果标识
		if textContent.Text == "" {
			t.Error("返回的文本内容为空")
		}
	})
}

// TestSkillMenuTool_WithoutRepo 测试仓库未初始化时的处理
func TestSkillMenuTool_WithoutRepo(t *testing.T) {
	// 临时保存原repo
	originalRepo := repo
	repo = nil
	defer func() {
		repo = originalRepo
	}()

	ctx := context.Background()

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"tag": "",
	}

	result, err := skillMenuTool(ctx, request)
	if err != nil {
		t.Fatalf("工具调用失败: %v", err)
	}

	if len(result.Content) == 0 {
		t.Error("期望返回内容，但实际为空")
		return
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Error("期望返回文本内容")
		return
	}

	expectedMsg := "数据库未初始化，无法获取技能列表"
	if textContent.Text != expectedMsg {
		t.Errorf("期望返回'%s'，实际返回'%s'", expectedMsg, textContent.Text)
	}
}

// TestFormatSkillList 测试技能列表格式化函数
func TestFormatSkillList(t *testing.T) {
	tests := []struct {
		name     string
		skills   []models.Skill
		title    string
		expected string
	}{
		{
			name:     "空技能列表",
			skills:   []models.Skill{},
			title:    "测试标题：",
			expected: "未找到匹配的技能",
		},
		{
			name: "单个技能",
			skills: []models.Skill{
				{Name: "test-skill", Description: "Test description"},
			},
			title:    "技能列表：",
			expected: "技能列表：\nname: test-skill description: Test description\n请调用 skill_detail 查技能详情",
		},
		{
			name: "多个技能",
			skills: []models.Skill{
				{Name: "skill-1", Description: "Description 1"},
				{Name: "skill-2", Description: "Description 2"},
			},
			title:    "搜索结果：",
			expected: "搜索结果：\nname: skill-1 description: Description 1\nname: skill-2 description: Description 2\n请调用 skill_detail 查技能详情",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSkillList(tt.skills, tt.title, 10)
			if result != tt.expected {
				t.Errorf("期望返回'%s'，实际返回'%s'", tt.expected, result)
			}
		})
	}
}

// TestSearchSkillsByTokens_CaseInsensitive 测试分词搜索大小写不敏感
func TestSearchSkillsByTokens_CaseInsensitive(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// 创建测试技能
	skill := createTestSkill(t, repo, "PDF-Parser", "Extract TEXT from PDF Documents")

	// 测试不同大小写的搜索
	testCases := []string{"pdf", "PDF", "Pdf", "text", "TEXT", "Text"}

	for _, keyword := range testCases {
		t.Run("搜索关键词"+keyword, func(t *testing.T) {
			results, err := repo.SearchSkillsByTokens(ctx, keyword)
			if err != nil {
				t.Fatalf("搜索失败: %v", err)
			}

			found := false
			for _, r := range results {
				if r.ID == skill.ID {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("搜索'%s'应该能找到技能%d", keyword, skill.ID)
			}
		})
	}
}

// TestSearchSkillsByTokens_MixedContent 测试混合中英文内容的分词搜索
func TestSearchSkillsByTokens_MixedContent(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// 创建混合中英文的测试技能
	skill1 := createTestSkill(t, repo, "pdf-tool", "PDF工具：用于处理PDF文件的实用工具")
	skill2 := createTestSkill(t, repo, "excel-tool", "Excel工具：用于处理Excel表格的工具")
	skill3 := createTestSkill(t, repo, "pdf-excel-converter", "PDF和Excel互转工具")

	// 测试搜索"工具"
	t.Run("搜索中文工具", func(t *testing.T) {
		results, err := repo.SearchSkillsByTokens(ctx, "工具")
		if err != nil {
			t.Fatalf("搜索失败: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("期望返回3个结果，实际返回%d个", len(results))
		}

		resultIDs := make(map[uint]bool)
		for _, s := range results {
			resultIDs[s.ID] = true
		}

		if !resultIDs[skill1.ID] || !resultIDs[skill2.ID] || !resultIDs[skill3.ID] {
			t.Error("结果应该包含所有三个技能")
		}
	})

	// 测试搜索"PDF"
	t.Run("搜索英文PDF", func(t *testing.T) {
		results, err := repo.SearchSkillsByTokens(ctx, "PDF")
		if err != nil {
			t.Fatalf("搜索失败: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("期望返回2个结果，实际返回%d个", len(results))
		}

		resultIDs := make(map[uint]bool)
		for _, s := range results {
			resultIDs[s.ID] = true
		}

		if !resultIDs[skill1.ID] || !resultIDs[skill3.ID] {
			t.Error("结果应该包含skill1和skill3")
		}
	})
}
