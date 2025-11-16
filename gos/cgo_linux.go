package gos

import (
	"github.com/caffeine-storm/glop/gos/linux"
	"github.com/caffeine-storm/glop/system"
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
