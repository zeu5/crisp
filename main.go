package crisp

import "testing"

var (
	rtEnabled = false
	rt        *runtime
)

func Initialize(t *testing.T) {
	rt = NewRuntime(t)
	rtEnabled = true
}

func Go(f func()) {
	if !rtEnabled || rt.getState() == stopped {
		go f()
		return
	}
	r := newRoutine(f)
	rt.addRoutine(r)
	// TODO: Pessimistic yield. If the current holder is the same as the current routine, then we can just return.
	rt.yield()
}

func Assert(b bool, msg string) {
	if !b {
		rt.Panic("assertion failed: " + msg)
	}
}

func CheckRandom(f func(), t *testing.T, iterations int) {
	rtEnabled = true
	rt = NewRuntime(t)
	rtConfig := &RuntimeConfig{
		MaxSteps:         100,
		RootFunc:         f,
		Scheduler:        NewRandomScheduler(),
		ScheduleSavePath: ".scheduler",
		MaxIterations:    iterations,
	}
	rt.SetConfig(rtConfig)
	rt.Run()
}
