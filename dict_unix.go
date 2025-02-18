//go:build !windows
// +build !windows

package stardict

import (
	"log/slog"
	"syscall"
)

// GetSequence returns data at the given offset
func (d *Dict) GetSequence(offset uint64, size uint64) []byte {
	if d.file == nil {
		slog.Warn("GetSequence: file is closed")
		return nil
	}
	p := make([]byte, size)
	if d.rawDictFile != nil {
		_, err := syscall.Pread(int(d.rawDictFile.Fd()), p, int64(offset))
		if err != nil {
			slog.Error("error while reading dict file", "err", err, "filename", d.filename)
			return nil
		}
	}
	// we are using .dict.dz reader which uses Seek() and is not concurrent-safe
	d.lock.Lock()
	defer d.lock.Unlock()
	_, err := d.file.ReadAt(p, int64(offset))
	if err != nil {
		slog.Error("error while reading dict file", "err", err, "filename", d.filename)
		return nil
	}
	return p
}
