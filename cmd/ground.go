/*
Copyright © 2021 Valentin Mayer-Eichberger <valentin@mayer-eichberger.de>

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
	bule "github.com/vale1410/bule/grounder"
	"os"
)

var (
	quantificationFlag bool
	withFactsFlag      bool
	textualFlag        bool
	constStringMapFlag map[string]int
	printInfoFlag      bool
)

func init() {
	rootCmd.AddCommand(groundCmd)
	groundCmd.PersistentFlags().BoolVarP(&quantificationFlag, "quant", "q", true, "Print Quantification")
	groundCmd.PersistentFlags().BoolVarP(&withFactsFlag, "facts", "f", true, "Output all facts.")
	groundCmd.PersistentFlags().BoolVarP(&textualFlag, "text", "t", true, "true: print grounded textual bule format. false: print dimacs format for QBF and SAT solvers.")
	groundCmd.PersistentFlags().StringToIntVarP(&constStringMapFlag, "const", "c", map[string]int{}, "Comma separated list of constant instantiations: c=d.")
}

// groundCmd represents the ground command
var groundCmd = &cobra.Command{
	Use:   "ground",
	Short: "Grounds to CNF from a program written in Bule format",
	Long: `Grounds a textual program written in Bule.
bule ground <program> [options].
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
