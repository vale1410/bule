package threshold

import (
	"fmt"
	"github.com/vale1410/bule/sat"
	"github.com/vale1410/bule/sorters"
	"math"
)

type EquationType int
type TranslationType int

const (
	AtMost EquationType = iota
	AtLeast
	Equal
	Optimization
)

const (
	SortingNetwork TranslationType = iota
	Facts
	SingleClause
	BDD
)

type Entry struct {
	Literal sat.Literal
	Weight  int64
}

//type Atom int
//
//type Literal struct {
//	Sign bool
//	Atom Atom
//}

type Threshold struct {
	Desc    string
	Entries []Entry
	K       int64
	Typ     EquationType
	Trans   TranslationType
	Pred    sat.Pred
	Clauses sat.ClauseSet
	Sorter  sorters.Sorter
	Bags    [][]sat.Literal
	LitIn   []sat.Literal //Bags flattened, input to Sorter
}

func (t *Threshold) OnlyFacts() (is bool) {

	t.Normalize()
	is = false

	if t.K == 0 {
		is = true
		t.Clauses = make(sat.ClauseSet, len(t.Entries))
		for i, x := range t.Entries {
			t.Clauses[i] = sat.Clause{"OF", []sat.Literal{x.Literal}}
		}

	}

	return is
}

func (t *Threshold) SingleClause() (yes bool, clause sat.Clause) {

	entries := make([]Entry, len(t.Entries))
	copy(entries, t.Entries)
	K := t.K

	if t.Typ == AtMost {
		K = -K
		for i, x := range entries {
			entries[i].Weight = int64(-1) * x.Weight
		}
	}

	// normalize to coefficients 1
	allOne := true
	clause.Literals = make([]sat.Literal, len(entries))

	for i, x := range entries {
		if x.Weight*x.Weight != 1 {
			allOne = false
			break
		}
		clause.Literals[i] = x.Literal

		if x.Weight == -1 {
			K += -x.Weight
			clause.Literals[i] = sat.Neg(clause.Literals[i])
		}
	}
	yes = allOne && K == 1
	return
}

func (t *Threshold) CreateSortingEncoding(typ sorters.SortingNetworkType) {

	total := t.Normalize()
	t.Print10()

	if total <= t.K {
		panic("sum of weights is too low to make a difference!")
	}
	if t.K == 0 {
		panic("Threshold is 0 with positive weights. All literals are facts!")
	}

	t.CreateBags()

	layers := make([]sorters.Sorter, len(t.Bags))
	bitsBag := len(t.Bags)

	for i, bag := range t.Bags {

		layers[i] = sorters.CreateSortingNetwork(len(bag), -1, typ)

		t.LitIn = append(t.LitIn, bag...) // this might have to be reversed
	}

	t.Sorter.In = make([]int, 0, len(t.LitIn))
	t.Sorter.Out = make([]int, 0, len(t.LitIn))

	offset := 2

	//fmt.Println("debug: layers", t.Bags)

	// determine the constant and what to add on both sides
	layerPow2 := int64(1 << uint(len(t.Bags)))

	tare := layerPow2 - ((t.K + 1) % layerPow2)
	tare = tare % layerPow2

	bTare := binary(tare)
	bitsTare := len(bTare)

	fmt.Println("debug: layerPow2", layerPow2)
	fmt.Println("debug: tare", tare)
	fmt.Println("debug: bitsBag", bitsBag, "bitsTare", bitsTare)
	fmt.Println("debug: bTare", bTare)

	// output of sorter in layer $i-1$
	bIn := make([]int, 0)

	finalMapping := make(map[int]int, len(t.Sorter.In))

	for i, layer := range layers {

		offset = layer.Normalize(offset, []int{})
		t.Sorter.Comparators = append(t.Sorter.Comparators, layer.Comparators...)

		//fmt.Println(i, "debug: bIn for this layer", bIn)

		//fmt.Println(i, "debug: layer", layer)

		t.Sorter.In = append(t.Sorter.In, layer.In...)

		size := len(bIn) + len(layers[i].In)

		//fmt.Println(i, "debug: size", size)

		mergeIn := make([]int, 0, size)
		mergeIn = append(mergeIn, bIn...)
		mergeIn = append(mergeIn, layer.Out...)

		//fmt.Println(i, "debug: merger preparation: size,cut", size, len(bIn))
		merger := sorters.CreateSortingNetwork(size, len(bIn), typ)
		offset = merger.Normalize(offset, mergeIn)
		//fmt.Println(i, "debug: mergeSorter", merger)

		// halving circuit:

		odd := 1

		if i < len(bTare) && bTare[i] == 1 {
			odd = 0
			bIn = make([]int, (len(merger.Out)+1)/2)
			//fmt.Println(i, "debug: lenMerger,tare i,odd", len(merger.Out), bTare[i], odd)
		} else {
			bIn = make([]int, len(merger.Out)/2)
			//fmt.Println(i, "debug: lenMerger,odd", len(merger.Out), odd)
		}

		// Alternate depending on bTare
		for j, x := range merger.Out {
			if j%2 == odd {
				bIn[j/2] = x
			} else if i < len(layers)-1 { // not in last layer, but else
				finalMapping[x] = -1
			}
		}

		//fmt.Println(i, "debug: merger", merger)

		t.Sorter.Comparators = append(t.Sorter.Comparators, merger.Comparators...)
		//fmt.Println(i, "debug: tSorter", t.Sorter)

	}

	// outLastLayer identifies the nth output in the last layer
	outLastLayer := ((t.K + 1 + tare) / int64(layerPow2)) - 1
	fmt.Println("debug: outLastLayer", outLastLayer)
	idSetToZero := bIn[outLastLayer]
	fmt.Println("which id is", idSetToZero)

	// and propagate the rest backwards
	setTo := -1
	for _, id := range t.Sorter.ComputeOut() {
		if id == idSetToZero {
			setTo = 0
		}
		if _, ok := finalMapping[id]; !ok {
			finalMapping[id] = setTo
		}
	}
	//fmt.Println("debug: finalMapping", finalMapping)

	t.Sorter.PropagateBackwards(finalMapping)

	t.Sorter.Normalize(2, []int{})

	//fmt.Println("final debug: tSorter", t.Sorter)
}

func (t *Threshold) IsNormalized() (yes bool) {
	yes = true

	for _, e := range t.Entries {
		if e.Weight <= 0 {
			yes = false
		}
	}
	return t.K > 0 && yes && t.Typ == AtMost
}

// Normalize: work in progress
// transform negative weights
// check if maximum reaches K at all
// sort by weight
// returns sum of weights
func (t *Threshold) Normalize() (total int64) {

	total = 0

	// remove 0 weights?
	if t.Typ == AtLeast {
		//set to AtMost, multiply by -1
		for i, e := range t.Entries {
			t.Entries[i].Weight = -e.Weight
		}
		t.K = -t.K
		t.Typ = AtMost
	} else if t.Typ == Equal {
		panic("Equal type for threshold function not supported")
	}

	for i, e := range t.Entries {
		if e.Weight == 0 {
			panic("Threshold contains a 0 weight element, should not occur")
		}
		if e.Weight < 0 {
			t.Entries[i].Weight = -e.Weight
			t.Entries[i].Literal = sat.Neg(e.Literal)
			t.K -= e.Weight
		}
		total += t.Entries[i].Weight
	}
	return
}

func (t *Threshold) CreateBags() {

	if !t.IsNormalized() {
		panic("Threshold is not normalized before creating bags")
	}

	nBags := len(binary(t.K))
	bins := make([][]int, len(t.Entries))
	bagPos := make([]int, nBags)
	bagSize := make([]int, nBags)

	maxWeight := int64(0)

	for i, e := range t.Entries {
		bins[i] = binary(e.Weight)

		for j, x := range bins[i] {
			bagSize[j] += x
		}

		if maxWeight < e.Weight {
			maxWeight = e.Weight
		}

	}

	t.Bags = make([][]sat.Literal, len(binary(maxWeight)))

	for i, _ := range t.Bags {
		t.Bags[i] = make([]sat.Literal, bagSize[i])
	}

	for i, e := range t.Entries {
		for j, x := range bins[i] {
			if x == 1 {
				t.Bags[j][bagPos[j]] = e.Literal
				bagPos[j]++
			}
		}
	}
}

func (t *Threshold) AddTare() {

}

// binary
// 23 = 10111
// special case if n==0 then return empty slice
// panic if n<0
func binary(n int64) (bin []int) {

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

func PrintBinary(n int64) {
	bin := binary(n)

	for i := len(bin) - 1; i >= 0; i-- {
		fmt.Print(bin[i])
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

		PrintBinary(x.Weight)

		if l.Sign {
			fmt.Print(" ")
		} else {
			fmt.Print(" ~")
		}
		//fmt.Print(l.Atom.P, "(", l.Atom.V1, ",", l.Atom.V2, ")")
		fmt.Print("x", l.Atom, " ")
	}
	switch t.Typ {
	case AtMost:
		fmt.Print(" <= ")
	case AtLeast:
		fmt.Print(" >= ")
	case Equal:
		fmt.Print(" == ")
	}

	PrintBinary(t.K)

	fmt.Println()
	fmt.Println()
}

func (t *Threshold) Print10() {
	fmt.Println(t.Desc)

	for _, x := range t.Entries {
		l := x.Literal

		if x.Weight > 0 {
			fmt.Printf("+")
		}

		fmt.Print(x.Weight, "")

		if l.Sign {
			fmt.Print(" ")
		} else {
			fmt.Print(" ~")
		}
		//fmt.Print(l.Atom.P, "(", l.Atom.V1, ",", l.Atom.V2, ")")
		fmt.Print("x", l.Atom, " ")
	}
	switch t.Typ {
	case AtMost:
		fmt.Print("<= ")
	case AtLeast:
		fmt.Print(">= ")
	case Equal:
		fmt.Print("== ")
	}
	fmt.Println(t.K, ";")

}
