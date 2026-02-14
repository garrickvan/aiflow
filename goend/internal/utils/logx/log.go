// logx 包提供日志功能，支持不同级别的日志输出
package logx

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// 日志等级常量
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// 全局日志实例
var (
	DebugLogger *log.Logger = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	InfoLogger  *log.Logger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	WarnLogger  *log.Logger = log.New(os.Stdout, "[WARN] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	ErrorLogger *log.Logger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lmicroseconds)
)

// InitLogger 初始化日志系统，根据配置的日志等级和输出形式设置日志输出
// logLevel: 日志等级，可选值：debug、info、warn、error
// outputType: 输出形式，可选值：std（标准输出）或 file（文件输出），默认std
// logDirPath: 当outputType为file时，指定日志文件夹路径，日志文件名为main.log
func InitLogger(logLevel string, outputType string, logDirPath string) {
	// 初始化输出目标
	var stdout io.Writer = os.Stdout
	var stderr io.Writer = os.Stderr

	// 处理输出形式
	if outputType == "file" && logDirPath != "" {
		// 确保日志文件夹存在
		if err := os.MkdirAll(logDirPath, 0755); err != nil {
			// 如果文件夹创建失败，回退到标准输出
			log.Printf("日志文件夹创建失败: %v, 回退到标准输出", err)
		} else {
			// 构建完整的日志文件路径
			logFilePath := logDirPath + "/main.log"
			// 打开或创建日志文件
			file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				// 如果文件打开失败，回退到标准输出
				log.Printf("日志文件打开失败: %v, 回退到标准输出", err)
			} else {
				stdout = file
				stderr = file
			}
		}
	}

	// 创建不同级别的日志实例
	DebugLogger = log.New(stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	InfoLogger = log.New(stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	WarnLogger = log.New(stdout, "[WARN] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	ErrorLogger = log.New(stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lmicroseconds)

	// 获取配置的日志等级
	logLevel = strings.ToLower(logLevel)

	// 根据日志等级禁用不需要的日志
	switch logLevel {
	case LogLevelError:
		DebugLogger.SetOutput(io.Discard)
		InfoLogger.SetOutput(io.Discard)
		WarnLogger.SetOutput(io.Discard)
	case LogLevelWarn:
		DebugLogger.SetOutput(io.Discard)
		InfoLogger.SetOutput(io.Discard)
	case LogLevelInfo:
		DebugLogger.SetOutput(io.Discard)
	case LogLevelDebug:
		// 调试级别，所有日志都启用
	default:
		// 默认使用info级别
		DebugLogger.SetOutput(io.Discard)
	}
}

// Debug 输出调试级别的日志
func Debug(format string, v ...interface{}) {
	DebugLogger.Printf(format, v...)
}

// Info 输出信息级别的日志
func Info(format string, v ...interface{}) {
	InfoLogger.Printf(format, v...)
}

// Warn 输出警告级别的日志
func Warn(format string, v ...interface{}) {
	WarnLogger.Printf(format, v...)
}

// Error 输出错误级别的日志
func Error(format string, v ...interface{}) {
	ErrorLogger.Printf(format, v...)
}

// Fatal 输出错误级别的日志并退出程序
func Fatal(format string, v ...interface{}) {
	ErrorLogger.Fatalf(format, v...)
}

// formatWithRequestID 格式化日志消息，添加请求ID前缀
// 如果请求ID为空，则直接返回原始格式化消息
func formatWithRequestID(requestID, format string, v ...interface{}) string {
	msg := fmt.Sprintf(format, v...)
	if requestID == "" {
		return msg
	}
	return fmt.Sprintf("[%s] %s", requestID, msg)
}

// DebugCtx 输出带上下文的调试级别日志
// 自动从上下文中提取请求ID并添加到日志前缀
func DebugCtx(ctx context.Context, format string, v ...interface{}) {
	requestID := getRequestIDFromContext(ctx)
	DebugLogger.Printf(formatWithRequestID(requestID, format, v...))
}

// InfoCtx 输出带上下文的信息级别日志
// 自动从上下文中提取请求ID并添加到日志前缀
func InfoCtx(ctx context.Context, format string, v ...interface{}) {
	requestID := getRequestIDFromContext(ctx)
	InfoLogger.Printf(formatWithRequestID(requestID, format, v...))
}

// WarnCtx 输出带上下文的警告级别日志
// 自动从上下文中提取请求ID并添加到日志前缀
func WarnCtx(ctx context.Context, format string, v ...interface{}) {
	requestID := getRequestIDFromContext(ctx)
	WarnLogger.Printf(formatWithRequestID(requestID, format, v...))
}

// ErrorCtx 输出带上下文的错误级别日志
// 自动从上下文中提取请求ID并添加到日志前缀
func ErrorCtx(ctx context.Context, format string, v ...interface{}) {
	requestID := getRequestIDFromContext(ctx)
	ErrorLogger.Printf(formatWithRequestID(requestID, format, v...))
}

// ctxKey 上下文键类型（与middleware包保持一致）
type ctxKey string

// requestIDKey 请求ID在上下文中的键
const requestIDKey ctxKey = "request_id"

// getRequestIDFromContext 从上下文获取请求ID
// 返回请求ID字符串，如果上下文中不存在则返回空字符串
func getRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}
