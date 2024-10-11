package utils

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"yokogcache/internal/db/dbservice"
)

func TestFind(t *testing.T) {
	dbop := dbservice.NewStudentDB(context.Background())

	var val float64
	if err := dbop.Find(&val, "name=?", "Ella Robinson"); err == nil {
		tmp := strconv.Itoa(int(val))
		fmt.Println(tmp)
	} else {
		fmt.Println(err)
	}
}
