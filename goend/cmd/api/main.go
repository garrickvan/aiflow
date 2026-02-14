package main

import (
	"aiflow/internal/api"
	"aiflow/internal/api/handlers"
	"aiflow/internal/config"
	"aiflow/internal/mcp"
	"aiflow/internal/repositories"
	"aiflow/internal/utils"
	"aiflow/internal/utils/logx"
	"embed"
	"flag"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mark3labs/mcp-go/server"
)

// todo: 后台提供重新分词按钮，点击后重新对所有技能描述进行分词
// todo: 后台提供目录导入skill能力
// todo: 提供zip包导入skill能力
// todo: 提供skill基准目录注入，确保脚本能正常运行
// todo: 添加平台参数，细化跟踪信息
// todo: 编号生成，先大写处理英文

var (
	// httpAddr 定义HTTP服务器监听地址
	httpAddr = flag.String("http", "localhost:9900", "HTTP服务器监听地址")
	// configPath 定义配置文件路径
	configPath = flag.String("config", "./config.yml", "配置文件路径")
	// config 全局配置实例
	appConfig config.Config
)

// 嵌入前端静态资源
// 注意：构建脚本会将前端构建文件复制到这个位置
//
//go:embed static
var staticFiles embed.FS

// main 函数是程序入口点，初始化并启动MCP HTTP服务器
// 功能：加载配置、初始化日志、创建服务器实例、配置路由、初始化数据库连接
func main() {
	flag.Parse()

	// 加载配置文件
	var err error
	appConfig, err = config.LoadConfig(*configPath)
	if err != nil {
		// 在日志系统初始化前使用标准log
		logx.Error("警告: 无法加载配置文件 %s, 使用默认配置: %v", *configPath, err)
		// 使用默认配置
		var warning error
		appConfig, warning = config.GetDefaultConfig(*httpAddr)
		if warning != nil {
			// 在日志系统初始化前使用标准log
			logx.Error("%s", warning.Error())
		}
	}

	// 初始化日志系统
	// 注意：当outputType为file时，appConfig.Log.FilePath被视为日志文件夹路径，日志文件名为main.log
	logx.InitLogger(appConfig.Log.Level, appConfig.Log.OutputType, appConfig.Log.FilePath)

	// 创建MCP服务器实例
	mcpServer := server.NewMCPServer(appConfig.Server.Name, appConfig.Server.Version)

	// 创建HTTP服务器
	httpServer := server.NewStreamableHTTPServer(mcpServer)

	// 创建chi路由器
	r := chi.NewRouter()

	// 添加中间件
	r.Use(middleware.RequestID) // 请求ID中间件
	if appConfig.Log.Level == "debug" {
		r.Use(middleware.Logger) // 开启调试日志
	}
	r.Use(middleware.Recoverer) // 恢复中间件，防止服务器崩溃
	// 添加CORS中间件
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// 配置HTTP路由
	// 根路径和MCP路径使用MCP服务器处理
	r.Handle(appConfig.Server.RootPath, httpServer)
	r.Handle(appConfig.Server.McpPath, httpServer)

	// 设置静态文件系统并注册WebHandler
	handlers.SetStaticFS(staticFiles)
	// 注册WebHandler处理/web路径的请求
	r.HandleFunc("/web", handlers.WebHandler)
	r.HandleFunc("/web/*", handlers.WebHandler)

	// 初始化数据库连接
	utils.CreateIfNotExist(config.DBPath)
	repo, err := repositories.NewRepository(config.DBPath)
	if err != nil {
		logx.Error("初始化数据库失败: %v", err)
		// 即使数据库初始化失败，也创建一个空的repo用于注册路由
		// 这样API会返回错误而不是404
		repo = repositories.NewEmptyRepository()
	}

	// 添加基础工具
	mcp.InitTools(mcpServer, repo)
	// 注册API路由（无论数据库是否初始化成功都注册）
	apiRouter := api.NewRouter(repo)
	apiRouter.RegisterRoutes(r)

	// 确定最终使用的监听地址
	listenAddr := appConfig.Server.Addr
	if listenAddr == "" {
		listenAddr = *httpAddr
	}

	// 启动HTTP服务器，打印核心配置信息
	logx.Info("MCP服务: %s v%s", appConfig.Server.Name, appConfig.Server.Version)
	logx.Info("服务路径:")
	logx.Info("  Web后台: http://%s%s", listenAddr, appConfig.Server.WebPath)
	logx.Info("  根路径: http://%s%s", listenAddr, appConfig.Server.RootPath)
	logx.Info("  MCP路径: http://%s%s", listenAddr, appConfig.Server.McpPath)
	logx.Info("日志配置:")
	logx.Info("  级别: %s, 输出: %s", appConfig.Log.Level, appConfig.Log.OutputType)
	if appConfig.Log.FilePath != "" && appConfig.Log.OutputType == "file" {
		logx.Info("  文件路径: %s", appConfig.Log.FilePath)
	}

	// 初始化系统托盘
	utils.InitTray(&appConfig)

	// 启动HTTP服务器（在后台运行）
	go func() {
		if err := http.ListenAndServe(listenAddr, r); err != nil {
			logx.Fatal("服务器启动失败: %v", err)
		}
	}()

	// 等待退出信号
	utils.WaitForExit()

	logx.Info("AiFlow服务已成功退出")
}
