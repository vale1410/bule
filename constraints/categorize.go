package constraints

import (
	"fmt"
	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/sat"
	"golang.org/x/tools/container/intsets"
)

type Group struct {
	pb    *Threshold
	simps []*Threshold
}

// adds pb to the map literal -> pb.Id; as well as recording
func addToCategory(nextId *int, pb *Threshold, cat map[sat.Literal][]int, lit2id map[sat.Literal]int, litSet *intsets.Sparse) {
	pb.Normalize(AtMost, true)
	for _, x := range pb.Entries {
		cat[x.Literal] = append(cat[x.Literal], pb.Id)
	}
	for _, e := range pb.Entries {
		if _, b := lit2id[e.Literal]; !b {
			lit2id[e.Literal] = *nextId
			*nextId++
		}
		litSet.Insert(lit2id[e.Literal])
	}
}

func Categorize(pbs []*Threshold) (groups []Group, rest []ThresholdTranslation) {
	// this is becoming much more elaborate NOW

	//	groups = make([]Group, 0, len(pbs))
	//	simps := make([]*Threshold, 0, len(pbs)) // AMO and EXK

	simplOcc := make(map[sat.Literal][]int, len(pbs)) // literal to list of implyfiers it occurs in
	complOcc := make(map[sat.Literal][]int, len(pbs)) // literal to list of complex pbs it occurs in
	litSets := make([]intsets.Sparse, len(pbs))       // pb.Id -> intsSet of literalIds

	nextId := 0
	lit2id := make(map[sat.Literal]int, 0) // literal to its id

	for _, pb := range pbs {

		t := CatSimpl(pb)

		switch t.Typ {
		case UNKNOWN:
			//groups = append(groups, Group{pb, []*Threshold{}})
			addToCategory(&nextId, pb, complOcc, lit2id, &litSets[pb.Id])
		case AtMostOne, ExactlyOne, ExactlyK:
			//simps = append(simps, pb)
			addToCategory(&nextId, pb, simplOcc, lit2id, &litSets[pb.Id])
		default: // already translated
			glob.A(t.Clauses.Size() > 0, "Translated pbs should contain clauses, special case?")
			rest = append(rest, t)
		}
	}

	//glob.A(len(simpls)+len(compls)+len(rest) == len(pbs), "partitioning into different types dont add up")

	//  id2lit := make([]sat.Literal, nextId)

	//	fmt.Println("not use\t: ", len(rest))
	//	fmt.Println("groups \t: ", len(groups))
	//	fmt.Println("simps  \t: ", len(simps))

	ex_checks := make(map[Match]int, 0)
	amo_checks := make(map[Match]int, 0)

	ex_matchings := make(map[int][]Matching, 0)
	amo_matchings := make(map[int][]Matching, 0)

	for lit, list := range complOcc {
		//id2lit[lit2id[lit]] = lit
		for _, c := range list {
			for _, s := range simplOcc[lit] {

				// of comp c and simpl s there is at least
				// count how many:

				if amo_checks[Match{c, s}] == 0 && ex_checks[Match{c, s}] == 0 {
					var inter intsets.Sparse
					inter.Intersection(&litSets[c], &litSets[s])
					if pbs[s].Typ == AtMost {
						amo_checks[Match{c, s}] = inter.Len()
						amo_matchings[c] = append(amo_matchings[c], Matching{s, inter})
					} else if pbs[s].Typ == Equal {
						ex_checks[Match{c, s}] = inter.Len()
						ex_matchings[c] = append(amo_matchings[c], Matching{s, inter})
					} else {
						glob.A(false, "case not treated")
					}
					if inter.Len() > 2 {
						pbs[c].Print10()
						pbs[s].Print10()
						fmt.Println("intersection of", litSets[c].String(), litSets[s].String())
						fmt.Println(c, s, "intersection", inter.String(), " len:", inter.Len())
					}
				}

			}
		}
	}

	if len(ex_checks) > 0 || len(amo_checks) > 0 {
		fmt.Println("total constraints", len(pbs))
		fmt.Println("amo check:", len(amo_checks))
		fmt.Println("ex check:", len(ex_checks))
		fmt.Println()
		fmt.Println()
		// for each Match compute the intersection
	}

	return
}

//Do the matching

type Match struct {
	comp, simp int // thresholdIds
}

type Matching struct {
	simp  int            // thresholdIds
	inter intsets.Sparse // intersection
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
				case AtMost: // check for binary, which is also a clause ( ~l1 \/ ~l2 )
					if len(pb.Entries) == 2 {
						t.Clauses.AddTaggedClause(pb.IdS()+"Cls", sat.Neg(pb.Entries[0].Literal), sat.Neg(pb.Entries[0].Literal))
						t.Typ = Clause
					} else {
						t.Typ = AtMostOne
					}
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
