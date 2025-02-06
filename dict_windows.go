//go:build windows
// +build windows

package stardict

import (
	"log/slog"
)

// GetSequence returns data at the given offset
func (d *Dict) GetSequence(offset uint64, size uint64) []byte {
	if d.file == nil {
		slog.Warn("GetSequence: file is closed")
		return nil
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	p := make([]byte, size)
	_, err := d.file.ReadAt(p, int64(offset))
	if err != nil {
		slog.Error("error while reading dict file", "err", err, "filename", d.filename)
		return nil
	}
	return p
}
