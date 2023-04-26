/*
Package dictzip provides a reader for files in the random access `dictzip` format.

Note: this is not concurrent-safe, since it calls fp.Seek()
*/
package dictzip

import (
	"compress/flate"
	"fmt"
	"io"
	"sync"
)

// Reader, This implements the io.ReadSeekCloser interface.
type Reader struct {
	fp        io.ReadSeekCloser
	offsets   []int64
	blockSize int64
	lock      sync.Mutex
}

func NewReader(rs io.ReadSeekCloser) (*Reader, error) {
	dz := &Reader{fp: rs}

	_, err := dz.fp.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	metadata := []byte{}

	p := 0

	h := make([]byte, 10)
	n, err := io.ReadFull(dz.fp, h)
	if err != nil {
		return nil, err
	}
	p += n

	if h[0] != 31 || h[1] != 139 {
		return nil, fmt.Errorf("invalid header: %02X %02X", h[0], h[1])
	}

	if h[2] != 8 {
		return nil, fmt.Errorf("unknown compression method: %v", h[2])
	}

	flg := h[3]

	if flg&4 != 0 {
		h := make([]byte, 2)
		n, err := io.ReadFull(dz.fp, h)
		if err != nil {
			return nil, err
		}
		p += n

		xLen := int(h[0]) + 256*int(h[1])
		h = make([]byte, xLen)
		n, err = io.ReadFull(dz.fp, h)
		if err != nil {
			return nil, err
		}
		p += n

		for q := 0; q < len(h); {
			si1 := h[q]
			si2 := h[q+1]
			ln := int(h[q+2]) + 256*int(h[q+3])

			if si1 == 'R' && si2 == 'A' {
				metadata = h[q+4 : q+4+ln]
			}

			q += 4 + ln
		}

	}

	// skip file name (8), file comment (16)
	for _, f := range []byte{8, 16} {
		if flg&f != 0 {
			h := make([]byte, 1)
			for {
				n, err := io.ReadFull(dz.fp, h)
				if err != nil {
					return nil, err
				}
				p += n
				if h[0] == 0 {
					break
				}
			}
		}
	}

	if flg&2 != 0 {
		h := make([]byte, 2)
		n, err := io.ReadFull(dz.fp, h)
		if err != nil {
			return nil, err
		}
		p += n
	}

	if len(metadata) < 6 {
		return nil, fmt.Errorf("missing dictzip metadata")
	}

	version := int(metadata[0]) + 256*int(metadata[1])

	if version != 1 {
		return nil, fmt.Errorf("unknown dictzip version: %v", version)
	}

	dz.blockSize = int64(metadata[2]) + 256*int64(metadata[3])
	blockCount := int(metadata[4]) + 256*int(metadata[5])

	dz.offsets = make([]int64, blockCount+1)
	dz.offsets[0] = int64(p)
	for i := 0; i < blockCount; i++ {
		dz.offsets[i+1] = dz.offsets[i] + int64(metadata[6+2*i]) + 256*int64(metadata[7+2*i])
	}

	return dz, nil
}

func (dz *Reader) Close() error {
	return dz.fp.Close()
}

func (dz *Reader) ReadAt(p []byte, off int64) (n int, err error) {
	want := len(p)
	b, err := dz.Get(off, int64(want))
	if b != nil {
		got := len(b)
		copy(p, b)
		if want == got {
			return got, nil
		} else {
			return got, err
		}
	} else {
		return 0, err
	}
}

func (dz *Reader) Get(start, size int64) ([]byte, error) {
	if size == 0 {
		return []byte{}, nil
	}

	if start < 0 || size < 0 {
		return nil, fmt.Errorf("negative start or size")
	}

	if int(start/dz.blockSize) >= len(dz.offsets) {
		return nil, fmt.Errorf("start passed end of archive")
	}

	start1 := dz.blockSize * (start / dz.blockSize)
	size1 := size + (start - start1)

	dz.lock.Lock()
	defer dz.lock.Unlock()

	_, err := dz.fp.Seek(dz.offsets[start/dz.blockSize], 0)
	if err != nil {
		return nil, err
	}
	rd := flate.NewReader(dz.fp)

	data := make([]byte, size1)
	_, err = io.ReadFull(rd, data)
	if err != nil {
		return nil, err
	}

	return data[start-start1:], nil
}

// Start and size in base64 notation, such as used by the `dictunzip` program.
func (dz *Reader) GetB64(start, size string) ([]byte, error) {
	start2, err := decode(start)
	if err != nil {
		return nil, err
	}
	size2, err := decode(size)
	if err != nil {
		return nil, err
	}
	return dz.Get(start2, size2)
}

//. Base64 decoder

var index = []uint64{
	99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99,
	99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99,
	99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 62, 99, 99, 99, 63,
	52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 99, 99, 99, 99, 99, 99,
	99, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14,
	15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 99, 99, 99, 99, 99,
	99, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
	41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 99, 99, 99, 99, 99,
	99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99,
	99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99,
	99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99,
	99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99,
	99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99,
	99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99,
	99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99,
	99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99, 99,
}

func decode(val string) (int64, error) {
	var result uint64
	var offset uint64

	for i := len(val) - 1; i >= 0; i-- {
		tmp := index[val[i]]
		if tmp == 99 {
			return 0, fmt.Errorf("illegal character in base64 value: %v", val[i:i+1])
		}

		if (tmp<<offset)>>offset != tmp {
			return 0, fmt.Errorf("type uint64 cannot store decoded base64 value: %v", val)
		}

		result |= tmp << offset
		offset += 6
	}
	return int64(result), nil
}
