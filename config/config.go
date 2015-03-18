package config

import (
	"fmt"
)

// global configuration accessible from everywhere

var Complex_flag string
var Timeout_flag int
var MaxMDD_flag int

func PringConfig() {
	fmt.Println("Configuration")
	fmt.Println("Complex_flag :\t", Complex_flag)
	fmt.Println("Timeout_flag :\t", Timeout_flag)
	fmt.Println("MaxMDD:\t", MaxMDD_flag)
}

// These constants are for future implementations

// MDD translation by which clauses? implications over branchens? @ignasis implementation

// Use of what type of sorting networks:

// Use Mergers

var SortersT int
var EquationT int
