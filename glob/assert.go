package glob

import (
	"flag"
	"fmt"
)

func Init() {
	flag.Parse()
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

			A(Debug_file != nil)

			ss += "\n"
			if _, err := Debug_file.Write([]byte(ss)); err != nil {
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
