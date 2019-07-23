package historystore


import (
	"context"

	"github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
)

const (
	SUBSCRIBER = "HistoryStoraService"
	TxEventChanMax = 100
)

type HistoryStorageService struct {
	common.BaseService
	eventBus  *types.EventBus
	txEventC  chan *types.EventDataTx
	txHandler *transactionHandler
	logHis    log.Logger
}

func NewHistoryStorageService(DBType string, DBHost string, DBName string, eventBus *types.EventBus, logHis log.Logger) *HistoryStorageService {
	hss := &HistoryStorageService{eventBus: eventBus}

	hss.logHis = logHis

	hss.BaseService = *common.NewBaseService(nil, SUBSCRIBER, hss)
	hss.txEventC = make(chan *types.EventDataTx, TxEventChanMax)

	hss.txHandler = newTransactionHandler(DBType, DBHost, DBName, hss.txEventC, logHis)
	hss.txHandler.Start()

	hss.logHis.Info("HistoryDB service start sucessfully", "dbType", DBType, "dbHost", DBHost, "dbName", DBName)

	return hss
}

func (hss *HistoryStorageService) OnStart() error {
	blockHeadersSub, err := hss.eventBus.SubscribeUnbuffered(context.Background(), SUBSCRIBER, types.EventQueryNewBlockHeader)
	if err != nil {
		hss.logHis.Error("EventBus SubscribeUnbuffered  NewBlockHeader Failed", "error", err)
		return err
	}

	txsSub, err := hss.eventBus.SubscribeUnbuffered(context.Background(), SUBSCRIBER, types.EventQueryTx)
	if err != nil {
		hss.logHis.Error("EventBus SubscribeUnbuffered Tx Failed", "error", err)
		return err
	}

	go func() {
		for {
			msg := <-blockHeadersSub.Out()
			header := msg.Data().(types.EventDataNewBlockHeader).Header
			for i := int64(0); i < header.NumTxs; i++ {
				msg2 := <-txsSub.Out()
				txEv := msg2.Data().(types.EventDataTx)
                hss.txEventC <- &txEv
			}
		}
	}()

	return nil
}

func (hss *HistoryStorageService) OnStop() {
	if hss.eventBus.IsRunning() {
		_ = hss.eventBus.UnsubscribeAll(context.Background(), SUBSCRIBER)
	}

	hss.logHis.Info("EventBus stop all subs")
}




