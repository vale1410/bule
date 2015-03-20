package constraints

import (
	"fmt"
	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/sat"
)

type Group struct {
	pb    *Threshold
	simps []*Threshold
}

func Categorize(pbs []*Threshold) (groups []Group, rest []ThresholdTranslation) {
	// this is becoming much more elaborate NOW

	groups = make([]Group, 0, len(pbs))
	simps := make([]*Threshold, 0, len(pbs)) // AMO and EXK

	simpsOcc := make(map[sat.Literal][]int, 0)
	complOcc := make(map[sat.Literal][]int, 0)

	for _, pb := range pbs {

		t := CatSimpl(pb)

		switch t.Typ {
		case UNKNOWN:
			pb.Normalize(AtMost, true)
			groups = append(groups, Group{pb, []*Threshold{}})
			for _, x := range pb.Entries {
				complOcc[x.Literal] = append(complOcc[x.Literal], pb.Id)
			}
		case AtMostOne, ExactlyOne, ExactlyK:
			pb.Normalize(AtMost, true)
			simps = append(simps, pb)
			for _, x := range pb.Entries {
				simpsOcc[x.Literal] = append(simpsOcc[x.Literal], pb.Id)
			}
		default: // already translated
			rest = append(rest, t)
		}
	}

	glob.A(len(rest)+len(groups)+len(simps) == len(pbs), "partitioning into different types dont add up")

	//	fmt.Println("not use\t: ", len(rest))
	//	fmt.Println("groups \t: ", len(groups))
	//	fmt.Println("simps  \t: ", len(simps))

	ex_checks := make(map[Match]bool, 0)
	amo_checks := make(map[Match]bool, 0)

	for key, list := range complOcc {
		for _, c := range list {
			for _, s := range simpsOcc[key] {
				if pbs[s].Typ == AtMost {
					amo_checks[Match{c, s}] = true
				} else if pbs[s].Typ == Equal {
					ex_checks[Match{c, s}] = true
				} else {
					glob.A(false, "case not treated")
				}
			}
		}

		//if len(list) != 0 && len(simpsOcc[key]) != 0 {
		//	fmt.Println(key.ToTxt(), "complex:", list, "simps", simpsOcc[key])
		//}
	}
	fmt.Println("ex check:", len(ex_checks))
	fmt.Println("amo check:", len(amo_checks))

	return
}

type Match struct {
	comp, simp int // thresholdIds
}

type Matching struct {
	m     Match
	inter []sat.Literal
}

func CatSimpl(pb *Threshold) (t ThresholdTranslation) {

	glob.A(!pb.Empty(), "pb should not be empty")

	t.Clauses.AddClauseSet(pb.Simplify())

	if pb.Empty() {
		t.Typ = Facts
		return
	} else {

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
				case AtMost:
					t.Typ = AtMostOne
				case AtLeast: // its a clause!
					t.Clauses.AddTaggedClause(pb.IdS()+"Cls", literals...)
					t.Typ = Clause
				case Equal:
					t.Typ = ExactlyOne
				}
			} else { //cardinality
				switch pb.Typ {
				case AtMost, AtLeast:
					t.Clauses.AddClauseSet(CreateCardinality(pb))
					t.Typ = Cardinality
				case Equal:
					t.Typ = ExactlyK
				}
			}
		}
	}
	return
}
