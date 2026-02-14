/**
 * 格式化相关工具函数
 */

/**
 * 格式化时间戳为本地时间字符串
 * @param timestamp - 时间戳（毫秒）
 * @returns 格式化后的时间字符串
 * @example
 * formatTime(1707753600000) // "2024-02-12 18:00:00"
 */
export const formatTime = (timestamp: number): string => {
  if (!timestamp) return '-';
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
  if (!timestamp) return '-';
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
