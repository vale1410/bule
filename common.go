package main

import (
	"errors"
	"fmt"
)

func assert(condition bool) {
	if !condition {
		panic(errors.New(""))
	}
}

func asserts(condition bool, info string) {
	if !condition {
		fmt.Println(info)
		panic(errors.New(info))
	}
}

func asserte(err error) {
	if err != nil {
		panic(err)
	}
}

func makeSet(a, b int) (c []int) {
	if a >= b {
		return []int{}
	}
	c = make([]int, 0, b-a)
	for i := a; i <= b; i++ {
		c = append(c, i)
	}
	return
}
