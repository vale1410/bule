package constraints

import (
	"fmt"
	"testing"
)

func TestRemoveZeros(test *testing.T) {
	fmt.Println("TestRemoveZeros")

	pb1 := CreatePB([]int64{1, 2, 3, 0, 321, 0, 0, -123, 0}, 1347)
	c := len(pb1.Entries)
	pb1.RemoveZeros()

	if len(pb1.Entries) != c-4 {
		test.Fail()
	}
}

func TestCleanChain(test *testing.T) {
	fmt.Println("TestCleanChain")

	pb := CreatePB([]int64{1, 2, 3, 0, 321, 0, 1, -123, 0}, 1347)
	results := []int{3, 3, 2, 0, 3, 0}

	chain := pb.Literals()

	pb.RemoveZeros()

	for i := 0; i < len(chain)-4; i++ {
		c1 := chain[i : i+4]
		c2 := CleanChain(pb.Entries, c1)
		if results[i] != len(c2) {
			fmt.Println(results[i], len(c2))
			test.Fail()
		}
	}

}
