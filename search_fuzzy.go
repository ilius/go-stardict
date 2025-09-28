package stardict

import (
	"strings"
	"time"

	common "github.com/ilius/go-dict-commons"
	su "github.com/ilius/go-dict-commons/search_utils"
)

// SearchFuzzy: run a fuzzy search with similarity scores
// ranging from 140 (which means %70) to 200 (which means 100%)
func (d *dictionaryImp) SearchFuzzy(
	query string,
	workerCount int,
	timeout time.Duration,
) []*common.SearchResultLow {
	// if len(query) < 2 {
	// 	return d.searchVeryShort(query)
	// }

	idx := d.idx
	const minScore = uint8(64)

	query = strings.ToLower(strings.TrimSpace(query))
	queryWords := strings.Split(query, " ")
	queryRunes := []rune(query)

	mainWordIndex := 0
	for mainWordIndex < len(queryWords)-1 && queryWords[mainWordIndex] == "*" {
		mainWordIndex++
	}
	queryMainWord := []rune(queryWords[mainWordIndex])

	minWordCount := 1
	queryWordCount := 0

	for _, word := range queryWords {
		if word == "*" {
			minWordCount++
			continue
		}
		queryWordCount++
	}

	prefix := queryMainWord[0]
	entryIndexes := idx.byWordPrefix[prefix]

	args := &su.ScoreFuzzyArgs{
		Query:          query,
		QueryRunes:     queryRunes,
		QueryMainWord:  queryMainWord,
		QueryWordCount: queryWordCount,
		MinWordCount:   minWordCount,
		MainWordIndex:  mainWordIndex,
	}

	return su.RunWorkers(
		len(entryIndexes),
		workerCount,
		timeout,
		func(start int, end int) []*common.SearchResultLow {
			var results []*common.SearchResultLow
			buff := make([]uint16, 500)
			var entry *IdxEntry
			var score uint8
			var entryI int
			for entryI = start; entryI < end; entryI++ {
				entry = idx.entries[entryIndexes[entryI]]
				score = su.ScoreFuzzy(entry.terms, args, buff)
				if score < minScore {
					continue
				}
				results = append(results, d.newResult(entry, entryI, score))
			}
			return results
		},
	)
}
