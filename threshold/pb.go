package pb

import (
	"fmt"
	"math"
)

type EquationType int

const (
	AtMost EquationType = iota
	AtLeast
	Equal
)

type Entry struct {
	Lit    Literal
	Weight int64
}

type Atom int

type Literal struct {
	sign bool
	atom Atom
}

type Threshold struct {
	Desc    string
	Entries []Entry
	K       int64
	Typ     EquationType
}

// binary
// 23 = 10111
func binary(n int64) (bin []int) {

	s := int64(math.Logb(float64(n))) + 1
	bin = make([]int, s)

	i := s
	var m int64

	for n != 0 {
		i--
		m = n / 2
		//fmt.Println(i, n, m)
		if n != m*2 {
			bin[i] = 1
		}
		n = m
	}
	return
}

func (t *Threshold) Print2() {
	fmt.Println(t.Desc)

	first := true
	for _, x := range t.Entries {
		l := x.Lit
		if !first {
			fmt.Printf("+ ")
		}
		first = false

		bin := binary(x.Weight)

		for _, i := range bin {
			fmt.Print(i)
		}

		if l.sign {
			fmt.Print(" * ")
		} else {
			fmt.Print(" *~")
		}
		//fmt.Print(l.atom.P, "(", l.atom.V1, ",", l.atom.V2, ")")
		fmt.Print("x", l.atom, " ")
	}
	switch t.Typ {
	case AtMost:
		fmt.Print(" <= ")
	case AtLeast:
		fmt.Print(" >= ")
	case Equal:
		fmt.Print(" == ")
	}

	bin := binary(t.K)

	for _, i := range bin {
		fmt.Print(i)
	}

	fmt.Println()
	fmt.Println()
}

func (t *Threshold) Print10() {
	fmt.Println(t.Desc)

	first := true
	for _, x := range t.Entries {
		l := x.Lit
		if !first {
			fmt.Printf("+ ")
		}
		first = false

		fmt.Print(x.Weight)

		if l.sign {
			fmt.Print(" * ")
		} else {
			fmt.Print(" *~")
		}
		//fmt.Print(l.atom.P, "(", l.atom.V1, ",", l.atom.V2, ")")
		fmt.Print("x", l.atom, " ")
	}
	switch t.Typ {
	case AtMost:
		fmt.Print(" <= ")
	case AtLeast:
		fmt.Print(" >= ")
	case Equal:
		fmt.Print(" == ")
	}
	fmt.Println(t.K)

}
