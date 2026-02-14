// config包管理应用配置
package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// 配置默认常量
const (
	// DefaultAddr 默认服务器监听地址
	DefaultAddr = "localhost:8080"
	// DefaultName 默认服务名称
	DefaultName = "aiflow"
	// DefaultVersion 默认版本号
	DefaultVersion = "0.5.0"
	// DefaultLogLevel 默认日志等级
	DefaultLogLevel = "info"
	// DBPath 默认数据库文件路径
	DBPath = "./db/aiflow.db"
)

// 有效日志等级集合
var validLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

// 有效输出类型集合
var validOutputTypes = map[string]bool{
	"std":  true,
	"file": true,
}

// Config 定义整个应用的配置结构
type Config struct {
	Server `yaml:"server"`
	Log    LogConfig `yaml:"log"`
	DB     DBConfig  `yaml:"db"`
}

// Server 定义服务器相关配置
type Server struct {
	Addr     string `yaml:"addr"`
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	RootPath string `yaml:"root_path"`
	McpPath  string `yaml:"mcp_path"`
	WebPath  string `yaml:"web_path"`
}

// LogConfig 定义日志相关配置
type LogConfig struct {
	Level      string `yaml:"level"`       // 日志等级：debug, info, warn, error
	OutputType string `yaml:"output_type"` // 输出形式：std（标准输出）或 file（文件输出）
	FilePath   string `yaml:"file_path"`   // 当output_type为file时，指定日志文件夹路径，日志文件名为main.log
}

// DBConfig 定义数据库相关配置
type DBConfig struct {
	Path string `yaml:"path"` // 数据库文件路径
}

// defaultConfig 内部默认配置
var defaultConfig = &Config{
	Server: Server{
		Addr:     DefaultAddr,
		Name:     DefaultName,
		Version:  DefaultVersion,
		RootPath: "/",
		McpPath:  "/mcp",
		WebPath:  "/web",
	},
	Log: LogConfig{
		Level:      DefaultLogLevel, // 默认日志等级
		OutputType: "std",           // 默认输出形式
		FilePath:   "",              // 默认文件路径
	},
	DB: DBConfig{
		Path: DBPath, // 默认数据库路径
	},
}

// FixWithDefault 修复Server配置的默认值
func (s *Server) FixWithDefault() {
	if s.Name == "" {
		s.Name = defaultConfig.Server.Name
	}
	if s.Version == "" {
		s.Version = defaultConfig.Server.Version
	}
	if s.RootPath == "" {
		s.RootPath = defaultConfig.Server.RootPath
	}
	if s.McpPath == "" {
		s.McpPath = defaultConfig.Server.McpPath
	}
	if s.WebPath == "" {
		s.WebPath = defaultConfig.Server.WebPath
	}
}

// Validate 验证配置有效性
// 检查服务器地址格式、日志等级、输出类型等配置项
func (c *Config) Validate() error {
	// 验证服务器地址格式
	if c.Server.Addr != "" {
		if _, err := net.ResolveTCPAddr("tcp", c.Server.Addr); err != nil {
			return fmt.Errorf("服务器地址格式无效 '%s': %w", c.Server.Addr, err)
		}
	}

	// 验证日志等级
	if c.Log.Level != "" {
		if !validLogLevels[strings.ToLower(c.Log.Level)] {
			return fmt.Errorf("无效的日志等级 '%s'，有效值: debug, info, warn, error", c.Log.Level)
		}
	}

	// 验证输出类型
	if c.Log.OutputType != "" {
		if !validOutputTypes[strings.ToLower(c.Log.OutputType)] {
			return fmt.Errorf("无效的输出类型 '%s'，有效值: std, file", c.Log.OutputType)
		}
	}

	return nil
}

// ApplyDefaults 应用默认值
// 为空字段设置默认值
func (c *Config) ApplyDefaults() {
	// 应用服务器默认值
	if c.Server.Addr == "" {
		c.Server.Addr = DefaultAddr
	}
	if c.Server.Name == "" {
		c.Server.Name = DefaultName
	}
	if c.Server.Version == "" {
		c.Server.Version = DefaultVersion
	}
	if c.Server.RootPath == "" {
		c.Server.RootPath = "/"
	}
	if c.Server.McpPath == "" {
		c.Server.McpPath = "/mcp"
	}
	if c.Server.WebPath == "" {
		c.Server.WebPath = "/web"
	}

	// 应用日志默认值
	if c.Log.Level == "" {
		c.Log.Level = DefaultLogLevel
	}
	if c.Log.OutputType == "" {
		c.Log.OutputType = "std"
	}

	// 应用数据库默认值
	if c.DB.Path == "" {
		c.DB.Path = DBPath
	}
}

// LoadFromEnv 从环境变量加载配置
// 支持的环境变量: AIFLOW_ADDR, AIFLOW_LOG_LEVEL, AIFLOW_LOG_OUTPUT, AIFLOW_DB_PATH
func (c *Config) LoadFromEnv() {
	// AIFLOW_ADDR -> Server.Addr
	if addr := os.Getenv("AIFLOW_ADDR"); addr != "" {
		c.Server.Addr = addr
	}

	// AIFLOW_LOG_LEVEL -> Log.Level
	if level := os.Getenv("AIFLOW_LOG_LEVEL"); level != "" {
		c.Log.Level = strings.ToLower(level)
	}

	// AIFLOW_LOG_OUTPUT -> Log.OutputType
	if output := os.Getenv("AIFLOW_LOG_OUTPUT"); output != "" {
		c.Log.OutputType = strings.ToLower(output)
	}

	// AIFLOW_DB_PATH -> DB.Path
	if dbPath := os.Getenv("AIFLOW_DB_PATH"); dbPath != "" {
		c.DB.Path = dbPath
	}
}

// FixWithDefault 修复Config配置的默认值
// 已废弃，请使用 ApplyDefaults 代替
func (c *Config) FixWithDefault() {
	c.ApplyDefaults()
}

// defaultConfigYAML 默认配置文件内容模板
const defaultConfigYAML = `# MCP HTTP服务器配置
server:
  # HTTP服务器监听地址
  addr: "localhost:9990"
  # MCP服务器名称
  name: "智流"
  # MCP服务器版本
  version: "0.5.0"
  # 路由配置
  root_path: "/"
  mcp_path: "/mcp"
  web_path: "/web"

log:
  # 日志等级，可选值：debug、info、warn、error
  level: "info"
  # 日志输出形式，可选值：std（标准输出）、file（文件输出）
  output_type: "file"
  # 当output_type为file时，指定日志文件夹路径，日志文件名为main.log
  file_path: "./logs"
`

// LoadConfig 从指定路径加载YAML配置文件
// 当配置文件不存在时，自动生成默认配置文件
// 加载流程: 读取文件 -> 解析YAML -> 加载环境变量 -> 应用默认值 -> 验证配置
func LoadConfig(path string) (Config, error) {
	// 读取配置文件内容
	content, err := os.ReadFile(path)
	if err != nil {
		// 文件不存在时，创建默认配置文件
		if os.IsNotExist(err) {
			// 确保目录存在
			dir := filepath.Dir(path)
			if dir != "" && dir != "." {
				if mkdirErr := os.MkdirAll(dir, 0755); mkdirErr != nil {
					return Config{}, fmt.Errorf("创建配置目录失败: %w", mkdirErr)
				}
			}
			// 写入默认配置文件
			if writeErr := os.WriteFile(path, []byte(defaultConfigYAML), 0644); writeErr != nil {
				return Config{}, fmt.Errorf("创建默认配置文件失败: %w", writeErr)
			}
			// 重新读取刚创建的文件
			content, err = os.ReadFile(path)
			if err != nil {
				return Config{}, err
			}
		} else {
			return Config{}, err
		}
	}

	// 解析YAML内容到Config结构体
	var cfg Config
	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		return Config{}, err
	}

	// 从环境变量加载配置（环境变量优先级高于配置文件）
	cfg.LoadFromEnv()

	// 应用默认值
	cfg.ApplyDefaults()

	// 验证配置有效性
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("配置验证失败: %w", err)
	}

	return cfg, nil
}

// GetDefaultConfig 返回默认配置，验证addr有效性，无效时使用默认值
func GetDefaultConfig(addr string) (Config, error) {
	cfg := *defaultConfig

	// 验证地址有效性
	if addr != "" {
		// 检查是否是有效的网络地址格式
		if _, err := net.ResolveTCPAddr("tcp", addr); err == nil {
			cfg.Server.Addr = addr
			return cfg, nil
		}
		// 地址无效，使用默认值并返回警告
		warning := fmt.Sprintf("警告: 提供的地址 '%s' 无效，使用默认地址 '%s'", addr, defaultConfig.Server.Addr)
		return cfg, fmt.Errorf("%s", warning)
	}

	// 地址为空，使用默认值
	return cfg, nil
}

// DefConfig 返回配置的默认副本
func DefConfig() *Config {
	return &Config{
		Server: Server{
			Addr:     DefaultAddr,
			Name:     DefaultName,
			Version:  DefaultVersion,
			RootPath: "/",
			McpPath:  "/mcp",
			WebPath:  "/web",
		},
		Log: LogConfig{
			Level:      DefaultLogLevel,
			OutputType: "std",
			FilePath:   "",
		},
		DB: DBConfig{
			Path: DBPath,
		},
	}
}
