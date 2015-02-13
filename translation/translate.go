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
	//	PB        constraints.Threshold
	Trans   TranslationType
	Clauses sat.ClauseSet
}

func Translate(PB constraints.Threshold, typ TranslationType) (t ThresholdTranslation) {

	//	t.PB = PB

	//this will become much more elaborate in the future
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
			id := 0 // TODO

			pred := sat.Pred("auxSN" + strconv.Itoa(id))
			//fmt.Println("sorter", sn.Sorter)
			t.Clauses = sat.CreateEncoding(sn.PB.LitIn, which, []sat.Literal{}, "BnB", pred, sn.Sorter)
		} else { // BDD

			// maybe do some sorting or such kinds?

			bdd.Translate(PB)
		}
	}
	return
}
