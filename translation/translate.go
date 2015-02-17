package translation

import (
	"fmt"
	"github.com/vale1410/bule/bdd"
	"github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/sat"
	"github.com/vale1410/bule/sorting_network"
	"strconv"
)

type TranslationType int // replace by a configuration

const (
	Facts TranslationType = iota
	SingleClause
	SortingNetwork
	BDD
)

type ThresholdTranslation struct {
	Trans   TranslationType
	Clauses sat.ClauseSet
}

func Translate(PB constraints.Threshold, typ TranslationType) (t ThresholdTranslation) {

	// this will become much more elaborate in the future
	// several translation methods; heuristics on which one to use
	// different configurations, etc.

	if b, cls := PB.OnlyFacts(); b { // forced type
		fmt.Println("Bule: translate by facts", cls.Size())
		PB.Clauses = cls
		t.Trans = Facts
	} else if b, literals := PB.SingleClause(); b { // forced type
		fmt.Println("Bule: translate by single clause", cls.Size())
		PB.Clauses.AddTaggedClause("SC", literals...)
		t.Trans = SingleClause
	} else if false { // check for AtMostOne/ExacltyOne
	} else {

		if typ == SortingNetwork {

			fmt.Println("Bule: translate by sorting network")
			t.Trans = SortingNetwork
			sn := sorting_network.NewSortingNetwork(PB)
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

			pred := sat.Pred("auxSN_" + strconv.Itoa(PB.Id))
			//fmt.Println("sorter", sn.Sorter)
			t.Clauses = sat.CreateEncoding(sn.PB.LitIn, which, []sat.Literal{}, "BnB", pred, sn.Sorter)
			// should be in the translation package
		} else { // BDD

			fmt.Println("test")
			// maybe do some sorting or such kinds?
			b := bdd.Init(len(PB.Entries))
			topId, _, _ := b.CreateBdd(PB.K, PB.Entries)
			b.Debug(true)
			t.Clauses = convertBDDIntoClauses(PB, topId, b)
		}
	}
	return
}

// include some type of configuration
func convertSNIntoClauses(sn sorting_network.SortingNetwork) (clauses sat.ClauseSet) {

	return
}

// include some type of configuration
func convertBDDIntoClauses(pb constraints.Threshold, id int, b bdd.BddStore) (clauses sat.ClauseSet) {

	//	pred := sat.Pred("auxBDD_" + strconv.Itoa(pb.Id))

	fmt.Println("check")

	for _, n := range b.Nodes {
		if !n.IsZero() && !n.IsOne() {

			v, l, vds := b.ClauseIds(*n)
			for i, vd := range vds {
				if i > 0 {
					fmt.Println("-(", pb.Entries[len(pb.Entries)-l-1].Literal.ToTxt(), " >=  ", i, ")", -v, vd)
				} else {
					fmt.Println(-v, vd)
				}
			}
		}

	}

	return
}
