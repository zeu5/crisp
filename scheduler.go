package crisp

import "math/rand"

type Scheduler interface {
	Reset(int64)
	Next([]*routine) *ScheduleStep
}

type RandomScheduler struct {
	seed int64
	rand *rand.Rand
}

func NewRandomScheduler() *RandomScheduler {
	return &RandomScheduler{}
}

func (s *RandomScheduler) Reset(seed int64) {
	s.seed = seed
	s.rand = rand.New(rand.NewSource(seed))
}

func (s *RandomScheduler) Next(routines []*routine) *ScheduleStep {
	if len(routines) == 0 {
		return nil
	}
	return &ScheduleStep{
		routine: routines[s.rand.Intn(len(routines))].id,
	}
}
