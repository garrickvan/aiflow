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
