package context

import (
	"context"
	"math/big"

	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/store/appstore"
	"github.com/tendermint/iavl"
)

var bcContext ContextAKVM

type ContextAKVM interface {
	CreateCurrency(symbol string, currency *ankrcmm.CurrencyInfo) error
	CurrencyInfo(symbol string, height int64, prove bool) (*ankrcmm.CurrencyInfo, string, *iavl.RangeProof, []byte, error)
	SpendGas(gas *big.Int) bool
	SenderAddr() string
	SetBalance(address string, amount ankrcmm.Amount)
	Balance(address string, symbol string, height int64, prove bool) (*big.Int, string, *iavl.RangeProof, []byte, error)
	SetAllowance(addrSender string, addrSpender string, amount ankrcmm.Amount)
	Allowance(addrSender string, addrSpender string, symbol string) (*big.Int, error)
	OwnerAddr() string
	ContractAddr() string
	BuildCurrencyCAddrMap(symbol string, cAddr string) error
	LoadContract(cAddr string, height int64, prove bool) (*ankrcmm.ContractInfo, string, *iavl.RangeProof, []byte, error)
	IsContractNormal(cAddr string) bool
	UpdateContractState(cAddr string, state ankrcmm.ContractState) error
	ChangeContractOwner(cAddr string, ownerAddr string) error
	AddContractRelatedObject(cAddr string, key string, jsonObject string) error
	LoadContractRelatedObject(cAddr string, key string)(jsonObject string, err error)
	Height() int64
	Publish(ctx context.Context, msg interface{}) error
	PublishWithTags(ctx context.Context, msg interface{}, tags map[string]string) error
}

func SetBCContext(context ContextAKVM) {
	bcContext = context
}

func GetBCContext() ContextAKVM {
	return bcContext
}

type ContractStoreAKVM interface {
	BuildCurrencyCAddrMap(symbol string, cAddr string) error
	IsContractNormal(cAddr string) bool
	LoadContract(cAddr string, height int64, prove bool) (*ankrcmm.ContractInfo, string, *iavl.RangeProof, []byte, error)
	UpdateContractState(cAddr string, state ankrcmm.ContractState) error
	ChangeContractOwner(cAddr string, ownerAddr string) error
	AddContractRelatedObject(cAddr string, key string, jsonObject string) error
	LoadContractRelatedObject(cAddr string, key string)(jsonObject string, err error)
}

type ContextAKVMImpl struct {
	ContextContract
	ContractStoreAKVM
	appstore.BCStore
}

func CreateContextAKVM(context ContextContract, appStore appstore.AppStore) ContextAKVM {
	contAKVM :=  &ContextAKVMImpl{context,appStore, appStore}
	bcContext = contAKVM

	return contAKVM
}

