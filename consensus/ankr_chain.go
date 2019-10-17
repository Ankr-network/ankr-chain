package ankrchain

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/Ankr-network/ankr-chain/account"
	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/common/code"
	"github.com/Ankr-network/ankr-chain/contract"
	"github.com/Ankr-network/ankr-chain/router"
	"github.com/Ankr-network/ankr-chain/store/appstore"
	"github.com/Ankr-network/ankr-chain/store/appstore/iavl"
	"github.com/Ankr-network/ankr-chain/tx"
	_ "github.com/Ankr-network/ankr-chain/tx/metering"
	"github.com/Ankr-network/ankr-chain/tx/serializer"
	_ "github.com/Ankr-network/ankr-chain/tx/token"
	val "github.com/Ankr-network/ankr-chain/tx/validator"
	akver "github.com/Ankr-network/ankr-chain/version"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmpubsub "github.com/tendermint/tendermint/libs/pubsub"
	tmCoreTypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"
)

var _ types.Application = (*AnkrChainApplication)(nil)

type AnkrChainApplication struct {
	ChainId      ankrcmm.ChainID
	APPName      string
	app          appstore.AppStore
	txSerializer tx.TxSerializer
	contract     contract.Contract
	pubsubServer *tmpubsub.Server
	logger       log.Logger
	minGasPrice  ankrcmm.Amount
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

	//router.MsgRouterInstance().SetLogger(l.With("module", "AnkrChainRouter"))

	chainID := appStore.ChainID()

	return &AnkrChainApplication{
		ChainId:      ankrcmm.ChainID(chainID),
		APPName:      appName,
		app:          appStore,
		txSerializer: serializer.NewTxSerializerCDC(),
		contract:     contract.NewContract(appStore, l.With("module", "contract")),
		logger:       l,
		minGasPrice:  ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(0).Bytes()},
	}
}

func NewMockAnkrChainApplication(appName string, l log.Logger) *AnkrChainApplication {
	appStore := NewMockAppStore()

	account.AccountManagerInstance().Init(appStore)

	return &AnkrChainApplication{
		APPName:      appName,
		app:          appStore,
		txSerializer: serializer.NewTxSerializerCDC(),
		contract:     contract.NewContract(appStore, l.With("module", "contract")),
		logger:       l,
		minGasPrice:  ankrcmm.Amount{ankrcmm.Currency{"ANKR", 18}, new(big.Int).SetUint64(0).Bytes()},
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

func (app *AnkrChainApplication) PubSubServer() *tmpubsub.Server {
	return app.pubsubServer
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

func (app *AnkrChainApplication) dispossTx(tx []byte) (*tx.TxMsg, uint32, string) {
	txMsg, err := app.txSerializer.Deserialize(tx)
	if err != nil {
		if app.logger != nil {
			app.logger.Error("can't deserialize tx", "err", err)
		}
		return nil, code.CodeTypeDecodingError, fmt.Sprintf("can't deserialize tx: tx=%v, err=%s", tx, err.Error())
	} else {
		if txMsg.ChID != app.ChainId {
			return nil, code.CodeTypeMismatchChainID, fmt.Sprintf("can't mistach the chain id, txChainID=%s, appChainID=%s", txMsg.ChID, app.ChainId)
		}
	}

	return txMsg, code.CodeTypeOK, ""
}

// tx is either "val:pubkey/power" or "key=value" or just arbitrary bytes
func (app *AnkrChainApplication) DeliverTx(tx []byte) types.ResponseDeliverTx {
	txMsg, codeVal, logStr := app.dispossTx(tx)
	if codeVal == code.CodeTypeOK {
		return txMsg.DeliverTx(app)
	}

	return types.ResponseDeliverTx{ Code: codeVal, Log: logStr}
}

func (app *AnkrChainApplication) CheckTx(tx []byte) types.ResponseCheckTx {
	txMsg, codeVal, logStr := app.dispossTx(tx)
	if codeVal == code.CodeTypeOK {
		return txMsg.CheckTx(app)
	}

	return types.ResponseCheckTx{ Code: codeVal, Log: logStr}
}

// Commit will panic if InitChain was not called
func (app *AnkrChainApplication) Commit() types.ResponseCommit {
	return app.app.Commit()
}

func (app *AnkrChainApplication) Query(reqQuery types.RequestQuery) types.ResponseQuery {
	qHandler, subPath := router.QueryRouterInstance().QueryHandler(reqQuery.Path)
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
	val.ValidatorManagerInstance().ValBeginBlock(req, app.app)
	return types.ResponseBeginBlock{}
}

// Update the validator set
func (app *AnkrChainApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	return types.ResponseEndBlock{ValidatorUpdates: val.ValidatorManagerInstance().ValUpdates()}
}