package constraints

import (
	"fmt"
	"github.com/vale1410/bule/sat"
	"github.com/vale1410/bule/sorters"
	"strconv"
)

var sorterClauses int
var sorterType sorters.SortingNetworkType

// which sets the type of clauses translated from sorting networks
// see below CreateEncoding for ids wrt clauses
// 0: false, false, false, true, true, true, false, false
// 1: false, false, false, true, true, true, false, true
// 2: false, true, true, true, true, true, true, false
// 3: false, true, true, true, true, true, true, true
func SetUpSorterTranslation(which int, typ sorters.SortingNetworkType) {
	sorterClauses = which
	sorterType = typ
}

// CreateCardinality takes set of literals and creates a sorting network
// encoding.
func CreateCardinality(input []sat.Literal, K int, typ EquationType) sat.ClauseSet {

	sx := strconv.Itoa(K) + "\\" + strconv.Itoa(len(input))
	var s string
	var sorterEqTyp sorters.EquationType

	switch typ {
	case AtMost:
		SetUpSorterTranslation(4, sorters.Pairwise)
		sorterEqTyp = sorters.AtMost
		s = "pb<SN" + sx
	case AtLeast:
		SetUpSorterTranslation(4, sorters.Pairwise)
		sorterEqTyp = sorters.AtLeast
		s = "pb>SN" + sx
	case Equal:
		SetUpSorterTranslation(4, sorters.Pairwise)
		s = "pb=SN" + sx
		sorterEqTyp = sorters.Equal
	}

	var which [8]bool

	switch sorterClauses {
	case 1:
		which = [8]bool{false, false, false, true, true, true, false, false}
	case 2:
		which = [8]bool{false, false, false, true, true, true, false, true}
	case 3:
		which = [8]bool{false, true, true, true, true, true, true, false}
	case 4:
		which = [8]bool{false, true, true, true, true, true, true, true}
	default:
		panic("sorterClauses in sat module not set")
	}

	sorter := sorters.CreateCardinalityNetwork(len(input), K, sorterEqTyp, sorterType)

	sorter.RemoveOutput()

	pred := sat.Pred("sort" + strconv.Itoa(uniqueId))

	output := make([]sat.Literal, 0)

	return CreateEncoding(input, which, output, s, pred, sorter)
}

// Create Encoding for Sorting Network
// 0)  Omitted for clarity (ids as in paper)
// 1)  A or -D
// 2)  B or -D
// 3) -A or -B or D
// 4) -A or  C
// 5) -B or  C
// 6)  A or  B or -C
// 7)  C or -D
// -1,0,1 = *, false, true
func CreateEncoding(input []sat.Literal, which [8]bool, output []sat.Literal, tag string, pred sat.Pred, sorter sorters.Sorter) (cs sat.ClauseSet) {

	//	sorters.PrintSorterTikZ(sorter, "sorter1.tex")

	//cs.list = make([]Clause, 0, 7*len(sorter.Comparators))

	backup := make(map[int]sat.Literal, len(sorter.Out)+len(sorter.In))

	for i, x := range sorter.In {
		backup[x] = input[i]
	}

	for i, x := range sorter.Out {
		backup[x] = output[i]
	}

	for _, comp := range sorter.Comparators {

		if comp.D == 1 || comp.C == 0 {
			fmt.Println("something is wrong with the comparator", comp)
			panic("something is wrong with the comparator")
		}

		getLit := func(x int) sat.Literal {
			if lit, ok := backup[x]; ok {
				return lit
			} else {
				return sat.Literal{true, sat.NewAtomP1(pred, x)}
			}
		}

		a := getLit(comp.A)
		b := getLit(comp.B)
		c := getLit(comp.C)
		d := getLit(comp.D)

		if comp.C == 1 { // 6) A or B
			//if which[6] {
			cs.AddTaggedClause(tag, a, b)
			//}
		} else if comp.C > 0 { // 4) 5) 6)
			//4)
			if which[4] {
				cs.AddTaggedClause(tag, sat.Neg(a), c)
			}
			//5)
			if which[5] {
				cs.AddTaggedClause(tag, sat.Neg(b), c)
			}
			//6)
			if which[6] {
				cs.AddTaggedClause(tag, a, b, sat.Neg(c))
			}
		}
		if comp.D == 0 { //3)
			//if which[3] {
			cs.AddTaggedClause(tag, sat.Neg(a), sat.Neg(b))
			//}
		} else if comp.D > 0 { // 1) 2) 3)
			//1)
			if which[1] {
				cs.AddTaggedClause(tag, a, sat.Neg(d))
			}
			//2)
			if which[2] {
				cs.AddTaggedClause(tag, b, sat.Neg(d))
			}
			//3)
			if which[3] {
				cs.AddTaggedClause(tag, sat.Neg(a), sat.Neg(b), d)
			}
		}

		if which[7] && comp.C != 1 && comp.D != 0 && comp.C != -1 && comp.D != -1 { // 7)

			if comp.C == 0 || comp.D == 1 {
				panic("something is wrong with this comparator")
			}
			cs.AddTaggedClause(tag, c, sat.Neg(d))
		}
	}
	return
}
