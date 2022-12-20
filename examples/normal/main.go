package main

import (
	"fmt"
	"strconv"

	"github.com/fufuok/ringbuffer"
)

func main() {
	// The initial size of the buffer is 10,
	// unlimited expansion, never full.
	rb := ringbuffer.NewUnbounded(10)
	// or
	rb = ringbuffer.New(10, 0)

	// The buffer max size is fixed at 10.
	rb = ringbuffer.NewFixed(10)
	// or
	// The buffer max size is fixed at 20.
	rb = ringbuffer.New(10, 20)
	// or
	// The buffer max size is fixed at 5.
	rb.SetMaxSize(5)

	rb.SetOnDiscards(func(v interface{}) {
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
	// [0 1 2 3 4]
	// [2 3 4]
}
