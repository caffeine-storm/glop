package system

import (
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/glog"
)

type MockSystem struct {
	System
	mockOs *mockOs
}

type mockOs struct {
	Os
	currentTimeMs int64
}

func (mos *mockOs) Startup() int64 {
	mos.Os.Startup()
	mos.currentTimeMs = 42
	return 42
}

func (mos *mockOs) Think() int64 {
	mos.Os.Think()
	return mos.currentTimeMs
}

func (mos *mockOs) GetInputEvents() ([]gin.OsEvent, int64) {
	events, _ := mos.Os.GetInputEvents()
	// rewrite event timestamps to all be 'current time' or else they'll get real
	// timestamps.
	for idx := range events {
		events[idx].TimestampMs = mos.currentTimeMs
	}

	return events, mos.currentTimeMs
}

func makeMockedOs(realOs Os) *mockOs {
	return &mockOs{
		Os: realOs,
	}
}

func MakeMocked(realOs Os) *MockSystem {
	mockOs := makeMockedOs(realOs)
	mockInput := gin.MakeLogged(glog.VoidLogger())
	return &MockSystem{
		System: Make(mockOs, mockInput),
		mockOs: mockOs,
	}
}

func (ms *MockSystem) AdvanceTimeMillis(delta uint64) {
	ms.mockOs.currentTimeMs += int64(delta)
}
