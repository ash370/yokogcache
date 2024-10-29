package persistent

import (
	"container/list"
	"os"
	"yokogcache/utils/logger"
)

type SnapShot struct {
	file *os.File
}

func NewSnapshot(filename string) *SnapShot {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, os.ModeDevice)
	if err != nil {
		logger.LogrusObj.Errorln("Init snapshot failed, err:", err)
		return nil
	}
	return &SnapShot{file: file}
}

func (s *SnapShot) BgSave(data map[string]*list.Element) error {
	return nil
}
