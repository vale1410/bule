package translation

import (
	//	"fmt"
	"github.com/vale1410/bule/bdd"
	"github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/sat"
	"github.com/vale1410/bule/sorting_network"
	"strconv"
)

type TranslationType int // replace by a configuration

const (
	Facts TranslationType = iota
	Clause
	AtMostOne
	ExactlyOne
	Cardinality
	Complex
)

type ThresholdTranslation struct {
	Var     int
	Cls     int
	Trans   TranslationType
	Clauses sat.ClauseSet
}

func Categorize(pb constraints.Threshold) (t ThresholdTranslation) {

	// this will become much more elaborate in the future
	// several translation methods; heuristics on which one to use
	// different configurations, etc.

	if b, cls := pb.OnlyFacts(); b { // forced type
		//	fmt.Println(PB)
		//	fmt.Println("Bule: facts", cls.Size())
		t.Clauses = cls
		t.Trans = Facts
	} else if b, literals := pb.SingleClause(); b { // forced type
		//	fmt.Println("Bule: single clause", cls.Size())
		t.Clauses.AddTaggedClause("single", literals...)
		t.Trans = Clause
		//	} else if b, literals := pb.AtMostOne(); b {
		//		// isAtMostOne constraint
		//	} else if b, literals := pb.Cardinality(); b {
		//		// isCardinality constraint
	} else {
		t.Trans = Complex
	}
	return
}

func TranslateBySN(pb constraints.Threshold) (t ThresholdTranslation) {
	//	t.Trans = Complex
	//	pb.NormalizeAtMost()
	pb.Print10()
	sn := sorting_network.NewSortingNetwork(pb)
	sn.CreateSorter()
	wh := 2
	var which [8]bool

	switch wh {
	case 1:
		which = [8]bool{false, false, false, true, true, true, false, false}
	case 2:
		which = [8]bool{false, false, false, true, true, true, false, true}
	case 3:
		which = [8]bool{false, true, true, true, true, true, true, false}
	case 4:
		which = [8]bool{false, true, true, true, true, true, true, true}
	}

	pred := sat.Pred("auxSN_" + strconv.Itoa(pb.Id))
	t.Clauses = sat.CreateEncoding(sn.LitIn, which, []sat.Literal{}, "BnB", pred, sn.Sorter)
	t.Cls = t.Clauses.Size()
	return
}

func TranslateByBDD(pb constraints.Threshold) (t ThresholdTranslation) {
	pb.NormalizeAtMost()
	pb.Sort()
	// maybe do some sorting or such kinds?
	b := bdd.Init(len(pb.Entries))
	topId, _, _ := b.CreateBdd(pb.K, pb.Entries)
	t.Clauses = convertBDDIntoClauses(pb, topId, b)
	t.Cls = t.Clauses.Size()
	return
}

// include some type of configuration
// optimize to remove 1 and 0 nodes in each level
func convertBDDIntoClauses(pb constraints.Threshold, id int, b bdd.BddStore) (clauses sat.ClauseSet) {

	pred := sat.Pred("auxBDD_" + strconv.Itoa(pb.Id))

	top_lit := sat.Literal{true, sat.NewAtomP1(pred, id)}
	clauses.AddTaggedClause("Top", top_lit)
	for _, n := range b.Nodes {
		v_id, l, vds := b.ClauseIds(*n)
		//fmt.Println(v_id, l, vds)
		if !n.IsZero() && !n.IsOne() {

			v_lit := sat.Literal{false, sat.NewAtomP1(pred, v_id)}
			for i, vd_id := range vds {
				vd_lit := sat.Literal{true, sat.NewAtomP1(pred, vd_id)}
				if i > 0 {
					//if vd_id != 0 { // vd is not true
					clauses.AddTaggedClause("1B", v_lit, sat.Neg(pb.Entries[len(pb.Entries)-l].Literal), vd_lit)
					//} else {
					//	clauses.AddClause(sat.Neg(v_lit), sat.Neg(pb.Entries[len(pb.Entries)-l].Literal))
					//}
				} else {
					//if vd_id != 1 { // vd is not true
					clauses.AddTaggedClause("0B", v_lit, vd_lit)
					//}
				}
			}
		} else if n.IsZero() {
			v_lit := sat.Literal{false, sat.NewAtomP1(pred, v_id)}
			clauses.AddTaggedClause("False", v_lit)
		} else if n.IsOne() {
			v_lit := sat.Literal{true, sat.NewAtomP1(pred, v_id)}
			clauses.AddTaggedClause("True", v_lit)
		}

	}

	return
}
