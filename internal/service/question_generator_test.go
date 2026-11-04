package service

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseQuestions(t *testing.T) {
	var many strings.Builder
	many.WriteString(`{"questions":[`)
	for i := 0; i < 10; i++ {
		if i > 0 {
			many.WriteString(",")
		}
		fmt.Fprintf(&many, `{"question":"Question %d?","answer":"Answer %d"}`, i, i)
	}
	many.WriteString(`]}`)

	tests := []struct {
		name      string
		reply     string
		wantCount int
		wantErr   bool
	}{
		{"valid", `{"questions":[{"question":"What is X?","answer":"Y."}]}`, 1, false},
		{"wrapped in text", "Here you go:\n```json\n{\"questions\":[{\"question\":\"Q1?\",\"answer\":\"A1\"}]}\n```", 1, false},
		{"skips empty answer", `{"questions":[{"question":"Q1?","answer":""},{"question":"Q2?","answer":"A2"}]}`, 1, false},
		{"removes duplicates", `{"questions":[{"question":"Same?","answer":"A"},{"question":"same?","answer":"B"}]}`, 1, false},
		{"caps at 8", many.String(), 8, false},
		{"not JSON", "sorry, I cannot do that", 0, true},
		{"empty list", `{"questions":[]}`, 0, true},
		{"wrong shape", `{"items":[1]}`, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseQuestions(tc.reply)

			if (err != nil) != tc.wantErr {
				t.Fatalf("error: got %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr && len(got) != tc.wantCount {
				t.Errorf("count: got %d, want %d", len(got), tc.wantCount)
			}
		})
	}
}
