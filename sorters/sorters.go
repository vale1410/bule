package sorters

import (
	"log"
)

const (
	OddEven SortingNetworkType = iota
	Bitonic
	Bubble
	Pairwise
	ShellSort
	Insertion
)

func (s SortingNetworkType) String() string {
	switch s {
	case OddEven:
		return "OddEven"
	case Bitonic:
		return "Bitonic"
	case Bubble:
		return "Bubble"
	case Pairwise:
		return "Pairwise"
	case ShellSort:
		return "ShellSort"
	case Insertion:
		return "Insertion"
	default:
		panic("SortingNetworkType not existing")
	}
	return ""
}

const (
	AtMost CardinalityType = iota
	AtLeast
	Equal
)

type SortingNetworkType int
type CardinalityType int

// The slice of comparators must be in correct order,
// meaning that the comparator with input A and B must
// occur after the comparator with output A and B.
// expection are 0 and 1, that define true and false. They may only occur as
// output of a comparator (C or D), but not as input (as will be propagated
// and the comparator removed).
type Sorter struct {
	Comparators []Comparator
	In          []int
	Out         []int
}

// Ids for the connections (A,B,C,D) start at 2 and are incremented.
// Id 0 and 1 are reserved for true and false respectively
// Id -1 means dont care
// B --|-- D = A && B
//     |
// A --|-- C = A || B
type Comparator struct {
	A, B, C, D int
}

func CreateCardinalityNetwork(size int, k int, cType CardinalityType, sType SortingNetworkType) (sorter Sorter) {

	mapping := make(map[int]int, size)

	sorter = CreateSortingNetwork(size, -1, sType)

	switch cType {
	case AtMost:
		for i := k; i < size; i++ {
			mapping[sorter.Out[i]] = 0
			sorter.Out[i] = 0
		}
		sorter.PropagateBackwards(mapping)
		sorter.Out = sorter.Out[:k]
	case AtLeast:
		for i := 0; i < k; i++ {
			mapping[sorter.Out[i]] = 1
			sorter.Out[i] = 1
		}
		sorter.PropagateBackwards(mapping)
		sorter.Out = sorter.Out[k:]
	case Equal:
		for i := k; i < size; i++ {
			mapping[sorter.Out[i]] = 0
			sorter.Out[i] = 0
		}
		for i := 0; i < k; i++ {
			mapping[sorter.Out[i]] = 1
			sorter.Out[i] = 1
		}
		sorter.PropagateBackwards(mapping)
		sorter.Out = nil
	default:
		log.Panic("Cardnality Not implemented yet")
	}
	return
}

// CreateSortingNetworks creates a sorting network of arbitrary size, cut and type
func CreateSortingNetwork(s int, cut int, typ SortingNetworkType) (sorter Sorter) {

	//grow to be 2^n
	n := 1
	for n < s {
		n *= 2
	}

	comparators := make([]Comparator, 0)
	output := make([]int, n)

	offset := 2

	for i, _ := range output {
		output[i] = i + offset
	}
	input := make([]int, n)
	copy(input, output)

	newId := n + offset

	switch typ {
	case OddEven:
		oddevenSort(&newId, output, &comparators, 0, n-1)
	case Bitonic:
		triangleBitonic(&newId, output, &comparators, 0, n-1)
	case Bubble:
		bubbleSort(&newId, output, &comparators)
	case Pairwise:
		pairwiseSort(&newId, output, &comparators, 0, n-1)
	default:
		log.Println(typ)
		log.Panic("Type of sorting network not implemented yet")
	}

	sorter = Sorter{comparators, input, output}
    log.Println(sorter)
	sorter.changeSize(s)
	sorter.PropagateOrdering(cut)
    log.Println(sorter)

	return
}

// RemoveOutput()
// Treats all Output Ids as DontCare and propagates backwards
// This only makes sense for CardinalityNetworks
// sets -1 at a comparator if output is dont care
func (sorter *Sorter) RemoveOutput() {

	mapping := make(map[int]int, len(sorter.Out))

	for i, x := range sorter.Out {
		mapping[x] = -1
		sorter.Out[i] = -1
	}

	sorter.Out = nil
	sorter.PropagateBackwards(mapping)
}

// Normalize
// renames the ids in the comparators from 1 to |2*comparator|
// If array in is empty, replaces ids by offset, offset+1 ...
// Otherwise use ides in In array for renaming, and only replace all other ids
// Then Replaces In array with argument
// Then renames new ids starting with offset
// Returns last offset + 1
// All Ids with -1,0,1 in C,D of a comparator are ignored and not renamed
func (s *Sorter) Normalize(offset int, in []int) (maxId int) {

	mapping := make(map[int]int, len(s.In)+2*len(s.Comparators))

	if len(in) == 0 {
		in = make([]int, len(s.In))
		for i, _ := range in {
			in[i] = offset
			offset++
		}
	}

	if len(s.In) != len(in) {
		log.Panic("Input vector (1)  and size of sorter (2)  differ:", len(in), len(s.In))
	}

	if offset < 2 {
		log.Panic("Offset has to be at least 2 in sorter.Normalize")
	}

	for i, id := range s.In {
		mapping[id] = in[i]
		s.In[i] = in[i]
	}

	for i, comp := range s.Comparators {

		a, aok := mapping[comp.A]
		b, bok := mapping[comp.B]
		_, cok := mapping[comp.C]
		_, dok := mapping[comp.D]

		if !aok || !bok || cok || dok {
			log.Panic("Normalize: 1 cannot rename ids ", comp)
		}

		s.Comparators[i].A = a
		s.Comparators[i].B = b

		if comp.C > 1 {
			mapping[comp.C] = offset
			s.Comparators[i].C = offset
			offset++
		}

		if comp.D > 1 {
			mapping[comp.D] = offset
			s.Comparators[i].D = offset
			offset++
		}
	}

	for i, id := range s.Out {
		a, aok := mapping[id]

		if !aok {
			log.Panic("Normalize: 2 cannot rename ids ", id, mapping[id])
		}

		s.Out[i] = a
	}

	return offset
}

// PropagateOrdering
// cut >= 0, cut=-1 means no cut
// from 0..cut-1 sorted and from cut .. length-1 sorted
// propagated and remove comparators
func (sorter *Sorter) PropagateOrdering(cut int) {

	if cut < 0 { // signal for no cut
		return
	} else if cut == 0 || cut == len(sorter.In) {
		// the stuff is already sorted, remove comparators
		sorter.Comparators = []Comparator{}
		copy(sorter.Out, sorter.In)
	} else {

		mapping := make(map[int]int, len(sorter.Comparators))
		location := make(map[int]bool, len(sorter.Comparators))

		for i, x := range sorter.In {
			mapping[x] = x
			location[x] = i >= cut
		}

		nRemove := 0
		comparators := sorter.Comparators

		zero := Comparator{0, 0, 0, 0}

		for i, comp := range comparators {

			a, aok := mapping[comp.A]
			b, bok := mapping[comp.B]

			if aok {
				comparators[i].A = a
			} else {
				a = comp.A
			}

			if bok {
				comparators[i].B = b
			} else {
				b = comp.B
			}

			la, laok := location[comp.A]
			lb, lbok := location[comp.B]

			if laok && lbok && la == lb {
				// we have an already sorted input
				mapping[comp.C] = a
				mapping[comp.D] = b
				location[comp.C] = la
				location[comp.D] = lb
				comparators[i] = zero
				nRemove++
			}
		}
		//remove zeros and then return comparators
		out := make([]Comparator, 0, len(sorter.Comparators)-nRemove)

		for _, comp := range comparators {
			if comp != zero {
				out = append(out, comp)
			}
		}
		sorter.Comparators = out
		return
	}
}

// ChangeSize shrinks the sorter to a size s
func (sorter *Sorter) changeSize(s int) {

	n := len(sorter.In)

	mapping := make(map[int]int, n-s)

	for i := s; i < n; i++ {
		//setting the top n-s elements to zero
		mapping[sorter.In[i]] = 0
	}

	sorter.PropagateForward(mapping)

	//potential check for s..n being 0

	for i, x := range sorter.Out {
		if r, ok := mapping[x]; ok {
			sorter.Out[i] = r
		}
	}

	sorter.In = sorter.In[:s]
	sorter.Out = sorter.Out[:s]

	return
}

func (sorter *Sorter) PropagateForward(mapping map[int]int) {

	l := 0
	comparators := sorter.Comparators

	// remove is a comparator that marks removal

	remove := Comparator{0, 0, 0, 0}

    log.Println(mapping)

	for i, comp := range comparators {

		a, aok := mapping[comp.A]
		b, bok := mapping[comp.B]

		if aok {
			comparators[i].A = a
		} else {
			a = comp.A
		}

		if bok {
			comparators[i].B = b
		} else {
			b = comp.B
		}

		removed := false

		if a == 0 {
			mapping[comp.D] = 0
			mapping[comp.C] = b
			removed = true
		}

		if b == 0 {
			mapping[comp.D] = 0
			mapping[comp.C] = a
			removed = true
		}

		if a == 1 {
			mapping[comp.C] = 1
			mapping[comp.D] = b
			removed = true
		}

		if b == 1 {
			mapping[comp.C] = 1
			mapping[comp.D] = a
			removed = true
		}

		if a == 0 && b == 0 {
			mapping[comp.C] = 0
			removed = true
		}

		if a == 1 && b == 1 {
			mapping[comp.D] = 1
			removed = true
		}

		if removed {
			l++
			comparators[i] = remove
			removed = false
		}
	}

	//remove the unused comparators
	out := make([]Comparator, 0, l)
	for _, comp := range comparators {
		if comp != remove {
			out = append(out, comp)
		}
	}

    log.Println(mapping)

	sorter.Comparators = out
}

// determines the ids of the out-vector
func (sorter *Sorter) ComputeOut() (out []int) {

	mapping := make(map[int]int, len(sorter.In))
	out = make([]int, len(sorter.In))

	for i, id := range sorter.In {
		mapping[id] = i
		out[i] = id
	}

	for _, comp := range sorter.Comparators {

		mapping[comp.C] = mapping[comp.A]
		mapping[comp.D] = mapping[comp.B]
		out[mapping[comp.A]] = comp.C
		out[mapping[comp.B]] = comp.D
	}

	return
}

func (sorter *Sorter) PropagateBackwards(mapping map[int]int) {

	l := 0
	comparators := sorter.Comparators
	remove := Comparator{0, 0, 0, 0}

	cleanMapping := make(map[int]int, 0)

	mapping[-1] = -1
	//mapping[0] = -1 //why?
	//mapping[1] = -1 //why?

	for i := len(comparators) - 1; i >= 0; i-- {

		comp := comparators[i]

		removed := false

		valueC, okC := mapping[comp.C]
		valueD, okD := mapping[comp.D]

		if okC && valueC == 0 {
			mapping[comp.A] = 0
			mapping[comp.B] = 0
			cleanMapping[comp.D] = 0
			removed = true
		}

		if okD && valueD == 1 {
			mapping[comp.A] = 1
			mapping[comp.B] = 1
			cleanMapping[comp.C] = 1
			removed = true
		}

		if okC && okD && valueC == -1 && valueD == -1 {
			mapping[comp.A] = -1
			mapping[comp.B] = -1
			removed = true
		}

		if removed {
			l++
			comparators[i] = remove
		}
	}

	//remove the unused comparators
	out := make([]Comparator, 0, l)
	for _, comp := range comparators {
		if comp != remove {
			if value, ok := mapping[comp.C]; ok {
				comp.C = value
			}

			if value, ok := mapping[comp.D]; ok {
				comp.D = value
			}
			out = append(out, comp)
		}
	}

	sorter.Comparators = out

	if len(cleanMapping) > 0 {
		sorter.PropagateForward(cleanMapping)
	}

}

// Functions for creating sorters
// used in the implementations of bitonic, oddeven, pairwise etc.
func compareAndSwap(newId *int, array []int, comparators *[]Comparator, i int, j int) {
	*newId += 2
	*comparators = append(*comparators, Comparator{array[i], array[j], *newId - 2, *newId - 1})

	array[i] = *newId - 2
	array[j] = *newId - 1
}
