package lib

import (
	"errors"
	"fmt"
)

var (
	DebugLevel int
)

func debug(level int, s ...interface{}) {
	if level <= DebugLevel {
		fmt.Println(s...)
	}
}

func assert(condition bool) {
	if !condition {
		panic("ASSERT FAILED")
	}
}

func asserts(condition bool, info ...string) {
	if !condition {
		s := ""
		for _, x := range info {
			s += x + " "
		}
		fmt.Println(s)
		panic(errors.New(s))
	}
}

func asserte(err error) {
	if err != nil {
		panic(err)
	}
}

func assertx(err error, info ...string) {
	if err != nil {
		for _, s := range info {
			fmt.Print(s, " ")
		}
		fmt.Println()
		panic(err)
	}
}

func makeSet(a, b int) (c []int) {
	if a > b {
		return []int{}
	}
	c = make([]int, 0, b-a+1)
	for i := a; i <= b; i++ {
		c = append(c, i)
	}
	return
}


