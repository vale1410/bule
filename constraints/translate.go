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
	CMDDC
	TranslationTypes
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
	case CMDDC:
		s = "CMDDC"
	case TranslationTypes:
		s = "TranslationTypes"
	case CSNC: // not yet implemented
		s = "CSNC"
	default:
		panic("has not been implemented")
	}
	return
}

func (pb *Threshold) IsComplexTranslation() (b bool) {
	switch pb.TransTyp {
	case UNKNOWN, CMDD, CSN, CMDDC, CSNC:
		b = true
	default:
		b = false
	}
	return
}

func (pb *Threshold) CategorizeTranslate1() {

	pb.SortDescending()
	// per default all information that can be simplified will be in form of facts
	pb.Simplify()
	pb.TransTyp = Facts
	pb.Translated = true

	if len(pb.Entries) == 0 {
		//glob.D(pb.Id, "was simplified completely")
	} else {
		if b, literals := pb.Cardinality(); b {
			//	glob.D("debug")
			//			pb.Print10()

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
					if len(pb.Entries) == 2 {
						pb.Clauses.AddTaggedClause("Cls", sat.Neg(pb.Entries[0].Literal), sat.Neg(pb.Entries[1].Literal))
						pb.TransTyp = Clause
					} else {
						trans := TranslateAtMostOne(Heule, "H_AMO", literals)
						pb.Clauses.AddClauseSet(trans.Clauses)
						pb.TransTyp = AMO
					}
				case GE: // its a clause!
					pb.Clauses.AddTaggedClause("Cls", literals...)
					pb.TransTyp = Clause
				case EQ: // Ex1
					trans := TranslateExactlyOne(Heule, "H_EX1", literals)
					pb.Clauses.AddClauseSet(trans.Clauses)
					pb.TransTyp = EX1
				}
			} else {
				pb.CreateCardinality()
				pb.TransTyp = CARD
			}

		} else {
			// treat equality as two constraints!
			if pb.Typ == EQ {
				glob.D(pb.Id, " decompose in >= amd <=")
				pbLE := pb.Copy()
				pbLE.Typ = LE
				pbGE := pb.Copy()
				pbGE.Typ = GE
				pbGE.Id = -pb.Id
				pbLE.TranslateComplexThreshold()
				pbGE.TranslateComplexThreshold()
				pb.Clauses.AddClauseSet(pbLE.Clauses)
				pb.Clauses.AddClauseSet(pbGE.Clauses)
			} else {
				pb.TranslateComplexThreshold()
			}
		}
	}
	return
}

func (pb *Threshold) TranslateComplexThreshold() {

	glob.A(!pb.Empty(), "No Empty at this point.")
	glob.A(len(pb.Chains) == 0, "should not contain a chain")

	pb.Normalize(LE, true)
	pb.SortDescending()

	var err error
	switch glob.Complex_flag {
	case "mdd":
		pb.Print10()
		pb.TranslateByMDD()
		if pb.Err != nil {
			panic(err.Error())
		}
		glob.D(pb.Id, " mdd:", pb.Clauses.Size())
	case "sn":
		pb.TranslateBySN()
		if pb.Err != nil {
			panic(err.Error())
		}
		glob.D(pb.Id, " Complex, SN:", pb.Clauses.Size())
	case "hybrid":
		tSN := pb.Copy()
		tMDD := pb.Copy()
		tSN.TranslateBySN()
		tMDD.TranslateByMDD()

		if tSN.Err != nil {
			panic(tSN.Err.Error())
		}

		glob.D(pb.Id, "Complex, SN:", tSN.Clauses.Size(), " mdd:", tMDD.Clauses.Size())

		if tMDD.Err == nil && tMDD.Clauses.Size() < tSN.Clauses.Size() {
			pb.Clauses.AddClauseSet(tMDD.Clauses)
			pb.TransTyp = CMDD
		} else {
			pb.Clauses.AddClauseSet(tSN.Clauses)
			pb.TransTyp = CSN
		}
	default:
		panic("Complex_flag option not available: " + glob.Complex_flag)
	}

	glob.A(pb.Clauses.Size() > 0, pb.Id, " non-trivial pb should produce some clauses...")

	return
}

// finds trivially implied facts, returns set of facts
// removes such entries from the pb
// threshold can become empty!
func (pb *Threshold) Simplify() {

	if pb.Typ == OPT {
		glob.D(pb.IdS(), " is not simplyfied because is OPT")
		return
	}

	pb.Normalize(LE, true)

	entries := make([]Entry, 0, len(pb.Entries))

	for _, x := range pb.Entries {
		if x.Weight > pb.K {
			pb.Clauses.AddTaggedClause(pb.IdS()+"-simpl", sat.Neg(x.Literal))
		} else {
			entries = append(entries, x)
		}
	}

	pb.Entries = entries
	pb.Normalize(GE, true)

	if pb.SumWeights() == pb.K {
		for _, x := range pb.Entries {
			pb.Clauses.AddTaggedClause("Fact", x.Literal)
		}
		pb.Entries = []Entry{}
	}

	if pb.SumWeights() < pb.K {
		glob.D("c PB", pb.Id, "is UNSAT")
		pb.Entries = []Entry{}
		pb.K = -1
		// is unsatisfied: how to do that?
	}

	pb.Normalize(LE, true)
	if pb.SumWeights() <= pb.K {
		glob.D("c PB", pb.Id, "is redundant")
		pb.Entries = []Entry{}
	}

	if pb.Empty() {
		pb.Translated = true
	}

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
func TranslatePBwithAMO(pb *Threshold, amo CardTranslation) {

	b := PreprocessPBwithAMO(pb, amo)
	if !b {
		panic("Translate PB with AMO called on wrong input")
	}
	chain := CleanChain(pb.Entries, amo.Aux)
	pb.TranslateByMDDChain(Chains{chain})
	if pb.Err != nil {
		panic(pb.Err.Error())
	}
}

func TranslateBySNChain(pb *Threshold, literals []sat.Literal) {
	// check for overlap of variables
	// just do a rewrite, and call translateByMDD, reuse variables
	return
}
