/**
 * 任务相关类型定义
 */

/**
 * 执行记录
 */
export interface ExecutionRecord {
  /** 执行序号 */
  sequence: number;
  /** 执行状态 */
  status: string;
  /** 执行结果 */
  result: string;
  /** 解决方案 */
  solution: string;
  /** 关联文件列表 */
  relatedFiles: string[];
  /** 验收标准 */
  acceptStd?: string;
  /** 使用的技能列表 */
  skills?: string[];
  /** 创建时间戳（毫秒） */
  createdAt?: number;
  /** 更新时间戳（毫秒） */
  updatedAt?: number;
}

/**
 * 任务实体
 */
export interface JobTask {
  /** 任务ID */
  id: number;
  /** 任务编号 */
  jobNo: string;
  /** 所属项目 */
  project: string;
  /** 任务类型 */
  type: string;
  /** 任务目标 */
  goal: string;
  /** 是否通过验收 */
  passAcceptStd: boolean;
  /** 任务状态 */
  status: string;
  /** 执行记录（JSON字符串） */
  executionRecords: string;
  /** 当前执行序号 */
  activeExecutionSequence: number;
  /** 创建时间戳（毫秒） */
  createdAt: number;
  /** 更新时间戳（毫秒） */
  updatedAt: number;
}

/**
 * 任务创建/编辑请求参数
 */
export interface JobTaskRequest {
  /** 任务编号 */
  jobNo: string;
  /** 所属项目 */
  project: string;
  /** 任务类型 */
  type: string;
  /** 任务目标 */
  goal: string;
  /** 是否通过验收 */
  passAcceptStd: boolean;
  /** 任务状态 */
  status: string;
}

/**
 * 任务类型选项
 */
export const JOBTASK_TYPE_OPTIONS = [
  { value: '', label: '全部' },
  { value: '新需求', label: '✨ 新需求' },
  { value: 'Bug修复', label: '🐛 Bug修复' },
  { value: '改进功能', label: '🚀 改进功能' },
  { value: '重构代码', label: '🔧 重构代码' },
  { value: '单元测试', label: '🧪 单元测试' },
  { value: '集成测试', label: '🔨 集成测试' },
  { value: '数据处理', label: '📊 数据处理' },
  { value: '版本控制', label: '📝 版本控制' },
];

/**
 * 任务状态选项
 */
export const JOBTASK_STATUS_OPTIONS = [
  { value: '', label: '全部' },
  { value: '已创建', label: '已创建' },
  { value: '处理中', label: '处理中' },
  { value: '处理失败', label: '处理失败' },
  { value: '处理完成', label: '处理完成' },
  { value: '验收通过', label: '验收通过' },
];

/**
 * 项目选项
 */
export const PROJECT_OPTIONS = [
  { value: '', label: '全部' },
  { value: '智流MCP', label: '智流MCP' },
  { value: 'AI助手', label: 'AI助手' },
  { value: '数据中台', label: '数据中台' },
  { value: '运维平台', label: '运维平台' },
];

/**
 * 验收标准选项
 */
export const ACCEPT_STD_OPTIONS = [
  { value: '人工验收', label: '人工验收' },
  { value: '脚本测试验收', label: '脚本测试验收' },
];
