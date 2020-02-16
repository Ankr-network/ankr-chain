package common

type QueryResp struct {
	RespData   []byte  `json:"respdata"`
	ProofValue []byte  `json:"proofvalue"`
}

type NonceQueryReq struct {
	Address string  `json:"address"`
}

type NonceQueryResp struct {
	Nonce      uint64  `json:"nonce"`
}

type BalanceQueryReq struct {
	Address string  `json:"address"`
	Symbol  string  `json:"symbol"`
}

type BalanceQueryResp struct {
	Amount string      `json:"amount"`
}

type CertKeyQueryReq struct {
	DCName string  `json:"dcname"`
}

type CertKeyQueryResp struct {
	PEMBase64 string  `json:"pembase64"`
}

type MeteringQueryReq struct {
	DCName string  `json:"dcname"`
	NSName string  `json:"nsname"`
}

type MeteringQueryResp struct {
	Value string      `json:"value"`
}

type ContractQueryReq struct {
	Address string  `json:"address"`
}

type ContractQueryResp struct {
	Addr       string   `json:"addr"`
	Name       string   `json:"name"`
	Owner      string   `json:"owneraddr"`
	Codes      []byte   `json:"codes"`
	CodesDesc  string   `json:"codesdesc"`
}

type ValidatorQueryReq struct {
	ValAddr  string  `json:"valaddr"`
}

type ValidatorQueryResp struct {
	Name         string     `json:"name"`
	ValAddress   string     `json:"valaddress"`
	PubKey       ValPubKey  `json:"pubkey"`
	Power        int64      `json:"power"`
	StakeAddress string     `json:"stakeaddress"`
	StakeAmount  Amount     `json:"stakeamount"`
	ValidHeight  uint64     `json:"validheight"`
}

type AccountQueryReq struct {
	Addr string  `json:"addr"`
}

type AccountQueryResp struct {
	AccType    AccountType `json:"accounttype"`
	Nonce      uint64      `json:"nonce"`
	Address    string      `json:"address"`
	PubKey     string      `json:"pubkey"`
	Amounts    []Amount    `json:"asserts"`
}

type CurrencyQueryReq struct {
	Symbol string  `json:"symbol"`
}

type CurrencyQueryResp struct {
	Symbol       string `json:"symbol"`
	Decimal      int64  `json:"decimal"`
	TotalSupply  string `json:"totalsupply"`
}

type StatisticalInfoReq struct {
}

type StatisticalInfoResp struct {
	Addrs      []string  `json:"addrs"`
    TotalTx    int64     `json:"totaltx"`
}

type RoleQueryReq struct {
	Name string  `json:"name"`
}

type RoleQueryResp struct {
	Name         string    `json:"name"`
	RType        RoleType `json:"roletype"`
	PubKey       string   `json:"pubkey"`
	ContractAddr string   `json:"contractaddr"`
}