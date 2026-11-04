package chunker

import (
	"slices"
	"testing"
	"unicode/utf8"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxChars int
		want     []string
	}{
		{"short text is one chunk", "Hello world.", 800, []string{"Hello world."}},
		{"small paragraphs are merged", "A.\n\nB.", 800, []string{"A.\n\nB."}},
		{"paragraphs split when they do not fit", "aaaa aaaa\n\nbbbb bbbb", 10, []string{"aaaa aaaa", "bbbb bbbb"}},
		{"long paragraph split at words", "one two three four five", 10, []string{"one two", "three four", "five"}},
		{"huge word is cut", "abcdefghij", 4, []string{"abcd", "efgh", "ij"}},
		{"empty text", "", 800, []string{}},
		{"only spaces", "   \n\n  ", 800, []string{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Split(tc.text, tc.maxChars)

			if !slices.Equal(got, tc.want) {
				t.Errorf("got %q, want %q", got, tc.want)
			}
			for _, chunk := range got {
				if utf8.RuneCountInString(chunk) > tc.maxChars {
					t.Errorf("chunk %q is longer than %d", chunk, tc.maxChars)
				}
			}
		})
	}
}
