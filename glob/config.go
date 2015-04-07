package glob

import (
	"fmt"
	"os"
)

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

var Opt_rewrite_flag bool
var Amo_reuse_flag bool
var Rewrite_same_flag bool
var Ex_chain_flag bool
var Amo_chain_flag bool
var Opt_bound_flag int64
var Cnf_tmp_flag string

const Len_rewrite_flag = 4

func PringConfig() {
	fmt.Println("Configuration")
	fmt.Println("Filename_flag :\t", Filename_flag)
	fmt.Println("Debug_flag :\t", Debug_flag)
	fmt.Println("Complex_flag :\t", Complex_flag)
	fmt.Println("Timeout_flag :\t", Timeout_flag)
	fmt.Println("MDD_max_flag :\t", MDD_max_flag)
	fmt.Println("MDD_redundant_flag :\t", MDD_redundant_flag)
	fmt.Println("Solver_flag :\t", Solver_flag)
	fmt.Println("Seed_flag :\t", Seed_flag)
	fmt.Println("Opt_rewrite_flag :\t", Opt_rewrite_flag)
	fmt.Println("Amo_reuse_flag :\t", Amo_reuse_flag)
	fmt.Println("Rewrite_same_flag :\t", Rewrite_same_flag)
	fmt.Println("Ex_chain_flag bool :\t", Ex_chain_flag)
	fmt.Println("Amo_chain_flag bool : \t", Amo_chain_flag)
	fmt.Println("Opt_bound_flag bool : \t", Opt_bound_flag)
}

func D(arg ...interface{}) {
	if Debug_flag {
		if Debug_filename == "" {
			for _, s := range arg {
				fmt.Print(s, " ")
			}
			fmt.Println()
		} else {
			ss := ""
			for _, s := range arg {
				ss += fmt.Sprintf("%v", s) + " "
			}
			ss += "\n"
			if _, err := Debug_output.Write([]byte(ss)); err != nil {
				panic(err)
			}
		}
	}
}

// An assert function
func A(check bool, arg ...interface{}) {
	if !check {
		fmt.Print("ASSERT FAILED: ")
		for _, s := range arg {
			fmt.Print(s, " ")
		}
		fmt.Println()
		panic(" ")
	}
}

func DT(check bool, arg ...interface{}) {
	if check {
		D(arg)
	}
}
