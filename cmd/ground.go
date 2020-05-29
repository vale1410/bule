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
How to run it: 
bule ground <program.bul> [options].
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			return
		}

		unitSlice := args[1:]

		bule.DebugLevel = debugFlag

		p, err := bule.ParseProgram(args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for key, val := range constStringMap {
			p.Constants[key] = val
		}

		debug(1, "Input:")
		p.PrintDebug(1)

		{
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

		stageInfo(&p, "Replace Constants and Math","(#const a=3. and Function Symbols (#mod)")
		p.ReplaceConstantsAndMathFunctions()

		{
			stageInfo(&p, "CheckUnboundVariables","Check for unbound variables that are not marked as such.")
			err := p.CheckUnboundVariables()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		{
			stageInfo(&p, "Do Fixpoint of TransformConstraintsToInstantiation.", ""+
				"For each constraint (X==v) rewrite clause with (X<-v) and remove constraint.")
			p.ConstraintSimplification()

			stageInfo(&p, "ExpandGroundRanges", "p[1..2]. and also X==1..2, but not Y==A..B.")
			p.ExpandGroundRanges()

			stageInfo(&p, "Do Fixpoint of TransformConstraintsToInstantiation.", "")
			p.ConstraintSimplification()

			stageInfo(&p, "CollectGroundFacts", "p[1,2]. r[1]. but not p[1],p[2]. and also not p[X,X], or p[1,X].")
			p.CollectGroundFacts()

			//		stageInfo(&p, "FindNewFacts()","Find clauses where all literals but 1 are facts. Resolve, add to tuples of fact and remove.")
			//		p.FindNewFacts()

			debug(2, "Now there should be no clauses entirely of facts!")

			stageInfo(&p, "InstantiateAndRemoveFactFromGenerator", " If a fact p(T1,T2) with tuples (v11,v12)..(vn2,vn1) occurs in clause, expand clause with (T1 == v11, T2 == v12).")
			p.InstantiateAndRemoveFactFromGenerator()

			debug(2, "Program is now fact free in all clauses!")

			stageInfo(&p, "Do Fixpoint of TransformConstraintsToInstantiation.", "")
			p.ConstraintSimplification()

			stageInfo(&p, "ExpandIterators", "")
			p.InstantiateAndRemoveFactFromIterator()

			stageInfo(&p, "RemoveClausesWithFacts", "")
			p.RemoveClausesWithFacts()
		}

		{
			err := p.CheckNoRemainingFacts()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		stageInfo(&p,"CollectGroundTuples","")
		err := p.CollectGroundTuples()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		{
			stageInfo(&p, "Ground non-Ground Lits", "Ground from all tuples the non-ground literals, until fixpoint.")
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
				stageInfo(&p, "Do Fixpoint of TransformConstraintsToInstantiation.", ""+
					"For each constraint (X==v) rewrite clause with (X<-v) and remove constraint.")
				p.ConstraintSimplification()

				stageInfo(&p, "RemoveClausesWithTuplesThatDontExist.", "")
				p.RemoveClausesWithTuplesThatDontExist()
			}
		}

		if quantificationFlag {
			stageInfo(&p,"Extract Quantors","")
			p.ExtractQuantors()
			stageInfo(&p,"Merge Quantification Levels","")
			p.MergeConsecutiveQuantificationLevels()
			debug(2, "Merged alternations:", p.Alternation)
		}

		debug(1, "Output")

		if textualFlag && withFactsFlag {
			p.PrintFacts()
		}

		if unitPropagationFlag || !textualFlag {
			units := convertArgsToUnits(unitSlice)
			clauseProgram := translateFromRuleProgram(p, units)
			clauseProgram.Print()
		} else {
			p.Print()
		}
	},
}

func runFixpoint(f func() (bool, error)) {
	ok := true
	var err error
	for ok {
		ok, err = f()
		if err != nil {
			fmt.Println("Error occurred in grounding.\n %w", err)
			os.Exit(1)
		}
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
