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

		debug(2, "Expand Ranges")
		p.ExpandRanges()

		debug(2, "Collect Facts")
		p.CollectFacts()

		debug(2, "Find Facts until fixpoint reached, i.e. all facts have been found!")
		for p.FindNewFacts() {}

		debug(2, "Now rewrite all rules with facts!")
		p.InstanciateAndRemoveFacts()
		//p.Print()

		debug(2, "Now rewrite all rules with facts!")
		p.CleanRules()

		{
			debug(2, "Output All Facts:")
			for pred,_ := range p.GroundFacts {
				for _,tuple := range p.PredicatToTuples[pred] {
					fmt.Print(pred)
					for i,t := range  tuple {
						if i == 0 {
							fmt.Print("[")
						}
						fmt.Print(t)
						if i == len(tuple)-1 {
							fmt.Print("]")
						} else {
							fmt.Print(",")
						}
					}
					fmt.Println(".")
				}
			}
		}



		debug(2, "Expand Conditionals")
		p.ExpandConditionals()


		// forget about heads now!
		debug(2, "\nRewrite Equivalences")
		p.RewriteEquivalencesAndImplications()


		// There are no equivalences and no generators anymore !

		{
			debug(2, "Grounding:")
			clauses, existQ, forallQ, maxIndex := p.Ground()

			// Do Unit Propagation

			// Find variables that need to be put in the quantifier alternation

			for i := 0; i <= maxIndex; i++ {

				if atoms, ok := forallQ[i]; ok {
					fmt.Print("a")
					for _, a := range atoms {
						fmt.Print(" ", a)
					}
					fmt.Println()
				}
				if atoms, ok := existQ[i]; ok {
					fmt.Print("e")
					for _, a := range atoms {
						fmt.Print(" ", a)
					}
					fmt.Println()
				}
			}

			for _, clause := range clauses {
				for i, lit := range clause {
					fmt.Print(lit.String())
					if i < len(clause)  -1 {
						fmt.Print(", ")
					}
				}
				fmt.Println(".")
			}
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
