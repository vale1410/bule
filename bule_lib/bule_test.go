package bule_lib

import (
    "fmt"
    "testing"
)


func TestBule(t *testing.T) {

    input,_ := NewInput("test.lp");
    for _,s := range input.Lines {
        fmt.Println(s)
    }
    atoms := ParseAtoms(&input)
    fmt.Println(atoms)
}

