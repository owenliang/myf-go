package util

import (
	"strconv"
	"strings"
)

// URI中的数字替换
func ReplaceURI(uri string) string {
	// 按/切割
	parts := strings.Split(strings.Trim(uri, "/"), "/")

	// 数字替换为{num}
	replaced := make([]string, 0)
	for i := 0; i < len(parts); i++ {
		if _, err := strconv.Atoi(parts[i]); err == nil { // 数字替换为{num}
			replaced = append(replaced, "{num}")
		} else {
			replaced = append(replaced, parts[i])
		}
	}
	// 用/连接起来
	return strings.Join(replaced, "/")
}
