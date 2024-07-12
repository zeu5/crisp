package crisp

import (
	"encoding/json"
	"path"
	"strconv"
)

type Schedule struct {
	seed  int64
	steps []ScheduleStep
}

type ScheduleStep struct {
	routine int64
	err     error
	done    bool
}

func (s *ScheduleStep) IsRoutine() bool {
	return s.routine != 0
}

func (s *ScheduleStep) IsError() bool {
	return s.err != nil
}

func (s *ScheduleStep) IsDone() bool {
	return s.done
}

func (s *ScheduleStep) GetRoutine() int64 {
	return s.routine
}

func (s *ScheduleStep) GetError() error {
	return s.err
}

func NewSchedule(seed int64) *Schedule {
	return &Schedule{
		seed: seed,
	}
}

func (s *Schedule) AddRoutineStep(routine int64) {
	s.steps = append(s.steps, ScheduleStep{
		routine: routine,
	})
}

func (s *Schedule) AddErrorStep(err error) {
	s.steps = append(s.steps, ScheduleStep{
		err: err,
	})
}

func (s *Schedule) AddDoneStep() {
	s.steps = append(s.steps, ScheduleStep{
		done: true,
	})
}

func (s *Schedule) Save(savePath string) error {
	savePath = path.Join(savePath, strconv.Itoa(int(s.seed))+".json")
	out := make(map[string]interface{})
	out["seed"] = s.seed
	steps := make([]map[string]interface{}, len(s.steps))
	for i, step := range s.steps {
		steps[i] = make(map[string]interface{})
		if step.IsRoutine() {
			steps[i]["routine"] = step.GetRoutine()
		} else if step.IsError() {
			steps[i]["error"] = step.GetError().Error()
		} else if step.IsDone() {
			steps[i]["done"] = true
		}
	}
	out["steps"] = steps
	bs, err := json.Marshal(out)
	if err != nil {
		return err
	}
	return writeToFile(savePath, bs)
}
