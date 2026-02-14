package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// generateRandomDirName 生成随机目录名
// 格式：4个随机字母 + 时间戳
func GenerateRandomDirName() string {
	// 生成4个随机字母
	letters := "abcdefghijklmnpqrstuvwxyz"
	randomStr := make([]byte, 4)
	for i := 0; i < 4; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			// 如果出错，使用默认值
			randomStr[i] = 'a'
			continue
		}
		randomStr[i] = letters[num.Int64()]
	}

	// 添加时间戳
	timestamp := time.Now().UnixMilli()

	return fmt.Sprintf("%s_%d", randomStr, timestamp)
}
