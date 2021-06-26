package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	bule "github.com/vale1410/bule/grounder"
)

func stage0Prerequisites(p *bule.Program) {
	for key, val := range constStringMapFlag {
		p.Constants[key] = strconv.Itoa(val)
	}

	//		debug(1, "Input:")
	//		p.PrintDebug(1)

	{
		// This inserts arity 1 with term 0 in all zero arity literals!
		err := p.CheckArityOfLiterals()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	{
		err := p.CheckFactsInIterators()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	stageInfo(p, "Replace Constants and Math", "(#const a=3. and Function Symbols (#mod)")
	p.ReplaceConstantsAndMathFunctions()

	{
		stageInfo(p, "CollectStringTermsToIntegers", "If there is a q[a] or p(a) somewhere, replace by q[4] , and mark IndexToString[4]='a'. Also remember that first entry of q is Int2String.")
		err := p.CollectStringTermsToIntegers()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	{
		stageInfo(p, "CheckUnboundVariables", "Check for unbound variables that are not marked as such.")
		err := p.CheckUnboundVariables()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func stage1GeneratorsAndFacts(p *bule.Program) {
	round := 0
	changed := true
	for changed {

		debug(2, "STAGE 0: This is round: ", round)
		round++
		changed = false

		stage(p, &changed,
			runFixpoint(p.ExpandGroundRanges),
			"ExpandGroundRanges",
			"p[1..2]. and also X==1..2, but not Y==A..B.")

		stage(p, &changed,
			p.ConstraintSimplification,
			"ConstraintSimplification.",
			"For each constraint (X==v) rewrite clause with (X<-v) and remove constraint.")

		stage(p, &changed,
			p.CollectGroundFacts,
			"CollectGroundFacts",
			"Example: p[1,2]. r[1]. but not p[1],p[2]. and also not p[X,X], or p[1,X].")

		stage(p, &changed,
			p.FindFactsThatAreFullyCollected,
			"FindFactsThatAreFullyCollected",
			"Of all facts that do not occur in the head, they will be set to FinishedCollection.")

		stage(p, &changed,
			p.InstantiateAndRemoveFactFromGenerator,
			"InstantiateAndRemoveFactFromGenerator",
			"Example: a fact p[T1,T2] with tuples (v11,v12)..(vn2,vn1) occurs in clause, expand clause with (T1 == v11, T2 == v12).")

		stage(p, &changed,
			p.ConstraintSimplification,
			"ConstraintSimplification.",
			"For each constraint (X==v) rewrite clause with (X<-v) and remove constraint.")

		stage(p, &changed,
			p.RemoveRulesWithNegatedGroundGenerator,
			"RemoveRulesWithNegatedGroundGenerator.",
			"~p[1,2]=> q(1,2) and p[1,2] is not a fact then remove rule!")

		stage(p, &changed,
			p.RemoveNegatedGroundGenerator,
			"RemoveNegatedGroundGenerator",
			"~p[1,2]=> q(1,2) and p[1,2] is a fact, then remove from generators!")
	}

	stage(p, &changed,
		p.CollectExplicitTupleDefinitions,
		"CollectExplicitTupleDefinitions.",
		"#exists[3] :: p(1,2)? %% Then remove this rule.")

	stage(p, &changed,
		p.RemoveRulesWithGenerators,
		"RemoveRulesWithGeneratorsBecauseTheyHaveEmptyDomains",
		"Because they have empty domains, e.g. \n edge[_,_,V] :: vertex[V]. %% there are no edges!")
}

// Unroll all iterators.
func stage2Iterators(p *bule.Program) {
	round := 0
	changed := true
	for changed {

		debug(2, "Stage 2 round: ", round)
		round++
		changed = false

		stage(p, &changed,
			runFixpoint(p.TransformConstraintsToInstantiationIterator),
			"Fixpoint of TransformConstraintsToInstantiationIterator.",
			"If we have p(X,Y):q[X,Y]:X==3+1, then simplify this to: p(4,Y):q[4,Y].")

		stage(p, &changed,
			p.InstantiateAndRemoveFactFromIterator,
			"InstantiateAndRemoveFactFromIterator", "")

		stage(p, &changed,
			p.CleanIteratorFromGroundBoolExpressions,
			"CleanIteratorFromGroundBoolExpressions.",
			"p(X,Y):q[X,Y]:#true :: p(X,Y):q[X,Y]")

		stage(p, &changed,
			p.ConvertHeadOnlyIteratorsToLiterals,
			"ConvertHeadOnlyIteratorsToLiterals.",
			"p(1,2) -> p(1,2) %% but now as literal!")
	}

	stage(p, &changed,
		p.RemoveLiteralsWithEmptyIterators,
		"RemoveLiteralsWithEmptyIterators",
		"win(E):edgeId[1,E]. %% edgeId is empty \n Remove!")

}

func stage3Clauses(p *bule.Program) {

	var err error
	{
		err = p.CheckNoGeneratorsOrIterators()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	changed := true
	round := 0
	for changed {
		debug(2, "Stage 3 iteration; round: ", round)
		round++
		changed = false

		stage(p, &changed,
			p.InstantiateExplicitNonGroundLiterals,
			"fixpoint(InstantiateExplicitNonGroundLiterals.)",
			"p(X,Y),q(X).% p is explicit -> p(1,2), q(X).")

		stage(p, &changed,
			p.ConstraintSimplification,
			"ConstraintSimplification.",
			"For each constraint (X==v) rewrite clause with (X<-v) and remove constraint.")

		stage(p, &changed,
			p.RemoveClausesWithExplicitLiteralAndTuplesThatDontExist,
			"RemoveClausesWithExplicitLiteralAndTuplesThatDontExist",
			"")
	}

	debug(2, "No more non-ground explicit variables!")
	{
		err = p.CheckNoExplicitDeclarationAndNonGroundExplicit()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	stage(p, &changed,
		p.CollectGroundTuples,
		"CollectGroundTuples", "")

	{
		stageInfo(p, "Ground non-Ground Lits", "Ground from all tuples the non-ground literals, until fixpoint.")
		ok := true
		i := 0
		for ok {
			i++
			var err error
			ok, err = p.InstantiateNonGroundLiterals()
			if err != nil {
				fmt.Printf("Error occurred in grounding when instantiating non-ground literals. Iteration %v.\n %v\n", i, err)
				os.Exit(1)
			}
			stageInfo(p, "Do Fixpoint of TransformConstraintsToInstantiation.", ""+
				"For each constraint (X==v) rewrite clause with (X<-v) and remove constraint.")
			p.ConstraintSimplification()

			stageInfo(p, "RemoveClausesWithTuplesThatDontExist.", "")
			p.RemoveClausesWithTuplesThatDontExist()
		}
	}

	if quantificationFlag {
		stageInfo(p, "Merge Quantification Levels", "")
		p.MergeConsecutiveQuantificationLevels()
		debug(2, "Merged alternations:", p.Alternation)
	}

}

func stage4Printing(p *bule.Program, args []string) {

	if textualFlag && withFactsFlag {
		p.PrintFacts()
	}

	clauseProgram := translateFromRuleProgram(*p)
	sb := clauseProgram.StringBuilder()
	fmt.Println(sb.String())

}

func stage(p *bule.Program, change *bool, f func() (bool, error), stage string, info string) {
	start := time.Now()
	stageInfo(p, stage, info)
	tmp, err := f()
	if err != nil {
		fmt.Println("ERROR IN STAGE:")
		fmt.Println(err)
		os.Exit(1)
	}
	if tmp {
		debug(2, "Stage changed Program!")
	}
	*change = *change || tmp

	elapsed := time.Since(start)
	log.Printf("DEBUGTIME %6.2f %v %v", elapsed.Seconds(), stage, len(p.Rules))
}

func runFixpoint(f func() (bool, error)) func() (bool, error) {
	return func() (changed bool, err error) {
		ok := true
		for ok {
			ok, err = f()
			changed = changed || ok
			if err != nil {
				return changed, fmt.Errorf("Error occurred in grounding.\n %w", err)
			}
		}
		return
	}
}

func stageInfo(p *bule.Program, stage string, info string) {
	p.PrintDebug(2)
	debug(2, "===================================================")
	debug(2, stage)
	debug(2, "===================================================")
	debug(3, info)
	debug(3, "---------------------------------------------------\n\n")
}
