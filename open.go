package stardict

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	common "github.com/ilius/go-dict-commons"
)

const ifoExt = ".ifo"

// Open open directories
func Open(dirPathList []string, order map[string]int) ([]common.Dictionary, error) {
	var dicList []common.Dictionary

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	for _, dirPath := range dirPathList {
		// dirPath = pathFromUnix(dirPath) // not needed for relative paths
		if !filepath.IsAbs(dirPath) {
			dirPath = filepath.Join(homeDir, dirPath)
		}

		dirEntries, err := os.ReadDir(dirPath)
		if err != nil {
			ErrorHandler(err)
			continue
		}
		for _, fi := range dirEntries {
			dic, err := checkDirEntry(dirPath, fi)
			if err != nil {
				ErrorHandler(err)
				continue
			}
			if dic == nil {
				continue
			}
			if order[dic.DictName()] < 0 {
				dic.disabled = true
				dicList = append(dicList, dic)
				continue
			}
			dicList = append(dicList, dic)
		}
	}
	slog.Info("Starting to load indexes")
	var wg sync.WaitGroup
	load := func(dic common.Dictionary) {
		defer wg.Done()
		t0 := time.Now()
		err = dic.Load()
		if err != nil {
			ErrorHandler(fmt.Errorf("error loading %#v: %w", dic.DictName(), err))
		} else {
			slog.Info("Loaded index", "path", dic.IndexPath(), "dt", time.Since(t0))
		}
	}
	for _, dic := range dicList {
		if dic.Disabled() {
			continue
		}
		wg.Add(1)
		go load(dic)
	}
	wg.Wait()
	return dicList, nil
}

type DirEntryFromFileInfo struct {
	fs.FileInfo
}

func (e *DirEntryFromFileInfo) Type() fs.FileMode {
	return e.Mode().Type()
}

func (e *DirEntryFromFileInfo) Info() (fs.FileInfo, error) {
	return e.FileInfo, nil
}

func checkDirEntry(parentDir string, entry os.DirEntry) (*dictionaryImp, error) {
	path := filepath.Join(parentDir, entry.Name())
	dictDir := parentDir
	if entry.IsDir() {
		_, ifoFi, err := findIfoFile(path)
		if err != nil {
			return nil, err
		}
		if ifoFi == nil {
			return nil, nil
		}
		entry = &DirEntryFromFileInfo{FileInfo: ifoFi}
		dictDir = path
	}
	name := entry.Name()
	if filepath.Ext(name) != ifoExt {
		return nil, nil
	}
	slog.Info("Initializing dictionary", "directory", dictDir)
	dic, err := NewDictionary(
		dictDir,
		name[:len(name)-len(ifoExt)],
	)
	if err != nil {
		return nil, err
	}
	resDir := filepath.Join(dictDir, "res")
	if isDir(resDir) {
		dic.resDir = resDir
		dic.resURL = "file://" + pathToUnix(resDir)
	}
	return dic, nil
}

func isDir(pathStr string) bool {
	stat, _ := os.Stat(pathStr)
	if stat == nil {
		return false
	}
	return stat.IsDir()
}

func findIfoFile(path string) (string, os.FileInfo, error) {
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return "", nil, err
	}
	for _, de := range dirEntries {
		if filepath.Ext(de.Name()) != ifoExt {
			continue
		}
		fi, err := de.Info()
		if err != nil {
			return "", nil, err
		}
		if fi == nil {
			return "", nil, nil
		}
		return filepath.Join(path, fi.Name()), fi, nil
	}
	return "", nil, nil
}

func pathToUnix(pathStr string) string {
	if runtime.GOOS != "windows" {
		return pathStr
	}
	return "/" + strings.ReplaceAll(pathStr, `\`, `/`)
}
