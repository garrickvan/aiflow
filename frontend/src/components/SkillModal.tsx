/**
 * 技能查看/编辑弹窗组件
 * 参考ui-skill-edtior.html设计，采用左右分栏布局
 * 左侧：元数据配置区 | 右侧：Markdown编辑器区
 * 使用Zustand状态管理
 */
import React, { useEffect, useState, useRef } from "react";
import { Modal, Form, Input, Button, Select, AutoComplete, message, Tooltip } from "antd";
import ReactMarkdown from "react-markdown";
import { skillApi } from "../services";
import {
  EditOutlined,
  EyeOutlined,
  SaveOutlined,
  CloseOutlined,
  DeleteOutlined,
  PlusOutlined,
  FullscreenOutlined,
  FullscreenExitOutlined,
  LockOutlined,
  CheckCircleOutlined,
  CodeOutlined,
  FileTextOutlined,
  BoldOutlined,
  ItalicOutlined,
  UnorderedListOutlined,
  FontSizeOutlined,

  BarsOutlined,
} from "@ant-design/icons";
import { useAppStore } from "../stores/appStore";
import { useModalStore, ModalType, type ModalDataMap } from "../stores/modalStore";

const { TextArea } = Input;

/**
 * 获取技能头像样式
 */
const getSkillAvatarConfig = (
  name: string,
): { gradient: string; icon: string } => {
  const firstChar = (name || '').charAt(0).toUpperCase();
  const gradients = [
    "linear-gradient(135deg, #8b5cf6, #3b82f6)",
    "linear-gradient(135deg, #10b981, #3b82f6)",
    "linear-gradient(135deg, #f59e0b, #ef4444)",
    "linear-gradient(135deg, #ec4899, #8b5cf6)",
    "linear-gradient(135deg, #06b6d4, #3b82f6)",
    "linear-gradient(135deg, #84cc16, #10b981)",
  ];
  const index = name.length % gradients.length;
  return {
    gradient: gradients[index],
    icon: firstChar,
  };
};

/**
 * 格式化时间戳
 */
const formatTime = (timestamp: number): string => {
  const date = new Date(timestamp);
  return date.toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
};

/**
 * 解析授权工具列表
 */
const parseAllowedTools = (allowedTools: string): string[] => {
  if (!allowedTools) return [];
  return allowedTools
    .split(/[,\s]+/)
    .map((t) => t.trim())
    .filter((t) => t);
};

/**
 * 编辑器视图模式
 */
type EditorViewMode = "edit" | "split" | "preview";

/**
 * 技能查看/编辑弹窗组件
 */
const SkillModal: React.FC = () => {
  const [form] = Form.useForm();
  const [viewMode, setViewMode] = useState<EditorViewMode>("edit");
  const [toolInput, setToolInput] = useState<string>("");
  const [tools, setTools] = useState<string[]>([]);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [detailContent, setDetailContent] = useState<string>("");
  const [cursorPos, setCursorPos] = useState({ line: 1, col: 1 });
  const [docStats, setDocStats] = useState({ chars: 0, lines: 0 });
  const detailTextareaRef = useRef<HTMLTextAreaElement>(null);

  // 从Zustand Store获取状态和actions
  const { tags, setSkills } = useAppStore();
  const { currentModal, modalData, closeSkillModal } = useModalStore();

  // 计算当前模态框状态和编辑数据
  const isSkillModalOpen = currentModal === ModalType.SKILL;
  const editingSkill = (modalData as ModalDataMap[typeof ModalType.SKILL]) ?? null;

  const isEditing = !!editingSkill;

  // 当编辑的技能变化时，重置表单
  useEffect(() => {
    if (isSkillModalOpen) {
      if (editingSkill) {
        const tagIds = editingSkill.tags?.map((tag: { id: number }) => tag.id) || [];
        setTools(parseAllowedTools(editingSkill.allowedTools));
        setDetailContent(editingSkill.detail || "");
        // 使用setTimeout确保Form组件已挂载
        setTimeout(() => {
          form.setFieldsValue({
            ...editingSkill,
            tags: tagIds,
            description: editingSkill.description || "",
            metadata: editingSkill.metadata || "",
          });
          updateDocStats(editingSkill.detail || "");
        }, 0);
      } else {
        setTools([]);
        setDetailContent("");
        setViewMode("edit");
        // 使用setTimeout确保Form组件已挂载
        setTimeout(() => {
          form.resetFields();
        }, 0);
      }
      setToolInput("");
    }
  }, [isSkillModalOpen, editingSkill, form]);

  // 更新文档统计
  const updateDocStats = (text: string) => {
    const lines = text.split("\n").length;
    const chars = text.length;
    setDocStats({ chars, lines });
  };

  // 处理表单提交
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      const submitData = {
        ...values,
        allowedTools: tools.join(" "),
        detail: detailContent,
        description: values.description?.trim(),
      };

      if (editingSkill) {
        // 更新技能
        const updatedSkill = await skillApi.updateSkill(editingSkill.id, {
          name: submitData.name,
          resourceDir: submitData.resourceDir,
          description: submitData.description,
          version: submitData.version,
          detail: submitData.detail,
          license: submitData.license,
          compatibility: submitData.compatibility,
          metadata: submitData.metadata,
          allowedTools: submitData.allowedTools,
          tags: submitData.tags,
        });
        // 使用函数式更新避免闭包问题，确保基于最新状态更新
        setSkills(
          (prevSkills) =>
            prevSkills.map((skill) =>
              skill.id === updatedSkill.id ? updatedSkill : skill,
            ),
        );
        message.success("技能更新成功");
      } else {
        // 新增技能
        const newSkill = await skillApi.createSkill({
          name: submitData.name,
          resourceDir: submitData.resourceDir,
          description: submitData.description,
          version: submitData.version,
          detail: submitData.detail,
          license: submitData.license,
          compatibility: submitData.compatibility,
          metadata: submitData.metadata,
          allowedTools: submitData.allowedTools,
          tags: submitData.tags,
        });
        // 使用函数式更新避免闭包问题
        setSkills((prevSkills) => [...prevSkills, newSkill]);
        message.success("技能创建成功");
      }
      closeSkillModal();
    } catch (error) {
      console.error("表单验证失败:", error);
    }
  };

  // 处理取消
  const handleCancel = () => {
    form.resetFields();
    setViewMode("edit");
    setTools([]);
    setDetailContent("");
    closeSkillModal();
  };

  // 处理删除
  const handleDelete = async () => {
    if (editingSkill) {
      try {
        await skillApi.deleteSkill(editingSkill.id);
        // 使用函数式更新避免闭包问题，确保基于最新状态更新
        setSkills((prevSkills) =>
          prevSkills.filter((skill) => skill.id !== editingSkill.id),
        );
        message.success("技能已移至回收站");
        closeSkillModal();
      } catch (error) {
        message.error("删除失败，请重试");
        console.error("Skill delete failed:", error);
      }
    }
  };

  // 添加工具
  const handleAddTool = () => {
    if (toolInput.trim() && !tools.includes(toolInput.trim())) {
      setTools([...tools, toolInput.trim()]);
      setToolInput("");
    }
  };

  // 移除工具
  const handleRemoveTool = (tool: string) => {
    setTools(tools.filter((t) => t !== tool));
  };

  // 切换全屏
  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen);
  };

  // 处理编辑器Tab键
  const handleEditorKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Tab") {
      e.preventDefault();
      const target = e.target as HTMLTextAreaElement;
      const start = target.selectionStart;
      const end = target.selectionEnd;
      const newValue =
        detailContent.substring(0, start) + "  " + detailContent.substring(end);
      setDetailContent(newValue);
      updateDocStats(newValue);
      setTimeout(() => {
        target.selectionStart = target.selectionEnd = start + 2;
      }, 0);
    }
  };

  // 更新光标位置
  const handleEditorKeyUp = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    const target = e.target as HTMLTextAreaElement;
    const pos = target.selectionStart;
    const val = target.value;
    const line = val.substr(0, pos).split("\n").length;
    const col = pos - val.lastIndexOf("\n", pos - 1);
    setCursorPos({ line, col });
  };

  // 处理编辑器内容变化
  const handleDetailChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    setDetailContent(value);
    updateDocStats(value);
  };

  // 插入Markdown标记
  const insertMarkdown = (before: string, after: string) => {
    const textarea = detailTextareaRef.current;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selected = detailContent.substring(start, end);
    const newValue =
      detailContent.substring(0, start) +
      before +
      selected +
      after +
      detailContent.substring(end);

    setDetailContent(newValue);
    updateDocStats(newValue);

    setTimeout(() => {
      textarea.selectionStart = start + before.length;
      textarea.selectionEnd = end + before.length;
      textarea.focus();
    }, 0);
  };

  // 获取标签颜色
  const getTagColor = (index: number) => {
    const colors = [
      { bg: "#eff6ff", text: "#1d4ed8", border: "#bfdbfe" },
      { bg: "#f5f3ff", text: "#7c3aed", border: "#ddd6fe" },
      { bg: "#f0fdf4", text: "#15803d", border: "#bbf7d0" },
      { bg: "#fff7ed", text: "#c2410c", border: "#fed7aa" },
      { bg: "#fdf2f8", text: "#be185d", border: "#fbcfe8" },
      { bg: "#eef2ff", text: "#4338ca", border: "#c7d2fe" },
    ];
    return colors[index % colors.length];
  };

  // 渲染头部
  const renderHeader = () => {
    // 只有当技能模态框打开且没有编辑数据时才显示新建界面
    if (!isSkillModalOpen || !editingSkill) {
      return (
        <div className="skill-modal-header">
          <div className="skill-modal-header-left">
            <div
              className="skill-avatar-large"
              style={{
                background: "linear-gradient(135deg, #3b82f6, #6366f1)",
              }}
            >
              <PlusOutlined />
            </div>
            <div className="skill-header-info">
              <h2 className="skill-header-title">新建技能</h2>
              <span className="skill-header-subtitle">创建一个新的技能</span>
            </div>
          </div>
          <div className="skill-modal-header-right">
            <Button
              type="text"
              onClick={toggleFullscreen}
              className="header-btn"
            >
              {isFullscreen ? (
                <FullscreenExitOutlined />
              ) : (
                <FullscreenOutlined />
              )}
            </Button>
            <Button
              type="text"
              onClick={handleCancel}
              className="header-btn close-btn"
            >
              <CloseOutlined />
            </Button>
          </div>
        </div>
      );
    }

    const avatarConfig = getSkillAvatarConfig(editingSkill.name);

    return (
      <div className="skill-modal-header">
        <div className="skill-modal-header-left">
          <div
            className="skill-avatar-large"
            style={{ background: avatarConfig.gradient }}
          >
            {avatarConfig.icon}
          </div>
          <div className="skill-header-info">
            <div className="skill-header-title-row">
              <h2 className="skill-header-title">{editingSkill.name}</h2>
              <span className="skill-status-badge success">
                <CheckCircleOutlined /> 已发布
              </span>
            </div>
            <div className="skill-header-meta">
              <span className="skill-id">ID: {editingSkill.id}</span>
              <span className="separator">•</span>
              <span>更新于 {formatTime(editingSkill.updatedAt)}</span>
            </div>
          </div>
        </div>
        <div className="skill-modal-header-right">
          <Button type="text" onClick={toggleFullscreen} className="header-btn">
            {isFullscreen ? <FullscreenExitOutlined /> : <FullscreenOutlined />}
          </Button>
          <Button
            type="text"
            onClick={handleCancel}
            className="header-btn close-btn"
          >
            <CloseOutlined />
          </Button>
        </div>
      </div>
    );
  };

  // 渲染左侧元数据面板
  const renderLeftPanel = () => {
    const selectedTagIds = Form.useWatch("tags", form) || [];
    const selectedTags = tags.filter((tag) => selectedTagIds.includes(tag.id));

    return (
      <div className="skill-modal-left-panel">
        <Form form={form} layout="vertical" className="skill-form">
          {/* 基础标识区 */}
          <div className="form-section">
            <div className="section-header">
              <span className="section-title">基础标识</span>
              <LockOutlined
                className="section-lock"
                title="系统生成，不可修改"
              />
            </div>

            <Form.Item
              name="name"
              label="技能名称"
              rules={[
                { required: true, message: "请输入技能名称" },
                {
                  min: 1,
                  max: 128,
                  message: "技能名称长度应在 1-128 字符之间",
                },
              ]}
            >
              <Input
                placeholder="如: pdf-processing"
                autoComplete="off"
                readOnly={isEditing}
                className={isEditing ? "readonly-input" : ""}
                suffix={isEditing ? <LockOutlined /> : null}
              />
            </Form.Item>

            <Form.Item name="resourceDir" label="资源目录">
              <Input
                placeholder="如: pdf_processing"
                autoComplete="off"
                readOnly={isEditing}
                className={isEditing ? "readonly-input" : ""}
                prefix={<span className="input-prefix">/skills/</span>}
              />
            </Form.Item>
          </div>

          {/* 配置信息区 */}
          <div className="form-section">
            <div className="section-header">
              <span className="section-title">配置信息</span>
            </div>

            <div className="form-row-2">
              <Form.Item
                name="license"
                label="许可证"
                className="form-item-compact"
              >
                <AutoComplete
                  placeholder="选择或输入许可证"
                  allowClear
                  options={[
                    { value: "MIT" },
                    { value: "Apache-2.0" },
                    { value: "GPL-3.0" },
                    { value: "BSD" },
                    { value: "LGPL" },
                    { value: "MPL-2.0" },
                    { value: "Proprietary" },
                  ]}
                  filterOption={(inputValue, option) =>
                    option!.value!.toUpperCase().indexOf(inputValue.toUpperCase()) !== -1
                  }
                />
              </Form.Item>

              <Form.Item
                name="version"
                label="版本"
                className="form-item-compact"
              >
                <Input placeholder="如: 1.0.0" />
              </Form.Item>
            </div>

            <Form.Item name="compatibility" label="兼容性">
              <TextArea
                rows={2}
                placeholder="环境要求、依赖版本..."
                className="compact-textarea"
              />
            </Form.Item>

            <Form.Item
              name="description"
              label="功能描述"
              rules={[
                { required: true, message: "请输入描述" },
                { min: 1, max: 1024, message: "描述长度应在 1-1024 字符之间" },
              ]}
            >
              <TextArea
                rows={6}
                placeholder="描述功能 + 触发时机；第三人称；含关键词便于技能发现..."
                className="compact-textarea"
              />
            </Form.Item>

            <Form.Item label="允许工具">
              <div className="tools-input-container">
                {tools.map((tool, index) => (
                  <span key={index} className="tool-tag">
                    {tool}
                    <span
                      className="tool-remove"
                      onClick={() => handleRemoveTool(tool)}
                    >
                      ×
                    </span>
                  </span>
                ))}
                <div className="tool-input-wrapper">
                  <Input
                    value={toolInput}
                    onChange={(e) => setToolInput(e.target.value)}
                    onPressEnter={handleAddTool}
                    placeholder="输入工具按回车添加"
                    className="tool-input"
                    size="small"
                  />
                  <Button
                    type="text"
                    size="small"
                    onClick={handleAddTool}
                    className="tool-add-btn"
                  >
                    <PlusOutlined />
                  </Button>
                </div>
              </div>
            </Form.Item>
          </div>
          {/* 标签区 */}
          <div className="form-section">
            <div className="section-header">
              <span className="section-title">标签</span>
            </div>

            <div className="selected-tags-preview">
              {selectedTags.map((tag, index) => {
                const color = getTagColor(index);
                return (
                  <span
                    key={tag.id}
                    className="tag-badge-preview"
                    style={{
                      backgroundColor: color.bg,
                      color: color.text,
                      borderColor: color.border,
                    }}
                  >
                    {tag.name}
                  </span>
                );
              })}
            </div>

            <Form.Item name="tags" className="form-item-no-margin">
              <Select
                mode="multiple"
                placeholder="选择标签"
                maxTagCount={0}
                maxTagPlaceholder={() => null}
                options={tags.map((tag) => ({
                  value: tag.id,
                  label: tag.name,
                }))}
                className="tags-select"
              />
            </Form.Item>
          </div>

            <Form.Item
              name="metadata"
              label="MetaData"
              rules={[
                {
                  validator: (_, value) => {
                    if (!value || value.trim() === '') {
                      return Promise.resolve();
                    }
                    try {
                      JSON.parse(value);
                      return Promise.resolve();
                    } catch {
                      return Promise.reject(new Error('请输入有效的JSON格式'));
                    }
                  },
                },
              ]}
            >
              <TextArea
                rows={6}
                placeholder='{"author": "example-org", "version": "1.0"}'
                className="compact-textarea"
              />
            </Form.Item>

        </Form>
      </div>
    );
  };

  // 渲染右侧编辑器面板
  const renderRightPanel = () => {
    return (
      <div className="skill-modal-right-panel">
        {/* Markdown编辑器区 */}
        <div className="editor-section">
          {/* 编辑器工具栏 */}
          <div className="editor-toolbar">
            <div className="toolbar-left">
              <h3 className="editor-title">
                <FileTextOutlined className="editor-icon" />
                Detail 文档
                <span className="editor-badge">Markdown</span>
              </h3>
              <div className="toolbar-divider" />
              <div className="view-mode-tabs">
                <button
                  className={`view-mode-btn ${viewMode === "edit" ? "active" : ""}`}
                  onClick={() => setViewMode("edit")}
                >
                  <EditOutlined /> 编辑
                </button>
                <button
                  className={`view-mode-btn ${viewMode === "split" ? "active" : ""}`}
                  onClick={() => setViewMode("split")}
                >
                  <BarsOutlined /> 分屏
                </button>
                <button
                  className={`view-mode-btn ${viewMode === "preview" ? "active" : ""}`}
                  onClick={() => setViewMode("preview")}
                >
                  <EyeOutlined /> 预览
                </button>
              </div>
            </div>
            <div className="toolbar-right">
              <Tooltip title="粗体">
                <button
                  className="toolbar-btn"
                  onClick={() => insertMarkdown("**", "**")}
                >
                  <BoldOutlined />
                </button>
              </Tooltip>
              <Tooltip title="斜体">
                <button
                  className="toolbar-btn"
                  onClick={() => insertMarkdown("*", "*")}
                >
                  <ItalicOutlined />
                </button>
              </Tooltip>
              <Tooltip title="代码">
                <button
                  className="toolbar-btn"
                  onClick={() => insertMarkdown("`", "`")}
                >
                  <CodeOutlined />
                </button>
              </Tooltip>
              <Tooltip title="标题">
                <button
                  className="toolbar-btn"
                  onClick={() => insertMarkdown("## ", "")}
                >
                  <FontSizeOutlined />
                </button>
              </Tooltip>
              <Tooltip title="列表">
                <button
                  className="toolbar-btn"
                  onClick={() => insertMarkdown("- ", "")}
                >
                  <UnorderedListOutlined />
                </button>
              </Tooltip>
              <div className="toolbar-divider" />
              <Tooltip title="格式化">
                <button className="toolbar-btn" onClick={() => {}}>
                  <BarsOutlined />
                </button>
              </Tooltip>
            </div>
          </div>

          {/* 编辑器内容区 */}
          <div className={`editor-content ${viewMode}`}>
            {/* 编辑区 */}
            {(viewMode === "edit" || viewMode === "split") && (
              <div
                className={`editor-pane ${viewMode === "split" ? "half" : "full"}`}
              >
                <TextArea
                  ref={detailTextareaRef}
                  value={detailContent}
                  onChange={handleDetailChange}
                  onKeyDown={handleEditorKeyDown}
                  onKeyUp={handleEditorKeyUp}
                  className="markdown-editor"
                  placeholder="在此输入 Markdown 内容..."
                  spellCheck={false}
                />
              </div>
            )}

            {/* 分隔线 */}
            {viewMode === "split" && <div className="editor-resizer" />}

            {/* 预览区 */}
            {(viewMode === "preview" || viewMode === "split") && (
              <div
                className={`preview-pane ${viewMode === "split" ? "half" : "full"}`}
              >
                <div className="markdown-preview">
                  {detailContent ? (
                    <ReactMarkdown>{detailContent}</ReactMarkdown>
                  ) : (
                    <span className="preview-placeholder">
                      预览内容将显示在这里...
                    </span>
                  )}
                </div>
              </div>
            )}
          </div>

          {/* 编辑器状态栏 */}
          <div className="editor-statusbar">
            <div className="statusbar-left">
              <span className="status-item">
                <FileTextOutlined />
                {docStats.chars.toLocaleString()} 字符 | {docStats.lines} 行
              </span>
              <span className="status-divider" />
              <span className="status-item">
                Ln {cursorPos.line}, Col {cursorPos.col}
              </span>
            </div>
            <div className="statusbar-right">
              <span className="status-item success">
                <CheckCircleOutlined /> Markdown 有效
              </span>
              <span className="status-divider" />
              <span className="status-item">自动保存</span>
            </div>
          </div>
        </div>
      </div>
    );
  };

  // 渲染底部操作栏
  const renderFooter = () => {
    return (
      <div className="skill-modal-footer">
        <div className="footer-left">
          {isEditing && (
            <>
              <Button danger onClick={handleDelete} className="footer-btn">
                <DeleteOutlined /> 删除
              </Button>
            </>
          )}
        </div>
        <div className="footer-right">
          <Button onClick={handleCancel} className="footer-btn">
            取消
          </Button>
          <Button
            type="primary"
            onClick={handleSubmit}
            className="footer-btn-primary"
          >
            <SaveOutlined /> {isEditing ? "保存更改" : "创建技能"}
          </Button>
        </div>
      </div>
    );
  };

  return (
    <Modal
      open={isSkillModalOpen}
      onCancel={handleCancel}
      footer={null}
      width={isFullscreen ? "100vw" : 1200}
      className={`skill-editor-modal ${isFullscreen ? "fullscreen" : ""}`}
      centered={!isFullscreen}
      closeIcon={null}
      destroyOnClose
      style={{
        top: isFullscreen ? 0 : undefined,
        padding: 0,
        maxWidth: isFullscreen ? "100vw" : undefined,
      }}
      styles={{
        body: { padding: 0, height: isFullscreen ? "100vh" : "85vh" },
      }}
    >
      <div className="skill-editor-container">
        {renderHeader()}
        <div className="skill-editor-body">
          {renderLeftPanel()}
          {renderRightPanel()}
        </div>
        {renderFooter()}
      </div>
    </Modal>
  );
};

export default SkillModal;
