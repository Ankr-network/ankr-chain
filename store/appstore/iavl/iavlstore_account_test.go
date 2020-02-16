package iavl

import (
	"math/big"
	"testing"

	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/stretchr/testify/assert"
)

func TestIavlStoreAccountCount(t *testing.T) {
	storeApp := NewMockIavlStoreApp()

	var accInfo1 ankrcmm.AccountInfo
	accInfo1.Nonce = 1
	accInfo1.Address = "5AEBA6EB8BC51DA277CCF1EF229F0C05D9535FA36CC872"
	accInfo1.PubKey = "430EF1CCC9C24958779C84E6EDE748F8B743E60475D5098837B8A4971C3468FA"
	accInfo1.Amounts = []ankrcmm.Amount{{ankrcmm.Currency{Symbol: "ANKR", Decimal: 18}, new(big.Int).SetUint64(100).Bytes()}}

	var accInfo2 ankrcmm.AccountInfo
	accInfo2.Nonce = 1
	accInfo2.Address = "8AEBA6EB8BC51DA277CCF1EF229F0C05D9535FA36CC873"
	accInfo2.PubKey = "930EF1CCC9C24958779C84E6EDE748F8B743E60475D5098837B8A4971C3468FB"
	accInfo2.Amounts = []ankrcmm.Amount{{ankrcmm.Currency{Symbol: "ANKR", Decimal: 18}, new(big.Int).SetUint64(100).Bytes()}}

	storeApp.addAccountInfo(&accInfo1)
	storeApp.addAccountInfo(&accInfo2)

	_, accCnt := storeApp.AccountList(0)

	assert.Equal(t, accCnt, uint64(2))
}

func TestBalance(t *testing.T) {
	storeApp := NewMockIavlStoreApp()

	var accInfo1 ankrcmm.AccountInfo
	accInfo1.Nonce = 1
	accInfo1.Address = "5AEBA6EB8BC51DA277CCF1EF229F0C05D9535FA36CC872"
	accInfo1.PubKey  = "430EF1CCC9C24958779C84E6EDE748F8B743E60475D5098837B8A4971C3468FA"
	accInfo1.Amounts = []ankrcmm.Amount{{ankrcmm.Currency{Symbol: "ANKR", Decimal: 18}, new(big.Int).SetUint64(100).Bytes()}}

	storeApp.addAccountInfo(&accInfo1)

	storeApp.SetBalance("5AEBA6EB8BC51DA277CCF1EF229F0C05D9535FA36CC872", ankrcmm.Amount{ankrcmm.Currency{Symbol: "ANKR", Decimal: 18}, new(big.Int).SetUint64(1000).Bytes()})

	bal, _, _, _, err := storeApp.Balance("5AEBA6EB8BC51DA277CCF1EF229F0C05D9535FA36CC872", "ANKR", 0, false)
	assert.Equal(t, err, nil)
	assert.Equal(t, bal.String(), "1000")
}
