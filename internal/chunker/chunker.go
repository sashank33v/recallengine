package chunker

import (
	"strings"
	"unicode/utf8"
)

func Split(text string, maxChars int) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	paragraphs := strings.Split(text, "\n\n")

	chunks := []string{}
	current := ""

	flush := func() {
		if current != "" {
			chunks = append(chunks, current)
			current = ""
		}
	}

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if utf8.RuneCountInString(p) > maxChars {
			flush()
			chunks = append(chunks, SplitLong(p, maxChars)...)
			continue
		}
		if current != "" && utf8.RuneCountInString(current)+2+utf8.RuneCountInString(p) > maxChars {
			flush()
		}

		if current == "" {
			current = p
		} else {
			current = current + "\n\n" + p
		}
	}
	flush()

	return chunks
}

func SplitLong(p string, maxChars int) []string {
	chunks := []string{}
	current := ""

	for _, word := range strings.Fields(p) {
		for utf8.RuneCountInString(word) > maxChars {
			if current != "" {
				chunks = append(chunks, current)
				current = ""
			}
			runes := []rune(word)
			chunks = append(chunks, string(runes[:maxChars]))
			word = string(runes[maxChars:])
		}

		if current == "" {
			current = word
		} else if utf8.RuneCountInString(current)+1+utf8.RuneCountInString(word) > maxChars {
			chunks = append(chunks, current)
			current = word
		} else {
			current = current + " " + word
		}
	}
	if current != "" {
		chunks = append(chunks, current)
	}

	return chunks
}
