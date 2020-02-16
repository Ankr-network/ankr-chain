package module

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	ankrcontext "github.com/Ankr-network/ankr-chain/context"
	"github.com/Ankr-network/wagon/exec"
	"github.com/tendermint/tendermint/types"
)

const (
	PrintSFunc                 = "print_s"
	PrintIFunc                 = "print_i"
	StrlenFunc                 = "strlen"
	StrcmpFunc                 = "strcmp"
	StrcatFunc                 = "strcat"
	AtoiFunc                   = "Atoi"
    ItoaFunc                   = "Itoa"
	BigIntSubFunc              =  "BigIntSub"
	BigIntAddFunc              =  "BigIntAdd"
	BigIntCmpFunc              =  "BigIntCmp"
	JsonObjectIndexFunc        = "JsonObjectIndex"
	JsonCreateObjectFunc       = "JsonCreateObject"
	JsonGetIntFunc             = "JsonGetInt"
	JsonGetStringFunc          = "JsonGetString"
	JsonPutIntFunc             = "JsonPutInt"
	JsonPutStringFunc          = "JsonPutString"
	JsonToStringFunc           = "JsonToString"
	ContractCallFunc           = "ContractCall"
	ContractDelegateCallFunc   = "ContractDelegateCall"
	TrigEventFunc              = "TrigEvent"
	SenderAddrFunc             = "SenderAddr"
	OwnerAddrFunc              = "OwnerAddr"
	ChangeContractOwnerFunc    = "ChangeContractOwner"
	SetBalanceFunc             = "SetBalance"
	BalanceFunc                = "Balance"
	SetAllowanceFunc           = "SetAllowance"
	AllowanceFunc              = "Allowance"
	CreateCurrencyFunc         = "CreateCurrency"
	ContractAddrFunc           = "ContractAddr"
	BuildCurrencyCAddrMapFunc  = "BuildCurrencyCAddrMap"
	HeightFunc                 = "Height"
	IsContractNormalFunc       = "IsContractNormal"
	SuspendContractFunc        = "SuspendContract"
	UnsuspendContractFunc      = "UnsuspendContract"
	StoreJsonObjectFunc        = "StoreJsonObject"
	LoadJsonObjectFunc         = "LoadJsonObject"
)

func Print_s(proc *exec.Process, strIdx int32) {
	str, err := proc.ReadString(int64(strIdx))
	if err != nil {
		proc.VM().Logger().Error("Print_s", "err", err)
		return
	}
	proc.VM().Logger().Info("Print_s", "str", str)
}

func Print_i(proc *exec.Process, v int32) {
	proc.VM().Logger().Info("Print_i", "v", v)
}

func Strlen(proc *exec.Process, strIdx int32) int32 {
	len, err := proc.VM().Strlen(uint(strIdx))
	if err != nil {
		return -1
	}

	return int32(len)
}

func Strcmp(proc *exec.Process, strIdx1 int32, strIdx2 int32) int32 {
	cmpR, _ := proc.VM().Strcmp(uint(strIdx1), uint(strIdx2))
	return int32(cmpR)
}

func Strcat(proc *exec.Process, strIdx1 int32, strIdx2 int32) uint64 {
	str1, err := proc.ReadString(int64(strIdx1))
	if err != nil {
		proc.VM().Logger().Error("Strcat str1", "err", err)
		return 0
	}

	str2, err := proc.ReadString(int64(strIdx2))
	if err != nil {
		proc.VM().Logger().Error("Strcat str2", "err", err)
		return 0
	}

	str := str1 + str2

	pointer, err := proc.VM().SetBytes([]byte(str))
	if err != nil {
		proc.VM().Logger().Error("Strcat SetBytes", "err", err)
		return 0
	}

	return pointer
}

func Atoi(proc *exec.Process, strIdx int32) int32 {
	str, err := proc.ReadString(int64(strIdx))
	if err != nil {
		proc.VM().Logger().Error("Atoi", "err", err)
		return -1
	}

	iRtn, err := strconv.Atoi(str)
	if err != nil {
		proc.VM().Logger().Error("Atoi convert error", "err", err)
		return -1
	}

	return int32(iRtn)
}

func Itoa(proc *exec.Process, iValue int32) uint64 {
	valStr := strconv.FormatInt(int64(iValue), 10)

	pointer, err := proc.VM().SetBytes([]byte(valStr))
	if err != nil {
		proc.VM().Logger().Error("ItoA SetBytes", "err", err)
		return 0
	}

	return pointer
}

func BigIntSub(proc *exec.Process, bigIntIndex1 int32, bigIntIndex2 int32) uint64 {
	bigIntStr1, err := proc.ReadString(int64(bigIntIndex1))
	if err != nil {
		proc.VM().Logger().Error("BigIntSub bigIntStr1", "err", err)
		return 0
	}
	bigInt1, isSucess := new(big.Int).SetString(bigIntStr1, 10)
	if !isSucess {
		proc.VM().Logger().Error("BigIntSub bigInt1", "bigIntStr1", bigIntStr1)
		return 0
	}

	bigIntStr2, err := proc.ReadString(int64(bigIntIndex2))
	if err != nil {
		proc.VM().Logger().Error("BigIntSub bigIntStr2", "err", err)
	}
	bigInt2, isSucess := new(big.Int).SetString(bigIntStr2, 10)
	if !isSucess {
		proc.VM().Logger().Error("BigIntSub bigInt2", "bigIntStr2", bigIntStr2)
		return 0
	}

	subBigIntStr := new(big.Int).Sub(bigInt1, bigInt2).String()

	pointer, err := proc.VM().SetBytes([]byte(subBigIntStr))
	if err != nil {
		proc.VM().Logger().Error("BigIntSub SetBytes", "err", err)
	}

	return pointer
}

func BigIntAdd(proc *exec.Process, bigIntIndex1 int32, bigIntIndex2 int32) uint64 {
	bigIntStr1, err := proc.ReadString(int64(bigIntIndex1))
	if err != nil {
		proc.VM().Logger().Error("BigIntAdd bigIntStr1", "err", err)
		return 0
	}
	bigInt1, isSucess := new(big.Int).SetString(bigIntStr1, 10)
	if !isSucess {
		proc.VM().Logger().Error("BigIntAdd bigInt1", "bigIntStr1", bigIntStr1)
		return 0
	}

	bigIntStr2, err := proc.ReadString(int64(bigIntIndex2))
	if err != nil {
		proc.VM().Logger().Error("BigIntAdd bigIntStr2", "bigIntStr2", bigIntStr2)
		return 0
	}
	bigInt2, isSucess := new(big.Int).SetString(bigIntStr2, 10)
	if !isSucess {
		proc.VM().Logger().Error("BigIntAdd bigInt2", "err", err)
		return 0
	}

	addBigIntStr := new(big.Int).Add(bigInt1, bigInt2).String()

	pointer, err := proc.VM().SetBytes([]byte(addBigIntStr))
	if err != nil {
		proc.VM().Logger().Error("BigIntAdd SetBytes", "err", err)
	}


	return pointer
}

func BigIntCmp(proc *exec.Process, bigIntIndex1 int32, bigIntIndex2 int32) int32 {
	bigIntStr1, err := proc.ReadString(int64(bigIntIndex1))
	if err != nil {
		proc.VM().Logger().Error("BigIntCmp bigIntStr1", "err", err)
		return 0
	}
	bigInt1, isSucess := new(big.Int).SetString(bigIntStr1, 10)
	if !isSucess || bigInt1 == nil {
		proc.VM().Logger().Error("BigIntCmp bigInt1", "bigIntStr1", bigIntStr1)
		return 0
	}

	bigIntStr2, err := proc.ReadString(int64(bigIntIndex2))
	if err != nil {
		proc.VM().Logger().Error("BigIntCmp bigIntStr2", "err", err)
	}
	bigInt2, isSucess := new(big.Int).SetString(bigIntStr2, 10)
	if !isSucess || bigInt2 == nil {
		proc.VM().Logger().Error("BigIntCmp bigInt2", "bigIntStr2", bigIntStr2)
		return 0
	}

	cmpR := bigInt1.Cmp(bigInt2)

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

	cInfo, _, _, _, err := ankrcontext.GetBCContext().LoadContract(toReadContractAddr, 0, false)
	if err != nil {
		proc.VM().Logger().Error("ContractCall LoadContract err", "err", err)
		return -1
	}

	params := make([]*ankrcmm.Param, 0)
    err =  json.Unmarshal([]byte(toReadJsonParam), params)
    if err != nil {
		proc.VM().Logger().Error("ContractCall json.Unmarshal err", "JsonParam", toReadJsonParam, "err", err)
		return -1
	}

    contrInvoker := proc.VM().ContrInvoker()
    if contrInvoker == nil {
		proc.VMContext().PushVM(proc.VM())
		rtnIndex, _ := proc.VM().ContrInvoker().InvokeInternal(cInfo.Addr, cInfo.Owner, proc.VM().OwnerAddr(), proc.VMContext(), cInfo.Codes[ankrcmm.CodePrefixLen:], cInfo.Name, toReadMethodName, params, toReadRTNType)
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

	cInfo, _, _, _, err := ankrcontext.GetBCContext().LoadContract(toReadContractAddr, 0, false)
	if err != nil {
		proc.VM().Logger().Error("ContractDelegateCall LoadContract err", "err", err)
		return -1
	}

	params := make([]*ankrcmm.Param, 0)
	err =  json.Unmarshal([]byte(toReadJsonParam), params)
	if err != nil {
		proc.VM().Logger().Error("ContractDelegateCall json.Unmarshal err", "JsonParam", toReadJsonParam, "err", err)
		return -1
	}

	contrInvoker := proc.VM().ContrInvoker()
	if contrInvoker == nil {
		proc.VMContext().PushVM(proc.VM())
		rtnIndex, _ := proc.VM().ContrInvoker().InvokeInternal(cInfo.Addr, cInfo.Owner, proc.VM().CallerAddr(), proc.VMContext(), cInfo.Codes[ankrcmm.CodePrefixLen:], cInfo.Name, toReadMethodName, params, toReadRTNType)
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
		proc.VM().Logger().Info("TrigEvent event source", "evSrc", evSrc)
	}

	evData, err := proc.ReadString(int64(dataIndex))
	if err != nil {
		proc.VM().Logger().Error("TrigEvent read event data", "err", err)
		return -1
	} else {
		proc.VM().Logger().Info("TrigEvent event data", "evData", evData)
	}

	runningContractAddr := proc.VMContext().RunningVM().ContractAddr()

	evSrcSegs := strings.Split(evSrc, "(")
	if len(evSrcSegs) != 2 {
		proc.VM().Logger().Error("TrigEvent event invalid evSrc", "evSrc", evSrc)
		return -1
	}

	method := evSrcSegs[0]
	//argStr := strings.TrimRight(evSrcSegs[1], ")")
	//argTyps := strings.Split(argStr, ",")

	var params []*ankrcmm.Param
	err =  json.Unmarshal([]byte(evData), &params)
	if err != nil {
		proc.VM().Logger().Error("TrigEvent event json.Unmarshal err", "evData", evData, "err", err)
		return -1
	}

	tags := make(map[string]string)
	tags["contract.addr"]   = runningContractAddr
	tags["contract.method"] = method
	for _, param := range params {
		tagName  := fmt.Sprintf("contract.method.%s", param.Name)
		tagValue :=  param.Value.(string)
		tags[tagName] = tagValue
	}

	proc.VMContext().Publisher().PublishWithTags(context.Background(), types.EventDataString("contract"), tags)

	return 0
}

func SenderAddr(proc *exec.Process) uint64 {
	addr := ankrcontext.GetBCContext().SenderAddr()
	pointer, err := proc.VM().SetBytes([]byte(addr))
	if err != nil {
		proc.VM().Logger().Error("SenderAddr SetBytes", "err", err)
		return 0
	}

	return pointer
}

func OwnerAddr(proc *exec.Process) uint64 {
	addr := ankrcontext.GetBCContext().OwnerAddr()
	pointer, err := proc.VM().SetBytes([]byte(addr))
	if err != nil {
		proc.VM().Logger().Error("OwnerAddr SetBytes", "err", err)
		return 0
	}

	return pointer
}

func ChangeContractOwner(proc *exec.Process, cAddrIndex int32, addrIndex int32) int32 {
	cAddr, err := proc.ReadString(int64(cAddrIndex))
	if err != nil {
		proc.VM().Logger().Error("ChangeContractOwner can't read addr", "err", err)
		return -1
	}

	addr, err := proc.ReadString(int64(addrIndex))
	if err != nil {
		proc.VM().Logger().Error("ChangeContractOwner can't read addr", "err", err)
		return -1
	}

	err = ankrcontext.GetBCContext().ChangeContractOwner(cAddr, addr)
	if err != nil {
		return -1
	}

	return 0
}

func SetBalance(proc *exec.Process, addrIndex int32, symbolIndex int32, amountIndex int32) int32 {
	addr, err := proc.ReadString(int64(addrIndex))
	if err != nil {
		proc.VM().Logger().Error("SetBalance can't read addr", "err", err)
		return -1
	}

	symbol, err := proc.ReadString(int64(symbolIndex))
	if err != nil {
		proc.VM().Logger().Error("SetBalance can't read symbol", "err", err)
		return -1
	}

	curInfo, _, _, _, err := ankrcontext.GetBCContext().CurrencyInfo(symbol, 0, false)
	if err != nil {
		proc.VM().Logger().Error("SetBalance can't get currency", "err", err, "symbol", symbol)
	}

	amount, err := proc.ReadString(int64(amountIndex))
	if err != nil {
		proc.VM().Logger().Error("SetBalance can't read amount", "err", err)
		return -1
	}

	amountInt, isSuc := new(big.Int).SetString(amount, 10)
	if !isSuc {
		proc.VM().Logger().Error("SetBalance invalid amountstr", "amountstr", amount)
		return -1
	}

	ankrcontext.GetBCContext().SetBalance(addr, ankrcmm.Amount{ankrcmm.Currency{symbol, curInfo.Decimal}, amountInt.Bytes()})

	return 0
}

func Balance(proc *exec.Process,  addrIndex int32, symbolIndex int32) uint64 {
	addr, err := proc.ReadString(int64(addrIndex))
	if err != nil {
		proc.VM().Logger().Error("Balance can't read addr", "err", err)
		return 0
	}

	symbol, err := proc.ReadString(int64(symbolIndex))
	if err != nil {
		proc.VM().Logger().Error("Balance can't read symbol", "err", err)
		return 0
	}

	balInt, _, _, _, err := ankrcontext.GetBCContext().Balance(addr, symbol, 0, false)
	if err != nil || balInt == nil{
		proc.VM().Logger().Error("Balance load balance", "err", err, "addr", addr, "symbol", symbol)
		balInt = new(big.Int).SetUint64(0)
	}

	pointer, err := proc.VM().SetBytes([]byte(balInt.String()))
	if err != nil {
		proc.VM().Logger().Error("SenderAddr SetBytes", "err", err)
		return 0
	}

	return pointer
}

func SetAllowance(proc *exec.Process, addrSenderIndex int32, addrSpenderIndex int32, symbolIndex int32, amountIndex int32) int32 {
	addrSender, err := proc.ReadString(int64(addrSenderIndex))
	if err != nil {
		proc.VM().Logger().Error("SetAllowance can't read addrSender", "err", err)
		return -1
	}

	addrSpender, err := proc.ReadString(int64(addrSpenderIndex))
	if err != nil {
		proc.VM().Logger().Error("SetAllowance can't read addrSpender", "err", err)
		return -1
	}

	symbol, err := proc.ReadString(int64(symbolIndex))
	if err != nil {
		proc.VM().Logger().Error("SetAllowance can't read symbol", "err", err)
		return -1
	}

	curInfo, _, _, _, err := ankrcontext.GetBCContext().CurrencyInfo(symbol, 0, false)
	if err != nil || curInfo == nil {
		proc.VM().Logger().Error("SetAllowance can't get currency", "err", err, "symbol", symbol)
		return -1
	}

	amount, err := proc.ReadString(int64(amountIndex))
	if err != nil {
		proc.VM().Logger().Error("SetAllowance can't read amount", "err", err)
		return -1
	}

	amountInt, isSuc := new(big.Int).SetString(amount, 10)
	if !isSuc {
		proc.VM().Logger().Error("SetAllowance invalid amountstr", "amountstr", amount)
		return -1
	}

	ankrcontext.GetBCContext().SetAllowance(addrSender, addrSpender, ankrcmm.Amount{ankrcmm.Currency{symbol, curInfo.Decimal}, amountInt.Bytes()})

	return 0
}

func Allowance(proc *exec.Process, addrSenderIndex int32, addrSpenderIndex int32, symbolIndex int32) uint64 {
	addrSender, err := proc.ReadString(int64(addrSenderIndex))
	if err != nil {
		proc.VM().Logger().Error("Allowance can't read addrSender", "err", err)
		return 0
	}

	addrSpender, err := proc.ReadString(int64(addrSpenderIndex))
	if err != nil {
		proc.VM().Logger().Error("Allowance can't read addrSpender", "err", err)
		return 0
	}

	symbol, err := proc.ReadString(int64(symbolIndex))
	if err != nil {
		proc.VM().Logger().Error("Allowance can't read symbol", "err", err)
		return 0
	}

	amountInt, err := ankrcontext.GetBCContext().Allowance(addrSender, addrSpender, symbol)
	if err != nil {
		proc.VM().Logger().Error("Allowance error", "err", err, "addrSender", addrSender, "addrSpender", addrSpender, "symbol", symbol)
		amountInt = new(big.Int).SetUint64(0)
	}

	pointer, err := proc.VM().SetBytes([]byte(amountInt.String()))
	if err != nil {
		proc.VM().Logger().Error("SenderAddr SetBytes", "err", err)
		return 0
	}

	return pointer
}

func CreateCurrency(proc *exec.Process, symbolIndex int32, decimal int32, totalSupplyIndex int32) int32 {
	symbol, err := proc.ReadString(int64(symbolIndex))
	if err != nil {
		proc.VM().Logger().Error("CreateCurrency can't read symbol", "err", err)
		return -1
	}

	totalSupply, err := proc.ReadString(int64(totalSupplyIndex))
	if err != nil {
		proc.VM().Logger().Error("CreateCurrency can't read total supply", "err", err)
		return -1
	}

	err = ankrcontext.GetBCContext().CreateCurrency(symbol, &ankrcmm.CurrencyInfo{symbol, int64(decimal), totalSupply})
	if err != nil {
		proc.VM().Logger().Error("CreateCurrency can't read symbol", "err", err)
		return -1
	}

	return 0
}

func ContractAddr(proc *exec.Process) uint64 {
	cAddr := ankrcontext.GetBCContext().ContractAddr()
	pointer, err := proc.VM().SetBytes([]byte(cAddr))
	if err != nil {
		proc.VM().Logger().Error("SenderAddr SetBytes", "err", err)
		return 0
	}

	return pointer
}

func BuildCurrencyCAddrMap(proc *exec.Process, symbolIndex int32, cAddrIndex int32) int32 {
	symbol, err := proc.ReadString(int64(symbolIndex))
	if err != nil {
		proc.VM().Logger().Error("BuildCurrencyCAddrMap can't read symbol", "err", err)
		return -1
	}

	cAddr, err := proc.ReadString(int64(cAddrIndex))
	if err != nil {
		proc.VM().Logger().Error("BuildCurrencyCAddrMap can't read cAddr", "err", err)
		return -1
	}

	ankrcontext.GetBCContext().BuildCurrencyCAddrMap(symbol, cAddr)

	return 0
}

func Height(proc *exec.Process) int32 {
	height := ankrcontext.GetBCContext().Height()

	return int32(height)
}

func IsContractNormal(proc *exec.Process, cAddrIndex int32) int32 {
	cAddr, err := proc.ReadString(int64(cAddrIndex))
	if err != nil {
		proc.VM().Logger().Error("IsContractNormal can't read cAddr", "err", err)
		return -1
	}

	isNormal := ankrcontext.GetBCContext().IsContractNormal(cAddr)

	if !isNormal {
		return -1
	}

	return 1
}

func SuspendContract(proc *exec.Process, cAddrIndex int32) int32 {
	cAddr, err := proc.ReadString(int64(cAddrIndex))
	if err != nil {
		proc.VM().Logger().Error("SuspendContract can't read cAddr", "err", err)
		return -1
	}

	err = ankrcontext.GetBCContext().UpdateContractState(cAddr, ankrcmm.ContractSuspend)
	if err != nil {
		proc.VM().Logger().Error("SuspendContract can't UpdateContractState", "err", err)
		return -1
	}

	return 0
}

func UnsuspendContract(proc *exec.Process, cAddrIndex int32) int32 {
	cAddr, err := proc.ReadString(int64(cAddrIndex))
	if err != nil {
		proc.VM().Logger().Error("SuspendContract can't read cAddr", "err", err)
		return -1
	}

	err = ankrcontext.GetBCContext().UpdateContractState(cAddr, ankrcmm.ContractNormal)
	if err != nil {
		proc.VM().Logger().Error("SuspendContract can't UpdateContractState", "err", err)
		return -1
	}

	return  0
}

func StoreJsonObject(proc *exec.Process, cAddrIndex int32, keyIndex int32, jsonObjIndex int32) int32 {
	cAddr, err := proc.ReadString(int64(cAddrIndex))
	if err != nil {
		proc.VM().Logger().Error("StoreJsonObject can't read addr", "err", err)
		return -1
	}

	key, err := proc.ReadString(int64(keyIndex))
	if err != nil {
		proc.VM().Logger().Error("StoreJsonObject can't read key", "err", err)
		return -1
	}

	jsonObject, err := proc.ReadString(int64(jsonObjIndex))
	if err != nil {
		proc.VM().Logger().Error("StoreJsonObject can't read jsonObject", "err", err)
		return -1
	}

	err = ankrcontext.GetBCContext().AddContractRelatedObject(cAddr, key, jsonObject)
	if err != nil {
		return -1
	}

	return 0
}

func LoadJsonObject(proc *exec.Process, cAddrIndex int32, keyIndex int32) uint64 {
	cAddr, err := proc.ReadString(int64(cAddrIndex))
	if err != nil {
		proc.VM().Logger().Error("LoadJsonObject can't read addr", "err", err)
		return 0
	}

	key, err := proc.ReadString(int64(keyIndex))
	if err != nil {
		proc.VM().Logger().Error("LoadJsonObject can't read key", "err", err)
		return 0
	}

	jsonObj, err := ankrcontext.GetBCContext().LoadContractRelatedObject(cAddr, key)
	if err != nil {
		proc.VM().Logger().Error("LoadJsonObject LoadContractRelatedObject err", "err", err)
		return 0
	}

	pointer, err := proc.VM().SetBytes([]byte(jsonObj))
	if err != nil {
		proc.VM().Logger().Error("LoadJsonObject SetBytes", "err", err)
		return 0
	}

	return pointer
}


