package scheduler

import (
	"math"
	"time"
)

type State struct {
	EaseFactor   float64
	IntervalDays int
	Repetitions  int
}

func NewState() State {
	return State{EaseFactor: 2.5, IntervalDays: 0, Repetitions: 0}
}

func Next(s State, quality int, now time.Time) (State, time.Time) {
	if quality < 3 {

		s.Repetitions = 0
		s.IntervalDays = 1
	} else {
		switch s.Repetitions {
		case 0:
			s.IntervalDays = 1
		case 1:
			s.IntervalDays = 6
		default:
			s.IntervalDays = int(math.Round(float64(s.IntervalDays) * s.EaseFactor))
		}
		s.Repetitions++

		q := float64(quality)
		s.EaseFactor = s.EaseFactor + (0.1 - (5-q)*(0.08+(5-q)*0.02))
		if s.EaseFactor < 1.3 {
			s.EaseFactor = 1.3
		}
	}

	dueAt := now.Add(time.Duration(s.IntervalDays) * 24 * time.Hour)
	return s, dueAt
}
