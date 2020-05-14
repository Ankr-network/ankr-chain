package iavl

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/iavl"
	"testing"

	"github.com/stretchr/testify/assert"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

var cdc = amino.NewCodec()

func TestIavlStoreVersion(t *testing.T) {
	db := dbm.NewMemDB()
	storeLog := log.NewNopLogger()

	iavlStore := NewIavlStore(db, 100, 10, storeLog)

	iavlStore.Set([]byte("testkey1"), []byte("testvalue1"))
	iavlStore.Commit()
	t.Logf("version=%d", iavlStore.tree.Version())

	val1, err := iavlStore.Get([]byte("testkey1"))
	assert.Equal(t, nil, err)
	assert.Equal(t, "testvalue1", string(val1))

	iavlStore.Set([]byte("testkey1"), []byte("testvalue2"))
	t.Logf("version=%d", iavlStore.tree.Version())

	val1, err = iavlStore.Get([]byte("testkey1"))
	assert.Equal(t, nil, err)
	assert.Equal(t, "testvalue2", string(val1))

	iavlStore.Rollback()
	t.Logf("version=%d", iavlStore.tree.Version())

	val1, err = iavlStore.Get([]byte("testkey1"))
	assert.Equal(t, nil, err)
	assert.Equal(t, "testvalue1", string(val1))
}

func TestIavlStoreVersionCommitRollback(t *testing.T) {
	db := dbm.NewMemDB()
	storeLog := log.NewNopLogger()

	iavlStore := NewIavlStore(db, 100, 10, storeLog)

	iavlStore.Set([]byte("testkey1"), []byte("testvalue1"))
	iavlStore.Commit()
	t.Logf("version=%d", iavlStore.tree.Version())

	val1, err := iavlStore.Get([]byte("testkey1"))
	assert.Equal(t, nil, err)
	assert.Equal(t, "testvalue1", string(val1))

	iavlStore.Set([]byte("testkey1"), []byte("testvalue2"))
	iavlStore.Commit()
	t.Logf("version=%d", iavlStore.tree.Version())

	val1, err = iavlStore.Get([]byte("testkey1"))
	assert.Equal(t, nil, err)
	assert.Equal(t, "testvalue2", string(val1))

	iavlStore.Rollback()
	t.Logf("version=%d", iavlStore.tree.Version())

	val1, err = iavlStore.Get([]byte("testkey1"))
	assert.Equal(t, nil, err)
	assert.Equal(t, "testvalue2", string(val1))
}

func TestRollback(t *testing.T) {
	require := require.New(t)

	tree := iavl.NewMutableTree(dbm.NewMemDB(), 0)

	tree.Set([]byte("k"), []byte("v"))
	tree.SaveVersion()

	t.Logf("version=%d", tree.Version())

	tree.Set([]byte("r"), []byte("v"))
	tree.Set([]byte("s"), []byte("v"))

	tree.Rollback()

	t.Logf("version=%d", tree.Version())

	tree.Set([]byte("t"), []byte("v"))

	tree.SaveVersion()

	t.Logf("version=%d", tree.Version())

	require.Equal(int64(2), tree.Size())

	_, val := tree.Get([]byte("r"))
	require.Nil(val)

	_, val = tree.Get([]byte("s"))
	require.Nil(val)

	_, val = tree.Get([]byte("t"))
	require.Equal([]byte("v"), val)
}

func TestIavlStoreMultiOpEncode(t *testing.T) {
	op := IavlStoreMultiOp {
		key: []byte{0x01, 0x08},
		Proof: &IavlStoreMultiProof{
			commitInfo{
				Version: 81,
			    AppHash:[]byte{0x18, 0x37},
				Commits: []storeCommitID{{Name: "test"}},
			},
	   },
	}

	dataBytes := cdc.MustMarshalBinaryLengthPrefixed(op)
	t.Logf("dataBytes=%v", dataBytes)

	var op1 IavlStoreMultiOp
	cdc.UnmarshalBinaryLengthPrefixed(dataBytes, &op1)

	t.Logf("op1=%v", op1)
}

func TestRollBack(t *testing.T)  {
	storeAPP := NewMockIavlStoreApp()

	storeAPP.IncTotalTx()
	storeAPP.IncTotalTx()

	storeAPP.Commit()

	storeAPP.IncTotalTx()
	storeAPP.Rollback()

	totalTx1, _ := storeAPP.totalTxLatestWithoutlock()
	totalTx2, _, _, _, _ := storeAPP.totalTxWithoutlock(0, false)
	fmt.Printf("totalTx1=%d, totalTx2=%d\n", totalTx1, totalTx2)
}

func TestQueryInfo(t *testing.T) {
	storeLog := log.NewNopLogger()


	dbApp := NewIavlStoreApp("F:\\dccntendermint\\mainnode\\arthurlog\\data", storeLog)
	countTx, _, _, _, _ := dbApp.TotalTx(0, false)

	fmt.Printf("countTx=%d\n", countTx)
}

