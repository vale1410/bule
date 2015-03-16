package translation

import (
	//"fmt"
	"github.com/vale1410/bule/bdd"
	"github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/sat"
	"github.com/vale1410/bule/sorters"
	"github.com/vale1410/bule/sorting_network"
	"math"
	"strconv"
)

type TranslationType int // replace by a configuration

const (
	UNKNOWN TranslationType = iota
	Facts
	Clause
	AtMostOne
	ExactlyOne
	Cardinality
	ComplexBDD
	ComplexSN
	TranslationTypes
)

type ThresholdTranslation struct {
	Var     int // number of auxiliary variables introduced by this encoding
	Cls     int // number of clauses used
	Typ     TranslationType
	Clauses sat.ClauseSet
}

type CombinedTranslation struct {
	Var     int // number of auxiliary variables introduced by this encoding
	Cls     int // number of clauses used
	Typ     TranslationType
	Clauses sat.ClauseSet
}

func Categorize(pb *constraints.Threshold) (t ThresholdTranslation) {

	// this will become much more elaborate in the future

	t.Clauses.AddClauseSet(pb.Simplify())
	// per default all information is simplified and in form of facts
	t.Typ = Facts

	if len(pb.Entries) > 0 {

		if b, literals := pb.Cardinality(); b {

			if pb.K == int64(len(pb.Entries)-1) {
				switch pb.Typ {
				case constraints.AtMost:
					pb.Normalize(constraints.AtLeast, true)
					for i, x := range literals {
						literals[i] = sat.Neg(x)
					}
				case constraints.AtLeast:
					for i, x := range literals {
						literals[i] = sat.Neg(x)
					}
					pb.Normalize(constraints.AtMost, true)
				case constraints.Equal:
					for i, x := range literals {
						literals[i] = sat.Neg(x)
					}
					pb.Multiply(-1)
					pb.NormalizePositiveCoefficients()
				}
			}

			if pb.K == 1 {
				switch pb.Typ {
				case constraints.AtMost: // AMO
					t.Clauses.AddClauseSet(constraints.AtMostOne(constraints.Heule, "pb-HeuleAMO", literals))
					t.Typ = AtMostOne
				case constraints.AtLeast: // its a clause!
					t.Clauses.AddTaggedClause("pb->Cls", literals...)
					t.Typ = Clause
				case constraints.Equal: // Ex1
					t.Clauses.AddClauseSet(constraints.ExactlyOne(constraints.Heule, "pb-HeuleEX1", literals))
					t.Typ = ExactlyOne
				}
			} else {
				var s string
				var typ sorters.EquationType
				sx := strconv.Itoa(int(pb.K)) + "\\" + strconv.Itoa(len(pb.Entries))
				switch pb.Typ {
				case constraints.AtMost:
					sat.SetUp(4, sorters.Pairwise)
					typ = sorters.AtMost
					s = "pb<SN" + sx
				case constraints.AtLeast:
					sat.SetUp(4, sorters.Pairwise)
					typ = sorters.AtLeast
					s = "pb>SN" + sx
				case constraints.Equal:
					sat.SetUp(4, sorters.Pairwise)
					s = "pb=SN" + sx
					typ = sorters.Equal
				}
				t.Clauses.AddClauseSet(sat.CreateCardinality(s, literals, int(pb.K), typ))
				t.Cls = t.Clauses.Size()
				t.Typ = Cardinality
			}

		} else {
			// treat equality as two constraints!
			if pb.Typ == constraints.Equal {
				//fmt.Println("decompose in >= amd <=")
				pb.Typ = constraints.AtMost
				t = TranslateComplexThreshold(pb)
				pb.Normalize(constraints.AtMost, true)
				pb.Typ = constraints.AtLeast
				tt := TranslateComplexThreshold(pb)
				pb.Typ = constraints.Equal
				t.Var += tt.Var
				t.Cls += tt.Cls
				t.Clauses.AddClauseSet(t.Clauses)
			} else {
				t = TranslateComplexThreshold(pb)
			}
		}
	}
	return
}

func TranslateComplexThreshold(pb *constraints.Threshold) (t ThresholdTranslation) {
	tSN := TranslateBySN(pb)
	tBDD := TranslateByBDD(pb)
	//	fmt.Println("Complex, SN:", tSN.Cls, " BDD:", tBDD.Cls)
	if tBDD.Cls < tSN.Cls {
		t.Clauses = tBDD.Clauses
		t.Typ = ComplexBDD
	} else {
		t.Clauses = tSN.Clauses
		t.Typ = ComplexSN
	}
	return
}

func TranslateByBDDandEX1(pb *constraints.Threshold, ex1 []sat.Literal) (t ThresholdTranslation) {
	// check for overlap of variables
	// just do a rewrite, and call translateByBDDandAMO, reuse variables
	return
}

func TranslateByBDDandAMO(pb *constraints.Threshold, literals []sat.Literal) (t ThresholdTranslation) {
	// check for overlap of variables
	// just do a rewrite, and call translateByBDDandAMO, reuse variables
	return
}

func TranslateBySNandAMO(pb *constraints.Threshold, literals []sat.Literal) (t ThresholdTranslation) {
	// check for overlap of variables
	// just do a rewrite, and call translateByBDDandAMO, reuse variables
	return

}

func TranslateBySN(pb *constraints.Threshold) (t ThresholdTranslation) {
	pb.Normalize(constraints.AtMost, true)
	pb.Sort()
	sn := sorting_network.NewSortingNetwork(*pb)
	sn.CreateSorter()
	//sorting_network.PrintThresholdTikZ("sn.tex", []sorting_network.SortingNetwork{sn})
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

func TranslateByBDD(pb *constraints.Threshold) (t ThresholdTranslation) {
	pb.Normalize(constraints.AtMost, true)
	pb.Sort()
	// maybe do some sorting or such kinds?
	b := bdd.Init(len(pb.Entries), 300000) //space-out for nodes for one BDD construction
	topId, _, _, err := b.CreateBdd(pb.K, pb.Entries)
	if err != nil {
		//fmt.Println(err.Error())
		t.Cls = math.MaxInt32
	} else {
		t.Clauses = convertBDDIntoClauses(pb, topId, b)
		t.Cls = t.Clauses.Size()
	}
	return
}

// include some type of configuration
// optimize to remove 1 and 0 nodes in each level
func convertBDDIntoClauses(pb *constraints.Threshold, id int, b bdd.BddStore) (clauses sat.ClauseSet) {

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
