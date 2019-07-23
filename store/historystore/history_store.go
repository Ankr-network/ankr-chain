package historystore

import (
	"github.com/Ankr-network/ankr-chain/store/historystore/db/mongodb"
	"github.com/Ankr-network/ankr-chain/store/historystore/db/mysql"
	"github.com/Ankr-network/ankr-chain/store/historystore/types"
	"github.com/tendermint/tendermint/libs/log"
)

type HistoryStorage interface {
	AddSendTx(tx *types.TransactionSendTx) error
	AddMetering(tx *types.TransactionMetering) error
	AddSetBalanceTx(tx *types.TransactionSetBalanceTx) error
	AddSetStakeTx(tx *types.TransactionSetStakeTx) error
	AddSetValidatorTx(tx *types.TransactionSetValidatorTx) error

	AddAccount(account *types.Account) error
	UpdateAccount(Address string, amount string) error
	GetAccount(Address string) (*types.Account, error)
}

func NewHistoryStorage(dbType string, dbHost string, dbName string,  logHis log.Logger) HistoryStorage {
	if dbType == "mongodb" {
		return mongodb.NewMongoDB(dbHost, dbName)
	}else if dbType == "mysql" {
		return mysql.NewMySql(dbHost, dbName, logHis)
	}

	return nil

}
