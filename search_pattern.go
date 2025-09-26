package stardict

import (
	"regexp"
	"time"

	"github.com/ilius/glob"
	common "github.com/ilius/go-dict-commons"
	su "github.com/ilius/go-dict-commons/search_utils"
)

func (d *dictionaryImp) searchPattern(
	workerCount int,
	timeout time.Duration,
	checkTerm func(string) uint8,
) []*common.SearchResultLow {
	idx := d.idx
	const minScore = uint8(140)

	N := len(idx.entries)
	return su.RunWorkers(
		N,
		workerCount,
		timeout,
		func(start int, end int) []*common.SearchResultLow {
			var results []*common.SearchResultLow
			var entry *IdxEntry
			var score uint8
			var entryI int
			for entryI = start; entryI < end; entryI++ {
				entry = idx.entries[entryI]
				score = uint8(0)
				for _, term := range entry.terms {
					termScore := checkTerm(term)
					if termScore > score {
						score = termScore
						break
					}
				}
				if score < minScore {
					continue
				}
				results = append(results, d.newResult(entry, entryI, score))
			}
			return results
		},
	)
}

func (d *dictionaryImp) SearchRegex(
	query string,
	workerCount int,
	timeout time.Duration,
) ([]*common.SearchResultLow, error) {
	re, err := regexp.Compile("^" + query + "$")
	if err != nil {
		return nil, err
	}
	return d.searchPattern(workerCount, timeout, func(term string) uint8 {
		if !re.MatchString(term) {
			return 0
		}
		if len(term) < 20 {
			return 200 - uint8(len(term))
		}
		return 180
	}), nil
}

func (d *dictionaryImp) SearchGlob(
	query string,
	workerCount int,
	timeout time.Duration,
) ([]*common.SearchResultLow, error) {
	pattern, err := glob.Compile(query)
	if err != nil {
		return nil, err
	}
	return d.searchPattern(workerCount, timeout, func(term string) uint8 {
		if !pattern.Match(term) {
			return 0
		}
		if len(term) < 20 {
			return 200 - uint8(len(term))
		}
		return 180
	}), nil
}
