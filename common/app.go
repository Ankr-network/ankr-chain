package common

const (
	APPName = "AnkrApp"

	KeyAddressLen = 46
	ValidatorSetChangePrefix string = "val:"
	AccountBlancePrefix string = "bal:"
	AccountStakePrefix string = "stk"
	CertPrefix string = "crt:"
	MeteringPrefix string = "mtr:"
	AllAccountsPrefix string = "all_accounts"
	AllCrtsPrefix string = "all_crts"

	SetMeteringPrefix string = "set_mtr="
	TrxSendPrefix string = "trx_send="
	SetBalancePrefix string = "set_bal="
	SetOpPrefix string = "set_op="
	SetStakePrefix string = "set_stk="
	SetCertPrefix string = "set_crt="
	RemoveCertPrefix string = "rmv_crt="

	SET_CRT_NONCE string = "set_crt_nonce"
	RMV_CRT_NONCE string = "rmv_crt_nonce"
	SET_OP_NONCE string = "admin_nonce"
	SET_VAL_NONCE string = "val_nonce"
	ADMIN_OP_VAL_PUBKEY_NAME string = "admin_op_val_pubkey"
	ADMIN_OP_FUND_PUBKEY_NAME string = "admin_op_fund_pubkey"
	ADMIN_OP_METERING_PUBKEY_NAME string = "admin_op_metering_pubkey"
)

type RunMode int
const (
	_ RunMode = iota
	RunModeTesting
	RunModeProd
)

var RM = RunModeProd

type ChainID string

type Address []byte

// CommitID contains the tree version number and its merkle root.
type CommitID struct {
	Version int64
	Hash    []byte
}

func PrefixBalanceKey(key []byte) []byte {
	return append([]byte(AccountBlancePrefix), key...)
}


