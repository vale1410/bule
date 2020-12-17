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

const defaultInstance string = "default"

// Flags
var (
	withInstance string
)

// solveCmd represents the solve command
var solveCmd = &cobra.Command{
	Use:   "solve",
	Short: "Grounds the bule formula and passes it to a solver instance, then it outputs a model if it exists. ",
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
		if err := s.load(); !err.isNil() {
			BuleExit(os.Stderr, err)
		}

		start := time.Now()
		bule.DebugLevel = debugFlag

		p, err := bule.ParseProgram(args)
		if err != nil {
			BuleExit(os.Stderr, errParse)
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
			dimacsOut := dimacsTidyUp(sb.String(), p.IsSATProblem())

			_, err = io.WriteString(file, dimacsOut)
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
		fmt.Fprintf(os.Stderr, "C program grounded in %s.\n", time.Since(start))

		var cmdOutput []byte
		var si *SolverInstance
		var bErr BuleErrorT
		var hint string

		if withInstance != defaultInstance {
			if si, bErr = s.get(withInstance); !bErr.isNil() {
				BuleExit(os.Stderr, bErr)
			} else {
				switch si.Type {
				case SAT:
					if p.IsSATProblem() {
						// OK, solving SAT problem with SAT instance
					} else {
						// Error, solving QBF problem with SAT instance
						BuleExit(os.Stderr, newBuleErrInadequateSolver(QBF, si.Prog))
					}
				case QBF:
					if p.IsSATProblem() {
						// Fair, solving SAT problem with QBF instance
						hint = "Use a dedicated SAT solver for this problem."
					} else {
						// OK, solving QBF problem with QBF instance
					}
				default:
					BuleExit(os.Stderr, errUnknownSolverType)
				}
			}
		} else {
			// try to infer a default solver instance
			if p.IsSATProblem() {
				if si, bErr = s.getSat(defaultInstance); !bErr.isNil() {
					// no default SAT solvers, try QBF
					if si, bErr = s.getQbf(defaultInstance); !bErr.isNil() {
						// no default QBF solvers, error
						BuleExit(os.Stderr, newBuleErrInadequateSolver(SAT, defaultInstance))
					} else {
						hint = "Use a dedicated SAT solver for this problem."
					}
				}
			} else {
				if si, bErr = s.getQbf(defaultInstance); !bErr.isNil() {
					// no default QBF solvers, error (can't solve with default SAT!)
					BuleExit(os.Stderr, newBuleErrInadequateSolver(QBF, defaultInstance))
				}
			}
		}
		// printsolver details and some hints
		if p.IsSATProblem() {
			fmt.Fprintf(os.Stderr, "This is a SAT problem\n")
		} else {
			fmt.Fprintf(os.Stderr, "This is a QBF problem\n")
		}
		fmt.Fprintf(os.Stderr, "Using a %v solver instance '%s %s'\n", si.Type, si.Prog, si.Flags)
		if len(hint) > 0 {
			fmt.Fprintf(os.Stderr, "*hint* %s\n", hint)
		}

		// prepare fields
		flagsSplit := strings.Fields(si.Flags)
		progName := strings.ToUpper(filepath.Base(si.Prog))

		fmt.Fprintf(os.Stderr, "Solving. . .\n\n")

		isTrue := true
		{
			cmdOutput, err = exec.Command(si.Prog, append(flagsSplit, outputGroundFile)...).Output()
			if exitError, ok := err.(*exec.ExitError); ok {
				debug(1, fmt.Sprintf("%s exist status:", si.Prog), exitError.ExitCode())
				switch UnifySolverOutput(progName, exitError.ExitCode()) {
				case SOLVER_TRUE:
					isTrue = true
				case SOLVER_FALSE:
					isTrue = false
				case SOLVER_ERROR:
					fallthrough
				default:
					log.Println(fmt.Sprintf("Exit error of %s is", progName), exitError.ExitCode())
					log.Println(fmt.Sprintf("Error %s log:\n ", progName), string(cmdOutput))
					log.Println("Omitting parsing because of error in solving: ", err)
					return
				}
			} else if err != nil {
				log.Println(fmt.Sprintf("Error %s log:\n ", progName), string(cmdOutput))
				log.Println("Omitting parsing because of error in solving: ", err)
				return
			}
		}

		// Parse output and return result
		{
			debug(1, fmt.Sprintf("Output by %s", progName))
			scanner := bufio.NewScanner(strings.NewReader(string(cmdOutput)))
			result := []int{}
			fmt.Println("-------- Solver output ----- ")
			for scanner.Scan() {
				s := scanner.Text()
				if strings.HasPrefix(s, "V ") || strings.HasPrefix(s, "v ") {
					fields := strings.Fields(s)
					for _,field := range fields[1:] {
						v, err := strconv.Atoi(field)
						if err != nil {
							log.Println("Error in parsing result: ", err)
							os.Exit(1)
						}
						if v != 0 {
							result = append(result, v)
						}
					}
					continue
				} else {
					log.Println("solver>> ", s)
				}
			}
			reverseMap := map[int]string{}
			for k, v := range clauseProgram.idMap {
				reverseMap[v] = k
			}

			fmt.Println("-------- Solver output end ----- ")

			if isTrue {
				fmt.Println("s SAT\n--------")
				for _, id := range result {
					if id > 0 {
						fmt.Printf("%s.\n", reverseMap[id])
					} else {
						fmt.Printf("~%s.\n", reverseMap[-1*id])
					}
				}
			} else {
				fmt.Println("s UNSAT")
			}
			fmt.Println("--------")
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

func dimacsTidyUp(dimacsOut string, isSat bool) string {

	if !isSat {
		// is QBF, keep full notation
		return dimacsOut
	}
	lines := strings.Split(dimacsOut, "\n")
	xs := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(line) > 0 {
			switch string(line[0]) {
			case "e":
			case "a":
			default:
				xs = append(xs, line)
			}
		}
	}
	return strings.Join(xs, "\n")
}
