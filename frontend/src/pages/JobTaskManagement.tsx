/**
 * 任务管理页面组件
 * 采用卡片网格布局展示任务列表，使用Zustand状态管理
 */
import React, { useState, useEffect, useRef, useCallback } from "react";
import {
  Button,
  Popconfirm,
  Select,
  Empty,
  Pagination,
  Badge,
  Card,
  Row,
  Col,
  Tooltip,
  DatePicker,
  Popover,
  Timeline,
  Switch,
} from "antd";
import type { Dayjs } from "dayjs";
import {
  EditOutlined,
  DeleteOutlined,
  FileTextOutlined,
  CheckCircleOutlined,
  SyncOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  FolderOutlined,
  SafetyCertificateOutlined,
  CopyOutlined,
  ReloadOutlined,
  VerticalAlignTopOutlined,
  RestOutlined,
  RollbackOutlined,
  CloseCircleOutlined,
  ExportOutlined,
  DownloadOutlined,
} from "@ant-design/icons";
import { message } from "antd";
import ReactMarkdown from "react-markdown";
import { useAppStore } from "../stores/appStore";
import { useModalStore } from "../stores/modalStore";
import type { JobTask, ExecutionRecord } from "../types";
import { JOBTASK_STATUS_OPTIONS, JOBTASK_TYPE_OPTIONS } from "../types";
import { jobtaskApi } from "../services";
import { formatTime, getTypeConfig, getRecordStatusColor, copyToClipboard } from "../utils";

const { RangePicker } = DatePicker;

/**
 * 获取状态对应的样式配置（带图标版本）
 */
const getStatusConfigWithIcon = (
  status: string,
): { color: string; icon: React.ReactNode; className: string } => {
  const configMap: Record<
    string,
    { color: string; icon: React.ReactNode; className: string }
  > = {
    已创建: {
      color: "#6b7280",
      icon: <ClockCircleOutlined />,
      className: "status-created",
    },
    处理中: {
      color: "#3b82f6",
      icon: <SyncOutlined spin />,
      className: "status-processing",
    },
    处理失败: {
      color: "#ef4444",
      icon: <ExclamationCircleOutlined />,
      className: "status-failed",
    },
    处理完成: {
      color: "#10b981",
      icon: <CheckCircleOutlined />,
      className: "status-success",
    },
    验收通过: {
      color: "#8b5cf6",
      icon: <SafetyCertificateOutlined />,
      className: "status-accepted",
    },
  };
  return (
    configMap[status] || {
      color: "#6b7280",
      icon: <ClockCircleOutlined />,
      className: "status-created",
    }
  );
};

/**
 * 解析执行记录JSON数组，返回处理总数
 */
const getProcessCount = (executionRecords: string | undefined): number => {
  if (!executionRecords) return 0;
  try {
    const records: ExecutionRecord[] = JSON.parse(executionRecords);
    return Array.isArray(records) ? records.length : 0;
  } catch {
    return 0;
  }
};

/**
 * 解析执行记录JSON数组，返回处理记录列表
 */
const getExecutionRecordsList = (
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
 * 过滤空字符串数组
 */
const filterEmptyStrings = (arr: string[] | null | undefined): string[] => {
  if (!arr || !Array.isArray(arr)) return [];
  return arr.filter(s => s && s.trim() !== '');
};

/**
 * 构建处理记录的Markdown内容
 */
const buildRecordMarkdown = (record: ExecutionRecord): string => {
  const parts: string[] = [];
  const validRelatedFiles = filterEmptyStrings(record.relatedFiles);

  if (record.solution) {
    parts.push(`**方案：**\n${record.solution}`);
  }

  if (validRelatedFiles.length > 0) {
    parts.push(`**关联文件：** ${validRelatedFiles.join(", ")}`);
  } else {
    parts.push(`**关联文件：** 无`);
  }

  if (record.result) {
    parts.push(`**处理结果：**\n${record.result}`);
  }

  if (record.acceptStd) {
    parts.push(`**验收标准：** ${record.acceptStd}`);
  }

  return parts.join("\n\n");
};

/**
 * 处理记录弹窗内容组件
 */
const ExecutionRecordsContent: React.FC<{
  records: ExecutionRecord[];
}> = ({ records }) => {
  if (records.length === 0) {
    return <div className="execution-records-empty">暂无处理记录</div>;
  }

  return (
    <div
      className="execution-records-content"
      style={{
        maxHeight: "500px",
        overflowY: "auto",
      }}
    >
      <div className="execution-records-title">处理记录 ({records.length})</div>
      <Timeline
        items={records.map((record, index) => ({
          key: index,
          color: getRecordStatusColor(record.status),
          children: (
            <div className="execution-record-item">
              <div
                className="execution-record-status"
                style={{ color: getRecordStatusColor(record.status) }}
              >
                {record.status}
              </div>
              <div className="execution-record-result">
                <ReactMarkdown>{buildRecordMarkdown(record)}</ReactMarkdown>
              </div>
            </div>
          ),
        }))}
      />
    </div>
  );
};

/**
 * 从执行记录中获取最新的验收标准
 */
const getLatestAcceptStd = (executionRecords: string | undefined): string | null => {
  if (!executionRecords) return null;
  try {
    const records: ExecutionRecord[] = JSON.parse(executionRecords);
    if (!Array.isArray(records) || records.length === 0) return null;
    const latestRecord = records[records.length - 1];
    return latestRecord.acceptStd || null;
  } catch {
    return null;
  }
};

/**
 * 任务卡片组件
 */
const JobCard: React.FC<{
  job: JobTask;
  onEdit: (job: JobTask) => void;
  onDelete: (id: number) => void;
  isTrashMode?: boolean;
  onRestore?: (id: number, onSuccess?: () => void) => void;
  onPermanentDelete?: (id: number, onSuccess?: () => void) => void;
  onRefresh?: () => void;
  isSelectMode?: boolean;
  isSelected?: boolean;
  onSelect?: (id: number) => void;
}> = ({ job, onEdit, onDelete, isTrashMode = false, onRestore, onPermanentDelete, onRefresh, isSelectMode = false, isSelected = false, onSelect }) => {
  const statusConfig = getStatusConfigWithIcon(job.status);
  const typeConfig = getTypeConfig(job.type);
  const processCount = getProcessCount(job.executionRecords);
  const executionRecordsList = getExecutionRecordsList(job.executionRecords);
  const latestAcceptStd = getLatestAcceptStd(job.executionRecords);

  const handleCopyJobNo = (): void => {
    copyToClipboard(job.jobNo, `任务号 ${job.jobNo} 已复制`);
  };

  const handleDo = (): void => {
    const command = `job_redo:${job.jobNo};`;
    copyToClipboard(command, `命令 ${command} 已复制`);
  };

  const handleCopyTitle = (): void => {
    copyToClipboard(job.goal, `目标已复制`);
  };

  const handleCopyProject = (): void => {
    copyToClipboard(job.project, `项目已复制`);
  };

  const handleCardClick = (): void => {
    if (isSelectMode && onSelect) {
      onSelect(job.id);
    }
  };

  return (
    <div
      className={`task-card ${statusConfig.className} ${isSelectMode ? 'selectable' : ''} ${isSelected ? 'selected' : ''}`}
      onClick={handleCardClick}
    >
      {isSelectMode && (
        <div className="task-select-checkbox">
          <input
            type="checkbox"
            checked={isSelected}
            onChange={() => onSelect?.(job.id)}
            onClick={(e) => e.stopPropagation()}
          />
        </div>
      )}
      <div className="task-card-header">
        <div className="task-id-wrapper">
          <Tooltip title="点击复制任务号">
            <span
              className="task-id task-id-clickable"
              onClick={handleCopyJobNo}
            >
              {job.jobNo}
            </span>
          </Tooltip>
          <Tooltip title="复制redo命令">
            <button
              className="btn-redo-mini btn-icon"
              style={{ width: "28px", height: "28px" }}
              onClick={handleDo}
            >
              <CopyOutlined />
            </button>
          </Tooltip>
        </div>
        <div className="task-header-actions">
          <span
            className="task-type"
            style={{
              color: typeConfig.color,
              backgroundColor: typeConfig.bgColor,
            }}
          >
            <span>{typeConfig.emoji}</span>
            <span>{job.type}</span>
          </span>
        </div>
      </div>

      <Tooltip title="点击复制目标">
        <h3
          className="task-title task-clickable"
          title={job.goal}
          onClick={handleCopyTitle}
        >
          {job.goal}
        </h3>
      </Tooltip>

      <div className="task-meta">
        <div className="meta-item">
          <FolderOutlined className="meta-icon" />
          <span className="meta-label">项目</span>
          <Tooltip title="点击复制项目">
            <span
              className="meta-value task-clickable"
              onClick={handleCopyProject}
            >
              {job.project}
            </span>
          </Tooltip>
        </div>
        <div className="meta-item">
          <SafetyCertificateOutlined className="meta-icon" />
          <span className="meta-label">验收</span>
          <Tooltip title={`验收方法: ${latestAcceptStd || '未指定'}`}>
            {job.passAcceptStd ? (
              <span className="accept-badge accept-passed">已通过({latestAcceptStd || '未指定'})</span>
            ) : job.status === "处理完成" ? (
              <span className="accept-badge accept-pending">待验收({latestAcceptStd || '未指定'})</span>
            ) : (
              <span className="accept-badge accept-not-passed">未通过({latestAcceptStd || '未指定'})</span>
            )}
          </Tooltip>
        </div>
      </div>

      <div className="task-footer">
        <div className="task-status" style={{ color: statusConfig.color }}>
          <Popover
            content={<ExecutionRecordsContent records={executionRecordsList} />}
            title={null}
            trigger="click"
            placement="topLeft"
            overlayClassName="execution-records-popover"
          >
            <span
              className={`process-count-badge ${statusConfig.className}`}
              style={{ cursor: "pointer" }}
            >
              {processCount}
            </span>
          </Popover>
          <Badge
            color={statusConfig.color}
            className={job.status === "处理中" ? "status-pulse" : ""}
          />
          <span>{statusConfig.icon}</span>
          <span>{job.status}</span>
          <div className="task-time">创建于 {formatTime(job.createdAt)}</div>
        </div>
        <div className="task-actions">
          {!isSelectMode && (!isTrashMode ? (
            <>
              <button
                className="btn-icon"
                title="编辑"
                onClick={(e) => { e.stopPropagation(); onEdit(job); }}
              >
                <EditOutlined />
              </button>
              <Popconfirm
                title="确定要删除这个任务吗？删除后可在回收站恢复。"
                onConfirm={() => onDelete(job.id)}
                okText="确定"
                cancelText="取消"
                placement="topRight"
              >
                <button className="btn-icon btn-delete" title="删除" onClick={(e) => e.stopPropagation()}>
                  <DeleteOutlined />
                </button>
              </Popconfirm>
            </>
          ) : (
            <>
              <Popconfirm
                title="确定要恢复这个任务吗？"
                onConfirm={() => onRestore?.(job.id, onRefresh)}
                okText="确定"
                cancelText="取消"
                placement="topRight"
              >
                <button className="btn-icon" title="恢复" style={{ color: '#10b981' }}>
                  <RollbackOutlined />
                </button>
              </Popconfirm>
              <Popconfirm
                title="确定要彻底删除这个任务吗？删除后无法恢复！"
                onConfirm={() => onPermanentDelete?.(job.id, onRefresh)}
                okText="确定"
                cancelText="取消"
                placement="topRight"
              >
                <button className="btn-icon btn-delete" title="彻底删除" onClick={(e) => e.stopPropagation()}>
                  <CloseCircleOutlined />
                </button>
              </Popconfirm>
            </>
          ))}
        </div>
      </div>
    </div>
  );
};

/**
 * 统计卡片组件
 */
const StatCard: React.FC<{
  title: string;
  value: number;
  icon: React.ReactNode;
  color: string;
}> = ({ title, value, icon, color }) => (
  <Card className="stat-card" variant="borderless">
    <div className="stat-content">
      <div>
        <div className="stat-title">{title}</div>
        <div className="stat-value" style={{ color }}>
          {value}
        </div>
      </div>
      <div
        className="stat-icon"
        style={{
          backgroundColor: `${color}1A`,
          color,
        }}
      >
        {icon}
      </div>
    </div>
  </Card>
);

/**
 * 任务管理页面组件
 */
const JobTaskManagement: React.FC = () => {
  // 从Zustand Store获取状态和actions
  const {
    jobTasks,
    projects,
    selectedProject,
    selectedJobTaskType,
    selectedJobTaskStatus,
    selectedJobTaskDateRange,
    autoRefresh,
    jobTaskPagination,
    collapsed,
    setJobTasks,
    setProjects,
    setSelectedProject,
    setSelectedJobTaskType,
    setSelectedJobTaskStatus,
    setSelectedJobTaskDateRange,
    setAutoRefresh,
    setJobTaskPagination,
  } = useAppStore();
  const { openJobTaskModal } = useModalStore();

  // 对 jobTasks 做空值保护，确保始终是数组
  const safeJobTasks = jobTasks || [];

  // 解构分页状态
  const { page: jobTaskPage, pageSize: jobTaskPageSize, total: jobTaskTotal } = jobTaskPagination;

  // 构建项目选项
  const projectOptions = [
    { value: "", label: "全部" },
    ...projects.map((project) => ({ value: project, label: project })),
  ];

  // 计算统计数据
  const totalCount = jobTaskPagination.total;
  const successCount = safeJobTasks.filter(
    (j) => j.status === "处理完成" || j.status === "验收通过",
  ).length;
  const failedCount = safeJobTasks.filter((j) => j.status === "处理失败").length;
  const pendingCount = safeJobTasks.filter((j) => j.status === "已创建").length;

  // 筛选栏悬浮状态
  const [isFilterSticky, setIsFilterSticky] = useState(false);
  const sentinelRef = useRef<HTMLDivElement>(null);
  // 窗口宽度状态
  const [windowWidth, setWindowWidth] = useState(window.innerWidth);
  // 移动端断点
  const MOBILE_BREAKPOINT = 768;
  // 回到顶部按钮显示状态
  const [showBackToTop, setShowBackToTop] = useState(false);
  // 页面容器ref
  const containerRef = useRef<HTMLDivElement>(null);
  // 回收站模式状态
  const [isTrashMode, setIsTrashMode] = useState(false);
  // 批量选择模式状态
  const [isSelectMode, setIsSelectMode] = useState(false);
  // 选中的任务ID列表
  const [selectedTaskIds, setSelectedTaskIds] = useState<number[]>([]);
  // 导出弹窗显示状态
  const [exportModalVisible, setExportModalVisible] = useState(false);
  // 导出格式
  const [exportFormat, setExportFormat] = useState<'csv' | 'json' | 'md'>('csv');

  /**
   * 加载任务数据
   */
  const loadJobTaskData = useCallback(async (
    project?: string,
    jobTaskType?: string,
    jobTaskStatus?: string,
    dateRange?: [Dayjs | null, Dayjs | null] | null,
    page?: number,
    pageSize?: number,
  ) => {
    try {
      const currentPage = page ?? jobTaskPage;
      const currentPageSize = pageSize ?? jobTaskPageSize;
      const startDate = dateRange?.[0]?.valueOf();
      const endDate = dateRange?.[1]?.valueOf();
      const [loadedJobTasks, loadedProjects] = await Promise.all([
        jobtaskApi.getJobTasks(
          project,
          jobTaskType,
          jobTaskStatus,
          currentPage,
          currentPageSize,
          startDate,
          endDate,
        ),
        jobtaskApi.getProjects(),
      ]);
      setJobTasks(loadedJobTasks.items);
      setJobTaskPagination({ total: loadedJobTasks.pagination.total });
      setProjects(loadedProjects);
    } catch (error) {
      message.error("加载任务数据失败");
      console.error("Load jobtask data failed:", error);
    }
  }, [jobTaskPage, jobTaskPageSize, setJobTasks, setProjects, setJobTaskPagination]);

  /**
   * 加载回收站任务列表
   */
  const loadTrashJobTasks = useCallback(async (page?: number, pageSize?: number) => {
    try {
      const currentPage = page ?? jobTaskPage;
      const currentPageSize = pageSize ?? jobTaskPageSize;
      const loadedTrashJobs = await jobtaskApi.getTrashJobTasks(currentPage, currentPageSize);
      setJobTasks(loadedTrashJobs.items);
      setJobTaskPagination({ total: loadedTrashJobs.pagination.total });
    } catch (error) {
      message.error("加载回收站数据失败");
      console.error("Load trash jobtask data failed:", error);
    }
  }, [jobTaskPage, jobTaskPageSize, setJobTasks, setJobTaskPagination]);

  /**
   * 刷新回收站数据
   */
  const refreshTrashData = useCallback(() => {
    if (isTrashMode) {
      loadTrashJobTasks(jobTaskPage, jobTaskPageSize);
    }
  }, [isTrashMode, loadTrashJobTasks, jobTaskPage, jobTaskPageSize]);

  /**
   * 组件挂载时加载数据
   */
  useEffect(() => {
    loadJobTaskData(
      selectedProject,
      selectedJobTaskType,
      selectedJobTaskStatus,
      selectedJobTaskDateRange,
      jobTaskPage,
      jobTaskPageSize,
    );
  }, []);

  /**
   * 切换回收站模式时加载对应数据
   */
  useEffect(() => {
    if (isTrashMode) {
      loadTrashJobTasks(jobTaskPage, jobTaskPageSize);
    } else {
      loadJobTaskData(
        selectedProject,
        selectedJobTaskType,
        selectedJobTaskStatus,
        selectedJobTaskDateRange,
        jobTaskPage,
        jobTaskPageSize,
      );
      setTimeout(() => {
        const sentinel = sentinelRef.current;
        if (sentinel) {
          const rect = sentinel.getBoundingClientRect();
          setIsFilterSticky(rect.bottom <= 0);
        }
      }, 0);
    }
  }, [isTrashMode]);

  // 使用IntersectionObserver API监听哨兵元素，实现筛选栏悬浮
  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        setIsFilterSticky(!entry.isIntersecting);
      },
      {
        root: null,
        threshold: 0,
        rootMargin: "0px 0px 0px 0px",
      },
    );

    observer.observe(sentinel);

    return () => {
      observer.disconnect();
    };
  }, []);

  /**
   * 监听窗口大小变化
   */
  useEffect(() => {
    const handleResize = () => {
      setWindowWidth(window.innerWidth);
    };

    window.addEventListener("resize", handleResize);
    return () => {
      window.removeEventListener("resize", handleResize);
    };
  }, []);

  /**
   * 监听页面滚动，控制回到顶部按钮显示
   */
  useEffect(() => {
    const container = document.body;

    const handleScroll = () => {
      setShowBackToTop(container.scrollTop > 300);
    };

    container.addEventListener("scroll", handleScroll);
    handleScroll();

    return () => {
      container.removeEventListener("scroll", handleScroll);
    };
  }, []);

  /**
   * 回到顶部
   */
  const handleBackToTop = (): void => {
    document.body.scrollTo({
      top: 0,
      behavior: "smooth",
    });
  };

  /**
   * 切换任务选择状态
   */
  const toggleTaskSelection = (id: number): void => {
    setSelectedTaskIds(prev => {
      if (prev.includes(id)) {
        return prev.filter(taskId => taskId !== id);
      } else {
        return [...prev, id];
      }
    });
  };

  /**
   * 全选/取消全选
   */
  const toggleSelectAll = (): void => {
    if (selectedTaskIds.length === safeJobTasks.length) {
      setSelectedTaskIds([]);
    } else {
      setSelectedTaskIds(safeJobTasks.map(task => task.id));
    }
  };

  /**
   * 处理导出
   */
  const handleExport = (): void => {
    const ids = selectedTaskIds.length > 0 ? selectedTaskIds : undefined;
    jobtaskApi.exportJobTasks(ids, exportFormat);
    setExportModalVisible(false);
    setIsSelectMode(false);
    setSelectedTaskIds([]);
  };

  /**
   * 取消选择模式
   */
  const cancelSelectMode = (): void => {
    setIsSelectMode(false);
    setSelectedTaskIds([]);
  };

  // 判断是否为移动端
  const isMobile = windowWidth < MOBILE_BREAKPOINT;

  /**
   * 处理删除任务
   */
  const handleDeleteJobTask = async (id: number) => {
    try {
      await jobtaskApi.deleteJobTask(id);
      const updatedJobTasks = jobTasks.filter((jt) => jt.id !== id);
      setJobTasks(updatedJobTasks);
      message.success("任务已移至回收站");
    } catch (error) {
      message.error("删除失败，请重试");
      console.error("JobTask delete failed:", error);
    }
  };

  /**
   * 恢复回收站中的任务
   */
  const handleRestoreJobTask = async (id: number, onSuccess?: () => void) => {
    try {
      await jobtaskApi.restoreJobTask(id);
      message.success("任务恢复成功");
      onSuccess?.();
    } catch (error) {
      message.error("恢复失败，请重试");
      console.error("JobTask restore failed:", error);
    }
  };

  /**
   * 彻底删除任务
   */
  const handlePermanentDeleteJobTask = async (id: number, onSuccess?: () => void) => {
    try {
      await jobtaskApi.permanentDeleteJobTask(id);
      message.success("任务已彻底删除");
      onSuccess?.();
    } catch (error) {
      message.error("删除失败，请重试");
      console.error("JobTask permanent delete failed:", error);
    }
  };

  /**
   * 处理分页变化
   */
  const handlePageChange = (page: number, pageSize: number) => {
    setJobTaskPagination({ page, pageSize });
    if (isTrashMode) {
      loadTrashJobTasks(page, pageSize);
    } else {
      loadJobTaskData(
        selectedProject,
        selectedJobTaskType,
        selectedJobTaskStatus,
        selectedJobTaskDateRange,
        page,
        pageSize,
      );
    }
  };

  /**
   * 处理项目变化
   */
  const handleProjectChange = (project: string | undefined) => {
    setSelectedProject(project);
    setJobTaskPagination({ page: 1 });
    loadJobTaskData(
      project,
      selectedJobTaskType,
      selectedJobTaskStatus,
      selectedJobTaskDateRange,
      1,
      jobTaskPageSize,
    );
  };

  /**
   * 处理类型变化
   */
  const handleTypeChange = (type: string | undefined) => {
    setSelectedJobTaskType(type);
    setJobTaskPagination({ page: 1 });
    loadJobTaskData(
      selectedProject,
      type,
      selectedJobTaskStatus,
      selectedJobTaskDateRange,
      1,
      jobTaskPageSize,
    );
  };

  /**
   * 处理状态变化
   */
  const handleStatusChange = (status: string | undefined) => {
    setSelectedJobTaskStatus(status);
    setJobTaskPagination({ page: 1 });
    loadJobTaskData(
      selectedProject,
      selectedJobTaskType,
      status,
      selectedJobTaskDateRange,
      1,
      jobTaskPageSize,
    );
  };

  /**
   * 处理日期范围变化
   */
  const handleDateRangeChange = (dates: [Dayjs | null, Dayjs | null] | null) => {
    setSelectedJobTaskDateRange(dates);
    setJobTaskPagination({ page: 1 });
    loadJobTaskData(
      selectedProject,
      selectedJobTaskType,
      selectedJobTaskStatus,
      dates,
      1,
      jobTaskPageSize,
    );
  };

  /**
   * 处理重置筛选
   */
  const handleResetFilters = () => {
    setSelectedProject(undefined);
    setSelectedJobTaskType(undefined);
    setSelectedJobTaskStatus(undefined);
    setSelectedJobTaskDateRange(null);
    setJobTaskPagination({ page: 1 });
    loadJobTaskData(undefined, undefined, undefined, null, 1, jobTaskPageSize);
  };

  /**
   * 处理自动刷新变化
   */
  const handleAutoRefreshChange = (enabled: boolean) => {
    setAutoRefresh(enabled);
  };

  /**
   * 自动刷新定时器
   */
  useEffect(() => {
    if (!autoRefresh || isTrashMode) return;

    const intervalId = setInterval(() => {
      loadJobTaskData(
        selectedProject,
        selectedJobTaskType,
        selectedJobTaskStatus,
        selectedJobTaskDateRange,
        jobTaskPage,
        jobTaskPageSize,
      );
    }, 5000);

    return () => {
      clearInterval(intervalId);
    };
  }, [
    autoRefresh,
    isTrashMode,
    selectedProject,
    selectedJobTaskType,
    selectedJobTaskStatus,
    selectedJobTaskDateRange,
    jobTaskPage,
    jobTaskPageSize,
    loadJobTaskData,
  ]);

  return (
    <div className="worktask-management" ref={containerRef}>
      {/* 头部区域 */}
      <div className="worktask-header">
        <div className="header-top">
          <h1 className="header-title">{isTrashMode ? '回收站' : '任务管理'}</h1>
          <div style={{ display: 'flex', gap: '12px' }}>
            <Button
              className="btn-secondary-gradient"
              icon={<RestOutlined />}
              onClick={() => setIsTrashMode(!isTrashMode)}
            >
              {isTrashMode ? '返回列表' : '回收站'}
            </Button>
            {/* {!isTrashMode && ( 暂时不开放
              <Button
                type="primary"
                className="btn-primary-gradient"
                icon={<PlusOutlined />}
                onClick={() => openJobTaskModal()}
              >
                新增任务
              </Button>
            )} */}
          </div>
        </div>
      </div>

      {/* 筛选栏 */}
      {!isTrashMode && (
        <div
          className={`filter-bar-wrapper ${isFilterSticky && !isMobile ? "filter-bar-wrapper-sticky" : ""}`}
          style={{ '--sider-width': collapsed ? '80px' : '200px' } as React.CSSProperties}
        >
          <div ref={sentinelRef} className="filter-sentinel" />
          <div className="filter-bar">
            <div className="filter-group">
              <span className="filter-label">项目</span>
              <Select
                className="filter-select"
                placeholder="选择项目"
                value={selectedProject}
                onChange={(value) => handleProjectChange(value || undefined)}
                options={projectOptions}
                variant="borderless"
              />
            </div>
            <div className="filter-group">
              <span className="filter-label">类型</span>
              <Select
                className="filter-select"
                placeholder="选择类型"
                value={selectedJobTaskType}
                onChange={(value) => handleTypeChange(value || undefined)}
                options={JOBTASK_TYPE_OPTIONS}
                variant="borderless"
                listHeight={320}
                popupMatchSelectWidth={false}
              />
            </div>
            <div className="filter-group">
              <span className="filter-label">状态</span>
              <Select
                className="filter-select"
                placeholder="选择状态"
                value={selectedJobTaskStatus}
                onChange={(value) => handleStatusChange(value || undefined)}
                options={JOBTASK_STATUS_OPTIONS}
                variant="borderless"
              />
            </div>
            <div className="filter-group">
              <span className="filter-label">创建时间</span>
              <RangePicker
                className="filter-date-range"
                value={selectedJobTaskDateRange}
                onChange={handleDateRangeChange}
                placeholder={["开始时间", "结束时间"]}
                format="YYYY-MM-DD HH:mm"
                showTime={{ format: "HH:mm" }}
              />
            </div>
            <Button
              className="btn-secondary-gradient"
              icon={<ReloadOutlined />}
              onClick={handleResetFilters}
            >
              重置
            </Button>
            <Button
              className="btn-secondary-gradient"
              icon={<ExportOutlined />}
              onClick={() => setIsSelectMode(true)}
            >
              导出
            </Button>
            <div className="filter-group auto-refresh-group">
              <span className="filter-label">自动更新</span>
              <Switch
                checked={autoRefresh}
                onChange={handleAutoRefreshChange}
                size="small"
              />
            </div>
          </div>
        </div>
      )}

      {/* 统计卡片 */}
      {!isTrashMode && (
        <Row gutter={[20, 20]} className="stats-row">
          <Col xs={12} sm={12} lg={6}>
            <StatCard
              title="总任务数"
              value={totalCount}
              icon={<FileTextOutlined />}
              color="#3b82f6"
            />
          </Col>
          <Col xs={12} sm={12} lg={6}>
            <StatCard
              title="处理完成"
              value={successCount}
              icon={<CheckCircleOutlined />}
              color="#10b981"
            />
          </Col>
          <Col xs={12} sm={12} lg={6}>
            <StatCard
              title="处理失败"
              value={failedCount}
              icon={<ExclamationCircleOutlined />}
              color="#ef4444"
            />
          </Col>
          <Col xs={12} sm={12} lg={6}>
            <StatCard
              title="待处理"
              value={pendingCount}
              icon={<ClockCircleOutlined />}
              color="#ef4444"
            />
          </Col>
        </Row>
      )}

      {/* 任务卡片网格 */}
      {safeJobTasks.length > 0 ? (
        <>
          <div className={`tasks-grid ${isSelectMode ? 'select-mode' : ''}`}>
            {safeJobTasks.map((job) => (
              <JobCard
                key={job.id}
                job={job}
                onEdit={openJobTaskModal}
                onDelete={handleDeleteJobTask}
                isTrashMode={isTrashMode}
                onRestore={handleRestoreJobTask}
                onPermanentDelete={handlePermanentDeleteJobTask}
                onRefresh={refreshTrashData}
                isSelectMode={isSelectMode}
                isSelected={selectedTaskIds.includes(job.id)}
                onSelect={toggleTaskSelection}
              />
            ))}
          </div>

          {/* 分页 */}
          <div className="pagination-wrapper">
            <Pagination
              total={jobTaskTotal}
              current={jobTaskPage}
              pageSize={jobTaskPageSize}
              showSizeChanger
              showQuickJumper
              onChange={handlePageChange}
              pageSizeOptions={["20", "50", "100"]}
              locale={{
                items_per_page: '条/页',
                jump_to: '跳至',
                jump_to_confirm: '确定',
                page: '页',
                prev_page: '上一页',
                next_page: '下一页',
                prev_5: '向前 5 页',
                next_5: '向后 5 页',
                prev_3: '向前 3 页',
                next_3: '向后 3 页',
              }}
            />
          </div>
        </>
      ) : (
        <Empty
          className="worktask-empty"
          description={isTrashMode ? "回收站为空" : "暂无任务数据"}
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
      )}

      {/* 回到顶部按钮 */}
      {showBackToTop && (
        <button
          className="back-to-top-btn"
          onClick={handleBackToTop}
          title="回到顶部"
        >
          <VerticalAlignTopOutlined />
        </button>
      )}

      {/* 导出弹窗 */}
      {exportModalVisible && (
        <div className="modal-overlay" onClick={() => setExportModalVisible(false)}>
          <div className="export-modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">选择导出格式</h3>
              <button
                className="modal-close"
                onClick={() => setExportModalVisible(false)}
              >
                ×
              </button>
            </div>
            <div className="modal-body">
              <div className="export-options">
                <div className="export-option-group">
                  <label className="export-label">导出格式</label>
                  <div className="export-format-options">
                    {[
                      { value: 'csv', label: 'CSV', desc: '表格格式，适合Excel查看' },
                      { value: 'json', label: 'JSON', desc: '数据格式，适合程序处理' },
                      { value: 'md', label: 'Markdown', desc: '文档格式，适合阅读' },
                    ].map((fmt) => (
                      <button
                        key={fmt.value}
                        className={`export-format-btn ${exportFormat === fmt.value ? 'active' : ''}`}
                        onClick={() => setExportFormat(fmt.value as 'csv' | 'json' | 'md')}
                      >
                        <span className="format-name">{fmt.label}</span>
                        <span className="format-desc">{fmt.desc}</span>
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            </div>
            <div className="modal-footer">
              <Button onClick={() => setExportModalVisible(false)}>
                取消
              </Button>
              <Button
                type="primary"
                icon={<DownloadOutlined />}
                onClick={handleExport}
              >
                导出
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* 选择模式工具栏 */}
      {isSelectMode && (
        <div className="select-mode-toolbar">
          <div className="select-mode-info">
            <span>已选择 {selectedTaskIds.length} 项</span>
            <Button type="link" onClick={toggleSelectAll}>
              {selectedTaskIds.length === safeJobTasks.length ? '取消全选' : '全选'}
            </Button>
          </div>
          <div className="select-mode-actions">
            <Button onClick={cancelSelectMode}>取消</Button>
            <Button
              type="primary"
              icon={<DownloadOutlined />}
              onClick={() => setExportModalVisible(true)}
            >
              导出
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};

export default JobTaskManagement;
