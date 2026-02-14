/**
 * 标签编辑模态框组件
 * 用于新增和编辑标签信息
 * 使用Zustand状态管理
 */
import React from 'react'
import {
  Modal,
  Form,
  Input,
  Button,
  Space,
  message,
  Tag
} from 'antd'
import { 
  TagOutlined, 
  EditOutlined,
  PlusOutlined,
  SaveOutlined,
  CloseCircleOutlined
} from '@ant-design/icons'
import { tagApi } from '../services'
import { useAppStore } from '../stores/appStore'
import { useModalStore, ModalType, type ModalDataMap } from '../stores/modalStore'

/**
 * 标签编辑模态框组件
 */
const GroupModal: React.FC = () => {
  // 创建表单实例
  const [form] = Form.useForm()

  // 从Zustand Store获取状态和actions
  const { tags, setTags } = useAppStore()
  const { currentModal, modalData, closeGroupModal } = useModalStore()

  // 计算当前模态框状态和编辑数据
  const isGroupModalOpen = currentModal === ModalType.GROUP
  const editingTag = (modalData as ModalDataMap[typeof ModalType.GROUP]) ?? null

  // 当编辑标签变化且模态框打开时，更新表单值
  React.useEffect(() => {
    if (isGroupModalOpen) {
      // 使用setTimeout确保Form组件已挂载
      setTimeout(() => {
        if (editingTag) {
          form.setFieldsValue(editingTag)
        } else {
          form.resetFields()
        }
      }, 0)
    }
  }, [editingTag, form, isGroupModalOpen])

  /**
   * 处理表单提交
   */
  const handleSubmit = async (values: { name: string }) => {
    try {
      if (editingTag) {
        // 更新标签
        const updatedTag = await tagApi.updateTag(editingTag.id, values.name)
        const updatedTags = tags.map((tag) =>
          tag.id === updatedTag.id ? updatedTag : tag,
        )
        setTags(updatedTags)
        message.success("标签更新成功")
      } else {
        // 新增标签
        const newTag = await tagApi.createTag(values.name)
        setTags([...tags, newTag])
        message.success("标签创建成功")
      }
      closeGroupModal()
    } catch (error) {
      message.error("操作失败，请重试")
      console.error("Tag submit failed:", error)
    }
  }

  /**
   * 处理取消操作
   */
  const handleCancel = () => {
    form.resetFields()
    closeGroupModal()
  }

  return (
    <Modal
      title={
        <div className="tag-modal-header">
          <div className={`tag-modal-icon ${editingTag ? 'edit' : 'add'}`}>
            {editingTag ? <EditOutlined /> : <PlusOutlined />}
          </div>
          <div className="tag-modal-title">
            <span className="title-text">{editingTag ? '编辑标签' : '新增标签'}</span>
            <span className="title-subtext">
              {editingTag ? `修改「${editingTag.name}」的标签名称` : '创建一个新标签用于分类技能'}
            </span>
          </div>
        </div>
      }
      open={isGroupModalOpen}
      onCancel={handleCancel}
      footer={null}
      width={480}
      className="tag-edit-modal"
      destroyOnClose
    >
      <div className="tag-modal-body">
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          className="tag-form"
        >
          <Form.Item
            name="name"
            label={
              <span className="tag-form-label">
                <TagOutlined /> 标签名称
              </span>
            }
            rules={[
              { required: true, message: '请输入标签名称' },
              { min: 1, max: 100, message: '标签名称长度应在 1-100 字符之间' }
            ]}
          >
            <Input 
              placeholder="请输入标签名称，如：文档处理" 
              autoComplete="off"
              size="large"
              className="tag-input"
              prefix={<Tag className="input-prefix-tag">Tag</Tag>}
              allowClear
            />
          </Form.Item>

          <Form.Item className="tag-form-footer">
            <Space size="middle">
              <Button 
                onClick={handleCancel}
                size="middle"
                className="tag-btn-cancel"
                icon={<CloseCircleOutlined />}
              >
                取消
              </Button>
              <Button 
                type="primary" 
                htmlType="submit"
                size="middle"
                className={`tag-btn-submit ${editingTag ? 'edit' : 'add'}`}
                icon={<SaveOutlined />}
              >
                {editingTag ? '保存修改' : '创建标签'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </div>
    </Modal>
  )
}

export default GroupModal
