package tx

import (
	"fmt"
	"math/big"

	"github.com/Ankr-network/ankr-chain/account"
	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/common/code"
	ankrcrypto "github.com/Ankr-network/ankr-chain/crypto"
	"github.com/Ankr-network/ankr-chain/store/appstore"
	ankrtxcmm "github.com/Ankr-network/ankr-chain/tx/common"
	"github.com/Ankr-network/wagon/exec/gas"
	"github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type ImplTxMsg interface {
	SignerAddr() []string
	Type() string
	Bytes(txSerializer TxSerializer) []byte
	SetSecretKey(sk ankrcrypto.SecretKey)
	SecretKey() ankrcrypto.SecretKey
	PermitKey(store appstore.AppStore, pubKey []byte) bool
	ProcessTx(context ContextTx, metric gas.GasMetric, flag TxExeFlag) (uint32, string, []cmn.KVPair)
}

type TxExeFlag uint32
const (
	_ TxExeFlag = iota
	TxExeFlag_OnlyCheck = 0x01
	TxExeFlag_PreRun    = 0x02
	TxExeFlag_Run       = 0x03
)

type txExeFlagInfo struct {
	Flag     TxExeFlag
	GasLimit []byte
	GasUsed  *big.Int
}

type TxMsg struct {
	ChID        ankrcmm.ChainID         `json:"chainid"`
	Nonce       uint64                  `json:"nonce"`
    GasLimit    []byte                  `json:"gaslimit"`
	GasPrice    ankrcmm.Amount          `json:"gasprice"`
	GasUsed     *big.Int                `json:"gasused"`
	Signs       []ankrcrypto.Signature  `json:"signs"`
	Memo        string                  `json:"memo"`
	Version     string                  `json:"version"`
    ImplTxMsg                           `json:"data"`
}

type txSignMsg struct {
	ChID     ankrcmm.ChainID   `json:"chainid"`
	Nonce    uint64            `json:"nonce"`
	GasLimit    []byte         `json:"gaslimit"`
	GasPrice ankrcmm.Amount    `json:"gasprice"`
	Memo     string            `json:"memo"`
	Version  string            `json:"version"`
	Data     []byte            `json:"data"`
}

func spendGasExe(gasLimit []byte, gasUsed *big.Int, gas *big.Int) bool {
	gasUsedT := new(big.Int).SetUint64(gasUsed.Uint64())
	gasUsedT = new(big.Int).Add(gasUsedT, gas)

	subGas := new(big.Int).Sub(gasUsedT, new(big.Int).SetBytes(gasLimit))

	if subGas.Cmp(big.NewInt(0)) == 1 || subGas.Cmp(big.NewInt(0)) == 0 {
		return false
	}

	gasUsed.SetUint64(gasUsedT.Uint64())

	return true
}

func NewTxStateInfo(gasLimit []byte) *txExeFlagInfo {
	gasLimitT := make([]byte, len(gasLimit))
	copy(gasLimitT, gasLimit)

	gasUsedT := new(big.Int).SetUint64(0)

	return &txExeFlagInfo{TxExeFlag_PreRun, gasLimitT, gasUsedT}
}

func (txState *txExeFlagInfo) SpendGas(gas *big.Int) bool {
	return spendGasExe(txState.GasLimit, txState.GasUsed, gas)
}

func (ts txSignMsg) Bytes(txSerializer TxSerializer) []byte {
	msgBytes, err := txSerializer.MarshalJSON(&ts)
	if err != nil {
		panic(err)
	}

	return msgBytes
}

func (tx *TxMsg) signMsg(txSerializer TxSerializer) *txSignMsg {
	return &txSignMsg{
		ChID:     tx.ChID,
		Nonce:    tx.Nonce,
		GasLimit: tx.GasLimit,
		GasPrice: tx.GasPrice,
		Memo:     tx.Memo,
		Version:  tx.Version,
		Data:     tx.ImplTxMsg.Bytes(txSerializer),
	}
}

func (tx *TxMsg) SignAndMarshal(txSerializer TxSerializer, key ankrcrypto.SecretKey) ([]byte, error) {
	signMsg := tx.signMsg(txSerializer)
	if signMsg != nil {
		signMsgBytes := signMsg.Bytes(txSerializer)
		signature, err := key.Sign(signMsgBytes)
		if err != nil {
			panic(err)
		}

		tx.Signs = []ankrcrypto.Signature{*signature}

		return txSerializer.Serialize(tx)
	}

	return nil, nil
}

func (tx *TxMsg) SpendGas(gas *big.Int) bool {
	if tx.GasUsed == nil {
		tx.GasUsed = new(big.Int).SetUint64(0)
	}
	return spendGasExe(tx.GasLimit, tx.GasUsed, gas)
}

func (tx *TxMsg) verifySignature(store appstore.AppStore, txSerializer TxSerializer) (uint32, string) {
	signMsg := tx.signMsg(txSerializer)
	toVerifyBytes := signMsg.Bytes(txSerializer)
	for i, signerAddr := range tx.SignerAddr() {
		if len(signerAddr) != ankrcmm.KeyAddressLen {
			return  code.CodeTypeInvalidAddress, fmt.Sprintf("Unexpected signer address. Got %v, len=%d", signerAddr, len(signerAddr))
		}

		var pubKeyBytes []byte
		if tx.Signs[i].PubKey != nil {
			pubKeyBytes = tx.Signs[i].PubKey.Bytes()
		}

		if !tx.PermitKey(store, pubKeyBytes) {
			return code.CodeTypeNotPermitPubKey, fmt.Sprintf("not permit public key: %v", tx.Signs[i].PubKey.Bytes())
		}

		isOk := tx.SecretKey().Verify(toVerifyBytes, &tx.Signs[i])
		if !isOk {
			return code.CodeTypeVerifySignaError, fmt.Sprintf("can't pass sign verifying for signer: pubKey=%v", tx.Signs[i].PubKey.Bytes())
		}
	}

	return code.CodeTypeOK, ""
}

func (tx *TxMsg) verifyMinGasPrice(context ContextTx) (uint32, string) {
	minGasPrice := context.MinGasPrice()
	gasPriceVal := new(big.Int).SetBytes(tx.GasPrice.Value)
	minGasPriceVal :=  new(big.Int).SetBytes(minGasPrice.Value)
	if tx.GasPrice.Cur.Symbol != minGasPrice.Cur.Symbol || gasPriceVal.Cmp(minGasPriceVal) == -1 {
		return code.CodeTypeGasPriceIrregular, fmt.Sprintf("irregular tx gas price: txGasSymbol=%s, txGasPriceVal=%s, minGasSymbol=%s, minGasPriceVal=%s",
			tx.GasPrice.Cur.Symbol, gasPriceVal.String(),
			minGasPrice.Cur.Symbol, minGasPriceVal.String())
	}

	return code.CodeTypeOK, ""
}

func (tx *TxMsg) BasicVerify(context ContextTx) (uint32, string) {
	codeV, log := tx.verifySignature(context.AppStore(), context.TxSerializer())
	if codeV != code.CodeTypeOK {
		return codeV, log
	}

	onceStore, _, _, _, err := context.AppStore().Nonce(tx.SignerAddr()[0], 0, false)
	if err != nil {
		return code.CodeTypeGetStoreNonceError, err.Error()
	}
	if onceStore != tx.Nonce && (tx.Nonce == 0 && onceStore != tx.Nonce + 1){
		return code.CodeTypeBadNonce, fmt.Sprintf("bad nonce: address=%s, onceStore=%d, tx.Nonce=%d", tx.SignerAddr()[0], onceStore, tx.Nonce)
	}

    return code.CodeTypeOK, ""
}

func (tx *TxMsg) verifyFromAddress() (uint32, string) {
	if len(tx.SignerAddr()) > 0 && tx.SignerAddr()[0] != "" {
		sk := tx.SecretKey()
		switch sk.(type) {
		case *ankrcrypto.SecretKeyEd25519:
			if tx.Type() == ankrtxcmm.TxMsgTypeSetCertMsg || tx.Type() != ankrtxcmm.TxMsgTypeRemoveCertMsg {
				return code.CodeTypeOK, ""
			}

			addr := tx.Signs[0].PubKey.Address().String()
			if string(addr) != tx.SignerAddr()[0] {
				return code.CodeTypeInvalidFromAddr, fmt.Sprintf("mismatch from addr: got addr=%s, expected addr=%s", tx.SignerAddr()[0], addr)
			}
		}
	}

	return code.CodeTypeOK, ""
}

func (tx *TxMsg) preRunForCheckTx(context ContextTx) types.ResponseCheckTx {
	txSInfo := NewTxStateInfo(tx.GasLimit)
	codeT, log, _ := tx.ProcessTx(context, txSInfo, TxExeFlag_PreRun)
	if codeT != code.CodeTypeOK {
		context.AppStore().Rollback()
		return types.ResponseCheckTx{Code: codeT, Log: log}
	}

	if txSInfo.GasUsed == nil || txSInfo.GasUsed.Cmp(big.NewInt(0)) == 0 {
		return types.ResponseCheckTx{Code: code.CodeTypeOK, GasWanted: 0}
	}

	subGas := new(big.Int).Sub(txSInfo.GasUsed, new(big.Int).SetBytes(txSInfo.GasLimit))
	if subGas.Cmp(big.NewInt(0)) == 1 || subGas.Cmp(big.NewInt(0)) == 0 {
		context.AppStore().Rollback()
		return types.ResponseCheckTx{Code: code.CodeTypeGasNotEnough, Log: fmt.Sprintf("TxMsg CheckTx, gas not enough, got %s", tx.GasUsed.String())}
	}

	usedFee := new(big.Int).Mul(txSInfo.GasUsed, new(big.Int).SetBytes(tx.GasPrice.Value))
	balFrom, _, _, _, err := context.AppStore().Balance(tx.SignerAddr()[0], tx.GasPrice.Cur.Symbol, 0, false)
	if err != nil {
		context.AppStore().Rollback()
		return types.ResponseCheckTx{Code: code.CodeTypeLoadBalError, Log: fmt.Sprintf("TxMsg CheckTx, get bal err=%s， addr=%s", err.Error(), tx.SignerAddr()[0])}
	}
	if usedFee.Cmp(balFrom) == 1 || usedFee.Cmp(balFrom) == 0 {
		context.AppStore().Rollback()
		return types.ResponseCheckTx{Code: code.CodeTypeFeeNotEnough, Log: fmt.Sprintf("TxMsg CheckTx, fee not enough, got %s, expected %s", usedFee.String(), balFrom.String())}
	}

	context.AppStore().Rollback()

	return types.ResponseCheckTx{Code: code.CodeTypeOK, GasWanted: new(big.Int).SetBytes(tx.GasLimit).Int64()}
}

func (tx *TxMsg) CheckTx(context ContextTx) (respCheckTx types.ResponseCheckTx) {
	defer func() {
		if rErr := recover(); rErr != nil {
			context.AppStore().Rollback()
			respCheckTx = types.ResponseCheckTx{Code: code.CodeTypeCheckTxError, Log: fmt.Sprintf("TxMsg CheckTx, err %v", rErr)}
		}
	}()

	codeT, log := tx.BasicVerify(context)
	if codeT != code.CodeTypeOK {
		return types.ResponseCheckTx{Code: codeT, Log: log}
	}

	codeT, log = tx.verifyMinGasPrice(context)
	if codeT != code.CodeTypeOK {
		return types.ResponseCheckTx{Code: codeT, Log: log}
	}

	codeT, log = tx.verifyFromAddress()
	if codeT != code.CodeTypeOK {
		return types.ResponseCheckTx{Code: codeT, Log: log}
	}

	return tx.preRunForCheckTx(context)
}

func (tx *TxMsg) gasCharge(context ContextTx, usedFee *big.Int) error {
	foundBal, _, _, _, err := context.AppStore().Balance(account.AccountManagerInstance().FoundAccountAddress(), tx.GasPrice.Cur.Symbol, 0, false)
	if err != nil {
		return  fmt.Errorf("TxMsg DeliverTx, get bal err=%s， addr=%s", err.Error(), account.AccountManagerInstance().FoundAccountAddress())
	}
	foundBal = new(big.Int).Add(foundBal, usedFee)
	context.AppStore().SetBalance(account.AccountManagerInstance().FoundAccountAddress(), ankrcmm.Amount{ankrcmm.Currency{tx.GasPrice.Cur.Symbol, 18}, foundBal.Bytes()})

	return nil
}

func (tx *TxMsg) DeliverTx(context ContextTx) (respDeliverTx types.ResponseDeliverTx) {
	defer func() {
		if rErr := recover(); rErr != nil {
			context.AppStore().Rollback()
			respDeliverTx = types.ResponseDeliverTx{Code: code.CodeTypeDeliverTxError, Log: fmt.Sprintf("TxMsg DeliverTx, err %v", rErr)}
		}
	}()

	context.AppStore().IncTotalTx()

	codeT, log, tags := tx.ProcessTx(context, tx, TxExeFlag_Run)
	if codeT != code.CodeTypeOK {
		context.AppStore().Rollback()
		return types.ResponseDeliverTx{Code: codeT, Log: log}
	}

	if tx.GasUsed == nil || tx.GasUsed.Cmp(big.NewInt(0)) == 0 {
		return types.ResponseDeliverTx{Code: code.CodeTypeOK, Log: log, GasWanted: 0, GasUsed: 0, Tags: tags}
	}

	subGas := new(big.Int).Sub(tx.GasUsed, new(big.Int).SetBytes(tx.GasLimit))
	if subGas.Cmp(big.NewInt(0)) == 1 || subGas.Cmp(big.NewInt(0)) == 0 {
		context.AppStore().Rollback()
		return types.ResponseDeliverTx{Code: code.CodeTypeGasNotEnough, Log: fmt.Sprintf("TxMsg DeliverTx, gas not enough, got %s", tx.GasUsed.String())}
	}

	usedFee := new(big.Int).Mul(tx.GasUsed, new(big.Int).SetBytes(tx.GasPrice.Value))
	balFrom, _, _, _, err := context.AppStore().Balance(tx.SignerAddr()[0], tx.GasPrice.Cur.Symbol, 0, false)
	if err != nil {
		context.AppStore().Rollback()
		return types.ResponseDeliverTx{Code: code.CodeTypeLoadBalError, Log: fmt.Sprintf("TxMsg DeliverTx, get bal err=%s， addr=%s", err.Error(), tx.SignerAddr()[0])}
	}
	if usedFee.Cmp(balFrom) == 1 || usedFee.Cmp(balFrom) == 0 {
		context.AppStore().Rollback()
		err = tx.gasCharge(context, balFrom)
		if err != nil {
			context.AppStore().Rollback()
		}
		return types.ResponseDeliverTx{Code: code.CodeTypeFeeNotEnough, Log: fmt.Sprintf("TxMsg DeliverTx, fee not enough, got %s, expected %s", usedFee.String(), balFrom.String())}
	}

	balFrom = new(big.Int).Sub(balFrom, usedFee)

	context.AppStore().SetBalance(tx.SignerAddr()[0], ankrcmm.Amount{ankrcmm.Currency{tx.GasPrice.Cur.Symbol, 18}, balFrom.Bytes()})

	err = tx.gasCharge(context, usedFee)
	if err != nil {
		context.AppStore().Rollback()
		return types.ResponseDeliverTx{Code: code.CodeTypeGasChargeError, Log: fmt.Sprintf("TxMsg DeliverTx, gas charge err=%s", err.Error())}
	}

	return types.ResponseDeliverTx{Code: code.CodeTypeOK, Log: log, GasWanted: new(big.Int).SetBytes(tx.GasLimit).Int64(), GasUsed: tx.GasUsed.Int64(), Tags: tags}
}
