package translation

import (
	"github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/sat"
	"github.com/vale1410/bule/sorting_network"
)

type TranslationType int

const (
	Facts TranslationType = iota
	SingleClause
	SortingNetwork
	BDD
)

type ThresholdTranslation struct {
	PB    constraints.Threshold
	Trans TranslationType
}

func Translate(t ThresholdTranslation) {

	//this will become much more elaborate in the future
	// several translation methods; heurist on which one to use
	// different configurations, etc.

	if b, cls := t.PB.OnlyFacts(); b {
		//fmt.Println("Bule: translate by facts", len(cls))
		t.PB.Clauses = cls
		t.Trans = Facts
	} else if b, literals := t.PB.SingleClause(); b {
		//fmt.Println("Bule: translate by single clause", len(cls))
		t.PB.Clauses.AddTaggedClause("SC", literals...)
		t.Trans = SingleClause
	} else {
		//fmt.Println("Bule: translate by sorting network")
		t.Trans = SortingNetwork
		transation := sorting_network.NewTranslation(t.PB)
		transation.CreateSorter()
	}

}
