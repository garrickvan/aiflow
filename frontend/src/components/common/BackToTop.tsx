/**
 * 回到顶部组件
 * 提供页面滚动回顶部的功能按钮
 */

import React from 'react';
import { VerticalAlignTopOutlined } from '@ant-design/icons';
import { BACK_TO_TOP_THRESHOLD } from '../../types';

/**
 * 回到顶部组件属性
 */
export interface BackToTopProps {
  /** 是否显示按钮 */
  visible: boolean;
  /** 点击回调 */
  onClick: () => void;
  /** 自定义提示文本 */
  title?: string;
}

/**
 * 回到顶部组件
 * @example
 * <BackToTop visible={showBackToTop} onClick={handleBackToTop} />
 */
export const BackToTop: React.FC<BackToTopProps> = ({
  visible,
  onClick,
  title = '回到顶部',
}) => {
  if (!visible) return null;

  return (
    <button
      className="back-to-top-btn"
      onClick={onClick}
      title={title}
      aria-label={title}
    >
      <VerticalAlignTopOutlined />
    </button>
  );
};

/**
 * 使用回到顶部功能的 Hook
 * @param threshold - 显示阈值（像素）
 * @returns [visible, scrollToTop]
 */
export const useBackToTop = (
  threshold: number = BACK_TO_TOP_THRESHOLD,
): [boolean, () => void] => {
  const [visible, setVisible] = React.useState(false);

  React.useEffect(() => {
    const handleScroll = () => {
      setVisible(document.body.scrollTop > threshold);
    };

    document.body.addEventListener('scroll', handleScroll);
    handleScroll();

    return () => {
      document.body.removeEventListener('scroll', handleScroll);
    };
  }, [threshold]);

  const scrollToTop = React.useCallback(() => {
    document.body.scrollTo({
      top: 0,
      behavior: 'smooth',
    });
  }, []);

  return [visible, scrollToTop];
};
