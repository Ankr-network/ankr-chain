package common

import (
	"github.com/tendermint/go-amino"
)

type RoleType int
const (
	_ RoleType = iota
	RoleGeneral
	RoleContract
)

type RoleInfo struct {
	Name         string
	Type         RoleType
	PubKey       string
	ContractAddr string
}

type RoleBoundActionInfo struct {
	Name         string
	ContractAddr string
	ActionName   string
}

type RoleBoundActionInfoList struct {
	RoleBounds []*RoleBoundActionInfo
}

func RegisterPermissionCodec(cdc *amino.Codec) {
	cdc.RegisterConcrete(RoleInfo{}, "ankr-chain/role", nil)
	cdc.RegisterConcrete(RoleInfo{}, "ankr-chain/roleBoundAction", nil)
	cdc.RegisterConcrete(RoleBoundActionInfoList{}, "ankr-chain/roleBoundActionInfoList", nil)
}

func EncodeRoleInfo(cdc *amino.Codec, rInfo *RoleInfo) []byte {
	return cdc.MustMarshalBinaryBare(rInfo)
}

func DecodeRoleInfo(cdc *amino.Codec, bytes []byte) (rInfo RoleInfo) {
	cdc.MustUnmarshalBinaryBare(bytes, &rInfo)
	return
}

func EncodeBoundActionInfoList(cdc *amino.Codec, rbaInfoList *RoleBoundActionInfoList) []byte {
	return cdc.MustMarshalBinaryBare(rbaInfoList)
}

func DecodeBoundActionInfoList(cdc *amino.Codec, bytes []byte) (rbaInfoList RoleBoundActionInfoList) {
	cdc.MustUnmarshalBinaryBare(bytes, &rbaInfoList)
	return
}
