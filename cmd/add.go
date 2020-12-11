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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//-----------------------------------------------------------------------------
// Error definitions and error constructors
//-----------------------------------------------------------------------------

var NewErrDecode = func(receiver interface{}) error {
	return errors.New(fmt.Sprintf("Couldn't map %s file contents to program structure %s.",
		filepath.Ext(viper.ConfigFileUsed())[1:], reflect.TypeOf(receiver)))
}

var NewErrBinaryNotFound = func(prog string) error {
	return errors.New(fmt.Sprintf("Couldn't locate program '%s' on this system, please check your $PATH settings.", prog))
}

var ErrMalformedConfig = errors.New("Malformed configuration file.")
var ErrBadArgFormat = errors.New(fmt.Sprintf("Required format: prog | flag | type [%v, %v].", SAT, QBF))
var ErrUnknownSolverType = errors.New("Unknown solver type.")
var ErrMalformedFlags = errors.New("Malformed flags argument: pass flags with '@<flags str>' format or plain @ for no flags")
var ErrConfigSync = errors.New("Error during configuration sync.")
var ErrNoLabelSpecified = errors.New("Please specify at least one of [--label <label>, --setdefault]")

//-----------------------------------------------------------------------------
// Types & type logic
//-----------------------------------------------------------------------------

const MAIN_KEY = "solvers"

type SolverT string

const (
	SAT SolverT = "SAT"
	QBF         = "QBF"
)

type SolverInstance struct {
	Prog  string  `mapstructure:"prog"`
	Flags string  `mapstrucutre:"flags"`
	Type  SolverT `mapstructure:"type"`
}

type Solvers struct {
	Sat map[string]SolverInstance `mapstructure:"sat"`
	Qbf map[string]SolverInstance `mapstructure:"qbf"`
}

func newSolverInstance(prog string, flags string, t string) (*SolverInstance, error) {
	if _, err := exec.LookPath(prog); err != nil {
		return nil, NewErrBinaryNotFound(prog)
	}

	if len(flags) == 0 {
		return nil, ErrMalformedFlags
	}
	// flag string needs to be empty (only @) or start with special character: @
	if string(flags[0]) == "@" {
		flags = strings.TrimSpace(flags[1:])
	} else {
		return nil, ErrMalformedFlags
	}

	stype := SolverT(strings.ToUpper(t))
	if _, valid := map[SolverT]bool{
		SAT: true,
		QBF: true,
	}[stype]; !valid {
		return nil, ErrBadArgFormat
	}
	return &SolverInstance{prog, flags, stype}, nil
}

func (s *Solvers) sync() error {
	viper.Set(MAIN_KEY, s)
	if err := viper.WriteConfig(); err != nil {
		return err
	}
	return nil
}

func (s *Solvers) load() error {
	var solversUntyped interface{}
	if solversUntyped = viper.Get(MAIN_KEY); solversUntyped == nil {
		return ErrMalformedConfig
	}
	if err := mapstructure.Decode(solversUntyped, s); err != nil {
		return NewErrDecode(s)
	}
	// OK
	return nil
}

func (s *Solvers) set(label string, si *SolverInstance) error {
	switch si.Type {
	case SAT:
		s.Sat[label] = *si
	case QBF:
		s.Qbf[label] = *si
	default:
		ErrExit(ErrUnknownSolverType, 1)
	}
	return s.sync()
}

func (s *Solvers) setDefault(si *SolverInstance) error {
	return s.set("default", si)
}

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
			ErrExit(ErrBadArgFormat, 1)
		}
		if n == 3 {
			if _, err := newSolverInstance(args[0], args[1], args[2]); err != nil {
				ErrExit(err, 1)
			}
			if label == "" && !setDefault {
				ErrExit(ErrNoLabelSpecified, 1)
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
					ErrExit(err, 1)
				} else {
					defer fd.Close()
					viper.SetConfigName(file)
					s.sync()
				}
			}
			return
		}

		if err := s.load(); err != nil {
			ErrExit(err, 1)
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
	addCmd.Flags().StringVarP(&label, "label", "", "", "label for solver instance")
	addCmd.Flags().BoolVarP(&setDefault, "setdefault", "", false, "set instance as default solver for SAT/QBF")
	addCmd.Flags().BoolVarP(&newConfig, "newconfig", "", false, "create empty Bule configuration in current directory")
}
