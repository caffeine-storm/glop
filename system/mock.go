package system

import (
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/glog"
)

type MockSystem interface {
	System
	AdvanceTime(delta uint64)
}

type MockOs interface {
	Os
}

type mockSystem struct {
	System
	mockOs *mockOs
}

type mockOs struct {
	Os
	currentTime int64
}

func (mos *mockOs) Startup() int64 {
	mos.Os.Startup()
	mos.currentTime = 42
	return 42
}

func (mos *mockOs) Think() int64 {
	mos.Os.Think()
	return mos.currentTime
}

func (mos *mockOs) GetInputEvents() ([]gin.OsEvent, int64) {
	events, _ := mos.Os.GetInputEvents()
	// rewrite event timestamps to all be 'current time' or else they'll get real
	// timestamps.
	for idx := range events {
		events[idx].Timestamp = mos.currentTime
	}

	return events, mos.currentTime
}

func MakeMockedOs(realOs Os) *mockOs {
	return &mockOs{
		Os: realOs,
	}
}

func MakeMocked(realOs Os) *mockSystem {
	mockOs := MakeMockedOs(realOs)
	mockInput := gin.MakeLogged(glog.VoidLogger())
	return &mockSystem{
		System: Make(mockOs, mockInput),
		mockOs: mockOs,
	}
}

func (ms *mockSystem) AdvanceTime(delta uint64) {
	ms.mockOs.currentTime += int64(delta)
}

var _ MockSystem = (*mockSystem)(nil)
