package main

import (
	"flag"
	"math/rand"

	"github.com/vale1410/bule/constraints"
)

var number = flag.Int("n", 5, "Number of PBs.")
var length = flag.Int("l", 5, "Length of PBs.")
var domain = flag.Int64("dom", 100, "Domain of Coefficients.")
var threshold = flag.Int64("t", 50, "Threshold is fraction of sum of coefficients (x/100).")

func main() {
	flag.Parse()

	weights := make([]int64, *length)
	for i := 0; i < *length; i++ {
		weights[i] = 1
	}
	opt := constraints.CreatePBOffset(1, weights, 1)
	opt.Typ = constraints.OPT
	opt.PrintPBO()

	for k := 0; k < *number; k++ {
		K := int64(0)
		for i := 0; i < *length; i++ {
			weights[i] = ((*domain) / 2) - rand.Int63n(*domain)
			K += weights[i]
		}
		pb := constraints.CreatePBOffset(1, weights, (K * (*threshold) / 100))
		pb.PrintPBO()
	}

}
