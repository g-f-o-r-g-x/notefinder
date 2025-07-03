package background

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"notefinder/internal/notefinder/types"
)

type BusReader interface {
	ReadBus() (*types.Note, bool)
}

type Consumer struct {
	reader BusReader
	mx     sync.Mutex
}

func NewConsumer(reader BusReader) *Consumer {
	return &Consumer{reader: reader}
}

func (c *Consumer) Run() {
	for {
		note, ok := c.reader.ReadBus()
		if !ok {
			return
		}

		fmt.Println(note.Title)
		for k, hits := range words(note) {
			fmt.Printf("\"%s\": %d\n", k, hits)
		}
		fmt.Println("---------------------------------")
	}
}

var (
	/* Below handles a0 byte sequence as well (Unicode for &nbsp;) */
	allWhiteSpace  = regexp.MustCompile(`(?:\s| |&nbsp;)+`)
	allPunctuation = regexp.MustCompile(`^[[:punct:]\p{P}\p{S}“”‘’„‚«»…–—‐‑‑‒−­]+|[[:punct:]\p{P}\p{S}“”‘’„‚«»…–—‐‑‑‒−­]+$`)
	rules          = NewRuleTable(strings.Split(defaultRules, "\n"))
)

func words(item *types.Note) map[string]int {
	ret := make(map[string]int)
	return ret
	for _, desc := range item.Mapping() {
		if !desc.Searchable {
			continue
		}
		value := desc.Ptr.(*string)

		text, err := url.QueryUnescape(*value)
		if err != nil {
			continue
		}
		all := allWhiteSpace.Split(text, -1)
		for _, w := range all {
			cleanWord := strings.ToLower(allPunctuation.ReplaceAllString(w, ""))

			onlyDigits := true
			for _, r := range cleanWord {
				if !unicode.IsDigit(r) {
					onlyDigits = false
					break
				}
			}

			if onlyDigits || len([]rune(cleanWord)) < 3 {
				continue
			}
			ret[rules.Stem(cleanWord)]++
		}
	}

	return ret
}
