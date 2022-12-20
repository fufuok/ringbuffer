# ringbuffer

Infinite ring buffer with auto-expanding or fixed capacity.

**It is not thread-safe(goroutine-safe) so you must use the lock-like synchronization primitive to use it in multiple writers and multiple readers.**

*forked from smallnest/chanx*

## DOC

see: [DOC.md](DOC.md)

```go
package ringbuffer // import "github.com/fufuok/ringbuffer"

var ErrIsEmpty = errors.New("ringbuffer is empty")
type RingBuffer struct{ ... }
    func New(initialSize int, maxBufferSize ...int) *RingBuffer
    func NewFixed(initialSize int) *RingBuffer
    func NewUnbounded(initialSize int) *RingBuffer
type RingBufferOf[T any] struct{ ... }
    func NewFixedOf[T any](initialSize int) *RingBufferOf[T]
    func NewOf[T any](initialSize int, maxBufferSize ...int) *RingBufferOf[T]
    func NewUnboundedOf[T any](initialSize int) *RingBufferOf[T]
type T interface{}
```

## Examples

see: [examples](examples)

```go
//go:build go1.18
// +build go1.18

package main

import (
	"fmt"
	"strconv"

	"github.com/fufuok/ringbuffer"
)

func main() {
	// The initial size of the buffer is 10,
	// unlimited expansion, never full.
	rb := ringbuffer.NewUnboundedOf[string](10)
	// or
	rb = ringbuffer.NewOf[string](10, 0)

	// The buffer max size is fixed at 10.
	rb = ringbuffer.NewFixedOf[string](10)
	// or
	// The buffer max size is fixed at 20.
	rb = ringbuffer.NewOf[string](10, 20)
	// or
	// The buffer max size is fixed at 5.
	rb.SetMaxSize(5)

	rb.SetOnDiscards(func(v string) {
		fmt.Println("discards:", v)
	})

	// write
	rb.Write("A")
	fmt.Println(rb.Len())

	// read
	v, _ := rb.Read()
	fmt.Println(v)

	for i := 0; i < 10; i++ {
		rb.Write(strconv.Itoa(i))
	}

	latest, _ := rb.RPeek()
	fmt.Println(latest)

	popLatest, _ := rb.RRead()
	fmt.Println(popLatest)

	all := rb.PeekAll()
	fmt.Println(all)

	top3 := rb.RPeekN(3)
	fmt.Println(top3)

	rb.Reset()

	// Output:
	// 1
	// A
	// discards: 5
	// discards: 6
	// discards: 7
	// discards: 8
	// discards: 9
	// 4
	// 4
	// [0 1 2 3]
	// [1 2 3]
}
```









*ff*

