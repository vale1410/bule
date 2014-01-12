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

	layers := make([]sorters.Sorter, len(t.Bags))
	bitsBag := len(t.Bags)

	for i, bag := range t.Bags {

		layers[i] = sorters.CreateSortingNetwork(len(bag), -1, typ)

		t.LitIn = append(t.LitIn, bag...) // this might have to be reversed
	}

	t.Sorter.In = make([]int, 0, len(t.LitIn))

	offset := 2

	fmt.Println("debug: layers", t.Bags)

	// determine the constant and what to add on both sides
	nextPow2 := int64(1)
	bitsThreshold := 1

	for nextPow2 < t.K+1 {
		nextPow2 *= 2
		bitsThreshold++
	}

	tare := nextPow2 - (t.K + 1)

	bTare := binary(tare)
	bitsTare := len(bTare)

	fmt.Println("debug: bitsBag", bitsBag, "bitsThreshold", bitsThreshold, "bitsTare", bitsTare)

	// layerTare indicates if odd or even output is taken

	layerTare := bTare

	if bitsBag < bitsTare {
		layerTare = layerTare[:bitsBag]
	}

	// consistLastLayer identifies the nth output in the last layer
	consistLastLayer := int64(0)
	p := int64(1)

	if bitsBag < bitsTare {
		for _, b := range bTare[bitsBag:] {
			if b == 1 {
				consistLastLayer += p
			}
			p *= 2
		}
	}

	fmt.Println("debug: layerTare", layerTare, "nth output in last layer:", p)

	// output of sorter in layer $i-1$
	bIn := make([]int, 0)

	for i, layer := range layers {

		offset = layer.Normalize(offset, []int{})
		t.Sorter.Comparators = append(t.Sorter.Comparators, layer.Comparators...)

		fmt.Println(i, "debug: bIn for this layer", bIn)

		fmt.Println(i, "debug: layer", layer)

		t.Sorter.In = append(t.Sorter.In, layer.In...)

		size := len(bIn) + len(layers[i].In)

		fmt.Println(i, "debug: size", size)

		mergeIn := make([]int, 0, size)
		mergeIn = append(mergeIn, bIn...)
		mergeIn = append(mergeIn, layer.Out...)

        fmt.Println(i, "debug: merger preparation: size,cut", size, len(bIn))
		merger := sorters.CreateSortingNetwork(size, len(bIn), typ)
		offset = merger.Normalize(offset, mergeIn)
		fmt.Println(i, "debug: mergeSorter", merger)

        // halving circuit:

		mapping := make(map[int]int, size)

		odd := 1

		if i < len(bTare) && bTare[i] == 1 {
			odd = 0
			bIn = make([]int, (len(merger.Out)+1)/2)
			fmt.Println(i, "debug: lenMerger,tare i,odd", len(merger.Out), bTare[i], odd)
		} else {
			bIn = make([]int, len(merger.Out)/2)
			fmt.Println(i, "debug: lenMerger,odd", len(merger.Out), odd)
		}

		// Alternate depending on bTare
		for j, x := range merger.Out {
			if j%2 == odd {
				bIn[j/2] = x
				fmt.Println(i, "debug: bIn,j,odd", bIn, j, odd)
			} else {
				fmt.Println(i, "debug: bIn,j,odd", bIn, j, odd)
				mapping[merger.Out[j]] = -1
				merger.Out[j] = -1
			}
		}

		fmt.Println(i, "debug: merger", merger)

		t.Sorter.Comparators = append(t.Sorter.Comparators, merger.Comparators...)

		fmt.Println(i, "debug: mapping", mapping)

		t.Sorter.PropagateBackwards(mapping)

		fmt.Println(i, "debug: tSorter", t.Sorter)

	}

    // take the K power of two, and p (which is added... ) and 
    // figure out the output element, set rest to -1 and backprop

	offset = t.Sorter.Normalize(2, []int{})
	t.Sorter.Out = make([]int, 1)
	t.Sorter.Out[0] = offset - 1
	fmt.Println("final debug: tSorter", t.Sorter)
}

// Normalize: work in progress
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
		t.Bags[i] = make([]Literal, bagSize[i])
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
