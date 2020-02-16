package common

import (
	"github.com/tendermint/go-amino"
)

const (
	CodePrefixLen = 10
)

type ContractType int
const (
	_ ContractType = iota
	ContractTypeNative  = 0x01
	ContractTypeRuntime = 0x02
	ContractTypeUnknown = 0x03
)

type ContractVMType int
const (
	_ ContractVMType = iota
	ContractVMTypeWASM    = 0x01
	ContractVMTypeUnknown = 0x02
)

type ContractPatternType int
const (
	_  ContractPatternType = iota
	ContractPatternType1       = 0x01
	ContractPatternType2       = 0x02
	ContractPatternTypeUnknown = 0x03
)

type ContractState int
const(
	_ ContractState = iota
	ContractNormal  = 0x01
	ContractSuspend = 0x02
)

type Param struct {
	Index     int          `json:"index"`
	Name      string       `json:"name"`
	ParamType string       `json:"paramType"`
	Value     interface{} `json:"value"`
}

type ParamSort []*Param

func (p ParamSort) Len() int {return len(p)}
func (p ParamSort) Swap(i, j int) { p[i], p[j] = p[j], p[i]}
func (p ParamSort) Less(i, j int) bool {
	return p[i].Index < p[j].Index
}

type ContractInterface interface {
	OwnerAddr() string
	ContractAddr() string
}

type ContractResult struct {
	IsSuccess bool          `json:"issuccess"`
	ResultType string       `json:"resultType"`
	Value      interface{} `json:"value"`
}

type ContractInfo struct {
	Addr         string             `json:"addr"`
	Name         string             `json:"name"`
	Owner        string             `json:"owneraddr"`
	Codes        []byte             `json:"codes"`
	CodesDesc    string             `json:"codesdesc"`
	State        ContractState      `json:"state"`
	RelatedInfos map[string]string  `json:"relatedinfos"`
}

func GenerateContractCodePrefix(cType ContractType, cVMType ContractVMType, cPatternType ContractPatternType) []byte {
	return []byte{byte(cType), byte(cVMType), byte(cPatternType), 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
}

func (c *ContractInfo) OwnerAddr() string {
	return c.Owner
}

func (c *ContractInfo) ContractAddr() string {
	return c.Addr
}

func EncodeContractInfo(cdc *amino.Codec, cInfo *ContractInfo) []byte {
	jsonBytes, _ := cdc.MarshalJSON(cInfo)

	return jsonBytes
}

func DecodeContractInfo(cdc *amino.Codec, bytes []byte) (cInfo ContractInfo) {
	cdc.UnmarshalJSON(bytes, &cInfo)
	return
}


