/**
 * 标签管理页面组件
 * 用于展示和管理标签列表
 */
import React from 'react'
import {
  Table,
  Button,
  Typography,
  Breadcrumb,
  Space,
  Popconfirm,
  Divider
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import type { Skill, Tag } from '../services'

interface GroupManagementProps {
  tags: Tag[];
  skills: Skill[];
  onAddGroup: () => void;
  onEditGroup: (tag: Tag) => void;
  onDeleteGroup: (id: number) => void;
  pagination: {
    total: number;
    current: number;
    pageSize: number;
  };
  onPageChange: (page: number, pageSize: number) => void;
}

/**
 * 标签管理页面组件
 */
const GroupManagement: React.FC<GroupManagementProps> = ({
  tags,
  skills,
  onAddGroup,
  onEditGroup,
  onDeleteGroup,
  pagination,
  onPageChange
}) => {
  // 对 skills 做空值保护，确保始终是数组
  const safeSkills = skills || [];

  // 标签表格列定义
  const groupColumns: ColumnsType<Tag> = [
    {
      title: '标签名称',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true
    },
    {
      title: '技能数量',
      key: 'skillCount',
      width: 90,
      ellipsis: true,
      render: (_: unknown, record: Tag) => {
        // 计算使用此标签的技能数量
        let count = 0;
        safeSkills.forEach(skill => {
          if (skill.tags && skill.tags.some(tag => tag.id === record.id)) {
            count++;
          }
        });
        return <span>{count}</span>;
      }
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      ellipsis: true,
      render: (createdAt) => {
        // 验证时间是否有效
        const date = new Date(createdAt);
        const isDateValid = !isNaN(date.getTime());

        return (
          <span>{isDateValid ? date.toLocaleString() : '无'}</span>
        );
      }
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      fixed: 'right',
      render: (_: unknown, record: Tag) => (
        <Space size="small">
          <Button
            type="primary"
            ghost
            icon={<EditOutlined />}
            size="small"
            onClick={() => onEditGroup(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个标签吗？"
            onConfirm={() => onDeleteGroup(record.id)}
            okText="确定"
            cancelText="取消"
            placement="topRight"
          >
            <Button
              type="primary"
              danger
              icon={<DeleteOutlined />}
              size="small"
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      )
    }
  ]

  return (
    <>
      {/* 面包屑 */}
      <Breadcrumb
        className="app-breadcrumb"
        items={[
          { title: '首页' },
          { title: '标签管理' }
        ]}
      />

      {/* 页面标题和按钮 */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
        <Typography.Title level={4} style={{ margin: 0, color: '#1E293B' }}>标签管理</Typography.Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={onAddGroup}
        >
          新增标签
        </Button>
      </div>

      <Divider style={{ marginBottom: '24px' }} />

      {/* 标签表格 */}
      <Table
        columns={groupColumns}
        dataSource={tags}
        rowKey="id"
        pagination={{
          total: pagination.total,
          current: pagination.current,
          pageSize: pagination.pageSize,
          showSizeChanger: true,
          showQuickJumper: true,
          onChange: onPageChange,
          pageSizeOptions: ['20', '50', '100'],
          locale: {
            items_per_page: '条/页',
            jump_to: '跳至',
            jump_to_confirm: '确定',
            page: '页',
            prev_page: '上一页',
            next_page: '下一页',
            prev_5: '向前 5 页',
            next_5: '向后 5 页',
            prev_3: '向前 3 页',
            next_3: '向后 3 页',
          }
        }}
        scroll={{ x: 800 }}
      />
    </>
  )
}

export default GroupManagement
