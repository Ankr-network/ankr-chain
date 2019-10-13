package integratetesting

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/Ankr-network/ankr-chain/account"
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
		Amounts: []account.Amount{account.Amount{account.Currency{"ANKR", 18}, new(big.Int).SetUint64(6000000000000000000).Bytes()}},
	}

	txMsg := &tx.TxMsg{ChID: "ankr-chain", Nonce: 0, GasPrice: account.Amount{account.Currency{"ANKR", 18}, new(big.Int).SetUint64(1000000000000000000).Bytes()}, Memo: "transfermsg testing", ImplTxMsg: tfMsg}
	t.Logf("txMsg:%v", txMsg)

	txSerializer := serializer.NewTxSerializerCDC()

	key := crypto.NewSecretKeyEd25519("wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg==")

	sendBytes, err := txMsg.SignAndMarshal(txSerializer, key)
	assert.Equal(t, err, nil)

	log := log.NewNopLogger()
	txContext := ankrchain.NewMockAnkrChainApplication("testApp", log)

	txM, err := txContext.TxSerializer().Deserialize(sendBytes)
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