package utils

import (
	"os"
	"path/filepath"
)

// CreateIfNotExist 检查传递的是文件还是目录，如果是文件则创建其父目录，如果是目录则创建目录
func CreateIfNotExist(path string) error {
	// 获取路径信息
	info, err := os.Stat(path)
	if err == nil {
		// 路径已存在，无需创建
		_ = info
		return nil
	}

	// 如果不是"不存在"的错误，返回错误
	if !os.IsNotExist(err) {
		return err
	}

	// 判断路径是文件还是目录
	// 如果路径包含扩展名，认为是文件；否则认为是目录
	ext := filepath.Ext(path)
	if ext != "" {
		// 是文件，创建其父目录
		dir := filepath.Dir(path)
		return os.MkdirAll(dir, 0755)
	}

	// 是目录，创建目录
	return os.MkdirAll(path, 0755)
}
