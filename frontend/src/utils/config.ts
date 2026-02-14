/**
 * 配置获取相关工具函数
 */

import { SKILL_AVATAR_GRADIENTS } from '../types';

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
  const firstChar = (name || '').charAt(0).toUpperCase();
  const index = (name || '').length % SKILL_AVATAR_GRADIENTS.length;
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
  /** 背景颜色 */
  bgColor: string;
  /** 边框颜色 */
  borderColor: string;
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
      bgColor: '#f8fafc',
      borderColor: '#e2e8f0',
    },
    处理中: {
      color: '#3b82f6',
      bgColor: '#eff6ff',
      borderColor: '#bfdbfe',
    },
    处理失败: {
      color: '#ef4444',
      bgColor: '#fef2f2',
      borderColor: '#fecaca',
    },
    处理完成: {
      color: '#10b981',
      bgColor: '#f0fdf4',
      borderColor: '#bbf7d0',
    },
    验收通过: {
      color: '#8b5cf6',
      bgColor: '#faf5ff',
      borderColor: '#e9d5ff',
    },
  };

  return configMap[status] || {
    color: '#6b7280',
    bgColor: '#f8fafc',
    borderColor: '#e2e8f0',
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
