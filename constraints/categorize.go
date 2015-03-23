package constraints

import (
	"fmt"
	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/sat"
	"golang.org/x/tools/container/intsets"
	"sort"
)

// adds pb to the map literal -> pb.Id; as well as recording
func addToCategory(nextId *int, pb *Threshold, cat map[sat.Literal][]int, lit2id map[sat.Literal]int, litSet *intsets.Sparse) {
	pb.Normalize(LE, true)
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

func Categorize2(pbs []*Threshold) {
	// this is becoming much more elaborate NOW

	simplOcc := make(map[sat.Literal][]int, len(pbs)) // literal to list of implyfiers it occurs in
	complOcc := make(map[sat.Literal][]int, len(pbs)) // literal to list of complex pbs it occurs in
	litSets := make([]intsets.Sparse, len(pbs))       // pb.Id -> intsSet of literalIds

	nextId := 0
	lit2id := make(map[sat.Literal]int, 0) // literal to its id
	translated := make([]bool, len(pbs))

	for i, pb := range pbs {

		glob.A(!pb.Empty(), "pb should be non-empty. pb.Id", pb.Id)

		t := CatSimpl(pb)

		switch t.Typ {
		case UNKNOWN:
			//groups = append(groups, Group{pb, []*Threshold{}})
			addToCategory(&nextId, pb, complOcc, lit2id, &litSets[pb.Id])
		case AMO, EX1, EXK:
			//simps = append(simps, pb)
			addToCategory(&nextId, pb, simplOcc, lit2id, &litSets[pb.Id])
		default: // already translated
			glob.A(t.Clauses.Size() > 0, "Translated pbs should contain clauses, special case?")
			translated[i] = true
		}
	}

	//glob.A(len(simpls)+len(compls)+len(rest) == len(pbs), "partitioning into different types dont add up")

	//  id2lit := make([]sat.Literal, nextId)

	//	fmt.Println("not use\t: ", len(rest))
	//	fmt.Println("groups \t: ", len(groups))
	//	fmt.Println("simps  \t: ", len(simps))

	ex_checks := make(map[Match]int, 0)
	amo_checks := make(map[Match]int, 0)

	ex_matchings := make(map[int][]Matching, 0)  // compl_id -> []Matchings
	amo_matchings := make(map[int][]Matching, 0) // compl_id ->  []Matchings

	for lit, list := range complOcc {
		//id2lit[lit2id[lit]] = lit
		for _, c := range list {
			for _, s := range simplOcc[lit] {

				// of comp c and simpl s there is at least
				// count how many:

				if amo_checks[Match{c, s}] == 0 && ex_checks[Match{c, s}] == 0 {
					// 0 means it has not been checked,
					// as there is at least one intersection
					var inter intsets.Sparse
					inter.Intersection(&litSets[c], &litSets[s])
					if pbs[s].Typ == LE {
						amo_checks[Match{c, s}] = inter.Len()
						if inter.Len() > 2 {
							amo_matchings[c] = append(amo_matchings[c], Matching{s, &inter})
						}
					} else if pbs[s].Typ == EQ {
						ex_checks[Match{c, s}] = inter.Len()
						if inter.Len() > 1 {
							ex_matchings[c] = append(amo_matchings[c], Matching{s, &inter})
						}
					} else {
						glob.A(false, "case not treated")
					}
					//if inter.Len() > 2 {
					//	pbs[c].Print10()
					//	pbs[s].Print10()
					//	fmt.Println("intersection of", litSets[c].String(), litSets[s].String())
					//	fmt.Println(c, s, "intersection", inter.String(), " len:", inter.Len())
					//}
				}

			}
		}
	}

	// for each comp
	// choose longest matching and align pbs simp and comp
	for comp, matchings := range amo_matchings {
		pre_t, _ := TranslateByMDD(pbs[comp]) // TODO: remove, just here to compare sizes
		pbs[comp].SortVar()

		//fmt.Println("new PB", comp)
		//pbs[comp].Print10()

		glob.A(!translated[comp], "comp is should not have been translated yet")
		translated[comp] = true

		sort.Sort(MatchingsBySize(matchings))

		var inter *intsets.Sparse
		var simp int
		for _, matching := range matchings { // find the next non-translated one
			if !translated[matching.simp] {
				// choose longest matching, that is not translated yet
				inter = matching.inter
				simp = matching.simp
				translated[simp] = true
				break //take this one!
			}
		}
		//pbs[simp].Print10()

		ind_entries := make(IndEntries, inter.Len())
		comp_rest := make([]*Entry, len(pbs[comp].Entries)-inter.Len())
		simp_rest := make([]*Entry, len(pbs[simp].Entries)-inter.Len())

		ind_pos := 0
		rest_pos := 0
		for i, x := range pbs[comp].Entries {
			if inter.Has(lit2id[x.Literal]) {
				ind_entries[ind_pos].c = &pbs[comp].Entries[i]
				ind_pos++
			} else {
				comp_rest[rest_pos] = &pbs[comp].Entries[i]
				rest_pos++
			}
		}

		ind_pos = 0
		rest_pos = 0
		for i, x := range pbs[simp].Entries {
			if inter.Has(lit2id[x.Literal]) {
				ind_entries[ind_pos].s = &pbs[simp].Entries[i]
				ind_pos++
			} else {
				simp_rest[rest_pos] = &pbs[simp].Entries[i]
				rest_pos++
			}
		}

		//fmt.Println("intersection of", litSets[comp].String(), litSets[simp].String())
		//fmt.Println("intersection", inter.String(), " len:", inter.Len())
		//fmt.Println(ind_entries)
		//fmt.Println(comp_rest)
		//fmt.Println(simp_rest)

		sort.Sort(ind_entries)

		compEntries := make([]Entry, len(pbs[comp].Entries))
		simpEntries := make([]Entry, len(pbs[simp].Entries))
		for i, ie := range ind_entries {
			compEntries[i] = *ie.c
			simpEntries[i] = *ie.s
		}
		for i := len(ind_entries); i < len(pbs[comp].Entries); i++ {
			compEntries[i] = *comp_rest[i-len(ind_entries)]
		}
		for i := len(ind_entries); i < len(pbs[simp].Entries); i++ {
			simpEntries[i] = *simp_rest[i-len(ind_entries)]
		}

		pbs[comp].Entries = compEntries
		pbs[simp].Entries = simpEntries
		//fmt.Println("reordering accoring to weights:")
		//pbs[comp].Print10()
		//pbs[simp].Print10()

		simp_translation := TranslateAtMostOne(Count, pbs[simp].IdS()+"count", pbs[simp].Literals())
		simp_translation.PB = pbs[simp]
		// replaces Preprocesss with AMO
		last := int64(0)
		for i, _ := range ind_entries {
			tmp := compEntries[i].Weight
			compEntries[i].Weight -= last
			glob.A(compEntries[i].Weight >= 0, "After rewriting PB weights cannot be negative")
			compEntries[i].Literal = simp_translation.Aux[i]
			last = tmp
		}
		pbs[comp].RemoveZeros()
		chain := CleanChain(pbs[comp].Entries, simp_translation.Aux)
		//fmt.Println("chain:")
		//chain.Print()
		//fmt.Println("rewritten:")
		//pbs[comp].Print10()

		t, err := TranslateByMDDChain(pbs[comp], Chains{chain})
		if err != nil {
			panic(err.Error())
		}

		//fmt.Println("normal", pre_t.Clauses.Size(), "chain:", t.Clauses.Size(), "overlap", len(ind_entries))
		fmt.Println(comp, "\t;", pre_t.Clauses.Size(), "\t;", t.Clauses.Size(), "\t;", inter.Len())
		//fmt.Println()

	}

	// translate the rest

	for i, _ := range pbs {
		if !translated[i] {
			// translate, simple or complex ...
			//x.Print10()
		}
	}

	return
}

type Match struct {
	comp, simp int // thresholdIds
}

type Matching struct {
	simp  int             // thresholdIds
	inter *intsets.Sparse // intersection
}

type IndEntry struct {
	c *Entry
	s *Entry
}

type IndEntries []IndEntry

func (a IndEntries) Len() int { return len(a) }
func (a IndEntries) Swap(i, j int) {
	*a[i].c, *a[j].c = *a[j].c, *a[i].c
	*a[i].s, *a[j].s = *a[j].s, *a[i].s
}
func (a IndEntries) Less(i, j int) bool { return (*a[i].c).Weight < (*a[j].c).Weight }

type MatchingsBySize []Matching

func (a MatchingsBySize) Len() int           { return len(a) }
func (a MatchingsBySize) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a MatchingsBySize) Less(i, j int) bool { return a[i].inter.Len() > a[j].inter.Len() }

func CatSimpl(pb *Threshold) (t ThresholdTranslation) {

	glob.A(!pb.Empty(), "pb should not be empty")

	t.Clauses.AddClauseSet(pb.Simplify())
	t.PB = pb

	if pb.Empty() {
		t.Typ = Facts
		return
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
				case LE: // check for binary, which is also a clause ( ~l1 \/ ~l2 )
					if len(pb.Entries) == 2 {
						t.Clauses.AddTaggedClause(pb.IdS()+"Cls", sat.Neg(pb.Entries[0].Literal), sat.Neg(pb.Entries[0].Literal))
						t.Typ = Clause
					} else {
						t.Typ = AMO
					}
				case GE: // its a clause!
					t.Clauses.AddTaggedClause(pb.IdS()+"Cls", literals...)
					t.Typ = Clause
				case EQ:
					t.Typ = EX1
				}
			} else { //cardinality
				switch pb.Typ {
				case LE, GE:
					t.Clauses.AddClauseSet(CreateCardinality(pb))
					t.Typ = CARD
				case EQ:
					t.Typ = EXK
				}
			}
		}
	}
	return
}
