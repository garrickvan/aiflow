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
