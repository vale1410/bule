package threshold

import (
	"github.com/vale1410/bule/sat"
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
	Literal sat.Literal
	Weight  int64
}

type Threshold struct {
	Desc    string
	Entries []Entry
	K       int64
	Tare    int64
	Typ     EquationType
	Pred    sat.Pred
	Clauses sat.ClauseSet
	Trans   TranslationType
	Sorter  sorters.Sorter
	Bags    [][]sat.Literal
	LitIn   []sat.Literal //Bags flattened, input to Sorter
}

type TranslationType int

const (
	Facts TranslationType = iota
	SingleClause
	SortingNetwork
	BDD
)

func (t *Threshold) Translate() {

	if b, cls := t.OnlyFacts(); b {
		//fmt.Println("Bule: translate by facts", len(cls))
		t.Clauses = cls
		t.Trans = Facts
	} else if b, literals := t.SingleClause(); b {
		//fmt.Println("Bule: translate by single clause", len(cls))
		t.Clauses.AddTaggedClause("SC", literals...)
		t.Trans = SingleClause
	} else {
		//fmt.Println("Bule: translate by sorting network")
		typ := sorters.OddEven
		t.Trans = SortingNetwork
		t.CreateSorter(typ)
	}

}

func (t *Threshold) OnlyFacts() (is bool, cs sat.ClauseSet) {

	t.NormalizeAtMost()
	is = false

	if t.K <= 0 {
		is = true
		cs = sat.NewClauseSet(len(t.Entries))
		for _, x := range t.Entries {
			cs.AddTaggedClause("OF", sat.Neg(x.Literal))
		}
	}

	return
}

func (t *Threshold) SingleClause() (is bool, literals []sat.Literal) {

	t.NormalizeAtLeast(false)

	var clause sat.Clause

	entries := make([]Entry, len(t.Entries))
	copy(entries, t.Entries)
	K := t.K

	// normalize to coefficients 1
	allOne := true
	literals = make([]sat.Literal, len(entries))

	for i, x := range entries {
		if x.Weight*x.Weight != 1 {
			allOne = false
			break
		}
		literals[i] = x.Literal

		if x.Weight == -1 {
			K += -x.Weight
			literals[i] = sat.Neg(clause.Literals[i])
		}
	}

	return allOne && K == 1, literals
}

func (t *Threshold) CreateSorter(typ sorters.SortingNetworkType) {

	total := t.NormalizeAtMost()

	//t.Print10()

	if total <= t.K {
		panic("sum of weights is too low to make a difference!")
	}
	if t.K == 0 {
		panic("Threshold is 0 with positive weights. All literals are facts!")
	}

	t.CreateBags()

	layers := make([]sorters.Sorter, len(t.Bags))

	for i, bag := range t.Bags {

		layers[i] = sorters.CreateSortingNetwork(len(bag), -1, typ)

		t.LitIn = append(t.LitIn, bag...) // this might have to be reversed
	}

	t.Sorter.In = make([]int, 0, len(t.LitIn))
	t.Sorter.Out = make([]int, 0, len(t.LitIn))

	offset := 2

	// determine the constant and what to add on both sides
	layerPow2 := int64(1 << uint(len(t.Bags)))

	tare := layerPow2 - ((t.K + 1) % layerPow2)
	tare = tare % layerPow2
	t.Tare = tare
	bTare := binary(tare)

	// output of sorter in layer $i-1$
	bIn := make([]int, 0)

	finalMapping := make(map[int]int, len(t.Sorter.In))

	for i, layer := range layers {

		offset = layer.Normalize(offset, []int{})
		t.Sorter.Comparators = append(t.Sorter.Comparators, layer.Comparators...)

		t.Sorter.In = append(t.Sorter.In, layer.In...)

		size := len(bIn) + len(layers[i].In)

		mergeIn := make([]int, 0, size)
		mergeIn = append(mergeIn, bIn...)
		mergeIn = append(mergeIn, layer.Out...)

		merger := sorters.CreateSortingNetwork(size, len(bIn), typ)
		offset = merger.Normalize(offset, mergeIn)

		// halving circuit:

		odd := 1

		if i < len(bTare) && bTare[i] == 1 {
			odd = 0
			bIn = make([]int, (len(merger.Out)+1)/2)
		} else {
			bIn = make([]int, len(merger.Out)/2)
		}

		// Alternate depending on bTare
		for j, x := range merger.Out {
			if j%2 == odd {
				bIn[j/2] = x
			} else if i < len(layers)-1 { // not in last layer, but else
				finalMapping[x] = -1
			}
		}

		t.Sorter.Comparators = append(t.Sorter.Comparators, merger.Comparators...)

	}

	// outLastLayer identifies the nth output in the last layer
	outLastLayer := ((t.K + 1 + tare) / int64(layerPow2)) - 1
	idSetToZero := bIn[outLastLayer]

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

// Normalize: AtLeast
// transform to AtLeast, only positive variables
// b=true: only positive variables
// b=false: only positive weights
func (t *Threshold) NormalizeAtLeast(posVariables bool) {
	posWeights := !posVariables

	// remove 0 weights?
	if t.Typ == AtMost {
		//set to AtMost, multiply by -1
		t.K = -t.K
		t.Typ = AtLeast
		for i, e := range t.Entries {
			t.Entries[i].Weight = -e.Weight
		}
	} else if t.Typ == Equal {
		panic("Equal type for threshold function not supported yet")
	}

	for i, e := range t.Entries {
		if (posWeights && t.Entries[i].Weight < 0) ||
			(posVariables && e.Literal.Sign == false) {
			t.Entries[i].Literal = sat.Neg(e.Literal)
			t.K -= t.Entries[i].Weight
			t.Entries[i].Weight = -t.Entries[i].Weight
		}
	}

	return
}

// Normalize: AtMost
// transform to AtMost, only positive weights
// transform negative weights
// check if maximum reaches K at all
// todo: sort by weight
// returns sum of weights
func (t *Threshold) NormalizeAtMost() (total int64) {

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
		t.Print10()
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
