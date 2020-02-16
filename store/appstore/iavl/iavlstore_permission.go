package iavl

import (
	"crypto/sha256"
	"errors"
	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/tendermint/iavl"
)

const (
	StoreRolePrefix        = "rolestore:"
	StoreBoundActionPrefix = "bacstore:"
)

func containRolePrefix(address string) string {
	return containPrefix(address, StoreRolePrefix)
}

func containBoundActionPrefix(address string) string {
	return containPrefix(address, StoreBoundActionPrefix)
}

func stripRoleKeyPrefix(key string) (string, error) {
	return stripKeyPrefix(key, StoreRolePrefix)
}

func boundInfoKey(contractAddr string, actionName string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(contractAddr))
	hasher.Write([]byte(actionName))

	return hasher.Sum(nil)[:16]
}

func (sp *IavlStoreApp) AddRole(rType ankrcmm.RoleType, name string, pubKey string, contractAddr string) {
	if !sp.iavlSM.IavlStore(IavlStorePermKey).Has([]byte(containRolePrefix(name))) {
		rInfo := &ankrcmm.RoleInfo{name, rType, pubKey, contractAddr}
		bytes := ankrcmm.EncodeRoleInfo(sp.cdc, rInfo)

		sp.iavlSM.IavlStore(IavlStorePermKey).Set([]byte(containRolePrefix(rInfo.Name)), bytes)
	}
}

func (sp *IavlStoreApp) LoadRole(name string, height int64, prove bool) (*ankrcmm.RoleInfo, string, *iavl.RangeProof, []byte, error) {
	if name == "" {
		return nil, "", nil, nil, errors.New("LoadRole, blank name")
	}

	rInfoBytes, proof, err := sp.iavlSM.IavlStore(IavlStorePermKey).GetWithVersionProve([]byte(containRolePrefix(name)), height, prove)
	if err != nil || len(rInfoBytes) == 0 {
		sp.storeLog.Error("can't get the role info", "name", name)
		return nil, containRolePrefix(name), nil, nil, err
	}

	rInfo := ankrcmm.DecodeRoleInfo(sp.cdc, rInfoBytes)

	return &rInfo, containRolePrefix(name), proof, rInfoBytes, nil
}

func (sp *IavlStoreApp) RoleQuery(name string, height int64, prove bool) (*ankrcmm.QueryResp, string, *iavl.RangeProof, error) {
	rInfo, rPrefix, rp, rInfoBytes, err := sp.LoadRole(name, height, prove)
	if err == nil && rInfo != nil  {
		rRespInfo := &ankrcmm.RoleQueryResp{
			rInfo.Name,
			rInfo.Type,
			rInfo.PubKey,
			rInfo.ContractAddr,
		}

		respData, err := sp.cdc.MarshalJSON(rRespInfo)
		if err != nil {
			return nil, containRolePrefix(name), rp, err
		}

		return &ankrcmm.QueryResp{respData, rInfoBytes}, rPrefix, rp, nil
	}

	return nil, rPrefix, rp, err
}

func (sp *IavlStoreApp) AddBoundAction(roleName string, contractAddr string, actionName string) {
	key := boundInfoKey(contractAddr, actionName)
	if !sp.iavlSM.IavlStore(IavlStorePermKey).Has(key) {

		rbaInfo     := &ankrcmm.RoleBoundActionInfo{roleName, contractAddr, actionName}
		rbaInfoList := &ankrcmm.RoleBoundActionInfoList{[]*ankrcmm.RoleBoundActionInfo{rbaInfo}}
		bytes       := ankrcmm.EncodeBoundActionInfoList(sp.cdc, rbaInfoList)

		sp.iavlSM.IavlStore(IavlStorePermKey).Set(key, bytes)
	}else {
		bytes, err  := sp.iavlSM.IavlStore(IavlStorePermKey).Get(key)
		if err == nil && bytes != nil{
			rbaInfoList := ankrcmm.DecodeBoundActionInfoList(sp.cdc, bytes)
			for _, rbaInfo := range rbaInfoList.RoleBounds {
				if rbaInfo != nil && rbaInfo.Name == roleName {
					sp.storeLog.Debug("Role bound info existed", "role", roleName, "contractAddr", contractAddr, "actionName", actionName)
					return
				}
			}
			rbaInfo := &ankrcmm.RoleBoundActionInfo{roleName, contractAddr, actionName}
			rbaInfoList.RoleBounds = append(rbaInfoList.RoleBounds, rbaInfo)

			bytes := ankrcmm.EncodeBoundActionInfoList(sp.cdc, &rbaInfoList)

			sp.iavlSM.IavlStore(IavlStorePermKey).Set(key, bytes)
		}else {
			sp.storeLog.Error("can't load role bound info", "contractAddr", contractAddr, "actionName", actionName)
		}
	}
}

func (sp *IavlStoreApp) LoadBoundAction(contractAddr string, actionName string) ankrcmm.RoleBoundActionInfoList {
	key := boundInfoKey(contractAddr, actionName)
	if !sp.iavlSM.IavlStore(IavlStorePermKey).Has(key) {
		sp.storeLog.Error("can't load role bound info", "contractAddr", contractAddr, "actionName", actionName)
		return ankrcmm.RoleBoundActionInfoList{}
	}

	bytes, err := sp.iavlSM.IavlStore(IavlStorePermKey).Get(key)
	if err == nil && bytes != nil {
		rbaInfoList := ankrcmm.DecodeBoundActionInfoList(sp.cdc, bytes)
		return rbaInfoList
	}

	return ankrcmm.RoleBoundActionInfoList{}
}


