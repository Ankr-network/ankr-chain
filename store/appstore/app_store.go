package appstore

import (
	"github.com/Ankr-network/ankr-chain/store/appstore/kv"
	ankrtypes "github.com/Ankr-network/ankr-chain/types"
	"github.com/tendermint/tendermint/abci/types"
)

type AppStore interface {
	Info(req types.RequestInfo) (resInfo types.ResponseInfo)
	DeliverTx(tx []byte) types.ResponseDeliverTx
	CheckTx(tx []byte) types.ResponseCheckTx
	Commit() types.ResponseCommit
	Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery)
	SetOption(req types.RequestSetOption) types.ResponseSetOption
	Get(key []byte) []byte
	Set(key []byte, val []byte)
	Delete(key []byte)
	Has(key []byte) bool
	Validators(judgeValidatorTx ankrtypes.JudgeValidatorTx) (validators []types.ValidatorUpdate)
	TotalValidatorPowers(judgeValidatorTx ankrtypes.JudgeValidatorTx) int64
	Size() int64
	IncSize()
	Height() int64
	APPHash() []byte
}

func NewAppStore(dbDir string) AppStore {
	return kv.NewKVStoreApplication(dbDir)
}