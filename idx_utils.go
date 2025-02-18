package stardict

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type WordPrefixMap map[rune]map[int]struct{}

func (wpm WordPrefixMap) Add(term string, termIndex int) {
	for _, word := range strings.Split(strings.ToLower(term), " ") {
		if word == "" {
			continue
		}
		prefix, _ := utf8.DecodeRuneInString(word)
		if prefix == utf8.RuneError {
			ErrorHandler(fmt.Errorf(
				"RuneError from DecodeRuneInString for word %#v in term %#v",
				word,
				term,
			))
			continue
		}
		m, ok := wpm[prefix]
		if !ok {
			m = map[int]struct{}{}
			wpm[prefix] = m
		}
		m[termIndex] = struct{}{}
	}
}
