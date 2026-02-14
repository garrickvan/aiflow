/**
 * 文件上传Modal组件
 * 支持拖拽和点击上传md、zip格式的文件
 * 使用Zustand状态管理
 */
import React from 'react'
import {
  Modal,
  Upload,
  Button,
  message,
  Typography
} from 'antd'
import {
  InboxOutlined,
  UploadOutlined,
  CloseOutlined,
  FileMarkdownOutlined,
  FileZipOutlined
} from '@ant-design/icons'
import type { UploadProps } from 'antd'
import { API_BASE_URL } from '../types/constant'
import { useModalStore, ModalType } from '../stores/modalStore'

const { Dragger } = Upload
const { Text } = Typography

/**
 * 文件上传Modal组件
 */
const UploadModal: React.FC = () => {
  // 从Zustand Store获取状态和actions
  const { currentModal, closeUploadModal } = useModalStore()

  // 计算当前模态框状态
  const isUploadModalOpen = currentModal === ModalType.UPLOAD

  const processType = 'import_skill';

  // 上传前的文件验证
  const beforeUpload: UploadProps['beforeUpload'] = (file) => {
    // 验证文件类型
    const isMd = file.name.endsWith('.md')
    const isZip = file.name.endsWith('.zip')
    if (!isMd && !isZip) {
      message.error('只能上传md或zip格式的文件！')
      return Upload.LIST_IGNORE
    }

    // 验证文件大小（50MB限制）
    const isLt50M = file.size / 1024 / 1024 < 50
    if (!isLt50M) {
      message.error('文件大小不能超过50MB！')
      return Upload.LIST_IGNORE
    }

    return true
  }

  // 上传成功处理
  const handleUploadSuccess = () => {
    message.success('文件上传成功！')
    closeUploadModal()
  }

  // 上传失败处理
  const handleUploadError = () => {
    message.error('文件上传失败，请重试！')
  }

  // 上传配置
  const uploadProps: UploadProps = {
    name: 'file',
    multiple: false,
    action: `${API_BASE_URL}/upload_data`,
    method: 'POST',
    headers: {
      'Content-Type': 'multipart/form-data',
    },
    data: {
      process_type: processType,
    },
    beforeUpload,
    onChange(info) {
      if (info.file.status === 'done') {
        handleUploadSuccess()
      }
    },
    customRequest: async ({ file, onSuccess, onError }) => {
      const formData = new FormData()
      formData.append('file', file)
      formData.append('process_type', processType)

      try {
        const res = await fetch(`${API_BASE_URL}/upload_data`, {
          method: 'POST',
          body: formData,
        })
        if (!res.ok) throw new Error('上传失败')
        onSuccess?.(res.body)
      } catch (err) {
        handleUploadError()
        onError?.(err as Error)
      }
    },
    showUploadList: false,
    className: 'upload-dragger',
  }

  return (
    <Modal
      open={isUploadModalOpen}
      onCancel={closeUploadModal}
      footer={null}
      width={520}
      className="upload-modal"
      closeIcon={null}
      centered
    >
      {/* 自定义头部 */}
      <div className="upload-modal-header">
        <div className="upload-modal-header-left">
          <div className="upload-modal-icon">
            <UploadOutlined />
          </div>
          <div className="upload-modal-title">
            <h3>导入技能文件</h3>
            <span className="upload-modal-subtitle">支持 .md 或 .zip 格式</span>
          </div>
        </div>
        <Button
          type="text"
          onClick={closeUploadModal}
          className="upload-modal-close"
        >
          <CloseOutlined />
        </Button>
      </div>

      {/* 上传区域 */}
      <div className="upload-modal-body">
        <Dragger {...uploadProps} className="upload-dragger-modern">
          <div className="upload-dragger-content">
            <div className="upload-icon-wrapper">
              <InboxOutlined className="upload-main-icon" />
            </div>
            <p className="upload-text-primary">点击或拖拽文件到此区域上传</p>
            <p className="upload-text-secondary">
              支持 .md 或 .zip 格式，文件大小不超过 50MB
            </p>
            <div className="upload-format-hints">
              <span className="format-tag">
                <FileMarkdownOutlined /> Markdown
              </span>
              <span className="format-tag">
                <FileZipOutlined /> ZIP
              </span>
            </div>
          </div>
        </Dragger>

        <div className="upload-hint-box">
          <Text type="secondary">
            提示：导入时如果存在相同名称的技能，将会覆盖原有技能数据
          </Text>
        </div>
      </div>

      {/* 底部操作栏 */}
      <div className="upload-modal-footer">
        <Button onClick={closeUploadModal} className="upload-btn-cancel">
          取消
        </Button>
      </div>
    </Modal>
  )
}

export default UploadModal
