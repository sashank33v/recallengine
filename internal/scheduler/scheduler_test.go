package scheduler

import (
	"math"
	"testing"
	"time"
)

func TestNext(t *testing.T) {
	now := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name         string
		start        State
		quality      int
		wantInterval int
		wantReps     int
		wantEase     float64
	}{
		{"new question, perfect answer", State{2.5, 0, 0}, 5, 1, 1, 2.6},
		{"new question, hard pass", State{2.5, 0, 0}, 3, 1, 1, 2.36},
		{"second success", State{2.5, 1, 1}, 4, 6, 2, 2.5},
		{"third success grows the interval", State{2.5, 6, 2}, 4, 15, 3, 2.5},
		{"failed answer starts again", State{2.5, 15, 3}, 2, 1, 0, 2.5},
		{"ease factor never below 1.3", State{1.3, 6, 2}, 3, 8, 3, 1.3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, dueAt := Next(tc.start, tc.quality, now)

			if got.IntervalDays != tc.wantInterval {
				t.Errorf("interval: got %d, want %d", got.IntervalDays, tc.wantInterval)
			}
			if got.Repetitions != tc.wantReps {
				t.Errorf("repetions: got %d, want %d", got.Repetitions, tc.wantReps)
			}
			if math.Abs(got.EaseFactor-tc.wantEase) > 0.0001 {
				t.Errorf("ease factor: got %.4f, want %.4f", got.EaseFactor, tc.wantEase)
			}
			wantDue := now.Add(time.Duration(tc.wantInterval) * 24 * time.Hour)
			if !dueAt.Equal(wantDue) {
				t.Errorf("due: got %v, want %v", dueAt, wantDue)
			}
		})
	}
}
