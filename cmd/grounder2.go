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

func computeAssignments(gLits []bule.Literal, constraints []bule.Constraint, current map[string]string, assignments []map[string]string) {

	// current assignment is empty

	var i int
	var cons bule.Constraint
	var is bool
	var variable string
	var value string
	var err error
	for i, cons = range constraints {
		is, variable, value, err = cons.IsInstantiation()
		if err != nil {
			panic(err)
		}
		if is {
			break
		}
	}

	if is {
		updated_constraints := append(constraints[:i], constraints[i+1:]...)
		current[variable] = value
	}

	// 1. iterate through constraints, if is instantiation then:
	//			add to final assignments,
	//			remove from constraints and apply to rest constraints and to literals
	// 2. iterate through gLits, if ground, then check in tuples if exists, otherwise stop
	// 3. if all constraints are assigned, Roll out first gLits, and go to 1.
	// If no

	return
}
