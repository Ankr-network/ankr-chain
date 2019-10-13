package module

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Ankr-network/ankr-chain/context"
	"github.com/Ankr-network/ankr-chain/types"
	"github.com/go-interpreter/wagon/exec"
)

const (
	PrintSFunc               = "print_s"
	PrintIFunc               = "print_i"
	StrlenFunc               = "strlen"
	StrcmpFunc               = "strcmp"
	JsonObjectIndexFunc      = "JsonObjectIndex"
	JsonCreateObjectFunc     = "JsonCreateObject"
	JsonGetIntFunc           = "JsonGetInt"
	JsonGetStringFunc        = "JsonGetString"
	JsonPutIntFunc           = "JsonPutInt"
	JsonPutStringFunc        = "JsonPutString"
	JsonToStringFunc         = "JsonToString"
	ContractCallFunc         = "ContractCall"
	ContractDelegateCallFunc = "ContractDelegateCall"
	TrigEventFunc            = "TrigEvent"
)

func Print_s(proc *exec.Process, strIdx int32) {
	str, err := proc.ReadString(int64(strIdx))
	if err != nil {
		proc.VM().Logger().Error("Print_s", "err", err)
	}
	proc.VM().Logger().Info("Print_s", "str", str)
}

func Print_i(proc *exec.Process, v int32) {
	proc.VM().Logger().Info("Print_i", "v", v)
}

func Strlen(proc *exec.Process, strIdx int32) int {
	len, err := proc.VM().Strlen(uint(strIdx))
	if err != nil {
		return -1
	}

	return len
}

func Strcmp(proc *exec.Process, strIdx1 int32, strIdx2 int32) int32 {
	cmpR, _ := proc.VM().Strcmp(uint(strIdx1), uint(strIdx2))
	return int32(cmpR)
}

func JsonObjectIndex(proc *exec.Process, jsonStrIdx int32) int32 {
	jsonStr, err := proc.ReadString(int64(jsonStrIdx))
	if err != nil {
		proc.VM().Logger().Error("JsonObjectIndex read json string", "err", err)
		return -1
	}

	jsonStruc := make(map[string]json.RawMessage)
	if err = json.Unmarshal([]byte(jsonStr), &jsonStruc); err != nil {
		proc.VM().Logger().Error(" JsonObjectIndex", "jsonStr", jsonStr, "err", err)
		return -1
	}

	curLen := len(proc.VMContext().JsonObjectCache)

	proc.VMContext().JsonObjectCache = append(proc.VMContext().JsonObjectCache, jsonStruc)

	return int32(curLen)
}

func JsonCreateObject(proc *exec.Process) int32 {
	jsonStruc := make(map[string]json.RawMessage)
	curLen := len(proc.VMContext().JsonObjectCache)
	proc.VMContext().JsonObjectCache= append(proc.VMContext().JsonObjectCache, jsonStruc)
	return int32(curLen)
}

func JsonGetInt(proc *exec.Process, jsonObjectIndex int32, argIndex int32) int64 {
	jsonStruc := proc.VMContext().JsonObjectCache[jsonObjectIndex]
	if jsonStruc == nil {
		proc.VM().Logger().Error(" JsonGetInt jsonObjectIndex invalid", "jsonIndex", jsonObjectIndex)
		return -1
	}

	argName, err := proc.ReadString(int64(argIndex))
	if err != nil {
		proc.VM().Logger().Error("JsonGetInt read arg name", "err", err)
		return -1
	}

	argVBytes, ok := jsonStruc[argName]
	if !ok {
		proc.VM().Logger().Error("JsonGetInt can't find the responding arg", "argName", argName)
		return -1
	}

	argVBytes = bytes.Trim(argVBytes, "\"")
	argV, err := strconv.ParseInt(string(argVBytes), 0, 64)
	if err != nil {
		proc.VM().Logger().Error("JsonGetInt ParseInt error", "err", err)
		return -1
	}

	return int64(argV)
}

func JsonGetString(proc *exec.Process, jsonObjectIndex int32, argIndex int32) uint64 {
	jsonStruc := proc.VMContext().JsonObjectCache[jsonObjectIndex]
	if jsonStruc == nil {
		proc.VM().Logger().Error(" JsonGetInt jsonObjectIndex invalid", "jsonIndex", jsonObjectIndex)
		return 0
	}

	argName, err := proc.ReadString(int64(argIndex))
	if err != nil {
		proc.VM().Logger().Error("JsonGetInt read arg name", "err", err)
		return 0
	}

	argVBytes, ok := jsonStruc[argName]
	if !ok {
		proc.VM().Logger().Error("JsonGetInt can't find the responding arg", "argName", argName)
		return 0
	}

	lenV := len(argVBytes)
	argVBytes = argVBytes[1 : lenV-1]

	pointer, err := proc.VM().SetBytes(argVBytes)
	if err != nil {
		proc.VM().Logger().Error("JsonGetInt SetBytes", "err", err)
	}

	return pointer
}

func JsonPutInt(proc *exec.Process, jsonObjectIndex int32, keyIndex int32,  valIndex int32) int32 {
	jsonStruc := proc.VMContext().JsonObjectCache[jsonObjectIndex]
	if jsonStruc == nil {
		proc.VM().Logger().Error("JsonPutInt jsonObjectIndex invalid", "jsonIndex", jsonObjectIndex)
		return -1
	}

	key, err := proc.ReadString(int64(keyIndex))
	if err != nil {
		proc.VM().Logger().Error("JsonPutInt read key", "err", err)
		return -1
	}

	jsonStruc[key], err = json.Marshal(int(valIndex))
	if err != nil {
		proc.VM().Logger().Error("JsonPutInt value json Marshal", "err", err)
		return -1
	}

    return 0
}

func JsonPutString(proc *exec.Process, jsonObjectIndex int32, keyIndex int32,  valIndex int32) int32 {
	jsonStruc := proc.VMContext().JsonObjectCache[jsonObjectIndex]
	if jsonStruc == nil {
		proc.VM().Logger().Error("JsonPutString jsonObjectIndex invalid", "jsonIndex", jsonObjectIndex)
		return -1
	}

	key, err := proc.ReadString(int64(keyIndex))
	if err != nil {
		proc.VM().Logger().Error("JsonPutString read key", "err", err)
		return -1
	}
	val, err := proc.ReadString(int64(valIndex))
	if err != nil {
		proc.VM().Logger().Error("JsonPutString read value", "err", err)
		return -1
	}

	jsonStruc[key], err = json.Marshal(strings.ToLower(val))
	if err != nil {
		proc.VM().Logger().Error("JsonPutString value json Marshal", "err", err)
		return -1
	}

	return 0
}

func JsonToString (proc *exec.Process, jsonObjectIndex int32) uint64{
	jsonStruc := proc.VMContext().JsonObjectCache[jsonObjectIndex]
	if jsonStruc == nil {
		proc.VM().Logger().Error("JsonPutString jsonObjectIndex invalid", "jsonIndex", jsonObjectIndex)
		return 0
	}

	jsonBytes, err := json.Marshal(&jsonStruc)
	if err != nil {
		proc.VM().Logger().Error(" JsonToString, Marshal", "err", err)
		return 0
	}

	pointer, err := proc.VM().SetBytes(jsonBytes)
	if err != nil {
		proc.VM().Logger().Error("JsonGetInt SetBytes", "err", err)
	}

	return pointer
}

func ContractCall(proc *exec.Process, contractIndex int32, methodIndex int32, paramJsonIndex int32, rtnType int32) int64 {
	toReadContractAddr, err := proc.ReadString(int64(contractIndex))
	if err != nil {
		proc.VM().Logger().Error("ContractCall read ContractName err", "err", err)
		return -1
	}

	toReadMethodName, err := proc.ReadString(int64(methodIndex))
	if err != nil {
		proc.VM().Logger().Error("ContractCall read MethodName err", "err", err)
		return -1
	}

	toReadJsonParam, err := proc.ReadString(int64(paramJsonIndex))
	if err != nil {
		proc.VM().Logger().Error("ContractCall read jsonParam err", "err", err)
		return -1
	}

	toReadRTNType, err := proc.ReadString(int64(rtnType))
	if err != nil {
		proc.VM().Logger().Error("ContractCall read rtnType err", "err", err)
		return -1
	}

	cInfo, err := context.GetBCContext().LoadContract(toReadContractAddr)
	if err != nil {
		proc.VM().Logger().Error("ContractCall LoadContract err", "err", err)
		return -1
	}

	params := make([]*types.Param, 0)
    err =  json.Unmarshal([]byte(toReadJsonParam), params)
    if err != nil {
		proc.VM().Logger().Error("ContractCall json.Unmarshal err", "JsonParam", toReadJsonParam, "err", err)
		return -1
	}

    contrInvoker := proc.VM().ContrInvoker()
    if contrInvoker == nil {
		proc.VMContext().PushVM(proc.VM())
		rtnIndex, _ := proc.VM().ContrInvoker().InvokeInternal(cInfo.Addr, cInfo.Owner, proc.VM().OwnerAddr(), proc.VMContext(), cInfo.Codes[types.CodePrefixLen:], cInfo.Name, toReadMethodName, params, toReadRTNType)
		lastVM, _:= proc.VMContext().PopVM()
		proc.VMContext().SetRunningVM(lastVM)
		switch rtnIndex.(type) {
		case int32:
			return int64(rtnIndex.(int32))
		case int64:
			return rtnIndex.(int64)
		case string:
			lastVM.SetBytes([]byte(rtnIndex.(string)))
		default:
			return -1
		}
	}else {
		proc.VM().Logger().Error("ContractCall there is no contrInvoker set")
	}

    return -1
}

func ContractDelegateCall(proc *exec.Process, contractIndex int32, methodIndex int32, paramJsonIndex int32, rtnType int32) int64 {
	toReadContractAddr, err := proc.ReadString(int64(contractIndex))
	if err != nil {
		proc.VM().Logger().Error("ContractDelegateCall read ContractName err", "err", err)
		return -1
	}

	toReadMethodName, err := proc.ReadString(int64(methodIndex))
	if err != nil {
		proc.VM().Logger().Error("ContractDelegateCall read MethodName err", "err", err)
		return -1
	}

	toReadJsonParam, err := proc.ReadString(int64(paramJsonIndex))
	if err != nil {
		proc.VM().Logger().Error("ContractDelegateCall read jsonParam err", "err", err)
		return -1
	}

	toReadRTNType, err := proc.ReadString(int64(rtnType))
	if err != nil {
		proc.VM().Logger().Error("ContractDelegateCall read rtnType err", "err", err)
		return -1
	}

	cInfo, err := context.GetBCContext().LoadContract(toReadContractAddr)
	if err != nil {
		proc.VM().Logger().Error("ContractDelegateCall LoadContract err", "err", err)
		return -1
	}

	params := make([]*types.Param, 0)
	err =  json.Unmarshal([]byte(toReadJsonParam), params)
	if err != nil {
		proc.VM().Logger().Error("ContractDelegateCall json.Unmarshal err", "JsonParam", toReadJsonParam, "err", err)
		return -1
	}

	contrInvoker := proc.VM().ContrInvoker()
	if contrInvoker == nil {
		proc.VMContext().PushVM(proc.VM())
		rtnIndex, _ := proc.VM().ContrInvoker().InvokeInternal(cInfo.Addr, cInfo.Owner, proc.VM().CallerAddr(), proc.VMContext(), cInfo.Codes[types.CodePrefixLen:], cInfo.Name, toReadMethodName, params, toReadRTNType)
		lastVM, _:= proc.VMContext().PopVM()
		proc.VMContext().SetRunningVM(lastVM)
		switch rtnIndex.(type) {
		case int32:
			return int64(rtnIndex.(int32))
		case int64:
			return rtnIndex.(int64)
		case string:
			lastVM.SetBytes([]byte(rtnIndex.(string)))
		default:
			return -1
		}
	}else {
		proc.VM().Logger().Error("ContractDelegateCall there is no contrInvoker set")
	}

	return -1
}

func TrigEvent(proc *exec.Process, evSrcIndex int32, dataIndex int32) int32 {
	evSrc, err := proc.ReadString(int64(evSrcIndex))
	if err != nil {
		proc.VM().Logger().Error("TrigEvent read event source", "err", err)
		return -1
	} else {
		proc.VM().Logger().Error("TrigEvent event source", "evSrc", evSrc)
	}

	evData, err := proc.ReadString(int64(dataIndex))
	if err != nil {
		proc.VM().Logger().Error("TrigEvent read event data", "err", err)
		return -1
	} else {
		proc.VM().Logger().Error("TrigEvent event data", "evData", evData)
	}

	return 0
}