package validator

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/Ankr-network/ankr-chain/common/code"
	"github.com/Ankr-network/ankr-chain/store/appstore"
	ankrtypes "github.com/Ankr-network/ankr-chain/types"
	"github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

var (
	onceVM     sync.Once
	instanceVM *ValidatorManager
)

type ValidatorManager struct {
	valUpdates []types.ValidatorUpdate
}

func (vm *ValidatorManager) InitValidator(appStore appstore.AppStore) {
	appStore.Set([]byte(ankrtypes.SET_VAL_NONCE), []byte("0"))
	appStore.Set(PrefixStakeKey([]byte("")), []byte("0:1"))
	/*value := []byte("")
	value = appStore.Get([]byte(ankrtypes.AccountStakePrefix))
	if value == nil || string(value) == "" {
		appStore.Set(PrefixStakeKey([]byte("")), []byte("0:1"))
	}*/
}

// add, update, or remove a validator
func (v *ValidatorManager) UpdateValidator(valUP types.ValidatorUpdate, appStore appstore.AppStore) (uint32, string,  []cmn.KVPair) {
	key := []byte("val:" + string(valUP.PubKey.Data))
	if valUP.Power == 0 {
		// remove validator
		if !appStore.Has(key) {
			return code.CodeTypeUnauthorized, fmt.Sprintf("Cannot remove non-existent validator %X", key), nil
		}
		appStore.Delete(key)
	} else {
		// add or update validator
		value := bytes.NewBuffer(make([]byte, 0))
		if err := types.WriteMessage(&valUP, value); err != nil {
			return code.CodeTypeEncodingError, fmt.Sprintf("Error encoding validator: %v", err), nil
		}
		appStore.Set(key, value.Bytes())
	}

	// we only update the changes array if we successfully updated the tree
	ValidatorManagerInstance().valUpdates = append(ValidatorManagerInstance().valUpdates, valUP)

	tags := []cmn.KVPair{
		{Key: []byte("app.type"), Value: []byte("UpdateValidator")},
	}

	return code.CodeTypeOK, "", tags
}

func (vm *ValidatorManager) Reset() {
	vm.valUpdates = make([]types.ValidatorUpdate, 0)
}

func (vm *ValidatorManager) ValUpdates() []types.ValidatorUpdate {
	return vm.valUpdates
}

func (vm *ValidatorManager) Validators(appStore appstore.AppStore) (validators []types.ValidatorUpdate) {
	itr := appStore.DB().Iterator(nil, nil)
	for ; itr.Valid(); itr.Next() {
		if isValidatorTx(itr.Key()) {
			validator := new(types.ValidatorUpdate)
			err := types.ReadMessage(bytes.NewBuffer(itr.Value()), validator)
			if err != nil {
				panic(err)
			}
			validators = append(validators, *validator)
		}
	}
	return
}

func (vm *ValidatorManager) TotalValidatorPowers(appStore appstore.AppStore) int64 {
	var totalValPowers int64 = 0
	it := appStore.DB().Iterator(nil, nil)
	if it != nil && it.Valid(){
		it.Next()
		for it.Valid() {
			if isValidatorTx(it.Key()) {
				validator := new(types.ValidatorUpdate)
				err := types.ReadMessage(bytes.NewBuffer(it.Value()), validator)
				if err != nil {
					panic(err)
				}

				totalValPowers += validator.Power
				fmt.Printf("validator = %v\n", validator)
			}
			it.Next()
		}
	}
	it.Close()

	return  totalValPowers
}

func ValidatorManagerInstance() *ValidatorManager {
	onceVM.Do(func(){
		instanceVM = &ValidatorManager{make([]types.ValidatorUpdate, 0)}
	})

	return instanceVM
}

