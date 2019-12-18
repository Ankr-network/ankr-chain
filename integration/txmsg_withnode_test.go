package integration

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"sync"
	"testing"

	"github.com/Ankr-network/ankr-chain/account"
	"github.com/Ankr-network/ankr-chain/tx/contract"
	"github.com/Ankr-network/ankr-chain/tx/metering"
	"github.com/Ankr-network/ankr-chain/client"
	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/crypto"
	"github.com/Ankr-network/ankr-chain/tx/serializer"
	"github.com/Ankr-network/ankr-chain/tx/token"
	"github.com/stretchr/testify/assert"
)

const TEST_CERT = `
-----BEGIN CERTIFICATE-----
MIICKDCCAc6gAwIBAgIUVoDB7Av8NH3bhQDPajOX/AHq/zIwCgYIKoZIzj0EAwIw
dDELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAkNBMQswCQYDVQQHEwJTRjEUMBIGA1UE
CRMLTUlTU0lPTiBTVC4xDjAMBgNVBBETBTk0MTA1MQ4wDAYDVQQKEwVIVUJDQTEV
MBMGA1UEAxMMbXlodWItY2EuY29tMB4XDTE5MDYyNDA3NDk1NloXDTI5MDYyNDA3
NDk1NlowfTELMAkGA1UEBhMCVVMxCzAJBgNVBAgTAkNBMQswCQYDVQQHEwJTRjEU
MBIGA1UECRMLTUlTU0lPTiBTVC4xDjAMBgNVBBETBTk0MTA1MRMwEQYDVQQKEwpE
YXRhQ2VudGVyMRkwFwYDVQQDExBteWRhdGFjZW50ZXIuY29tMFkwEwYHKoZIzj0C
AQYIKoZIzj0DAQcDQgAEE4x4SoWjyQit98+NDaAApQIbNIUOh/wGi4rR6EmcGmFa
qKW0jHxoYr3093CQHQ5X+BVVAjsLZCSy5melIcgPLqM1MDMwDgYDVR0PAQH/BAQD
AgeAMBMGA1UdJQQMMAoGCCsGAQUFBwMCMAwGA1UdEwEB/wQCMAAwCgYIKoZIzj0E
AwIDSAAwRQIhAPqre8XQqNr6JFvEhfaZz5XHf7854zDC4H/wmLcRv5b3AiAGgiuI
PvDQFLYt8PkvJk9hH2ynYEyI6zId1KFGxBrd/g==
-----END CERTIFICATE-----`

const TEST_KEY = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIHMyEp01U2qMTNYLdQEyc9NX8F9JowMM7ODVD9ap77ENoAoGCCqGSM49
AwEHoUQDQgAEE4x4SoWjyQit98+NDaAApQIbNIUOh/wGi4rR6EmcGmFaqKW0jHxo
Yr3093CQHQ5X+BVVAjsLZCSy5melIcgPLg==
-----END EC PRIVATE KEY-----`

func TestTxTransferWithNode(t *testing.T) {
	c := client.NewClient("localhost:26657")

	msgHeader := client.TxMsgHeader{
		ChID: "test-chain-FRIERy",
		GasLimit: new(big.Int).SetUint64(20000).Bytes(),
		GasPrice: ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(100000000000000000).Bytes()},
		Memo: "test transfer",
		Version: "",
	}

	amount, _ := new(big.Int).SetString("6000000000000000000", 10)

	tfMsg := &token.TransferMsg{FromAddr: "B150DF65F0179C218DFE34E8E450331A8450661D60FCA9",
		ToAddr:  "C053853FA561AFDC33CF3CC5CB97D655121A99B4982F72",
		Amounts: []ankrcmm.Amount{ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, amount.Bytes()}},
	}

	txSerializer := serializer.NewTxSerializerCDC()

	key := crypto.NewSecretKeyEd25519("WawC9+JxxZW5Re731mvYodryGDWk3nQvlu6rN4j+0Vvs04Ms+dGzEPbu145bl0lqoocqep7tZvorggxnXR0oYA==")

	builder := client.NewTxMsgBuilder(msgHeader, tfMsg,  txSerializer, key)

	txHash, cHeight, _, err := builder.BuildAndCommit(c)

	assert.Equal(t, err, nil)

	t.Logf("TestTxTransferWithNode sucessful: txHash=%s, cHeight=%d", txHash, cHeight)

	resp := &ankrcmm.BalanceQueryResp{}
	c.Query("/store/balance", &ankrcmm.BalanceQueryReq{"C053853FA561AFDC33CF3CC5CB97D655121A99B4982F72", "ANKR"}, resp)

	t.Logf("addr=C053853FA561AFDC33CF3CC5CB97D655121A99B4982F72, bal=%s", resp.Amount)

	resp = &ankrcmm.BalanceQueryResp{}
	c.Query("/store/balance", &ankrcmm.BalanceQueryReq{"B150DF65F0179C218DFE34E8E450331A8450661D60FCA9", "ANKR"}, resp)

	t.Logf("addr=B150DF65F0179C218DFE34E8E450331A8450661D60FCA9, bal=%s", resp.Amount)
}

func TestBroadcastTxAsyncWithNode(t *testing.T) {
	c := client.NewClient("localhost:26657")

	msgHeader := client.TxMsgHeader{
		ChID: "test-chain-tPbTdZ",
		GasLimit: new(big.Int).SetUint64(1000).Bytes(),
		GasPrice: ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(100000000000000000).Bytes()},
		Memo: "TestBroadcastTxAsync",
		Version: "1.0.2",
	}

	tfMsg := &token.TransferMsg{FromAddr: "B508ED0D54597D516A680E7951F18CAD24C7EC9FCFCD67",
		ToAddr:  "065E37B3FC243B9FABB1519AB876E7632C510DC9324031",
		Amounts: []ankrcmm.Amount{ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(6000000000000000000).Bytes()}},
	}

	txSerializer := serializer.NewTxSerializerCDC()

	key := crypto.NewSecretKeyEd25519("wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg==")

	builder := client.NewTxMsgBuilder(msgHeader, tfMsg,  txSerializer, key)

	data, txHash, log, err := builder.BuildAndBroadcastAsync(c)

	assert.Equal(t, err, nil)

	t.Logf("TestBroadcastTxAsyncWithNode sucessful: data=%v, txHash=%s, log=%s", data, txHash, log)
}

func TestBroadcastTxAsyncParallelWithNode(t *testing.T) {
	c := client.NewClient("localhost:26657")

	msgHeader := client.TxMsgHeader{
		ChID: "test-chain-NoqWuO",
		GasLimit: new(big.Int).SetUint64(1000).Bytes(),
		GasPrice: ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(100000000000000000).Bytes()},
		Memo: "TestBroadcastTxAsync",
		Version: "1.0.2",
	}

	tfMsg := &token.TransferMsg{FromAddr: "B508ED0D54597D516A680E7951F18CAD24C7EC9FCFCD67",
		ToAddr:  "065E37B3FC243B9FABB1519AB876E7632C510DC9324031",
		Amounts: []ankrcmm.Amount{ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(6000000000000000000).Bytes()}},
	}

	txSerializer := serializer.NewTxSerializerCDC()

	key := crypto.NewSecretKeyEd25519("wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg==")

	builder := client.NewTxMsgBuilder(msgHeader, tfMsg,  txSerializer, key)

	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			data, txHash, log, err := builder.BuildAndBroadcastAsync(c)
			t.Logf("TestBroadcastTxAsyncWithNode sucessful: data=%v, txHash=%s, log=%s, err=%v", data, txHash, log, err)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestBroadcastTxSyncWithNode(t *testing.T) {
	c := client.NewClient("localhost:26657")

	msgHeader := client.TxMsgHeader{
		ChID: "test-chain-NoqWuO",
		GasLimit: new(big.Int).SetUint64(1000).Bytes(),
		GasPrice: ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(100000000000000000).Bytes()},
		Memo: "TestBroadcastTxSyncWithNode",
		Version: "1.0.2",
	}

	tfMsg := &token.TransferMsg{FromAddr: "B508ED0D54597D516A680E7951F18CAD24C7EC9FCFCD67",
		ToAddr:  "065E37B3FC243B9FABB1519AB876E7632C510DC9324031",
		Amounts: []ankrcmm.Amount{ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(6000000000000000000).Bytes()}},
	}

	txSerializer := serializer.NewTxSerializerCDC()

	key := crypto.NewSecretKeyEd25519("wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg==")

	builder := client.NewTxMsgBuilder(msgHeader, tfMsg,  txSerializer, key)

	data, txHash, log, err := builder.BuildAndBroadcastSync(c)

	assert.Equal(t, err, nil)

	t.Logf("TestBroadcastTxSyncWithNode sucessful: data=%v, txHash=%s, log=%s", data, txHash, log)
}

func TestBroadcastTxSyncParallelWithNode(t *testing.T) {
	c := client.NewClient("localhost:26657")

	msgHeader := client.TxMsgHeader{
		ChID: "test-chain-NoqWuO",
		GasLimit: new(big.Int).SetUint64(1000).Bytes(),
		GasPrice: ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(100000000000000000).Bytes()},
		Memo: "TestBroadcastTxSyncParallelWithNode",
		Version: "1.0.2",
	}

	tfMsg := &token.TransferMsg{FromAddr: "B508ED0D54597D516A680E7951F18CAD24C7EC9FCFCD67",
		ToAddr:  "065E37B3FC243B9FABB1519AB876E7632C510DC9324031",
		Amounts: []ankrcmm.Amount{ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(6000000000000000000).Bytes()}},
	}

	txSerializer := serializer.NewTxSerializerCDC()

	key := crypto.NewSecretKeyEd25519("wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg==")

	builder := client.NewTxMsgBuilder(msgHeader, tfMsg,  txSerializer, key)

	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			data, txHash, log, err := builder.BuildAndBroadcastSync(c)
			t.Logf("TestBroadcastTxSyncParallelWithNode sucessful: data=%v, txHash=%s, log=%s, err=%v", data, txHash, log, err)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestCertMsgWithNode(t *testing.T) {
	c := client.NewClient("chain-dev.dccn.ankr.com:26657")

	msgHeader := client.TxMsgHeader{
		ChID: "Ankr-dev-chain",
		GasLimit: new(big.Int).SetUint64(1000).Bytes(),
		GasPrice: ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(100000000000000000).Bytes()},
		Memo: "test CertMsg",
		Version: "1.0.2",
	}

	pubBS64 := account.AccountManagerInstance().AdminOpAccount(ankrcmm.AccountAdminMetering)
	addrFrom := crypto.CreateCertAddress(pubBS64,"cls-e9242b31-3f8e-4d0a-b04f-913ff9f01ffe", crypto.CertAddrTypeSet)

	t.Logf("certMsgFromAddr=%s", addrFrom)


	certMsg := &metering.SetCertMsg{FromAddr: addrFrom,
		DCName: "cls-e9242b31-3f8e-4d0a-b04f-913ff9f01ffe",
		PemBase64: TEST_CERT,
	}

	txSerializer := serializer.NewTxSerializerCDC()

	key := crypto.NewSecretKeyEd25519("wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg==")

	builder := client.NewTxMsgBuilder(msgHeader, certMsg,  txSerializer, key)

	txHash, cHeight, _, err := builder.BuildAndCommit(c)

	assert.Equal(t, err, nil)

	t.Logf("TestCertMsgWithNode:94 sucessful: txHash=%s, cHeight=%d", txHash, cHeight)

	resp := &ankrcmm.CertKeyQueryResp{}
	c.Query("/store/certkey", &ankrcmm.CertKeyQueryReq{"dc1"}, resp)

	t.Logf("pembase64=%s", resp.PEMBase64)
}

func TestMeteringWithNode(t *testing.T) {
	c := client.NewClient("localhost:26657")

	msgHeader := client.TxMsgHeader{
		ChID: "test-chain-50L9ea",
		GasLimit: new(big.Int).SetUint64(1000).Bytes(),
		GasPrice: ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(100000000000000000).Bytes()},
		Memo: "test metering",
		Version: "1.0.2",
	}

	resp := &ankrcmm.CertKeyQueryResp{}
	c.Query("/store/certkey", &ankrcmm.CertKeyQueryReq{"dc1"}, resp)

	key := crypto.NewSecretKeyPem(TEST_KEY, resp.PEMBase64,"@mert:"+"dc1_"+"ns1")

	addr, _ := key.Address()

	t.Logf("meteringtMsgFromAddr=%s", string(addr))

	certMsg := &metering.MeteringMsg{FromAddr: string(addr),
		DCName: "dc1",
		NSName: "ns1",
		Value: "value1",
	}

	txSerializer := serializer.NewTxSerializerCDC()

	builder := client.NewTxMsgBuilder(msgHeader, certMsg,  txSerializer, key)

	txHash, cHeight, _, err := builder.BuildAndCommit(c)

	assert.Equal(t, err, nil)

	t.Logf("TestCertMsgWithNode:94 sucessful: txHash=%s, cHeight=%d", txHash, cHeight)

	respMetering := &ankrcmm.MeteringQueryResp{}
	c.Query("/store/metering", &ankrcmm.MeteringQueryReq{"dc1", "ns1"}, respMetering)

	t.Logf("value=%s", respMetering.Value)
}

func TestContractDeployWithNode(t *testing.T) {
	c := client.NewClient("localhost:26657")

	msgHeader := client.TxMsgHeader{
		ChID: "test-chain-dltzyF",
		GasLimit: new(big.Int).SetUint64(10000000).Bytes(),
		GasPrice: ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(100000000000000000).Bytes()},
		Memo: "test ContractDeploy",
		Version: "1.0.2",
	}

	rawBytes, err := ioutil.ReadFile("F:/GoPath/src/github.com/Ankr-network/ankr-chain/contract/example/cpp/TestContract.wasm")
	if err != nil {
		t.Errorf("can't read wasm file: %s", err.Error())
	}

	cdMsg := &contract.ContractDeployMsg{FromAddr: "B508ED0D54597D516A680E7951F18CAD24C7EC9FCFCD67",
		Name:     "TestContract",
		Codes:     rawBytes,
		CodesDesc: "",
	}

	txSerializer := serializer.NewTxSerializerCDC()

	key := crypto.NewSecretKeyEd25519("wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg==")

	builder := client.NewTxMsgBuilder(msgHeader, cdMsg,  txSerializer, key)

	txHash, cHeight, contractAddr, err := builder.BuildAndCommit(c)

	assert.Equal(t, err, nil)

	t.Logf("TestTxTransferWithNode sucessful: txHash=%s, cHeight=%d, contractAddr=%s", txHash, cHeight, contractAddr)

	resp := &ankrcmm.ContractQueryResp{}
	c.Query("/store/contract", &ankrcmm.ContractQueryReq{contractAddr}, resp)

	t.Logf("conract=%v", resp)
}

func TestContractInvokeWithNode(t *testing.T) {
	c := client.NewClient("localhost:26657")

	msgHeader := client.TxMsgHeader{
		ChID: "test-chain-dltzyF",
		GasLimit: new(big.Int).SetUint64(10000000).Bytes(),
		GasPrice: ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(100000000000000000).Bytes()},
		Memo: "test ContractInvoke",
		Version: "1.0.2",
	}

	jsonArg := "[{\"index\":1,\"Name\":\"args\",\"ParamType\":\"string\",\"Value\":{\"testStr\":\"testFuncWithInt arg\"}}]"

	cdMsg := &contract.ContractInvokeMsg{
		FromAddr: "B508ED0D54597D516A680E7951F18CAD24C7EC9FCFCD67",
		ContractAddr: "5448AE3966FABECB07F1827EC38F87848D88CB2FB2B000",
		Method:       "testFuncWithString",
		Args:         jsonArg,
		RtnType:      "string",
	}

	txSerializer := serializer.NewTxSerializerCDC()

	key := crypto.NewSecretKeyEd25519("wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg==")

	builder := client.NewTxMsgBuilder(msgHeader, cdMsg,  txSerializer, key)

	txHash, cHeight, contractResultJson, err := builder.BuildAndCommit(c)

	assert.Equal(t, err, nil)

	var contractR ankrcmm.ContractResult
	json.Unmarshal([]byte(contractResultJson), &contractR)

	t.Logf("TestTxTransferWithNode sucessful: txHash=%s, cHeight=%d, contractR=%v", txHash, cHeight, contractR)
}

func TestContractDeployWithNodePattern1(t *testing.T) {
	c := client.NewClient("localhost:26657")

	msgHeader := client.TxMsgHeader{
		ChID: "test-chain-dltzyF",
		GasLimit: new(big.Int).SetUint64(10000000).Bytes(),
		GasPrice: ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(100000000000000000).Bytes()},
		Memo: "test ContractDeploy",
		Version: "1.0.2",
	}

	//rawBytes, err := ioutil.ReadFile("F:/GoPath/src/github.com/Ankr-network/ankr-chain/contract/example/cpp/TestContract2.wasm")
	rawBytes, err := ioutil.ReadFile("F:/GoPath/src/github.com/Ankr-network/ankr-chain/contract/example/cpp/TestContract2.wasm")
	if err != nil {
		t.Errorf("can't read wasm file: %s", err.Error())
	}

	cdMsg := &contract.ContractDeployMsg{FromAddr: "B508ED0D54597D516A680E7951F18CAD24C7EC9FCFCD67",
		Name:     "TestContract",
		Codes:     rawBytes,
		CodesDesc: "",
	}

	txSerializer := serializer.NewTxSerializerCDC()

	key := crypto.NewSecretKeyEd25519("wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg==")

	builder := client.NewTxMsgBuilder(msgHeader, cdMsg,  txSerializer, key)

	txHash, cHeight, contractAddr, err := builder.BuildAndCommit(c)

	assert.Equal(t, err, nil)

	t.Logf("TestTxTransferWithNode sucessful: txHash=%s, cHeight=%d, contractAddr=%s", txHash, cHeight, contractAddr)

	resp := &ankrcmm.ContractQueryResp{}
	c.Query("/store/contract", &ankrcmm.ContractQueryReq{contractAddr}, resp)

	t.Logf("conract=%v", resp)
}

func TestContractInvokeWithNodePattern1(t *testing.T) {
	c := client.NewClient("localhost:26657")

	msgHeader := client.TxMsgHeader{
		ChID: "test-chain-dltzyF",
		GasLimit: new(big.Int).SetUint64(10000000).Bytes(),
		GasPrice: ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(100000000000000000).Bytes()},
		Memo: "test ContractInvoke",
		Version: "1.0.2",
	}

	jsonArg := "[{\"index\":1,\"Name\":\"args\",\"ParamType\":\"string\",\"Value\":\"testFuncWithInt arg\"}]"



	cdMsg := &contract.ContractInvokeMsg{
		FromAddr: "B508ED0D54597D516A680E7951F18CAD24C7EC9FCFCD67",
		ContractAddr: "73FC39E8B6A2B8BA5D372A8A663CFA74A15B3A782FBC24",
		Method:       "testFuncWithString",
		Args:         jsonArg,
		RtnType:      "string",
	}

	txSerializer := serializer.NewTxSerializerCDC()

	key := crypto.NewSecretKeyEd25519("wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg==")

	builder := client.NewTxMsgBuilder(msgHeader, cdMsg,  txSerializer, key)

	txHash, cHeight, contractResultJson, err := builder.BuildAndCommit(c)

	assert.Equal(t, err, nil)

	var contractR ankrcmm.ContractResult
	json.Unmarshal([]byte(contractResultJson), &contractR)

	t.Logf("TestTxTransferWithNode sucessful: txHash=%s, cHeight=%d, contractR=%v", txHash, cHeight, contractR)
}

func TestQueryAccountInfoWithNode(t *testing.T) {
	c := client.NewClient("localhost:26657")

	resp := &ankrcmm.AccountQueryResp{}
	c.Query("/store/account", &ankrcmm.AccountQueryReq{"5AEBA6EB8BC51DA277CCF1EF229F0C05D9535FA36CC872"}, resp)

	t.Logf("account=%v", new(big.Int).SetBytes(resp.Amounts[0].Value).String())
}