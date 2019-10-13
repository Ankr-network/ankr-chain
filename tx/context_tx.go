package tx

import (
	"github.com/Ankr-network/ankr-chain/account"
	"github.com/Ankr-network/ankr-chain/contract"
	"github.com/Ankr-network/ankr-chain/store/appstore"
	"github.com/tendermint/tendermint/libs/log"
)

type ContextTx interface {
	MinGasPrice() account.Amount
	AppStore() appstore.AppStore
	TxSerializer() TxSerializer
	Contract() contract.Contract
	Logger() log.Logger
}