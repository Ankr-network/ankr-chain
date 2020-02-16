package common

import (
	"github.com/tendermint/go-amino"
)

type PayRepositoryState int
const (
	_ PayRepositoryState = iota
   PayRepositoryNormal
   PayRepositoryFrozen
)

type PayRepository struct {
	AccAddress string             `json:"accaddress"`
	Amounts    []Amount           `json:"amounts"`
	State      PayRepositoryState `json:"state"`
}

type PayRecord struct {
	FromAddress string   `json:"fromaddress"`
	ToAddress   string   `json:"toaddress"`
	Amounts     []Amount `json:"amounts"`
}

func RegisterRSPayCodec(cdc *amino.Codec) {
	cdc.RegisterConcrete(PayRepository{}, "ankr-chain/payRepository", nil)
	cdc.RegisterConcrete(PayRecord{}, "ankr-chain/payRecord", nil)
}

func EncodePayRepositoryInfo(cdc *amino.Codec, pRInfo *PayRepository) []byte {
	return cdc.MustMarshalBinaryBare(pRInfo)
}

func DecodePayRepositoryInfo(cdc *amino.Codec, bytes []byte) (pRInfo PayRepository) {
	cdc.MustUnmarshalBinaryBare(bytes, &pRInfo)
	return
}

func EncodePayRecordInfo(cdc *amino.Codec, pRInfo *PayRecord) []byte {
	return cdc.MustMarshalBinaryBare(pRInfo)
}

func DecodePayRecordInfo(cdc *amino.Codec, bytes []byte) (pRInfo PayRecord) {
	cdc.MustUnmarshalBinaryBare(bytes, &pRInfo)
	return
}
