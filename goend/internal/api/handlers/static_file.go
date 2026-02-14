package handlers

import (
	"io/fs"
	"net/http"
)

// 静态文件系统
var staticFS fs.FS

// SetStaticFS 设置静态文件系统
func SetStaticFS(fs fs.FS) {
	staticFS = fs
}

// WebHandler 处理Web请求，返回静态资源目录下的前端资源
var WebHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// 当访问以 /web 开头的路径时，返回静态资源
	// 所有的静态资源都从这里返回
	if len(path) >= 4 && path[:4] == "/web" {

		// 检查静态文件系统是否已设置
		if staticFS == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("静态文件系统未初始化"))
			return
		}

		// 处理请求，将 /web 路径映射到静态资源目录
		// 移除 /web 前缀，将请求传递给文件服务器
		// 对于 /web 路径，映射到根目录
		if path == "/web" || path == "/web/index.html" {
			r.URL.Path = "/"
		} else {
			// 对于其他 /web/xxx 路径，移除 /web 前缀
			r.URL.Path = path[4:]
			// 处理 /static 前缀，因为前端构建生成的路径包含 /static
			if len(r.URL.Path) >= 7 && r.URL.Path[:7] == "/static" {
				r.URL.Path = r.URL.Path[7:]
			}
		}

		// 直接读取并返回index.html文件
		if r.URL.Path == "/" {
			data, err := fs.ReadFile(staticFS, "static/dist/index.html")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("读取index.html文件失败: " + err.Error()))
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write(data)
			return
		}

		// 创建文件服务器，使用static/dist目录作为根目录
		distFS, err := fs.Sub(staticFS, "static/dist")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("创建dist文件系统失败: " + err.Error()))
			return
		}

		// 使用文件服务器处理静态文件请求
		http.FileServer(http.FS(distFS)).ServeHTTP(w, r)
		return
	}

	// 其他路径返回默认信息
	w.Write([]byte("Web路由初始化完成: " + path))
})
