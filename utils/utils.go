package utils

import (
	"context"
	"strings"
	"yokogcache/internal/db/dbservice"
)

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

func InitDB() {
	//dbservice.NewStudentDB(context.Background()).CreateTable()
	name := []string{
		"Ella Robinson", "Alexander Williams", "James Franklin",
	}
	score := []float64{
		67.8, 56.5, 34,
	}

	dbop := dbservice.NewStudentDB(context.Background())
	for _, n := range name {
		for _, s := range score {
			dbop.CreateRecord(n, s)
		}
	}
}
