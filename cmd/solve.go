/*
Copyright © 2020 Valentin Mayer-Eichberger <valentin@mayer-eichberger.de>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	bule "github.com/vale1410/bule/lib"
)

//-----------------------------------------------------------------------------
// Command  logic
//-----------------------------------------------------------------------------

// Autocompletion
func autoCompleteBuleFiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"bul", "bule"}, cobra.ShellCompDirectiveFilterFileExt
}

func autoCompleteSolverInstance(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var s Solvers
	if err := s.load(); err != nil {
		ErrExit(err, 1)
	}
	labelsSat := make([]string, 0, len(s.Sat))
	labelsQbf := make([]string, 0, len(s.Qbf))
	labelsAll := make([]string, 0, cap(labelsSat)+cap(labelsQbf))
	// add Sat instances
	for label := range s.Sat {
		labelsSat = append(labelsSat, satPrefix+label)
	}
	// add Qbf instances
	for label := range s.Qbf {
		labelsQbf = append(labelsQbf, qbfPrefix+label)
	}
	// sort them and make default instance 1st
	sortSwap(&labelsSat, satPrefix+"default")
	sortSwap(&labelsQbf, qbfPrefix+"default")
	// merge sorted instances
	for _, label := range labelsSat {
		labelsAll = append(labelsAll, label)
	}
	for _, label := range labelsQbf {
		labelsAll = append(labelsAll, label)
	}
	return labelsAll, cobra.ShellCompDirectiveDefault
}

// Flags
const defaultInstance string = "default"
const satPrefix string = "[SAT]"
const qbfPrefix string = "[QBF]"

var (
	withInstance string
)

// solveCmd represents the solve command
var solveCmd = &cobra.Command{
	Use:   "solve",
	Short: "Grounds the bule formula and passes it to DEPQBF, then it outputs a model if it exists. ",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.
`,
	// ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// 	if len(args) != 0 {
	// 		return nil, cobra.ShellCompDirectiveNoFileComp
	// 	}
	// 	return []string{"foo", "bar", "baz"}, cobra.ShellCompDirectiveNoFileComp
	// },
	ValidArgsFunction: autoCompleteBuleFiles,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			return
		}

		var s Solvers
		if err := s.load(); err != nil {
			ErrExit(err, 1)
		}

		start := time.Now()
		bule.DebugLevel = debugFlag

		p, err := bule.ParseProgram(args)
		if err != nil {
			fmt.Println("Error parsing program")
			fmt.Println(err)
			os.Exit(1)
		}

		stage0Prerequisites(&p)
		stage1GeneratorsAndFacts(&p)
		stage2Iterators(&p)
		stage3Clauses(&p)

		unitSlice := []string{}
		units := convertArgsToUnits(unitSlice)
		clauseProgram := translateFromRuleProgram(p, units)
		tmpFolder := os.TempDir()
		timestamp := time.Now().Format("2006_01_02T15_04_05")
		outputGroundFile := filepath.Join(tmpFolder, "tmp_"+timestamp+".bul")

		{ // Output to tmp File
			file, err := os.Create(outputGroundFile)
			if err != nil {
				fmt.Println("Error creating file.", err)
				os.Exit(1)
			}
			defer file.Close()

			sb := clauseProgram.StringBuilder()
			_, err = io.WriteString(file, sb.String())
			if err != nil {
				fmt.Println("Error writing file.", err)
				os.Exit(1)
			}
			err = file.Sync()
			if err != nil {
				fmt.Println("Error flushing file.", err)
				os.Exit(1)
			}
		}

		debug(1, "Ground program writen to ", outputGroundFile)
		fmt.Printf("c program grounded in %s. Solving...\n", time.Since(start))

		var cmdOutput []byte

		// if user specified -w --with flag
		if withInstance != defaultInstance {
			// trim the helper prefix
			if strings.HasPrefix(withInstance, satPrefix) {
				withInstance = withInstance[len(satPrefix):]
			} else {
				if strings.HasPrefix(withInstance, qbfPrefix) {
					withInstance = withInstance[len(qbfPrefix):]
				}
			}
		}
		// TODO: make this implicitly chose right solver by .cnf output
		var si *SolverInstance
		if si, err = s.get(withInstance); err != nil {
			ErrExit(ErrNoSuchLabel, 1)
		} else {
			fmt.Printf(">>> Using %v solver instance: %s %s\n", si.Type, si.Prog, si.Flags)
		}
		execFlags := strings.Split(si.Flags, " ")
		_ = execFlags

		// Reason on SAT and QBF solver wrt. implicit problem type!!

		// IF its a SAT problem and it is called with a SAT solver -> remove e line

		// If it is a SAT problem called with a QBF solver -> fine -> same, give hint

		// If it is a QBF problem called with a SAT solver -> PROBLEM!!! ABORT

		isTrue := true
		{
			flagsSplit := strings.Fields(si.Flags)
			fmt.Println(si.Prog, flagsSplit, outputGroundFile)
			cmdOutput, err = exec.Command(si.Prog, append(flagsSplit, outputGroundFile)...).Output()
			if exitError, ok := err.(*exec.ExitError); ok {
				debug(1, "DEPQBF exist status:", exitError.ExitCode())
				if exitError.ExitCode() == 10 {
					isTrue = true
				} else if exitError.ExitCode() == 20 {
					isTrue = false
				} else {
					log.Println("exitError of DEPQBF is", exitError.ExitCode())
					log.Println("Error DEPQBF log:\n ", string(cmdOutput))
					log.Println("Omitting parsing because of error in solving: ", err)
					return
				}
			} else if err != nil {
				log.Println("Error DEPQBF log:\n ", string(cmdOutput))
				log.Println("Omitting parsing because of error in solving: ", err)
				return
			}
		}

		// Parse output and return result
		{
			debug(1, "Output by DEPQBF")
			scanner := bufio.NewScanner(strings.NewReader(string(cmdOutput)))
			result := []int{}
			for scanner.Scan() {
				s := scanner.Text()
				//fmt.Println(s)
				//if strings.HasPrefix(s, "s ") {
				//	continue
				//}
				if strings.HasPrefix(s, "V ") {
					fields := strings.Fields(s)
					v, err := strconv.Atoi(fields[1])
					if err != nil {
						log.Println("Error in parsing result: ", err)
						os.Exit(1)
					}
					result = append(result, v)
					continue
				}
			}
			reverseMap := map[int]string{}
			for k, v := range clauseProgram.idMap {
				reverseMap[v] = k
			}

			if isTrue {
				fmt.Println("TRUE")
				for _, id := range result {
					if id > 0 {
						fmt.Println(reverseMap[id])
					} else {
						fmt.Printf("~%s\n", reverseMap[-1*id])
					}
				}
				fmt.Println()
			} else {
				fmt.Println("FALSE")

			}
		}
	},
}

func init() {
	rootCmd.AddCommand(solveCmd)
	solveCmd.Flags().StringVarP(&withInstance, "with", "w", defaultInstance, "solve problem with particular solver instance")
	solveCmd.RegisterFlagCompletionFunc("with", autoCompleteSolverInstance)
}

// Utils
func sortSwap(sPtr *[]string, key string) {
	s := *sPtr
	sort.Strings(s)
	if i := sort.SearchStrings(s, key); i < len(s) {
		// index i is in the pre-sorted array, now check if it's exact
		if s[i] == key && i != 0 {
			// perform safe swap
			s[0], s[i] = key, s[0]
		}
	}
	*sPtr = s
}
