/**
 * 统计卡片组件
 * 用于展示统计数据，支持标题、数值、图标和副标题
 */

import React from 'react';
import { Card } from 'antd';

/**
 * 统计卡片属性
 */
export interface StatCardProps {
  /** 标题 */
  title: string;
  /** 数值（支持数字或字符串） */
  value: number | string;
  /** 图标 */
  icon: React.ReactNode;
  /** 主题颜色 */
  color: string;
  /** 副标题（可选） */
  subtitle?: string;
}

/**
 * 统计卡片组件
 * @example
 * <StatCard
 *   title="技能总数"
 *   value={42}
 *   icon={<FileTextOutlined />}
 *   color="#8b5cf6"
 * />
 */
export const StatCard: React.FC<StatCardProps> = ({
  title,
  value,
  icon,
  color,
  subtitle,
}) => (
  <Card className="stat-card" variant="borderless">
    <div className="stat-content">
      <div>
        <div className="stat-title">{title}</div>
        <div className="stat-value" style={{ color }}>
          {value}
        </div>
        {subtitle && <div className="stat-subtitle">{subtitle}</div>}
      </div>
      <div
        className="stat-icon"
        style={{
          backgroundColor: `${color}1A`,
          color,
        }}
      >
        {icon}
      </div>
    </div>
  </Card>
);
