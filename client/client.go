package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/common/code"
	"github.com/Ankr-network/ankr-chain/store/appstore/iavl"
	"github.com/tendermint/go-amino"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/pubsub/query"
	tmliteErr "github.com/tendermint/tendermint/lite/errors"
	tmliteProxy "github.com/tendermint/tendermint/lite/proxy"
	"github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type Client struct {
	cHttp *client.HTTP
	cdc   *amino.Codec
}

var storeNameMap = map[string]string {
	"/store/nonce" : iavl.IavlStoreAccountKey,
	"/store/balance" : iavl.IavlStoreAccountKey,
	"/store/certkey" : iavl.IAvlStoreMainKey,
	"/store/metering" : iavl.IAvlStoreMainKey,
	"/store/validator" : iavl.IAvlStoreMainKey,
	"/store/contract" : iavl.IAvlStoreContractKey,
	"/store/account" : iavl.IavlStoreAccountKey,
	"/store/currency" : iavl.IAvlStoreContractKey,
	"/store/statisticalinfo" : iavl.IAvlStoreMainKey,
}

func getStoreName(path string) (string, error) {
	if storeName, ok := storeNameMap[path]; ok {
		return storeName, nil
	}

	return "", fmt.Errorf("invalid path: %s", path)
}

func NewClient(nodeUrl string) *Client {
	cHttp := client.NewHTTP(nodeUrl, "/websocket")

	return &Client{cHttp, amino.NewCodec()}
}

func (c *Client) Query(path string, req interface{}, resp interface{}) (err error) {
	reqDataBytes, err := c.cdc.MarshalJSON(req)
	if err != nil {
		return err
	}

	resultQ, err := c.cHttp.ABCIQuery(path, reqDataBytes)
	if err != nil {
		return err
	}

	if resultQ.Response.Code != code.CodeTypeOK {
		return fmt.Errorf("Client query response code not ok, code=%d, log=%s", resultQ.Response.Code, resultQ.Response.Log)
	}

	var qResp ankrcmm.QueryResp
	err = c.cdc.UnmarshalJSON(resultQ.Response.Value, &qResp)
	if err != nil {
		return err
	}

	c.cdc.UnmarshalJSON(qResp.RespData, resp)

	return nil
}

func (c *Client) verifyProof(home string, reqPath string, resp abcitypes.ResponseQuery, proofVal []byte) error {
     if home == "" {
     	return errors.New("home dir can't blank when need to verify")
	 }

	st, err := c.cHttp.Status()
	if err !=nil {
		return err
	}

	verifier, err := tmliteProxy.NewVerifier(
		st.NodeInfo.Network,
		home,
		c.cHttp,
		log.NewNopLogger(),
		10,
	)
	if err != nil {
		return err
	}

	sHeader, err := tmliteProxy.GetCertifiedCommit(resp.Height, c.cHttp,  verifier)
	switch {
	case tmliteErr.IsErrCommitNotFound(err):
		return fmt.Errorf("can't find commit info: %w", err)
	case err != nil:
		return err
	}

	prt := iavl.DefaultProofRuntime()

	storeName, err := getStoreName(reqPath)
	if err != nil {
		return err
	}

	kp := merkle.KeyPath{}
	kp = kp.AppendKey([]byte(storeName), merkle.KeyEncodingURL)
	kp = kp.AppendKey(resp.Key, merkle.KeyEncodingURL)

	if resp.Value == nil {
		err = prt.VerifyAbsence(resp.Proof, sHeader.Header.AppHash, kp.String())
		if err != nil {
			return fmt.Errorf("failed to prove merkle proof:%w", err)
		}
		return nil
	}
	err = prt.VerifyValue(resp.Proof, sHeader.Header.AppHash, kp.String(), proofVal)
	if err != nil {
		return fmt.Errorf("failed to prove merkle proof: %w", err)
	}
	
	return nil	
}

func (c *Client) QueryWithOption(path string, height int64, needProofVerify bool, home string, req interface{}, resp interface{}) (err error) {
	reqDataBytes, err := c.cdc.MarshalJSON(req)
	if err != nil {
		return err
	}

	resultQ, err := c.cHttp.ABCIQueryWithOptions(path, reqDataBytes, client.ABCIQueryOptions{height, needProofVerify})
	if err != nil {
		return err
	}

	if resultQ.Response.Code != code.CodeTypeOK {
		return fmt.Errorf("Client query response code not ok, code=%d, log=%s", resultQ.Response.Code, resultQ.Response.Log)
	}

	var qResp ankrcmm.QueryResp
	err = c.cdc.UnmarshalJSON(resultQ.Response.Value, &qResp)
	if err != nil {
		return err
	}

	if needProofVerify {
		err = c.verifyProof(home, path, resultQ.Response, qResp.ProofValue)
		if err != nil {
			return err
		}
	}

	c.cdc.UnmarshalJSON(qResp.RespData, resp)

	return nil
}

func (c *Client) BroadcastTxCommitWithRawResult(txBytes []byte) (*ctypes.ResultBroadcastTxCommit, error){
	result, err := c.cHttp.BroadcastTxCommit(txBytes)
	if err != nil {
		return nil, err
	}

	if result.CheckTx.Code != code.CodeTypeOK {
		return result, fmt.Errorf("Client BroadcastTxCommit CheckTx response code not ok, code=%d, log=%s", result.CheckTx.Code, result.CheckTx.Log)
	}

	err = client.WaitForHeight(c.cHttp, result.Height+1, nil)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (c *Client) BroadcastTxCommit(txBytes []byte) (txHash string, commitHeight int64, log string, err error) {
	result, err := c.cHttp.BroadcastTxCommit(txBytes)
	if err != nil {
		return "", -1, "", err
	}

	if result.CheckTx.Code != code.CodeTypeOK {
		return "", -1, "", fmt.Errorf("Client BroadcastTxCommit CheckTx response code not ok, code=%d, log=%s", result.CheckTx.Code, result.CheckTx.Log)
	}

	if result.DeliverTx.Code != code.CodeTypeOK {
		return result.Hash.String(), result.Height, "", fmt.Errorf("Client BroadcastTxCommit DeliverTx response code not ok, code=%d, log=%s", result.DeliverTx.Code, result.DeliverTx.Log)
	}

	err = client.WaitForHeight(c.cHttp, result.Height+1, nil)
	if err != nil {
		return result.Hash.String(), result.Height, "", err
	}

	return result.Hash.String(), result.Height, result.DeliverTx.Log, nil
}

func (c *Client) BroadcastTxCommitWithoutWaiting(txBytes []byte) (txHash string, commitHeight int64, log string, err error) {
	result, err := c.cHttp.BroadcastTxCommit(txBytes)
	if err != nil {
		return "", -1, "", err
	}

	if result.CheckTx.Code != code.CodeTypeOK {
		return "", -1, "", fmt.Errorf("Client BroadcastTxCommit CheckTx response code not ok, code=%d, log=%s", result.CheckTx.Code, result.CheckTx.Log)
	}

	if result.DeliverTx.Code != code.CodeTypeOK {
		return result.Hash.String(), result.Height, "", fmt.Errorf("Client BroadcastTxCommit DeliverTx response code not ok, code=%d, log=%s", result.DeliverTx.Code, result.DeliverTx.Log)
	}

	return result.Hash.String(), result.Height, result.DeliverTx.Log, nil
}

func (c *Client) Status() (*ctypes.ResultStatus, error) {
	return c.cHttp.Status()
}

func (c *Client) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	return c.cHttp.ABCIInfo()
}

func (c *Client) BroadcastTxAsync(txBytes []byte) (*ctypes.ResultBroadcastTx, error) {
	return c.cHttp.BroadcastTxAsync(txBytes)
}

func (c *Client) BroadcastTxSync(txBytes []byte) (*ctypes.ResultBroadcastTx, error) {
	return c.cHttp.BroadcastTxSync(txBytes)
}

func (c *Client) UnconfirmedTxs(limit int) (*ctypes.ResultUnconfirmedTxs, error) {
	return c.cHttp.UnconfirmedTxs(limit)
}

func (c *Client) NumUnconfirmedTxs() (*ctypes.ResultUnconfirmedTxs, error) {
	return c.cHttp.NumUnconfirmedTxs()
}

func (c *Client) NetInfo() (*ctypes.ResultNetInfo, error) {
	return c.cHttp.NetInfo()
}

func (c *Client) DumpConsensusState() (*ctypes.ResultDumpConsensusState, error) {
	return c.cHttp.DumpConsensusState()
}

func (c *Client) ConsensusState() (*ctypes.ResultConsensusState, error) {
	return c.cHttp.ConsensusState()
}

func (c *Client) Health() (*ctypes.ResultHealth, error) {
	return c.cHttp.Health()
}

func (c *Client) BlockchainInfo(minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error) {
	return c.cHttp.BlockchainInfo(minHeight, maxHeight)
}

func (c *Client) Genesis() (*ctypes.ResultGenesis, error) {
	return c.cHttp.Genesis()
}

func (c *Client) Block(height *int64) (*ctypes.ResultBlock, error) {
	return c.cHttp.Block(height)
}

func (c *Client) BlockResults(height *int64) (*ctypes.ResultBlockResults, error) {
	return c.cHttp.BlockResults(height)
}

func (c *Client) Commit(height *int64) (*ctypes.ResultCommit, error) {
	return c.cHttp.Commit(height)
}

func (c *Client) Tx(hash []byte, prove bool) (*ctypes.ResultTx, error) {
	return c.cHttp.Tx(hash, prove)
}

func (c *Client) TxSearch(query string, prove bool, page, perPage int) (*ctypes.ResultTxSearch, error) {
	return c.cHttp.TxSearch(query, prove, page, perPage)
}

func (c *Client) Validators(height *int64) (*ctypes.ResultValidators, error) {
	return c.cHttp.Validators(height)
}

func (c *Client) SubscribeAndWait(subscriber string, queryStr string, waitTimeOut time.Duration, maxSubCap int, out chan ctypes.ResultEvent) error {
	c.cHttp.Start()

	q := query.MustParse(queryStr)

	ctx, cancel := context.WithTimeout(context.Background(), waitTimeOut)
	defer cancel()

	eventCh, err := c.cHttp.Subscribe(ctx, subscriber, q.String(), maxSubCap)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case evt := <-eventCh:
				out <- evt
			}
		}
	}()

	return nil
}