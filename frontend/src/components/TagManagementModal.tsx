/**
 * 标签管理模态框组件
 * 用于在技能管理页面中管理标签列表
 * 使用Zustand状态管理
 */
import React from 'react'
import {
  Modal,
  Table,
  Button,
  Space,
  Popconfirm,
  message,
  Badge,
  Tag as AntTag
} from 'antd'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  TagOutlined,
  ClockCircleOutlined,
  AppstoreOutlined
} from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import type { Tag } from '../services'
import { tagApi } from '../services'
import { useAppStore } from '../stores/appStore'
import { useModalStore, ModalType } from '../stores/modalStore'

/**
 * 标签管理模态框组件
 */
const TagManagementModal: React.FC = () => {
  // 从Zustand Store获取状态和actions
  const { tags, skills, tagPagination, setTags, setTagPagination } = useAppStore()
  const { currentModal, closeTagManagementModal, openGroupModal } = useModalStore()

  // 计算当前模态框状态
  const isTagManagementModalOpen = currentModal === ModalType.TAG_MANAGEMENT

  // 处理删除标签
  const handleDeleteTag = async (id: number) => {
    try {
      await tagApi.deleteTag(id);
      const updatedTags = tags.filter((tag) => tag.id !== id);
      setTags(updatedTags);
      message.success("标签删除成功");
    } catch (error) {
      message.error("删除失败，请重试");
      console.error("Tag delete failed:", error);
    }
  };

  // 处理分页变化
  const handlePageChange = (page: number, pageSize: number) => {
    setTagPagination({ page, pageSize });
  };

  // 计算标签使用数量
  const getTagSkillCount = (tagId: number): number => {
    return skills.filter(skill => 
      skill.tags && skill.tags.some(tag => tag.id === tagId)
    ).length
  }

  // 标签表格列定义
  const tagColumns: ColumnsType<Tag> = [
    {
      title: '标签名称',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      render: (name: string) => (
        <div className="tag-name-cell">
          <AntTag 
            className="tag-badge-primary"
            icon={<TagOutlined />}
          >
            {name}
          </AntTag>
        </div>
      )
    },
    {
      title: '技能数量',
      key: 'skillCount',
      width: 120,
      align: 'center',
      render: (_: unknown, record: Tag) => {
        const count = getTagSkillCount(record.id)
        return (
          <Badge 
            count={count} 
            className="tag-skill-count"
            style={{ 
              backgroundColor: count > 0 ? '#8b5cf6' : '#d1d5db',
              fontSize: '12px',
              fontWeight: 600
            }}
          />
        )
      }
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
      render: (createdAt) => {
        const date = new Date(createdAt)
        const isDateValid = !isNaN(date.getTime())
        return (
          <div className="tag-time-cell">
            <ClockCircleOutlined className="time-icon" />
            <span className="time-text">
              {isDateValid ? date.toLocaleString('zh-CN', {
                year: 'numeric',
                month: '2-digit',
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit'
              }) : '无'}
            </span>
          </div>
        )
      }
    },
    {
      title: '操作',
      key: 'action',
      width: 160,
      fixed: 'right',
      align: 'center',
      render: (_: unknown, record: Tag) => (
        <Space size="small" className="tag-actions">
          <Button
            type="primary"
            ghost
            icon={<EditOutlined />}
            size="small"
            className="tag-edit-btn"
            onClick={() => openGroupModal(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个标签吗？"
            description="删除后，相关技能将不再关联此标签"
            onConfirm={() => handleDeleteTag(record.id)}
            okText="确定"
            cancelText="取消"
            placement="topRight"
            okButtonProps={{ danger: true }}
          >
            <Button
              type="primary"
              danger
              icon={<DeleteOutlined />}
              size="small"
              className="tag-delete-btn"
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      )
    }
  ]

  // 计算统计数据
  const totalTags = tags.length
  const totalUsedTags = tags.filter(tag => getTagSkillCount(tag.id) > 0).length

  return (
    <Modal
      title={
        <div className="tag-modal-header">
          <div className="tag-modal-icon">
            <AppstoreOutlined />
          </div>
          <div className="tag-modal-title">
            <span className="title-text">标签管理</span>
            <span className="title-subtext">共 {totalTags} 个标签，{totalUsedTags} 个正在使用</span>
          </div>
        </div>
      }
      open={isTagManagementModalOpen}
      onCancel={closeTagManagementModal}
      footer={null}
      width={720}
      className="tag-management-modal"
      destroyOnClose
    >
      <div className="tag-modal-content">
        {/* 新增标签按钮 */}
        <div className="tag-modal-toolbar">
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => openGroupModal()}
            className="tag-add-btn"
            size="middle"
          >
            新增标签
          </Button>
        </div>

        {/* 标签表格 */}
        <Table
          columns={tagColumns}
          dataSource={tags}
          rowKey="id"
          pagination={{
            total: tagPagination.total,
            current: tagPagination.page,
            pageSize: tagPagination.pageSize,
            showSizeChanger: true,
            showQuickJumper: true,
            onChange: handlePageChange,
            pageSizeOptions: ['10', '20', '50'],
            showTotal: (total) => `共 ${total} 条记录`,
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
          scroll={{ x: 600 }}
          size="middle"
          className="tag-table"
          rowClassName={() => 'tag-table-row'}
          locale={{
            emptyText: (
              <div className="tag-empty-state">
                <TagOutlined className="empty-icon" />
                <p>暂无标签数据</p>
                <span>点击上方按钮创建第一个标签</span>
              </div>
            )
          }}
        />
      </div>
    </Modal>
  )
}

export default TagManagementModal
