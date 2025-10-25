package stardict

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	common "github.com/ilius/go-dict-commons"
	su "github.com/ilius/go-dict-commons/search_utils"
)

func (d *dictionaryImp) SearchExact(
	query string,
	workerCount int,
	timeout time.Duration,
) []*common.SearchResultLow {
	idx := d.idx
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
			var entryI int
			for entryI = start; entryI < end; entryI++ {
				entry = idx.entries[entryIndexes[entryI]]
				for _, term := range entry.terms {
					if strings.ToLower(term) == query {
						results = append(results, d.newResult(entry, entryI, 200))
						break
					}
				}
			}
			return results
		},
	)
}
