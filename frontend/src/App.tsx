import {
  useState,
  useEffect,
  useCallback,
  Suspense,
  lazy,
} from "react";
import { useSearchParams } from "react-router-dom";
import { Menu, Spin, Layout, Button, Drawer } from "antd";
import { ToolOutlined, FileTextOutlined, MenuOutlined, CloseOutlined } from "@ant-design/icons";

const { Header, Sider, Content } = Layout;
import type { MenuProps } from "antd";
import "./App.css";

// 导入Zustand Store
import { useAppStore } from "./stores/appStore";

// 导入组件
import SkillModal from "./components/SkillModal";
import GroupModal from "./components/GroupModal";
import UploadModal from "./components/UploadModal";
import JobTaskModal from "./components/JobTaskModal";
import TagManagementModal from "./components/TagManagementModal";

// 异步加载页面组件，避免一次性过多渲染
const SkillManagement = lazy(() => import("./pages/SkillManagement"));
const JobTaskManagement = lazy(() => import("./pages/JobTaskManagement"));

/**
 * 页面参数配置常量
 */
const PAGE_PARAM_KEY = "page";
const PAGE_SKILL = "skill";
const PAGE_JOB = "job";
const PAGE_DEFAULT = PAGE_SKILL;

/**
 * 主应用组件
 * 使用查询参数进行页面导航，使用Zustand进行状态管理
 */
function App() {
  const [searchParams, setSearchParams] = useSearchParams();

  // 从Zustand Store获取状态和actions
  const {
    collapsed,
    mobileMenuOpen,
    toggleCollapsed,
    setMobileMenuOpen,
  } = useAppStore();

  // 窗口宽度状态（本地状态，不涉及全局共享）
  const [windowWidth, setWindowWidth] = useState(window.innerWidth);
  // 移动端断点
  const MOBILE_BREAKPOINT = 768;

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

  // 判断是否为移动端
  const isMobile = windowWidth < MOBILE_BREAKPOINT;

  /**
   * 获取当前页面参数
   */
  const getCurrentPage = useCallback((): string => {
    const page = searchParams.get(PAGE_PARAM_KEY);
    if (page === PAGE_JOB) {
      return PAGE_JOB;
    }
    return PAGE_SKILL;
  }, [searchParams]);

  /**
   * 切换到指定页面
   * @param page - 目标页面
   */
  const navigateToPage = useCallback(
    (page: string) => {
      setSearchParams({ [PAGE_PARAM_KEY]: page });
    },
    [setSearchParams],
  );

  // 初始化时，如果没有page参数，设置为默认值
  useEffect(() => {
    const currentPage = searchParams.get(PAGE_PARAM_KEY);
    if (!currentPage) {
      setSearchParams({ [PAGE_PARAM_KEY]: PAGE_DEFAULT });
    }
  }, [searchParams, setSearchParams]);

  // 菜单项定义
  const menuItems: MenuProps["items"] = [
    {
      key: PAGE_SKILL,
      icon: <ToolOutlined />,
      label: "技能管理",
      onClick: () => {
        navigateToPage(PAGE_SKILL);
        if (isMobile) {
          setMobileMenuOpen(false);
        }
      },
    },
    {
      key: PAGE_JOB,
      icon: <FileTextOutlined />,
      label: "任务管理",
      onClick: () => {
        navigateToPage(PAGE_JOB);
        if (isMobile) {
          setMobileMenuOpen(false);
        }
      },
    },
  ];

  // 获取当前页面
  const currentPage = getCurrentPage();

  return (
    <Layout style={{ minHeight: "100vh", overflow: "hidden" }}>
      {/* 移动端汉堡菜单按钮 */}
      {isMobile && (
        <Button
          className="mobile-menu-btn"
          type="primary"
          icon={<MenuOutlined />}
          onClick={() => setMobileMenuOpen(true)}
        />
      )}

      {/* 移动端侧边栏抽屉 */}
      <Drawer
        className="mobile-sider-drawer"
        placement="left"
        closable={false}
        onClose={() => setMobileMenuOpen(false)}
        open={mobileMenuOpen}
        size={200}
        styles={{
          body: {
            padding: 0,
            backgroundColor: "var(--color-sider-bg)",
          },
          header: {
            display: "none",
          },
        }}
      >
        {/* 移动端抽屉头部 */}
        <Header
          className="app-header mobile-header"
          style={{
            color: "white",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            textAlign: "center",
            boxShadow: "var(--shadow-md)",
            zIndex: 1,
            transition: "var(--transition-base)",
            padding: "16px",
            borderBottom: "1px solid rgba(255, 255, 255, 0.1)",
          }}
        >
          <div
            className="app-logo"
            style={{
              fontSize: "18px",
              fontWeight: "600",
              whiteSpace: "nowrap",
              display: "flex",
              alignItems: "center",
              gap: "8px",
            }}
          >
            <img
              src="/web/static/skill.svg"
              alt="智流MCP"
              style={{
                width: "32px",
                height: "32px",
                transition: "var(--transition-base)",
              }}
            />
            {"智流MCP"}
          </div>
          <Button
            type="text"
            icon={<CloseOutlined style={{ color: "white" }} />}
            onClick={() => setMobileMenuOpen(false)}
          />
        </Header>
        <Menu
          mode="inline"
          selectedKeys={[currentPage]}
          items={menuItems}
          style={{
            height: "100%",
            borderRight: 0,
            backgroundColor: "var(--color-sider-bg)",
            padding: "8px 0",
          }}
          theme="dark"
          className="app-menu"
        />
      </Drawer>

      {/* 主体内容 */}
      <Layout style={{ flex: 1, display: "flex" }}>
        {/* 桌面端侧边栏 */}
        {!isMobile && (
          <Sider
            className="app-sider"
            collapsible
            collapsed={collapsed}
            onCollapse={(value) => {
              if (value !== collapsed) {
                toggleCollapsed();
              }
            }}
            style={{
              backgroundColor: "var(--color-sider-bg)",
              boxShadow: "var(--shadow-lg)",
              flexShrink: 0,
              transition: "var(--transition-base)",
              position: "fixed",
              left: 0,
              top: 0,
              bottom: 0,
              zIndex: 100,
            }}
          >
            {/* 头部 */}
            <Header
              className="app-header"
              style={{
                color: "white",
                display: "flex",
                justifyContent: "center",
                alignItems: "center",
                textAlign: "center",
                boxShadow: "var(--shadow-md)",
                zIndex: 1,
                transition: "var(--transition-base)",
                padding: "16px 0",
                borderBottom: "1px solid rgba(255, 255, 255, 0.1)",
              }}
            >
              <div
                className="app-logo"
                style={{
                  fontSize: "18px",
                  fontWeight: "600",
                  whiteSpace: "nowrap",
                  display: "flex",
                  alignItems: "center",
                  gap: "8px",
                }}
              >
                <img
                  src="/web/static/skill.svg"
                  alt="智流MCP"
                  style={{
                    width: "32px",
                    height: "32px",
                    transition: "var(--transition-base)",
                  }}
                />
                {!collapsed && "智流MCP"}
              </div>
            </Header>
            <Menu
              mode="inline"
              selectedKeys={[currentPage]}
              items={menuItems}
              style={{
                height: "100%",
                borderRight: 0,
                backgroundColor: "var(--color-sider-bg)",
                padding: "8px 0",
              }}
              theme="dark"
              className="app-menu"
            />
          </Sider>
        )}

        {/* 右侧内容 */}
        <Layout
          className="main-content-layout"
          style={{
            flex: 1,
            display: "flex",
            flexDirection: "column",
            backgroundColor: "var(--color-gray-100)",
            transition: "var(--transition-base)",
            marginLeft: isMobile ? 0 : collapsed ? "80px" : "200px",
          }}
        >
          <Content
            style={{
              flex: 1,
              overflow: "auto",
              display: "flex",
              flexDirection: "column",
            }}
          >
            {/* 内容区域 - 使用Suspense处理异步组件加载 */}
            <Suspense
              fallback={
                <div
                  style={{
                    display: "flex",
                    justifyContent: "center",
                    alignItems: "center",
                    height: "400px",
                  }}
                >
                  <Spin size="large" />
                </div>
              }
            >
              {/* 技能管理页面 */}
              {currentPage === PAGE_SKILL && <SkillManagement />}

              {/* 任务管理页面 */}
              {currentPage === PAGE_JOB && <JobTaskManagement />}
            </Suspense>
          </Content>
        </Layout>
      </Layout>

      {/* 技能模态框 */}
      <SkillModal />

      {/* 标签管理模态框 */}
      <TagManagementModal />

      {/* 标签编辑模态框 */}
      <GroupModal />

      {/* 导入技能模态框 */}
      <UploadModal />

      {/* 任务模态框 */}
      <JobTaskModal />
    </Layout>
  );
}

export default App;
