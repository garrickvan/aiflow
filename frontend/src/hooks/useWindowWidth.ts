/**
 * 窗口宽度监听 Hook
 * 提供响应式布局所需的窗口宽度信息
 */

import { useState, useEffect } from 'react';
import { MOBILE_BREAKPOINT } from '../types';

/**
 * 窗口宽度信息
 */
export interface WindowWidthInfo {
  /** 当前窗口宽度（像素） */
  width: number;
  /** 是否为移动端 */
  isMobile: boolean;
}

/**
 * 监听窗口宽度变化
 * @returns 窗口宽度信息
 * @example
 * const { width, isMobile } = useWindowWidth();
 */
export const useWindowWidth = (): WindowWidthInfo => {
  const [width, setWidth] = useState(window.innerWidth);

  useEffect(() => {
    const handleResize = () => {
      setWidth(window.innerWidth);
    };

    window.addEventListener('resize', handleResize);
    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, []);

  return {
    width,
    isMobile: width < MOBILE_BREAKPOINT,
  };
};
