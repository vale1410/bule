package bule_lib

import (
    "fmt"
    "testing"
)


func TestBule(t *testing.T) {

    input,_ := NewInput("test2.lp");
    fmt.Println("string parsed and chopped into pieces:")
    for _,s := range input.Lines {
        fmt.Println(s)
    }
    fmt.Println("\n\n")
    problem := ParseLines(&input)
    fmt.Print(problem.abox.String())
    fmt.Print(problem.tbox)
}

