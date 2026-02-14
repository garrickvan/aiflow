package utils

import (
	"aiflow/internal/config"
	"aiflow/internal/utils/logx"
	"encoding/base64"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/getlantern/systray"
)

var (
	// 服务状态
	serverRunning = false
	// 退出通道
	exitChan = make(chan struct{})
	// 退出通道是否已关闭
	exitChanClosed = false
	_cfg, _        = config.GetDefaultConfig("")
)

// 简单的默认图标 (base64编码的16x16像素图标)
const iconBase64 = "AAABAAEAGBgAAAEAIADaAwAAFgAAAIlQTkcNChoKAAAADUlIRFIAAAAYAAAAGAgGAAAA4Hc9+AAAAAlwSFlzAAAOxAAADsQBlSsOGwAAA4xJREFUSImVll9oW1Ucxz/n5OYmN8OkW7WZss5VKGXCirqiW2mttUzqkxYcU6RTYQ+VISo+ijpfJ8j0qQVhWFlBdBNkYFG6dWvWKROhVey/p7VoTf3TJNrc5ObmHB9uktY2t6nft3v5nd/3nN/ve76/I7TWGh8k04rEnMv0UpHFPxUZ2wuNWoL99ZJDjQE6Wgz2xqRfCkQ1gpWMYnjCYXLBRfnSe5ACjjYbnOwwiVch2kIwPuMydCWP7dTIvAnhoGCgJ8RjBw1/gi++d/h4wqmaIBYRPNkaRANfTRUq5dqMFzpN+trMrQTjMy7nRnNVFwkB7z8fqXxrDW+MZPHr3uu9YbpKJ5HgNXNwLF89Gti9S9B0l+Ts5RxnL+e4r0FSFxG+8YNjeZIZtU4wnHDIFfxrnlrT/LqqOH0sxOljIX5ZVaSz/vF2QfNJqdQymVbcXHB9gwGUhjOXbGKWIBYRnLlk11TX5IJLMq0wJuZqSxFgJaNZySiEEPyeqb1AaUjMucgfl4o1g/ftkbz1dJjDTQYPHQjwdp/Fvj3+l6uM6aUi4sWhf/Tqmv+OeluDnOoO8VtKcf66J4SXHg2xt07y0dU8o9MF37W7dwkMPz2X8cC9AeaWi7zzuY3rCYOp21nefcbiwQOBbQkytqb2OYG/bV1JDuAq799OIKPWVj0b0vMYALnNFioxwluzGTFLYOyvl6yurTe6r83k2aMmTkEzu1ykrcnAVZ4FfPadV47jjwQ53GRgSHjzqTAtdwcIBQUXbuT58of1kjXWS4xDjQGmFj2CO+8Q9HeYnBvNcU+d5MQRkw+/zlEXkRx/OEj3/UEAzACMTDqks5pXnghx8ZbD7T8Ur/aGGZ9xKz7V2hjA6GwxGJl0UBqCAYEAllOKYEBgFzTXZ12KCq7+XODEEc/EPv3WIZXVSAmnuk2Sac3SXwqxqWwdLQZGPCZpbzZIzLsspxTf/FTgveciKA3nr+UplpqbymqGrvzXr5SCCzccXu4JgYCLtzxigPZmg3hMem6azCheG7axS34Uj0kcV7Pd/diIqCWQgkpyyxR80G/REJWeTONRyUBPiLKekmm14+Tg6T21wfwGHg/REPVkVRFX10GDk50m/iZcGwJPbV0bptqWkXlt1mVw7P+PTKs0Mru2G5llrGQUwwmHyfmdDf32ZoP+TpN4dAdDfyOSaUVi3mV6sfqzpbX0bKn2mijjX84XmagmwylBAAAAAElFTkSuQmCC"

// InitTray 初始化系统托盘
func InitTray(cfg *config.Config) {
	if cfg != nil {
		_cfg = *cfg
	}
	go systray.Run(onTrayReady, onTrayExit)
}

// onTrayReady 系统托盘准备就绪时调用
func onTrayReady() {
	logx.Info("系统托盘准备初始化...")

	// 解码图标
	logx.Debug("解码托盘图标...")
	iconBytes, err := base64.StdEncoding.DecodeString(iconBase64)
	if err != nil {
		logx.Error("解码图标失败: %v", err)
		return
	}
	logx.Debug("图标解码成功")

	// 设置托盘图标
	systray.SetIcon(iconBytes)
	// 设置托盘标题
	systray.SetTitle("AiFlow")
	// 设置托盘提示
	systray.SetTooltip("AiFlow 服务运行中")
	logx.Debug("托盘图标和标题设置完成")

	// 创建托盘菜单
	logx.Debug("创建托盘菜单...")
	createTrayMenu()
	logx.Debug("托盘菜单创建完成")

	// 标记服务为运行状态
	serverRunning = true
	logx.Info("系统托盘初始化完成")
}

// createTrayMenu 创建托盘菜单
func createTrayMenu() {
	// 打开Web后台菜单
	webMenu := systray.AddMenuItem("打开Web后台", "打开AiFlow Web后台管理页面")
	go func() {
		for {
			select {
			case <-webMenu.ClickedCh:
				openWebAdmin()
			case <-exitChan:
				return
			}
		}
	}()

	// 打开日志目录菜单
	logMenu := systray.AddMenuItem("打开日志目录", "打开AiFlow日志文件所在目录")
	go func() {
		for {
			select {
			case <-logMenu.ClickedCh:
				openLogDirectory()
			case <-exitChan:
				return
			}
		}
	}()

	// 分隔线
	systray.AddSeparator()

	// 退出菜单
	exitMenu := systray.AddMenuItem("退出", "退出AiFlow服务")
	go func() {
		select {
		case <-exitMenu.ClickedCh:
			exitService()
		case <-exitChan:
			return
		}
	}()
}

// openWebAdmin 打开Web后台管理页面
func openWebAdmin() {
	logx.Debug("准备打开Web后台管理页面...")
	listenAddr := _cfg.Server.Addr
	if listenAddr == "" {
		logx.Error("未配置监听地址，无法打开Web后台")
		return
	}
	webURL := fmt.Sprintf("http://%s%s", listenAddr, _cfg.Server.WebPath)
	logx.Info("打开Web后台: %s", webURL)

	// 使用默认浏览器打开
	var cmd string
	var args []string

	cmd = "cmd"
	args = []string{"/c", "start", webURL}

	logx.Debug("执行命令: %s %v", cmd, args)
	if err := exec.Command(cmd, args...).Start(); err != nil {
		logx.Error("打开浏览器失败: %v", err)
		return
	}
	logx.Debug("浏览器打开命令执行成功")
}

// openLogDirectory 打开日志目录
func openLogDirectory() {
	logx.Debug("准备打开日志目录...")
	logDir := _cfg.Log.FilePath
	if logDir == "" {
		logDir = "./logs"
		logx.Debug("日志目录未配置，使用默认值: %s", logDir)
	}

	// 检查是否为绝对路径，不是则转换为绝对路径
	if !isAbsolutePath(logDir) {
		logx.Debug("日志目录不是绝对路径，转换为绝对路径...")
		absPath, err := getAbsolutePath(logDir)
		if err != nil {
			logx.Error("获取绝对路径失败: %v", err)
			return
		}
		logDir = absPath
		logx.Debug("转换后的绝对路径: %s", logDir)
	}

	logx.Info("打开日志目录: %s", logDir)

	// 使用资源管理器打开目录
	var cmd string
	var args []string

	cmd = "explorer"
	args = []string{logDir}

	logx.Debug("执行命令: %s %v", cmd, args)
	if err := exec.Command(cmd, args...).Start(); err != nil {
		logx.Error("打开日志目录失败: %v", err)
		return
	}
	logx.Debug("资源管理器打开命令执行成功")
}

// isAbsolutePath 检查路径是否为绝对路径
func isAbsolutePath(path string) bool {
	// Windows 绝对路径格式: C:\path 或 \\server\share
	if len(path) >= 2 && path[1] == ':' {
		return true
	}
	if len(path) >= 2 && path[0] == '\\' && path[1] == '\\' {
		return true
	}
	return false
}

// getAbsolutePath 获取路径的绝对路径
func getAbsolutePath(path string) (string, error) {
	// 使用标准库获取绝对路径
	return filepath.Abs(path)
}

// exitService 退出服务
func exitService() {
	logx.Info("退出服务...")
	// 标记服务为停止状态
	serverRunning = false
	// 退出系统托盘
	systray.Quit()
	// 通知主程序退出
	closeExitChan()
}

// onTrayExit 系统托盘退出时调用
func onTrayExit() {
	logx.Info("系统托盘已退出")
	// 标记服务为停止状态
	serverRunning = false
	// 通知主程序退出
	closeExitChan()
}

// closeExitChan 安全关闭退出通道
func closeExitChan() {
	if !exitChanClosed {
		exitChanClosed = true
		close(exitChan)
	}
}

// WaitForExit 等待退出信号
func WaitForExit() {
	<-exitChan
	logx.Info("收到退出信号，服务正在停止...")
}
