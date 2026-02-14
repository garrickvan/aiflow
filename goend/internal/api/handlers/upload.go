package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aiflow/internal/api/helpers"
	"aiflow/internal/errors"
	"aiflow/internal/models"
	"aiflow/internal/repositories"

	"gopkg.in/yaml.v3"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	repo *repositories.Repository
}

// NewUploadHandler 创建文件上传处理器
func NewUploadHandler(repo *repositories.Repository) *UploadHandler {
	return &UploadHandler{repo: repo}
}

// UploadData 处理文件上传
func (h *UploadHandler) UploadData(w http.ResponseWriter, req *http.Request) {
	// 解析表单数据
	const maxUploadSize = 50 << 20 // 50MB限制
	err := req.ParseMultipartForm(maxUploadSize)
	if err != nil {
		helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeBadRequest, "解析表单数据失败", err))
		return
	}

	// 获取process_type参数
	processType := req.FormValue("process_type")
	if processType == "" {
		processType = "import_skill" // 默认值
	}

	// 获取上传的文件
	file, fileHeader, err := req.FormFile("file")
	if err != nil {
		helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeBadRequest, "获取上传文件失败", err))
		return
	}
	defer file.Close()

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".md" && ext != ".zip" {
		helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeBadRequest, "只支持上传.md和.zip格式的文件", nil))
		return
	}

	// 创建upload_data目录（如果不存在）
	uploadDir := "upload_data"
	if _, err = os.Stat(uploadDir); os.IsNotExist(err) {
		const dirPermission = 0755
		if err = os.MkdirAll(uploadDir, dirPermission); err != nil {
			helpers.RenderError(w, req, errors.NewInternalError(errors.ErrCodeInternalError, "创建上传目录失败", err))
			return
		}
	}

	// 生成保存文件的路径，先检查文件是否存在，如果存在则添加时间戳
	fileName := filepath.Join(uploadDir, fileHeader.Filename)
	if _, err = os.Stat(fileName); err == nil {
		// 文件存在，添加时间戳
		fileName = filepath.Join(uploadDir, fmt.Sprintf("%s_%d%s", strings.TrimSuffix(fileHeader.Filename, ext), time.Now().Unix(), ext))
	}

	// 创建目标文件
	dst, err := os.Create(fileName)
	if err != nil {
		helpers.RenderError(w, req, errors.NewInternalError(errors.ErrCodeInternalError, "创建文件失败", err))
		return
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, file); err != nil {
		helpers.RenderError(w, req, errors.NewInternalError(errors.ErrCodeInternalError, "保存文件失败", err))
		return
	}

	switch processType {
	case "import_skill":
		switch ext {
		case ".md":
			// 处理技能导入
			err = h.handleSkillImport(fileName)
			if err != nil {
				helpers.RenderError(w, req, errors.NewInternalError(errors.ErrCodeSkillCreate, err.Error(), err))
				return
			}
		// case ".zip":
		// 处理技能分组导入
		default:
			helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeBadRequest, "无效的文件类型", nil))
			return
		}
	default:
		helpers.RenderError(w, req, errors.NewInvalidParamError(errors.ErrCodeBadRequest, fmt.Sprintf("无效的处理类型：%s", processType), nil))
		return
	}

	// 返回成功响应
	helpers.RenderSuccessWithMessage(w, req, "文件上传成功", map[string]interface{}{
		"filename":    fileHeader.Filename,
		"size":        fileHeader.Size,
		"processType": processType,
		"path":        fileName,
	})
}

// handleSkillImport 处理技能导入
// 解析.md文件内容，提取YAML头部信息并保存到数据库
func (h *UploadHandler) handleSkillImport(fileName string) error {
	// 1. 读取.md文件内容
	fileContent, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	// 2. 提取YAML头部信息
	yamlHeader, content := extractYAMLHeader(string(fileContent))
	if yamlHeader == "" {
		return fmt.Errorf("文件中未找到YAML头部信息，请确保文件以---开头和结束")
	}

	// 3. 解析YAML头部信息（支持多种命名风格）
	skillData, err := parseSkillYAML(yamlHeader)
	if err != nil {
		return err
	}

	// 4. 验证必填字段
	if strings.TrimSpace(skillData.Name) == "" {
		return fmt.Errorf("技能名称(name)不能为空")
	}
	if strings.TrimSpace(skillData.Description) == "" {
		return fmt.Errorf("技能描述(description)不能为空")
	}

	// 5. 构建Skill对象
	skill := &models.Skill{
		Name:          strings.TrimSpace(skillData.Name),
		ResourceDir:   strings.TrimSpace(skillData.ResourceDir),
		Description:   strings.TrimSpace(skillData.Description),
		License:       skillData.License,
		Compatibility: skillData.Compatibility,
		Metadata:      skillData.Metadata,
		AllowedTools:  skillData.AllowedTools,
		Detail:        content,
		CreatedAt:     time.Now().UnixMilli(),
		UpdatedAt:     time.Now().UnixMilli(),
	}

	// 如果资源目录为空，使用名称生成
	if skill.ResourceDir == "" {
		skill.ResourceDir = generateResourceDir(skill.Name)
	}

	// 6. 保存技能到数据库（同名则更新）
	existingSkill, err := h.repo.GetSkillByName(context.Background(), skill.Name)
	if err == nil && existingSkill != nil {
		// 技能已存在，更新
		skill.ID = existingSkill.ID
		skill.CreatedAt = existingSkill.CreatedAt
		err = h.repo.UpdateSkill(context.Background(), skill)
		if err != nil {
			return fmt.Errorf("更新技能到数据库失败: %v", err)
		}
	} else {
		// 技能不存在，创建新技能
		err = h.repo.CreateSkill(context.Background(), skill)
		if err != nil {
			return fmt.Errorf("保存技能到数据库失败: %v", err)
		}
	}

	// 7. 处理标签关联
	if len(skillData.Tags) > 0 {
		// 如果是更新操作，先清除旧标签关联
		if existingSkill != nil {
			oldTags, _ := h.repo.GetTagsBySkillID(context.Background(), skill.ID)
			for _, oldTag := range oldTags {
				h.repo.RemoveTagFromSkill(context.Background(), skill.ID, oldTag.ID)
			}
		}
		// 添加新标签关联
		for _, tagName := range skillData.Tags {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}
			// 查找或创建标签
			tag, err := h.repo.GetTagByName(context.Background(), tagName)
			if err != nil {
				// 标签不存在，创建新标签
				tag = &models.Tag{
					Name:      tagName,
					CreatedAt: time.Now().UnixMilli(),
					UpdatedAt: time.Now().UnixMilli(),
				}
				err = h.repo.CreateTag(context.Background(), tag)
				if err != nil {
					return fmt.Errorf("创建标签'%s'失败: %v", tagName, err)
				}
			}
			// 关联标签到技能
			err = h.repo.AddTagToSkill(context.Background(), skill.ID, tag.ID)
			if err != nil {
				return fmt.Errorf("关联标签'%s'到技能失败: %v", tagName, err)
			}
		}
	}

	return nil
}

// skillYAMLData 技能YAML数据结构（支持多种字段命名）
type skillYAMLData struct {
	Name          string   `yaml:"name"`
	ResourceDir   string   `yaml:"resource_dir"`
	ResourceDir2  string   `yaml:"resourceDir"`
	ResourceDir3  string   `yaml:"resource-dir"`
	Description   string   `yaml:"description"`
	License       string   `yaml:"license"`
	Compatibility string   `yaml:"compatibility"`
	Metadata      string   `yaml:"metadata"`
	AllowedTools  string   `yaml:"allowed_tools"`
	AllowedTools2 string   `yaml:"allowedTools"`
	AllowedTools3 string   `yaml:"allowed-tools"`
	Tags          []string `yaml:"tags"`
}

// parseSkillYAML 解析技能YAML，支持多种字段命名风格
func parseSkillYAML(yamlHeader string) (*skillYAMLData, error) {
	var data skillYAMLData
	err := yaml.Unmarshal([]byte(yamlHeader), &data)
	if err != nil {
		return nil, fmt.Errorf("YAML格式错误: %v", err)
	}

	// 合并多种命名风格的字段
	if data.ResourceDir == "" && data.ResourceDir2 != "" {
		data.ResourceDir = data.ResourceDir2
	}
	if data.ResourceDir == "" && data.ResourceDir3 != "" {
		data.ResourceDir = data.ResourceDir3
	}
	if data.AllowedTools == "" && data.AllowedTools2 != "" {
		data.AllowedTools = data.AllowedTools2
	}
	if data.AllowedTools == "" && data.AllowedTools3 != "" {
		data.AllowedTools = data.AllowedTools3
	}

	return &data, nil
}

// generateResourceDir 根据技能名称生成资源目录
func generateResourceDir(name string) string {
	// 将名称转换为小写，替换空格和特殊字符为下划线
	dir := strings.ToLower(name)
	dir = strings.ReplaceAll(dir, " ", "_")
	dir = strings.ReplaceAll(dir, "-", "_")
	// 移除非字母数字下划线的字符
	var result strings.Builder
	for _, r := range dir {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// extractYAMLHeader 提取MD文件中的YAML头部信息
// 返回YAML头部内容和剩余的文件内容
func extractYAMLHeader(content string) (string, string) {
	lines := strings.Split(content, "\n")
	var yamlLines []string
	inYAML := false

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "---" {
			if !inYAML {
				// 开始YAML头部
				inYAML = true
			} else {
				// 结束YAML头部
				inYAML = false
				// 拼接YAML头部内容
				yamlHeader := strings.Join(yamlLines, "\n")
				// 拼接剩余内容
				remainingContent := strings.Join(lines[i+1:], "\n")
				return yamlHeader, remainingContent
			}
		} else if inYAML {
			yamlLines = append(yamlLines, line)
		}
	}

	return "", content
}
