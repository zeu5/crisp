package crisp

import (
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
)

type runtimeState int

const (
	running runtimeState = iota
	notStarted
	stopped
)

type runtime struct {
	root       *routine
	scheduler  Scheduler
	maxSteps   int
	randSource *rand.Rand
	config     *RuntimeConfig
	t          *testing.T
	state      runtimeState

	routines   map[int64]*routine
	curRoutine *routine
	nextStep   *ScheduleStep
	stepsTaken int
	lock       *sync.Mutex
}

type RuntimeConfig struct {
	MaxSteps         int
	Scheduler        Scheduler
	RootFunc         func()
	ScheduleSavePath string
	Rand             *rand.Rand
	MaxIterations    int
}

func NewRuntime(t *testing.T) *runtime {
	return &runtime{
		routines:   make(map[int64]*routine),
		t:          t,
		state:      notStarted,
		curRoutine: nil,
		nextStep:   nil,
		stepsTaken: 0,
		lock:       new(sync.Mutex),
	}
}

func (r *runtime) Panic(msg string) {
	// TODO: save the schedule
	r.t.Fatal(msg)
}

func (r *runtime) SetConfig(c *RuntimeConfig) {
	r.root = newRoutine(c.RootFunc)
	r.scheduler = c.Scheduler
	r.maxSteps = c.MaxSteps
	r.config = c
	if c.Rand != nil {
		r.randSource = c.Rand
	} else {
		r.randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	if c.ScheduleSavePath != "" {
		_, err := os.Stat(c.ScheduleSavePath)
		if err == nil {
			os.RemoveAll(c.ScheduleSavePath)
		}
	}
}

func (r *runtime) reset() {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.routines = make(map[int64]*routine)
	r.curRoutine = nil
	r.nextStep = nil
	r.stepsTaken = 0
}

func (r *runtime) addRoutine(rt *routine) {
	r.lock.Lock()
	rid := int64(len(r.routines))
	rt.id = rid
	r.routines[rid] = rt
	r.lock.Unlock()

	go rt.run()
}

func (r *runtime) me() *routine {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.curRoutine
}

// Should be called with the lock held. Not thread-safe.
func (r *runtime) availableRoutines() []*routine {
	var routines []*routine
	for _, routine := range r.routines {
		if routine.state == available {
			routines = append(routines, routine)
		}
	}
	return routines
}

func (r *runtime) scheduleNext() {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.nextStep != nil {
		// Already scheduled
		return
	}

	r.nextStep = r.scheduler.Next(r.availableRoutines())
	r.stepsTaken++
}

func (r *runtime) yield() {
	r.lock.Lock()
	state := r.state
	routine := r.curRoutine
	r.lock.Unlock()
	if state == stopped {
		return
	}
	r.t.Logf("Yielding routine %d", routine.id)
	routine.yield()
}

func (r *runtime) getState() runtimeState {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.state
}

func (r *runtime) RunIteration() {
	schedule := NewSchedule(r.randSource.Int63())
	r.state = running
	r.addRoutine(r.root)
	r.scheduler.Reset(schedule.seed)
	r.lock.Lock()
	r.curRoutine = r.root
	r.lock.Unlock()

	// TODO: main loop
ROUTINE_LOOP:
	for {
		r.lock.Lock()
		curRoutine := r.curRoutine
		if r.stepsTaken >= r.maxSteps {
			r.lock.Unlock()
			r.t.Logf("Max steps reached! Stopping")
			break ROUTINE_LOOP
		}
		r.lock.Unlock()
		r.scheduleNext()

		// Resume only if available, otherwise move on
		// We need to do this check because we optimistically decided the next routine to run
		// if it is the same as the current routine and in the previous step, it ran to completion
		// then we cannot resume it again.
		if curRoutine.state == available {
			r.t.Logf("Resuming routine %d", curRoutine.id)
			curRoutine.resume()
		}

		schedule.AddRoutineStep(curRoutine.id)
		r.lock.Lock()
		nextStep := r.nextStep
		r.nextStep = nil
		root := r.root
		r.lock.Unlock()

		if nextStep.IsDone() || root.done() {
			break ROUTINE_LOOP
		} else if nextStep.IsError() {
			r.t.Fatalf("Error running routine: %v", nextStep.GetError())
		} else {
			r.lock.Lock()
			r.curRoutine = r.routines[nextStep.GetRoutine()]
			r.lock.Unlock()
		}
	}

	// Finish all routines
	r.cleanup()

	// Save the schedule
	err := schedule.Save(r.config.ScheduleSavePath)
	if err != nil {
		r.t.Fatalf("Error saving schedule: %v", err)
	}
}

func (r *runtime) cleanup() {
	r.t.Logf("Cleaning up routines!")
	r.lock.Lock()
	r.state = stopped
	r.lock.Unlock()

	// Resume all available routines
	r.lock.Lock()
	for _, routine := range r.routines {
		if routine.state == available {
			routine.resumePermanently()
		}
	}
	r.lock.Unlock()

ROUTINE_CLEANUP_LOOP:
	for {
		allDone := true
		r.lock.Lock()
	ROUTINE_STATE_CHECK_LOOP:
		for _, routine := range r.routines {
			if routine.state != finished {
				allDone = false
				break ROUTINE_STATE_CHECK_LOOP
			}
		}
		r.lock.Unlock()
		if allDone {
			break ROUTINE_CLEANUP_LOOP
		}
		time.Sleep(1 * time.Millisecond)
	}
	r.reset()
}

func (r *runtime) Run() {
	for i := 0; i < r.config.MaxIterations; i++ {
		r.t.Logf("Running iteration %d", i)
		r.RunIteration()
	}
}
