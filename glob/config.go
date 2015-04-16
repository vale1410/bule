package glob

import "os"

// global configuration accessible from everywhere

var Filename_flag string
var Debug_output *os.File
var Debug_filename string
var Debug_flag bool
var Complex_flag string
var Timeout_flag int
var MDD_max_flag int
var MDD_redundant_flag bool
var Solver_flag string
var Seed_flag int64

var Amo_reuse_flag bool
var Rewrite_opt_flag bool
var Rewrite_same_flag bool
var Ex_chain_flag bool
var Amo_chain_flag bool
var Opt_bound_flag int64
var Cnf_tmp_flag string
var Search_strategy_flag string

const Len_rewrite_same_flag = 3
const Len_rewrite_amo_flag = 3
const Len_rewrite_ex_flag = 3
