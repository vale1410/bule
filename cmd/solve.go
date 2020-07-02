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
	"github.com/spf13/cobra"
	bule "github.com/vale1410/bule/lib"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// solveCmd represents the solve command
var solveCmd = &cobra.Command{
	Use:   "solve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			return
		}

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

		log.Println("Ground program writen to ", outputGroundFile)

		var cmdOutput []byte
		// Run File with DEPQBF
		isTrue := true
		{
				log.Println("depqbf", "--qdo", "--no-dynamic-nenofex", outputGroundFile)
				cmdOutput, err = exec.Command("depqbf", "--qdo", "--no-dynamic-nenofex", outputGroundFile).Output()
				if exitError, ok := err.(*exec.ExitError); ok {
					log.Println("DEPQBF exist status:", exitError.ExitCode())
					if  exitError.ExitCode() == 10 {
						isTrue = true
					} else if  exitError.ExitCode() == 20 {
						isTrue = false
					} else {
						log.Println("Error DEPQBF log:\n ", string(cmdOutput))
						log.Println("Omitting parsing because of error in solving: ", err)
						return
					}
				} else if err != nil {
					log.Println("Error DEPQBF log:\n ", string(cmdOutput))
					log.Println("Omitting parsing because of error in solving: ", err)
					return
				}
				log.Println("success: depqbf")
		}

		// Parse output and return result
		{
			fmt.Println("Output by DEPQBF")
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
			for k,v:= range clauseProgram.idMap {
				reverseMap[v] = k
			}

			if isTrue {
				fmt.Println("TRUE")
				for _, id := range result {
					if id > 0 {
						fmt.Print(reverseMap[id]," ")
					} else {
						fmt.Print("~",reverseMap[-1*id]," ")
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// solveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only prepare when this command
	// is called directly, e.g.:
	// solveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
