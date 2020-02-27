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
	progFlag  string //= flag.String("f", "", "Path to file.")
)

func debug(level int, s ...interface{}) {
	if level <= debugFlag {
		fmt.Println()
		fmt.Print("%d: ")
		fmt.Print(s...)
		fmt.Println()
	}
}

// groundCmd represents the ground command
var groundCmd = &cobra.Command{
	Use:   "ground",
	Short: "Grounds to CNF from a program written in Bule format",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Bule is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		debug(2, "Bule started")
		p := bule.ParseProgram(progFlag)
		bule.DebugLevel = debugFlag

		debug(2, "Replace Constants")
		p.ReplaceConstants()

		debug(2, "Expand ground Ranges in literals.")
		p.ExpandRanges()

		debug(2, "Collect ground facts.")
		p.CollectGroundFacts()

		debug(2, "Find clauses where all literals but 1 are facts. Resolve, add to tuples of fact and remove.")
		for p.FindNewFacts() {}

		// Now there should be no clauses entirely of facts!

		p.PrintDebug(2)
		debug(2, "If a fact p(T1,T2) with tuples (v11,v12)..(vn2,vn1) occurs in clause, expand clause with (T1 == v11, T2 == v12).")
		for p.InstanciateAndRemoveFacts() {}
		p.PrintDebug(2)

		debug(2, "For each constraint (X==v) rewrite clause with (X<-v) and remove constraint.")
		for p.TransformConstraintsToInstantiation() {}
		p.PrintDebug(2)

		debug(2, "Remove clauses with contradictions (1==2) and remove true constraints (1>2, 1==1).")
		p.CleanRules()

		debug(2, "Expand Conditionals")
		p.ExpandConditionals()

		debug(2, "All Rules should be clauses with search predicates. No more ground facts.")
		p.PrintDebug(2)

		debug(2, "Collect all ground literals in all clauses")
		p.CollectGroundTuples()
//		p.Debug()

		debug(2, "Collect all ground literals in all clauses")

		p.GroundFromTuples() //

		p.Print()





		// // forget about heads now!
		// debug(2, "\nRewrite Equivalences")
		// p.RewriteEquivalencesAndImplications()


		// There are no equivalences and no generators anymore !

		{
			// debug(2, "Grounding:")
			// clauses, existQ, forallQ, maxIndex := p.Ground()

			// Do Unit Propagation

			// Find variables that need to be put in the quantifier alternation

			// for i := 0; i <= maxIndex; i++ {

			// 	if atoms, ok := forallQ[i]; ok {
			// 		fmt.Print("a")
			// 		for _, a := range atoms {
			// 			fmt.Print(" ", a)
			// 		}
			// 		fmt.Println()
			// 	}
			// 	if atoms, ok := existQ[i]; ok {
			// 		fmt.Print("e")
			// 		for _, a := range atoms {
			// 			fmt.Print(" ", a)
			// 		}
			// 		fmt.Println()
			// 	}
			// }
//
//			for _, clause := range clauses {
//				for i, lit := range clause {
//					fmt.Print(lit.String())
//					if i < len(clause)  -1 {
//						fmt.Print(", ")
//					}
//				}
//				fmt.Println(".")
//			}
		}
	},
}

func init() {
	rootCmd.AddCommand(groundCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	//debugFlag int    //= flag.Int("d", 0, "Debug Level .")
	//progFlag  string //= flag.String("f", "", "Path to file.")
	groundCmd.PersistentFlags().StringVarP(&progFlag, "file", "f", "", "Path to File")
	groundCmd.PersistentFlags().IntVarP(&debugFlag, "debug", "d", 0, "Debug level")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// groundCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
