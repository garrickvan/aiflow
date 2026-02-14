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
