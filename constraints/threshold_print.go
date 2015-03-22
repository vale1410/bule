package constraints

import (
	"fmt"
	"math"
	"os"
	//"sort"
	"strconv"
)

func (c Lits) Print() {

	for i, x := range c {
		fmt.Print(x.ToTxt())
		if i != len(c)-1 {
			fmt.Print(",")
		} else {
			fmt.Println()
		}
	}
}

func (c Chain) Print() {

	for i, x := range c {
		fmt.Print(x.ToTxt())
		if i != len(c)-1 {
			fmt.Print(" <-")
		} else {
			fmt.Println()
		}
	}
}

func (t *Threshold) Print2() {
	fmt.Println(t.Desc)

	first := true
	for _, x := range t.Entries {
		l := x.Literal
		if !first {
			fmt.Printf("+ ")
		}
		first = false

		fmt.Print(BinaryStr(x.Weight))
		fmt.Print(x.Weight)
		fmt.Print(l.ToTex())
	}
	switch t.Typ {
	case LE:
		fmt.Print(" <= ")
		fmt.Print(BinaryStr(t.K))
	case GE:
		fmt.Print(" >= ")
		fmt.Print(BinaryStr(t.K))
	case EQ:
		fmt.Print(" == ")
		fmt.Print(BinaryStr(t.K))
	case OPT:
	}

	fmt.Println()
	fmt.Println()
}

func (t *Threshold) WriteFormula(base int, file *os.File) {

	var w string

	file.Write([]byte("$$"))
	for _, x := range t.Entries {
		if base == 2 {
			w = BinaryStr(x.Weight)
		} else if base == 10 {
			w = fmt.Sprintf("%v", x.Weight)
		} else {
			panic("only base 2 and 10 supported")
		}
		if x.Weight < 0 {
			file.Write([]byte(fmt.Sprintf("%v \\cdot %v ", w, x.Literal.ToTex())))
		} else {
			file.Write([]byte(fmt.Sprintf("+%v \\cdot %v ", w, x.Literal.ToTex())))
		}
	}
	switch t.Typ {
	case LE:
		file.Write([]byte(" \\leq "))
	case GE:
		file.Write([]byte(" \\geq "))
	case EQ:
		file.Write([]byte(" = "))
	}
	if base == 2 {
		w = BinaryStr(t.K)
	} else if base == 10 {
		w = fmt.Sprintf("%v", t.K)
	} else {
		panic("only base 2 and 10 supported")
	}
	file.Write([]byte(fmt.Sprintf("%v $$\n", w)))
}

func (t *Threshold) PrintGurobi() {
	for _, x := range t.Entries {
		l := x.Literal

		if x.Weight > 0 {
			fmt.Printf(" +")
		}

		if x.Weight != 1 {
			fmt.Printf(" ")
			fmt.Print(x.Weight)
		}
		fmt.Print(l.ToTxt())
	}
	switch t.Typ {
	case LE:
		fmt.Print(" <= ")
	case GE:
		fmt.Print(" >= ")
	case EQ:
		fmt.Print(" = ")
	}
	fmt.Println(t.K)

}

func (t *Threshold) PrintPBO() {
	for _, x := range t.Entries {
		l := x.Literal

		if x.Weight > 0 {
			fmt.Printf("+")
		}
		fmt.Print(x.Weight, " ", l.ToPBO(), " ")
	}
	switch t.Typ {
	case LE:
		fmt.Print("<= ")
	case GE:
		fmt.Print(">= ")
	case EQ:
		fmt.Print("= ")
	}
	fmt.Println(t.K, ";")
}

func (t *Threshold) Print10() {
	for _, x := range t.Entries {
		l := x.Literal

		if x.Weight > 0 {
			fmt.Printf(" +")
		}
		if x.Weight == 1 {
			fmt.Print(l.ToTxt())
		} else {
			fmt.Print(" ", x.Weight, l.ToTxt())
		}
	}
	switch t.Typ {
	case LE:
		fmt.Print(" <= ")
	case GE:
		fmt.Print(" >= ")
	case EQ:
		fmt.Print(" = ")
	}
	fmt.Println(t.K, ";")

}

func (t *Threshold) PrintGringo() {

	if len(t.Entries) > 0 {

		switch t.Typ {
		case LE:
			fmt.Print(":- ", t.K+1, " [ ")
		case GE:
			fmt.Print(":- [ ")
		case EQ:
			fmt.Print(":- not ", t.K, " [ ")
		case OPT:
			fmt.Print("#minimize[")
		}

		for i, x := range t.Entries {
			if i != 0 {
				fmt.Print(" , ")
			}
			if x.Weight != 1 {
				fmt.Print(x.Literal.ToTxt(), "=", x.Weight)
			} else {
				fmt.Print(x.Literal.ToTxt())
			}
		}

		switch t.Typ {
		case LE:
			fmt.Print(" ]")
		case GE:
			fmt.Print(" ] ", t.K-1)
		case EQ:
			fmt.Print(" ] ", t.K)
		case OPT:
			fmt.Print("]")
		}
		fmt.Println(".")
	}

}

func BinaryStr(n int64) (s string) {
	bin := Binary(n)
	for i := len(bin) - 1; i >= 0; i-- {
		s += strconv.Itoa(bin[i])
	}
	return
}

// binary
// 23 = 10111
// special case if n==0 then return empty slice
// panic if n<0
func Binary(n int64) (bin []int) {

	var s int64

	if n < 0 {
		panic("binary representation of number smaller than 0")
	} else if n == 0 {
		s = 0
	} else {
		s = int64(math.Logb(float64(n))) + 1
	}

	bin = make([]int, s)

	i := 0
	var m int64

	for n != 0 {
		m = n / 2
		//fmt.Println(i, n, m)
		if n != m*2 {
			bin[i] = 1
		}
		n = m
		i++
	}
	return
}
