package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"

	"github.com/spf13/viper"
)

//-----------------------------------------------------------------------------
// Error definitions and error constructors
//-----------------------------------------------------------------------------

// error class
const (
	BULE_ERR_CLASS_SYSTEM  int = 100
	BULE_ERR_CLASS_CONFIG  int = 200
	BULE_ERR_CLASS_ARGS    int = 300
	BULE_ERR_CLASS_MISSING int = 400
	BULE_ERR_CLASS_RUNTIME int = 500
	BULE_ERR_CLASS_PARSER  int = 600
	BULE_ERR_CLASS_SOLVER  int = 700
)

type BuleErrorT struct {
	err  error
	code int
}

func (be *BuleErrorT) equals(other BuleErrorT) bool {
	// compare by code only
	return be.code == other.code
}

func (be *BuleErrorT) is(other BuleErrorT) bool {
	// strict compare: by message and code
	return be.equals(other) && (be.err.Error() == other.err.Error())
}

func (be *BuleErrorT) isNil() bool {
	return be.equals(buleOK)
}

// No error
var buleOK = BuleErrorT{errors.New(""), 0}

// System
var errPathResolve = BuleErrorT{errors.New("Couldn't resolve $HOME dir"), BULE_ERR_CLASS_SYSTEM + 1}
var newBuleFileCreate = func(msg string) BuleErrorT {
	return BuleErrorT{
		errors.New(msg),
		BULE_ERR_CLASS_SYSTEM + 2}
}
var newBuleErrNotFound = func(prog string) BuleErrorT {
	return BuleErrorT{
		errors.New(fmt.Sprintf("Couldn't locate program '%s' on this system, please check your $PATH settings", prog)),
		BULE_ERR_CLASS_SYSTEM + 3}
}

// Config file
var errMalformedConfig = BuleErrorT{errors.New("Malformed configuration file"), BULE_ERR_CLASS_CONFIG + 1}
var errConfigSync = BuleErrorT{errors.New("Error during configuration sync"), BULE_ERR_CLASS_CONFIG + 2}

// User arguments
var errBadArgFormat = BuleErrorT{errors.New(fmt.Sprintf("Required format: 'prog\t flag\t type[%v, %v]'", SAT, QBF)), BULE_ERR_CLASS_ARGS + 1}
var errMalformedFlags = BuleErrorT{errors.New("Malformed flags argument: pass flags with @'flags string' format or plain @ for no flags"), BULE_ERR_CLASS_ARGS + 2}
var errNoLabelSpecified = BuleErrorT{errors.New("Please specify at least one of [--label 'label', --setdefault]"), BULE_ERR_CLASS_ARGS + 3}

// Missing key
var errUnknownSolverType = BuleErrorT{errors.New("Unknown solver type"), BULE_ERR_CLASS_MISSING + 1}
var errNoSuchInstance = BuleErrorT{errors.New("No such instance in configuration file"), BULE_ERR_CLASS_MISSING + 2}

// Runtime
var errRuntimeCommand = BuleErrorT{errors.New("While executing cobra command"), BULE_ERR_CLASS_RUNTIME + 1}
var newBuleErrDecode = func(receiver interface{}) BuleErrorT {
	return BuleErrorT{
		errors.New(fmt.Sprintf("Couldn't map %s file contents to program structure %",
			filepath.Ext(viper.ConfigFileUsed())[1:], reflect.TypeOf(receiver))),
		BULE_ERR_CLASS_RUNTIME + 2}
}

// Parser
var errParse = BuleErrorT{errors.New("Error parsing the program"), BULE_ERR_CLASS_PARSER + 1}

// Solver
var newBuleErrInadequateSolver = func(t SolverT, prog string) BuleErrorT {
	return BuleErrorT{
		errors.New(fmt.Sprintf("Error during solve step: inadequate %v solver '%s'", t, prog)),
		BULE_ERR_CLASS_SOLVER + 1}
}

// Log routine
func BuleLogErr(wr io.Writer, be BuleErrorT) {
	wr.Write([]byte(fmt.Sprintf("%s (%d)\n", be.err.Error(), be.code)))
}

// Exit routine
func BuleExit(wr io.Writer, be BuleErrorT) {
	var bytes []byte
	if !be.isNil() {
		bytes = []byte(fmt.Sprintf("*****\t%s.\n\tExiting with code %d.\n", be.err.Error(), be.code))
	}
	wr.Write(bytes)
	os.Exit(be.code)
}

//-----------------------------------------------------------------------------
// Error mappings for different solvers
//-----------------------------------------------------------------------------

type SolverOutputT int

const (
	SOLVER_TRUE SolverOutputT = iota
	SOLVER_FALSE
	SOLVER_ERROR
)

func UnifySolverOutput(progName string, code int) SolverOutputT {
	// use uppper-case for solver names
	switch progName {

	// case "CAQE"

	// case ....

	case "DEPQBF":
		fallthrough
	default:
		switch code {
		case 10:
			return SOLVER_TRUE
		case 20:
			return SOLVER_FALSE
		default:
			return SOLVER_ERROR
		}
	}
}
