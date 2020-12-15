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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//-----------------------------------------------------------------------------
// Command  logic
//-----------------------------------------------------------------------------

// Flags
var (
	label      string
	setDefault bool
	newConfig  bool
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a SAT or QBF solver to Bule.",
	Long:  "Add a SAT or QBF solver to Bule.",

	Args: func(cmd *cobra.Command, args []string) error {
		var n int
		if n = len(args); n != 0 && n != 3 {
			BuleExit(os.Stderr, errBadArgFormat)
		}
		if n == 3 {
			if _, err := newSolverInstance(args[0], args[1], args[2]); !err.isNil() {
				BuleExit(os.Stderr, err)
			}
			if label == "" && !setDefault {
				BuleExit(os.Stderr, errNoLabelSpecified)
			}
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {

		var s Solvers

		if len(args) == 0 {
			if newConfig {
				file := ".bule.yaml"
				if _, err := os.Stat(file); !os.IsNotExist(err) {
					fmt.Printf("File ./%s already exists. Won't change.\n", file)
					return
				}
				var fd, err = os.Create(file)
				if err != nil {
					BuleExit(os.Stderr, newBuleFileCreate(err.Error()))
				} else {
					defer fd.Close()
					viper.SetConfigName(file)
					s.sync()
				}
			}
			return
		}

		if err := s.load(); !err.isNil() {
			BuleExit(os.Stderr, err)
		}
		// always valid, already checked args
		si, _ := newSolverInstance(args[0], args[1], args[2])

		if setDefault {
			s.setDefault(si)
		}
		if label != "" {
			s.set(label, si)
		}
		fmt.Println("OK.")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&label, "label", "l", "", "label for solver instance")
	addCmd.Flags().BoolVarP(&setDefault, "setdefault", "", false, "set instance as default solver for SAT/QBF")
	addCmd.Flags().BoolVarP(&newConfig, "newconfig", "", false, "create empty Bule configuration in current directory")
}
