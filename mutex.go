package crisp

import "sync"

type Mutex struct {
	c *cMutex
	m sync.Mutex
}

func (m *Mutex) Lock() {
	if rtEnabled {
		m.m.Lock()
	} else {
		if m.c == nil {
			m.c = newCMutex()
		}
		m.c.Lock()
	}
}

func (m *Mutex) Unlock() {
	if rtEnabled {
		m.m.Unlock()
	} else {
		m.c.Unlock()
	}
}

// TODO: Implement this
type cMutex struct {
	waitingRoutines map[int64]bool
	currentHolder   int64
}

func (m *cMutex) Lock() {
	me := rt.me()
	if m.currentHolder == me.id {
		rt.Panic("deadlock! acquiring mutex held by current routine")
	}
	if m.currentHolder != -1 {
		m.waitingRoutines[me.id] = true
		me.block()
	} else {
		m.currentHolder = me.id
	}
	// TODO: Pessimistic yield. If the current holder is the same as the current routine, then we can just return.
	rt.yield()
}

func (m *cMutex) Unlock() {
	// TODO: unblock all, reset current holder maybe yield?
}

func newCMutex() *cMutex {
	return &cMutex{
		waitingRoutines: make(map[int64]bool),
		currentHolder:   -1,
	}
}
