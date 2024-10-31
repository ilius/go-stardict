package stardict

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/ilius/go-stardict/v2/dictzip"
)

type DictFile interface {
	ReadAt(p []byte, off int64) (n int, err error)
	Close() error
}

// Dict implements in-memory dictionary
type Dict struct {
	filename string

	file DictFile
	lock sync.Mutex

	// rawDictFile is only set if we are using .dict, not .dict.dz
	rawDictFile *os.File
}

func (d *Dict) Open() error {
	filename := d.filename
	rawFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	var file DictFile
	if strings.HasSuffix(filename, ".dz") {
		var err error
		file, err = dictzip.NewReader(rawFile)
		if err != nil {
			return err
		}
	} else {
		file = rawFile
		d.rawDictFile = rawFile
	}

	d.file = file
	return nil
}

func (d *Dict) Close() {
	if d.file == nil {
		return
	}
	log.Println("Closing", d.filename)
	closeCloser(d.file)
	d.file = nil
}

// ReadDict creates Dict and opens .dict file
func ReadDict(filename string) (*Dict, error) {
	dict := &Dict{
		filename: filename,
	}
	err := dict.Open()
	if err != nil {
		return nil, err
	}
	return dict, nil
}
