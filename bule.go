package main

import (
	"fmt"
	"io/ioutil"
	"bytes"
    "bule_lib"
)

func main() {
	fmt.Printf("This is bule (pronounced Boul-ee), the state of the art CNF grounder!\n")
	fileinput, err := ioutil.ReadFile("test.lp")
	input := &Input{Name: "test", Content: fileinput}


	if err != nil {
		fmt.Printf("problem...:\n ")
	}
	fmt.Printf(input.Name + "\n")
    b :=  bytes.NewBuffer(input.Content)
    fmt.Printf(b.String())

}
