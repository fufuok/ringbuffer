package ringbuffer

import (
	"testing"

	"github.com/fufuok/ringbuffer/internal/assert"
)

func TestRingBuffer(t *testing.T) {
	rb := NewUnbounded(10)
	v, err := rb.Read()
	assert.Nil(t, v)
	assert.NotNil(t, err, ErrIsEmpty)

	write := 0
	read := 0

	// write one and read it
	rb.Write(0)
	v, err = rb.Read()
	assert.Nil(t, err)
	assert.Equal(t, 0, v)
	assert.Equal(t, 1, rb.r)
	assert.Equal(t, 1, rb.w)
	assert.True(t, rb.IsEmpty())

	// then write 10
	for i := 0; i < 9; i++ {
		rb.Write(i)
		write += i
	}
	assert.Equal(t, 10, rb.Capacity())
	assert.Equal(t, 9, rb.Len())

	// write one more, the buffer is full so it grows
	rb.Write(10)
	write += 10
	assert.Equal(t, 20, rb.Capacity())
	assert.Equal(t, 10, rb.Len())

	for i := 0; i < 90; i++ {
		rb.Write(i)
		write += i
	}

	assert.Equal(t, 160, rb.Capacity())
	assert.Equal(t, 100, rb.Len())

	for {
		v, err := rb.Read()
		if err == ErrIsEmpty {
			break
		}

		read += v.(int)
	}

	assert.Equal(t, write, read)

	rb.Reset()
	assert.Equal(t, 10, rb.Capacity())
	assert.Equal(t, 0, rb.Len())
	assert.True(t, rb.IsEmpty())
}

func TestRingBuffer_One(t *testing.T) {
	rb := New(1)
	v, err := rb.Read()
	assert.Nil(t, v)
	assert.NotNil(t, err, ErrIsEmpty)

	write := 0
	read := 0

	// write one and read it
	rb.Write(0)
	v, err = rb.Read()
	assert.Nil(t, err)
	assert.Equal(t, 0, v)
	assert.Equal(t, 1, rb.r)
	assert.Equal(t, 1, rb.w)
	assert.True(t, rb.IsEmpty())

	// then write 10
	for i := 0; i < 9; i++ {
		rb.Write(i)
		write += i
	}
	assert.Equal(t, 16, rb.Capacity())
	assert.Equal(t, 9, rb.Len())

	// write one more, the buffer is full so it grows
	rb.Write(10)
	write += 10
	assert.Equal(t, 16, rb.Capacity())
	assert.Equal(t, 10, rb.Len())

	for i := 0; i < 90; i++ {
		rb.Write(i)
		write += i
	}

	assert.Equal(t, 128, rb.Capacity())
	assert.Equal(t, 100, rb.Len())

	for {
		v, err := rb.Read()
		if err == ErrIsEmpty {
			break
		}

		read += v.(int)
	}

	assert.Equal(t, write, read)

	rb.Reset()
	assert.Equal(t, minBufferSize, rb.Capacity())
	assert.Equal(t, 0, rb.Len())
	assert.True(t, rb.IsEmpty())
}

func TestRingBuffer_MaxSize(t *testing.T) {
	rb := NewFixed(10)
	v, err := rb.Read()
	assert.Nil(t, v)
	assert.NotNil(t, err, ErrIsEmpty)

	// write one and read it
	rb.Write(0)
	v, err = rb.Read()
	assert.Nil(t, err)
	assert.Equal(t, 0, v)
	assert.Equal(t, 1, rb.r)
	assert.Equal(t, 1, rb.w)
	assert.True(t, rb.IsEmpty())

	// then write 10
	for i := 0; i < 9; i++ {
		rb.Write(i)
	}
	assert.Equal(t, 10, rb.Capacity())
	assert.Equal(t, 9, rb.Len())

	// write one more, the buffer is full so it grows
	rb.Write(10)
	assert.Equal(t, 20, rb.Capacity())
	assert.Equal(t, 10, rb.Len())

	for i := 0; i < 90; i++ {
		rb.Write(i)
	}

	assert.Equal(t, 20, rb.Capacity())
	assert.Equal(t, 10, rb.Len())
	assert.Equal(t, uint64(90), rb.Discards())
	assert.Equal(t, 10, rb.MaxSize())

	// Unbounded
	rb.SetMaxSize(0)

	for i := 0; i < 90; i++ {
		rb.Write(i)
	}

	assert.Equal(t, 160, rb.Capacity())
	assert.Equal(t, 100, rb.Len())
	assert.Equal(t, 0, rb.MaxSize())

	maxSize := 2
	rb.SetMaxSize(maxSize)
	callbackDiscardsCount := 0
	rb.SetOnDiscards(func(v interface{}) {
		callbackDiscardsCount++
	})

	for i := 0; i < 90; i++ {
		rb.Write(i)
	}

	assert.Equal(t, maxSize+1, rb.Capacity())
	assert.Equal(t, maxSize, rb.Len())
	assert.Equal(t, uint64(180), rb.Discards())
	assert.Equal(t, maxSize, rb.MaxSize())
	assert.Equal(t, 90, callbackDiscardsCount)

	for {
		if _, err := rb.Read(); err == ErrIsEmpty {
			break
		}
	}
	assert.Equal(t, maxSize+1, rb.Capacity())
	assert.Equal(t, 0, rb.Len())
	assert.True(t, rb.IsEmpty())
}

func TestRingBuffer_PeekAll(t *testing.T) {
	rb := New(3, 4)
	assert.Nil(t, rb.PeekAll())
	assert.Nil(t, rb.LPeekN(1))
	assert.Nil(t, rb.RPeekN(1))

	rb.Write(1)
	assert.Equal(t, []T{1}, rb.PeekAll())
	assert.Equal(t, []T{1}, rb.LPeekN(3))
	assert.Equal(t, []T{1}, rb.RPeekN(3))

	for i := 0; i < 10; i++ {
		rb.Write(i)
	}
	assert.Equal(t, []T{1, 0, 1, 2}, rb.PeekAll())
	assert.Equal(t, []T{1, 0, 1}, rb.LPeekN(3))
	assert.Equal(t, []T{0, 1, 2}, rb.RPeekN(3))

	rb.SetMaxSize(0)

	for i := 0; i < 5; i++ {
		rb.Write(i)
	}
	assert.Equal(t, []T{1, 0, 1, 2, 0, 1, 2, 3, 4}, rb.PeekAll())
	assert.Equal(t, []T{1, 0, 1}, rb.LPeekN(3))
	assert.Equal(t, []T{2, 3, 4}, rb.RPeekN(3))
}

func TestRingBuffer_RRead(t *testing.T) {
	rb := NewFixed(2)

	rb.Write(1)
	v, err := rb.RPeek()
	assert.Nil(t, err)
	assert.Equal(t, 1, v)
	assert.Equal(t, 1, rb.Len())

	v, err = rb.RRead()
	assert.Nil(t, err)
	assert.Equal(t, 1, v)
	assert.True(t, rb.IsEmpty())

	v, err = rb.RPeek()
	assert.NotNil(t, err, ErrIsEmpty)
	assert.Equal(t, nil, v)
	v, err = rb.RRead()
	assert.NotNil(t, err, ErrIsEmpty)
	assert.Equal(t, nil, v)
	assert.Equal(t, 0, rb.w)

	for i := 0; i < 5; i++ {
		rb.Write(i)
		_, _ = rb.Read()
	}
	rb.Write(7)
	assert.Equal(t, 0, rb.w)

	v, err = rb.RPeek()
	assert.Nil(t, err)
	assert.Equal(t, 7, v)
	assert.Equal(t, 1, rb.Len())

	v, err = rb.RRead()
	assert.Nil(t, err)
	assert.Equal(t, 7, v)
	assert.True(t, rb.IsEmpty())
	assert.Equal(t, 1, rb.w)
}

func TestRingBuffer_Overwrite(t *testing.T) {
	rb := NewUnbounded(5)
	for i := 0; i < 10; i++ {
		rb.Write(i)
	}
	assert.Equal(t, 10, rb.Len())

	rb.SetMaxSize(3)
	assert.Equal(t, 3, rb.Len())
	assert.Equal(t, []T{7, 8, 9}, rb.PeekAll())

	rb.Write(10)
	assert.Equal(t, []T{7, 8, 9}, rb.PeekAll())

	rb.Overwrite(10)
	assert.Equal(t, []T{8, 9, 10}, rb.PeekAll())
	rb.Overwrite(11)
	assert.Equal(t, []T{9, 10, 11}, rb.PeekAll())

	rb.Truncate(2)
	assert.Equal(t, []T{10, 11}, rb.PeekAll())
}
