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
