package integration

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"
	"testing"

	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/common/code"
	"github.com/Ankr-network/ankr-chain/consensus"
	"github.com/Ankr-network/ankr-chain/crypto"
	"github.com/Ankr-network/ankr-chain/tx"
	"github.com/Ankr-network/ankr-chain/tx/serializer"
	"github.com/Ankr-network/ankr-chain/tx/token"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
)

func TestTxTransfer(t *testing.T) {
	tfMsg := &token.TransferMsg{FromAddr: "B508ED0D54597D516A680E7951F18CAD24C7EC9FCFCD67",
		ToAddr:  "454D92DC842F532683E820DF6C3784473AD9CCF222D8FB",
		Amounts: []ankrcmm.Amount{ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(6000000000000000000).Bytes()}},
	}

	txGasLimit := new(big.Int).SetUint64(1000).Bytes()
	txMsg := &tx.TxMsg{ChID: "ankr-chain", Nonce: 0, GasLimit: txGasLimit, GasPrice: ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(10000000000000).Bytes()}, Memo: "transfermsg testing", ImplTxMsg: tfMsg}
	t.Logf("txMsg:%v\n", txMsg)

	txSerializer := serializer.NewTxSerializerCDC()

	key := crypto.NewSecretKeyEd25519("wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg==")

	sendBytes, err := txMsg.SignAndMarshal(txSerializer, key)
	assert.Equal(t, err, nil)

	log := log.NewNopLogger()
	txContext := ankrchain.NewMockAnkrChainApplication("testApp", log)

	txM, err := txContext.TxSerializer().DeserializeCDCV1(sendBytes)
	assert.Equal(t, err, nil)

	t.Logf("txM:%v", txM)

	t.Logf("tx type:%s", txM.Type())

	tfMsgD := txM.ImplTxMsg.(*token.TransferMsg)
	fmt.Printf("tfMsgD: %v\n", tfMsgD)

	_, errLog := txM.BasicVerify(txContext)
	assert.Equal(t, errLog, "")

	respCheckTx := txM.CheckTx(txContext)
	assert.Equal(t, respCheckTx.Code, code.CodeTypeOK)

	respDeliverTx := txM.DeliverTx(txContext)
	assert.Equal(t, respDeliverTx.Code, code.CodeTypeOK)
}

func TestBigCmp(t *testing.T) {
	bigOne := new(big.Int).SetUint64(10000000000000)
	bigTwo, _ := new(big.Int).SetString("100000000000000", 10)

	cmpR := bigOne.Cmp(bigTwo)

	t.Logf("cmpR=%d", cmpR)
}

func TestReadContract(t *testing.T) {
	rawBytes, err := ioutil.ReadFile("F:/GoPath/src/github.com/Ankr-network/ankr-chain/contract/example/cpp/ResourceEscrow.wasm")
	if err != nil {
		t.Errorf("can't read wasm file: %s", err.Error())
	}

	fmt.Printf("%s\n", strings.Replace(strings.Trim(fmt.Sprint(rawBytes), "[]"), " ", ",", -1))
}
