package contract

import (
	"encoding/json"
	"fmt"
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
	"github.com/Ankr-network/wagon/exec/gas"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type BlankSpeedGas struct { }

type ContractInvokeMsg struct {
	FromAddr     string  `json:"fromaddr"`
	ContractAddr string  `json:"contractaddr"`
	Method       string  `json:"name"`
	Args         string  `json:"args"`
	RtnType      string  `json:"rtnType"`
}

func (bs *BlankSpeedGas)SpendGas(gas *big.Int) bool {
	return true
}

func (cd *ContractInvokeMsg) SignerAddr() []string {
	return []string {cd.FromAddr}
}

func (cd *ContractInvokeMsg) Type() string {
	return txcmm.TxMsgTypeContractInvokeMsg
}

func (ci *ContractInvokeMsg) Bytes(txSerializer tx.TxSerializer) []byte {
	bytes, _ := txSerializer.MarshalJSON(ci)
	return bytes
}

func (ci *ContractInvokeMsg) SetSecretKey(sk ankrcrypto.SecretKey) {

}

func (ci *ContractInvokeMsg) SecretKey() ankrcrypto.SecretKey {
	return &ankrcrypto.SecretKeyEd25519{}
}

func (ci *ContractInvokeMsg) PermitKey(store appstore.AppStore, pubKey []byte) bool {
	return true
}

func (ci *ContractInvokeMsg) SenderAddr() string {
	return ci.FromAddr
}

func (ci *ContractInvokeMsg) ProcessTx(context tx.ContextTx, metric gas.GasMetric, flag tx.TxExeFlag) (uint32, string, []cmn.KVPair) {
	if len(ci.FromAddr) != ankrcmm.KeyAddressLen {
		return code.CodeTypeInvalidAddress, fmt.Sprintf("ContractInvokeMsg ProcessTx, unexpected from address. Got %s, addr len=%d", ci.FromAddr, len(ci.FromAddr)), nil
	}

	if len(ci.ContractAddr) != ankrcmm.KeyAddressLen {
		return code.CodeTypeContractInvalidAddr, fmt.Sprintf("ContractInvokeMsg ProcessTx, unexpected contract address. Got %s, addr len=%d", ci.ContractAddr, len(ci.ContractAddr)), nil
	}

	cInfo, _, _, _, err := context.AppStore().LoadContract(ci.ContractAddr, 0, false)
	if err != nil {
		return code.CodeTypeLoadContractErr, fmt.Sprintf("ContractInvokeMsg ProcessTx, load contract err: contractAddr=%s", ci.ContractAddr), nil
	} else if cInfo == nil {
		return code.CodeTypeContractAddrTakenUp, fmt.Sprintf("ContractInvokeMsg ProcessTx, the contract adress has been taken up:contractAddr=%s", ci.ContractAddr), nil
	}

	if flag == tx.TxExeFlag_OnlyCheck {
		return code.CodeTypeOK, "", nil
	}


	var params []*ankrcmm.Param
	json.Unmarshal([]byte(ci.Args), &params)

	metricInjected := metric

	abi := ankrcmm.NewABIUtil(cInfo.CodesDesc)
	isAction := abi.IsAction(ci.Method)
	if !isAction {
		metricInjected = new(BlankSpeedGas)
	}

	contractType    := ankrcmm.ContractType(cInfo.Codes[0])
	contractPatt    := ankrcmm.ContractPatternType(cInfo.Codes[2])
	contractContext := ankrcontext.NewContextContract(context.AppStore(), metricInjected, ci, cInfo, context.AppStore(), context.AppStore(), context.Publisher())
	rtn, err := context.Contract().Call(contractContext, context.AppStore(), contractType, contractPatt, cInfo.Codes[ankrcmm.CodePrefixLen:], cInfo.Name, ci.Method, params, ci.RtnType)
	if err != nil {
		return code.CodeTypeCallContractErr, fmt.Sprintf("call contract err: contract=%s, method=%s, err=%v", ci.ContractAddr, ci.Method, err), nil
	}

	if !rtn.IsSuccess {
		return code.CodeTypeCallContractErr, fmt.Sprintf("call contract err: contract=%s, method=%s", ci.ContractAddr, ci.Method), nil
	}

	if flag == tx.TxExeFlag_PreRun {
		return code.CodeTypeOK, "", nil
	}

	context.AppStore().IncNonce(ci.FromAddr)

	tvalue := time.Now().UnixNano()
	tags := []cmn.KVPair{
		{Key: []byte("app.fromaddress"), Value: []byte(ci.FromAddr)},
		{Key: []byte("app.contractaddr"), Value: []byte(ci.ContractAddr)},
		{Key: []byte("app.method"), Value: []byte(ci.Method)},
		{Key: []byte("app.args"), Value: []byte(ci.Args)},
		{Key: []byte("app.timestamp"), Value: []byte(strconv.FormatInt(tvalue, 10))},
		{Key: []byte("app.type"), Value: []byte(txcmm.TxMsgTypeContractInvokeMsg)},
	}

	rtnJson, _ := json.Marshal(rtn)

	return code.CodeTypeOK, string(rtnJson), tags
}
