# 前端代码重构落地方案

> 文档版本：v1.0
> 创建日期：2026-02-12
> 项目名称：智流MCP

---

## 一、重构目标

### 1.1 核心指标对比表

| 指标 | 重构前 | 重构后 | 改善幅度 |
|------|--------|--------|----------|
| 类型定义文件 | 混合在 services/index.ts | 独立 types/ 目录 | 职责分离 |
| 公共组件复用 | StatCard 重复定义 2 次 | 提取为公共组件 | 代码减少 ~100 行 |
| 工具函数复用 | formatTime 重复定义 2 次 | 提取为公共函数 | 代码减少 ~30 行 |
| 魔法数字 | 硬编码散落各处 | 统一常量管理 | 可维护性提升 |
| CSS 文件行数 | App.css 4000+ 行 | 按模块拆分 | 可读性提升 |
| 自定义 Hooks | 无 | 3+ 个可复用 Hook | 逻辑复用 |

### 1.2 重构原则

1. **渐进式重构**：分阶段实施，每个阶段独立可验证
2. **向后兼容**：保持 API 接口不变，仅调整内部实现
3. **测试驱动**：重构后必须通过编译检查
4. **文档同步**：代码变更同步更新注释

---

## 二、重构阶段规划

### 优先级说明

| 优先级 | 说明 | 预计工时 |
|--------|------|----------|
| P0 | 核心架构优化，影响后续开发 | 2-3 天 |
| P1 | 代码质量提升，减少重复 | 1-2 天 |
| P2 | 样式优化，提升可维护性 | 1 天 |

### 阶段总览

```
阶段 P0（核心架构）
├── 任务 1：类型定义独立化
├── 任务 2：常量提取与统一管理
└── 任务 3：工具函数提取

阶段 P1（组件优化）
├── 任务 4：公共组件提取
├── 任务 5：自定义 Hooks 提取
└── 任务 6：页面组件瘦身

阶段 P2（样式优化）
├── 任务 7：CSS 样式拆分
└── 任务 8：样式命名规范化
```

---

## 三、阶段 P0：核心架构优化

### 任务 1：类型定义独立化

#### 1.1 问题描述

当前 `services/index.ts` 文件混合了类型定义和 API 实现，违反单一职责原则：

```typescript
// 当前结构（问题代码）
// services/index.ts
export interface Tag { ... }        // 类型定义
export interface Skill { ... }      // 类型定义
export const tagApi = { ... }       // API 实现
export const skillApi = { ... }     // API 实现
```

#### 1.2 目标结构

```
frontend/src/
├── types/
│   ├── index.ts          # 类型统一导出
│   ├── skill.ts          # 技能相关类型
│   ├── jobtask.ts        # 任务相关类型
│   ├── tag.ts            # 标签相关类型
│   ├── common.ts         # 通用类型（分页等）
│   └── constant.ts       # 常量定义（已存在）
└── services/
    └── index.ts          # 仅保留 API 实现
```

#### 1.3 实施步骤

**步骤 1：创建类型文件**

创建 `types/common.ts`：

```typescript
/**
 * 通用类型定义
 * 包含分页、响应等基础类型
 */

/**
 * 分页参数结构
 */
export interface Pagination {
  /** 总记录数 */
  total: number;
  /** 当前页码 */
  page: number;
  /** 每页条数 */
  pageSize: number;
  /** 总页数 */
  totalPage: number;
}

/**
 * 带分页的响应结构
 * @template T - 数据项类型
 */
export interface PaginatedResponse<T> {
  /** 数据列表 */
  items: T[];
  /** 分页信息 */
  pagination: Pagination;
}
```

创建 `types/tag.ts`：

```typescript
/**
 * 标签相关类型定义
 */

/**
 * 标签实体
 */
export interface Tag {
  /** 标签ID */
  id: number;
  /** 标签名称 */
  name: string;
  /** 创建时间戳（毫秒） */
  createdAt: number;
  /** 更新时间戳（毫秒） */
  updatedAt: number;
}
```

创建 `types/skill.ts`：

```typescript
/**
 * 技能相关类型定义
 */

import type { Tag } from './tag';

/**
 * 技能实体
 */
export interface Skill {
  /** 技能ID */
  id: number;
  /** 技能名称 */
  name: string;
  /** 资源目录 */
  resourceDir: string;
  /** 描述 */
  description: string;
  /** 详细说明 */
  detail: string;
  /** 许可证 */
  license: string;
  /** 兼容性说明 */
  compatibility: string;
  /** 元数据 */
  metadata: string;
  /** 允许的工具列表 */
  allowedTools: string;
  /** 关联标签 */
  tags: Tag[];
  /** 创建时间戳（毫秒） */
  createdAt: number;
  /** 更新时间戳（毫秒） */
  updatedAt: number;
}

/**
 * 技能创建/编辑请求参数
 */
export interface SkillRequest {
  /** 技能名称 */
  name: string;
  /** 资源目录 */
  resourceDir: string;
  /** 描述 */
  description: string;
  /** 版本号 */
  version: string;
  /** 详细说明 */
  detail: string;
  /** 许可证 */
  license: string;
  /** 兼容性说明 */
  compatibility: string;
  /** 元数据 */
  metadata: string;
  /** 允许的工具列表 */
  allowedTools: string;
  /** 关联标签ID列表 */
  tags: number[];
}
```

创建 `types/jobtask.ts`：

```typescript
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
] as const;

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
] as const;

/**
 * 项目选项
 */
export const PROJECT_OPTIONS = [
  { value: '', label: '全部' },
  { value: '智流MCP', label: '智流MCP' },
  { value: 'AI助手', label: 'AI助手' },
  { value: '数据中台', label: '数据中台' },
  { value: '运维平台', label: '运维平台' },
] as const;

/**
 * 验收标准选项
 */
export const ACCEPT_STD_OPTIONS = [
  { value: '人工验收', label: '人工验收' },
  { value: '脚本测试验收', label: '脚本测试验收' },
] as const;
```

创建 `types/index.ts`：

```typescript
/**
 * 类型定义统一导出
 * 所有类型从此文件导出，便于统一管理
 */

// 通用类型
export type { Pagination, PaginatedResponse } from './common';

// 标签类型
export type { Tag } from './tag';

// 技能类型
export type { Skill, SkillRequest } from './skill';

// 任务类型
export type {
  ExecutionRecord,
  JobTask,
  JobTaskRequest,
} from './jobtask';
export {
  JOBTASK_TYPE_OPTIONS,
  JOBTASK_STATUS_OPTIONS,
  PROJECT_OPTIONS,
  ACCEPT_STD_OPTIONS,
} from './jobtask';

// 常量
export { API_BASE_URL } from './constant';
```

**步骤 2：更新 services/index.ts**

重构后的 `services/index.ts`：

```typescript
/**
 * API服务层
 * 用于对接后端API接口
 */

import { API_BASE_URL } from '../types';
import type {
  Tag,
  Skill,
  SkillRequest,
  JobTask,
  JobTaskRequest,
  PaginatedResponse,
} from '../types';

// 通用请求函数
async function request<T>(url: string, options?: RequestInit): Promise<T> {
  try {
    const response = await fetch(`${API_BASE_URL}${url}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    if (!data.success) {
      throw new Error(data.error || '操作失败');
    }

    return data.data as T;
  } catch (error) {
    console.error('API request failed:', error);
    throw error;
  }
}

// 标签API
export const tagApi = {
  async getTags(page: number = 1, pageSize: number = 10): Promise<PaginatedResponse<Tag>> {
    return request<PaginatedResponse<Tag>>(`/tags?page=${page}&pageSize=${pageSize}`);
  },

  async createTag(name: string): Promise<Tag> {
    return request<Tag>('/tags', {
      method: 'POST',
      body: JSON.stringify({ name }),
    });
  },

  async updateTag(id: number, name: string): Promise<Tag> {
    return request<Tag>(`/tags/${id}`, {
      method: 'PUT',
      body: JSON.stringify({ name }),
    });
  },

  async deleteTag(id: number): Promise<void> {
    await request<void>(`/tags/${id}`, { method: 'DELETE' });
  },
};

// 技能API
export const skillApi = {
  async getSkills(
    tagId?: number,
    page: number = 1,
    pageSize: number = 10,
    startDate?: number,
    endDate?: number
  ): Promise<PaginatedResponse<Skill>> {
    let url = '/skills';
    const params = new URLSearchParams();
    if (tagId) params.append('tagId', tagId.toString());
    if (startDate) params.append('startDate', startDate.toString());
    if (endDate) params.append('endDate', endDate.toString());
    params.append('page', page.toString());
    params.append('pageSize', pageSize.toString());

    const queryString = params.toString();
    if (queryString) url += `?${queryString}`;

    return request<PaginatedResponse<Skill>>(url);
  },

  async createSkill(skill: SkillRequest): Promise<Skill> {
    return request<Skill>('/skills', {
      method: 'POST',
      body: JSON.stringify(skill),
    });
  },

  async updateSkill(id: number, skill: SkillRequest): Promise<Skill> {
    return request<Skill>(`/skills/${id}`, {
      method: 'PUT',
      body: JSON.stringify(skill),
    });
  },

  async deleteSkill(id: number): Promise<void> {
    await request<void>(`/skills/${id}`, { method: 'DELETE' });
  },

  async getTrashSkills(page: number = 1, pageSize: number = 10): Promise<PaginatedResponse<Skill>> {
    const params = new URLSearchParams();
    params.append('page', page.toString());
    params.append('pageSize', pageSize.toString());
    return request<PaginatedResponse<Skill>>(`/skills/trash?${params.toString()}`);
  },

  async restoreSkill(id: number): Promise<void> {
    await request<void>(`/skills/${id}/restore`, { method: 'POST' });
  },

  async permanentDeleteSkill(id: number): Promise<void> {
    await request<void>(`/skills/${id}/permanent`, { method: 'DELETE' });
  },

  async exportSkill(id: number): Promise<void> {
    try {
      const response = await fetch(`${API_BASE_URL}/skills/${id}/export`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = 'SKILL.md';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Export failed:', error);
      throw error;
    }
  },
};

// 任务API
export const jobtaskApi = {
  async getJobTasks(
    project?: string,
    jobType?: string,
    status?: string,
    page: number = 1,
    pageSize: number = 10,
    startDate?: number,
    endDate?: number
  ): Promise<PaginatedResponse<JobTask>> {
    let url = '/jobtasks';
    const params = new URLSearchParams();
    if (project && project !== '') params.append('project', project);
    if (jobType && jobType !== '') params.append('type', jobType);
    if (status && status !== '') params.append('status', status);
    if (startDate) params.append('startDate', startDate.toString());
    if (endDate) params.append('endDate', endDate.toString());
    params.append('page', page.toString());
    params.append('pageSize', pageSize.toString());

    const queryString = params.toString();
    if (queryString) url += `?${queryString}`;

    return request<PaginatedResponse<JobTask>>(url);
  },

  async getProjects(): Promise<string[]> {
    return request<string[]>('/jobtasks/projects');
  },

  async createJobTask(jobTask: JobTaskRequest): Promise<JobTask> {
    return request<JobTask>('/jobtasks', {
      method: 'POST',
      body: JSON.stringify(jobTask),
    });
  },

  async updateJobTask(id: number, jobTask: JobTaskRequest): Promise<JobTask> {
    return request<JobTask>(`/jobtasks/${id}`, {
      method: 'PUT',
      body: JSON.stringify(jobTask),
    });
  },

  async deleteJobTask(id: number): Promise<void> {
    await request<void>(`/jobtasks/${id}`, { method: 'DELETE' });
  },

  async getTrashJobTasks(page: number = 1, pageSize: number = 10): Promise<PaginatedResponse<JobTask>> {
    const params = new URLSearchParams();
    params.append('page', page.toString());
    params.append('pageSize', pageSize.toString());
    return request<PaginatedResponse<JobTask>>(`/jobtasks/trash?${params.toString()}`);
  },

  async restoreJobTask(id: number): Promise<void> {
    await request<void>(`/jobtasks/${id}/restore`, { method: 'POST' });
  },

  async permanentDeleteJobTask(id: number): Promise<void> {
    await request<void>(`/jobtasks/${id}/permanent`, { method: 'DELETE' });
  },

  async exportJobTasks(ids?: number[], format: 'csv' | 'json' | 'md' = 'csv'): Promise<void> {
    try {
      const response = await fetch(`${API_BASE_URL}/jobtasks/export`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ids, format }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const timestamp = new Date().toISOString().slice(0, 19).replace(/:/g, '-');
      const extensionMap = { csv: 'csv', json: 'json', md: 'md' } as const;
      const filename = `jobtasks_${timestamp}.${extensionMap[format]}`;

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Export failed:', error);
      throw error;
    }
  },
};

// 文件上传API
export const uploadApi = {
  async uploadFile(file: File, processType: string): Promise<{ success: boolean; message: string }> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('process_type', processType);

    try {
      const response = await fetch(`${API_BASE_URL}/upload_data`, {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      if (!data.success) {
        throw new Error(data.error || '上传失败');
      }

      return data;
    } catch (error) {
      console.error('Upload failed:', error);
      throw error;
    }
  },
};
```

**步骤 3：更新导入路径**

更新所有使用类型的文件，将导入路径从 `../services` 改为 `../types`：

```typescript
// 修改前
import type { Skill, Tag, JobTask } from '../services';

// 修改后
import type { Skill, Tag, JobTask } from '../types';
```

#### 1.4 验证步骤

```powershell
# 进入前端目录
cd d:\CodeHub\personal-project\aiflow\frontend

# 执行 TypeScript 编译检查
npx tsc --noEmit
```

---

### 任务 2：常量提取与统一管理

#### 2.1 问题描述

当前代码中存在大量魔法数字和硬编码值：

```typescript
// 问题代码示例
const MOBILE_BREAKPOINT = 768;  // 在 App.tsx 和页面组件中重复定义
const scrollTop = container.scrollTop > 300;  // 硬编码滚动阈值
```

#### 2.2 目标结构

更新 `types/constant.ts`：

```typescript
/**
 * 应用常量定义
 * 统一管理所有魔法数字和配置值
 */

// ==================== API 配置 ====================

/**
 * API 基础 URL
 * 构建模式下使用相对路径，开发模式下使用绝对路径
 */
export const API_BASE_URL = import.meta.env.PROD ? '/api' : 'http://localhost:9900/api';

// ==================== 响应式断点 ====================

/**
 * 移动端断点（像素）
 * 屏幕宽度小于此值时启用移动端布局
 */
export const MOBILE_BREAKPOINT = 768;

// ==================== 滚动相关 ====================

/**
 * 回到顶部按钮显示阈值（像素）
 * 滚动超过此距离时显示回到顶部按钮
 */
export const BACK_TO_TOP_THRESHOLD = 300;

/**
 * 筛选栏悬浮阈值（像素）
 */
export const FILTER_STICKY_THRESHOLD = 0;

// ==================== 分页配置 ====================

/**
 * 默认页码
 */
export const DEFAULT_PAGE = 1;

/**
 * 默认每页条数
 */
export const DEFAULT_PAGE_SIZE = 20;

/**
 * 可选的每页条数选项
 */
export const PAGE_SIZE_OPTIONS = [20, 50, 100] as const;

// ==================== 自动刷新配置 ====================

/**
 * 自动刷新间隔（毫秒）
 */
export const AUTO_REFRESH_INTERVAL = 5000;

// ==================== UI 配置 ====================

/**
 * 侧边栏展开宽度（像素）
 */
export const SIDER_WIDTH_EXPANDED = 200;

/**
 * 侧边栏收起宽度（像素）
 */
export const SIDER_WIDTH_COLLAPSED = 80;

/**
 * 头部高度（像素）
 */
export const HEADER_HEIGHT = 64;

// ==================== 时间格式 ====================

/**
 * 日期时间显示格式
 */
export const DATETIME_FORMAT = 'YYYY-MM-DD HH:mm:ss';

/**
 * 日期显示格式
 */
export const DATE_FORMAT = 'YYYY-MM-DD';

/**
 * 时间显示格式
 */
export const TIME_FORMAT = 'HH:mm';

// ==================== 标签显示 ====================

/**
 * 技能卡片最大显示标签数
 */
export const MAX_VISIBLE_TAGS = 3;

// ==================== 头像配置 ====================

/**
 * 技能头像渐变色配置
 */
export const SKILL_AVATAR_GRADIENTS = [
  'linear-gradient(135deg, #8b5cf6, #3b82f6)',
  'linear-gradient(135deg, #10b981, #3b82f6)',
  'linear-gradient(135deg, #f59e0b, #ef4444)',
  'linear-gradient(135deg, #ec4899, #8b5cf6)',
  'linear-gradient(135deg, #06b6d4, #3b82f6)',
  'linear-gradient(135deg, #84cc16, #10b981)',
] as const;
```

#### 2.3 实施步骤

**步骤 1：更新 types/index.ts 导出**

```typescript
// 在 types/index.ts 末尾添加
export * from './constant';
```

**步骤 2：替换页面组件中的魔法数字**

以 `SkillManagement.tsx` 为例：

```typescript
// 修改前
const MOBILE_BREAKPOINT = 768;
const [showBackToTop, setShowBackToTop] = useState(false);
// ...
setShowBackToTop(container.scrollTop > 300);

// 修改后
import {
  MOBILE_BREAKPOINT,
  BACK_TO_TOP_THRESHOLD,
  DEFAULT_PAGE,
  DEFAULT_PAGE_SIZE,
  PAGE_SIZE_OPTIONS,
} from '../types';

// 使用常量
setShowBackToTop(container.scrollTop > BACK_TO_TOP_THRESHOLD);
```

**步骤 3：更新分页配置**

```typescript
// 修改前
pageSizeOptions={['20', '50', '100']}

// 修改后
pageSizeOptions={PAGE_SIZE_OPTIONS.map(String)}
```

#### 2.4 验证步骤

```powershell
# 执行 TypeScript 编译检查
npx tsc --noEmit

# 搜索是否还有硬编码的魔法数字
# 检查 MOBILE_BREAKPOINT 是否还有重复定义
```

---

### 任务 3：工具函数提取

#### 3.1 问题描述

多个页面组件中存在重复的工具函数：

```typescript
// formatTime 在 SkillManagement.tsx 和 JobTaskManagement.tsx 中重复定义
const formatTime = (timestamp: number): string => { ... };

// getSkillAvatarConfig 在 SkillManagement.tsx 中定义
const getSkillAvatarConfig = (name: string): { ... } => { ... };

// getStatusConfig 在 JobTaskManagement.tsx 中定义
const getStatusConfig = (status: string): { ... } => { ... };
```

#### 3.2 目标结构

```
frontend/src/
└── utils/
    ├── index.ts          # 工具函数统一导出
    ├── format.ts         # 格式化相关函数
    ├── config.ts         # 配置获取函数
    └── clipboard.ts      # 剪贴板相关函数
```

#### 3.3 实施步骤

**步骤 1：创建 utils/format.ts**

```typescript
/**
 * 格式化相关工具函数
 */

import { DATETIME_FORMAT, DATE_FORMAT, TIME_FORMAT } from '../types';

/**
 * 格式化时间戳为本地时间字符串
 * @param timestamp - 时间戳（毫秒）
 * @returns 格式化后的时间字符串
 * @example
 * formatTime(1707753600000) // "2024-02-12 18:00:00"
 */
export const formatTime = (timestamp: number): string => {
  const date = new Date(timestamp);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
};

/**
 * 格式化时间戳为日期字符串
 * @param timestamp - 时间戳（毫秒）
 * @returns 格式化后的日期字符串
 */
export const formatDate = (timestamp: number): string => {
  const date = new Date(timestamp);
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  });
};

/**
 * 格式化文件大小
 * @param bytes - 字节数
 * @returns 格式化后的文件大小字符串
 */
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B';

  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const k = 1024;
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${units[i]}`;
};
```

**步骤 2：创建 utils/config.ts**

```typescript
/**
 * 配置获取相关工具函数
 */

import { SKILL_AVATAR_GRADIENTS, MAX_VISIBLE_TAGS } from '../types';

/**
 * 技能头像配置
 */
export interface SkillAvatarConfig {
  /** 渐变背景色 */
  gradient: string;
  /** 显示的图标/字符 */
  icon: string;
}

/**
 * 获取技能头像样式配置
 * @param name - 技能名称
 * @returns 头像配置对象
 */
export const getSkillAvatarConfig = (name: string): SkillAvatarConfig => {
  const firstChar = name.charAt(0).toUpperCase();
  const index = name.length % SKILL_AVATAR_GRADIENTS.length;
  return {
    gradient: SKILL_AVATAR_GRADIENTS[index],
    icon: firstChar,
  };
};

/**
 * 状态配置
 */
export interface StatusConfig {
  /** 状态颜色 */
  color: string;
  /** 状态图标 */
  icon: React.ReactNode;
  /** CSS 类名 */
  className: string;
}

/**
 * 获取任务状态对应的样式配置
 * @param status - 任务状态
 * @returns 状态配置对象
 */
export const getStatusConfig = (status: string): StatusConfig => {
  const configMap: Record<string, StatusConfig> = {
    已创建: {
      color: '#6b7280',
      icon: 'ClockCircleOutlined',
      className: 'status-created',
    },
    处理中: {
      color: '#3b82f6',
      icon: 'SyncOutlined',
      className: 'status-processing',
    },
    处理失败: {
      color: '#ef4444',
      icon: 'ExclamationCircleOutlined',
      className: 'status-failed',
    },
    处理完成: {
      color: '#10b981',
      icon: 'CheckCircleOutlined',
      className: 'status-success',
    },
    验收通过: {
      color: '#8b5cf6',
      icon: 'SafetyCertificateOutlined',
      className: 'status-accepted',
    },
  };

  return configMap[status] || {
    color: '#6b7280',
    icon: 'ClockCircleOutlined',
    className: 'status-created',
  };
};

/**
 * 任务类型配置
 */
export interface TypeConfig {
  /** 类型颜色 */
  color: string;
  /** 背景颜色 */
  bgColor: string;
  /** Emoji 图标 */
  emoji: string;
}

/**
 * 获取任务类型对应的样式配置
 * @param type - 任务类型
 * @returns 类型配置对象
 */
export const getTypeConfig = (type: string): TypeConfig => {
  const configMap: Record<string, TypeConfig> = {
    新需求: {
      color: '#3b82f6',
      bgColor: 'rgba(59, 130, 246, 0.1)',
      emoji: '✨',
    },
    Bug修复: {
      color: '#ef4444',
      bgColor: 'rgba(239, 68, 68, 0.1)',
      emoji: '🐛',
    },
    改进功能: {
      color: '#10b981',
      bgColor: 'rgba(16, 185, 129, 0.1)',
      emoji: '🚀',
    },
    重构代码: {
      color: '#8b5cf6',
      bgColor: 'rgba(139, 92, 246, 0.1)',
      emoji: '🔧',
    },
    单元测试: {
      color: '#f59e0b',
      bgColor: 'rgba(245, 158, 11, 0.1)',
      emoji: '🧪',
    },
    集成测试: {
      color: '#f59e0b',
      bgColor: 'rgba(245, 158, 11, 0.1)',
      emoji: '🔨',
    },
    数据处理: {
      color: '#06b6d4',
      bgColor: 'rgba(6, 182, 212, 0.1)',
      emoji: '📊',
    },
    版本控制: {
      color: '#ec4899',
      bgColor: 'rgba(236, 72, 153, 0.1)',
      emoji: '📝',
    },
  };

  return configMap[type] || {
    color: '#6b7280',
    bgColor: 'rgba(107, 114, 128, 0.1)',
    emoji: '📋',
  };
};

/**
 * 获取状态对应的颜色（用于执行记录）
 * @param status - 状态
 * @returns 颜色值
 */
export const getRecordStatusColor = (status: string): string => {
  const colorMap: Record<string, string> = {
    处理完成: '#10b981',
    处理失败: '#ef4444',
    处理中: '#3b82f6',
    已创建: '#6b7280',
    验收通过: '#8b5cf6',
  };
  return colorMap[status] || '#6b7280';
};
```

**步骤 3：创建 utils/clipboard.ts**

```typescript
/**
 * 剪贴板相关工具函数
 */

import { message } from 'antd';

/**
 * 复制文本到剪贴板
 * @param text - 要复制的文本
 * @param successMsg - 成功提示消息
 */
export const copyToClipboard = async (
  text: string,
  successMsg: string = '已复制到剪贴板',
): Promise<void> => {
  /**
   * 降级复制方案（用于不支持 Clipboard API 的环境）
   */
  const fallbackCopy = (): boolean => {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();

    const successful = document.execCommand('copy');
    document.body.removeChild(textArea);
    return successful;
  };

  try {
    // 检查是否在 iframe 中（VSCode Webview 环境）
    if (window.parent !== window) {
      window.parent.postMessage({
        command: 'copyToClipboard',
        text: text,
      }, '*');

      const result = await Promise.race([
        new Promise<{ success: boolean; error?: string }>((resolve) => {
          const handler = (event: MessageEvent) => {
            if (event.data && event.data.command === 'copyToClipboardResult') {
              window.removeEventListener('message', handler);
              resolve({ success: event.data.success, error: event.data.error });
            }
          };
          window.addEventListener('message', handler);
        }),
        new Promise<{ success: boolean }>((_, reject) =>
          setTimeout(() => reject(new Error('timeout')), 1000),
        ),
      ]);

      if (result.success) {
        message.success(successMsg);
      } else {
        throw new Error('复制失败');
      }
      return;
    }

    // 尝试使用降级方案
    if (fallbackCopy()) {
      message.success(successMsg);
    } else {
      message.error('复制失败');
    }
  } catch (err) {
    // 最终降级方案
    if (fallbackCopy()) {
      message.success(successMsg);
    } else {
      message.error('复制失败');
    }
  }
};
```

**步骤 4：创建 utils/index.ts**

```typescript
/**
 * 工具函数统一导出
 */

// 格式化函数
export { formatTime, formatDate, formatFileSize } from './format';

// 配置获取函数
export {
  getSkillAvatarConfig,
  getStatusConfig,
  getTypeConfig,
  getRecordStatusColor,
} from './config';
export type { SkillAvatarConfig, StatusConfig, TypeConfig } from './config';

// 剪贴板函数
export { copyToClipboard } from './clipboard';
```

**步骤 5：更新页面组件导入**

```typescript
// 修改前
const formatTime = (timestamp: number): string => { ... };
const getSkillAvatarConfig = (name: string): { ... } => { ... };

// 修改后
import { formatTime, getSkillAvatarConfig } from '../utils';
```

#### 3.4 验证步骤

```powershell
# 执行 TypeScript 编译检查
npx tsc --noEmit
```

---

## 四、阶段 P1：组件优化

### 任务 4：公共组件提取

#### 4.1 问题描述

`StatCard` 组件在 `SkillManagement.tsx` 和 `JobTaskManagement.tsx` 中重复定义：

```typescript
// SkillManagement.tsx 中的定义
const StatCard: React.FC<{
  title: string;
  value: number | string;
  icon: React.ReactNode;
  color: string;
  subtitle?: string;
}> = ({ title, value, icon, color, subtitle }) => ( ... );

// JobTaskManagement.tsx 中的定义（略有不同）
const StatCard: React.FC<{
  title: string;
  value: number;
  icon: React.ReactNode;
  color: string;
}> = ({ title, value, icon, color }) => ( ... );
```

#### 4.2 目标结构

```
frontend/src/
└── components/
    ├── common/
    │   ├── StatCard.tsx       # 统计卡片组件
    │   ├── BackToTop.tsx      # 回到顶部组件
    │   └── index.ts           # 公共组件导出
    ├── SkillModal.tsx
    ├── JobTaskModal.tsx
    └── ...
```

#### 4.3 实施步骤

**步骤 1：创建 components/common/StatCard.tsx**

```typescript
/**
 * 统计卡片组件
 * 用于展示统计数据，支持标题、数值、图标和副标题
 */

import React from 'react';
import { Card } from 'antd';

/**
 * 统计卡片属性
 */
export interface StatCardProps {
  /** 标题 */
  title: string;
  /** 数值（支持数字或字符串） */
  value: number | string;
  /** 图标 */
  icon: React.ReactNode;
  /** 主题颜色 */
  color: string;
  /** 副标题（可选） */
  subtitle?: string;
}

/**
 * 统计卡片组件
 * @example
 * <StatCard
 *   title="技能总数"
 *   value={42}
 *   icon={<FileTextOutlined />}
 *   color="#8b5cf6"
 * />
 */
export const StatCard: React.FC<StatCardProps> = ({
  title,
  value,
  icon,
  color,
  subtitle,
}) => (
  <Card className="stat-card" variant="borderless">
    <div className="stat-content">
      <div>
        <div className="stat-title">{title}</div>
        <div className="stat-value" style={{ color }}>
          {value}
        </div>
        {subtitle && <div className="stat-subtitle">{subtitle}</div>}
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
```

**步骤 2：创建 components/common/BackToTop.tsx**

```typescript
/**
 * 回到顶部组件
 * 提供页面滚动回顶部的功能按钮
 */

import React from 'react';
import { VerticalAlignTopOutlined } from '@ant-design/icons';
import { BACK_TO_TOP_THRESHOLD } from '../../types';

/**
 * 回到顶部组件属性
 */
export interface BackToTopProps {
  /** 是否显示按钮 */
  visible: boolean;
  /** 点击回调 */
  onClick: () => void;
  /** 自定义提示文本 */
  title?: string;
}

/**
 * 回到顶部组件
 * @example
 * <BackToTop visible={showBackToTop} onClick={handleBackToTop} />
 */
export const BackToTop: React.FC<BackToTopProps> = ({
  visible,
  onClick,
  title = '回到顶部',
}) => {
  if (!visible) return null;

  return (
    <button
      className="back-to-top-btn"
      onClick={onClick}
      title={title}
      aria-label={title}
    >
      <VerticalAlignTopOutlined />
    </button>
  );
};

/**
 * 使用回到顶部功能的 Hook
 * @param threshold - 显示阈值（像素）
 * @returns [visible, scrollToTop]
 */
export const useBackToTop = (
  threshold: number = BACK_TO_TOP_THRESHOLD,
): [boolean, () => void] => {
  const [visible, setVisible] = React.useState(false);

  React.useEffect(() => {
    const handleScroll = () => {
      setVisible(document.body.scrollTop > threshold);
    };

    document.body.addEventListener('scroll', handleScroll);
    handleScroll();

    return () => {
      document.body.removeEventListener('scroll', handleScroll);
    };
  }, [threshold]);

  const scrollToTop = React.useCallback(() => {
    document.body.scrollTo({
      top: 0,
      behavior: 'smooth',
    });
  }, []);

  return [visible, scrollToTop];
};
```

**步骤 3：创建 components/common/index.ts**

```typescript
/**
 * 公共组件统一导出
 */

export { StatCard } from './StatCard';
export type { StatCardProps } from './StatCard';

export { BackToTop, useBackToTop } from './BackToTop';
export type { BackToTopProps } from './BackToTop';
```

**步骤 4：更新页面组件**

```typescript
// 修改前
const StatCard: React.FC<{ ... }> = ({ ... }) => ( ... );

// 修改后
import { StatCard, BackToTop, useBackToTop } from '../components/common';

// 使用 Hook
const [showBackToTop, handleBackToTop] = useBackToTop();

// 在 JSX 中
<BackToTop visible={showBackToTop} onClick={handleBackToTop} />
```

#### 4.4 验证步骤

```powershell
# 执行 TypeScript 编译检查
npx tsc --noEmit
```

---

### 任务 5：自定义 Hooks 提取

#### 5.1 问题描述

窗口宽度监听、筛选栏悬浮等逻辑在多个组件中重复：

```typescript
// 重复代码 1：窗口宽度监听
const [windowWidth, setWindowWidth] = useState(window.innerWidth);
useEffect(() => {
  const handleResize = () => setWindowWidth(window.innerWidth);
  window.addEventListener("resize", handleResize);
  return () => window.removeEventListener("resize", handleResize);
}, []);
const isMobile = windowWidth < MOBILE_BREAKPOINT;

// 重复代码 2：筛选栏悬浮
const [isFilterSticky, setIsFilterSticky] = useState(false);
const sentinelRef = useRef<HTMLDivElement>(null);
useEffect(() => {
  const sentinel = sentinelRef.current;
  if (!sentinel) return;
  const observer = new IntersectionObserver(...);
  observer.observe(sentinel);
  return () => observer.disconnect();
}, []);
```

#### 5.2 目标结构

```
frontend/src/
└── hooks/
    ├── index.ts              # Hooks 统一导出
    ├── useWindowWidth.ts     # 窗口宽度监听
    ├── useFilterSticky.ts    # 筛选栏悬浮
    └── useDebounce.ts        # 防抖 Hook
```

#### 5.3 实施步骤

**步骤 1：创建 hooks/useWindowWidth.ts**

```typescript
/**
 * 窗口宽度监听 Hook
 * 提供响应式布局所需的窗口宽度信息
 */

import { useState, useEffect } from 'react';
import { MOBILE_BREAKPOINT } from '../types';

/**
 * 窗口宽度信息
 */
export interface WindowWidthInfo {
  /** 当前窗口宽度（像素） */
  width: number;
  /** 是否为移动端 */
  isMobile: boolean;
}

/**
 * 监听窗口宽度变化
 * @returns 窗口宽度信息
 * @example
 * const { width, isMobile } = useWindowWidth();
 */
export const useWindowWidth = (): WindowWidthInfo => {
  const [width, setWidth] = useState(window.innerWidth);

  useEffect(() => {
    const handleResize = () => {
      setWidth(window.innerWidth);
    };

    window.addEventListener('resize', handleResize);
    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, []);

  return {
    width,
    isMobile: width < MOBILE_BREAKPOINT,
  };
};
```

**步骤 2：创建 hooks/useFilterSticky.ts**

```typescript
/**
 * 筛选栏悬浮 Hook
 * 使用 IntersectionObserver 实现筛选栏吸顶效果
 */

import { useState, useEffect, useRef, useCallback } from 'react';

/**
 * 筛选栏悬浮状态
 */
export interface FilterStickyState {
  /** 是否处于悬浮状态 */
  isSticky: boolean;
  /** 哨兵元素 ref */
  sentinelRef: React.RefObject<HTMLDivElement>;
}

/**
 * 监听筛选栏悬浮状态
 * @returns 悬浮状态和哨兵元素 ref
 * @example
 * const { isSticky, sentinelRef } = useFilterSticky();
 * // JSX:
 * // <div ref={sentinelRef} className="filter-sentinel" />
 * // <div className={`filter-bar ${isSticky ? 'sticky' : ''}`}>
 */
export const useFilterSticky = (): FilterStickyState => {
  const [isSticky, setIsSticky] = useState(false);
  const sentinelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        setIsSticky(!entry.isIntersecting);
      },
      {
        root: null,
        threshold: 0,
        rootMargin: '0px 0px 0px 0px',
      },
    );

    observer.observe(sentinel);

    return () => {
      observer.disconnect();
    };
  }, []);

  return {
    isSticky,
    sentinelRef,
  };
};
```

**步骤 3：创建 hooks/useDebounce.ts**

```typescript
/**
 * 防抖 Hook
 * 延迟执行值更新，适用于搜索输入等场景
 */

import { useState, useEffect } from 'react';

/**
 * 防抖值
 * @param value - 原始值
 * @param delay - 延迟时间（毫秒）
 * @returns 防抖后的值
 * @example
 * const [searchTerm, setSearchTerm] = useState('');
 * const debouncedSearchTerm = useDebounce(searchTerm, 300);
 * // 使用 debouncedSearchTerm 进行 API 调用
 */
export const useDebounce = <T>(value: T, delay: number): T => {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(timer);
    };
  }, [value, delay]);

  return debouncedValue;
};
```

**步骤 4：创建 hooks/index.ts**

```typescript
/**
 * 自定义 Hooks 统一导出
 */

export { useWindowWidth } from './useWindowWidth';
export type { WindowWidthInfo } from './useWindowWidth';

export { useFilterSticky } from './useFilterSticky';
export type { FilterStickyState } from './useFilterSticky';

export { useDebounce } from './useDebounce';
```

**步骤 5：更新页面组件**

```typescript
// 修改前
const [windowWidth, setWindowWidth] = useState(window.innerWidth);
useEffect(() => { ... }, []);
const isMobile = windowWidth < MOBILE_BREAKPOINT;

// 修改后
import { useWindowWidth, useFilterSticky } from '../hooks';

const { isMobile } = useWindowWidth();
const { isSticky, sentinelRef } = useFilterSticky();
```

#### 5.4 验证步骤

```powershell
# 执行 TypeScript 编译检查
npx tsc --noEmit
```

---

### 任务 6：页面组件瘦身

#### 6.1 问题描述

页面组件（如 `SkillManagement.tsx`）过于庞大，包含过多内部组件和逻辑：

- 文件行数超过 700 行
- 内部定义了 `SkillCard`、`StatCard` 等组件
- 业务逻辑和 UI 渲染混合

#### 6.2 目标结构

```
frontend/src/
└── pages/
    └── SkillManagement/
        ├── index.tsx              # 页面主组件
        ├── SkillCard.tsx          # 技能卡片组件
        ├── SkillFilterBar.tsx     # 筛选栏组件
        └── useSkillData.ts        # 数据加载 Hook
```

#### 6.3 实施步骤

**步骤 1：创建 SkillCard 组件**

`pages/SkillManagement/SkillCard.tsx`：

```typescript
/**
 * 技能卡片组件
 */

import React from 'react';
import { Popconfirm } from 'antd';
import { EditOutlined, DeleteOutlined, DownloadOutlined, RollbackOutlined, CloseCircleOutlined } from '@ant-design/icons';
import type { Skill } from '../../types';
import { formatTime, getSkillAvatarConfig } from '../../utils';
import { MAX_VISIBLE_TAGS } from '../../types';

/**
 * 技能卡片属性
 */
export interface SkillCardProps {
  /** 技能数据 */
  skill: Skill;
  /** 编辑回调 */
  onEdit: (skill: Skill) => void;
  /** 删除回调 */
  onDelete: (id: number) => void;
  /** 导出回调 */
  onExport: (id: number) => void;
  /** 是否为回收站模式 */
  isTrashMode?: boolean;
  /** 恢复回调 */
  onRestore?: (id: number, onSuccess?: () => void) => void;
  /** 彻底删除回调 */
  onPermanentDelete?: (id: number, onSuccess?: () => void) => void;
  /** 刷新回调 */
  onRefresh?: () => void;
}

/**
 * 技能卡片组件
 */
export const SkillCard: React.FC<SkillCardProps> = ({
  skill,
  onEdit,
  onDelete,
  onExport,
  isTrashMode = false,
  onRestore,
  onPermanentDelete,
  onRefresh,
}) => {
  const avatarConfig = getSkillAvatarConfig(skill.name);

  return (
    <div className="skill-card">
      <div className="skill-card-header">
        <div className="skill-identity">
          <div
            className="skill-avatar"
            style={{ background: avatarConfig.gradient }}
          >
            {avatarConfig.icon}
          </div>
          <div className="skill-title-group">
            <span className="skill-name">{skill.name}</span>
            <span className="skill-dir">{skill.resourceDir}</span>
          </div>
        </div>
        <div className="skill-tags">
          {skill.tags?.slice(0, MAX_VISIBLE_TAGS).map((tag) => (
            <span key={tag.id} className="skill-tag">
              {tag.name}
            </span>
          ))}
          {skill.tags && skill.tags.length > MAX_VISIBLE_TAGS && (
            <span className="skill-tag more">
              +{skill.tags.length - MAX_VISIBLE_TAGS}
            </span>
          )}
        </div>
      </div>

      <div className="skill-description">{skill.description}</div>

      {skill.compatibility && (
        <div className="compatibility-alert">
          <span>⚡</span>
          <span>{skill.compatibility}</span>
        </div>
      )}

      <div className="skill-footer">
        <div className="skill-time">创建于 {formatTime(skill.createdAt)}</div>
        <div className="skill-actions">
          {!isTrashMode ? (
            <>
              <button
                className="btn-icon"
                title="编辑"
                onClick={() => onEdit(skill)}
              >
                <EditOutlined />
              </button>
              <button
                className="btn-icon export"
                title="导出"
                onClick={() => onExport(skill.id)}
              >
                <DownloadOutlined />
              </button>
              <Popconfirm
                title="确定要删除这个技能吗？删除后可在回收站恢复。"
                onConfirm={() => onDelete(skill.id)}
                okText="确定"
                cancelText="取消"
                placement="topRight"
              >
                <button className="btn-icon delete" title="删除">
                  <DeleteOutlined />
                </button>
              </Popconfirm>
            </>
          ) : (
            <>
              <Popconfirm
                title="确定要恢复这个技能吗？"
                onConfirm={() => onRestore?.(skill.id, onRefresh)}
                okText="确定"
                cancelText="取消"
                placement="topRight"
              >
                <button
                  className="btn-icon"
                  title="恢复"
                  style={{ color: '#10b981' }}
                >
                  <RollbackOutlined />
                </button>
              </Popconfirm>
              <Popconfirm
                title="确定要彻底删除这个技能吗？删除后无法恢复！"
                onConfirm={() => onPermanentDelete?.(skill.id, onRefresh)}
                okText="确定"
                cancelText="取消"
                placement="topRight"
              >
                <button className="btn-icon delete" title="彻底删除">
                  <CloseCircleOutlined />
                </button>
              </Popconfirm>
            </>
          )}
        </div>
      </div>
    </div>
  );
};
```

**步骤 2：创建数据加载 Hook**

`pages/SkillManagement/useSkillData.ts`：

```typescript
/**
 * 技能数据加载 Hook
 */

import { useState, useCallback, useEffect } from 'react';
import { message } from 'antd';
import type { Dayjs } from 'dayjs';
import type { Skill, Tag, Pagination } from '../../types';
import { skillApi, tagApi } from '../../services';
import { DEFAULT_PAGE, DEFAULT_PAGE_SIZE } from '../../types';

/**
 * 技能数据状态
 */
export interface SkillDataState {
  /** 技能列表 */
  skills: Skill[];
  /** 标签列表 */
  tags: Tag[];
  /** 分页信息 */
  pagination: Pagination;
  /** 加载状态 */
  loading: boolean;
}

/**
 * 技能数据操作
 */
export interface SkillDataActions {
  /** 加载技能数据 */
  loadSkills: (tagId?: number, dateRange?: [Dayjs | null, Dayjs | null] | null, page?: number, pageSize?: number) => Promise<void>;
  /** 加载回收站数据 */
  loadTrashSkills: (page?: number, pageSize?: number) => Promise<void>;
  /** 删除技能 */
  deleteSkill: (id: number) => Promise<void>;
  /** 导出技能 */
  exportSkill: (id: number) => Promise<void>;
  /** 恢复技能 */
  restoreSkill: (id: number, onSuccess?: () => void) => Promise<void>;
  /** 彻底删除技能 */
  permanentDeleteSkill: (id: number, onSuccess?: () => void) => Promise<void>;
}

/**
 * 使用技能数据
 */
export const useSkillData = (): SkillDataState & SkillDataActions => {
  const [skills, setSkills] = useState<Skill[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [pagination, setPagination] = useState<Pagination>({
    page: DEFAULT_PAGE,
    pageSize: DEFAULT_PAGE_SIZE,
    total: 0,
  });
  const [loading, setLoading] = useState(false);

  const loadSkills = useCallback(async (
    tagId?: number,
    dateRange?: [Dayjs | null, Dayjs | null] | null,
    page: number = DEFAULT_PAGE,
    pageSize: number = DEFAULT_PAGE_SIZE,
  ) => {
    setLoading(true);
    try {
      const startDate = dateRange?.[0]?.valueOf();
      const endDate = dateRange?.[1]?.valueOf();
      const [loadedSkills, loadedTags] = await Promise.all([
        skillApi.getSkills(tagId, page, pageSize, startDate, endDate),
        tagApi.getTags(1, 100),
      ]);
      setSkills(loadedSkills.items);
      setTags(loadedTags.items);
      setPagination((prev) => ({
        ...prev,
        page,
        pageSize,
        total: loadedSkills.pagination.total,
      }));
    } catch (error) {
      message.error('加载技能数据失败');
      console.error('Load skill data failed:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  const loadTrashSkills = useCallback(async (
    page: number = DEFAULT_PAGE,
    pageSize: number = DEFAULT_PAGE_SIZE,
  ) => {
    setLoading(true);
    try {
      const loadedTrashSkills = await skillApi.getTrashSkills(page, pageSize);
      setSkills(loadedTrashSkills.items);
      setPagination((prev) => ({
        ...prev,
        page,
        pageSize,
        total: loadedTrashSkills.pagination.total,
      }));
    } catch (error) {
      message.error('加载回收站数据失败');
      console.error('Load trash skills failed:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  const deleteSkill = useCallback(async (id: number) => {
    try {
      await skillApi.deleteSkill(id);
      setSkills((prev) => prev.filter((skill) => skill.id !== id));
      message.success('技能已移至回收站');
    } catch (error) {
      message.error('删除失败，请重试');
      console.error('Skill delete failed:', error);
    }
  }, []);

  const exportSkill = useCallback(async (id: number) => {
    try {
      await skillApi.exportSkill(id);
    } catch (error) {
      message.error('导出失败，请重试');
      console.error('Skill export failed:', error);
    }
  }, []);

  const restoreSkill = useCallback(async (id: number, onSuccess?: () => void) => {
    try {
      await skillApi.restoreSkill(id);
      message.success('技能恢复成功');
      onSuccess?.();
    } catch (error) {
      message.error('恢复失败，请重试');
      console.error('Skill restore failed:', error);
    }
  }, []);

  const permanentDeleteSkill = useCallback(async (id: number, onSuccess?: () => void) => {
    try {
      await skillApi.permanentDeleteSkill(id);
      message.success('技能已彻底删除');
      onSuccess?.();
    } catch (error) {
      message.error('删除失败，请重试');
      console.error('Skill permanent delete failed:', error);
    }
  }, []);

  return {
    skills,
    tags,
    pagination,
    loading,
    loadSkills,
    loadTrashSkills,
    deleteSkill,
    exportSkill,
    restoreSkill,
    permanentDeleteSkill,
  };
};
```

**步骤 3：重构页面主组件**

`pages/SkillManagement/index.tsx`：

```typescript
/**
 * 技能管理页面
 */

import React, { useState, useEffect, useCallback } from 'react';
import { Button, Select, Empty, Pagination, Row, Col, Space, DatePicker } from 'antd';
import type { Dayjs } from 'dayjs';
import {
  PlusOutlined,
  UploadOutlined,
  AppstoreOutlined,
  TagOutlined,
  ClockCircleOutlined,
  ReloadOutlined,
  RestOutlined,
  FileTextOutlined,
} from '@ant-design/icons';
import { useAppStore } from '../../stores/appStore';
import { useModalStore } from '../../stores/modalStore';
import { useWindowWidth, useFilterSticky, useBackToTop } from '../../hooks';
import { StatCard, BackToTop } from '../../components/common';
import { SkillCard } from './SkillCard';
import { useSkillData } from './useSkillData';
import { PAGE_SIZE_OPTIONS } from '../../types';

const { RangePicker } = DatePicker;

/**
 * 技能管理页面组件
 */
const SkillManagement: React.FC = () => {
  // 全局状态
  const {
    selectedTagId,
    selectedSkillDateRange,
    skillPagination,
    collapsed,
    setSelectedTagId,
    setSelectedSkillDateRange,
    setSkillPagination,
  } = useAppStore();
  const { openSkillModal, openUploadModal, openTagManagementModal } = useModalStore();

  // 自定义 Hooks
  const { isMobile } = useWindowWidth();
  const { isSticky, sentinelRef } = useFilterSticky();
  const [showBackToTop, handleBackToTop] = useBackToTop();

  // 数据加载
  const {
    skills,
    tags,
    pagination,
    loadSkills,
    loadTrashSkills,
    deleteSkill,
    exportSkill,
    restoreSkill,
    permanentDeleteSkill,
  } = useSkillData();

  // 本地状态
  const [isTrashMode, setIsTrashMode] = useState(false);

  // 初始加载
  useEffect(() => {
    loadSkills(selectedTagId, selectedSkillDateRange, skillPagination.page, skillPagination.pageSize);
  }, []);

  // 切换回收站模式
  useEffect(() => {
    if (isTrashMode) {
      loadTrashSkills(skillPagination.page, skillPagination.pageSize);
    } else {
      loadSkills(selectedTagId, selectedSkillDateRange, skillPagination.page, skillPagination.pageSize);
    }
  }, [isTrashMode]);

  // 事件处理
  const handlePageChange = (page: number, pageSize: number) => {
    setSkillPagination({ page, pageSize });
    if (isTrashMode) {
      loadTrashSkills(page, pageSize);
    } else {
      loadSkills(selectedTagId, selectedSkillDateRange, page, pageSize);
    }
  };

  const handleTagChange = (tagId: number | undefined) => {
    setSelectedTagId(tagId);
    setSkillPagination({ page: 1 });
    loadSkills(tagId, selectedSkillDateRange, 1, pagination.pageSize);
  };

  const handleDateRangeChange = (dates: [Dayjs | null, Dayjs | null] | null) => {
    setSelectedSkillDateRange(dates);
    setSkillPagination({ page: 1 });
    loadSkills(selectedTagId, dates, 1, pagination.pageSize);
  };

  const handleResetFilters = () => {
    setSelectedTagId(undefined);
    setSelectedSkillDateRange(null);
    setSkillPagination({ page: 1 });
    loadSkills(undefined, null, 1, pagination.pageSize);
  };

  const refreshTrashData = useCallback(() => {
    if (isTrashMode) {
      loadTrashSkills(pagination.page, pagination.pageSize);
    }
  }, [isTrashMode, loadTrashSkills, pagination.page, pagination.pageSize]);

  return (
    <div className="skill-management">
      {/* 头部区域 */}
      <div className="skill-header">
        <div className="header-top">
          <h1 className="header-title">{isTrashMode ? '回收站' : '技能管理'}</h1>
          <Space className="header-actions">
            <Button
              className="btn-secondary-gradient"
              icon={<RestOutlined />}
              onClick={() => setIsTrashMode(!isTrashMode)}
            >
              {isTrashMode ? '返回列表' : '回收站'}
            </Button>
            {!isTrashMode && (
              <>
                <Button
                  className="btn-secondary-gradient"
                  icon={<UploadOutlined />}
                  onClick={openUploadModal}
                >
                  导入
                </Button>
                <Button
                  type="primary"
                  className="btn-primary-gradient"
                  icon={<PlusOutlined />}
                  onClick={() => openSkillModal()}
                >
                  新增技能
                </Button>
              </>
            )}
          </Space>
        </div>
      </div>

      {/* 筛选栏 */}
      {!isTrashMode && (
        <div
          className={`filter-bar-wrapper ${isSticky && !isMobile ? 'filter-bar-wrapper-sticky' : ''}`}
          style={{ '--sider-width': collapsed ? '80px' : '200px' } as React.CSSProperties}
        >
          <div ref={sentinelRef} className="filter-sentinel" />
          <div className="filter-bar">
            <div className="filter-group">
              <span className="filter-label">标签</span>
              <Select
                className="filter-select"
                placeholder="选择标签"
                value={selectedTagId}
                onChange={(value) => handleTagChange(value || undefined)}
                options={[
                  { value: '', label: '全部标签' },
                  ...tags.map((tag) => ({ value: tag.id, label: tag.name })),
                ]}
                variant="borderless"
              />
            </div>
            <div className="filter-group">
              <span className="filter-label">创建时间</span>
              <RangePicker
                value={selectedSkillDateRange}
                onChange={handleDateRangeChange}
                placeholder={['开始时间', '结束时间']}
                format="YYYY-MM-DD HH:mm"
                showTime={{ format: 'HH:mm' }}
              />
            </div>
            <Button
              className="btn-secondary-gradient"
              icon={<AppstoreOutlined />}
              onClick={openTagManagementModal}
            >
              管理标签
            </Button>
            <Button
              className="btn-secondary-gradient"
              icon={<ReloadOutlined />}
              onClick={handleResetFilters}
            >
              重置
            </Button>
          </div>
        </div>
      )}

      {/* 统计卡片 */}
      {!isTrashMode && (
        <Row gutter={[20, 20]} className="stats-row">
          <Col xs={12} sm={8} lg={8}>
            <StatCard
              title="技能总数"
              value={pagination.total}
              icon={<FileTextOutlined />}
              color="#8b5cf6"
            />
          </Col>
          <Col xs={12} sm={8} lg={8}>
            <StatCard
              title="活跃标签"
              value={tags.length}
              icon={<TagOutlined />}
              color="#10b981"
            />
          </Col>
          <Col xs={12} sm={8} lg={8}>
            <StatCard
              title="当前展示"
              value={skills.length}
              icon={<ClockCircleOutlined />}
              color="#f59e0b"
            />
          </Col>
        </Row>
      )}

      {/* 技能卡片网格 */}
      {skills.length > 0 ? (
        <>
          <div className="skills-grid">
            {skills.map((skill) => (
              <SkillCard
                key={skill.id}
                skill={skill}
                onEdit={openSkillModal}
                onDelete={deleteSkill}
                onExport={exportSkill}
                isTrashMode={isTrashMode}
                onRestore={restoreSkill}
                onPermanentDelete={permanentDeleteSkill}
                onRefresh={refreshTrashData}
              />
            ))}
          </div>

          {/* 分页 */}
          <div className="pagination-wrapper">
            <Pagination
              total={pagination.total}
              current={pagination.page}
              pageSize={pagination.pageSize}
              showSizeChanger
              showQuickJumper
              onChange={handlePageChange}
              pageSizeOptions={PAGE_SIZE_OPTIONS.map(String)}
              locale={{
                items_per_page: '条/页',
                jump_to: '跳至',
                jump_to_confirm: '确定',
                page: '页',
              }}
            />
          </div>
        </>
      ) : (
        <Empty
          className="skill-empty"
          description={isTrashMode ? '回收站为空' : '暂无技能数据'}
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
      )}

      {/* 回到顶部 */}
      <BackToTop visible={showBackToTop} onClick={handleBackToTop} />
    </div>
  );
};

export default SkillManagement;
```

#### 6.4 验证步骤

```powershell
# 执行 TypeScript 编译检查
npx tsc --noEmit
```

---

## 五、阶段 P2：样式优化

### 任务 7：CSS 样式拆分

#### 7.1 问题描述

`App.css` 文件超过 4000 行，包含所有样式，难以维护：

- 设计 Tokens 与组件样式混合
- 不同模块的样式耦合
- 响应式样式分散

#### 7.2 目标结构

```
frontend/src/
└── styles/
    ├── index.css           # 样式入口（导入所有模块）
    ├── tokens.css          # 设计 Tokens（CSS 变量）
    ├── base.css            # 基础样式重置
    ├── layout.css          # 布局相关（侧边栏、头部等）
    ├── components.css      # 公共组件样式
    ├── skill.css           # 技能管理页面样式
    ├── jobtask.css         # 任务管理页面样式
    ├── modal.css           # 弹窗样式
    └── responsive.css      # 响应式样式
```

#### 7.3 实施步骤

**步骤 1：创建 styles/tokens.css**

```css
/**
 * 设计 Tokens
 * 定义所有 CSS 变量，统一管理设计系统
 */

:root {
  /* ==================== 颜色系统 ==================== */

  /* 主色调 */
  --color-primary: #3b82f6;
  --color-primary-hover: #2563eb;
  --color-secondary: #8b5cf6;
  --color-success: #10b981;
  --color-warning: #f59e0b;
  --color-error: #ef4444;

  /* 中性色 */
  --color-gray-50: #f9fafb;
  --color-gray-100: #f3f4f6;
  --color-gray-200: #e5e7eb;
  --color-gray-300: #d1d5db;
  --color-gray-400: #9ca3af;
  --color-gray-500: #6b7280;
  --color-gray-600: #4b5563;
  --color-gray-700: #374151;
  --color-gray-800: #1f2937;
  --color-gray-900: #111827;

  /* 侧边栏 */
  --color-sider-bg: #1E293B;
  --color-sider-hover: #2D3748;
  --color-menu-text: #CBD5E1;
  --color-menu-selected: #3B82F6;

  /* ==================== 圆角 ==================== */

  --radius-sm: 6px;
  --radius-md: 8px;
  --radius-lg: 12px;
  --radius-xl: 16px;
  --radius-2xl: 20px;

  /* ==================== 阴影 ==================== */

  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.07);
  --shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.1);
  --shadow-xl: 0 20px 25px rgba(0, 0, 0, 0.1);
  --shadow-card: 0 4px 20px rgba(0, 0, 0, 0.08);
  --shadow-hover: 0 12px 40px rgba(0, 0, 0, 0.15);

  /* ==================== 过渡动画 ==================== */

  --transition-fast: 150ms cubic-bezier(0.4, 0, 0.2, 1);
  --transition-base: 300ms cubic-bezier(0.4, 0, 0.2, 1);
  --transition-slow: 500ms cubic-bezier(0.4, 0, 0.2, 1);

  /* ==================== 间距 ==================== */

  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-5: 20px;
  --space-6: 24px;
  --space-8: 32px;

  /* ==================== 字体 ==================== */

  --font-sans: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  --font-mono: 'Courier New', monospace;
}
```

**步骤 2：创建 styles/base.css**

```css
/**
 * 基础样式重置
 */

/* 全局重置 */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

/* HTML 和 Body */
html, body {
  width: 100%;
  height: 100%;
  overflow-x: hidden;
  font-family: var(--font-sans);
}

/* 根容器 */
#root {
  width: 100%;
  height: 100%;
}

/* 滚动条美化 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.5);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(107, 114, 128, 0.7);
}

::-webkit-scrollbar-corner {
  background: transparent;
}
```

**步骤 3：创建 styles/index.css**

```css
/**
 * 样式入口文件
 * 按顺序导入各模块样式
 */

/* 设计 Tokens（必须最先导入） */
@import './tokens.css';

/* 基础样式 */
@import './base.css';

/* 布局样式 */
@import './layout.css';

/* 公共组件样式 */
@import './components.css';

/* 页面样式 */
@import './skill.css';
@import './jobtask.css';

/* 弹窗样式 */
@import './modal.css';

/* 响应式样式（必须最后导入） */
@import './responsive.css';
```

**步骤 4：更新 main.tsx 导入**

```typescript
// 修改前
import './App.css';

// 修改后
import './styles/index.css';
```

#### 7.4 验证步骤

```powershell
# 执行构建检查
yarn build

# 检查样式是否正常加载
```

---

### 任务 8：样式命名规范化

#### 8.1 命名规范

采用 BEM 命名规范：

```
.block {}
.block__element {}
.block--modifier {}
```

#### 8.2 示例转换

```css
/* 修改前 */
.skill-card {}
.skill-card-header {}
.skill-card.selected {}

/* 修改后 */
.skill-card {}
.skill-card__header {}
.skill-card--selected {}
```

---

## 六、重构检查清单

### 阶段 P0 检查清单

- [ ] 创建 `types/common.ts` 并定义通用类型
- [ ] 创建 `types/tag.ts` 并定义标签类型
- [ ] 创建 `types/skill.ts` 并定义技能类型
- [ ] 创建 `types/jobtask.ts` 并定义任务类型
- [ ] 创建 `types/index.ts` 统一导出
- [ ] 更新 `services/index.ts` 移除类型定义
- [ ] 更新所有文件的类型导入路径
- [ ] 更新 `types/constant.ts` 添加所有常量
- [ ] 创建 `utils/format.ts` 提取格式化函数
- [ ] 创建 `utils/config.ts` 提取配置函数
- [ ] 创建 `utils/clipboard.ts` 提取剪贴板函数
- [ ] 创建 `utils/index.ts` 统一导出
- [ ] 执行 TypeScript 编译检查通过

### 阶段 P1 检查清单

- [ ] 创建 `components/common/StatCard.tsx`
- [ ] 创建 `components/common/BackToTop.tsx`
- [ ] 创建 `components/common/index.ts` 统一导出
- [ ] 创建 `hooks/useWindowWidth.ts`
- [ ] 创建 `hooks/useFilterSticky.ts`
- [ ] 创建 `hooks/useDebounce.ts`
- [ ] 创建 `hooks/index.ts` 统一导出
- [ ] 创建 `pages/SkillManagement/SkillCard.tsx`
- [ ] 创建 `pages/SkillManagement/useSkillData.ts`
- [ ] 重构 `pages/SkillManagement/index.tsx`
- [ ] 执行 TypeScript 编译检查通过

### 阶段 P2 检查清单

- [ ] 创建 `styles/tokens.css`
- [ ] 创建 `styles/base.css`
- [ ] 创建 `styles/layout.css`
- [ ] 创建 `styles/components.css`
- [ ] 创建 `styles/skill.css`
- [ ] 创建 `styles/jobtask.css`
- [ ] 创建 `styles/modal.css`
- [ ] 创建 `styles/responsive.css`
- [ ] 创建 `styles/index.css` 统一导入
- [ ] 更新 `main.tsx` 样式导入
- [ ] 执行构建检查通过

---

## 七、预期收益表

| 收益项 | 具体效果 | 影响范围 |
|--------|----------|----------|
| **代码复用** | StatCard、formatTime 等复用，减少 ~150 行重复代码 | 全局 |
| **类型安全** | 类型定义独立，IDE 智能提示更准确 | 开发体验 |
| **可维护性** | 常量统一管理，修改配置只需一处 | 全局 |
| **可测试性** | 工具函数独立，便于单元测试 | utils/ |
| **可扩展性** | Hooks 提取后，新页面可直接复用 | hooks/ |
| **样式隔离** | CSS 按模块拆分，避免样式冲突 | styles/ |
| **构建优化** | 样式模块化后可按需加载 | 生产构建 |
| **团队协作** | 清晰的目录结构，降低沟通成本 | 团队 |

---

## 八、风险与应对

| 风险 | 可能性 | 影响 | 应对措施 |
|------|--------|------|----------|
| 导入路径错误 | 中 | 编译失败 | 逐文件检查，使用 IDE 重构功能 |
| 类型不兼容 | 低 | 类型错误 | 保持接口不变，仅调整内部实现 |
| 样式丢失 | 低 | UI 异常 | 按模块拆分后逐一验证 |
| 功能回归 | 低 | 功能异常 | 每阶段完成后进行功能测试 |

---

## 九、附录

### A. 目录结构总览

```
frontend/src/
├── components/
│   ├── common/
│   │   ├── StatCard.tsx
│   │   ├── BackToTop.tsx
│   │   └── index.ts
│   ├── SkillModal.tsx
│   ├── JobTaskModal.tsx
│   └── ...
├── hooks/
│   ├── useWindowWidth.ts
│   ├── useFilterSticky.ts
│   ├── useDebounce.ts
│   └── index.ts
├── pages/
│   ├── SkillManagement/
│   │   ├── index.tsx
│   │   ├── SkillCard.tsx
│   │   └── useSkillData.ts
│   └── JobTaskManagement/
│       └── ...
├── services/
│   └── index.ts
├── stores/
│   ├── appStore.ts
│   └── modalStore.ts
├── styles/
│   ├── index.css
│   ├── tokens.css
│   ├── base.css
│   ├── layout.css
│   ├── components.css
│   ├── skill.css
│   ├── jobtask.css
│   ├── modal.css
│   └── responsive.css
├── types/
│   ├── index.ts
│   ├── common.ts
│   ├── skill.ts
│   ├── jobtask.ts
│   ├── tag.ts
│   └── constant.ts
├── utils/
│   ├── index.ts
│   ├── format.ts
│   ├── config.ts
│   └── clipboard.ts
├── App.tsx
├── main.tsx
└── index.css
```

### B. 命令速查

```powershell
# TypeScript 编译检查
npx tsc --noEmit

# 开发服务器
yarn dev

# 生产构建
yarn build

# 代码检查
yarn lint
```

---

> 文档维护：前端架构师
> 最后更新：2026-02-12
