package stardict

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"

	common "codeberg.org/ilius/go-dict-commons"
	"github.com/ilius/go-stardict/v2/murmur3"
)

// dictionaryImp stardict dictionary
type dictionaryImp struct {
	*Info

	dict     *Dict
	idx      *Idx
	ifoPath  string
	idxPath  string
	dictPath string
	synPath  string
	resDir   string
	resURL   string

	decodeData func(data []byte) []*common.SearchResultItem
}

func (d *dictionaryImp) Disabled() bool {
	return d.disabled
}

func (d *dictionaryImp) Loaded() bool {
	return d.dict != nil
}

func (d *dictionaryImp) SetDisabled(disabled bool) {
	d.disabled = disabled
}

func (d *dictionaryImp) ResourceDir() string {
	return d.resDir
}

func (d *dictionaryImp) ResourceURL() string {
	return d.resURL
}

func (d *dictionaryImp) IndexPath() string {
	return d.idxPath
}

func (d *dictionaryImp) InfoPath() string {
	return d.ifoPath
}

func (d *dictionaryImp) Close() {
	d.dict.Close()
}

func (d *dictionaryImp) CalcHash() ([]byte, error) {
	file, err := os.Open(d.idxPath)
	if err != nil {
		return nil, err
	}
	defer closeCloser(file)
	hash := murmur3.New128()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

func (d *dictionaryImp) newResult(entry *IdxEntry, entryIndex int, score uint8) *common.SearchResultLow {
	return &common.SearchResultLow{
		F_Score: score,
		F_Terms: entry.terms,
		Items: func() []*common.SearchResultItem {
			return d.decodeData(d.dict.GetSequence(entry.offset, entry.size))
		},
		F_EntryIndex: uint64(entryIndex),
	}
}

func (d *dictionaryImp) EntryByIndex(index int) *common.SearchResultLow {
	if index >= len(d.idx.entries) {
		return nil
	}
	entry := d.idx.entries[index]
	return d.newResult(entry, index, 0)
}

func (d *dictionaryImp) decodeWithSametypesequence(data []byte) (items []*common.SearchResultItem) {
	seq := d.Options[I_sametypesequence]

	seqLen := len(seq)

	var dataPos int
	dataSize := len(data)

	for i, t := range seq {
		switch t {
		case 'm', 'l', 'g', 't', 'x', 'y', 'k', 'w', 'h', 'r':
			// if last seq item
			if i == seqLen-1 {
				items = append(items, &common.SearchResultItem{Type: t, Data: data[dataPos:dataSize]})
			} else {
				end := bytes.IndexRune(data[dataPos:], '\000')
				items = append(items, &common.SearchResultItem{Type: t, Data: data[dataPos : dataPos+end+1]})
				dataPos += end + 1
			}
		case 'W', 'P':
			if i == seqLen-1 {
				items = append(items, &common.SearchResultItem{Type: t, Data: data[dataPos:dataSize]})
			} else {
				size := binary.BigEndian.Uint32(data[dataPos : dataPos+4])
				items = append(items, &common.SearchResultItem{Type: t, Data: data[dataPos+4 : dataPos+int(size)+5]})
				dataPos += int(size) + 5
			}
		}
	}

	return
}

func (d *dictionaryImp) decodeWithoutSametypesequence(data []byte) (items []*common.SearchResultItem) {
	var dataPos int
	dataSize := len(data)

	for {
		t := data[dataPos]

		dataPos++

		switch t {
		case 'm', 'l', 'g', 't', 'x', 'y', 'k', 'w', 'h', 'r':
			end := bytes.IndexRune(data[dataPos:], '\000')

			if end < 0 { // last item
				items = append(items, &common.SearchResultItem{Type: rune(t), Data: data[dataPos:dataSize]})
				dataPos = dataSize
			} else {
				items = append(items, &common.SearchResultItem{Type: rune(t), Data: data[dataPos : dataPos+end+1]})
				dataPos += end + 1
			}
		case 'W', 'P':
			size := binary.BigEndian.Uint32(data[dataPos : dataPos+4])
			items = append(items, &common.SearchResultItem{Type: rune(t), Data: data[dataPos+4 : dataPos+int(size)+5]})
			dataPos += int(size) + 5
		}

		if dataPos >= dataSize {
			break
		}
	}

	return
}

// DictName returns book name
func (d *dictionaryImp) DictName() string {
	return d.Options[I_bookname]
}

// NewDictionary returns a new Dictionary
// path - path to dictionary files
// name - name of dictionary to parse
func NewDictionary(path string, name string) (*dictionaryImp, error) {
	d := &dictionaryImp{}

	path = filepath.Clean(path)

	ifoPath := filepath.Join(path, name+".ifo")
	idxPath := filepath.Join(path, name+".idx")
	synPath := filepath.Join(path, name+".syn")

	dictDzPath := filepath.Join(path, name+".dict.dz")
	dictPath := filepath.Join(path, name+".dict")

	if _, err := os.Stat(ifoPath); err != nil {
		return nil, err
	}
	if _, err := os.Stat(idxPath); err != nil {
		return nil, err
	}
	if _, err := os.Stat(synPath); err != nil {
		synPath = ""
	}

	// we should have either .dict or .dict.dz file
	if _, err := os.Stat(dictPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if _, errDz := os.Stat(dictDzPath); errDz != nil {
			return nil, err
		}
		dictPath = dictDzPath
	}

	info, err := ReadInfo(ifoPath)
	if err != nil {
		return nil, err
	}
	d.Info = info

	d.ifoPath = ifoPath
	d.idxPath = idxPath
	d.synPath = synPath
	d.dictPath = dictPath

	if _, ok := info.Options[I_sametypesequence]; ok {
		d.decodeData = d.decodeWithSametypesequence
	} else {
		d.decodeData = d.decodeWithoutSametypesequence
	}

	return d, nil
}

func (d *dictionaryImp) Load() error {
	{
		idx, err := ReadIndex(d.idxPath, d.synPath, d.Info)
		if err != nil {
			return err
		}
		d.idx = idx
	}
	{
		dict, err := ReadDict(d.dictPath)
		if err != nil {
			return err
		}
		d.dict = dict
	}
	return nil
}
