package gos

import (
	"github.com/runningwild/glop/gos/linux"
	"github.com/runningwild/glop/system"
)

type linuxSystemObject struct {
	linux.SystemObject
}

var (
	_ system.Os = (*linuxSystemObject)(nil)
)

func GetSystemInterface() *linuxSystemObject {
	return &linuxSystemObject{
		SystemObject: linux.SystemObject{},
	}
}
