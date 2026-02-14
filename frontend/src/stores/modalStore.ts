/**
 * 模态框状态管理
 * 统一管理中心，避免appStore膨胀
 * 使用Zustand替代React Context，避免不必要的重渲染
 */
import { create } from "zustand";
import type { Skill, Tag, JobTask } from "../types";

/**
 * 模态框类型常量
 * 使用const对象替代enum以兼容erasableSyntaxOnly配置
 */
export const ModalType = {
  NONE: "none",
  SKILL: "skill",
  GROUP: "group",
  UPLOAD: "upload",
  JOB_TASK: "job_task",
  TAG_MANAGEMENT: "tag_management",
} as const;

/**
 * 模态框类型类型定义
 */
export type ModalTypeValue = typeof ModalType[keyof typeof ModalType];

/**
 * 模态框数据类型映射
 */
export interface ModalDataMap {
  [ModalType.NONE]: null;
  [ModalType.SKILL]: Skill | null;
  [ModalType.GROUP]: Tag | null;
  [ModalType.UPLOAD]: null;
  [ModalType.JOB_TASK]: JobTask | null;
  [ModalType.TAG_MANAGEMENT]: null;
}

/**
 * 模态框状态接口
 */
interface ModalState {
  // 当前打开的模态框类型
  currentModal: ModalTypeValue;
  // 当前模态框的数据
  modalData: ModalDataMap[ModalTypeValue];

  // Actions - 通用模态框操作
  openModal: <T extends ModalTypeValue>(
    type: T,
    data?: ModalDataMap[T]
  ) => void;
  closeModal: () => void;

  // Actions - 便捷方法（保持向后兼容）
  openSkillModal: (skill?: Skill) => void;
  closeSkillModal: () => void;
  openGroupModal: (tag?: Tag) => void;
  closeGroupModal: () => void;
  openUploadModal: () => void;
  closeUploadModal: () => void;
  openJobTaskModal: (jobTask?: JobTask) => void;
  closeJobTaskModal: () => void;
  openTagManagementModal: () => void;
  closeTagManagementModal: () => void;
}

/**
 * 创建模态框状态Store
 * 采用统一的状态管理，避免多个独立状态字段
 */
export const useModalStore = create<ModalState>((set) => ({
  // 初始状态
  currentModal: ModalType.NONE,
  modalData: null,

  // 通用打开模态框方法
  openModal: (type, data) =>
    set({
      currentModal: type,
      modalData: data ?? null,
    }),

  // 通用关闭模态框方法
  closeModal: () =>
    set({
      currentModal: ModalType.NONE,
      modalData: null,
    }),

  // 技能模态框便捷方法
  openSkillModal: (skill) =>
    set({
      currentModal: ModalType.SKILL,
      modalData: skill || null,
    }),
  closeSkillModal: () =>
    set({
      currentModal: ModalType.NONE,
      modalData: null,
    }),

  // 标签编辑模态框便捷方法
  openGroupModal: (tag) =>
    set({
      currentModal: ModalType.GROUP,
      modalData: tag || null,
    }),
  closeGroupModal: () =>
    set({
      currentModal: ModalType.NONE,
      modalData: null,
    }),

  // 上传模态框便捷方法
  openUploadModal: () =>
    set({
      currentModal: ModalType.UPLOAD,
      modalData: null,
    }),
  closeUploadModal: () =>
    set({
      currentModal: ModalType.NONE,
      modalData: null,
    }),

  // 任务模态框便捷方法
  openJobTaskModal: (jobTask) =>
    set({
      currentModal: ModalType.JOB_TASK,
      modalData: jobTask || null,
    }),
  closeJobTaskModal: () =>
    set({
      currentModal: ModalType.NONE,
      modalData: null,
    }),

  // 标签管理模态框便捷方法
  openTagManagementModal: () =>
    set({
      currentModal: ModalType.TAG_MANAGEMENT,
      modalData: null,
    }),
  closeTagManagementModal: () =>
    set({
      currentModal: ModalType.NONE,
      modalData: null,
    }),
}));

/**
 * 模态框状态选择器Hook
 * 用于获取特定模态框的状态
 */
export const useModal = (type: ModalTypeValue) => {
  const { currentModal, modalData, closeModal } = useModalStore();
  return {
    isOpen: currentModal === type,
    data: modalData as ModalDataMap[typeof type],
    close: closeModal,
  };
};
