package sync

import "time"

const stackSize = 5

type State struct {
	Stack []Outcome
}

func NewState() *State {
	return &State{
		Stack: []Outcome{},
	}
}

type Outcome struct {
	Timestamp time.Time
	Success   bool
}

func NewOutcome(success bool) *Outcome {
	return &Outcome{
		Timestamp: time.Now(),
		Success:   success,
	}
}

func (s *State) Add(outcome Outcome) {
	s.Stack = append([]Outcome{outcome}, s.Stack...)

	if len(s.Stack) > stackSize {
		s.Stack = s.Stack[:stackSize]
	}
}

func (s *State) OnSuccess() {
	s.Add(*NewOutcome(true))
}

func (s *State) OnFailure(err error) {
	s.Add(*NewOutcome(false))
}
