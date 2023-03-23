package stardict

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// Sense has information belonging to single item position in dictionary
type Sense = [2]uint64

// Idx implements an in-memory index for a dictionary
type Idx struct {
	items map[string][]Sense

	db *sql.DB
}

// NewIdx initializes idx struct
func NewIdx() *Idx {
	idx := new(Idx)
	idx.items = make(map[string][]Sense)
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("CREATE TABLE idx (keyword TEXT, offset INTEGER, size INTEGER);")
	if err != nil {
		panic(err)
	}
	idx.db = db
	return idx
}

// Add adds an item to in-memory index
func (idx *Idx) Add(item string, offset uint64, size uint64) {
	_, err := idx.db.Exec(
		"insert into idx (keyword, offset, size) values (?, ?, ?)",
		item, offset, size,
	)
	if err != nil {
		fmt.Println(err)
	}
}

// Get gets all translations for an item
func (idx Idx) Get(item string) []Sense {
	rows, err := idx.db.Query("select offset, size from idx where keyword = ?", item)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()
	senses := []Sense{}
	for rows.Next() {
		var offset int
		var size int
		err = rows.Scan(&offset, &size)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		senses = append(senses, Sense{uint64(offset), uint64(size)})
	}
	err = rows.Err()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return senses
}

type IdxSearchResult struct {
	Keyword string
	Offset  uint64
	Size    uint64
}

func (idx Idx) search(cond string, arg string) []*IdxSearchResult {
	rows, err := idx.db.Query(
		"select keyword, offset, size from idx where "+cond,
		arg,
	)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()
	results := []*IdxSearchResult{}
	for rows.Next() {
		var keyword string
		var offset int
		var size int
		err = rows.Scan(&keyword, &offset, &size)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		results = append(results, &IdxSearchResult{
			Keyword: keyword,
			Offset:  uint64(offset),
			Size:    uint64(size),
		})
	}
	err = rows.Err()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return results
}

// ReadIndex reads dictionary index into a memory and returns in-memory index structure
func ReadIndex(filename string, info *Info) (idx *Idx, err error) {
	data, err := ioutil.ReadFile(filename)
	// unable to read index
	if err != nil {
		return
	}

	idx = NewIdx()

	var a [255]byte // temporary buffer
	var aIdx int
	var expect int

	var dataStr string
	var dataOffset uint64
	var dataSize uint64

	var maxIntBytes int

	if info.Is64 == true {
		maxIntBytes = 8
	} else {
		maxIntBytes = 4
	}

	for _, b := range data {
		if expect == 0 {
			a[aIdx] = b
			if b == 0 {
				dataStr = string(a[:aIdx])

				aIdx = 0
				expect++
				continue
			}
			aIdx++
		} else {
			if expect == 1 {
				a[aIdx] = b
				if aIdx == maxIntBytes-1 {
					if info.Is64 {
						dataOffset = binary.BigEndian.Uint64(a[:maxIntBytes])
					} else {
						dataOffset = uint64(binary.BigEndian.Uint32(a[:maxIntBytes]))
					}

					aIdx = 0
					expect++
					continue
				}
				aIdx++
			} else {
				a[aIdx] = b
				if aIdx == maxIntBytes-1 {
					if info.Is64 {
						dataSize = binary.BigEndian.Uint64(a[:maxIntBytes])
					} else {
						dataSize = uint64(binary.BigEndian.Uint32(a[:maxIntBytes]))
					}

					aIdx = 0
					expect = 0

					// finished with one record
					idx.Add(dataStr, dataOffset, dataSize)

					continue
				}
				aIdx++
			}
		}
	}

	_, err = idx.db.Exec("CREATE INDEX keyword_idx ON idx(keyword);")
	if err != nil {
		return nil, err
	}

	return idx, err
}
