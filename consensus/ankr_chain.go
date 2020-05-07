package ankrchain

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/Ankr-network/ankr-chain/account"
	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/common/code"
	"github.com/Ankr-network/ankr-chain/contract"
	"github.com/Ankr-network/ankr-chain/router"
	"github.com/Ankr-network/ankr-chain/store/appstore"
	"github.com/Ankr-network/ankr-chain/store/appstore/iavl"
	"github.com/Ankr-network/ankr-chain/tx"
	txcmm "github.com/Ankr-network/ankr-chain/tx/common"
	_ "github.com/Ankr-network/ankr-chain/tx/metering"
	"github.com/Ankr-network/ankr-chain/tx/serializer"
	_ "github.com/Ankr-network/ankr-chain/tx/token"
	"github.com/Ankr-network/ankr-chain/tx/v0"
	val "github.com/Ankr-network/ankr-chain/tx/validator"
	akver "github.com/Ankr-network/ankr-chain/version"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"github.com/tendermint/tendermint/libs/log"
	tmpubsub "github.com/tendermint/tendermint/libs/pubsub"
	tmCoreTypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"
)

var _ types.Application = (*AnkrChainApplication)(nil)

type AnkrChainApplication struct {
	ChainId         ankrcmm.ChainID
	APPName         string
	latestHeight    int64
	latestAPPHash   []byte
	app             appstore.AppStore
	txSerializer    tx.TxSerializer
	contract        contract.Contract
	pubsubServer    *tmpubsub.Server
	logger          log.Logger
	minGasPrice     ankrcmm.Amount
	checkTxRecord   *sync.Map
	deliverTxRecord *sync.Map
}

func NewAppStore(dbDir string, l log.Logger) appstore.AppStore {
	appStore := iavl.NewIavlStoreApp(dbDir, l)
	router.QueryRouterInstance().AddQueryHandler("store", appStore)

	return  appStore
}

func NewMockAppStore() appstore.AppStore {
	appStore := iavl.NewMockIavlStoreApp()
	router.QueryRouterInstance().AddQueryHandler("store", appStore)

	return appStore
}

func NewAnkrChainApplication(dbDir string, appName string, l log.Logger) *AnkrChainApplication {
	appStore := NewAppStore(dbDir, l.With("module", "AppStore"))

	v0.MsgRouterInstance().SetLogger(l.With("module", "V0TxMsgRouter"))

	chainID := appStore.ChainID()

	return &AnkrChainApplication{
		ChainId:         ankrcmm.ChainID(chainID),
		APPName:         appName,
		app:             appStore,
		txSerializer:    serializer.NewTxSerializerCDC(),
		contract:        contract.NewContract(appStore, l.With("module", "contract")),
		logger:          l,
		minGasPrice:     ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(10000000000000).Bytes()},
		checkTxRecord:   &sync.Map{},
		deliverTxRecord: &sync.Map{},
	}
}

func NewMockAnkrChainApplication(appName string, l log.Logger) *AnkrChainApplication {
	appStore := NewMockAppStore()

	account.AccountManagerInstance().Init(appStore)

	return &AnkrChainApplication{
		APPName:         appName,
		app:             appStore,
		txSerializer:    serializer.NewTxSerializerCDC(),
		contract:        contract.NewContract(appStore, l.With("module", "contract")),
		logger:          l,
		minGasPrice:     ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(10000000000000).Bytes()},
		checkTxRecord:   &sync.Map{},
		deliverTxRecord: &sync.Map{},
	}
}

func (app *AnkrChainApplication) SetLogger(l log.Logger) {
	app.logger = l
}

func (app *AnkrChainApplication) MinGasPrice() ankrcmm.Amount {
	return app.minGasPrice
}

func (app *AnkrChainApplication) AppStore() appstore.AppStore {
	return app.app
}

func (app *AnkrChainApplication) Logger() log.Logger {
	return app.logger
}

func (app *AnkrChainApplication) TxSerializer() tx.TxSerializer {
	return app.txSerializer
}

func (app *AnkrChainApplication) Contract() contract.Contract {
	return app.contract
}

func (app *AnkrChainApplication) SetPubSubServer(server *tmpubsub.Server) {
	app.pubsubServer = server
}

func (app *AnkrChainApplication) Publisher() tx.Publisher {
	return app
}

func (app *AnkrChainApplication) Publish(ctx context.Context, msg interface{}) error {
	if app.pubsubServer == nil {
		app.logger.Error("current Publish not available", "msg", msg)
		return fmt.Errorf("current Publish not available: msg=%v", msg)
	}

	return app.pubsubServer.Publish(ctx, msg)
}

func (app *AnkrChainApplication) PublishWithTags(ctx context.Context, msg interface{}, tags map[string]string) error {
	if app.pubsubServer == nil {
		app.logger.Error("current PublishWithTags not available", "msg", msg)
		return fmt.Errorf("current PublishWithTags not available: msg=%v", msg)
	}

	return app.pubsubServer.PublishWithTags(ctx, msg, tags)
}

func (app *AnkrChainApplication) LatestHeight() int64 {
	return app.latestHeight
}

func (app *AnkrChainApplication) ChainStateInfo() ankrcmm.ChainStateInfo {
	return app
}

func (app *AnkrChainApplication) Info(req types.RequestInfo) types.ResponseInfo {
	return types.ResponseInfo{
		Data:             app.APPName,
		Version:          version.ABCIVersion,
		AppVersion:       akver.APPVersion,
		LastBlockHeight:  app.app.Height(),
		LastBlockAppHash: app.app.APPHash(),
	}
}

func (app *AnkrChainApplication) SetOption(req types.RequestSetOption) types.ResponseSetOption {
	return types.ResponseSetOption{}
}

func (app *AnkrChainApplication) dispossTxWithCDCV1(tx []byte) (*tx.TxMsg, uint32, string) {
	txMsg, err := app.txSerializer.DeserializeCDCV1(tx)
	if err != nil {
		if app.logger != nil {
			app.logger.Error("can't deserialize tx", "err", err)
		}
		return nil, code.CodeTypeDecodingError, fmt.Sprintf("can't deserialize tx: tx=%v, err=%s", tx, err.Error())
	} else {
		if txMsg.ChID != app.ChainId {
			return nil, code.CodeTypeMismatchChainID, fmt.Sprintf("can't mistach the chain id, txChainID=%s, appChainID=%s", txMsg.ChID, app.ChainId)
		}

		if txMsg.Type() == txcmm.TxMsgTypeTransfer &&  txMsg.Version != "1.0.2" {
			return nil, code.CodeTypeMismatchTxVersion, fmt.Sprintf("expected version 1.0.2 for new tx transfer, txVersion=%s", txMsg.Version)
		}
	}

	return txMsg, code.CodeTypeOK, ""
}

func (app *AnkrChainApplication) dispossTxWithCDCV0(tx []byte) (*tx.TxMsgCDCV0, uint32, string) {
	txMsg, err := app.txSerializer.DeserializeCDCV0(tx)
	if err != nil {
		if app.logger != nil {
			app.logger.Error("can't deserialize tx", "err", err)
		}
		return nil, code.CodeTypeDecodingError, fmt.Sprintf("can't deserialize tx: tx=%v, err=%s", tx, err.Error())
	} else {
		if txMsg.ChID != app.ChainId {
			return nil, code.CodeTypeMismatchChainID, fmt.Sprintf("can't mistach the chain id, txChainID=%s, appChainID=%s", txMsg.ChID, app.ChainId)
		}

		if txMsg.Version != "0.31.5" &&  txMsg.Version != "1.0" && txMsg.Version != "" {
			return nil, code.CodeTypeMismatchTxVersion, fmt.Sprintf("can't mistach the tx version(0.31.5, 1.0 or 0), txVersion=%s", txMsg.Version)
		}
	}

	return txMsg, code.CodeTypeOK, ""
}

func (app *AnkrChainApplication) DeliverTx(tx []byte) types.ResponseDeliverTx {
	txHash := fmt.Sprintf("%X", tmhash.Sum(tx))
	if _, ok := app.deliverTxRecord.Load(txHash); ok {
		app.logger.Error("AnkrChainApplication DeliverTx duplicate tx error", "txHash", txHash)
		return types.ResponseDeliverTx{Code: code.CodeTypeDuplicateTxError, Log: fmt.Sprintf("duplicated tx, txHash=%s", txHash)}
	}

	app.deliverTxRecord.Store(txHash, true)

	txMsg, codeVal, logStr := app.dispossTxWithCDCV1(tx)
	if codeVal == code.CodeTypeOK {
		rd := txMsg.DeliverTx(app)
		if rd.Code != code.CodeTypeOK {
			app.deliverTxRecord = &sync.Map{}
		}
		return rd
	} else {
		app.logger.Info("AnkrChainApplication DeliverTx new tx cdcv1 serialize error, switch to cdcv0 tx", "logStr", logStr)
		txMsgCDCV0, codeVal, logStr := app.dispossTxWithCDCV0(tx)
		if codeVal == code.CodeTypeOK {
			rdCDCV0 := txMsgCDCV0.DeliverTx(app)
			if rdCDCV0.Code != code.CodeTypeOK {
				app.deliverTxRecord = &sync.Map{}
			}
			return rdCDCV0
		}

		app.logger.Info("AnkrChainApplication DeliverTx new tx cdcv0 serialize error, switch to V0 tx", "logStr", logStr)
	}

	rdV0 := v0.MsgRouterInstance().DeliverTx(tx, app.AppStore())
	if rdV0.Code != code.CodeTypeOK {
		app.deliverTxRecord = &sync.Map{}
	}
	return rdV0
}

func (app *AnkrChainApplication) CheckTx(tx []byte) types.ResponseCheckTx {
	txHash := fmt.Sprintf("%X", tmhash.Sum(tx))
	if _, ok := app.checkTxRecord.Load(txHash); ok {
		app.logger.Error("AnkrChainApplication CheckTx duplicate tx error", "txHash", txHash)
		return types.ResponseCheckTx{Code: code.CodeTypeDuplicateTxError, Log: fmt.Sprintf("duplicated tx, txHash=%s", txHash)}
	}

	app.checkTxRecord.Store(txHash, true)

	txMsg, codeVal, logStr := app.dispossTxWithCDCV1(tx)
	if codeVal == code.CodeTypeOK {
		rc := txMsg.CheckTx(app)
		if rc.Code != code.CodeTypeOK {
			app.checkTxRecord.Delete(txHash)
		}
		return rc
	} else {
		app.checkTxRecord.Delete(txHash)
	}

	return types.ResponseCheckTx{ Code: codeVal, Log: logStr}
}

// Commit will panic if InitChain was not called
func (app *AnkrChainApplication) Commit() types.ResponseCommit {
	/*
	if app.app.KVState().Size > 0 && app.latestHeight == app.app.KVState().Height {
		rtnResp :=  types.ResponseCommit{Data: app.app.KVState().AppHash}

		app.app.SetTotalTx(app.app.KVState().Size)
		app.app.ResetKVState()

		return rtnResp
	}*/


	appHashH := app.app.APPHashByHeight(app.latestHeight-1)
	if appHashH == nil {
		app.logger.Info("AnkrChainApplication Commit appHashH nil\n")
	}

	if app.latestAPPHash == nil {
		app.logger.Info("AnkrChainApplication Commit app.latestAPPHash nil\n")
	}

	if  appHashH != nil && app.latestAPPHash != nil && !bytes.Equal(appHashH, app.latestAPPHash) {
		panic(fmt.Errorf("AnkrChainApplication Commit appHash check error, height=%d. Got %X, expected %X", app.latestHeight, appHashH, app.latestAPPHash))
	}

	rc := app.app.Commit()

	app.deliverTxRecord = &sync.Map{}

	return rc
}

func (app *AnkrChainApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	defer func() {
		if rErr := recover(); rErr != nil {
			resQuery.Code = code.CodeTypeQueryInvalidQueryReqData
			resQuery.Log  = fmt.Sprintf("AnkrChainApplication Query, excetion catched, invalid query req data, path=%s, err=%v", reqQuery.Path, rErr)
			app.logger.Error("AnkrChainApplication Query, excetion catched, invalid query req data", "path", reqQuery.Path, "err", rErr)
		}
	}()

	qHandler, subPath := router.QueryRouterInstance().QueryHandler(reqQuery.Path)
	if qHandler == nil {
		resQuery.Code = code.CodeTypeQueryInvalidQueryReqData
		resQuery.Log  = fmt.Sprintf("AnkrChainApplication Query, invalid query req data, path=%s", reqQuery.Path)
		app.logger.Error("AnkrChainApplication Query, invalid query req data", "path", reqQuery.Path)
		return
	}

	reqQuery.Path = subPath
	return qHandler.Query(reqQuery)
}

// Save the validators in the merkle tree
func (app *AnkrChainApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	var initTotalPowers int64
	for _, v := range req.Validators {
		initTotalPowers += v.Power

		if initTotalPowers > tmCoreTypes.MaxTotalVotingPower {
			app.logger.Error("The init total validator powers reach max %d", "maxtotalvalidatorpower", tmCoreTypes.MaxTotalVotingPower)
			return types.ResponseInitChain{}
		}

		err := val.ValidatorManagerInstance().InitValidator(&v, app.app)
		if err != nil {
			app.logger.Error("InitChain error updating validators", "err", err)
		}
	}

	sbytes := string(req.AppStateBytes)
	if len(sbytes) > 0 {
		sbytes = sbytes[1 : len(sbytes)-1]
		addressAndBalance := strings.Split(sbytes, ":")
		if len(addressAndBalance) != 2 {
			app.logger.Error("Error read app states", "appstate", sbytes)
			return types.ResponseInitChain{}
		}
		addressS, balanceS := addressAndBalance[0], addressAndBalance[1]
		fmt.Println(addressS)
		fmt.Println(balanceS)
		//app.app.state.db.Set(prefixBalanceKey([]byte(addressS)), []byte(balanceS+":1"))
		//app.app.state.Size += 1
		//app.app.Commit()
	}

	app.ChainId = ankrcmm.ChainID(req.ChainId)

    app.app.SetChainID(req.ChainId)

	account.AccountManagerInstance().Init(app.app)

	return types.ResponseInitChain{}
}

// Track the block hash and header information
func (app *AnkrChainApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	app.logger.Info(fmt.Sprintf("AnkrChainApplication BeginBlock appHash=%X, height=%d", req.Header.AppHash, req.Header.Height))

	val.ValidatorManagerInstance().ValBeginBlock(req, app.app)
	app.latestHeight = req.Header.Height
	app.latestAPPHash = req.Header.AppHash
	return types.ResponseBeginBlock{}
}

// Update the validator set
func (app *AnkrChainApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	app.latestHeight = req.Height
	return types.ResponseEndBlock{ValidatorUpdates: val.ValidatorManagerInstance().ValUpdates()}
}