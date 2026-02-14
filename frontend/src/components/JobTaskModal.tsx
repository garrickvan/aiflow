/**
 * 任务查看/编辑弹窗组件
 * 参考ui-editor.html设计，采用左右分栏布局
 * 左侧任务信息，右侧执行记录时间轴
 */
import React, { useEffect, useState, useMemo, memo } from "react";
import { Modal, Form, Input, Select, Button, message } from "antd";
import type { ExecutionRecord } from "../types";
import {
  JOBTASK_STATUS_OPTIONS,
  JOBTASK_TYPE_OPTIONS,
  PROJECT_OPTIONS,
} from "../types";
import { jobtaskApi } from "../services";
import {
  EditOutlined,
  EyeOutlined,
  HistoryOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  SyncOutlined,
  ExclamationCircleOutlined,
  SafetyCertificateOutlined,
  SaveOutlined,
  CloseOutlined,
  FolderOpenOutlined,
  TagOutlined,
  BranchesOutlined,
  FileTextOutlined,
  FlagOutlined,
  BulbOutlined,
  CalendarOutlined,
  DownOutlined,
  ReloadOutlined,
  FileOutlined,
  ThunderboltFilled,
} from "@ant-design/icons";
import { useAppStore } from "../stores/appStore";
import { useModalStore, ModalType, type ModalDataMap } from "../stores/modalStore";

/**
 * 获取状态配置
 */
const getStatusConfig = (
  status: string,
): {
  color: string;
  icon: React.ReactNode;
  label: string;
  bgColor: string;
  borderColor: string;
  dotColor: string;
} => {
  const configMap: Record<
    string,
    {
      color: string;
      icon: React.ReactNode;
      label: string;
      bgColor: string;
      borderColor: string;
      dotColor: string;
    }
  > = {
    已创建: {
      color: "#64748b",
      icon: <ClockCircleOutlined />,
      label: "已创建",
      bgColor: "#f8fafc",
      borderColor: "#e2e8f0",
      dotColor: "#94a3b8",
    },
    处理中: {
      color: "#3b82f6",
      icon: <SyncOutlined spin />,
      label: "处理中",
      bgColor: "#eff6ff",
      borderColor: "#bfdbfe",
      dotColor: "#3b82f6",
    },
    处理失败: {
      color: "#ef4444",
      icon: <ExclamationCircleOutlined />,
      label: "处理失败",
      bgColor: "#fef2f2",
      borderColor: "#fecaca",
      dotColor: "#ef4444",
    },
    处理完成: {
      color: "#10b981",
      icon: <CheckCircleOutlined />,
      label: "处理完成",
      bgColor: "#f0fdf4",
      borderColor: "#bbf7d0",
      dotColor: "#10b981",
    },
    验收通过: {
      color: "#8b5cf6",
      icon: <SafetyCertificateOutlined />,
      label: "验收通过",
      bgColor: "#faf5ff",
      borderColor: "#e9d5ff",
      dotColor: "#8b5cf6",
    },
  };
  return (
    configMap[status] || {
      color: "#64748b",
      icon: <ClockCircleOutlined />,
      label: status,
      bgColor: "#f8fafc",
      borderColor: "#e2e8f0",
      dotColor: "#94a3b8",
    }
  );
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
 * 格式化日期时间（完整）
 */
const formatDateTime = (timestamp: number): string => {
  const date = new Date(timestamp);
  return date.toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
};

/**
 * 执行记录时间轴项组件 - 使用memo优化
 */
interface ExecutionTimelineItemProps {
  record: ExecutionRecord;
  index: number;
  isExpanded: boolean;
  isActive: boolean;
  isLast: boolean;
  onToggle: (sequence: number) => void;
}

const ExecutionTimelineItem = memo<ExecutionTimelineItemProps>(
  ({ record, isExpanded, isActive, isLast, onToggle }) => {
    const recordStatus = getStatusConfig(record.status || "已创建");

    return (
      <div
        key={record.sequence}
        className={`job-timeline-item ${isActive ? "active" : ""}`}
      >
        {/* 时间轴线 */}
        {!isLast && <div className="job-timeline-line" />}

        {/* 时间节点 */}
        <div
          className={`job-timeline-dot ${isActive ? "active" : ""}`}
          onClick={() => onToggle(record.sequence)}
        />

        {/* 执行记录卡片 */}
        <div
          className={`job-timeline-card ${isActive ? "active" : ""} ${isExpanded ? "expanded" : ""}`}
          onClick={() => onToggle(record.sequence)}
        >
          {/* 卡片头部 */}
          <div className="job-card-header">
            <div className="job-card-header-left">
              <div className={`job-sequence-badge ${isActive ? "active" : ""}`}>
                {String(record.sequence).padStart(2, "0")}
                {isActive && (
                  <div className="job-active-indicator">
                    <ThunderboltFilled />
                  </div>
                )}
              </div>
              <div className="job-card-info">
                <div className="job-card-tags">
                  <span
                    className="job-status-tag"
                    style={{
                      color: recordStatus.color,
                      background: recordStatus.bgColor,
                      borderColor: recordStatus.borderColor,
                    }}
                  >
                    {record.status || "已创建"}
                  </span>
                  {isActive && (
                    <span className="job-active-tag">
                      <ThunderboltFilled /> 当前活跃
                    </span>
                  )}
                  <span className="job-time-tag">
                    <CalendarOutlined />{" "}
                    {formatTime(record.createdAt || Date.now())}
                  </span>
                </div>
                <p className="job-card-title">{record.result || "执行记录"}</p>
              </div>
            </div>
            <DownOutlined
              className={`job-card-chevron ${isExpanded ? "expanded" : ""}`}
            />
          </div>

          {/* 折叠内容 */}
          {isExpanded && (
            <div className="job-card-content">
              {/* 解决方案 */}
              {record.solution && (
                <div className="job-card-section">
                  <label className="job-card-section-label">
                    <BulbOutlined /> 解决方案
                  </label>
                  <div className="job-solution-block">{record.solution}</div>
                </div>
              )}

              {/* 执行结果 */}
              {record.result && (
                <div className="job-card-section">
                  <label className="job-card-section-label">
                    <FileTextOutlined /> 执行结果
                  </label>
                  <div className="job-code-block">{record.result}</div>
                </div>
              )}

              {/* 关联文件 */}
              {record.relatedFiles && record.relatedFiles.length > 0 && (
                <div className="job-card-section">
                  <label className="job-card-section-label">
                    <FileTextOutlined /> 关联文件
                  </label>
                  <div className="job-file-list">
                    {record.relatedFiles.map((file, idx) => (
                      <div key={idx} className="job-file-tag">
                        <FileOutlined className="job-file-icon" />
                        <span className="job-file-name">{file}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* 验收标准 */}
              {record.acceptStd && (
                <div className="job-card-section">
                  <label className="job-card-section-label">
                    <CheckCircleOutlined /> 验收标准
                  </label>
                  <div className="job-accept-block">
                    <div className="job-accept-indicator">
                      <div className="job-accept-dot" />
                    </div>
                    <p>{record.acceptStd}</p>
                  </div>
                </div>
              )}

              {/* 使用技能 */}
              {record.skills && record.skills.length > 0 && (
                <div className="job-card-section">
                  <label className="job-card-section-label">
                    <ThunderboltFilled /> 使用技能
                  </label>
                  <div className="job-skill-list">
                    {record.skills.map((skill, idx) => (
                      <div key={idx} className="job-skill-tag">
                        <span className="job-skill-name">{skill}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    );
  },
);

ExecutionTimelineItem.displayName = "ExecutionTimelineItem";

/**
 * 验收通过复选框包装组件
 * 用于根据checked状态动态切换样式
 * 支持Form.Item的valuePropName="checked"传值
 */
const CheckboxWrapper: React.FC<{
  checked?: boolean;
  value?: boolean;
  onChange?: (checked: boolean) => void;
}> = ({ checked, value, onChange }) => {
  // 兼容Form.Item的valuePropName="checked"和普通value传值
  const isChecked = checked !== undefined ? checked : value;
  return (
    <label className={`job-accept-checkbox ${!isChecked ? "unchecked" : ""}`}>
      <input
        type="checkbox"
        className="job-checkbox-input"
        checked={isChecked}
        onChange={(e) => onChange?.(e.target.checked)}
      />
      <span className="job-checkbox-text">验收通过</span>
      <SafetyCertificateOutlined className="job-checkbox-icon" />
    </label>
  );
};

/**
 * 解析执行记录
 */
const parseExecutionRecords = (
  executionRecords: string | undefined,
): ExecutionRecord[] => {
  if (!executionRecords) return [];
  try {
    const records: ExecutionRecord[] = JSON.parse(executionRecords);
    return Array.isArray(records) ? records : [];
  } catch {
    return [];
  }
};

/**
 * 任务查看/编辑弹窗组件
 */
const JobTaskModal: React.FC = () => {
  const [form] = Form.useForm();
  const [goalCount, setGoalCount] = useState(0);
  const [expandedRecords, setExpandedRecords] = useState<Set<number>>(
    new Set([1]),
  );
  const [isWideMode, setIsWideMode] = useState(false);

  // 从Zustand Store获取状态和actions
  const { jobTasks, setJobTasks } = useAppStore();
  const { currentModal, modalData, closeJobTaskModal } = useModalStore();

  // 计算当前模态框状态和编辑数据
  const isJobTaskModalOpen = currentModal === ModalType.JOB_TASK;
  const editingJobTask = (modalData as ModalDataMap[typeof ModalType.JOB_TASK]) ?? null;

  const isEditing = !!editingJobTask;

  // 当编辑的任务变化时，重置表单
  useEffect(() => {
    if (isJobTaskModalOpen) {
      if (editingJobTask) {
        const goal = editingJobTask.goal || "";
        setGoalCount(goal.length);
        // 默认展开第一条执行记录
        setExpandedRecords(
          new Set([editingJobTask.activeExecutionSequence || 1]),
        );
        setTimeout(() => {
          form.setFieldsValue({
            jobNo: editingJobTask.jobNo,
            project: editingJobTask.project,
            type: editingJobTask.type,
            goal: editingJobTask.goal,
            status: editingJobTask.status,
            passAcceptStd: editingJobTask.passAcceptStd,
          });
        }, 0);
      } else {
        setGoalCount(0);
        setTimeout(() => {
          form.resetFields();
          form.setFieldsValue({
            status: "已创建",
            type: "新需求",
            passAcceptStd: false,
          });
        }, 0);
      }
    }
  }, [isJobTaskModalOpen, editingJobTask, form]);

  // 处理表单提交
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();

      if (editingJobTask) {
        // 编辑时传递原project字段，后端不允许修改项目
        const updatedJobTask = await jobtaskApi.updateJobTask(
          editingJobTask.id,
          {
            jobNo: values.jobNo,
            project: editingJobTask.project,
            type: values.type,
            goal: values.goal,
            status: values.status,
            passAcceptStd: values.passAcceptStd,
          },
        );
        setJobTasks(
          jobTasks.map((jt) =>
            jt.id === updatedJobTask.id ? updatedJobTask : jt,
          ),
        );
        message.success("任务更新成功");
      } else {
        const newJobTask = await jobtaskApi.createJobTask({
          jobNo: values.jobNo,
          project: values.project,
          type: values.type,
          goal: values.goal,
          status: values.status,
          passAcceptStd: false,
        });
        setJobTasks([...jobTasks, newJobTask]);
        message.success("任务创建成功");
      }
      closeJobTaskModal();
    } catch (error) {
      console.error("表单验证失败:", error);
    }
  };

  // 处理取消
  const handleCancel = () => {
    form.resetFields();
    setExpandedRecords(new Set());
    setIsWideMode(false);
    closeJobTaskModal();
  };

  // 切换执行记录展开状态
  const toggleAccordion = (sequence: number) => {
    setExpandedRecords((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(sequence)) {
        newSet.delete(sequence);
      } else {
        newSet.add(sequence);
      }
      return newSet;
    });
  };

  // 切换宽屏模式
  const toggleWideMode = () => {
    setIsWideMode(!isWideMode);
  };

  // 获取当前状态值
  const currentStatus = form.getFieldValue("status") || "已创建";

  // 解析执行记录 - 使用useMemo缓存
  const executionRecords = useMemo(
    () => parseExecutionRecords(editingJobTask?.executionRecords),
    [editingJobTask?.executionRecords],
  );

  // 缓存状态配置
  const statusConfig = useMemo(
    () => getStatusConfig(currentStatus),
    [currentStatus],
  );

  // 渲染左侧任务信息面板
  const renderLeftPanel = () => (
    <div className={`job-modal-left-panel ${isWideMode ? "hidden" : ""}`}>
      <div className="job-modal-panel-content">
        {/* 任务编号和所属项目 - 创建模式显示编辑框，编辑模式保留隐藏字段 */}
        {!isEditing ? (
          <>
            <div className="job-form-section">
              <label className="job-form-label">
                <FileTextOutlined className="job-form-icon" />
                任务编号
              </label>
              <Form.Item
                name="jobNo"
                noStyle
                rules={[{ required: true, message: "请输入任务编号" }]}
              >
                <Input className="job-jobno-input" placeholder="输入任务编号" />
              </Form.Item>
            </div>
            <div className="job-form-section">
              <label className="job-form-label">
                <FolderOpenOutlined className="job-form-icon" />
                所属项目
              </label>
              <Form.Item
                name="project"
                noStyle
                rules={[{ required: true, message: "请选择所属项目" }]}
              >
                <Select
                  className="job-project-select"
                  placeholder="选择所属项目"
                  options={PROJECT_OPTIONS.filter((opt) => opt.value !== "")}
                />
              </Form.Item>
            </div>
          </>
        ) : (
          /* 编辑模式：保留隐藏的project字段，确保表单提交时包含该值 */
          <Form.Item name="project" noStyle>
            <input type="hidden" />
          </Form.Item>
        )}
        {/* 任务类型 */}
        <div className="job-form-section">
          <label className="job-form-label">
            <TagOutlined className="job-form-icon" />
            任务类型
          </label>
          <Form.Item
            name="type"
            noStyle
            rules={[{ required: true, message: "请选择任务类型" }]}
          >
            <Select
              className="job-type-select"
              options={JOBTASK_TYPE_OPTIONS.filter((opt) => opt.value !== "")}
              suffixIcon={<DownOutlined className="job-select-arrow" />}
            />
          </Form.Item>
        </div>

        {/* 完成阶段 */}
        <div className="job-form-section">
          <label className="job-form-label">
            <BranchesOutlined className="job-form-icon" />
            完成阶段
          </label>
          <div className="job-status-row">
            <Form.Item
              name="status"
              noStyle
              rules={[{ required: true, message: "请选择完成阶段" }]}
            >
              <Select
                className="job-status-select"
                options={JOBTASK_STATUS_OPTIONS.filter(
                  (opt) => opt.value !== "",
                )}
                suffixIcon={<DownOutlined className="job-select-arrow" />}
              />
            </Form.Item>

            {/* 验收通过复选框 */}
            <Form.Item name="passAcceptStd" valuePropName="checked" noStyle>
              <CheckboxWrapper />
            </Form.Item>
          </div>
        </div>

        {/* 任务目标 */}
        <div className="job-form-section">
          <label className="job-form-label job-form-label-between">
            <span>
              <FlagOutlined className="job-form-icon" />
              任务目标
            </span>
            <span
              className={`job-char-count ${goalCount > 100 ? "exceeded" : ""}`}
            >
              {goalCount}/100
            </span>
          </label>
          <Form.Item
            name="goal"
            noStyle
            rules={[{ required: true, message: "请输入任务目标" }]}
          >
            <Input.TextArea
              rows={4}
              className="job-goal-textarea"
              placeholder="简洁明确的任务目标描述..."
              onChange={(e) => setGoalCount(e.target.value.length)}
            />
          </Form.Item>
        </div>

        {/* 元信息 - 仅编辑模式显示 */}
        {editingJobTask && (
          <div className="job-meta-section">
            <div className="job-meta-item">
              <span className="job-meta-label">
                <ClockCircleOutlined /> 创建时间
              </span>
              <span className="job-meta-value">
                {formatDateTime(editingJobTask.createdAt)}
              </span>
            </div>
            <div className="job-meta-item">
              <span className="job-meta-label">
                <ReloadOutlined /> 更新时间
              </span>
              <span className="job-meta-value">
                {formatDateTime(editingJobTask.updatedAt)}
              </span>
            </div>
            <div className="job-meta-item">
              <span className="job-meta-label">
                <ThunderboltFilled /> 活跃序号
              </span>
              <span className="job-meta-value job-meta-highlight">
                #
                {String(editingJobTask.activeExecutionSequence || 1).padStart(
                  2,
                  "0",
                )}
              </span>
            </div>
          </div>
        )}
      </div>
    </div>
  );

  // 渲染执行记录时间轴
  const renderExecutionTimeline = () => {
    if (!editingJobTask) {
      return (
        <div className="job-timeline-empty">
          <HistoryOutlined className="job-empty-icon" />
          <p>新建任务，暂无执行记录</p>
        </div>
      );
    }

    return (
      <div className="job-timeline-container">
        {executionRecords.length === 0 ? (
          <div className="job-timeline-empty">
            <HistoryOutlined className="job-empty-icon" />
            <p>暂无执行记录</p>
          </div>
        ) : (
          executionRecords.map((record, index) => (
            <ExecutionTimelineItem
              key={record.sequence}
              record={record}
              index={index}
              isExpanded={expandedRecords.has(record.sequence)}
              isActive={
                record.sequence === editingJobTask.activeExecutionSequence
              }
              isLast={index === executionRecords.length - 1}
              onToggle={toggleAccordion}
            />
          ))
        )}
      </div>
    );
  };

  return (
    <Modal
      open={isJobTaskModalOpen}
      onCancel={handleCancel}
      width={isWideMode ? 900 : 1100}
      className="job-task-modal"
      centered
      footer={null}
      closable={false}
      destroyOnClose
      style={{ margin: "20px auto" }}
    >
      <Form form={form} layout="vertical" className="job-modal-form">
        <div className="job-modal-container">
          {/* 头部 */}
          <div className="job-modal-header">
            <div className="job-header-left">
              <div className="job-header-icon">
                <FileTextOutlined />
              </div>
              <div className="job-header-info">
                <div className="job-header-row">
                  <span className="job-header-label">TASK</span>
                  {isEditing && editingJobTask && (
                    <span className="job-jobno-display">
                      {editingJobTask.jobNo}
                    </span>
                  )}
                  <span
                    className="job-status-badge"
                    style={{
                      color: statusConfig.color,
                      background: statusConfig.bgColor,
                      borderColor: statusConfig.borderColor,
                    }}
                  >
                    <span
                      className="job-status-dot"
                      style={{ background: statusConfig.dotColor }}
                    />
                    {currentStatus}
                  </span>
                </div>
                <div className="job-header-row job-header-sub">
                  <FolderOpenOutlined className="job-header-sub-icon" />
                  {isEditing && editingJobTask ? (
                    <span className="job-project-display">
                      {editingJobTask.project}
                    </span>
                  ) : null}
                </div>
              </div>
            </div>
            <div className="job-header-actions">
              <Button
                className="job-header-btn"
                icon={isWideMode ? <EyeOutlined /> : <EditOutlined />}
                onClick={toggleWideMode}
                title={isWideMode ? "显示左侧面板" : "隐藏左侧面板"}
              />
              <Button
                className="job-header-btn job-close-btn"
                icon={<CloseOutlined />}
                onClick={handleCancel}
              />
            </div>
          </div>

          {/* 主体内容 - 左右分栏 */}
          <div className="job-modal-body">
            {renderLeftPanel()}
            <div
              className={`job-modal-right-panel ${isWideMode ? "full" : ""}`}
            >
              {/* 执行记录头部 */}
              <div className="job-timeline-header">
                <div className="job-timeline-title">
                  <div className="job-timeline-icon">
                    <HistoryOutlined />
                  </div>
                  <div>
                    <h3>执行记录</h3>
                    <p>共 {executionRecords.length} 条执行历史</p>
                  </div>
                </div>
              </div>

              {/* 时间轴 */}
              {renderExecutionTimeline()}
            </div>
          </div>

          {/* 底部操作栏 */}
          <div className="job-modal-footer">
            <div className="job-footer-actions">
              <Button onClick={handleCancel} className="job-cancel-btn">
                取消
              </Button>
              <Button
                type="primary"
                onClick={handleSubmit}
                className="job-save-btn"
                icon={<SaveOutlined />}
              >
                {isEditing ? "保存修改" : "创建任务"}
              </Button>
            </div>
          </div>
        </div>
      </Form>
    </Modal>
  );
};

export default JobTaskModal;
