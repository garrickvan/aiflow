/**
 * 全局应用状态管理
 * 使用Zustand替代React Context，避免不必要的重渲染
 * 模态框状态已迁移至modalStore.ts
 */
import { create } from "zustand";
import type { Dayjs } from "dayjs";
import type { Skill, Tag, JobTask } from "../types";

/**
 * 分页状态接口
 */
interface PaginationState {
  page: number;
  pageSize: number;
  total: number;
}

/**
 * 应用状态接口
 */
interface AppState {
  // 数据状态
  skills: Skill[];
  tags: Tag[];
  jobTasks: JobTask[];
  projects: string[];

  // 加载状态
  loading: boolean;

  // UI状态 - 侧边栏
  collapsed: boolean;
  mobileMenuOpen: boolean;

  // 筛选状态 - 技能
  selectedTagId: number | undefined;
  selectedSkillDateRange: [Dayjs | null, Dayjs | null] | null;

  // 筛选状态 - 任务
  selectedProject: string | undefined;
  selectedJobTaskType: string | undefined;
  selectedJobTaskStatus: string | undefined;
  selectedJobTaskDateRange: [Dayjs | null, Dayjs | null] | null;

  // 自动刷新
  autoRefresh: boolean;

  // 分页状态
  skillPagination: PaginationState;
  tagPagination: PaginationState;
  jobTaskPagination: PaginationState;

  // Actions - 数据设置（支持函数式更新以避免闭包问题）
  setSkills: (skills: Skill[] | ((prev: Skill[]) => Skill[])) => void;
  setTags: (tags: Tag[] | ((prev: Tag[]) => Tag[])) => void;
  setJobTasks: (jobTasks: JobTask[] | ((prev: JobTask[]) => JobTask[])) => void;
  setProjects: (projects: string[] | ((prev: string[]) => string[])) => void;

  // Actions - 加载状态
  setLoading: (loading: boolean) => void;

  // Actions - 侧边栏
  toggleCollapsed: () => void;
  setMobileMenuOpen: (open: boolean) => void;

  // Actions - 筛选
  setSelectedTagId: (id: number | undefined) => void;
  setSelectedSkillDateRange: (range: [Dayjs | null, Dayjs | null] | null) => void;
  setSelectedProject: (project: string | undefined) => void;
  setSelectedJobTaskType: (type: string | undefined) => void;
  setSelectedJobTaskStatus: (status: string | undefined) => void;
  setSelectedJobTaskDateRange: (range: [Dayjs | null, Dayjs | null] | null) => void;

  // Actions - 自动刷新
  setAutoRefresh: (autoRefresh: boolean) => void;

  // Actions - 分页
  setSkillPagination: (pagination: Partial<PaginationState>) => void;
  setTagPagination: (pagination: Partial<PaginationState>) => void;
  setJobTaskPagination: (pagination: Partial<PaginationState>) => void;

  // Actions - 重置
  resetFilters: () => void;
}

/**
 * 创建应用状态Store
 * 模态框状态已分离至modalStore.ts
 */
export const useAppStore = create<AppState>((set) => ({
  // 初始数据状态
  skills: [],
  tags: [],
  jobTasks: [],
  projects: [],

  // 初始加载状态
  loading: true,

  // 初始侧边栏状态
  collapsed: false,
  mobileMenuOpen: false,

  // 初始筛选状态
  selectedTagId: undefined,
  selectedSkillDateRange: null,
  selectedProject: undefined,
  selectedJobTaskType: undefined,
  selectedJobTaskStatus: undefined,
  selectedJobTaskDateRange: null,

  // 初始自动刷新
  autoRefresh: true,

  // 初始分页状态
  skillPagination: { page: 1, pageSize: 20, total: 0 },
  tagPagination: { page: 1, pageSize: 20, total: 0 },
  jobTaskPagination: { page: 1, pageSize: 20, total: 0 },

  // Actions - 数据设置（支持函数式更新以避免闭包问题）
  setSkills: (skills) =>
    set((state) => ({
      skills: typeof skills === "function" ? skills(state.skills) : skills,
    })),
  setTags: (tags) =>
    set((state) => ({
      tags: typeof tags === "function" ? tags(state.tags) : tags,
    })),
  setJobTasks: (jobTasks) =>
    set((state) => ({
      jobTasks: typeof jobTasks === "function" ? jobTasks(state.jobTasks) : jobTasks,
    })),
  setProjects: (projects) =>
    set((state) => ({
      projects: typeof projects === "function" ? projects(state.projects) : projects,
    })),

  // Actions - 加载状态
  setLoading: (loading) => set({ loading }),

  // Actions - 侧边栏
  toggleCollapsed: () => set((state) => ({ collapsed: !state.collapsed })),
  setMobileMenuOpen: (mobileMenuOpen) => set({ mobileMenuOpen }),

  // Actions - 筛选
  setSelectedTagId: (selectedTagId) => set({ selectedTagId }),
  setSelectedSkillDateRange: (selectedSkillDateRange) =>
    set({ selectedSkillDateRange }),
  setSelectedProject: (selectedProject) => set({ selectedProject }),
  setSelectedJobTaskType: (selectedJobTaskType) =>
    set({ selectedJobTaskType }),
  setSelectedJobTaskStatus: (selectedJobTaskStatus) =>
    set({ selectedJobTaskStatus }),
  setSelectedJobTaskDateRange: (selectedJobTaskDateRange) =>
    set({ selectedJobTaskDateRange }),

  // Actions - 自动刷新
  setAutoRefresh: (autoRefresh) => set({ autoRefresh }),

  // Actions - 分页
  setSkillPagination: (pagination) =>
    set((state) => ({
      skillPagination: { ...state.skillPagination, ...pagination },
    })),
  setTagPagination: (pagination) =>
    set((state) => ({
      tagPagination: { ...state.tagPagination, ...pagination },
    })),
  setJobTaskPagination: (pagination) =>
    set((state) => ({
      jobTaskPagination: { ...state.jobTaskPagination, ...pagination },
    })),

  // Actions - 重置筛选
  resetFilters: () =>
    set({
      selectedTagId: undefined,
      selectedSkillDateRange: null,
      selectedProject: undefined,
      selectedJobTaskType: undefined,
      selectedJobTaskStatus: undefined,
      selectedJobTaskDateRange: null,
    }),
}));
