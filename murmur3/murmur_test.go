package murmur3

import (
	"testing"
)

var data = []struct {
	h32   uint32
	h64_1 uint64
	h64_2 uint64
	s     string
}{
	{0x00000000, 0x0000000000000000, 0x0000000000000000, ""},
	{0x248bfa47, 0xcbd8a7b341bd9b02, 0x5b1e906a48ae1d19, "hello"},
	{0x149bbb7f, 0x342fac623a5ebc8e, 0x4cdcbc079642414d, "hello, world"},
	{0xe31e8a70, 0xb89e5988b737affc, 0x664fc2950231b2cb, "19 Jan 2038 at 3:14:07 AM"},
	{0xd5c48bfc, 0xcd99481f9ee902c9, 0x695da1a38987b6e7, "The quick brown fox jumps over the lazy dog."},
}

func TestRef(t *testing.T) {
	for _, elem := range data {
		var h128 Hash128 = New128()
		h128.Write([]byte(elem.s))
		if v1, v2 := h128.Sum128(); v1 != elem.h64_1 || v2 != elem.h64_2 {
			t.Errorf("'%s': 0x%x-0x%x (want 0x%x-0x%x)", elem.s, v1, v2, elem.h64_1, elem.h64_2)
		}

		if v1, v2 := Sum128([]byte(elem.s)); v1 != elem.h64_1 || v2 != elem.h64_2 {
			t.Errorf("'%s': 0x%x-0x%x (want 0x%x-0x%x)", elem.s, v1, v2, elem.h64_1, elem.h64_2)
		}
	}
}

func TestIncremental(t *testing.T) {
	for _, elem := range data {
		h128 := New128()
		for i, j, k := 0, 0, len(elem.s); i < k; i = j {
			j = 2*i + 3
			if j > k {
				j = k
			}
			s := elem.s[i:j]
			print(s + "|")
			h128.Write([]byte(s))
		}
		println()
		if v1, v2 := h128.Sum128(); v1 != elem.h64_1 || v2 != elem.h64_2 {
			t.Errorf("'%s': 0x%x-0x%x (want 0x%x-0x%x)", elem.s, v1, v2, elem.h64_1, elem.h64_2)
		}
	}
}

func bench128(b *testing.B, length int) {
	buf := make([]byte, length)
	b.SetBytes(int64(length))
	b.ResetTimer()
	for range b.N {
		Sum128(buf)
	}
}

func Benchmark128_1(b *testing.B) {
	bench128(b, 1)
}

func Benchmark128_2(b *testing.B) {
	bench128(b, 2)
}

func Benchmark128_4(b *testing.B) {
	bench128(b, 4)
}

func Benchmark128_8(b *testing.B) {
	bench128(b, 8)
}

func Benchmark128_16(b *testing.B) {
	bench128(b, 16)
}

func Benchmark128_32(b *testing.B) {
	bench128(b, 32)
}

func Benchmark128_64(b *testing.B) {
	bench128(b, 64)
}

func Benchmark128_128(b *testing.B) {
	bench128(b, 128)
}

func Benchmark128_256(b *testing.B) {
	bench128(b, 256)
}

func Benchmark128_512(b *testing.B) {
	bench128(b, 512)
}

func Benchmark128_1024(b *testing.B) {
	bench128(b, 1024)
}

func Benchmark128_2048(b *testing.B) {
	bench128(b, 2048)
}

func Benchmark128_4096(b *testing.B) {
	bench128(b, 4096)
}

func Benchmark128_8192(b *testing.B) {
	bench128(b, 8192)
}

//---
