package constraints

import (
	"fmt"
	"github.com/vale1410/bule/config"
	"github.com/vale1410/bule/sat"
	"math"
)

type TranslationType int // replace by a configuration

const (
	UNKNOWN TranslationType = iota
	Facts
	Clause
	AtMostOne
	ExactlyOne
	Cardinality
	ComplexMDD
	ComplexSN
	ComplexMDDChain
	ComplexSNChain
	TranslationTypes
)

type ThresholdTranslation struct {
	PB      *Threshold
	Var     int // number of auxiliary variables introduced by this encoding
	Cls     int // number of clauses used
	Typ     TranslationType
	Clauses sat.ClauseSet
}

func group(pb []*Threshold) {
}

func Translate(pb *Threshold) (t ThresholdTranslation) {

	// this will become much more elaborate in the future

	t.Clauses.AddClauseSet(pb.Simplify())
	// per default all information is simplified and in form of facts
	t.Typ = Facts
	t.PB = pb

	if len(pb.Entries) > 0 {

		if b, literals := pb.Cardinality(); b {

			if pb.K == int64(len(pb.Entries)-1) {
				switch pb.Typ {
				case AtMost:
					pb.Normalize(AtLeast, true)
					for i, x := range literals {
						literals[i] = sat.Neg(x)
					}
				case AtLeast:
					for i, x := range literals {
						literals[i] = sat.Neg(x)
					}
					pb.Normalize(AtMost, true)
				case Equal:
					for i, x := range literals {
						literals[i] = sat.Neg(x)
					}
					pb.Multiply(-1)
					pb.NormalizePositiveCoefficients()
				}
			}

			if pb.K == 1 {
				switch pb.Typ {
				case AtMost: // AMO
					trans := TranslateAtMostOne(Heule, "HeuleAMO", literals)
					t.Clauses.AddClauseSet(trans.Clauses)
					t.Typ = AtMostOne
				case AtLeast: // its a clause!
					t.Clauses.AddTaggedClause("pb->Cls", literals...)
					t.Typ = Clause
				case Equal: // Ex1
					trans := TranslateExactlyOne(Heule, "HeuleEX1", literals)
					t.Clauses.AddClauseSet(trans.Clauses)
					t.Typ = ExactlyOne
				}
			} else {
				t.Clauses.AddClauseSet(CreateCardinality(literals, int(pb.K), pb.Typ))
				t.Cls = t.Clauses.Size()
				t.Typ = Cardinality
			}

		} else {
			// treat equality as two constraints!
			if pb.Typ == Equal {
				//fmt.Println("decompose in >= amd <=")
				pb.Typ = AtMost
				t = TranslateComplexThreshold(pb)
				pb.Normalize(AtMost, true)
				pb.Typ = AtLeast
				tt := TranslateComplexThreshold(pb)
				pb.Typ = Equal
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

func TranslateComplexThreshold(pb *Threshold) (t ThresholdTranslation) {

	switch config.Complex_flag {
	case "mdd":
		t = TranslateByMDD(pb)
	case "sn":
		t = TranslateBySN(pb)
	case "hybrid":
		tSN := TranslateBySN(pb)
		tMDD := TranslateByMDD(pb)

		//	fmt.Println("Complex, SN:", tSN.Cls, " mdd:", tMDD.Cls)

		if tMDD.Cls < tSN.Cls {
			t.Clauses = tMDD.Clauses
			t.Typ = ComplexMDD
		} else {
			t.Clauses = tSN.Clauses
			t.Typ = ComplexSN
		}
	default:
		panic("Complex_flag option not available: " + config.Complex_flag)
	}
	return
}

// returns if preprocessing was successful
// returns if it cant do the preprocessing
func PreprocessExactly(pb1 *Threshold, pb2 *Threshold) bool {

	//assumptions:
	//check for correct property of pb2
	//check for overlap of literals
	//both pb1 and pb2 are sorted in variable ordering!

	if pb2.Typ == Equal {
		b, _ := pb2.Cardinality()
		if !b {
			return false
		}
	}

	pb1.SortVar()
	pb2.SortVar()
	pb1.NormalizePositiveCoefficients()

	//pb1 is positiveCoefficients
	//pb2 is an exactly1 where all coefficients are 1

	//find min coefficient, to subtract
	pos := make([]int, len(pb2.Entries))
	mw := int64(math.MaxInt64) // min weight

	//position of current entry in pb2
	j := 0
	for i, x := range pb1.Entries {
		if j == len(pos) {
			break
		}
		if x.Literal == pb2.Entries[j].Literal {
			if x.Weight < mw {
				mw = x.Weight
			}
			pos[j] = i
			j++
		}
	}

	if j != len(pos) {
		return false
	}

	//fmt.Printf("%#v %#v \n", mw, pos)

	for _, i := range pos {
		pb1.Entries[i].Weight -= mw
	}
	pb1.K -= mw
	pb1.RemoveZeros()

	return true
}

// returns if preprocessing was successful
// Uses the translation of pb2 (count translation) and rewrites pb1
func TranslatePBwithAMO(pb *Threshold, amo CardTranslation) (t ThresholdTranslation) {

	//assumptions:
	//check for correct property of pb2
	//check for overlap of literals
	//both pb1 and pb2 are sorted in variable ordering!

	fmt.Println(amo)

	b, es := CommonSlice(pb.Entries, amo.Aux)

	if !b {
		panic("Check if amo fits  with the pb1")
	}

	last := int64(0)
	for i, e := range es {
		es[i].Weight = e.Weight - last
		last = e.Weight
	}

	pb.RemoveZeros()
	t.PB = pb

	// mdd := createMDDChain(pb1, amoLiterals)
	// translateByMDDChain(pb1, amoLiterals)

	return
}

func TranslateByMDDChain(pb *Threshold, literals []sat.Literal) (t ThresholdTranslation) {
	// check for overlap of variables
	// just do a rewrite, and call translateByMDD, reuse variables
	return
}

func TranslateBySNChain(pb *Threshold, literals []sat.Literal) (t ThresholdTranslation) {
	// check for overlap of variables
	// just do a rewrite, and call translateByMDD, reuse variables
	return
}
