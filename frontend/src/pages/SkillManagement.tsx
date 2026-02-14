/**
 * 技能管理页面组件
 * 采用卡片网格布局展示技能列表，使用Zustand状态管理
 */
import React, { useState, useEffect, useRef, useCallback } from "react";
import {
  Button,
  Popconfirm,
  Select,
  Empty,
  Pagination,
  Card,
  Row,
  Col,
  Space,
  DatePicker,
} from "antd";
import type { Dayjs } from "dayjs";
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  UploadOutlined,
  DownloadOutlined,
  AppstoreOutlined,
  FileTextOutlined,
  TagOutlined,
  ClockCircleOutlined,
  ReloadOutlined,
  VerticalAlignTopOutlined,
  RollbackOutlined,
  CloseCircleOutlined,
  RestOutlined,
} from "@ant-design/icons";
import { useAppStore } from "../stores/appStore";
import { useModalStore } from "../stores/modalStore";
import type { Skill } from "../types";
import { skillApi, tagApi } from "../services";
import { formatTime, getSkillAvatarConfig } from "../utils";
import { message } from "antd";

const { RangePicker } = DatePicker;

/**
 * 技能卡片组件
 */
const SkillCard: React.FC<{
  skill: Skill;
  onEdit: (skill: Skill) => void;
  onDelete: (id: number) => void;
  onExport: (id: number) => void;
  isTrashMode?: boolean;
  onRestore?: (id: number, onSuccess?: () => void) => void;
  onPermanentDelete?: (id: number, onSuccess?: () => void) => void;
  onRefresh?: () => void;
}> = ({ skill, onEdit, onDelete, onExport, isTrashMode = false, onRestore, onPermanentDelete, onRefresh }) => {
  const avatarConfig = getSkillAvatarConfig(skill.name);

  return (
    <div className="skill-card">
      <div className="skill-card-header">
        <div className="skill-identity">
          <div
            className="skill-avatar"
            style={{ background: avatarConfig.gradient }}
          >
            {avatarConfig.icon}
          </div>
          <div className="skill-title-group">
            <span className="skill-name">{skill.name}</span>
            <span className="skill-dir">{skill.resourceDir}</span>
          </div>
        </div>
        <div className="skill-tags">
          {skill.tags?.slice(0, 3).map((tag) => (
            <span key={tag.id} className="skill-tag">
              {tag.name}
            </span>
          ))}
          {skill.tags && skill.tags.length > 3 && (
            <span className="skill-tag more">+{skill.tags.length - 3}</span>
          )}
        </div>
      </div>

      <div className="skill-description">
        {skill.description}
      </div>

      {skill.compatibility && (
        <div className="compatibility-alert">
          <span>⚡</span>
          <span>{skill.compatibility}</span>
        </div>
      )}

      <div className="skill-footer">
        <div className="skill-time">创建于 {formatTime(skill.createdAt)}</div>
        <div className="skill-actions">
          {!isTrashMode ? (
            <>
              <button
                className="btn-icon"
                title="编辑"
                onClick={() => onEdit(skill)}
              >
                <EditOutlined />
              </button>
              <button
                className="btn-icon export"
                title="导出"
                onClick={() => onExport(skill.id)}
              >
                <DownloadOutlined />
              </button>
              <Popconfirm
                title="确定要删除这个技能吗？删除后可在回收站恢复。"
                onConfirm={() => onDelete(skill.id)}
                okText="确定"
                cancelText="取消"
                placement="topRight"
              >
                <button className="btn-icon delete" title="删除">
                  <DeleteOutlined />
                </button>
              </Popconfirm>
            </>
          ) : (
            <>
              <Popconfirm
                title="确定要恢复这个技能吗？"
                onConfirm={() => onRestore?.(skill.id, onRefresh)}
                okText="确定"
                cancelText="取消"
                placement="topRight"
              >
                <button className="btn-icon" title="恢复" style={{ color: '#10b981' }}>
                  <RollbackOutlined />
                </button>
              </Popconfirm>
              <Popconfirm
                title="确定要彻底删除这个技能吗？删除后无法恢复！"
                onConfirm={() => onPermanentDelete?.(skill.id, onRefresh)}
                okText="确定"
                cancelText="取消"
                placement="topRight"
              >
                <button className="btn-icon delete" title="彻底删除">
                  <CloseCircleOutlined />
                </button>
              </Popconfirm>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

/**
 * 统计卡片组件
 */
const StatCard: React.FC<{
  title: string;
  value: number | string;
  icon: React.ReactNode;
  color: string;
  subtitle?: string;
}> = ({ title, value, icon, color, subtitle }) => (
  <Card className="stat-card" variant="borderless">
    <div className="stat-content">
      <div>
        <div className="stat-title">{title}</div>
        <div className="stat-value" style={{ color }}>{value}</div>
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

/**
 * 技能管理页面组件
 */
const SkillManagement: React.FC = () => {
  // 从Zustand Store获取状态和actions
  const {
    skills,
    tags,
    selectedTagId,
    selectedSkillDateRange,
    skillPagination,
    collapsed,
    setSkills,
    setTags,
    setSelectedTagId,
    setSelectedSkillDateRange,
    setSkillPagination,
  } = useAppStore();
  const { openSkillModal, openUploadModal, openTagManagementModal } = useModalStore();

  // 对 skills 做空值保护，确保始终是数组
  const safeSkills = skills || [];

  // 计算统计数据
  const totalCount = skillPagination.total;
  const tagsCount = tags.length;

  // 筛选栏悬浮状态
  const [isFilterSticky, setIsFilterSticky] = useState(false);
  const sentinelRef = useRef<HTMLDivElement>(null);

  // 窗口宽度状态
  const [windowWidth, setWindowWidth] = useState(window.innerWidth);
  // 移动端断点
  const MOBILE_BREAKPOINT = 768;
  // 回到顶部按钮显示状态
  const [showBackToTop, setShowBackToTop] = useState(false);
  // 页面容器ref
  const containerRef = useRef<HTMLDivElement>(null);
  // 回收站模式状态
  const [isTrashMode, setIsTrashMode] = useState(false);

  // 解构分页状态
  const { page: skillPage, pageSize: skillPageSize, total: skillTotal } = skillPagination;

  /**
   * 加载技能数据
   */
  const loadSkillData = useCallback(async (
    tagId?: number,
    dateRange?: [Dayjs | null, Dayjs | null] | null,
    page?: number,
    pageSize?: number,
  ) => {
    try {
      const currentPage = page ?? skillPage;
      const currentPageSize = pageSize ?? skillPageSize;
      const startDate = dateRange?.[0]?.valueOf();
      const endDate = dateRange?.[1]?.valueOf();
      const [loadedSkills, loadedTags] = await Promise.all([
        skillApi.getSkills(tagId, currentPage, currentPageSize, startDate, endDate),
        tagApi.getTags(1, 100),
      ]);
      setSkills(loadedSkills.items);
      setSkillPagination({ total: loadedSkills.pagination.total });
      setTags(loadedTags.items);
    } catch (error) {
      message.error("加载技能数据失败");
      console.error("Load skill data failed:", error);
    }
  }, [skillPage, skillPageSize, setSkills, setTags, setSkillPagination]);

  /**
   * 加载回收站技能列表
   */
  const loadTrashSkills = useCallback(async (page?: number, pageSize?: number) => {
    try {
      const currentPage = page ?? skillPage;
      const currentPageSize = pageSize ?? skillPageSize;
      const loadedTrashSkills = await skillApi.getTrashSkills(currentPage, currentPageSize);
      setSkills(loadedTrashSkills.items);
      setSkillPagination({ total: loadedTrashSkills.pagination.total });
    } catch (error) {
      message.error("加载回收站数据失败");
      console.error("Load trash skills failed:", error);
    }
  }, [skillPage, skillPageSize, setSkills, setSkillPagination]);

  /**
   * 使用IntersectionObserver API监听哨兵元素，实现筛选栏悬浮
   */
  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        setIsFilterSticky(!entry.isIntersecting);
      },
      {
        root: null,
        threshold: 0,
        rootMargin: '0px 0px 0px 0px',
      }
    );

    observer.observe(sentinel);

    return () => {
      observer.disconnect();
    };
  }, []);

  /**
   * 监听窗口大小变化
   */
  useEffect(() => {
    const handleResize = () => {
      setWindowWidth(window.innerWidth);
    };

    window.addEventListener("resize", handleResize);
    return () => {
      window.removeEventListener("resize", handleResize);
    };
  }, []);

  /**
   * 监听页面滚动，控制回到顶部按钮显示
   */
  useEffect(() => {
    const container = document.body;

    const handleScroll = () => {
      setShowBackToTop(container.scrollTop > 300);
    };

    container.addEventListener("scroll", handleScroll);
    handleScroll();

    return () => {
      container.removeEventListener("scroll", handleScroll);
    };
  }, []);

  /**
   * 刷新回收站数据
   */
  const refreshTrashData = useCallback(() => {
    if (isTrashMode) {
      loadTrashSkills(skillPage, skillPageSize);
    }
  }, [isTrashMode, loadTrashSkills, skillPage, skillPageSize]);

  /**
   * 组件挂载时加载数据
   */
  useEffect(() => {
    loadSkillData(selectedTagId, selectedSkillDateRange, skillPage, skillPageSize);
  }, []);

  /**
   * 切换回收站模式时加载对应数据
   */
  useEffect(() => {
    if (isTrashMode) {
      loadTrashSkills(skillPage, skillPageSize);
    } else {
      // 从回收站返回列表时，刷新普通列表数据
      loadSkillData(selectedTagId, selectedSkillDateRange, skillPage, skillPageSize);
      setTimeout(() => {
        const sentinel = sentinelRef.current;
        if (sentinel) {
          const rect = sentinel.getBoundingClientRect();
          setIsFilterSticky(rect.bottom <= 0);
        }
      }, 0);
    }
  }, [isTrashMode]);

  /**
   * 回到顶部
   */
  const handleBackToTop = (): void => {
    document.body.scrollTo({
      top: 0,
      behavior: "smooth",
    });
  };

  // 判断是否为移动端
  const isMobile = windowWidth < MOBILE_BREAKPOINT;

  /**
   * 处理编辑技能
   */
  const handleEditSkill = (skill: Skill) => {
    openSkillModal(skill);
  };

  /**
   * 处理删除技能
   */
  const handleDeleteSkill = async (id: number) => {
    try {
      await skillApi.deleteSkill(id);
      // 使用函数式更新避免闭包问题，确保基于最新状态更新
      setSkills((prevSkills) => prevSkills.filter((skill) => skill.id !== id));
      message.success("技能已移至回收站");
    } catch (error) {
      message.error("删除失败，请重试");
      console.error("Skill delete failed:", error);
    }
  };

  /**
   * 处理导出技能
   */
  const handleExportSkill = async (id: number) => {
    try {
      await skillApi.exportSkill(id);
    } catch (error) {
      message.error("导出失败，请重试");
      console.error("Skill export failed:", error);
    }
  };

  /**
   * 恢复回收站中的技能
   */
  const handleRestoreSkill = async (id: number, onSuccess?: () => void) => {
    try {
      await skillApi.restoreSkill(id);
      message.success("技能恢复成功");
      onSuccess?.();
    } catch (error) {
      message.error("恢复失败，请重试");
      console.error("Skill restore failed:", error);
    }
  };

  /**
   * 彻底删除技能
   */
  const handlePermanentDeleteSkill = async (id: number, onSuccess?: () => void) => {
    try {
      await skillApi.permanentDeleteSkill(id);
      message.success("技能已彻底删除");
      onSuccess?.();
    } catch (error) {
      message.error("删除失败，请重试");
      console.error("Skill permanent delete failed:", error);
    }
  };

  /**
   * 处理分页变化
   */
  const handlePageChange = (page: number, pageSize: number) => {
    setSkillPagination({ page, pageSize });
    if (isTrashMode) {
      loadTrashSkills(page, pageSize);
    } else {
      loadSkillData(selectedTagId, selectedSkillDateRange, page, pageSize);
    }
  };

  /**
   * 处理标签变化
   */
  const handleTagChange = (tagId: number | undefined) => {
    setSelectedTagId(tagId);
    setSkillPagination({ page: 1 });
    loadSkillData(tagId, selectedSkillDateRange, 1, skillPageSize);
  };

  /**
   * 处理日期范围变化
   */
  const handleDateRangeChange = (dates: [Dayjs | null, Dayjs | null] | null) => {
    setSelectedSkillDateRange(dates);
    setSkillPagination({ page: 1 });
    loadSkillData(selectedTagId, dates, 1, skillPageSize);
  };

  /**
   * 处理重置筛选
   */
  const handleResetFilters = () => {
    setSelectedTagId(undefined);
    setSelectedSkillDateRange(null);
    setSkillPagination({ page: 1 });
    loadSkillData(undefined, null, 1, skillPageSize);
  };

  return (
    <div className="skill-management" ref={containerRef}>
      {/* 头部区域 */}
      <div className="skill-header">
        <div className="header-top">
          <h1 className="header-title">{isTrashMode ? '回收站' : '技能管理'}</h1>
          <Space className="header-actions">
            <Button
              className="btn-secondary-gradient"
              icon={<RestOutlined />}
              onClick={() => setIsTrashMode(!isTrashMode)}
            >
              {isTrashMode ? '返回列表' : '回收站'}
            </Button>
            {!isTrashMode && (
              <>
                <Button
                  className="btn-secondary-gradient"
                  icon={<UploadOutlined />}
                  onClick={openUploadModal}
                >
                  导入
                </Button>
                <Button
                  type="primary"
                  className="btn-primary-gradient"
                  icon={<PlusOutlined />}
                  onClick={() => openSkillModal()}
                >
                  新增技能
                </Button>
              </>
            )}
          </Space>
        </div>
      </div>

      {/* 筛选栏 - 独立出来实现sticky效果，小屏状态下不浮动，回收站模式下不显示 */}
      {!isTrashMode && (
        <div
          className={`filter-bar-wrapper ${isFilterSticky && !isMobile ? 'filter-bar-wrapper-sticky' : ''}`}
          style={{ '--sider-width': collapsed ? '80px' : '200px' } as React.CSSProperties}
        >
          {/* 哨兵元素 - 用于IntersectionObserver检测滚动位置 */}
          <div ref={sentinelRef} className="filter-sentinel" />
          <div className="filter-bar">
            <div className="filter-group">
              <span className="filter-label">标签</span>
              <Select
                className="filter-select"
                placeholder="选择标签"
                value={selectedTagId}
                onChange={(value) => handleTagChange(value || undefined)}
                options={[
                  { value: '', label: '全部标签' },
                  ...tags.map((tag) => ({
                    value: tag.id,
                    label: tag.name,
                  })),
                ]}
                variant="borderless"
              />
            </div>
            <div className="filter-group">
              <span className="filter-label">创建时间</span>
              <RangePicker
                className="filter-date-range"
                value={selectedSkillDateRange}
                onChange={handleDateRangeChange}
                placeholder={['开始时间', '结束时间']}
                format="YYYY-MM-DD HH:mm"
                showTime={{ format: 'HH:mm' }}
              />
            </div>
            <Button
              className="btn-secondary-gradient"
              icon={<AppstoreOutlined />}
              onClick={openTagManagementModal}
            >
              管理标签
            </Button>
            <Button
              className="btn-secondary-gradient"
              icon={<ReloadOutlined />}
              onClick={handleResetFilters}
            >
              重置
            </Button>
          </div>

        </div>
      )}

      {/* 统计卡片 - 回收站模式下不显示 */}
      {!isTrashMode && (
        <Row gutter={[20, 20]} className="stats-row">
          <Col xs={12} sm={8} lg={8}>
            <StatCard
              title="技能总数"
              value={totalCount}
              icon={<FileTextOutlined />}
              color="#8b5cf6"
            />
          </Col>
          <Col xs={12} sm={8} lg={8}>
            <StatCard
              title="活跃标签"
              value={tagsCount}
              icon={<TagOutlined />}
              color="#10b981"
            />
          </Col>
          <Col xs={12} sm={8} lg={8}>
            <StatCard
              title="当前展示"
              value={safeSkills.length}
              icon={<ClockCircleOutlined />}
              color="#f59e0b"
            />
          </Col>
        </Row>
      )}

      {/* 技能卡片网格 */}
      {safeSkills.length > 0 ? (
        <>
          <div className="skills-grid">
            {safeSkills.map((skill) => (
              <SkillCard
                key={skill.id}
                skill={skill}
                onEdit={handleEditSkill}
                onDelete={handleDeleteSkill}
                onExport={handleExportSkill}
                isTrashMode={isTrashMode}
                onRestore={handleRestoreSkill}
                onPermanentDelete={handlePermanentDeleteSkill}
                onRefresh={refreshTrashData}
              />
            ))}
          </div>

          {/* 分页 */}
          <div className="pagination-wrapper">
            <Pagination
              total={skillTotal}
              current={skillPage}
              pageSize={skillPageSize}
              showSizeChanger
              showQuickJumper
              onChange={handlePageChange}
              pageSizeOptions={['20', '50', '100']}
              locale={{
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
              }}
            />
          </div>
        </>
      ) : (
        <Empty
          className="skill-empty"
          description={isTrashMode ? "回收站为空" : "暂无技能数据"}
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
      )}

      {/* 回到顶部按钮 */}
      {showBackToTop && (
        <button
          className="back-to-top-btn"
          onClick={handleBackToTop}
          title="回到顶部"
        >
          <VerticalAlignTopOutlined />
        </button>
      )}
    </div>
  );
};

export default SkillManagement;
