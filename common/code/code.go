package code

import (
	"github.com/tendermint/tendermint/abci/types"
)

const (
	CodeTypeOK                       uint32 = types.CodeTypeOK
	CodeTypeEncodingError            uint32 = 1
	CodeTypeDecodingError            uint32 = 2
	CodeTypeBadNonce                 uint32 = 3
	CodeTypeUnauthorized             uint32 = 4
	CodeTypeP2PFilterPathError       uint32 = 5
	CodeTypeP2PFilterNotPass         uint32 = 6
	CodeQueryDataLenZero             uint32 = 7
	CodeTypeUnknownError             uint32 = 8
	CodeTypeMismatchChainID          uint32 = 9
	CodeTypeGetStoreNonceError       uint32 = 10
	CodeTypeInvalidAddress           uint32 = 11
	CodeTypeVerifySignaError         uint32 = 12
	CodeTypeLoadBalError             uint32 = 13
	CodeTypeBalNotEnough             uint32 = 14
	CodeTypeGasNotEnough             uint32 = 15
	CodeTypeFeeNotEnough             uint32 = 16
	CodeTypeGasPriceIrregular        uint32 = 17
	CodeTypeGasChargeError           uint32 = 18
	CodeTypeLoadContractErr          uint32 = 19
	CodeTypeCallContractErr          uint32 = 20
	CodeTypeInvalidStakeCurrency     uint32 = 21
	CodeTypeInvalidValidatorPubKey   uint32 = 22
	CodeTypeNotPermitPubKey          uint32 = 23
	CodeTypeContractInvalidCodeSize  uint32 = 24
	CodeTypeContractAddrTakenUp      uint32 = 25
	CodeTypeContractInvalidAddr      uint32 = 26
	CodeTypeContractCantFound        uint32 = 27
	CodeTypeQueryInvalidStoreName    uint32 = 28
	CodeTypeQueryInvalidQueryReqData uint32 = 29
	CodeTypeNoV0TxHandler            uint32 = 30
	CodeTypeInvalidFromAddr          uint32 = 31
	CodeTypeCheckTxError             uint32 = 32
	CodeTypeDeliverTxError           uint32 = 33
	CodeTypeMismatchTxVersion        uint32 = 34
	CodeTypeDuplicateTxError         uint32 = 35
	CodeTypeRoleNameInvalid          uint32 = 36
	CodeTypeRoleContractAddrBlank    uint32 = 37
	CodeTypeRoleInvalidAccount       uint32 = 38
	CodeTypeRoleUnSupportedType      uint32 = 39
	CodeTypeRoleExisted              uint32 = 40
	CodeTypeRoleNotExisted           uint32 = 41
	CodeTypeRoleNotMismatch          uint32 = 42
)
