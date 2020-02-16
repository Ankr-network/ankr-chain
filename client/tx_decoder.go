package client

import (
	"errors"
	"fmt"
	"math/big"

	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	ankrcrypto "github.com/Ankr-network/ankr-chain/crypto"
	"github.com/Ankr-network/ankr-chain/tx"
	"github.com/Ankr-network/ankr-chain/tx/metering"
	"github.com/Ankr-network/ankr-chain/tx/serializer"
	"github.com/Ankr-network/ankr-chain/tx/token"
	"github.com/Ankr-network/ankr-chain/tx/v0"
)

type v0TxHandler func(txMsgData interface{}) (*tx.TxMsg, error)

type TxDecoder struct {
	cdcTxSerializer *serializer.TxSerializerCDC
	v0TxSerializer  *v0.TxSerializerV0
	v0TxHandlerMap  map[string]v0TxHandler
}

func NewTxDecoder() *TxDecoder {
	cdcTxSerializer := serializer.NewTxSerializerCDC()
	v0TxSerializer  := &v0.TxSerializerV0{}

	txD := &TxDecoder{ cdcTxSerializer: cdcTxSerializer, v0TxSerializer: v0TxSerializer}

	txD.v0TxHandlerMap = map[string]v0TxHandler{
		ankrcmm.TrxSendPrefix :     txD.sendTxHandler,
		ankrcmm.SetCertPrefix :     txD.setCertHandler,
		ankrcmm.RemoveCertPrefix :  txD.removeCertHandler,
		ankrcmm.SetMeteringPrefix : txD.meteringHandler,
	}

	return txD
}

func (d *TxDecoder) sendTxHandler(txMsgData interface{}) (*tx.TxMsg, error) {
	trxSendSlices, ok := txMsgData.([]string)
	if !ok || len(trxSendSlices) != 6 {
		return nil, errors.New("invalid send tx")
	}

	fromS   := trxSendSlices[0]
	toS     := trxSendSlices[1]
	amountS := trxSendSlices[2]
	nonceS  := trxSendSlices[3]
	pubkeyS := trxSendSlices[4]
	sigS    := trxSendSlices[5]

	amountInt, isSuccess := new(big.Int).SetString(amountS, 10)
	if !isSuccess {
		return nil, fmt.Errorf("invalid send tx amountS: %s", amountS)
	}

	tfmsg := &token.TransferMsg{
		fromS,
		 toS,
		[]ankrcmm.Amount {{ankrcmm.Currency{"ANKR", 18}, amountInt.Bytes()}},
	}

	nonce, isSuccess := new(big.Int).SetString(nonceS, 10)
	if !isSuccess {
		return nil, fmt.Errorf("invalid send tx nonceS: %s", nonceS)
	}

	pubObj, err := ankrcrypto.DeserilizePubKey(pubkeyS)
	if err != nil {
		return nil, err
	}

	return &tx.TxMsg {
		Nonce: nonce.Uint64(),
		Signs: []ankrcrypto.Signature{{PubKey: pubObj, Signed: []byte(sigS)}},
		ImplTxMsg: tfmsg,
	}, nil
}

func (d *TxDecoder) setCertHandler(txMsgData interface{}) (*tx.TxMsg, error) {
	trxSetCertSlices, ok := txMsgData.([]string)
	if !ok || len(trxSetCertSlices) != 4 {
		return nil, errors.New("invalid set cert tx")
	}

	dcS     := trxSetCertSlices[0]
	pemB64S := trxSetCertSlices[1]
	nonceS  := trxSetCertSlices[2]
	sigS    := trxSetCertSlices[3]

	scMsg := &metering.SetCertMsg{DCName: dcS, PemBase64: pemB64S}

	nonce, isSucess := new(big.Int).SetString(nonceS, 10)
	if !isSucess {
		return nil, fmt.Errorf("invalid set cert tx nonceS: %s", nonceS)
	}

	return &tx.TxMsg {
		Nonce: nonce.Uint64(),
		Signs: []ankrcrypto.Signature{{Signed: []byte(sigS)}},
		ImplTxMsg: scMsg,
	}, nil
}

func (d *TxDecoder) removeCertHandler(txMsgData interface{}) (*tx.TxMsg, error) {
	trxSetCertSlices, ok := txMsgData.([]string)
	if !ok || len(trxSetCertSlices) != 3 {
		return nil, errors.New("invalid remove cert tx")
	}

	dcS    := trxSetCertSlices[0]
	nonceS := trxSetCertSlices[1]
	sigS   := trxSetCertSlices[2]

	rcMsg := &metering.RemoveCertMsg{DCName: dcS}

	nonce, isSucess := new(big.Int).SetString(nonceS, 10)
	if !isSucess {
		return nil, fmt.Errorf("invalid remove cert tx nonceS: %s", nonceS)
	}

	return &tx.TxMsg {
		Nonce: nonce.Uint64(),
		Signs: []ankrcrypto.Signature{{Signed: []byte(sigS)}},
		ImplTxMsg: rcMsg,
	}, nil
}

func (d *TxDecoder) meteringHandler(txMsgData interface{}) (*tx.TxMsg, error) {
	trxSetMeteringSlices, ok := txMsgData.([]string)
	if !ok || len(trxSetMeteringSlices) != 6 {
		return nil, errors.New("invalid metering tx")
	}

	dcS    := trxSetMeteringSlices[0]
	nsS    := trxSetMeteringSlices[1]
	sigxS  := trxSetMeteringSlices[2]
	sigyS  := trxSetMeteringSlices[3]
	nonceS := trxSetMeteringSlices[4]
	valueS := trxSetMeteringSlices[5]

	rcMsg := &metering.MeteringMsg{DCName: dcS, NSName: nsS, Value: valueS}

	nonce, isSuccess := new(big.Int).SetString(nonceS, 10)
	if !isSuccess {
		return nil, fmt.Errorf("invalid metering tx nonceS: %s", nonceS)
	}

	return &tx.TxMsg {
		Nonce: nonce.Uint64(),
		Signs: []ankrcrypto.Signature{{R: sigxS, S: sigyS}},
		ImplTxMsg: rcMsg,
	}, nil
}

func (d *TxDecoder) Decode(tx []byte) (*tx.TxMsg, error){
	cdcTxSerializer := serializer.NewTxSerializerCDC()
	if cdcTxSerializer == nil {
		return nil, errors.New("can't creat tx cdc serializer")
	}

	txMsg, err := cdcTxSerializer.DeserializeCDCV1(tx)
	if err != nil {
		txType, data, err := d.v0TxSerializer.Deserialize(tx)
		if err != nil {
			return nil, err
		}

		if txVoHandler, ok := d.v0TxHandlerMap[txType]; ok {
			return txVoHandler(data)
		} else {
			return nil, fmt.Errorf("unknown tx msg: %v", tx)
		}
	}

	return txMsg, nil


}
