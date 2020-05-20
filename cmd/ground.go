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
	dimacsFlag         bool
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

		bule.DebugLevel = debugFlag

		p := bule.ParseProgram(args)

		debug(1, "Input")
		p.PrintDebug(1)

		{
			err := p.CheckArityOfLiterals()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		{
			err := p.CheckFactsInGenerators()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		debug(2, "Replace Constants (#const a=3. and Function Symbols (#mod)")
		p.ReplaceConstantsAndMathFunctions()


		{
			debug(2, "Check for unbound variables that are not marked as such.")
			err := p.CheckUnboundVariables()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		p.ConstraintSimplification()

		debug(2, "ExpandGroundRanges:\n p[1..2]. and also X==1..2, but not Y==A..B.")
		runFixpoint(p.ExpandGroundRanges)

		debug(2, "CollectGroundFacts:\n p[1,2]. r[1]. but not p[1],p[2]. and also not p[X,X], or p[1,X].")
		p.CollectGroundFacts()

		debug(2, "FindNewFacts():\nFind clauses where all literals but 1 are facts. Resolve, add to tuples of fact and remove.")
		runFixpoint(p.FindNewFacts)

		debug(2, "Now there should be no clauses entirely of facts!")

		p.PrintDebug(2)
		debug(2, "InstantiateAndRemoveFacts: If a fact p(T1,T2) with tuples (v11,v12)..(vn2,vn1) occurs in clause, expand clause with (T1 == v11, T2 == v12).")
		runFixpoint(p.InstantiateAndRemoveFacts)

		debug(2, "Program is now fact free in all clauses!")
		p.PrintDebug(2)

		p.ConstraintSimplification()

		p.PrintDebug(2)

		debug(2, "Expand Conditionals")
		p.ExpandConditionals()

		p.PrintDebug(2)

		debug(2, "All Rules should be clauses with search predicates. No more ground facts.")

		p.RemoveClausesWithFacts()

		{
			err := p.CheckNoRemainingFacts()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		debug(2, "Collect all ground literals in all clauses")
		p.CollectGroundTuples()
		//		p.PrintTuples()

		p.PrintDebug(2)

		debug(2, "Ground from all tuples the non-ground literals, until fixpoint.")
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
			p.PrintDebug(2)
			p.ConstraintSimplification()
			p.PrintDebug(2)
			debug(2, "RemoveClausesWithTuplesThatDontExist.")
			p.RemoveClausesWithTuplesThatDontExist()
			p.PrintDebug(2)
		}

		if quantificationFlag {
			debug(2, "Extract Quantors ")
			p.ExtractQuantors()
			debug(2, "Merge Quantification Levels")
			p.MergeConsecutiveQuantificationLevels()
			debug(2, "Merged alternations:", p.Alternation)
		}

		debug(1, "Output")

		if dimacsFlag {
			clauseProgram := translateFromRuleProgram(p)
			clauseProgram.Print()
		} else {
			p.Print(withFactsFlag)
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

func init() {
	rootCmd.AddCommand(groundCmd)

	//	groundCmd.PersistentFlags().StringVarP(&progFlag, "file", "f", "", "Path to File")
	groundCmd.PersistentFlags().BoolVarP(&quantificationFlag, "quant", "q", true, "Print Quantification")
	groundCmd.PersistentFlags().BoolVarP(&withFactsFlag, "facts", "f", false, "Output all facts.")
	groundCmd.PersistentFlags().BoolVarP(&dimacsFlag, "dimacs", "D", true, "false: print bule format. true: print dimacs format")
	groundCmd.PersistentFlags().BoolVarP(&printInfoFlag, "info", "i", true, "Print all units as well.")
	groundCmd.PersistentFlags().BoolVarP(&unitPropagationFlag, "up", "u", true, "Perform Unitpropagation.")
}
