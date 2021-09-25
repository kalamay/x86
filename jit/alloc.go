package jit

import (
	"fmt"
	"math/bits"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

type Alloc []byte

func (a Alloc) Addr() unsafe.Pointer {
	return (*sliceHeader)(unsafe.Pointer(&a)).data
}

func (a Alloc) String() string {
	p, buf, tail := 0, strings.Builder{}, [16]byte{}
	for i := 0; i < len(a); i++ {
		p = i & 15
		if p == 0 {
			if i > 0 {
				buf.WriteString("  ")
				buf.Write(tail[:])
				buf.WriteByte('\n')
			}
			fmt.Fprintf(&buf, "%08x:", i)
		}
		if (i & 1) == 0 {
			buf.WriteByte(' ')
		}
		fmt.Fprintf(&buf, "%02x", a[i])
		if ' ' <= a[i] && a[i] <= '~' {
			tail[p] = a[i]
		} else {
			tail[p] = '.'
		}
	}
	if r := 16 - p; r < 16 {
		fmt.Fprintf(&buf, "%*s", r*2+(r-1)/2, "")
		buf.Write(tail[:p])
		buf.WriteByte('\n')
	}
	return buf.String()
}

const (
	DefaultMinSize   = 256
	DefaultMaxSize   = 8192
	DefaultMapPages  = 32
	DefaultFill      = true
	DefaultFillValue = byte(0xc3)

	pageSize = 4096
	pageMask = ^uintptr(0) >> 12 << 12
	rx       = syscall.PROT_READ | syscall.PROT_EXEC
	rw       = syscall.PROT_READ | syscall.PROT_WRITE
	rwx      = syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_WRITE
	flags    = syscall.MAP_ANON | syscall.MAP_PRIVATE
)

type PoolConfig struct {
	MinSize int
	MaxSize int

	// MapPages controls how many pages are requested from the OS at a time.
	MapPages int

	// Fill sets the type of fill to apply to each allocation for any trailing
	// bytes. This is used to clear prior instructions. Any single-byte string
	// sets the fill byte (i.e. "\x00" for zero-fill). The special value "return"
	// may be used to fill with the RET instruction. The special value "none"
	// disables the fill. Any other value uses the defailt.
	Fill string

	// Unprotected disables the page protection mode. Turning off protection means
	// the allocation is both executable and writable. This can improve performance
	// by reducing syscalls, but this should only been when use of the allocation is
	// fully trusted.
	Unprotected bool
}

type Pool struct {
	mtx     sync.Mutex
	buckets [][][]byte
	stats   PoolStats
	minSize int
	maxSize int
	mapSize int
	fillMem bool
	fillVal byte
	protect bool
}

func NewPool(c PoolConfig) *Pool {
	minSize := DefaultMinSize
	maxSize := DefaultMaxSize
	mapSize := DefaultMapPages * pageSize
	fillMem := DefaultFill
	fillVal := DefaultFillValue

	if c.MinSize > 0 {
		minSize = p2(c.MinSize, 16)
	}
	if c.MaxSize > 0 {
		maxSize = p2(c.MaxSize, 16)
	}
	if c.MapPages > 0 {
		mapSize = c.MapPages * pageSize
	}
	if maxSize < minSize {
		maxSize = minSize
	}
	if mapSize < maxSize*2 {
		mapSize = maxSize * 2
	}

	switch c.Fill {
	case "none":
		fillMem = false
	case "return":
		fillMem = true
		fillVal = 0xc3
	default:
		if len(c.Fill) == 1 {
			fillMem = true
			fillVal = c.Fill[0]
		}
	}

	n := bits.TrailingZeros64(uint64(maxSize)) - bits.TrailingZeros64(uint64(minSize)) + 1

	return &Pool{
		buckets: make([][][]byte, n),
		stats:   PoolStats{Buckets: make([]PoolStatCounts, n)},
		minSize: minSize,
		maxSize: maxSize,
		mapSize: mapSize,
		fillMem: fillMem,
		fillVal: fillVal,
		protect: !c.Unprotected,
	}
}

func (p *Pool) String() string {
	poolSize := len(p.buckets)
	lens := make([]int, poolSize)

	p.mtx.Lock()
	for i := 0; i < poolSize; i++ {
		lens[i] = len(p.buckets[i])
	}
	p.mtx.Unlock()

	s := PoolStats{}
	p.Stats(&s)

	buf := strings.Builder{}
	fmt.Fprintf(&buf, "jit.Pool(%p) {\n\t      available  allocs/frees", p)
	blockSize := uint64(p.minSize)
	for i := 0; i < poolSize; i++ {
		sz, suf := truncateSize(blockSize)
		fmt.Fprintf(&buf, "\n\t%3d%s: %-10d %d/%d", sz, suf, lens[i], s.Buckets[i].Allocs, s.Buckets[i].Frees)
		blockSize <<= 1
	}
	fmt.Fprintf(&buf, `
	over: -          %d/%d
	min: %d, max: %d
}`,
		s.Overage.Allocs, s.Overage.Frees,
		s.MinSize, s.MaxSize)

	return buf.String()
}

func (p *Pool) reportMinMax(size uint64) {
	for {
		s := atomic.LoadUint64(&p.stats.MinSize)
		if (s > 0 && size > s) || atomic.CompareAndSwapUint64(&p.stats.MinSize, s, size) {
			break
		}
	}

	for {
		s := atomic.LoadUint64(&p.stats.MaxSize)
		if size < s || atomic.CompareAndSwapUint64(&p.stats.MaxSize, s, size) {
			break
		}
	}
}

func (p *Pool) alloc(s int) (dst []byte, err error) {
	p.reportMinMax(uint64(s))

	size, bucket := allocSize(s, p.minSize, p.maxSize)

	if bucket < 0 {
		atomic.AddUint64(&p.stats.Overage.Allocs, 1)
		return mmap(size)
	}

	atomic.AddUint64(&p.stats.Buckets[bucket].Allocs, 1)

	poolSize, min, max := len(p.buckets), p.minSize, p.maxSize

	p.mtx.Lock()
	defer p.mtx.Unlock()

	at := bucket
	for at < poolSize && len(p.buckets[at]) == 0 {
		at++
	}

	if at == bucket {
		n := len(p.buckets[bucket]) - 1
		dst, p.buckets[bucket] = p.buckets[bucket][n], p.buckets[bucket][:n]
		return
	}

	var val []byte

	if at == poolSize {
		if val, err = mmap(p.mapSize); err != nil {
			return
		}
		for len(val) > max {
			p.buckets[poolSize-1] = append(p.buckets[poolSize-1], val[:max])
			val = val[max:]
		}
		at -= 2
		atomic.AddUint64(&p.stats.Maps, 1)
	} else {
		n := len(p.buckets[at]) - 1
		val, p.buckets[at] = p.buckets[at][n], p.buckets[at][:n]
		at--
	}

	sz := min << at
	for ; at >= bucket; at-- {
		p.buckets[at] = append(p.buckets[at], val[:sz])
		val = val[sz:]
		sz >>= 1
	}

	return val, nil
}

func (p *Pool) Alloc(src []byte) (Alloc, error) {
	b, err := p.alloc(len(src))
	if err != nil {
		return nil, err
	}

	if p.protect {
		pg := pageOf(b)
		mwrite(pg)
		if p.fillMem {
			memset16(b[len(src) & ^127:], p.fillVal)
		}
		copy(b, src)
		mexec(pg)
	} else {
		copy(b, src)
	}

	return codeOf(b, len(src)), nil
}

func (p *Pool) Free(c *Alloc) {
	b := blockOf(*c)
	bucket := bucketOf(len(b), p.minSize, p.maxSize)
	*c = nil

	if bucket < 0 {
		atomic.AddUint64(&p.stats.Overage.Frees, 1)
		return
	}

	atomic.AddUint64(&p.stats.Buckets[bucket].Frees, 1)

	p.mtx.Lock()
	p.buckets[bucket] = append(p.buckets[bucket], b)
	p.mtx.Unlock()
}

func (p *Pool) Stats(s *PoolStats) {
	p.stats.Copy(s)
}

type PoolStats struct {
	Buckets []PoolStatCounts
	Overage PoolStatCounts
	Maps    uint64
	MinSize uint64
	MaxSize uint64
}

type PoolStatCounts struct {
	Allocs uint64
	Frees  uint64
}

func (s *PoolStats) Copy(dst *PoolStats) {
	poolSize := len(s.Buckets)
	if len(dst.Buckets) != poolSize {
		dst.Buckets = make([]PoolStatCounts, poolSize)
	}
	for i := 0; i < poolSize; i++ {
		dst.Buckets[i].Allocs = atomic.LoadUint64(&s.Buckets[i].Allocs)
		dst.Buckets[i].Frees = atomic.LoadUint64(&s.Buckets[i].Frees)
	}
	dst.Overage.Allocs = atomic.LoadUint64(&s.Overage.Allocs)
	dst.Overage.Frees = atomic.LoadUint64(&s.Overage.Frees)
	dst.Maps = atomic.LoadUint64(&s.Maps)
	dst.MinSize = atomic.LoadUint64(&s.MinSize)
	dst.MaxSize = atomic.LoadUint64(&s.MaxSize)
}

func mmap(n int) ([]byte, error) {
	return syscall.Mmap(-1, 0, n, rx, flags)
}

func mwrite(b []byte) {
	if err := syscall.Mprotect(b, rwx); err != nil {
		panic(err)
	}
}

func mexec(b []byte) {
	if err := syscall.Mprotect(b, rx); err != nil {
		panic(err)
	}
}

func codeOf(b []byte, n int) Alloc {
	hdr := *(*sliceHeader)(unsafe.Pointer(&b))
	return Alloc(*(*[]byte)(unsafe.Pointer(&sliceHeader{
		data: hdr.data,
		len:  n,
		cap:  hdr.cap,
	})))
}

func pageOf(b []byte) []byte {
	hdr := *(*sliceHeader)(unsafe.Pointer(&b))
	n := (hdr.len + (pageSize - 1)) / pageSize * pageSize
	return *(*[]byte)(unsafe.Pointer(&sliceHeader{
		data: unsafe.Pointer(uintptr(hdr.data) & pageMask),
		len:  n,
		cap:  n,
	}))
}

func blockOf(c Alloc) []byte {
	hdr := *(*sliceHeader)(unsafe.Pointer(&c))
	return *(*[]byte)(unsafe.Pointer(&sliceHeader{
		data: hdr.data,
		len:  hdr.cap,
		cap:  hdr.cap,
	}))
}
