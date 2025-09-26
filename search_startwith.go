package stardict

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	common "github.com/ilius/go-dict-commons"
	su "github.com/ilius/go-dict-commons/search_utils"
)

func (d *dictionaryImp) SearchStartWith(
	query string,
	workerCount int,
	timeout time.Duration,
) []*common.SearchResultLow {
	idx := d.idx
	const minScore = uint8(140)

	query = strings.ToLower(strings.TrimSpace(query))

	prefix, _ := utf8.DecodeRuneInString(query)
	if prefix == utf8.RuneError {
		ErrorHandler(fmt.Errorf(
			"RuneError from DecodeRuneInString for query: %#v",
			query,
		))
		return nil
	}
	entryIndexes := idx.byWordPrefix[prefix]
	return su.RunWorkers(
		len(entryIndexes),
		workerCount,
		timeout,
		func(start int, end int) []*common.SearchResultLow {
			var results []*common.SearchResultLow
			var entry *IdxEntry
			var score uint8
			var entryI int
			for entryI = start; entryI < end; entryI++ {
				entry = idx.entries[entryIndexes[entryI]]
				score = su.ScoreStartsWith(entry.terms, query)
				if score < minScore {
					continue
				}
				results = append(results, d.newResult(entry, entryI, score))
			}
			return results
		},
	)
}
