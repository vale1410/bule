package constraints

import (
	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/sat"
	"math"
)

type TranslationType int // replace by a configuration

const (
	UNKNOWN TranslationType = iota
	Facts
	Clause
	AMO
	EX1
	EXK
	CARD
	CMDD
	CSN
	TranslationTypes
	CMDDC
	CSNC
)

func (t TranslationType) String() (s string) {
	switch t {
	case Facts:
		s = "Fcts"
	case Clause:
		s = "Cls"
	case AMO:
		s = "AMO"
	case EX1:
		s = "EX1"
	case EXK:
		s = "EXK"
	case CARD:
		s = "Card"
	case CMDD:
		s = "CMDD"
	case CSN:
		s = "CSN"
	case TranslationTypes:
		s = "TranslationTypes"
	case CMDDC:
		s = "CMDDC"
	case CSNC:
		s = "CSNC"
	default:
		panic("has not been implemented")
	}
	return
}

type ThresholdTranslation struct {
	PB      *Threshold
	Typ     TranslationType
	Clauses sat.ClauseSet
	Chains  Chains
}

func Categorize1(pb *Threshold) (t ThresholdTranslation) {

	// per default all information that can be simplified will be in form of facts
	t.Clauses.AddClauseSet(pb.Simplify())

	t.Typ = Facts
	t.PB = pb
	if len(pb.Entries) == 0 {
		glob.D(pb.Id, "was simplified completely")
	} else {

		if b, literals := pb.Cardinality(); b {

			if pb.K == int64(len(pb.Entries)-1) {
				switch pb.Typ {
				case LE:
					pb.Normalize(GE, true)
					for i, x := range literals {
						literals[i] = sat.Neg(x)
					}
				case GE:
					for i, x := range literals {
						literals[i] = sat.Neg(x)
					}
					pb.Normalize(LE, true)
				case EQ:
					for i, x := range literals {
						literals[i] = sat.Neg(x)
					}
					pb.Multiply(-1)
					pb.NormalizePositiveCoefficients()
				}
			}

			if pb.K == 1 {
				switch pb.Typ {
				case LE: // AMO
					trans := TranslateAtMostOne(Heule, "H_AMO", literals)
					t.Clauses.AddClauseSet(trans.Clauses)
					t.Typ = AMO
				case GE: // its a clause!
					t.Clauses.AddTaggedClause("Cls", literals...)
					t.Typ = Clause
				case EQ: // Ex1
					trans := TranslateExactlyOne(Heule, "H_EX1", literals)
					t.Clauses.AddClauseSet(trans.Clauses)
					t.Typ = EX1
				}
			} else {
				t.Clauses.AddClauseSet(CreateCardinality(pb))
				t.Typ = CARD
			}

		} else {
			// treat equality as two constraints!
			if pb.Typ == EQ {
				glob.D(pb.Id, " decompose in >= amd <=")
				pb.Typ = LE
				t = TranslateComplexThreshold(pb)
				pb.Normalize(LE, true)
				pb.Typ = GE
				tt := TranslateComplexThreshold(pb) // TODO: same id problem for sorters and mdds, needs attention
				t.Clauses.AddClauseSet(tt.Clauses)
				pb.Typ = EQ
			} else {
				t = TranslateComplexThreshold(pb)
			}
		}
	}

	return
}

func TranslateComplexThreshold(pb *Threshold) (t ThresholdTranslation) {
	pb.Normalize(LE, true)
	pb.SortWeight()

	var err error
	switch glob.Complex_flag {
	case "mdd":
		t, err = TranslateByMDD(pb)
		if err != nil {
			panic(err.Error())
		}
		glob.D(pb.Id, " mdd:", t.Clauses.Size())
	case "sn":
		t, err = TranslateBySN(pb)
		if err != nil {
			panic(err.Error())
		}
		glob.D(pb.Id, " Complex, SN:", t.Clauses.Size())
	case "hybrid":
		tSN, err1 := TranslateBySN(pb)
		tMDD, err2 := TranslateByMDD(pb)

		if err1 != nil {
			panic(err1.Error())
		}

		glob.D(pb.Id, "Complex, SN:", tSN.Clauses.Size(), " mdd:", tMDD.Clauses.Size())

		if err2 == nil && tMDD.Clauses.Size() < tSN.Clauses.Size() {
			t.Clauses = tMDD.Clauses
			t.Typ = CMDD
		} else {
			t.Clauses = tSN.Clauses
			t.Typ = CSN
		}
	default:
		panic("Complex_flag option not available: " + glob.Complex_flag)
	}

	glob.A(t.Clauses.Size() > 0, pb.Id, " non-trivial pb should produce some clauses...")

	return
}

// returns if preprocessing was successful
// returns if it cant do the preprocessing
func PreprocessPBwithExactly(pb1 *Threshold, pb2 *Threshold) bool {

	//assumptions:
	//check for correct property of pb2
	//check for overlap of literals
	//both pb1 and pb2 are sorted in variable ordering!

	if pb2.Typ == EQ {
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
// Uses the translation of pb2 (count translation)
// TODO deprecated
func PreprocessPBwithAMO(pb *Threshold, amo CardTranslation) bool {

	//assumptions:
	//check for correct property of pb2
	//check for overlap of literals
	//both pb1 and amo are in the same ordering

	glob.A(amo.PB != nil, "amo PB pointer is not set correctly!")
	b, es := CommonSlice(pb.Entries, amo.PB.Entries)
	//fmt.Println(amo.PB.Entries, es)

	if !b {
		panic("Check if amo fits  with the pb1")
	}

	last := int64(0)
	for i, e := range es {
		es[i].Weight = e.Weight - last
		es[i].Literal = amo.Aux[i]
		last = e.Weight
	}

	pb.RemoveZeros()

	return true
}

// returns if preprocessing was successful
// Uses the translation of pb2 (count translation)
func TranslatePBwithAMO(pb *Threshold, amo CardTranslation) (t ThresholdTranslation) {

	b := PreprocessPBwithAMO(pb, amo)
	if !b {
		panic("Translate PB with AMO called on wrong input")
	}
	chain := CleanChain(pb.Entries, amo.Aux)
	t, err := TranslateByMDDChain(pb, Chains{chain})
	if err != nil {
		panic(err.Error())
	}

	return t
}

func TranslateBySNChain(pb *Threshold, literals []sat.Literal) (t ThresholdTranslation) {
	// check for overlap of variables
	// just do a rewrite, and call translateByMDD, reuse variables
	return
}
