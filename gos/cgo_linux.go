package gos

import (
	"github.com/runningwild/glop/gos/linux"
	"github.com/runningwild/glop/system"
)

type linuxSystemObject struct {
	system.Os
}

var _ system.Os = (*linuxSystemObject)(nil)

func NewSystemInterface() *linuxSystemObject {
	return &linuxSystemObject{
		Os: linux.New(),
	}
}
