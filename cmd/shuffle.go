/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

var (
	identityFlag      bool
	polarityFlag      bool
	clauseShuffleFlag bool
	seedFlag          int64
)

// shuffleCmd represents the shuffle command
var shuffleCmd = &cobra.Command{
	Use:   "shuffle",
	Short: "Shuffles a CNF/QBF formula in DIMACS",
	Long: `Shuffle Ids, polarity and order of clauses.
		Input from file and output on stdout.`,

	Run: func(cmd *cobra.Command, args []string) {

		rand.Seed(seedFlag)

		var scanner *bufio.Scanner
		scanner = bufio.NewScanner(os.Stdin)

		//adjust the capacity to your need (max characters in line)
		const maxCapacity = 1024 * 1024
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		var perm []int
		var pol []int

		for scanner.Scan() {

			s := scanner.Text()

			if seedFlag < 1 || identityFlag { // dont shuffle !
				fmt.Println(s)
				continue
			}

			f := strings.Fields(s)

			if s == "" || strings.HasPrefix(s, "%") || strings.HasPrefix(s, "c") {
				fmt.Println(s)
				continue
			}

			if strings.HasPrefix(s, "p cnf") {
				nVars, _ := strconv.Atoi(f[2])
				perm = rand.Perm(nVars)
				pol = make([]int, nVars)
				for i, _ := range pol {
					pol[i] = 1 - 2*(rand.Int()%2)
				}
				fmt.Println(s)
				continue
			}

			if strings.HasPrefix(s, "e ") || strings.HasPrefix(s, "a ") {
				fmt.Print(f[0], " ")
				for _, x := range f[1 : len(f)-1] {
					y, _ := strconv.Atoi(x)
					fmt.Print(perm[y-1]+1, " ")
				}
				fmt.Println(0)
				continue
			}
			for _, x := range f[:len(f)-1] {
				y, _ := strconv.Atoi(x)
				if y < 0 {
					p := 1
					if polarityFlag {
						p = pol[-y-1]
					}
					fmt.Print(p*-(perm[-y-1]+1), " ")
				} else {
					p := 1
					if polarityFlag {
						p = pol[y-1]
					}
					fmt.Print(p*(perm[y-1]+1), " ")
				}
			}
			fmt.Println(0)
		}
		return
	},
}

func init() {
	rootCmd.AddCommand(shuffleCmd)
	shuffleCmd.PersistentFlags().BoolVarP(&identityFlag, "no", "", false, "dont shuffle and just pass identiy!")
	shuffleCmd.PersistentFlags().BoolVarP(&polarityFlag, "polarity", "p", false, "randomly change polarity.")
	shuffleCmd.PersistentFlags().BoolVarP(&clauseShuffleFlag, "order of clauses", "o", false, "Order sequence of clauses.")
	shuffleCmd.PersistentFlags().Int64VarP(&seedFlag, "seed", "s", 0, "random seed initializer. All seed value < 1 is interpreted as identity function (no shuffling)")
}
