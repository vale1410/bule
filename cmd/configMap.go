package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

//-----------------------------------------------------------------------------
// Type definitions
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

//-----------------------------------------------------------------------------
// Type logic
//-----------------------------------------------------------------------------
func newSolverInstance(prog string, flags string, t string) (*SolverInstance, BuleErrorT) {
	if _, err := exec.LookPath(prog); err != nil {
		return nil, newBuleErrNotFound(prog)
	}

	if len(flags) == 0 {
		return nil, errMalformedFlags
	}
	// flag string needs to be empty (only @) or start with special character: @
	if string(flags[0]) == "@" {
		flags = strings.TrimSpace(flags[1:])
	} else {
		return nil, errMalformedFlags
	}

	stype := SolverT(strings.ToUpper(t))
	if _, valid := map[SolverT]bool{
		SAT: true,
		QBF: true,
	}[stype]; !valid {
		return nil, errBadArgFormat
	}
	return &SolverInstance{prog, flags, stype}, buleOK
}

func (s *Solvers) sync() BuleErrorT {
	viper.Set(MAIN_KEY, s)
	if err := viper.WriteConfig(); err != nil {
		return errConfigSync
	}
	return buleOK
}

func (s *Solvers) load() BuleErrorT {
	var solversUntyped interface{}
	if solversUntyped = viper.Get(MAIN_KEY); solversUntyped == nil {
		return errMalformedConfig
	}
	if err := mapstructure.Decode(solversUntyped, s); err != nil {
		return newBuleErrDecode(s)
	}
	// OK
	return buleOK
}

func (s *Solvers) set(label string, si *SolverInstance) BuleErrorT {
	switch si.Type {
	case SAT:
		s.Sat[label] = *si
	case QBF:
		s.Qbf[label] = *si
	default:
		BuleExit(os.Stderr, errUnknownSolverType)
	}
	return s.sync()
}

func (s *Solvers) setDefault(si *SolverInstance) BuleErrorT {
	return s.set("default", si)
}

func (s *Solvers) get(label string) (*SolverInstance, BuleErrorT) {
	switch si, err := s.getSat(label); true {
	case err.equals(errNoSuchInstance):
		return s.getQbf(label)
	case err.equals(buleOK):
		return si, buleOK
	default:
		return nil, err
	}
}

func (s *Solvers) getSat(label string) (*SolverInstance, BuleErrorT) {
	if si, exists := s.Sat[label]; exists {
		return &si, buleOK
	}
	return nil, errNoSuchInstance
}

func (s *Solvers) getQbf(label string) (*SolverInstance, BuleErrorT) {
	if si, exists := s.Qbf[label]; exists {
		return &si, buleOK
	}
	return nil, errNoSuchInstance
}

//-----------------------------------------------------------------------------
// Extend type logic for removing instances
//-----------------------------------------------------------------------------
func (s *Solvers) remove(label string) BuleErrorT {
	switch err := s.removeSat(label); true {
	case err.equals(errNoSuchInstance):
		return s.removeQbf(label)
	case err.equals(buleOK):
		fallthrough
	default:
		return err
	}
}

func (s *Solvers) removeSat(label string) BuleErrorT {
	if v, exists := s.Sat[label]; exists {
		delete(s.Sat, label)
		if err := s.sync(); !err.isNil() {
			fmt.Fprintf(os.Stderr, "Sync failed, reverting changes")
			s.Sat[label] = v
			return err
		}
		return buleOK
	} else {
		return errNoSuchInstance
	}
}

func (s *Solvers) removeQbf(label string) BuleErrorT {
	if v, exists := s.Qbf[label]; exists {
		delete(s.Qbf, label)
		if err := s.sync(); !err.isNil() {
			fmt.Fprintf(os.Stderr, "Sync failed, reverting changes")
			s.Qbf[label] = v
			return err
		}
		return buleOK
	} else {
		return errNoSuchInstance
	}
}

func (s *Solvers) purge() BuleErrorT {
	s.Sat = make(map[string]SolverInstance)
	s.Qbf = make(map[string]SolverInstance)
	return s.sync()
}
