package cdcv0

import (
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

type TransferMsg struct {
	FromAddr string           `json:"fromaddr"`
	ToAddr   string           `json:"toaddr"`
	Amounts  []ankrcmm.Amount `json:"amounts"`
}

func (tf *TransferMsg) SignerAddr() []string {
	return []string {tf.FromAddr}
}

func (tf *TransferMsg) GasWanted() int64 {
	return 0
}

func (tf *TransferMsg) Type() string {
	return txcmm.TxMsgTypeTransfer
}

func (tf *TransferMsg) Bytes(txSerializer tx.TxSerializer) []byte {
	bytes, _ :=  txSerializer.MarshalJSON(tf)
	return bytes
}

func (tf *TransferMsg) SetSecretKey(sk ankrcrypto.SecretKey) {

}

func (tf *TransferMsg) SecretKey() ankrcrypto.SecretKey {
	return &ankrcrypto.SecretKeyEd25519{}
}

func (tf *TransferMsg) PermitKey(store appstore.AppStore, pubKey []byte) bool {
	return true
}

func (tf *TransferMsg) SenderAddr() string {
	return tf.FromAddr
}

func (tf *TransferMsg) ProcessTx(context tx.ContextTx, metric gas.GasMetric, isOnlyCheck bool) (uint32, string, []cmn.KVPair){
	if len(tf.FromAddr) != ankrcmm.KeyAddressLen {
		return  code.CodeTypeInvalidAddress, fmt.Sprintf("TransferMsg ProcessTx, unexpected from address. Got %s, addr len=%d", tf.FromAddr, len(tf.FromAddr)), nil
	}
	if len(tf.ToAddr) != ankrcmm.KeyAddressLen {
		return code.CodeTypeInvalidAddress, fmt.Sprintf("TransferMsg ProcessTx, unexpected to address. Got %s, addr len=%d", tf.ToAddr, len(tf.ToAddr)), nil
	}

	if isOnlyCheck {
		return code.CodeTypeOK, "", nil
	}

	trAmount := tf.Amounts[0]
	contractAddr, err := context.AppStore().ContractAddrBySymbol(trAmount.Cur.Symbol)
	if contractAddr == "" {
		return code.CodeTypeContractCantFound, fmt.Sprintf("TransferMsg ProcessTx, can't find the currency contract, symbol=%s", trAmount.Cur.Symbol), nil
	}

	tokenContract, _, _, _, err := context.AppStore().LoadContract(contractAddr, 0, false)
	if err != nil {
		return code.CodeTypeLoadContractErr, fmt.Sprintf("load contract err: contractAddr = %s", contractAddr), nil
	}

	params :=  []*ankrcmm.Param{{0, "", "string", tf.FromAddr},
		{1, "", "string", tf.ToAddr},
		{2, "", "string", new(big.Int).SetBytes(tf.Amounts[0].Value).String()},
	}

	contractType    := ankrcmm.ContractType(tokenContract.Codes[0])
	contractPatt    := ankrcmm.ContractPatternType(tokenContract.Codes[2])
	contractContext := ankrcontext.NewContextContract(context.AppStore(), metric, tf, tokenContract, context.AppStore(), context.AppStore(), context.Publisher())
	rtn, err := context.Contract().Call(contractContext, context.AppStore(), contractType, contractPatt, tokenContract.Codes[ankrcmm.CodePrefixLen:], trAmount.Cur.Symbol, "TransferFromCDCV0", params, "bool")
	if err != nil {
		return code.CodeTypeCallContractErr, fmt.Sprintf("call contract err: contract=%s, method=TransferFromCDCV0, err=%v", tf.Amounts[0].Cur.Symbol, err), nil
	}

	if !rtn.IsSuccess {
		return code.CodeTypeCallContractErr, fmt.Sprintf("call contract err: contract=%s, method=TransferFromCDCV0", tf.Amounts[0].Cur.Symbol), nil
	}

	context.AppStore().IncNonce(tf.FromAddr)

	tvalue := time.Now().UnixNano()
	addressIndexFrom := fmt.Sprintf("app.%s", tf.FromAddr)
	addressIndexTo   := fmt.Sprintf("app.%s", tf.ToAddr)
	tags := []cmn.KVPair{
		{Key: []byte("app.fromaddress"), Value: []byte(tf.FromAddr)},
		{Key: []byte("app.toaddress"), Value: []byte(tf.ToAddr)},
		{Key: []byte(addressIndexFrom), Value: []byte(strconv.FormatInt(1, 10))},
		{Key: []byte(addressIndexTo), Value: []byte(strconv.FormatInt(1, 10))},
		{Key: []byte("app.timestamp"), Value: []byte(strconv.FormatInt(tvalue, 10))},
		{Key: []byte("app.type"), Value: []byte(txcmm.TxMsgTypeTransfer)},
	}

	return code.CodeTypeOK, "", tags
}

