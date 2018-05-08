package main

import (
	"fmt"

	"github.com/mariadesouza/workerpool"
)

func findandStoreFibonacci(n int, fibonacci *int64) {
	var x, y int64
	x, y = 1, 1
	for i := 1; i < n; i++ {
		x, y = y, x+y
	}
	*fibonacci = y
}

func main() {
	s := 20
	var fibonacci [20]int64
	wp := workerpool.New(3)
	defer wp.Close()
	for i := 0; i < s; i++ {
		index := i
		wp.AddWorkToPool(
			func() {
				findandStoreFibonacci(index, &fibonacci[index])
			})
		//fmt.Printf("+")
	}
	fmt.Println(fibonacci)
}
