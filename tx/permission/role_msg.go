package permission

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/common/code"
	ankrcrypto "github.com/Ankr-network/ankr-chain/crypto"
	"github.com/Ankr-network/ankr-chain/store/appstore"
	"github.com/Ankr-network/ankr-chain/tx"
	txcmm "github.com/Ankr-network/ankr-chain/tx/common"
	"github.com/Ankr-network/wagon/exec/gas"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type AddRoleMsg struct {
	FromAddr     string            `json:"fromaddr"`
	Name         string            `json:"name"`
	RoleType     ankrcmm.RoleType  `json:"roletype"`
	tmcrypto.    PubKey            `json:"pubkey"`
	ContractAddr string            `json:"contractaddr"`
}

func (ar *AddRoleMsg) SignerAddr() []string {
	return []string {ar.FromAddr}
}

func (ar *AddRoleMsg) GasWanted() int64 {
	return 0
}

func (ar *AddRoleMsg) Type() string {
	return txcmm.TxMsgTypeAddRole
}

func (ar *AddRoleMsg) Bytes(txSerializer tx.TxSerializer) []byte {
	bytes, _ :=  txSerializer.MarshalJSON(ar)
	return bytes
}

func (ar *AddRoleMsg) SetSecretKey(sk ankrcrypto.SecretKey) {

}

func (ar *AddRoleMsg) SecretKey() ankrcrypto.SecretKey {
	return &ankrcrypto.SecretKeyEd25519{}
}

func (ar *AddRoleMsg) PermitKey(store appstore.AppStore, pubKey []byte) bool {
	return true
}

func (ar *AddRoleMsg) SenderAddr() string {
	return ar.FromAddr
}

func (ar *AddRoleMsg) ProcessTx(context tx.ContextTx, metric gas.GasMetric, flag tx.TxExeFlag) (uint32, string, []cmn.KVPair){
	if len(ar.FromAddr) != ankrcmm.KeyAddressLen {
		return code.CodeTypeInvalidAddress, fmt.Sprintf("AddRoleMsg ProcessTx, unexpected from address. Got %s, addr len=%d", ar.FromAddr, len(ar.FromAddr)), nil
	}

	regx := regexp.MustCompile(`[^a-zA-Z0-9]`)
	if !regx.MatchString(ar.Name) {
		return code.CodeTypeRoleNameInvalid, fmt.Sprintf("AddRoleMsg ProcessTx, invalid role name:%s", ar.Name), nil
	}

	if ar.RoleType == ankrcmm.RoleContract {
		if ar.ContractAddr == "" {
			return code.CodeTypeRoleContractAddrBlank, fmt.Sprintf("AddRoleMsg ProcessTx, blank contract address for contract role name:%s", ar.Name), nil
		}else {
			cInfo, _, _, _, err := context.AppStore().LoadContract(ar.ContractAddr, 0, false)
			if err != nil || cInfo == nil {
				errStr := ""
				if err != nil {
					errStr = err.Error()
				}
				return code.CodeTypeContractCantFound, fmt.Sprintf("AddRoleMsg ProcessTx, can't load contract %s, %s", ar.ContractAddr, errStr), nil
			}

			if ar.FromAddr != cInfo.Owner {
				return code.CodeTypeRoleInvalidAccount, fmt.Sprintf("AddRoleMsg ProcessTx, now contract %s owner address, expected %s, got %s", ar.ContractAddr, cInfo.Owner, ar.FromAddr), nil
			}
		}
	}else {
		//TBD
		return code.CodeTypeRoleUnSupportedType, fmt.Sprintf("AddRoleMsg ProcessTx, not support role type: %d", ar.RoleType), nil
	}


	if flag == tx.TxExeFlag_OnlyCheck || flag == tx.TxExeFlag_PreRun{
		return code.CodeTypeOK, "", nil
	}

	context.AppStore().AddRole(ar.RoleType, ar.Name, string(ar.PubKey.Bytes()), ar.ContractAddr)

	context.AppStore().IncNonce(ar.FromAddr)

	tvalue := time.Now().UnixNano()

	tags := []cmn.KVPair{
		{Key: []byte("app.timestamp"), Value: []byte(strconv.FormatInt(tvalue, 10))},
		{Key: []byte("app.type"), Value: []byte(txcmm.TxMsgTypeAddRole)},
	}

	return code.CodeTypeOK, "", tags
}
