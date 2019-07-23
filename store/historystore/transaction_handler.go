package historystore

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/Ankr-network/ankr-chain/consensus"
	"github.com/Ankr-network/ankr-chain/store/historystore/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/types/time"
)

type  txHandler func(string, string, int64, uint32, []byte) error

type transactionHandler struct {
	txEventC       chan *tmtypes.EventDataTx
	txHandlerMap   map[string]txHandler
	txPrefixLenMap map[string]int
	txHisStore     HistoryStorage
	accountLocker  sync.RWMutex
	logHis         log.Logger
}

func newTransactionHandler(DBType string, DBHost string, DBName string, txEventC chan *tmtypes.EventDataTx, logHis log.Logger)*transactionHandler {

	tranHandler := new(transactionHandler)

	tranHandler.txHandlerMap = make(map[string]txHandler)
	tranHandler.txHandlerMap["Send"]              = tranHandler.handlerSendTx
	tranHandler.txHandlerMap["SetMetering"]      = tranHandler.handlerMetering
	tranHandler.txHandlerMap["SetBalance"]       = tranHandler.handlerSetBalance
	tranHandler.txHandlerMap["SetStake"]         = tranHandler.handlerSetStake
	tranHandler.txHandlerMap["UpdateValidator"] = tranHandler.handlerValidatorTx

	tranHandler.txPrefixLenMap = make(map[string]int)
	tranHandler.txPrefixLenMap["Send"]              = len("trx_send=")
	tranHandler.txPrefixLenMap["SetMetering"]      = len("set_mtr=")
	tranHandler.txPrefixLenMap["SetBalance"]       = len("set_bal=")
	tranHandler.txPrefixLenMap["SetStake"]         = len("set_stk=")
	tranHandler.txPrefixLenMap["UpdateValidator"] = len("val:")

	tranHandler.txHisStore = NewHistoryStorage(DBType, DBHost, DBName, logHis)
	tranHandler.txEventC   = txEventC
	tranHandler.logHis     = logHis

	if tranHandler.txHisStore != nil {
		acc, err := tranHandler.txHisStore.GetAccount(ankrchain.INIT_ADDRESS)
		if err != nil || acc == nil {
			tranHandler.txHisStore.AddAccount(&types.Account{ankrchain.INIT_ADDRESS, "1000000000000000000000000"})
		}
	}

	return tranHandler
}

func (th *transactionHandler) updateAccount(address string, amount *big.Int, isFrom bool) error {
	th.accountLocker.Lock()
	defer th.accountLocker.Unlock()

	acc, err := th.txHisStore.GetAccount(address)
	if err != nil {
		if isFrom {
			th.logHis.Error("not exist from address", "err", err, "fromaddress", address)
			return err
		}else {
			return th.txHisStore.AddAccount(&types.Account{address, amount.String()})
		}
	}
	if acc == nil {
		if isFrom {
			th.logHis.Error("invalid from address", "err", err, "fromaddress", address)
			return errors.New("invalid address")
		}else {
			return th.txHisStore.AddAccount(&types.Account{address, amount.String()})
		}
	}

	balBig, _ := new(big.Int).SetString(acc.Balance, 10)
	if isFrom {
		acc.Balance = new(big.Int).Sub(balBig, amount).String()
	}else {
		acc.Balance = new(big.Int).Add(balBig, amount).String()
	}

	return th.txHisStore.UpdateAccount(acc.Address, acc.Balance)
}

func (th *transactionHandler) setAccountBalance(address string, amount *big.Int) error {
	return th.updateAccount(address, amount,false)
}

func (th *transactionHandler) handlerSendTx(txHash string, txType string, height int64, index uint32, txBody []byte) error {
	txSendSegs := strings.Split(string(txBody), ":")
	if len(txSendSegs) < 6 {
		 th.logHis.Error("invalid tx send", "tx send params count", len(txSendSegs))
		 return  fmt.Errorf("invalid tx send: %v", txSendSegs)
	}

	fromAddress := txSendSegs[0]
	toAddress   := txSendSegs[1]
	amountS     := txSendSegs[2]

	amountInt, isSucess := new(big.Int).SetString(amountS, 10)
	if !isSucess {
		th.logHis.Error("invalid amoun", "amount", amountS)
		return errors.New("invalid amount")
	}

	if th.txHisStore == nil {
		th.logHis.Error("txHisStore is nil")
		return errors.New("txHisStore is nil")
	}

	err := th.updateAccount(fromAddress, amountInt, true)
	if err != nil {
		th.logHis.Error("updateAccount error", "err", err, "fromaddress", fromAddress, "amount", amountInt.String(), "isFrom", true)
		return err
	}
	err = th.updateAccount(toAddress, amountInt, false)
	if err != nil {
		th.logHis.Error("updateAccount error", "err", err, "toaddress", toAddress, "amount", amountInt.String(), "isFrom", false)
		return err
	}

	txHead := types.TransactionHead{TxHash: txHash, TxType: txType, Height: height, Index: index, Time: time.Now()}
    txSendTx := &types.TransactionSendTx{txHead, fromAddress, toAddress, amountS}

	th.txHisStore.AddSendTx(txSendTx)

	return nil
}

func (th *transactionHandler) handlerMetering(txHash string, txType string, height int64, index uint32, txBody []byte) error {
	txMeteringSegs := strings.SplitN(string(txBody), ":", 6)
	if len(txMeteringSegs) != 6 {
		th.logHis.Error("invalid tx metering", "tx metering params count", len(txMeteringSegs))
		return  fmt.Errorf("invalid tx Metering: %v", txMeteringSegs)
	}
	dcS    := txMeteringSegs[0]
	nsS    := txMeteringSegs[1]
	valueS := txMeteringSegs[5]

	if th.txHisStore == nil {
		th.logHis.Error("txHisStore is nil")
		return errors.New("txHisStore is nil")
	}

	txHead := types.TransactionHead{TxHash: txHash, TxType: txType, Height: height, Index: index, Time: time.Now()}
	txMetering := &types.TransactionMetering{txHead, dcS, nsS, valueS}

	th.txHisStore.AddMetering(txMetering)

	return nil
}

func (th *transactionHandler) handlerSetBalance(txHash string, txType string, height int64, index uint32, txBody []byte) error {
	txSetBalanceSegs := strings.Split(string(txBody), ":")
	if len(txSetBalanceSegs) != 5 {
		th.logHis.Error("invalid tx setbalance", "tx setbalance params count", len(txSetBalanceSegs))
		return  fmt.Errorf("invalid tx set balance: %v", txSetBalanceSegs)
	}
	addressS := txSetBalanceSegs[0]
	amountS  := txSetBalanceSegs[1]

	amountInt, isSucess := new(big.Int).SetString(amountS, 10)
	if !isSucess {
		th.logHis.Error("invalid amount", "amount", amountS)
		return errors.New("invalid amount")
	}

	err := th.setAccountBalance(addressS, amountInt)
	if err != nil {
		return err
	}

	txHead := types.TransactionHead{TxHash: txHash, TxType: txType, Height: height, Index: index, Time: time.Now()}
	txSetBalance := &types.TransactionSetBalanceTx{txHead, addressS, amountS}
	th.txHisStore.AddSetBalanceTx(txSetBalance)

	return nil
}

func (th *transactionHandler) handlerSetStake(txHash string, txType string, height int64, index uint32, txBody []byte) error {
	txSetStakeSegs := strings.Split(string(txBody), ":")
	if len(txSetStakeSegs) != 4 {
		th.logHis.Error("invalid tx setstake", "tx setstake params count", len(txSetStakeSegs))
		return  fmt.Errorf("invalid tx set stake: %v", txSetStakeSegs)
	}

	amountS := txSetStakeSegs[0]

	txHead := types.TransactionHead{TxHash: txHash, TxType: txType, Height: height, Index: index, Time: time.Now()}
	txSetStake := &types.TransactionSetStakeTx{txHead, amountS}

	th.txHisStore.AddSetStakeTx(txSetStake)

	return nil
}

func (th *transactionHandler) handlerValidatorTx(txHash string, txType string, height int64, index uint32, txBody []byte) error {
	txVaidatorSegs := strings.Split(string(txBody), "/")
	if len(txVaidatorSegs) != 5 {
		th.logHis.Error("invalid tx validator", "tx validator params count", len(txVaidatorSegs))
		return  fmt.Errorf("invalid tx validator: %v", txVaidatorSegs)
	}
	pubkeyS := txVaidatorSegs[0]
	powerS  := txVaidatorSegs[1]

	txHead := types.TransactionHead{TxHash: txHash, TxType: txType, Height: height, Index: index, Time: time.Now()}
	txSetValidator:= &types.TransactionSetValidatorTx{txHead,pubkeyS, powerS}

	th.txHisStore.AddSetValidatorTx(txSetValidator)

	return nil
}

func (th *transactionHandler) handler(txEvData *tmtypes.EventDataTx) {
	if txEvData == nil {
		th.logHis.Error("txEvData == nil")
		return
	}

	height := txEvData.Height
	index  := txEvData.Index
	for _, tagKV := range txEvData.TxResult.Result.Tags {
		tagK := string(tagKV.Key)
		if tagK == "app.type" {
			txTypeS := string(tagKV.Value)
            if tHanddler, ok := th.txHandlerMap[txTypeS]; ok {
            	prefixLen := th.txPrefixLenMap[txTypeS]
				tHanddler(fmt.Sprintf("%X", txEvData.Tx.Hash()), txTypeS, height, index, txEvData.Tx[prefixLen:])
			}
			return
		}
	}
}

func (th *transactionHandler) Start() {
	go func() {
		for {
			select {
			case txEvData := <-th.txEventC:
				th.handler(txEvData)
			}
		}
	}()

}


