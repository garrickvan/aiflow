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
