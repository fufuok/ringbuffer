package ringbuffer

import (
	"errors"
)

const minBufferSize = 2

var ErrIsEmpty = errors.New("ringbuffer is empty")

type T interface{}

// RingBuffer is a ring buffer for common types.
// It is never full and always grows if it will be full.
// It is not thread-safe(goroutine-safe) so you must use the lock-like synchronization primitive
// to use it in multiple writers and multiple readers.
// Exceeding maxSize, data will be discarded.
type RingBuffer struct {
	buf         []T
	initialSize int
	size        int
	maxSize     int
	discards    uint64
	r           int // read pointer
	w           int // write pointer
	onDiscards  func(interface{})
}

func NewUnbounded(initialSize int) *RingBuffer {
	return New(initialSize, 0)
}

func NewFixed(initialSize int) *RingBuffer {
	return New(initialSize, initialSize)
}

func New(initialSize int, maxBufferSize ...int) *RingBuffer {
	if initialSize < minBufferSize {
		initialSize = minBufferSize
	}

	maxSize := 0
	if len(maxBufferSize) > 0 && maxBufferSize[0] >= minBufferSize {
		maxSize = maxBufferSize[0]
	}

	return &RingBuffer{
		buf:         make([]T, initialSize),
		initialSize: initialSize,
		size:        initialSize,
		maxSize:     maxSize,
	}
}

func (r *RingBuffer) Read() (T, error) {
	if r.r == r.w {
		return nil, ErrIsEmpty
	}

	v := r.buf[r.r]
	r.r++
	if r.r == r.size {
		r.r = 0
	}

	return v, nil
}

// RRead erases the last written data, and returns that data.
func (r *RingBuffer) RRead() (T, error) {
	if r.r == r.w {
		var t T
		return t, ErrIsEmpty
	}
	if r.w == 0 {
		r.w = r.size - 1
		return r.buf[r.w], nil
	}
	r.w--
	return r.buf[r.w], nil
}

func (r *RingBuffer) Peek() (T, error) {
	if r.r == r.w {
		return nil, ErrIsEmpty
	}
	return r.buf[r.r], nil
}

// RPeek get the latest written data.
func (r *RingBuffer) RPeek() (T, error) {
	if r.r == r.w {
		var t T
		return t, ErrIsEmpty
	}
	if r.w == 0 {
		return r.buf[r.size-1], nil
	}
	return r.buf[r.w-1], nil
}

func (r *RingBuffer) PeekAll() (buf []T) {
	if r.r == r.w {
		return
	}

	if r.w > r.r {
		buf = append(buf, r.buf[r.r:r.w]...)
		return
	}

	buf = append(buf, r.buf[r.r:]...)
	buf = append(buf, r.buf[:r.w]...)
	return
}

func (r *RingBuffer) RPeekN(n int) []T {
	if n <= 0 {
		return nil
	}

	buf := r.PeekAll()
	l := len(buf)
	if l >= n {
		return buf[l-n:]
	}
	return buf
}

func (r *RingBuffer) LPeekN(n int) []T {
	if n <= 0 {
		return nil
	}

	buf := r.PeekAll()
	if len(buf) >= n {
		return buf[:n]
	}
	return buf
}

func (r *RingBuffer) Write(v T) {
	if r.maxSize > 0 && r.Len() >= r.maxSize {
		r.discards++
		if r.onDiscards != nil {
			r.onDiscards(v)
		}
		return
	}

	r.buf[r.w] = v
	r.w++

	if r.w == r.size {
		r.w = 0
	}

	if r.w == r.r { // full
		r.grow()
	}
}

func (r *RingBuffer) grow() {
	var size int
	if r.size < 1024 {
		size = r.size * 2
	} else {
		size = r.size + r.size/4
	}

	buf := make([]T, size)

	copy(buf[0:], r.buf[r.r:])
	copy(buf[r.size-r.r:], r.buf[0:r.r])

	r.r = 0
	r.w = r.size
	r.size = size
	r.buf = buf
}

// Truncate discards all but the first n unread bytes from the buffer
// but continues to use the same allocated storage.
func (r *RingBuffer) Truncate(n int) {
	if n <= 0 {
		r.Reset()
		return
	}

	if r.Len() <= n {
		return
	}

	if r.size > n*2 {
		data := r.RPeekN(n)
		r.r = 0
		r.w = n
		r.size = n + 1
		r.buf = make([]T, r.size)
		copy(r.buf, data)
		return
	}

	if r.w > r.r {
		r.r = r.w - r.r - n
	} else {
		x := r.w - n
		if x >= 0 {
			r.r = x
		} else {
			r.r = r.size + x
		}
	}
}

func (r *RingBuffer) IsEmpty() bool {
	return r.r == r.w
}

// Capacity returns the size of the underlying buffer.
func (r *RingBuffer) Capacity() int {
	return r.size
}

func (r *RingBuffer) MaxSize() int {
	return r.maxSize
}

func (r *RingBuffer) Discards() uint64 {
	return r.discards
}

func (r *RingBuffer) Len() int {
	if r.r == r.w {
		return 0
	}

	if r.w > r.r {
		return r.w - r.r
	}

	return r.size - r.r + r.w
}

func (r *RingBuffer) Reset() {
	r.r = 0
	r.w = 0
	r.size = r.initialSize
	r.buf = make([]T, r.initialSize)
}

func (r *RingBuffer) SetMaxSize(n int) int {
	if n == 0 {
		// Unbounded
		r.maxSize = 0
	} else if n >= minBufferSize {
		// Reset maximum limit
		r.maxSize = n
		r.Truncate(n)
	}

	return r.maxSize
}

func (r *RingBuffer) SetOnDiscards(fn func(interface{})) {
	if fn != nil {
		r.onDiscards = fn
	}
}
