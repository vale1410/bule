package mdd

import "golang.org/x/tools/container/intsets"

func (mdd *MddStore) encodePrim() {

	parents := make([]intsets.Sparse, len(mdd.store))

	for _, p := range mdd.Nodes {
		for _, c := range p.Children {
			parents[c].Insert(p.Id)
		}
	}

	/// clauses 1 - 3 are units

}
