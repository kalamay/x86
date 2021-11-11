package jit

import (
	"math/bits"
	"unsafe"
)

type sliceHeader struct {
	data unsafe.Pointer
	len  int
	cap  int
}

type stringHeader struct {
	data unsafe.Pointer
	len  int
}

func Memset(b []byte, c byte) {
	// TODO: move this out
	n := len(b)
	if n >= 16 {
		memset16(b[:n&(^15)], c)
	}
	for i := n & (^15); i < n; i++ {
		b[i] = c
	}
}

func memset16(b []byte, c byte)

func truncateSize(n uint64) (uint64, string) {
	const (
		k = 1024
		m = 1024 * k
		g = 1024 * m
	)

	switch {
	case n >= g:
		return n / g, "G"
	case n >= m:
		return n / m, "M"
	case n >= k:
		return n / k, "K"
	default:
		return n, "B"
	}
}

func p2(n, min int) int {
	if n < min {
		return min
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n + 1
}

func allocSize(n, min, max int) (int, int) {
	switch {
	case n <= min:
		return min, 0
	case n > max:
		return (n + (pageSize - 1)) / pageSize * pageSize, -1
	default:
		size, bucket := min, 0
		for size <= n { // pad allocation by at least 1 byte
			size <<= 1
			bucket++
		}
		return size, bucket
	}
}

func bucketOf(n, min, max int) int {
	switch {
	case n <= min:
		return 0
	case n > max:
		return -1
	default:
		return bits.TrailingZeros32(uint32(n >> bits.TrailingZeros32(uint32(min))))
	}
}
