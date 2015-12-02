package glob

import (
	"flag"
	"math"
	"os"
)

// global configuration accessible from everywhere

var (
	Ver                   = flag.Bool("ver", false, "Show version info.")
	Debug_file            *os.File
	Debug_filename        = flag.String("df", "", "File to print debug information.")
	Debug_flag            = flag.Bool("d", false, "Print debug information.")
	Filename_flag         = flag.String("f", "test.pb", "Path of to PB file.")
	Cnf_tmp_flag          = flag.String("out", "", "If set: output cnf to this file.")
	Pbo_flag              = flag.Bool("pbo", false, "Reformat to pbo format, output to stdout.")
	Gringo_flag           = flag.Bool("gringo", false, "Reformat to Gringo format, output to stdout.")
	Gurobi_flag           = flag.Bool("gurobi", false, "Reformat to Gurobi input, output to stdout.")
	Solve_flag            = flag.Bool("solve", true, "Dont solve just categorize and analyze the constraints.")
	Dimacs_flag           = flag.Bool("dimacs", false, "Print readable format of clauses.")
	Stat_flag             = flag.Bool("stat", false, "Extended statistics on types of PBs in problem.")
	Cat_flag              = flag.Int("cat", 2, "Categorize method 1, or 2. (default 2, historic: 1).")
	Complex_flag          = flag.String("complex", "hybrid", "Solve complex PBs with mdd/sn/hybrid. Default is hybrid")
	Timeout_flag          = flag.Int("timeout", 600, "Timeout of the overall solving process")
	MDD_max_flag          = flag.Int("mdd-max", 2000000, "Maximal number of MDD Nodes in processing one PB.")
	MDD_redundant_flag    = flag.Bool("mdd-redundant", true, "Reduce MDD by redundant nodes.")
	Opt_bound_flag        = flag.Int64("opt-bound", math.MaxInt64, "Initial bound for optimization function <= given value. Negative values allowed.")
	Opt_half_flag         = flag.Bool("opt-half", false, "Sets opt-bound to half the sum of the weights of the optimization function.")
	Solver_flag           = flag.String("solver", "minisat", "Choose Solver: minisat/clasp/lingeling/glucose/CCandr/cmsat.")
	Seed_flag             = flag.Int64("seed", 42, "Random seed initializer.")
	Amo_reuse_flag        = flag.Bool("amo-reuse", false, "Reuses AMO constraints for rewriting complex PBs.")
	Rewrite_opt_flag      = flag.Bool("opt-rewrite", true, "Rewrites opt with chains from AMO and other constraint.")
	Rewrite_same_flag     = flag.Bool("rewrite-same", false, "Groups same coefficients and introduces sorter and chains for them.")
	Rewrite_equal_flag    = flag.Bool("rewrite-equal", false, "Rewrites complex == constraints into >= and <=.")
	Ex_chain_flag         = flag.Bool("ex-chain", false, "Rewrites PBs with matching EXK constraints.")
	Amo_chain_flag        = flag.Bool("amo-chain", true, "Rewrites PBs with matching AMO.")
	Search_strategy_flag  = flag.String("search", "iterative", "Search objective iterative or binary.")
	Len_rewrite_same_flag = flag.Int("len-rewrite-same", 3, "Min length to rewrite PB.")
	Len_rewrite_amo_flag  = flag.Int("len-rewrite-amo", 3, "Min length to rewrite PB.")
	Len_rewrite_ex_flag   = flag.Int("len-rewrite-ex", 3, "Min length to rewrite PB.")
)
