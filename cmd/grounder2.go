package cmd

import bule "github.com/vale1410/bule/grounder"

func stageX1GeneratorsAndFacts(p *bule.Program) {

	//	// COMPUTE ORDER IN WHICH TO PROCESS
	//	// Stage 1: collect tuples for g-lits and s-lits
	//	var order []*bule.Rule
	//	for _, r := range order {
	//		// until fixpoint:
	//		// - roll out g-literal
	//		// - clean up instantiations
	//		// - potentially remove rules
	//		// if head is ground -> add to tuples.
	//	}

}

func stageX2Iterators(p *bule.Program) {

}

func stageX3Clauses(p *bule.Program) {

}

func computeAssignment(gLits []bule.Literal, constraints []bule.Constraint) (assignments []map[string]string) {
	return
}
