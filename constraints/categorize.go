package constraints

import (
	"sort"

	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/sat"
	"golang.org/x/tools/container/intsets"
)

func CategorizeTranslate2(pbs []*Threshold) {

	//1) Categorize
	simplOcc := make(map[sat.Literal][]int, len(pbs)) // literal to list of simplifiers it occurs in
	complOcc := make(map[sat.Literal][]int, len(pbs)) // literal to list of complex pbs it occurs in
	litSets := make([]intsets.Sparse, len(pbs))       // pb.Id -> intsSet of literalIds

	nextId := 0
	lit2id := make(map[sat.Literal]int, 0) // literal to its id

	for i, pb := range pbs {

		if pb.Empty() {
			glob.D("pb is empty. pb.Id:", pb.Id)
			continue
		}

		pb.Normalize(LE, true)
		pb.SortVar()

		pb.CatSimpl()

		switch pb.TransTyp {
		case UNKNOWN:
			addToCategory(&nextId, pb, complOcc, lit2id, &litSets[i])
		case AMO, EX1:
			//fmt.Println(pb.Id, len(pbs))
			//pb.Print10()
			addToCategory(&nextId, pb, simplOcc, lit2id, &litSets[i])
		default: // already translated
			glob.DT(pb.Clauses.Size() == 0, "Translated pb should contain clauses, special case?", pb.Id)
			pb.Translated = true
		}
	}
	if glob.Amo_chain_flag || glob.Ex_chain_flag {
		doChaining(pbs, complOcc, simplOcc, lit2id, litSets)
	}

	for _, pb := range pbs {
		if pb.IsComplexTranslation() {

			sort.Sort(EntriesDescending(pb.Entries[pb.PosAfterChains():]))

			if glob.Rewrite_same_flag {

				pb.RewriteSameWeights()
				//glob.D("rewrite same weights:", pb.Id, len(pb.Chains))

			}
			//pb.Print10()

			if pb.Typ != OPT && len(pb.Chains) > 0 { //TODO: decide on what to use ...
				pb.TranslateByMDDChain(pb.Chains)
				//pbTMP := pb.Copy()
				//pbTMP.Chains = nil
				//pbTMP.TranslateBySN()
				//pb.Clauses = pbTMP.Clauses
				pb.Translated = true
			}
		}

		if !pb.Translated && pb.Typ != OPT {
			glob.A(len(pb.Chains) == 0, "At this point no chain.", pb)
			//glob.A(pb.Clauses.Size() == 0, pb.Id, pb, "not translation means also that there should not be any clauses.")
			pb.CategorizeTranslate1()
			pb.Translated = true
		}

		if pb.Err != nil {
			panic(pb.Err.Error())
		}
		//pb.Print10()
	}
}

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

func doChaining(pbs []*Threshold, complOcc map[sat.Literal][]int, simplOcc map[sat.Literal][]int,
	lit2id map[sat.Literal]int, litSets []intsets.Sparse) {

	//2) Prepare Matchings

	checked := make(map[Match]bool, 0)

	//ex_matchings := make(map[int][]Matching, 0)  // simpl_id -> []Matchings
	//currently ex and amo matchings are treated equivalently, the only
	//difference is that ex adds the unit clause of the ladder encoding, thus
	//the rewrite is correct and after UP the first value in the Ex is propagated.
	// TODO: explicitly rewrite and remove smallest value

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
						if inter.Len() >= glob.Len_rewrite_amo_flag {
							amo_matchings[c] = append(amo_matchings[c], Matching{s, &inter})
						}
					} else if pbs[s].Typ == EQ {
						if inter.Len() >= glob.Len_rewrite_ex_flag {
							amo_matchings[c] = append(amo_matchings[c], Matching{s, &inter})
							//ex_matchings[c] = append(amo_matchings[c], Matching{s, &inter})
						}
					} else {
						glob.A(false, "case not treated")
					}
				}
			}
		}
	}

	glob.D("amo/ex_matchings:", len(amo_matchings))

	//3) amo/ex matchings

	for comp := range pbs {
		if matchings, b := amo_matchings[comp]; b {
			workOnMatching(pbs, comp, matchings, lit2id, litSets)
		}
	}
}

func workOnMatching(pbs []*Threshold, comp int, matchings []Matching,
	lit2id map[sat.Literal]int, litSets []intsets.Sparse) {
	glob.A(!pbs[comp].Translated, "comp", comp, "should not have been translated yet")

	var chains Chains
	//inter := &intsets.Sparse{}
	var comp_offset int // the  new first position of Entries in comp

	if !glob.Amo_reuse_flag {
		//fmt.Println("before remove translated matches: len", len(matchings))
		// remove translated ones...
		p := len(matchings)
		for i := 0; i < p; i++ { // find the next non-translated one
			if pbs[matchings[i].simp].Translated {
				p--
				matchings[i] = matchings[p]
				i--
			}
		}
		matchings = matchings[:p]
		//fmt.Println("after removing translated matches: len", len(matchings))
	}

	for len(matchings) > 0 {

		//fmt.Println("len(matchings", len(matchings))
		sort.Sort(MatchingsBySize(matchings))
		matching := matchings[0]
		//glob.D(comp, matching.simp, matching.inter.Len())
		// choose longest matching, that is not translated yet
		//fmt.Println("check matching: simp", matching.simp, "inter", matching.inter.String())
		//matching.inter.IntersectionWith(&litSets[comp]) //update matching
		if matching.inter.Len() < glob.Len_rewrite_amo_flag {
			break
		}
		inter := matching.inter
		simp := matching.simp

		//pbs[comp].Print10()
		//pbs[simp].Print10()
		//glob.D("entries", comp, litSets[comp].String(), simp, litSets[simp].String(), inter.String())

		ind_entries := make(IndEntries, inter.Len())
		comp_rest := make([]*Entry, len(pbs[comp].Entries)-inter.Len()-comp_offset)
		simp_rest := make([]*Entry, len(pbs[simp].Entries)-inter.Len())
		simp_offset := len(simp_rest)

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

		{ // remove intersecting matchings
			//fmt.Println("before remove intersecting matches: len", len(matchings))
			p := len(matchings)
			for i := 1; i < p; i++ { // find the next non-translated one
				//var tmp intsets.Sparse
				//			glob.D("before", matchings[i].inter.String())
				//				glob.D("inter", inter.String())
				matchings[i].inter.DifferenceWith(inter)
				//			glob.D("after ", matchings[i].inter.String())
				//if tmp.Intersection(inter, matchings[i].inter); !tmp.IsEmpty() {
				if matchings[i].inter.IsEmpty() {
					//			fmt.Println("remove", matchings[i].inter.String())
					p--
					matchings[i] = matchings[p]
					i--
				}
			}
			matchings = matchings[1:p]
			//fmt.Println("after removing intersecting matches: len", len(matchings))
		}

		sort.Sort(ind_entries)

		compEntries := make([]Entry, len(pbs[comp].Entries))
		simpEntries := make([]Entry, len(pbs[simp].Entries))

		// fill the compEntries and simpEntries
		copy(compEntries, pbs[comp].Entries[:comp_offset])

		for i, ie := range ind_entries {
			glob.A(ie.c.Literal == ie.s.Literal, "Indicator entries should be aligned but", ie.c.Literal, ie.s.Literal)
			compEntries[i+comp_offset] = *ie.c
			simpEntries[i+simp_offset] = *ie.s
		}
		for i := range comp_rest {
			compEntries[comp_offset+len(ind_entries)+i] = *comp_rest[i]
		}
		for i := range simp_rest {
			simpEntries[i] = *simp_rest[i]
		}

		pbs[comp].Entries = compEntries
		pbs[simp].Entries = simpEntries

		var simp_translation CardTranslation
		if pbs[simp].Typ == EQ {
			simp_translation = TranslateExactlyOne(Count, pbs[simp].IdS()+"-cnt", pbs[simp].Literals())
		} else {
			glob.A(pbs[simp].Typ == LE)
			simp_translation = TranslateAtMostOne(Count, pbs[simp].IdS()+"-cnt", pbs[simp].Literals())
		}
		//simp_translation := TranslateAtMostOne(Count, pbs[simp].IdS()+"-cnt", pbs[simp].Conditionals())
		pbs[simp].Translated = true
		pbs[simp].Clauses.AddClauseSet(simp_translation.Clauses)
		simp_translation.PB = pbs[simp]
		// replaces entries with auxiliaries of the AMO
		last := int64(0)
		for i := range ind_entries {
			tmp := compEntries[i+comp_offset].Weight
			compEntries[i+comp_offset].Weight -= last
			glob.A(compEntries[i+comp_offset].Weight >= 0, "After rewriting PB weights cannot be negative")
			compEntries[i+comp_offset].Literal = simp_translation.Aux[i+simp_offset]
			last = tmp
		}
		pbs[comp].RemoveZeros()
		chain := CleanChain(pbs[comp].Entries, simp_translation.Aux[simp_offset:])
		//pbs[comp].Print10()
		//pbs[simp].Print10()
		//glob.D(len(chain))
		//glob.D(Chain(simp_translation.Aux))
		glob.A(len(chain) > 0, "chain has to have at least one element")
		comp_offset += len(chain)
		chains = append(chains, chain)

		//fmt.Println("chain:")
		//fmt.Println("rewritten:")
		pbs[simp].SortVar() // for reuse of other constraints
	}
	pbs[comp].Chains = chains
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

func (pb *Threshold) CatSimpl() {

	glob.A(!pb.Empty(), "pb should not be empty")

	if pb.Typ == OPT {
		glob.D(pb.IdS(), " is not simplyfied because is OPT")
		pb.TransTyp = UNKNOWN
		return
	}

	pb.Simplify()

	if pb.Empty() {
		pb.TransTyp = Facts
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
						pb.Clauses.AddTaggedClause("Cls", sat.Neg(pb.Entries[0].Literal), sat.Neg(pb.Entries[1].Literal))
						pb.TransTyp = Clause
					} else {
						pb.TransTyp = AMO
					}
				case GE: // its a clause!
					pb.Clauses.AddTaggedClause("Cls", literals...)
					pb.TransTyp = Clause
				case EQ:
					pb.TransTyp = EX1
				}
			} else { //cardinality
				switch pb.Typ {
				case LE, GE, EQ:
					pb.CreateCardinality()
					pb.TransTyp = CARD
				}
			}
		}
	}
	return
}
