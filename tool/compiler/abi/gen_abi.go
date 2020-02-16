package abi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/scanner"
)

var(
	//expression to match invoke_action and invoke_func labeled functions
	invokeRegexp = `(INVOKE_ACTION|INVOKE_FUNC) \( "\b[\w]+`
	baseRegexp   = `class [\w]+( : public akchain : : Contract)`
	TempCppFile = "temp.cpp"
	includeRegexp = `# include \"[^\"]+\"`
	headerRegexp = `\"[^\"]+\"`
	GenerateAbi bool
	ABIPrefix = NewBinPrefix(ContractTypeRuntime, ContractVMTypeWASM, ContractPatternTypeUnknown)
)


func (b *BinPrefix)SetOption(op [ExtensionLen]byte) *BinPrefix {
	b.Extension = op
	return b
}

func (b BinPrefix)Byte() (out [CodePrefixLen]byte) {
	out[0] = byte(b.TypeBin)
	out[1] = byte(b.VMTypeBin)
	out[2] = byte(b.PattenTypeBin)
	for i, b := range b.Extension {
		out[i + 3] = b
	}
	return out
}

type InvokeType struct {
	name string
	invokeType []string
}

func (cc *ContractClass)Execute(args []string) error  {
	file := args[0]
	err := parseClassFromFile(file, cc)
	if err != nil {
		return err
	}

	// collect functions in action entry
	functions := collectFunctions(file)
	if len(functions) != 0 {
		ABIPrefix.PattenTypeBin = ContractPatternType2
		ContractMainFile = file
	}else {
		ABIPrefix.PattenTypeBin = ContractPatternType1
		TempCppFile = filepath.Join(PrefixPath, TempCppFile)
		ContractMainFile = TempCppFile
		err = cc.GenCode(file)
		if err != nil {
			return err
		}
	}

	if GenerateAbi {
		return cc.genAbi(file, functions)
	}
	return nil
}

func collectFunctions(file string) []InvokeType {
	contractContent := readContract(file)
	functions := getActionAndFunc(contractContent)
	return functions
}

func (cc *ContractClass)genAbi(file string, functions []InvokeType) error {
	m := getActionEntry(functions, cc)
	jsonByte, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		return err
	}
	abiFile := getAbiFileName(file)
	err = writeABI(jsonByte, abiFile)
	if err != nil {
		return err
	}
	return nil
}

func getAbiFileName(srcFile string) string {
	// replace cpp or cc with json
	srcFile = filepath.Base(srcFile)
	abiFile := strings.TrimRight(srcFile, "cpp")
	abiFile = strings.TrimRight(abiFile, "cc")
	return  fmt.Sprintf("%sabi",abiFile)
}

func writeABI(abi []byte, fileName string) error {
	err := ioutil.WriteFile(fileName, abi, 0600)
	return err
}

func parseClassFromFile(file string, cc *ContractClass) error {
	cl := searchClass(file)
	PrefixPath = filepath.Dir(file)
	if len(cl) != 0 {
		ClassDefineFile = file
	}else {
		PrefixPath = filepath.Dir(file)
		includes := searchIncludes(file)
		for _, include := range includes {
			defineFile := filepath.Join(PrefixPath, include)
			cl = searchClass(defineFile)
			if len(cl) != 0 {
				ClassDefineFile = defineFile
				break
			}
		}
	}
	if len(cl) == 0 {
		return errors.New("no class found! ")
	}
	//fmt.Println(cl)
	err := cc.ParseClass(cl)
	return err
}

func searchIncludes(file string) (includes []string) {
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()
	var sc scanner.Scanner
	sc.Init(f)
	var contractTokens []string
	for tok := sc.Scan(); tok != scanner.EOF; tok = sc.Scan() {
		contractTokens = append(contractTokens, sc.TokenText())
	}
	reg, _ := regexp.Compile(includeRegexp)
	in := reg.FindAllString(strings.Join(contractTokens, " "), -1)
	reg, _  = regexp.Compile(headerRegexp)
	for _, v := range in {
		head := reg.FindString(v)
		head = strings.TrimLeft(head, "\"")
		head = strings.TrimRight(head, "\"")
		head = strings.TrimLeft(head, " ")
		head = strings.TrimRight(head, " ")
		includes = append(includes, head)
	}
	return
}

//read contract
func readContract(file string) (contract []string) {
	var sc scanner.Scanner
	fileBuffer, err := os.Open(file)
	if err != nil {
		return
	}
	defer fileBuffer.Close()
	sc.Init(fileBuffer)
	for tok := sc.Scan(); tok != scanner.EOF; tok = sc.Scan() {
		contract = append(contract, sc.TokenText())
	}
	return
}

// collect functions called in action entry, prepare for abi output
// require actionEntry defined by user
func getActionEntry(funcs []InvokeType, cc *ContractClass) []*Method {
	var m []*Method
	// if no action entry is defined in contract, output abi
	if len(funcs) == 0 {
		for _, v := range cc.FuncCache {
			m = append(m, cc.FuncSigs[v])
		}
		return m
	}

	for _, v := range funcs {
		funcSig := cc.FuncSigs[v.name]
		if funcSig != nil {
			if len(v.invokeType) != 0 {
				funcSig.Type = append(funcSig.Type, v.invokeType...)
			}
			m = append(m, funcSig)
		}
	}
	return m
}

//input shall be the content of a contract
func getActionAndFunc(contract []string) []InvokeType {
	actionAndEvent := make([]InvokeType, 0)
	reg, _ := regexp.Compile(invokeRegexp)
	//fmt.Println(contract)
	invokes := reg.FindAllString(strings.Join(contract, " "), -1)
	for _, invoke := range invokes {
		var invokeFunc InvokeType
		invoke = strings.TrimRight(invoke, " ")
		invokeSlice := strings.Split(invoke, " ")
		length := len(invokeSlice)
		 invokeFunc.name = strings.TrimLeft(invokeSlice[length-1], "\"")
		switch invokeSlice[0] {
		case "INVOKE_ACTION":
			invokeFunc.invokeType = []string{"action","event"}
		case "INVOKE_FUNC":
			invokeFunc.invokeType = []string{"action","event"}
			
		}
		actionAndEvent = append(actionAndEvent, invokeFunc)
	}
	return actionAndEvent
}

func searchClass(file string) (class []string) {
	var sc scanner.Scanner
	fileBuffer, err := os.Open(file)
	defer fileBuffer.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	sc.Init(fileBuffer)
	for tok := sc.Scan(); tok != scanner.EOF; tok = sc.Scan(){
		switch sc.TokenText() {
		case "class":
			classDeclare := readClass(sc)
			reg, _ := regexp.Compile(baseRegexp)
			if reg.MatchString(strings.Join(classDeclare, " ")) {
				return classDeclare
			}
		}
	}
	return
}

// read class definition and public functions
func readClass(sc scanner.Scanner) []string {
	var scope Scope
	args := make([]string,0)
	for ; !isOutScope(scope); sc.Scan() {
		switch sc.TokenText() {
		case "{":
			scope.entered = true
			scope.subScope++
			args = append(args, sc.TokenText())
		case "}":
			scope.subScope--
			args = append(args, sc.TokenText())
		case "public":
			if scope.entered {
				scope.publicScop = true
			}
		case "private", "protected":
			if scope.entered {
				scope.publicScop = false
			}
		}
		//only collect public functions
		if !scope.entered{
			args = append(args, sc.TokenText())
		}
		if isPublicScope(scope){
			args = append(args, sc.TokenText())
		}
	}
	return args
}

type Scope struct {
	entered bool //mark if we entered a class scope
	subScope int //sub scope counter
	publicScop bool
}

func isOutScope(s Scope) bool {
	if s.entered && s.subScope == 0 {
		return true
	}
	return false
}

func isPublicScope(s Scope) bool  {
	return s.entered && s.publicScop
}
