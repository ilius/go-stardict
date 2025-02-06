package stardict

import (
	"io"
	"log/slog"
)

func closeCloser(c io.Closer) {
	err := c.Close()
	if err != nil {
		slog.Error("error in Close", "err", err)
	}
}

func splitRunes(s []rune, c rune) [][]rune {
	if len(s) == 0 {
		return [][]rune{nil}
	}
	res := [][]rune{}
	var buf []rune
	lastPos := 0
	for i, x := range s {
		if x == c {
			res = append(res, buf)
			buf = nil
			lastPos = i + 1
			continue
		}
		buf = s[lastPos : i+1]
	}
	if buf != nil {
		res = append(res, buf)
	}
	return res
}
