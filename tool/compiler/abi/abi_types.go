package abi

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)
var (
	//funcRegexp = `(\[ \[ (ACTION|EVENT) \] \] ){0,2}(char|void) [\w]+ \(`
	//used inside class definition
	funcRegexp = `(\[ \[ (ACTION|EVENT|OWNERABLE|PAYABLE) \] \] ){0,4}(char|void|int|float|bool) (\*\s)*[\w]+ \( (([\w]+|,|\*)\s)*\)` //add more
	pureFunc = `(char|void|int|float|bool) (\*\s)*[\w]+ \( (([\w]+|,|\*)\s)*\)`
	inputRegexp = `\( (([\w]+|,|\*)\s)*\)`
	exportRegexp = `extern(\s)*"(C|c)"(\s)*{(\s)*([^\{]*\{[^\}]*\})*(\s)*}`
)
var (
	ClassDefineFile  string
	ContractMainFile string
	PrefixPath       string
)

const (
	CodePrefixLen = 10
	ExtensionLen = 7
)

type ContractType int
const (
	_ ContractType = iota
	ContractTypeNative  = 0x01
	ContractTypeRuntime = 0x02 //~
	ContractTypeUnknown = 0x03
)

type ContractVMType int
const (
	_ ContractVMType = iota
	ContractVMTypeWASM    = 0x01 //~
	ContractVMTypeUnknown = 0x02
)

type ContractPatternType int
const (
	_  ContractPatternType = iota
	ContractPatternType1       = 0x01 //action entry
	ContractPatternType2       = 0x02
	ContractPatternTypeUnknown = 0x03
)

type BinPrefix struct {
	TypeBin ContractType
	VMTypeBin ContractVMType
	PattenTypeBin ContractPatternType
	Extension [ExtensionLen]byte
}

func NewBinPrefix(contractType ContractType, contractVmType ContractVMType, contractPattenType ContractPatternType) *BinPrefix {
	return &BinPrefix{
		TypeBin: contractType,
		VMTypeBin: contractVmType,
		PattenTypeBin: contractPattenType,
	}
}

type Method struct {
	Name string `json:"name"`
	Inputs []*InputType `json:"inputs"`
	Outputs *OutputType `json:"outputs"`
	Type []string `json:"type"`  // function label, such as action and event
}

func NewMethod() *Method {
	return &Method{
		Inputs:make([]*InputType, 0),
		Type:make([]string, 0),
	}
}

//generate code
func (m Method) genCode(className string) []string {
	outputType := GetType(m.Outputs.Type)
	inputs := make([]string, 0)
	paramName := make([]string, 0)
	for _, v := range m.Inputs {
		ctype := GetType(v.Type)
		Arg := fmt.Sprintf("%s %s",ctype, v.Name)
		inputs = append(inputs, Arg)
		paramName = append(paramName, v.Name)
	}
	inputsStr := strings.Join(inputs, ", ")
	arg := strings.Join(paramName, ",")
	line0 := fmt.Sprintf("\n")
	line1 := fmt.Sprintf("EXPORT %s %s(%s){\n", outputType, m.Name, inputsStr)
	line2 := fmt.Sprintf("\t%s %s;\n", className, "co")
	line3 := fmt.Sprintf("\treturn co.%s(%s);\n",m.Name,arg)
	line4 := fmt.Sprintf("}\n")
	return []string{line0,line1, line2, line3, line4}
}

func (m *Method)addType(t string) *Method {
	m.Type = append(m.Type, t)
	return m
}

func (m *Method)addInputs(n *InputType) *Method {
	m.Inputs = append(m.Inputs, n)
	return m
}

func (m *Method)addOutPuts(n *OutputType) *Method {
	m.Outputs = n
	return m
}

type InputType struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type OutputType struct {
	Type string `json:"type"`
}

type ContractClass struct {
	ClassName string
	FuncSigs  map[string]*Method
	FuncCache []string
}

func (cc *ContractClass)addFunc(m *Method)  {
	cc.FuncSigs[m.Name] = m
	cc.FuncCache = append(cc.FuncCache, m.Name)
}

func NewContractClass() *ContractClass {
	return &ContractClass{
		FuncSigs:make(map[string]*Method),
	}
}

//copy contract source code to temp file
func (c ContractClass)GenCode(file string) error {
	contractContent, _ := ioutil.ReadFile(file)
	reg, _ := regexp.Compile(exportRegexp)

	matched := reg.FindAll(contractContent,-1)
	if len(matched) >1 {
		return errors.New("Too much code exported. ")
	}
	// write title
	title := fmt.Sprintf("extern \"C\"{\n")
	//var result strings.Builder
	//builder := strings.Builder{}
	var result = make([]byte,0)
	result = append(result, title...)
	for _,v := range c.FuncCache {
		m := c.FuncSigs[v]
		lines := m.genCode(c.ClassName)
		for _, line := range lines{
			result = append(result, fmt.Sprintf("\t%s",line)...)
		}
	}
	// write file ends
	//_, err = f.WriteString("}\n")
	codeEnd := fmt.Sprintf("}\n")
	result = append(result, codeEnd...)
	var newResult []byte
	if len(matched) == 0 {
		result = append([]byte("\n"),result...)
		newResult = append(contractContent, result...)
	}else {
		newResult = reg.ReplaceAll(contractContent, result)
	}

	f, err := os.OpenFile(TempCppFile, os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		return err
	}
	f.Write(newResult)
	return nil
}

type ActionEntry []*Method
func NewActionEntry() ActionEntry  {
	ac := make([]*Method, 0)
	return ac
}

func (a ActionEntry)Add(m *Method) ActionEntry {
	b := append(a, m)
	return b
}

// parse class into contractClass struct
// used for functions defined in derived class, both with
// and without event|action labeled
func (cc *ContractClass)ParseClass(args []string) error {
	cc.ClassName = args[1]
	//funcS := getFunctions(args)
	reg, err := regexp.Compile(funcRegexp)
	if err != nil {
		return err
	}
	funcs := reg.FindAllString(strings.Join(args, " "), -1)
	for _, v := range funcs {
		m := NewMethod()
		m.ParseFunction(v)
		cc.addFunc(m)
		//cc.FuncSigs[m.Name] = m
	}
	//fmt.Println(cc)
	return nil
}

// parse single function definition into `Method` struct
func (m *Method)ParseFunction(foo string)  {
	if strings.Contains(foo, "ACTION") {
		m.addType("action")
	}

	if strings.Contains(foo, "EVENT"){
		m.addType("event")
	}
	if strings.Contains(foo, "OWNERABLE"){
		m.addType("ownerable")
	}
	if strings.Contains(foo, "PAYABLE"){
		m.addType("payable")
	}

	reg, _ := regexp.Compile(pureFunc)
	filterFunc := reg.FindString(foo)
	m.Name = getFuncName(filterFunc)
	m.Inputs = getInputs(filterFunc)
	m.Outputs = getOutput(filterFunc)
	//fmt.Println(filterFunc)
}

func getInputs(foo string) []*InputType {
	res := make([]*InputType, 0)
	reg, _ := regexp.Compile(inputRegexp)
	inputs := reg.FindString(foo)
	inputs = strings.TrimRight(inputs, ")")
	inputs = strings.TrimLeft(inputs, "(")
	inputs = strings.TrimRight(inputs," ")
	// input pair
	inputsSlice := strings.Split(inputs, ",")

	// parse single input name and type
	for _, input := range inputsSlice {
		input = strings.TrimRight(input, " ")
		input = strings.TrimLeft(input, " ")
		if len(input) == 0 {
			continue
		}
		in := new(InputType)
		inputS := strings.Split(input, " ")
		length := len(inputS)
		in.Name = inputS[length-1]
		for _, v := range inputS {
			switch v {
			// find type, in case static or const  exist
 			case "int","void", "bool":
				in.Type = v
			case "char":
				in.Type = "string"
			}
		}
		res = append(res, in)
	}
	return res
}

func getFuncName(foo string) string {
	fooSlice := strings.Split(foo, " ")
	var index = 0
	for ; fooSlice[index] != "("; index ++ {

	}
	return fooSlice[index - 1]
}

func getOutput(foo string) *OutputType {
	out := new(OutputType)
	fooSlice := strings.Split(foo, " ")
	for _, v := range fooSlice{
		typeName := GetTypeName(v)
		out.Type = typeName
		return out
	}
	return out
}

// used to transform c type to contract abi type
// char * should be output as string
func GetTypeName(cType string) string {
	switch cType{
	case "char":
		return "string"
	default:
		return cType
	}
}

// string equal `char *` type
func GetType(typeName string) string {
	switch typeName {
	case "string":
		return "char*"
	default:
		return typeName
	}
}

