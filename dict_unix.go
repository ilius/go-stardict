//go:build !windows
// +build !windows

package stardict

import (
	"log"
	"syscall"
)

// GetSequence returns data at the given offset
func (d *Dict) GetSequence(offset uint64, size uint64) []byte {
	if d.file == nil {
		log.Println("GetSequence: file is closed")
		return nil
	}
	p := make([]byte, size)
	if d.rawDictFile != nil {
		_, err := syscall.Pread(int(d.rawDictFile.Fd()), p, int64(offset))
		if err != nil {
			log.Printf("error while reading dict file %#v: %v\n", d.filename, err)
			return nil
		}
	}
	// we are using .dict.dz reader which uses Seek() and is not concurrent-safe
	d.lock.Lock()
	defer d.lock.Unlock()
	_, err := d.file.ReadAt(p, int64(offset))
	if err != nil {
		log.Printf("error while reading dict file %#v: %v\n", d.filename, err)
		return nil
	}
	return p
}
