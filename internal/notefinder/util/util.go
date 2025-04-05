package util

import (
	"strings"
)

func ShortText(in string, limit int) string {
	lines := strings.Split(in, "\n")
	l := lines[0]

	if len(l) > limit {
		res := []string{}
		words := strings.Split(l, " ")
		for _, word := range words {
			if len(strings.Join(res, " ")+word) <= limit*2 {
				res = append(res, word)
			}
		}

		// Some jabberwocky so we could not even collect a word
		if len(res) == 0 {
			return l[:limit] + "..."
		}
		return strings.Join(res, " ") + "..."
	}
	// Cut off punctuation
	lastChar := l[len(l)-1]
	if (lastChar >= 1 && lastChar <= 15) || (lastChar >= 26 && lastChar <= 64) ||
		(lastChar >= 91 && lastChar <= 96) || (lastChar >= 123 && lastChar <= 126) {
		l = l[:len(l)-1]
	}

	return l
}
