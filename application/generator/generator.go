package main

import (
	"flag"
	"fmt"
)

var number = flag.Int("n", 1, "Number of cliques.")
var clique = flag.Int("clique", 2, "Size of cliques.")

func main() {
	flag.Parse()

	edges := ((*clique * (*clique - 1)) / 2) * *number
	fmt.Println("p edges", edges, *number**clique)

	for i := 0; i < *number; i++ {
		printClique(*clique, i * *clique)
	}

}

func printClique(n int, offset int) {

	for i := 1; i <= n; i++ {
		for j := i + 1; j <= n; j++ {
			fmt.Println("e", i+offset, j+offset)
		}
	}
}


