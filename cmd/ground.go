/*
Copyright © 2020 Valentin Mayer-Eichberger <valentin@mayer-eichberger.de>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	bule "github.com/vale1410/bule/lib"
	"os"
	"strconv"
)

var (
	quantificationFlag bool
	withFactsFlag      bool
	textualFlag        bool
	constStringMap     map[string]int
)

// groundCmd represents the ground command
var groundCmd = &cobra.Command{
	Use:   "ground",
	Short: "Grounds to CNF from a program written in Bule format",
	Long: `Grounds to CNF from a program written in Bule format
How to prepare it:
bule ground <program.bul> [options].
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			return
		}

		bule.DebugLevel = debugFlag

		p, err := bule.ParseProgram(args)
		if err != nil {
			fmt.Println("Error parsing program")
			fmt.Println(err)
			os.Exit(1)
		}

		stage0Prerequisites(&p)
		stage1GeneratorsAndFacts(&p)
		stage2Iterators(&p)
		stage3Clauses(&p)
		stage4Printing(&p, args)
	},
}

func stage0Prerequisites(p *bule.Program) {
	for key, val := range constStringMap {
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
			p.ConstraintSimplification,
			"Do Fixpoint of TransformConstraintsToInstantiation.",
			"For each constraint (X==v) rewrite clause with (X<-v) and remove constraint.")

		stage(p, &changed,
			runFixpoint(p.ExpandGroundRanges),
			"ExpandGroundRanges",
			"p[1..2]. and also X==1..2, but not Y==A..B.")
		stage(p, &changed,
			p.ConstraintSimplification,
			"Do Fixpoint of TransformConstraintsToInstantiation.",
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
			"Example: a fact p(T1,T2) with tuples (v11,v12)..(vn2,vn1) occurs in clause, expand clause with (T1 == v11, T2 == v12).")

		stage(p, &changed,
			p.ConstraintSimplification,
			"Do Fixpoint of TransformConstraintsToInstantiation.",
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
		p.RemoveRulesWithGenerators,
		"RemoveRulesWithGeneratorsBecauseTheyHaveEmptyDomains",
		"Because they have empty domains, e.g. \n edge[_,_,V] => vertex[V]. %% there are no edges!")
}

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

		// Can this be omitted ? All tests run through successfully
///		stage(p, &changed,
///			p.ConstraintSimplification,
///			"ConstraintSimplification.",
///			"For each constraint (X==v) rewrite clause with (X<-v) and remove constraint.")

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
	stage(p, &changed,
		p.CollectExplicitTupleDefinitions,
		"CollectExplicitTupleDefinitions.",
		"#exist(3), p(1,2)? %% Then remove this rule.")

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

	debug(1, "Output")

	if textualFlag && withFactsFlag {
		p.PrintFacts()
	}

	if unitPropagationFlag || !textualFlag {
		//unitSlice := args[1:] \\TODO FIXME, currently turned off
		unitSlice := []string{}
		units := convertArgsToUnits(unitSlice)
		clauseProgram := translateFromRuleProgram(*p, units)
		sb := clauseProgram.StringBuilder()
		fmt.Println(sb.String())
	} else {
		p.Print()
	}
}

func stage(p *bule.Program, change *bool, f func() (bool, error), stage string, info string) {
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

func init() {
	rootCmd.AddCommand(groundCmd)
	groundCmd.PersistentFlags().BoolVarP(&quantificationFlag, "quant", "q", true, "Print Quantification")
	groundCmd.PersistentFlags().BoolVarP(&withFactsFlag, "facts", "f", false, "Output all facts.")
	groundCmd.PersistentFlags().BoolVarP(&textualFlag, "text", "t", false, "true: print grounded textual bule format. false: print dimacs format for QBF and SAT solvers.")
	groundCmd.PersistentFlags().BoolVarP(&printInfoFlag, "info", "i", true, "Print all units as well.")
	groundCmd.PersistentFlags().BoolVarP(&unitPropagationFlag, "up", "u", true, "Perform Unitpropagation.")
	groundCmd.PersistentFlags().StringToIntVarP(&constStringMap, "const", "c", map[string]int{}, "Comma separated list of constant instantiations: c=d.")
}
