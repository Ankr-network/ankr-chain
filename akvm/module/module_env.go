package module

import (
	"reflect"

	"github.com/Ankr-network/wagon/wasm"
)

type ModuleEnv struct {
	wasmModule    *wasm.Module
	importedFuncs map[string]*wasm.Function
}

func NewModuleEnv() *ModuleEnv {
	mEnv :=  &ModuleEnv{
		wasm.NewModule(),
		make(map[string]*wasm.Function),
	}

	mEnv.RegisterImportedFunc(PrintIFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(Print_i),
		Name: PrintIFunc,
	})

	mEnv.RegisterImportedFunc(PrintSFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(Print_s),
		Name: PrintSFunc,
	})

	mEnv.RegisterImportedFunc(StrlenFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(Strlen),
		Name: StrlenFunc,
	})

	mEnv.RegisterImportedFunc(StrcmpFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(Strcmp),
		Name: StrcmpFunc,
	})

	mEnv.RegisterImportedFunc(StrcatFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(Strcat),
		Name: StrcatFunc,
	})

	mEnv.RegisterImportedFunc(AtoiFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(Atoi),
		Name: AtoiFunc,
	})

	mEnv.RegisterImportedFunc(ItoaFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(Itoa),
		Name: ItoaFunc,
	})

	mEnv.RegisterImportedFunc(BigIntSubFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(BigIntSub),
		Name: BigIntSubFunc,
	})

	mEnv.RegisterImportedFunc(BigIntAddFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(BigIntAdd),
		Name: BigIntAddFunc,
	})

	mEnv.RegisterImportedFunc(BigIntCmpFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(BigIntCmp),
		Name: BigIntCmpFunc,
	})

	mEnv.RegisterImportedFunc(JsonObjectIndexFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(JsonObjectIndex),
		Name: JsonObjectIndexFunc,
	})

	mEnv.RegisterImportedFunc(JsonCreateObjectFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(JsonCreateObject),
		Name: JsonCreateObjectFunc,
	})

	mEnv.RegisterImportedFunc(JsonGetIntFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(JsonGetInt),
		Name: JsonGetIntFunc,
	})

	mEnv.RegisterImportedFunc(JsonGetStringFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(JsonGetString),
		Name: JsonGetStringFunc,
	})

	mEnv.RegisterImportedFunc(JsonPutIntFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(JsonPutInt),
		Name: JsonPutIntFunc,
	})

	mEnv.RegisterImportedFunc(JsonPutStringFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(JsonPutString),
		Name: JsonPutStringFunc,
	})

	mEnv.RegisterImportedFunc(JsonToStringFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(JsonToString),
		Name: JsonToStringFunc,
	})

	mEnv.RegisterImportedFunc(ContractCallFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(ContractCall),
		Name: JsonToStringFunc,
	})

	mEnv.RegisterImportedFunc(ContractDelegateCallFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(ContractDelegateCall),
		Name: ContractDelegateCallFunc,
	})

	mEnv.RegisterImportedFunc(TrigEventFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(TrigEvent),
		Name: TrigEventFunc,
	})

	mEnv.RegisterImportedFunc(SenderAddrFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(SenderAddr),
		Name: SenderAddrFunc,
	})

	mEnv.RegisterImportedFunc(OwnerAddrFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(OwnerAddr),
		Name: OwnerAddrFunc,
	})

	mEnv.RegisterImportedFunc(ChangeContractOwnerFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32},ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(ChangeContractOwner),
		Name: ChangeContractOwnerFunc,
	})

	mEnv.RegisterImportedFunc(SetBalanceFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(SetBalance),
		Name: SetBalanceFunc,
	})

	mEnv.RegisterImportedFunc(BalanceFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(Balance),
		Name: BalanceFunc,
	})

	mEnv.RegisterImportedFunc(SetAllowanceFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(SetAllowance),
		Name: SetAllowanceFunc,
	})

	mEnv.RegisterImportedFunc(AllowanceFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(Allowance),
		Name: AllowanceFunc,
	})

	mEnv.RegisterImportedFunc(CreateCurrencyFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(CreateCurrency),
		Name: CreateCurrencyFunc,
	})

	mEnv.RegisterImportedFunc(ContractAddrFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(ContractAddr),
		Name: ContractAddrFunc,
	})

	mEnv.RegisterImportedFunc(BuildCurrencyCAddrMapFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(BuildCurrencyCAddrMap),
		Name: BuildCurrencyCAddrMapFunc,
	})

	mEnv.RegisterImportedFunc(HeightFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(Height),
		Name: HeightFunc,
	})

	mEnv.RegisterImportedFunc(IsContractNormalFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(IsContractNormal),
		Name: IsContractNormalFunc,
	})

	mEnv.RegisterImportedFunc(SuspendContractFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(SuspendContract),
		Name: SuspendContractFunc,
	})

	mEnv.RegisterImportedFunc(UnsuspendContractFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(UnsuspendContract),
		Name: UnsuspendContractFunc,
	})

	mEnv.RegisterImportedFunc(StoreJsonObjectFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(StoreJsonObject),
		Name: StoreJsonObjectFunc,
	})

	mEnv.RegisterImportedFunc(LoadJsonObjectFunc, &wasm.Function{
		Sig: &wasm.FunctionSig{ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32}, ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32}},
		Body: &wasm.FunctionBody{},
		Host: reflect.ValueOf(LoadJsonObject),
		Name: LoadJsonObjectFunc,
	})

	return mEnv
}

func (m *ModuleEnv) ImportedFuncTable() map[string]*wasm.Function {
	return m.importedFuncs
}

func (m *ModuleEnv) WasmModule() *wasm.Module {
	return m.wasmModule
}

func (m *ModuleEnv) RegisterImportedFunc(name string, importedFunc *wasm.Function) {
	m.importedFuncs[name] = importedFunc
}
