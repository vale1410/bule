package constraints

import (
	"errors"
	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/mdd"
	"github.com/vale1410/bule/sat"
	"math"
	"strconv"
)

func (pb *Threshold) TranslateByMDD() {
	pb.TranslateByMDDChain(Chains{})
}

// chains must be in order of pb and be subsets of its literals
func (pb *Threshold) TranslateByMDDChain(chains Chains) {
	glob.A(!pb.Empty(), pb.Id, "works only for non-empty mdds")
	glob.A(pb.Positive(), pb.Id, "Weights need to be positive")
	glob.A(pb.Typ == LE, pb.Id, "works only on LE, but is", pb.Typ, pb.String())

	if len(chains) == 0 {
		pb.TransTyp = CMDD
	} else {
		pb.TransTyp = CMDDC
	}

	store := mdd.InitIntervalMdd(len(pb.Entries))
	topId, _, _, err := CreateMDDChain(&store, pb.K, pb.Entries, chains)
	store.Top = topId
	//store.Debug(true)

	if err != nil {
		pb.Err = err
		return
	}

	if glob.MDD_redundant_flag {
		store.RemoveRedundants()
		//glob.D("remove redundant nodes in MDD", removed)
	}

	pb.Clauses.AddClauseSet(convertMDD2Clauses(store, pb))
}

// Translate monotone MDDs to SAT
// Together with AMO translation
func convertMDD2Clauses(store mdd.IntervalMddStore, pb *Threshold) (clauses sat.ClauseSet) {

	pred := sat.Pred("mdd" + strconv.Itoa(pb.Id))

	top_lit := sat.Literal{true, sat.NewAtomP1(pred, store.Top)}
	clauses.AddTaggedClause("Top", top_lit)
	for _, n := range store.Nodes {
		v_id, l, vds := store.ClauseIds(*n)
		if !n.IsZero() && !n.IsOne() {

			v_lit := sat.Literal{false, sat.NewAtomP1(pred, v_id)}
			last_id := -1
			for i, vd_id := range vds {
				if last_id != vd_id {
					vd_lit := sat.Literal{true, sat.NewAtomP1(pred, vd_id)}
					if i > 0 {
						literal := pb.Entries[len(pb.Entries)-l+i-1].Literal
						clauses.AddTaggedClause("1B", v_lit, sat.Neg(literal), vd_lit)
					} else {
						clauses.AddTaggedClause("0B", v_lit, vd_lit)
					}
				}
				last_id = vd_id
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

func CreateMDD(store *mdd.IntervalMddStore, K int64, entries []Entry) (int, int64, int64, error) {
	return CreateMDDChain(store, K, entries, Chains{})
}

// Chains is a set of chains in order of the PB
// Chain: there are clauses  xi <-xi+1 <- xi+2 ... <- xi+k, and xi .. xi+k are in order of PB
// assumption: chains are subsets of literals of PB and in their order
func CreateMDDChain(store *mdd.IntervalMddStore, K int64, entries []Entry, chains Chains) (int, int64, int64, error) {

	l := len(entries) ///level

	if store.MaxNodes < len(store.Nodes) {
		return 0, 0, 0, errors.New("mdd max nodes reached")
	}

	//chain.Print()
	//fmt.Println(l, K, entries)

	if id, wmin_cache, wmax_cache := store.GetByWeight(l, K); id != -1 {

		//	fmt.Println("exists", l, K, "[", wmin, wmax, "]")

		return id, wmin_cache, wmax_cache, nil

	} else {
		//domain of variable [0,1], extend to [0..n] soon (MDDs)
		// entry of variable domain, atom: Dom: 2

		var n mdd.IntervalNode
		var err error

		//glob.D(entries, chains)

		glob.A(len(chains) == 0 || len(chains[0]) > 0, "if exists, then chain must contain at least 1 element")
		if len(chains) > 0 && chains[0][0] == entries[0].Literal { //chain mode
			chain := chains[0]
			var jumpEntries []Entry
			if len(entries) <= len(chain) { // can this happen if entries and chains are perfectly aligned?
				jumpEntries = []Entry{}
			} else {
				jumpEntries = entries[len(chain):]
			}
			// iterate over the chain
			n.Level = l
			n.Children = make([]int, len(chain)+1)

			n.Children[0], n.Wmin, n.Wmax, err = CreateMDDChain(store, K, jumpEntries, chains[1:])

			if err != nil {
				return 0, 0, 0, err
			}

			acc := int64(0)

			//			fmt.Printf("entries:%v  chain: %v", entries, chain)
			for i, _ := range chain {

				glob.A(len(chain) <= len(entries), "chain and PB are not aligned!!!! ")
				glob.A(chain[i] == entries[i].Literal, "chain and PB are not aligned!!!! ")

				var wmin2, wmax2 int64
				acc += entries[i].Weight
				n.Children[i+1], wmin2, wmax2, err = CreateMDDChain(store, K-acc, jumpEntries, chains[1:])
				n.Wmin = maxx(n.Wmin, wmin2+acc)
				n.Wmax = min(n.Wmax, wmax2+acc)

				if err != nil {
					return 0, 0, 0, err
				}

			}

		} else { //usual mode or for int-variables
			dom := 2
			n.Level = l
			n.Children = make([]int, dom)
			n.Wmin = math.MinInt64
			n.Wmax = math.MaxInt64

			var err error
			for i := int64(0); i < int64(dom); i++ {
				var wmin2, wmax2 int64

				n.Children[i], wmin2, wmax2, err = CreateMDDChain(store, K-i*entries[0].Weight, entries[1:], chains)

				n.Wmin = maxx(n.Wmin, wmin2+i*entries[0].Weight)
				n.Wmax = min(n.Wmax, wmax2+i*entries[0].Weight)

				if err != nil {
					return 0, 0, 0, err
				}
			}
		}

		return store.Insert(n), n.Wmin, n.Wmax, nil
	}
}

func min(a, b int64) int64 {
	if a <= b {
		return a
	} else {
		return b
	}
}

func maxx(a, b int64) int64 {
	if a >= b {
		return a
	} else {
		return b
	}
}
