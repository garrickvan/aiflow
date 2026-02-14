package utils

import "time"

// 时间格式常量
const (
	// DateTimeLayout 默认日期时间格式
	DateTimeLayout = "2006-01-02 15:04:05"
	// DateLayout 日期格式
	DateLayout = "2006-01-02"
	// TimeLayout 时间格式
	TimeLayout = "15:04:05"
)

// FormatTimestamp 格式化毫秒时间戳为默认格式字符串
// 输出格式: "2006-01-02 15:04:05"
func FormatTimestamp(timestamp int64) string {
	t := time.UnixMilli(timestamp)
	return t.Format(DateTimeLayout)
}

// FormatTimestampWithLayout 格式化毫秒时间戳为指定格式字符串
// layout参数指定输出格式，如 "2006-01-02" 或 "15:04:05"
func FormatTimestampWithLayout(timestamp int64, layout string) string {
	t := time.UnixMilli(timestamp)
	return t.Format(layout)
}

// NowMilli 获取当前毫秒时间戳
// 返回从1970-01-01起的毫秒数
func NowMilli() int64 {
	return time.Now().UnixMilli()
}
