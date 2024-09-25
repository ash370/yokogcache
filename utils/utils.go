package utils

import "strings"

// 检查是否满足x.x.x.x:port格式
func ValidPeerAddr(addr string) bool {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return false
	}

	token := strings.Split(parts[0], ".")
	if parts[0] != "localhost" && len(token) != 4 {
		return false
	}
	return true
}
