package native

import (
	"fmt"
	"math/big"

	"github.com/Ankr-network/ankr-chain/account"
	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/context"
	"github.com/Ankr-network/ankr-chain/store/appstore"
	"github.com/Ankr-network/wagon/exec/gas"
	"github.com/tendermint/tendermint/libs/log"
)

type AnkrCoin struct {
	name        string
	symbol      string
	decimals    uint8
	totalSupply *big.Int
	context     context.ContextContract
	log         log.Logger
}

func NewAnkrCoin(store appstore.AppStore, log log.Logger) *AnkrCoin {
	addrBytes := make([]byte, ankrcmm.KeyAddressLen/2)
	addrBytes[ankrcmm.KeyAddressLen/2-1] = 0x01
	addr := fmt.Sprintf("%X", addrBytes)
	codePrefixBytes := ankrcmm.GenerateContractCodePrefix(ankrcmm.ContractTypeNative, ankrcmm.ContractVMTypeUnknown, ankrcmm.ContractPatternTypeUnknown)
	totalSup, _ := new(big.Int).SetString("10000000000000000000000000000", 10)

	store.CreateCurrency("ANKR", &ankrcmm.CurrencyInfo{"ANKR", 18, "10000000000000000000000000000"})

	conInfo, _, _, _, err := store.LoadContract(string(addr), 0, false)
	if err == nil && conInfo == nil {
		store.BuildCurrencyCAddrMap("ANKR", addr)
		store.SaveContract(string(addr), &ankrcmm.ContractInfo{addr, "ANKR", account.AccountManagerInstance().GenesisAccountAddress(), codePrefixBytes, "", ankrcmm.ContractNormal, make(map[string]string)})
		store.SetBalance(account.AccountManagerInstance().GenesisAccountAddress(), ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, totalSup.Bytes()})
	}

	return &AnkrCoin{
		"Ankr Network",
		"ANKR", 18,
		totalSup,
		nil,
		log,
	}
}

func (ac *AnkrCoin) SetContextContract(context context.ContextContract) {
	ac.context = context
}

func (ac *AnkrCoin) Name() string {
	return ac.name
}

func (ac *AnkrCoin) Symbol() string {
	return ac.symbol
}

func (ac *AnkrCoin) Decimals() uint8 {
	return ac.decimals
}

func (ac *AnkrCoin) TotalSupply() *big.Int {
	return ac.totalSupply
}

func (ac *AnkrCoin) BalanceOf(addr string) *big.Int {
	bal, _, _, _, err := ac.context.Balance(addr, ac.symbol, 0, false)
	if err != nil {
		ac.log.Error("AnkrCoin BalanceOf err", "addr", addr, "err", err)
	}

	return bal
}

func (ac *AnkrCoin) Transfer(toAddr string, amount string) bool {
	value, isSucess := new(big.Int).SetString(amount, 10)
	if !isSucess || value == nil{
		ac.log.Error("AnkrCoin Transfer invalid amount", "amount", amount, "isSucess", isSucess)
		return false
	}

	zeroAmount := new(big.Int).SetUint64(0)
	if value.Cmp(zeroAmount) <= 0 {
		ac.log.Error("AnkrCoin Transfer amount <= 0", "amount", amount)
		return false
	}

	if toAddr == "" {
		ac.log.Error("AnkrCoin Transfer toAddr blank")
		return false
	}

	if value.Cmp(ac.totalSupply) >= 0 {
		ac.log.Error("AnkrCoin Transfer amount >= totalSupply", "amount", amount, "totalSupply", ac.totalSupply.String())
		return false
	}

	balSender := ac.BalanceOf(ac.context.SenderAddr())
	if balSender == nil || balSender.Cmp(value) == -1 {
		if balSender == nil {
			ac.log.Error("AnkrCoin Transfer sender balance nil", "senderAddr", ac.context.SenderAddr())
		} else {
			ac.log.Error("AnkrCoin Transfer sender balance less than or equal to value", "senderAddr", ac.context.SenderAddr(), "balance", balSender.String(), "value", value.String())
		}

		return false
	}

	balTo := ac.BalanceOf(toAddr)
	if balTo == nil {
		balTo = new(big.Int).SetUint64(0)
	}

	balSender = new(big.Int).Sub(balSender, value)
	balTo     = new(big.Int).Add(balTo, value)

	stepGas := uint64(100000 * 2)
	isSucess = ac.context.SpendGas(new(big.Int).SetUint64(stepGas))
	if !isSucess {
		ac.log.Error("AnkrCoin Transfer gasUsed reach the limit value after gas slow step", "senderAddr", ac.context.SenderAddr())
		return false
	}

	ac.context.SetBalance(ac.context.SenderAddr(), ankrcmm.Amount{ankrcmm.Currency{ac.symbol, 18}, balSender.Bytes()})
	ac.context.SetBalance(toAddr, ankrcmm.Amount{ankrcmm.Currency{ac.symbol, 18}, balTo.Bytes()})

	//gasUsed := uint64(len(balSender.Bytes())) * gas.GasContractByte
	//ac.context.SpendGas(new(big.Int).SetUint64(gasUsed))

	//gasUsed = uint64(len(balTo.Bytes())) * gas.GasContractByte
	//ac.context.SpendGas(new(big.Int).SetUint64(gasUsed))

	toAddrParam   := fmt.Sprintf("\"%s\"", toAddr)
	amountParam   := fmt.Sprintf("\"%s\"", amount)

	jsonArgFromat := "[{\"index\":1,\"Name\":\"toAddr\",\"ParamType\":\"string\",\"Value\":%s}," +
		              "{\"index\":2,\"Name\":\"amount\",\"ParamType\":\"string\",\"Value\":%s}]"

	jsonArg := fmt.Sprintf(jsonArgFromat, toAddrParam, amountParam)

	TrigEvent("Transfer(string, string, string))", jsonArg, ac.log, ac.context)

	return true
}

func (ac *AnkrCoin) TransferFrom(fromAddr string, toAddr string, amount string) bool {
	value, isSuccess := new(big.Int).SetString(amount, 10)
	if !isSuccess || value == nil{
		ac.log.Error("AnkrCoin TransferFrom invalid amount", "isSuccess", isSuccess)
		return false
	}

	zeroAmount := new(big.Int).SetUint64(0)
	if value.Cmp(zeroAmount) <= 0 {
		ac.log.Error("AnkrCoin Transfer amount <= 0", "amount", amount)
		return false
	}

	if toAddr == "" {
		ac.log.Error("AnkrCoin TransferFrom toAddr blank")
		return false
	}

	if value.Cmp(ac.totalSupply) >= 0 {
		ac.log.Error("AnkrCoin Transfer amount >= totalSupply", "amount", amount, "totalSupply", ac.totalSupply.String())
		return false
	}

	allowAccount := ac.Allowance(ac.context.SenderAddr(), fromAddr)
	if allowAccount == nil {
		ac.log.Error("Get invalid allowance", "SenderAddr", ac.context.SenderAddr(), "fromAddr", fromAddr)
		return false
	}

	if value.Cmp(allowAccount) == 1 {
		ac.log.Error("AnkrCoin Transfer amount >= allowAccount", "amount", amount, "allowAccount", allowAccount.String())
		return false
	}

	balFrom := ac.BalanceOf(fromAddr)
	if balFrom == nil || balFrom.Cmp(value) == -1 {
		if balFrom == nil {
			ac.log.Error("AnkrCoin TransferFrom from balance nil", "fromAddr", ac.context.SenderAddr())
		} else {
			ac.log.Error("AnkrCoin TransferFrom from balance less than or equal to value", "fromAddr", ac.context.SenderAddr(), "balance", balFrom.String(), "value", value.String())
		}

		return false
	}

	balTo := ac.BalanceOf(toAddr)
	if balTo == nil {
		balTo = new(big.Int).SetUint64(0)
	}

	balFrom = new(big.Int).Sub(balFrom, value)
	balTo   = new(big.Int).Add(balTo, value)

	stepGas := gas.GasSlowStep * 2
	isSuccess = ac.context.SpendGas(new(big.Int).SetUint64(stepGas))
	if !isSuccess {
		ac.log.Error("AnkrCoin Transfer gasUsed reach the limit value after gas slow step", "senderAddr", ac.context.SenderAddr(), "stepGas", stepGas)
		return false
	}

	ac.context.SetBalance(ac.context.SenderAddr(), ankrcmm.Amount{ankrcmm.Currency{ac.symbol,18}, balFrom.Bytes()})
	ac.context.SetBalance(toAddr, ankrcmm.Amount{ankrcmm.Currency{ac.symbol, 18}, balTo.Bytes()})

	gasUsed := uint64(len(balFrom.Bytes())) * gas.GasContractByte
	isSuccess = ac.context.SpendGas(new(big.Int).SetUint64(gasUsed))
	if !isSuccess {
		ac.log.Error("AnkrCoin Transfer gasUsed reach the limit value after balFrom bytes gas", "senderAddr", ac.context.SenderAddr(), "gasUsed", gasUsed)
		return false
	}

	gasUsed = uint64(len(balTo.Bytes())) * gas.GasContractByte
	isSuccess = ac.context.SpendGas(new(big.Int).SetUint64(gasUsed))
	if !isSuccess {
		ac.log.Error("AnkrCoin Transfer gasUsed reach the limit value after balTo bytes gas", "senderAddr", ac.context.SenderAddr(), "gasUsed", gasUsed)
		return false
	}

	fromAddrParam := fmt.Sprintf("\"%s\"", fromAddr)
	toAddrParam   := fmt.Sprintf("\"%s\"", toAddr)
	amountParam   := fmt.Sprintf("\"%s\"", amount)

	jsonArgFromat := "[{\"index\":1,\"Name\":\"fromAddr\",\"ParamType\":\"string\",\"Value\":%s}," +
		              "{\"index\":2,\"Name\":\"toAddr\",\"ParamType\":\"string\",\"Value\":%s}," +
		              "{\"index\":3,\"Name\":\"amount\",\"ParamType\":\"string\",\"Value\":%s}]"

	jsonArg := fmt.Sprintf(jsonArgFromat, fromAddrParam, toAddrParam, amountParam)

	TrigEvent("transferFrom(string, string, string))", jsonArg, ac.log, ac.context)

	return true
}

func (ac *AnkrCoin) TransferFromCDCV0(fromAddr string, toAddr string, amount string) bool {
	value, isSuccess := new(big.Int).SetString(amount, 10)
	if !isSuccess || value == nil{
		ac.log.Error("AnkrCoin TransferFrom invalid amount", "isSucess", isSuccess)
	}

	if toAddr == "" {
		ac.log.Error("AnkrCoin TransferFrom toAddr blank")
		return false
	}

	balFrom := ac.BalanceOf(fromAddr)
	if balFrom == nil || balFrom.Cmp(value) == -1 || balFrom.Cmp(value) == 0 {
		if balFrom == nil {
			ac.log.Error("AnkrCoin TransferFrom from balance nil", "fromAddr", ac.context.SenderAddr())
		} else {
			ac.log.Error("AnkrCoin TransferFrom from balance less than or equal to value", "fromAddr", ac.context.SenderAddr(), "balance", balFrom.String(), "value", value.String())
		}

		return false
	}

	balTo := ac.BalanceOf(toAddr)
	if balTo == nil {
		balTo = new(big.Int).SetUint64(0)
	}

	balFrom = new(big.Int).Sub(balFrom, value)
	balTo   = new(big.Int).Add(balTo, value)

	stepGas := gas.GasSlowStep * 2
	ac.context.SpendGas(new(big.Int).SetUint64(stepGas))

	ac.context.SetBalance(ac.context.SenderAddr(), ankrcmm.Amount{ankrcmm.Currency{ac.symbol,18}, balFrom.Bytes()})
	ac.context.SetBalance(toAddr, ankrcmm.Amount{ankrcmm.Currency{ac.symbol, 18}, balTo.Bytes()})

	gasUsed := uint64(len(balFrom.Bytes())) * gas.GasContractByte
	ac.context.SpendGas(new(big.Int).SetUint64(gasUsed))

	gasUsed = uint64(len(balTo.Bytes())) * gas.GasContractByte
	ac.context.SpendGas(new(big.Int).SetUint64(gasUsed))

	fromAddrParam := fmt.Sprintf("\"%s\"", fromAddr)
	toAddrParam   := fmt.Sprintf("\"%s\"", toAddr)
	amountParam   := fmt.Sprintf("\"%s\"", amount)

	jsonArgFromat := "[{\"index\":1,\"Name\":\"fromAddr\",\"ParamType\":\"string\",\"Value\":%s}," +
		"{\"index\":2,\"Name\":\"toAddr\",\"ParamType\":\"string\",\"Value\":%s}," +
		"{\"index\":3,\"Name\":\"amount\",\"ParamType\":\"string\",\"Value\":%s}]"

	jsonArg := fmt.Sprintf(jsonArgFromat, fromAddrParam, toAddrParam, amountParam)

	TrigEvent("transferFrom(string, string, string))", jsonArg, ac.log, ac.context)

	return true
}

func (ac *AnkrCoin) Approve(spenderAddr string, amount string) bool {
	value, isSuccess := new(big.Int).SetString(amount, 10)
	if !isSuccess || value == nil{
		ac.log.Error("AnkrCoin Approve invalid amount", "isSucess", isSuccess)
		return false
	}

	zeroAmount := new(big.Int).SetUint64(0)
	if value.Cmp(zeroAmount) <= 0 {
		ac.log.Error("AnkrCoin Approve amount <= 0", "amount", amount)
		return false
	}

	if value.Cmp(ac.totalSupply) >= 0 {
		ac.log.Error("AnkrCoin Approve amount >= totalSupply", "amount", amount, "totalSupply", ac.totalSupply.String())
		return false
	}

	ac.context.SetAllowance(ac.context.SenderAddr(), spenderAddr, ankrcmm.Amount{ankrcmm.Currency{ac.symbol, 18},value.Bytes()})

	gasUsed := uint64(len(value.Bytes())) * gas.GasContractByte
	isSuccess = ac.context.SpendGas(new(big.Int).SetUint64(gasUsed))
	if !isSuccess {
		ac.log.Error("AnkrCoin Approve gasUsed reach the limit value after val bytes gas", "senderAddr", ac.context.SenderAddr(), "gasUsed", gasUsed)
		return false
	}

	spenderAddrParam   := fmt.Sprintf("\"%s\"", spenderAddr)
	amountParam   := fmt.Sprintf("\"%s\"", amount)

	jsonArgFromat := "[{\"index\":1,\"Name\":\"spenderAddr\",\"ParamType\":\"string\",\"Value\":%s}," +
		              "{\"index\":2,\"Name\":\"amount\",\"ParamType\":\"string\",\"Value\":%s}]"

	jsonArg := fmt.Sprintf(jsonArgFromat, spenderAddrParam, amountParam)

	TrigEvent("TransferFrom(string, string, string))", jsonArg, ac.log, ac.context)

	return true
}

func (ac *AnkrCoin) Allowance(ownerAddr string, spenderAddr string) *big.Int {
	allowVal, err := ac.context.Allowance(ownerAddr, spenderAddr, ac.symbol)
	if err != nil {
		ac.log.Error("AnkrCoin Allowance err", "ownerAddr", ownerAddr, "spenderAddr", spenderAddr, "err", err)
	}

	return allowVal
}

func (ac *AnkrCoin) IncreaseApproval(spenderAddr string, addedAmount string) bool{
	addedValue, isSuccess := new(big.Int).SetString(addedAmount, 10)
	if !isSuccess || addedValue == nil{
		ac.log.Error("AnkrCoin IncreaseApproval invalid addedAmount", "isSucess", isSuccess)
		return false
	}

	zeroAmount := new(big.Int).SetUint64(0)
	if addedValue.Cmp(zeroAmount) <= 0 {
		ac.log.Error("AnkrCoin IncreaseApproval addedValue <= 0", "amount", addedValue)
		return false
	}

	if addedValue.Cmp(ac.totalSupply) >= 0 {
		ac.log.Error("AnkrCoin IncreaseApproval addedValue >= totalSupply", "amount", addedValue, "totalSupply", ac.totalSupply.String())
		return false
	}

	allowVal := ac.Allowance(ac.context.SenderAddr(), spenderAddr)
	if allowVal == nil {
		ac.log.Error("AnkrCoin IncreaseApproval sender's allowance nil")
		return false
	}

	allowVal = new(big.Int).Add(allowVal, addedValue)

	stepGas := gas.GasSlowStep
	isSuccess = ac.context.SpendGas(new(big.Int).SetUint64(stepGas))
	if !isSuccess {
		ac.log.Error("AnkrCoin IncreaseApproval gasUsed reach the limit value after gas slow step", "senderAddr", ac.context.SenderAddr(), "gasUsed", stepGas)
		return false
	}

	ac.context.SetAllowance(ac.context.SenderAddr(), spenderAddr, ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18},allowVal.Bytes()})

	gasUsed := uint64(len(allowVal.Bytes())) * gas.GasContractByte
	isSuccess = ac.context.SpendGas(new(big.Int).SetUint64(gasUsed))
	if !isSuccess {
		ac.log.Error("AnkrCoin IncreaseApproval gasUsed reach the limit value after allowVal bytes gas", "senderAddr", ac.context.SenderAddr(), "gasUsed", gasUsed)
		return false
	}

	spenderAddrParam := fmt.Sprintf("\"%s\"", spenderAddr)
	addedAmountParam := fmt.Sprintf("\"%s\"", addedAmount)

	jsonArgFromat := "[{\"index\":1,\"Name\":\"spenderAddr\",\"ParamType\":\"string\",\"Value\":%s}," +
		"{\"index\":2,\"Name\":\"addedAmount\",\"ParamType\":\"string\",\"Value\":%s}]"

	jsonArg := fmt.Sprintf(jsonArgFromat, spenderAddrParam, addedAmountParam)

	TrigEvent("TransferFrom(string, string, string))", jsonArg, ac.log, ac.context)

	return true
}

func (ac *AnkrCoin) DecreaseApproval(spenderAddr string, subtractedAmount string) bool {
	subtractedValue, isSuccess := new(big.Int).SetString(subtractedAmount, 10)
	if !isSuccess || subtractedValue == nil{
		ac.log.Error("AnkrCoin DecreaseApproval invalid subtractedAmount", "isSucess", isSuccess)
		return false
	}

	zeroAmount := new(big.Int).SetUint64(0)
	if subtractedValue.Cmp(zeroAmount) <= 0 {
		ac.log.Error("AnkrCoin DecreaseApproval addedValue <= 0", "amount", subtractedValue)
		return false
	}

	if subtractedValue.Cmp(ac.totalSupply) >= 0 {
		ac.log.Error("AnkrCoin DecreaseApproval addedValue >= totalSupply", "amount", subtractedValue, "totalSupply", ac.totalSupply.String())
		return false
	}

	allowVal := ac.Allowance(ac.context.SenderAddr(), spenderAddr)
	if allowVal == nil {
		ac.log.Error("AnkrCoin DecreaseApproval sender's allowance nil")
		return false
	}

	allowVal = new(big.Int).Sub(allowVal, subtractedValue)
	stepGas := gas.GasSlowStep
	isSuccess = ac.context.SpendGas(new(big.Int).SetUint64(stepGas))
	if !isSuccess {
		ac.log.Error("AnkrCoin DecreaseApproval gasUsed reach the limit value after gas slow step", "senderAddr", ac.context.SenderAddr(), "gasUsed", stepGas)
		return false
	}

	ac.context.SetAllowance(ac.context.SenderAddr(), spenderAddr, ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18},allowVal.Bytes()})

	gasUsed := uint64(len(allowVal.Bytes())) * gas.GasContractByte
	isSuccess = ac.context.SpendGas(new(big.Int).SetUint64(gasUsed))
	if !isSuccess {
		ac.log.Error("AnkrCoin DecreaseApproval gasUsed reach the limit value after allowVal bytes gas", "senderAddr", ac.context.SenderAddr(), "gasUsed", gasUsed)
		return false
	}

	spenderAddrParam      := fmt.Sprintf("\"%s\"", spenderAddr)
	subtractedAmountParam := fmt.Sprintf("\"%s\"", subtractedAmount)

	jsonArgFromat := "[{\"index\":1,\"Name\":\"spenderAddr\",\"ParamType\":\"string\",\"Value\":%s}," +
		              "{\"index\":2,\"Name\":\"subtractedAmount\",\"ParamType\":\"string\",\"Value\":%s}]"

	jsonArg := fmt.Sprintf(jsonArgFromat, spenderAddrParam, subtractedAmountParam)

	TrigEvent("DecreaseApproval(string, string))", jsonArg, ac.log, ac.context)

	return true
}