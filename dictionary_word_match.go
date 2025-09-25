package stardict

import (
	"log/slog"
	"strings"
	"time"

	common "github.com/ilius/go-dict-commons"
	su "github.com/ilius/go-dict-commons/search_utils"
)

func (d *dictionaryImp) SearchWordMatch(
	query string,
	workerCount int,
	timeout time.Duration,
) []*common.SearchResultLow {
	idx := d.idx
	const minScore = uint8(140)

	query = strings.ToLower(strings.TrimSpace(query))

	prefix := []rune(strings.Split(query, " ")[0])[0]
	entryIndexes := idx.byWordPrefix[prefix]

	t1 := time.Now()
	N := len(entryIndexes)

	results := su.RunWorkers(
		N,
		workerCount,
		timeout,
		func(start int, end int) []*common.SearchResultLow {
			var results []*common.SearchResultLow
			var entry *IdxEntry
			var score uint8
			var entryI int
			for entryI = start; entryI < end; entryI++ {
				entry = idx.entries[entryIndexes[entryI]]
				score = su.ScoreWordMatch(entry.terms, query)
				if score < minScore {
					continue
				}
				results = append(results, d.newResult(entry, entryI, score))
			}
			return results
		},
	)

	dt := time.Since(t1)
	if dt > time.Millisecond {
		slog.Debug("SearchWordMatch index loop", "dt", dt, "query", query, "dictName", d.DictName())
	}
	return results
}
