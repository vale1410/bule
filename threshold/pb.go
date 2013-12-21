package pb

import (
	"fmt"
	"github.com/vale1410/bule/sorters"
	"math"
)

type EquationType int

const (
	AtMost EquationType = iota
	AtLeast
	Equal
	Optimization
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
	Bags    [][]Literal
	LitIn   []Literal //Bags flattened, input to Sorter
	Sorter  sorters.Sorter
}

func (t *Threshold) CreateSortingEncoding(typ sorters.SortingNetworkType) {

	t.CreateBags()

	//typ := sorters.Pairwise

	layers := make([]sorters.Sorter, len(t.Bags))

	for i, bag := range t.Bags {

		layers[i] = sorters.CreateSortingNetwork(len(bag), -1, typ)

		t.LitIn = append(t.LitIn, bag...)
	}

	t.Sorter.In = make([]int, 0, len(t.LitIn))

	//collaps all into one sorter, doing the halving ...

	bIn := make([]int, 0)

	offset := 2

	fmt.Println("debug: layers", t.Bags)

	for i, layer := range layers {

		halver := sorters.CreateSortingNetwork(len(bIn), -1, typ)

		offset = halver.Normalize(offset, bIn)
		t.Sorter.Comparators = append(t.Sorter.Comparators, halver.Comparators...)

		fmt.Println(i, "debug: halver", halver)

		offset = layer.Normalize(offset, []int{})
		t.Sorter.Comparators = append(t.Sorter.Comparators, layer.Comparators...)

		fmt.Println(i, "debug: layer", layer)

		t.Sorter.In = append(t.Sorter.In, layer.In...)

		size := len(bIn) + len(layers[i].In)

		fmt.Println(i, "debug: size", size)

		combinedIn := make([]int, 0, size)
		combinedIn = append(combinedIn, halver.Out...)
		combinedIn = append(combinedIn, layer.Out...)

		fmt.Println(i, "debug: combinedSorter size,cut", size, len(bIn))
		combinedSorter := sorters.CreateSortingNetwork(size, len(bIn), typ)
		fmt.Println(i, "debug: combinedSorter", combinedSorter)
		fmt.Println(i, "debug: combinedIn", combinedIn)
		offset = combinedSorter.Normalize(offset, combinedIn)

		bIn = make([]int, len(combinedSorter.Out)/2)

		// halving circuit

		mapping := make(map[int]int, size)

		for j, x := range combinedSorter.Out {
			if j%2 == 1 {
				bIn[j/2] = x
			} else {
				mapping[combinedSorter.Out[j]] = -1
				combinedSorter.Out[j] = -1
			}
		}

		fmt.Println(i, "debug: mapping", mapping)

		combinedSorter.PropagateBackwards(mapping)

		fmt.Println(i, "debug: combinedSorter", combinedSorter)

		t.Sorter.Comparators = append(t.Sorter.Comparators, combinedSorter.Comparators...)

		fmt.Println(i, "debug: tSorter", t.Sorter)
	}

	offset = t.Sorter.Normalize(2, []int{})
	t.Sorter.Out = make([]int, 1)
	t.Sorter.Out[0] = offset - 1
	fmt.Println("final debug: tSorter", t.Sorter)
}

// transform negative weights
// check if maximum reaches K at all
// sort by weight
func (t *Threshold) Normalize() {

	total := int64(0)

	for _, e := range t.Entries {
		total += e.Weight
	}

	if total < t.K {
		fmt.Println("sum of weights is too low!")
	}

}

func (t *Threshold) CreateBags() {

	nBags := len(binary(t.K))
	bins := make([][]int, len(t.Entries))
	bagPos := make([]int, nBags)
	bagSize := make([]int, nBags)

	maxWeight := int64(0)

	for i, e := range t.Entries {
		bins[i] = binary(e.Weight)

		for j, x := range bins[i] {
			bagSize[len(bins[i])-j-1] += x
		}

		if maxWeight < e.Weight {
			maxWeight = e.Weight
		}

	}

	t.Bags = make([][]Literal, len(binary(maxWeight)))

	for i, _ := range t.Bags {
		t.Bags[i] = make([]Literal, bagSize[i]+1)
		//t.Bags[i][bagSize[i]] = Literal{true, Atom(100 + i)}
	}

	for i, e := range t.Entries {
		for j, x := range bins[i] {
			pos := len(bins[i]) - j - 1
			if x == 1 {
				t.Bags[pos][bagPos[pos]] = e.Lit
				bagPos[pos]++
			}
		}
	}

	fmt.Println(t.Bags)

}

func (t *Threshold) AddTare() {

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
