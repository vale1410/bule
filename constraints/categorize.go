package constraints

import (
	"sort"

	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/sat"
	"golang.org/x/tools/container/intsets"
)

// adds pb to the map literal -> pb.Id; as well as recording litSets
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

func Categorize2(pbs []*Threshold) (clauses sat.ClauseSet, opt_trans ThresholdTranslation) {

	// 1) categorize constraints, fill litSets and occurence lists
	// 2) find matchings by scan through occurence lists
	// 3) for each complex pb and its list of matchings translate through chains

	//1)
	simplOcc := make(map[sat.Literal][]int, len(pbs)) // literal to list of simplifiers it occurs in
	complOcc := make(map[sat.Literal][]int, len(pbs)) // literal to list of complex pbs it occurs in
	litSets := make([]intsets.Sparse, len(pbs))       // pb.Id -> intsSet of literalIds

	nextId := 0
	lit2id := make(map[sat.Literal]int, 0) // literal to its id
	translated := make([]bool, len(pbs))

	for i, pb := range pbs {

		if pb.Empty() {
			glob.D("pb is empty. pb.Id:", pb.Id)
			continue
		}

		pb.Normalize(LE, true)
		pb.SortWeight()

		t := CatSimpl(pb)
		clauses.AddClauseSet(t.Clauses)

		switch t.Typ {
		case UNKNOWN:
			addToCategory(&nextId, pb, complOcc, lit2id, &litSets[pb.Id])
		case AMO, EX1, EXK:
			addToCategory(&nextId, pb, simplOcc, lit2id, &litSets[pb.Id])
		default: // already translated
			glob.A(t.Clauses.Size() > 0, "Translated pbs should contain clauses, special case?")
			translated[i] = true
		}
	}

	//2)

	checked := make(map[Match]bool, 0)

	ex_matchings := make(map[int][]Matching, 0)  // compl_id -> []Matchings
	amo_matchings := make(map[int][]Matching, 0) // compl_id -> []Matchings

	for lit, list := range complOcc {
		//id2lit[lit2id[lit]] = lit
		for _, c := range list {
			for _, s := range simplOcc[lit] {

				if !checked[Match{c, s}] {
					// of comp c and simpl s there is at least
					checked[Match{c, s}] = true
					// 0 means it has not been checked,
					// as there is at least one intersection
					var inter intsets.Sparse
					inter.Intersection(&litSets[c], &litSets[s])
					if pbs[s].Typ == LE {
						if inter.Len() > 2 {
							amo_matchings[c] = append(amo_matchings[c], Matching{s, &inter})
						}
					} else if pbs[s].Typ == EQ {
						if inter.Len() > 1 {
							ex_matchings[c] = append(amo_matchings[c], Matching{s, &inter})
						}
					} else {
						glob.A(false, "case not treated")
					}
				}
			}
		}
	}

	glob.D("amo_matchings:", len(amo_matchings))
	glob.D("ex_matchings:", len(ex_matchings))

	//3)

	for comp, _ := range pbs {
		if matchings, b := amo_matchings[comp]; b {
			//pre_t, _ := TranslateByMDD(pbs[comp]) // TODO: remove, just here to compare sizes
			//pbs[comp].SortWeight()                // because TranslateByMDD might reorder entries

			//fmt.Println("new PB", comp)
			//pbs[comp].Print10()

			glob.A(!translated[comp], "comp", comp, "should not have been translated yet")

			sort.Sort(MatchingsBySize(matchings))

			var chains Chains
			inter := &intsets.Sparse{}
			var comp_offset int                  // the  new first position of Entries in comp
			for _, matching := range matchings { // find the next non-translated one
				if !translated[matching.simp] {
					// choose longest matching, that is not translated yet
					//fmt.Println("check matching: simp", matching.simp, "inter", matching.inter.String())
					matching.inter.IntersectionWith(&litSets[comp]) //update matching
					inter = matching.inter
					if inter.Len() > 2 {
						simp := matching.simp

						//pbs[comp].Print10()
						//pbs[simp].Print10()
						//fmt.Println("entries", comp, litSets[comp].String(), simp, litSets[simp].String(), inter.String())

						ind_entries := make(IndEntries, inter.Len())
						comp_rest := make([]*Entry, len(pbs[comp].Entries)-inter.Len()-comp_offset)
						simp_rest := make([]*Entry, len(pbs[simp].Entries)-inter.Len())

						ind_pos := 0
						rest_pos := 0
						for i := comp_offset; i < len(pbs[comp].Entries); i++ {
							if inter.Has(lit2id[pbs[comp].Entries[i].Literal]) {
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
						litSets[comp].DifferenceWith(inter)
						//fmt.Println("litSets[comp] is now", litSets[comp].String())
						//fmt.Println(ind_entries)
						//fmt.Println(comp_rest)
						//fmt.Println(simp_rest)

						sort.Sort(ind_entries)

						compEntries := make([]Entry, len(pbs[comp].Entries))
						simpEntries := make([]Entry, len(pbs[simp].Entries))

						// fill the compEntries and simpEntries
						copy(compEntries, pbs[comp].Entries[:comp_offset])

						for i, ie := range ind_entries {
							compEntries[i+comp_offset] = *ie.c
							simpEntries[i] = *ie.s
						}
						for i, _ := range comp_rest {
							//fmt.Println(i, comp_offset, len(ind_entries), comp_rest)
							compEntries[comp_offset+len(ind_entries)+i] = *comp_rest[i]
						}
						for i := len(ind_entries); i < len(pbs[simp].Entries); i++ {
							simpEntries[i] = *simp_rest[i-len(ind_entries)]
						}

						pbs[comp].Entries = compEntries
						pbs[simp].Entries = simpEntries
						//pbs[comp].Print10()
						//pbs[simp].Print10()

						simp_translation := TranslateAtMostOne(Count, pbs[simp].IdS()+"count", pbs[simp].Literals())
						translated[simp] = true
						clauses.AddClauseSet(simp_translation.Clauses)
						simp_translation.PB = pbs[simp]
						// replaces entries with auxiliaries of the AMO
						last := int64(0)
						for i, _ := range ind_entries {
							tmp := compEntries[i+comp_offset].Weight
							compEntries[i+comp_offset].Weight -= last
							glob.A(compEntries[i+comp_offset].Weight >= 0, "After rewriting PB weights cannot be negative")
							compEntries[i+comp_offset].Literal = simp_translation.Aux[i]
							last = tmp
						}
						pbs[comp].RemoveZeros()
						chain := CleanChain(pbs[comp].Entries, simp_translation.Aux) // real intersection
						comp_offset += len(chain)
						chains = append(chains, chain)

						//fmt.Println("chain:")
						//chain.Print()
						//fmt.Println("rewritten:")
						//pbs[comp].Print10()
					}
					//else {
					//		fmt.Println("matching not taken, it looks after updated intersection:", inter.String())
					//	}
				}
			}

			// just add this to t!
			if pbs[comp].Typ == OPT {
				opt_trans.PB = pbs[comp]
				opt_trans.Chains = chains
				glob.D("translating optimization statement with chains: ", pbs[comp])
			} else {
				t, err := TranslateByMDDChain(pbs[comp], chains)
				if err != nil {
					panic(err.Error())
				}
				translated[comp] = true
				clauses.AddClauseSet(t.Clauses)

			}

			//fmt.Println(glob.Filename_flag, "\t;", pre_t.Clauses.Size(), "\t;", t.Clauses.Size(), "\t;", len(chains))
			//fmt.Println(glob.Filename_flag, "\t;", pre_t.Clauses.Size(), "\t;", t.Clauses.Size(), "\t;", len(chains))
		}
	}

	// translate the rest

	for i, _ := range pbs {
		if !translated[i] && pbs[i].Typ != OPT {
			t := Categorize1(pbs[i])
			clauses.AddClauseSet(t.Clauses)
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

	if pb.Typ == OPT {
		glob.D(pb.IdS(), " is not simplyfied because is OPT")
		t.Typ = UNKNOWN
		return t
	}

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
						t.Clauses.AddTaggedClause("Cls", sat.Neg(pb.Entries[0].Literal), sat.Neg(pb.Entries[1].Literal))
						t.Typ = Clause
					} else {
						t.Typ = AMO
					}
				case GE: // its a clause!
					t.Clauses.AddTaggedClause("Cls", literals...)
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
