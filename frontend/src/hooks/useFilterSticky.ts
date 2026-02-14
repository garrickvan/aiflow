/**
 * 筛选栏悬浮 Hook
 * 使用 IntersectionObserver 实现筛选栏吸顶效果
 */

import { useState, useEffect, useRef } from 'react';

/**
 * 筛选栏悬浮状态
 */
export interface FilterStickyState {
  /** 是否处于悬浮状态 */
  isSticky: boolean;
  /** 哨兵元素 ref */
  sentinelRef: React.RefObject<HTMLDivElement | null>;
}

/**
 * 监听筛选栏悬浮状态
 * @returns 悬浮状态和哨兵元素 ref
 * @example
 * const { isSticky, sentinelRef } = useFilterSticky();
 * // JSX:
 * // <div ref={sentinelRef} className="filter-sentinel" />
 * // <div className={`filter-bar ${isSticky ? 'sticky' : ''}`}>
 */
export const useFilterSticky = (): FilterStickyState => {
  const [isSticky, setIsSticky] = useState(false);
  const sentinelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        setIsSticky(!entry.isIntersecting);
      },
      {
        root: null,
        threshold: 0,
        rootMargin: '0px 0px 0px 0px',
      },
    );

    observer.observe(sentinel);

    return () => {
      observer.disconnect();
    };
  }, []);

  return {
    isSticky,
    sentinelRef,
  };
};
