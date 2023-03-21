package stardict

import (
	"fmt"
	"os"
	"path/filepath"
)

// Open open directories
func Open(d string) ([]*Dictionary, error) {
	var items []*Dictionary
	const ext = ".ifo"
	filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		name := info.Name()
		if filepath.Ext(info.Name()) != ext {
			return nil
		}
		fmt.Printf("Loading %#v\n", path)
		dir, err := NewDictionary(filepath.Dir(path), name[:len(name)-len(ext)])
		if err != nil {
			return err
		}
		items = append(items, dir)
		return nil
	})
	return items, nil
}
