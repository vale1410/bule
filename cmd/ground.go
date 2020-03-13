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
)

var (
	debugFlag int    //= flag.Int("d", 0, "Debug Level .")
)

func debug(level int, s ...interface{}) {
	if level <= debugFlag {
		fmt.Print("% ")
		fmt.Print(s...)
		fmt.Println()
	}
}

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

		debug(2, "Bule started")
		p := bule.ParseProgram(args[0])
		bule.DebugLevel = debugFlag

		debug(1, "Input")
		p.PrintDebug(1)
		debug(1,"Output")

		debug(2, "Replace Constants (#const a=3. and Function Symbols (#mod)")
		p.ReplaceConstantsAndMathFunctions()

		debug(2, "Expand ground Ranges in literals.")
		for p.ExpandRanges() {}

		debug(2, "Collect ground facts.")
		p.CollectGroundFacts()

		debug(2, "Find clauses where all literals but 1 are facts. Resolve, add to tuples of fact and remove.")
		for p.FindNewFacts() {}

		// Now there should be no clauses entirely of facts!

		p.PrintDebug(2)
		debug(2, "If a fact p(T1,T2) with tuples (v11,v12)..(vn2,vn1) occurs in clause, expand clause with (T1 == v11, T2 == v12).")
		for p.InstantiateAndRemoveFacts() {}
		p.PrintDebug(2)

		debug(2, "Fixpoint of TransformConstraintsToInstantiation.")
		debug(2, "For each constraint (X==v) rewrite clause with (X<-v) and remove constraint.")
		for p.TransformConstraintsToInstantiation() {}

		p.PrintDebug(2)
		debug(2, "Remove clauses with contradictions (1==2) and remove true constraints (1>2, 1==1).")
		p.CleanRules()

		debug(2, "Expand Conditionals")
		p.ExpandConditionals()

		debug(2, "All Rules should be clauses with search predicates. No more ground facts.")

		debug(2, "Collect all ground literals in all clauses")
		p.CollectGroundTuples()
		//p.PrintTuples()

		debug(2, "Collect all ground literals in all clauses")
		p.GroundFromTuples()

		debug(2, "Extract Quantors ")

//		debug(2, "Print Quantification")
//		p.ExtractQuantors()
//		p.PrintQuantification()

		p.Print()


	},
}

func init() {
	rootCmd.AddCommand(groundCmd)

	//debugFlag int    //= flag.Int("d", 0, "Debug Level .")
	//progFlag  string //= flag.String("f", "", "Path to file.")
//	groundCmd.PersistentFlags().StringVarP(&progFlag, "file", "f", "", "Path to File")
	groundCmd.PersistentFlags().IntVarP(&debugFlag, "debug", "d", 0, "Debug level")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// groundCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
