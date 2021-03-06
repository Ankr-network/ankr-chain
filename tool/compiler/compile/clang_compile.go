package compile

import (
	"errors"
	"fmt"
	"github.com/Ankr-network/ankr-chain/tool/compiler/abi"
	"os/exec"
	"reflect"
)

var (
	cPlusType = "c++"
	cType     = "c"
)

type ClangOptions struct {
	Compiler   string
	Compile    string
	Optimise   string
	Target     string
	Standard   string
	Extensions []string
}

var DefaultClangOptions = ClangOptions{
	Compiler: "clang++",
	Optimise: "-O3",
	Compile:  "-c",
	Target:   "--target=wasm32",
	Standard: "-std=c++14",
}

func (co *ClangOptions) Options() (args []string) {
	args = append(args, co.Optimise, co.Compile, co.Target)
	args = append(args, co.Extensions...)
	return
}

// set clang compile options
//func (*ClangOptions)setOption *ClangOptions{...}
func (cp *ClangOptions) withC() *ClangOptions {
	cp.Compiler = "clang"
	return cp
}

func (cp *ClangOptions) withCpp() *ClangOptions {
	cp.Compiler = "clang++"
	return cp
}

type srcContract struct {
	name     string
	fileType string // the contract whether is c or c++ type
}

func NewClangOption() *ClangOptions {
	return &DefaultClangOptions
}

// compile c/c++ file into object
func (co *ClangOptions) Execute(args []string) error {

	clangArgs := co.Options()
	clangArgs = append(clangArgs, abi.ContractMainFile)

	out, err := exec.Command(co.Compiler, clangArgs...).Output()
	if err != nil {
		return handleExeError(co.Compiler,err)
	}
	if len(out) != 0 {
		fmt.Println(string(out))
	}
	return nil
}

func handleExeError(cmdName string, err error) error {
	errType := reflect.TypeOf(err)
	name := errType.String()
	switch name {
	case "*exec.Error":
		ee := err.(*exec.Error)
		return errors.New(cmdName + ee.Error() )
	case "*exec.ExitError":
		ee := err.(*exec.ExitError)
		return errors.New(string(ee.Stderr))
	default:
		return errors.New(err.Error())
	}
}
