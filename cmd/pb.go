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
	"github.com/spf13/cobra"
	"math"

	"flag"
	"fmt"
	"os"

	"github.com/vale1410/bule/constraints"
	"github.com/vale1410/bule/glob"
	"github.com/vale1410/bule/parser"
	"github.com/vale1410/bule/sat"
)

// pbCmd represents the pb command
var pbCmd = &cobra.Command{
	Use:   "pb",
	Short: "A CNF Eager Pseudo Boolean Constraint Solver",
	Long: `This is essentially the Pseudo Boolean Solver Bule 1.0.
`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) >= 2 {
			fmt.Println("Command line flags not recognized", flag.Args())
			return
		}

		if len(args) == 1 {
			glob.Filename_flag = flag.Args()[0]
		}

		if glob.Debug_filename != "" {

			var err error
			glob.Debug_file, err = os.Create(glob.Debug_filename)

			if err != nil {
				panic(err)
			}
			defer glob.Debug_file.Close()
		}

		glob.D("Running Debug Mode...")

		problem := parser.New(glob.Filename_flag)

		if glob.Pbo_flag {
			problem.PrintPBO()
			return
		}

		if glob.Gringo_flag {
			problem.PrintGringo()
			return
		}

		if glob.Gurobi_flag {
			problem.PrintGurobi()
			return
		}

		pbs := problem.Pbs[1:] // opt is just a pointer to first in pbs.
		opt := problem.Opt

		primaryVars := make(map[string]bool, 0)

		for i, _ := range pbs {
			for _, x := range pbs[i].Entries {
				primaryVars[x.Literal.A.Id()] = true
			}
		}

		var clauses sat.ClauseSet

		// Categorize Version 1 (deprecated)
		switch glob.Cat_flag {
		case 1:
			{
				for _, pb := range pbs {
					pb.Print10()
					pb.CategorizeTranslate1()
					clauses.AddClauseSet(pb.Clauses)
				}
			}
		case 2:
			{
				constraints.CategorizeTranslate2(pbs)
				for _, pb := range pbs {
					clauses.AddClauseSet(pb.Clauses)
				}
			}
		default:
			panic("Category not implemented")
		}

		if glob.Dimacs_flag {
			clauses.PrintDebug()
		}

		if glob.Solve_flag {
			g := sat.IdGenerator(clauses.Size()*7 + 1)
			g.PrimaryVars = primaryVars
			opt.NormalizePositiveCoefficients()
			opt.Offset = opt.K
			glob.A(opt.Positive(), "opt only has positive coefficients")
			g.Solve(clauses, opt, glob.Opt_bound_flag, -opt.Offset)
			//fmt.Println()
		}

	},
}

func printStats(stats []int) {

	glob.A(len(stats) == int(constraints.TranslationTypes), "Stats for translation errornous")

	trans := constraints.Facts
	fmt.Print("Name;")
	for i := trans; i < constraints.TranslationTypes; i++ {
		if i > 0 {
			fmt.Printf("%v;", constraints.TranslationType(i))
		}
	}
	fmt.Println()
	fmt.Print(glob.Filename_flag, ";")
	for i := trans; i < constraints.TranslationTypes; i++ {
		if i > 0 {
			fmt.Printf("%v;", stats[i])
		}
	}
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(pbCmd)

	pbCmd.Flags().BoolVarP(&glob.Debug_flag, "debug", "", false, "Print debug information.")
	pbCmd.Flags().StringVarP(&glob.Debug_filename, "debugFile", "", "", "File to output debug information. If empty, then stdout.")
	pbCmd.Flags().StringVarP(&glob.Filename_flag, "file", "f", "test.pb", "Path of to PB file.")
	pbCmd.Flags().StringVarP(&glob.Cnf_tmp_flag, "out", "o", "", "If set: output cnf to this file.")
	pbCmd.Flags().StringVarP(&glob.Complex_flag, "complex", "", "hybrid", "Solve complex PBs with mdd/sn/hybrid. Default is hybrid")
	pbCmd.Flags().StringVarP(&glob.Solver_flag, "solver", "", "clasp", "Choose Solver: minisat/clasp/lingeling/glucose/CCandr/cmsat.")
	pbCmd.Flags().StringVarP(&glob.Search_strategy_flag, "search", "", "iterative", "Search objective iterative or binary.")
	pbCmd.Flags().BoolVarP(&glob.Pbo_flag, "pbo", "", false, "Reformat to pbo format, output to stdout.")
	pbCmd.Flags().BoolVarP(&glob.Gringo_flag, "gringo", "", false, "Reformat to Gringo format, output to stdout.")
	pbCmd.Flags().BoolVarP(&glob.Gurobi_flag, "gurobi", "", false, "Reformat to Gurobi input, output to stdout.")
	pbCmd.Flags().BoolVarP(&glob.Solve_flag, "solve", "", true, "Dont solve just categorize and analyze the constraints.")
	pbCmd.Flags().BoolVarP(&glob.Dimacs_flag, "dimacs", "", false, "Print readable format of clauses.")
	pbCmd.Flags().BoolVarP(&glob.Stat_flag, "stat", "", false, "Extended statistics on types of PBs in problem.")
	pbCmd.Flags().BoolVarP(&glob.Amo_reuse_flag, "amo-reuse", "", false, "Reuses AMO constraints for rewriting complex PBs.")
	pbCmd.Flags().BoolVarP(&glob.Rewrite_opt_flag, "opt-rewrite", "", true, "Rewrites opt with chains from AMO and other constraint.")
	pbCmd.Flags().BoolVarP(&glob.Rewrite_same_flag, "rewrite-same", "", false, "Groups same coefficients and introduces sorter and chains for them.")
	pbCmd.Flags().BoolVarP(&glob.Rewrite_equal_flag, "rewrite-equal", "", false, "Rewrites complex == constraints into >= and <=.")
	pbCmd.Flags().BoolVarP(&glob.Ex_chain_flag, "ex-chain", "", false, "Rewrites PBs with matching EXK constraints.")
	pbCmd.Flags().BoolVarP(&glob.Amo_chain_flag, "amo-chain", "", false, "Rewrites PBs with matching AMO. (buggy)")
	pbCmd.Flags().BoolVarP(&glob.Infer_var_ids, "infer-ids", "", false, "Tries to infer SAT variables Ids by convention v<id>.")
	pbCmd.Flags().BoolVarP(&glob.MDD_redundant_flag, "mdd-redundant", "", true, "Reduce MDD by redundant nodes.")
	pbCmd.Flags().BoolVarP(&glob.Opt_half_flag, "opt-half", "", false, "Sets opt-bound to half the sum of the weights of the optimization function.")
	pbCmd.Flags().IntVarP(&glob.Len_rewrite_same_flag, "len-rewrite-same", "", 3, "Min length to rewrite PB.")
	pbCmd.Flags().IntVarP(&glob.Len_rewrite_amo_flag, "len-rewrite-amo", "", 3, "Min length to rewrite PB.")
	pbCmd.Flags().IntVarP(&glob.Len_rewrite_ex_flag, "len-rewrite-ex", "", 3, "Min length to rewrite PB.")
	pbCmd.Flags().IntVarP(&glob.First_aux_id_flag, "aux-id", "", 1, "Set initial variable counter for auxiliary Ids.")
	pbCmd.Flags().IntVarP(&glob.Cat_flag, "cat", "", 2, "Categorize method 1, or 2. (default 2, historic: 1).")
	pbCmd.Flags().IntVarP(&glob.Timeout_flag, "timeout", "", 600, "Timeout of the overall solving process")
	pbCmd.Flags().IntVarP(&glob.MDD_max_flag, "mdd-max", "", 2000000, "Maximal number of MDD Nodes in processing one PB.")
	pbCmd.Flags().Int64VarP(&glob.Opt_bound_flag, "opt-bound", "", math.MaxInt64, "Initial bound for optimization function <= given value. Negative values allowed. Default is largest upper bound.")
	pbCmd.Flags().Int64VarP(&glob.Seed_flag, "seed", "", 42, "Random seed initializer.")
}
