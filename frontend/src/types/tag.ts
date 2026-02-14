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
