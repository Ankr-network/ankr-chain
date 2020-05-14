package integration

import (
	"fmt"
	"github.com/Ankr-network/ankr-chain/client"
	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/crypto"
	"github.com/Ankr-network/ankr-chain/tx/serializer"
	"github.com/Ankr-network/ankr-chain/tx/token"
	"math/big"
	"sync"
	"testing"
)

func TestTxTransferWithNodeDup(t *testing.T) {
	msgHeader := client.TxMsgHeader{
		ChID: "Ankr-chain",
		GasLimit: new(big.Int).SetUint64(20000000).Bytes(),
		GasPrice: ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(10000000000000).Bytes()},
		Memo: "test transfer",
		Version: "1.0.2",
	}

	c := client.NewClient("http://178.128.211.3:26657")

	resp := &ankrcmm.NonceQueryResp{}
	err := c.Query("/store/nonce", &ankrcmm.NonceQueryReq{"FF5245F641038D7AEA19863E6F6486421FA30865E3E34E"}, resp)
	if err != nil {
		return
	}

	nonce := resp.Nonce

	amount, _ := new(big.Int).SetString("150000000000000000000", 10)

	tfMsg := &token.TransferMsg{FromAddr: "FF5245F641038D7AEA19863E6F6486421FA30865E3E34E",
		ToAddr:  "C9DAEE774002099874EEF0EF1DFB742F5F256D96A8C@AK",
		Amounts: []ankrcmm.Amount{ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, amount.Bytes()}},
	}

	txSerializer := serializer.NewTxSerializerCDC()

	key := crypto.NewSecretKeyEd25519("6Fk/rOEtLA7WZLhFhNtriNzHmZkQnKqRIyxYxvcG2XcxF4vdfOPtVAZ4Gv24rEK3SBr3VJ/H0bFimJqqd9lmjQ==")

	builder := client.NewTxMsgBuilder(msgHeader, tfMsg,  txSerializer, key)

	txBytes, err := builder.BuildOnly(nonce)
	if err != nil {
		return
	}

	var wg1 sync.WaitGroup
	wg1.Add(2)

	go func() {
		c := client.NewClient("http://157.245.128.17:26657")
		txHash, cHeight, _, err := c.BroadcastTxCommitWithoutWaiting(txBytes)
		if err == nil {
			fmt.Printf("txHash=%s, cHeight=%d\n", txHash, cHeight)
		} else {
			fmt.Printf("txHash=%s, err=%s\n", txHash, err.Error())
		}

		wg1.Done()
	}()

	go func() {
		c := client.NewClient("http://138.197.153.52:26657")
		txHash, cHeight, _, err := c.BroadcastTxCommitWithoutWaiting(txBytes)
		if err == nil {
			fmt.Printf("txHash=%s, cHeight=%d\n", txHash, cHeight)
		} else {
			fmt.Printf("txHash=%s, err=%s\n", txHash, err.Error())
		}

		wg1.Done()
	}()

	wg1.Wait()
}


