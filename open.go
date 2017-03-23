package stardict

import (
	"fmt"
	"os"
	"path/filepath"
)

// Open open directories
func Open(d string) ([]*Dictionary, error) {
	var items []*Dictionary
	filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := ".ifo"
		name := info.Name()
		if filepath.Ext(info.Name()) == ext {
			fmt.Println(filepath.Dir(path), name[:len(name)-len(ext)])
			dir, err := NewDictionary(filepath.Dir(path), name[:len(name)-len(ext)])
			if err != nil {
				return err
			}
			items = append(items, dir)
		}
		return nil
	})
	return items, nil
}
