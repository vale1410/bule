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

func PringConfig() {
	fmt.Println("Configuration")
	fmt.Println("Filename_flag :\t", Filename_flag)
	fmt.Println("Debug_flag :\t", Debug_flag)
	fmt.Println("Complex_flag :\t", Complex_flag)
	fmt.Println("Timeout_flag :\t", Timeout_flag)
	fmt.Println("MDD_max_flag :\t", MDD_max_flag)
	fmt.Println("MDD_redundant_flag :\t", MDD_redundant_flag)
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
