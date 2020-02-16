package integration

import (
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/Ankr-network/ankr-chain/client"
	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/crypto"
	"github.com/Ankr-network/ankr-chain/tx/serializer"
	"github.com/Ankr-network/ankr-chain/tx/token"
	"github.com/stretchr/testify/assert"
)

func dumpAccountsBal(accounts []string, t *testing.T, c *client.Client) {
	for _, address := range accounts {
		resp := &ankrcmm.BalanceQueryResp{}
		c.Query("/store/balance", &ankrcmm.BalanceQueryReq{address, "ANKR"}, resp)

		fmt.Printf("addr=%s, bal=%s\n", address, resp.Amount)
	}
}

func TestMultiAccountToOneWithSingleNode(t *testing.T) {

	accountMap := map[string]string{
		"284BA3577E33A81A5A23C03A152EC492284FF6AA9141ED":"zHiP1sjD7nna88mYSRqMficV+Y0DVDKm7vEVeASuiQVSfhD2H3S4qBy2giP0DDuoy3nL6UL1A8Nbv/Ucb5cAOg==",
		"A1A908C22DFAD31DDB8FA16AEA0D0ECDC55C44B32D3342":"mNFXYwBpdO1+Xj9KJ6uExTGBtMGVjcLCFRLI3eL0+5/aEmnXKW7faPzWioduH1ljWtmRbt30s+Y57yXl9qwJ9w==",
		"A3CCCF3C744D653FC7211FA497015E02057B75118C79B8":"krEsrSWdZ2PPyzl6ecQZ04a+NDpDYP3z/MGM38xm9PRgfsYu3tGokwJFcCm+AY8Ac5yPZrB/uGjBoyZUpxsRhg==",
		"B37D4DDBC83B3A459E00C3153FF1325435E498BF02DC0E":"nx4MJa8Iv60bysCmJL5NHdfxu4sOH8z5aeEBe3IBDf8CGsGGut7tL1F5m4LUJlRVTFWsXP+k2i6XBZ50fXxcvg==",
		"F1B6713F3839ED9B609DE3605967778E2117E63F4A9E11":"SpxyN7N6KiQ9uReXJiK5w0Z1pFuOdKyb6Eua3v0XjJI4c0TzlN9dZJ1iQsLy+9h3dVA260cvI8Xc8rUnkguI4g==",
	} //adddress->pri

	c := client.NewClient("localhost:26657")

	msgHeader := client.TxMsgHeader{
		ChID: "test-chain-LqLg1M",
		GasLimit: new(big.Int).SetUint64(1000).Bytes(),
		GasPrice: ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(10000000000000).Bytes()},
		Memo: "test transfer",
		Version: "1.0.2",
	}

	var addArray []string
	for addr, _ := range accountMap {
		addArray = append(addArray, addr)
	}
	addArray = append(addArray, "64BC85F08C03F42B17EAAF5AFFAF9BFAF96CFCB85CA2F3")

	dumpAccountsBal(addArray, t, c)
	t.Logf("Begin desposit to other accounts")

	//tranfer ANKR to other accounts
	var wg1 sync.WaitGroup
	wg1.Add(5)
	for addr, _ := range accountMap {
		go func(toAddr string) {
			amount, _ := new(big.Int).SetString("100000000000000000000", 10)

			tfMsg := &token.TransferMsg{FromAddr: "B508ED0D54597D516A680E7951F18CAD24C7EC9FCFCD67",
				ToAddr:  toAddr,
				Amounts: []ankrcmm.Amount{ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, amount.Bytes()}},
			}

			txSerializer := serializer.NewTxSerializerCDC()

			key := crypto.NewSecretKeyEd25519("wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg==")

			builder := client.NewTxMsgBuilder(msgHeader, tfMsg, txSerializer, key)

			txHash, cHeight, _, err := builder.BuildAndCommit(c)

			assert.Equal(t, err, nil)

			fmt.Printf("TestTxTransferWithNode sucessful: txHash=%s, cHeight=%d, toAddr=%s\n", txHash, cHeight, toAddr)
			wg1.Done()
		}(addr)
	}
	wg1.Wait()

	dumpAccountsBal(addArray, t, c)
	fmt.Printf("Begin desposit between other accounts(one to multi)\n")

	var wg2 sync.WaitGroup
	wg2.Add(4)
    fromAddr := ""
    fromPri  := ""
	for addr, pri := range accountMap {
		if fromAddr == "" && fromPri == ""{
			fromAddr = addr
			fromPri = pri
			continue
		}

		go func(toAddr string) {
			amount, _ := new(big.Int).SetString("10000000000000000000", 10)

			tfMsg := &token.TransferMsg{FromAddr: fromAddr,
				ToAddr:  toAddr,
				Amounts: []ankrcmm.Amount{ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, amount.Bytes()}},
			}

			txSerializer := serializer.NewTxSerializerCDC()

			key := crypto.NewSecretKeyEd25519(fromPri)

			builder := client.NewTxMsgBuilder(msgHeader, tfMsg, txSerializer, key)

			txHash, cHeight, _, err := builder.BuildAndCommit(c)

			assert.Equal(t, err, nil)

			fmt.Printf("TestTxTransferWithNode sucessful: txHash=%s, cHeight=%d, toAddr=%s\n", txHash, cHeight, toAddr)
			wg2.Done()
		}(addr)
	}
	wg2.Wait()

	dumpAccountsBal(addArray, t, c)
	fmt.Printf("Begin desposit between other accounts(multi to one\n")

	var wg3 sync.WaitGroup
	wg3.Add(4)
	toAddr := ""

	for addr, pri := range accountMap {
		if toAddr == "" {
			toAddr = addr
			continue
		}

		go func(fromAddr string, fromPri string) {
			amount, _ := new(big.Int).SetString("10000000000000000000", 10)

			tfMsg := &token.TransferMsg{FromAddr: fromAddr,
				ToAddr:  toAddr,
				Amounts: []ankrcmm.Amount{ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, amount.Bytes()}},
			}

			txSerializer := serializer.NewTxSerializerCDC()

			key := crypto.NewSecretKeyEd25519(fromPri)

			builder := client.NewTxMsgBuilder(msgHeader, tfMsg, txSerializer, key)

			txHash, cHeight, _, err := builder.BuildAndCommit(c)

			assert.Equal(t, err, nil)

			fmt.Printf("TestTxTransferWithNode sucessful: txHash=%s, cHeight=%d, toAddr=%s\n", txHash, cHeight, toAddr)
			wg3.Done()
		}(addr, pri)
	}
	wg3.Wait()

	fmt.Printf("all end\n")
	dumpAccountsBal(addArray, t, c)
}

func TestMultiAccountToOneWithMultiNode(t *testing.T) {

	accountMap := map[string]string{
		"284BA3577E33A81A5A23C03A152EC492284FF6AA9141ED":"zHiP1sjD7nna88mYSRqMficV+Y0DVDKm7vEVeASuiQVSfhD2H3S4qBy2giP0DDuoy3nL6UL1A8Nbv/Ucb5cAOg==",
		"A1A908C22DFAD31DDB8FA16AEA0D0ECDC55C44B32D3342":"mNFXYwBpdO1+Xj9KJ6uExTGBtMGVjcLCFRLI3eL0+5/aEmnXKW7faPzWioduH1ljWtmRbt30s+Y57yXl9qwJ9w==",
		"A3CCCF3C744D653FC7211FA497015E02057B75118C79B8":"krEsrSWdZ2PPyzl6ecQZ04a+NDpDYP3z/MGM38xm9PRgfsYu3tGokwJFcCm+AY8Ac5yPZrB/uGjBoyZUpxsRhg==",
		"B37D4DDBC83B3A459E00C3153FF1325435E498BF02DC0E":"nx4MJa8Iv60bysCmJL5NHdfxu4sOH8z5aeEBe3IBDf8CGsGGut7tL1F5m4LUJlRVTFWsXP+k2i6XBZ50fXxcvg==",
		"F1B6713F3839ED9B609DE3605967778E2117E63F4A9E11":"SpxyN7N6KiQ9uReXJiK5w0Z1pFuOdKyb6Eua3v0XjJI4c0TzlN9dZJ1iQsLy+9h3dVA260cvI8Xc8rUnkguI4g==",
	} //adddress->pri

	cs :=  []*client.Client{
		client.NewClient("localhost:26697"),
		client.NewClient("localhost:26687"),
		client.NewClient("localhost:26677"),
		client.NewClient("localhost:26667"),
		client.NewClient("localhost:26657"),
	}

	msgHeader := client.TxMsgHeader{
		ChID: "Ankr-test-chain",
		GasLimit: new(big.Int).SetUint64(1000).Bytes(),
		GasPrice: ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(10000000000000).Bytes()},
		Memo: "test transfer",
		Version: "1.0.2",
	}

	var addArray []string
	for addr, _ := range accountMap {
		addArray = append(addArray, addr)
	}
	addArray = append(addArray, "64BC85F08C03F42B17EAAF5AFFAF9BFAF96CFCB85CA2F3")

	dumpAccountsBal(addArray, t, cs[0])
	t.Logf("Begin desposit to other accounts")

	//tranfer ANKR to other accounts
	var wg1 sync.WaitGroup
	wg1.Add(5)
	i := uint(0)
	for addr, _ := range accountMap {
		go func(toAddr string, cIndex uint) {
			amount, _ := new(big.Int).SetString("100000000000000000000", 10)

			tfMsg := &token.TransferMsg{FromAddr: "B508ED0D54597D516A680E7951F18CAD24C7EC9FCFCD67",
				ToAddr:  toAddr,
				Amounts: []ankrcmm.Amount{ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, amount.Bytes()}},
			}

			txSerializer := serializer.NewTxSerializerCDC()

			key := crypto.NewSecretKeyEd25519("wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg==")

			builder := client.NewTxMsgBuilder(msgHeader, tfMsg, txSerializer, key)

			txHash, cHeight, _, err := builder.BuildAndCommit(cs[cIndex])

			assert.Equal(t, err, nil)

			fmt.Printf("TestTxTransferWithNode sucessful: txHash=%s, cHeight=%d, toAddr=%s\n", txHash, cHeight, toAddr)
			wg1.Done()
		}(addr, i)
		i++
	}
	wg1.Wait()

	dumpAccountsBal(addArray, t, cs[0])
	fmt.Printf("Begin desposit between other accounts(one to multi)\n")

	var wg2 sync.WaitGroup
	wg2.Add(4)
	fromAddr := ""
	fromPri  := ""
	i = uint(0)
	for addr, pri := range accountMap {
		if fromAddr == "" && fromPri == ""{
			fromAddr = addr
			fromPri = pri
			continue
		}

		go func(toAddr string, cIndex uint) {
			amount, _ := new(big.Int).SetString("10000000000000000000", 10)

			tfMsg := &token.TransferMsg{FromAddr: fromAddr,
				ToAddr:  toAddr,
				Amounts: []ankrcmm.Amount{ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, amount.Bytes()}},
			}

			txSerializer := serializer.NewTxSerializerCDC()

			key := crypto.NewSecretKeyEd25519(fromPri)

			builder := client.NewTxMsgBuilder(msgHeader, tfMsg, txSerializer, key)

			txHash, cHeight, _, err := builder.BuildAndCommit(cs[cIndex])

			assert.Equal(t, err, nil)

			fmt.Printf("TestTxTransferWithNode sucessful: txHash=%s, cHeight=%d, toAddr=%s\n", txHash, cHeight, toAddr)
			wg2.Done()
		}(addr, i)
		i++
	}
	wg2.Wait()

	dumpAccountsBal(addArray, t, cs[0])
	fmt.Printf("Begin desposit between other accounts(multi to one\n")

	var wg3 sync.WaitGroup
	wg3.Add(4)
	toAddr := ""
	i = uint(0)
	for addr, pri := range accountMap {
		if toAddr == "" {
			toAddr = addr
			continue
		}

		go func(fromAddr string, fromPri string, cIndex uint) {
			amount, _ := new(big.Int).SetString("10000000000000000000", 10)

			tfMsg := &token.TransferMsg{FromAddr: fromAddr,
				ToAddr:  toAddr,
				Amounts: []ankrcmm.Amount{ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, amount.Bytes()}},
			}

			txSerializer := serializer.NewTxSerializerCDC()

			key := crypto.NewSecretKeyEd25519(fromPri)

			builder := client.NewTxMsgBuilder(msgHeader, tfMsg, txSerializer, key)

			txHash, cHeight, _, err := builder.BuildAndCommit(cs[cIndex])

			assert.Equal(t, err, nil)

			fmt.Printf("TestTxTransferWithNode sucessful: txHash=%s, cHeight=%d, toAddr=%s\n", txHash, cHeight, toAddr)
			wg3.Done()
		}(addr, pri, i)
		i++
	}
	wg3.Wait()

	fmt.Printf("all end\n")
	dumpAccountsBal(addArray, t, cs[i])
}

func TestDumpAllAccounts(t *testing.T) {
	accountMap := map[string]string{
		"284BA3577E33A81A5A23C03A152EC492284FF6AA9141ED":"zHiP1sjD7nna88mYSRqMficV+Y0DVDKm7vEVeASuiQVSfhD2H3S4qBy2giP0DDuoy3nL6UL1A8Nbv/Ucb5cAOg==",
		"A1A908C22DFAD31DDB8FA16AEA0D0ECDC55C44B32D3342":"mNFXYwBpdO1+Xj9KJ6uExTGBtMGVjcLCFRLI3eL0+5/aEmnXKW7faPzWioduH1ljWtmRbt30s+Y57yXl9qwJ9w==",
		"A3CCCF3C744D653FC7211FA497015E02057B75118C79B8":"krEsrSWdZ2PPyzl6ecQZ04a+NDpDYP3z/MGM38xm9PRgfsYu3tGokwJFcCm+AY8Ac5yPZrB/uGjBoyZUpxsRhg==",
		"B37D4DDBC83B3A459E00C3153FF1325435E498BF02DC0E":"nx4MJa8Iv60bysCmJL5NHdfxu4sOH8z5aeEBe3IBDf8CGsGGut7tL1F5m4LUJlRVTFWsXP+k2i6XBZ50fXxcvg==",
		"F1B6713F3839ED9B609DE3605967778E2117E63F4A9E11":"SpxyN7N6KiQ9uReXJiK5w0Z1pFuOdKyb6Eua3v0XjJI4c0TzlN9dZJ1iQsLy+9h3dVA260cvI8Xc8rUnkguI4g==",
	} //adddress->pri

	cs :=  []*client.Client{
		client.NewClient("localhost:26697"),
		client.NewClient("localhost:26687"),
		client.NewClient("localhost:26677"),
		client.NewClient("localhost:26667"),
		client.NewClient("localhost:26657"),
	}

	var addArray []string
	for addr, _ := range accountMap {
		addArray = append(addArray, addr)
	}
	addArray = append(addArray, "64BC85F08C03F42B17EAAF5AFFAF9BFAF96CFCB85CA2F3")

	dumpAccountsBal(addArray, t, cs[0])
}