package contract

import (
	"fmt"
	"github.com/Ankr-network/wagon/exec/gas"
	"math/big"
	"strconv"
	"time"

	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/common/code"
	ankrcontext "github.com/Ankr-network/ankr-chain/context"
	ankrcrypto "github.com/Ankr-network/ankr-chain/crypto"
	"github.com/Ankr-network/ankr-chain/store/appstore"
	"github.com/Ankr-network/ankr-chain/tx"
	txcmm "github.com/Ankr-network/ankr-chain/tx/common"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type ContractDeployMsg struct {
	FromAddr string   `json:"fromaddr"`
	Name     string   `json:"name"`
	Codes    []byte   `json:"codes"`
	CodesDesc string  `json:"codesdesc"`
}

func (cd *ContractDeployMsg) SignerAddr() []string {
	return []string {cd.FromAddr}
}

func (cd *ContractDeployMsg) Type() string {
	return txcmm.TxMsgTypeContractDeployMsg
}

func (cd *ContractDeployMsg) Bytes(txSerializer tx.TxSerializer) []byte {
	bytes, _ :=  txSerializer.MarshalJSON(cd)
	return bytes
}

func (cd *ContractDeployMsg) SetSecretKey(sk ankrcrypto.SecretKey) {

}

func (cd *ContractDeployMsg) SecretKey() ankrcrypto.SecretKey {
	return &ankrcrypto.SecretKeyEd25519{}
}

func (cd *ContractDeployMsg) PermitKey(store appstore.AppStore, pubKey []byte) bool {
	return true
}

func (cd *ContractDeployMsg) SenderAddr() string {
	return cd.FromAddr
}

func (cd *ContractDeployMsg) ProcessTx(context tx.ContextTx, metric gas.GasMetric, flag tx.TxExeFlag) (uint32, string, []cmn.KVPair){
	if len(cd.FromAddr) != ankrcmm.KeyAddressLen {
		return  code.CodeTypeInvalidAddress, fmt.Sprintf("ContractDeployMsg ProcessTx, unexpected from address. Got %s, addr len=%d", cd.FromAddr, len(cd.FromAddr)), nil
	}

	if len(cd.Codes) <= ankrcmm.CodePrefixLen {
		return code.CodeTypeContractInvalidCodeSize, fmt.Sprintf("ContractDeployMsg ProcessTx, invalid code size, Got %v, code size=%d", cd.Codes, len(cd.Codes)), nil
	}

	nonce, _, _, _, _ := context.AppStore().Nonce(cd.FromAddr, 0, false)
	contractAddr := ankrcrypto.CreateContractAddress(cd.FromAddr, nonce)
    cInfo, _, _, _, err := context.AppStore().LoadContract(contractAddr, 0, false)
    if err != nil {
    	return code.CodeTypeLoadContractErr, fmt.Sprintf("ContractDeployMsg ProcessTx, load contract err: contractAddr=%s", contractAddr), nil
	} else if cInfo != nil {
		return code.CodeTypeContractAddrTakenUp, fmt.Sprintf("ContractDeployMsg ProcessTx, the contract adress has been taken up:contractAddr=%s", contractAddr), nil
	}

	if flag == tx.TxExeFlag_OnlyCheck {
		return code.CodeTypeOK, "", nil
	}

    gasUsed := uint64(len(cd.Codes)) * gas.GasContractByte
    if !metric.SpendGas(new(big.Int).SetUint64(gasUsed)) {
    	return code.CodeTypeGasNotEnough, fmt.Sprintf("ContractDeployMsg ProcessTx, gas not enough, Got %d", gasUsed), nil
	}

	cInfo = &ankrcmm.ContractInfo{contractAddr, cd.Name, cd.FromAddr, cd.Codes, cd.CodesDesc, ankrcmm.ContractNormal, make(map[string]string)}

	contractType    := ankrcmm.ContractType(cInfo.Codes[0])
	contractPatt    := ankrcmm.ContractPatternType(cInfo.Codes[2])
	contractContext := ankrcontext.NewContextContract(context.AppStore(), metric, cd, cInfo, context.AppStore(), context.AppStore(), context.Publisher())
	rtn, err := context.Contract().Call(contractContext, context.AppStore(), contractType, contractPatt, cInfo.Codes[ankrcmm.CodePrefixLen:], cInfo.Name, "init", nil, "string")
	if err != nil {
		return code.CodeTypeCallContractErr, fmt.Sprintf("call contract err: contract=%s, method=init, err=%v", contractAddr, err), nil
	}

	if !rtn.IsSuccess {
		return code.CodeTypeCallContractErr, fmt.Sprintf("call contract err: contract=%s, method=init", contractAddr), nil
	}

	if flag == tx.TxExeFlag_PreRun {
		return code.CodeTypeOK, "", nil
	}

	context.AppStore().SaveContract(contractAddr, cInfo)

	context.AppStore().AddAccount(contractAddr, ankrcmm.AccountContract)

	context.AppStore().IncNonce(cd.FromAddr)

	tvalue := time.Now().UnixNano()
	tags := []cmn.KVPair{
		{Key: []byte("app.fromaddress"), Value: []byte(cd.FromAddr)},
		{Key: []byte("app.contractaddr"), Value: []byte(contractAddr)},
		{Key: []byte("app.timestamp"), Value: []byte(strconv.FormatInt(tvalue, 10))},
		{Key: []byte("app.type"), Value: []byte(txcmm.TxMsgTypeContractDeployMsg)},
	}

	return code.CodeTypeOK, contractAddr, tags
}