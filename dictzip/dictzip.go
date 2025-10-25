package dictzip

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type Reader struct {
	f          *os.File
	chunkSize  int
	chunkCount int
	offsets    []uint32
	dataStart  int64
}

func NewReader(f *os.File) (*Reader, error) {
	hdr := make([]byte, 10)
	if _, err := io.ReadFull(f, hdr); err != nil {
		return nil, fmt.Errorf("read gzip header: %w", err)
	}
	if hdr[0] != 0x1f || hdr[1] != 0x8b {
		return nil, fmt.Errorf("not a gzip/dictzip file")
	}
	if hdr[3]&0x04 == 0 {
		return nil, fmt.Errorf("missing FEXTRA flag (not dictzip)")
	}

	var xlenBuf [2]byte
	if _, err := io.ReadFull(f, xlenBuf[:]); err != nil {
		return nil, fmt.Errorf("read XLEN: %w", err)
	}
	xlen := int(binary.LittleEndian.Uint16(xlenBuf[:]))

	extra := make([]byte, xlen)
	if _, err := io.ReadFull(f, extra); err != nil {
		return nil, fmt.Errorf("read FEXTRA: %w", err)
	}

	var r Reader
	pos := 0
	for pos+4 <= len(extra) {
		si1, si2 := extra[pos], extra[pos+1]
		subLen := int(binary.LittleEndian.Uint16(extra[pos+2 : pos+4]))
		pos += 4
		if pos+subLen > len(extra) {
			return nil, fmt.Errorf("invalid subfield len")
		}
		if si1 == 'R' && si2 == 'A' {
			sub := extra[pos : pos+subLen]
			if len(sub) < 6 {
				return nil, fmt.Errorf("RA too short (%d)", len(sub))
			}
			version := binary.LittleEndian.Uint16(sub[0:2])
			if version != 1 {
				return nil, fmt.Errorf("unsupported RA version %d", version)
			}

			// try all known layouts
			type layout struct {
				chunkSizeOfs  int
				chunkSizeLen  int
				chunkCountOfs int
				chunkCountLen int
				offsetStart   int
			}
			candidates := []layout{
				{2, 2, 4, 2, 6},  // GNU
				{2, 4, 6, 4, 10}, // StarDict
				{2, 2, 4, 4, 8},  // mixed 8-byte header
			}
			var ok bool
			for _, l := range candidates {
				if len(sub) < l.offsetStart {
					continue
				}
				var cs, cc int
				switch l.chunkSizeLen {
				case 2:
					cs = int(binary.LittleEndian.Uint16(sub[l.chunkSizeOfs : l.chunkSizeOfs+2]))
				case 4:
					cs = int(binary.LittleEndian.Uint32(sub[l.chunkSizeOfs : l.chunkSizeOfs+4]))
				}
				switch l.chunkCountLen {
				case 2:
					cc = int(binary.LittleEndian.Uint16(sub[l.chunkCountOfs : l.chunkCountOfs+2]))
				case 4:
					cc = int(binary.LittleEndian.Uint32(sub[l.chunkCountOfs : l.chunkCountOfs+4]))
				}
				exp := l.offsetStart + 4*(cc+1)
				if len(sub) >= exp {
					r.chunkSize, r.chunkCount = cs, cc
					r.offsets = make([]uint32, cc+1)
					for i := 0; i <= cc; i++ {
						r.offsets[i] = binary.LittleEndian.Uint32(
							sub[l.offsetStart+i*4 : l.offsetStart+4+i*4])
					}
					ok = true
					break
				}
			}
			if !ok {
				return nil, fmt.Errorf("invalid RA offsets length (len=%d)", len(sub))
			}
		}
		pos += subLen
	}

	dataStart, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	r.f = f
	r.dataStart = dataStart
	return &r, nil
}

func (r *Reader) ReadAt(p []byte, off int64) (int, error) {
	if r.f == nil {
		return 0, fmt.Errorf("reader closed")
	}
	if r.chunkSize == 0 {
		return 0, fmt.Errorf("invalid chunk size")
	}
	if off < 0 {
		return 0, fmt.Errorf("negative offset")
	}
	startChunk := int(off / int64(r.chunkSize))
	if startChunk >= r.chunkCount {
		return 0, io.EOF
	}
	offsetInChunk := int(off % int64(r.chunkSize))

	n := 0
	for len(p) > 0 && startChunk < r.chunkCount {
		chunk, err := r.readChunk(startChunk)
		if err != nil {
			return n, err
		}
		if offsetInChunk >= len(chunk) {
			startChunk++
			offsetInChunk = 0
			continue
		}
		copied := copy(p, chunk[offsetInChunk:])
		n += copied
		p = p[copied:]
		startChunk++
		offsetInChunk = 0
	}
	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}

func (r *Reader) readChunk(i int) ([]byte, error) {
	if i < 0 || i >= r.chunkCount {
		return nil, fmt.Errorf("invalid chunk index")
	}
	start := int64(r.offsets[i])
	end := int64(r.offsets[i+1])
	if end <= start {
		return nil, fmt.Errorf("invalid offsets in chunk %d", i)
	}
	if _, err := r.f.Seek(r.dataStart+start, io.SeekStart); err != nil {
		return nil, err
	}
	comp := make([]byte, end-start)
	if _, err := io.ReadFull(r.f, comp); err != nil {
		return nil, err
	}
	zr, err := zlib.NewReader(bytes.NewReader(comp))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return io.ReadAll(zr)
}

func (r *Reader) Close() error {
	if r.f == nil {
		return nil
	}
	err := r.f.Close()
	r.f = nil
	return err
}
