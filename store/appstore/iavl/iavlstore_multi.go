package iavl

import (
	"encoding/binary"
	"fmt"
	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/tendermint/go-amino"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	IavlStoreAccountKey  = "ANKRCHAINACCOUNT"
	IAvlStoreMainKey     = "ANKRCHAINMAIN"
	IAvlStoreContractKey = "ANKRCHAINCONTRACT"
	IavlStorePermKey     = "ANKRCHAINPERM"

	IavlStoreAccountDefCacheSize  = 10000
	IAvlStoreTxDefCacheSize       = 10000
	IAvlStoreContractDefCacheSize = 10000
	IAvlStorePermDefCacheSize     = 10000

	IAVLStoreAccountKeepVersionNum  = 100
	IAVLStoreMainKeepVersionNum     = 100
	IAVLStoreContractKeepVersionNum = 100
	IAVLStorePermKeepVersionNum     = 100

    CommitInfoKey     = "cminfo%d"
    LatestVerKey      = "latestverkey"
)

type IavlStoreMulti struct {
	db           dbm.DB
	storeMap     map[string]*IavlStore
	log          log.Logger
	cdc          *amino.Codec
}

type storeCommitID struct {
	Name string
	CID  ankrcmm.CommitID
}

type commitInfo struct {
	Version int64
	AppHash []byte
	Commits []storeCommitID
}

func NewIavlStoreMulti(db dbm.DB, storeLog log.Logger) *IavlStoreMulti {
	if db == nil {
		panic("can't create IavlStoreMulti, db nil")
	}

	if storeLog == nil {
		panic("can't create IavlStoreMulti, storeLog nil")
	}

	storeMap := make(map[string]*IavlStore)

	dbAcc := dbm.NewPrefixDB(db, []byte("ankr:"+IavlStoreAccountKey+"/"))
	storeMap[IavlStoreAccountKey] = NewIavlStore(dbAcc, IavlStoreAccountDefCacheSize, IAVLStoreAccountKeepVersionNum, storeLog.With("module", "accountstore"))

	dbTran := dbm.NewPrefixDB(db, []byte("ankr:"+IAvlStoreMainKey+"/"))
	storeMap[IAvlStoreMainKey] = NewIavlStore(dbTran, IAvlStoreTxDefCacheSize, IAVLStoreMainKeepVersionNum, storeLog.With("module", "txstore"))

	dbMt := dbm.NewPrefixDB(db, []byte("ankr:"+IAvlStoreContractKey+"/"))
	storeMap[IAvlStoreContractKey] = NewIavlStore(dbMt, IAvlStoreContractDefCacheSize, IAVLStoreContractKeepVersionNum, storeLog.With("module", "contractstore"))

	dbR := dbm.NewPrefixDB(db, []byte("ankr:"+IavlStorePermKey+"/"))
	storeMap[IavlStorePermKey] = NewIavlStore(dbR, IAvlStorePermDefCacheSize, IAVLStorePermKeepVersionNum, storeLog.With("module", "permstore"))

	return &IavlStoreMulti{db, storeMap, storeLog, amino.NewCodec()}
}

func (ms *IavlStoreMulti) IavlStore(storeKey string) *IavlStore {
	if store, ok := ms.storeMap[storeKey]; ok {
		return store
	}

	ms.log.Error("can't find the responding iavlstore", "storeKey", storeKey)

	return nil
}

// Load the latest versioned tree from disk.
func (ms *IavlStoreMulti) Load() {
	for key, iavlS := range ms.storeMap {
		ver, err := iavlS.Load()
		if err != nil {
			ms.log.Error("Load the latest db failed", "key", key, "err", err)
		}else {
			ms.log.Info("Load the latest db successful", "key", key, "ver", ver)
		}
	}
}

func (ms *IavlStoreMulti) commitInfo(version int64) *commitInfo {
	infoKey := fmt.Sprintf(CommitInfoKey, version)
	infoV := ms.db.Get([]byte(infoKey))
	if infoV != nil {
		var cmmInfo commitInfo
		ms.cdc.MustUnmarshalBinaryLengthPrefixed(infoV, &cmmInfo)
		return &cmmInfo
	}else {
		ms.log.Error("can't get commit info", "infokey", infoKey)

		zeroAppHash := make([]byte, 8)
		binary.PutVarint(zeroAppHash, int64(0))
		return &commitInfo{version, zeroAppHash, nil}
	}
}

func (ms *IavlStoreMulti) setCommitInfo(batch dbm.Batch, version int64, info commitInfo) {
	infoBytes := ms.cdc.MustMarshalBinaryLengthPrefixed(info)
	infoKey := fmt.Sprintf(CommitInfoKey, version)
	batch.Set([]byte(infoKey), infoBytes)
}

func (ms *IavlStoreMulti) latestVersion() int64 {
	verBytes := ms.db.Get([]byte(LatestVerKey))
	if verBytes == nil {
		return 0
	}

	var latestVer int64
	err := ms.cdc.UnmarshalBinaryLengthPrefixed(verBytes, &latestVer)
	if err != nil {
		panic(err)
	}

	return latestVer
}

func (ms *IavlStoreMulti) setLatestVersion(batch dbm.Batch, version int64) {
	latestVerBtest, _ := ms.cdc.MarshalBinaryLengthPrefixed(version)
	batch.Set([]byte(LatestVerKey), latestVerBtest)
}

func (ms *IavlStoreMulti) lastCommit() ankrcmm.CommitID {
	latestVer := ms.latestVersion()
	if latestVer == 0 {
		return ankrcmm.CommitID{}
	}

    cmmInfos :=  ms.commitInfo(latestVer)
    if cmmInfos != nil {
    	if cmmInfos.Version != latestVer {
    		ms.log.Error("error commitinfo, mmInfos.version != latestVer", "latestVer", latestVer, "commitInfoVer", cmmInfos.Version)
		}

		return ankrcmm.CommitID{latestVer, cmmInfos.AppHash}
	}else {
		ms.log.Error("can't get the latest commitinfo", "latestVer", latestVer)
	}

    return ankrcmm.CommitID{}
}

func (ms *IavlStoreMulti) Commit(version int64, totalTx int64) ankrcmm.CommitID {
	var cmmInfo commitInfo

	version += 1

	appHash := make([]byte, 8)
	binary.PutVarint(appHash, totalTx)

	cmmInfo.Version = version
	cmmInfo.AppHash = appHash

	hashM := make(map[string][]byte)
	for k, s := range ms.storeMap {
		commitID, err := s.Commit()
		if err != nil {
			panic(err)
		}

		hashM[k] = commitID.Hash

		cmmInfo.Commits = append(cmmInfo.Commits, storeCommitID{k,commitID})
	}

	batch := ms.db.NewBatch()
	defer batch.Close()

	ms.setCommitInfo(batch, version, cmmInfo)
	ms.setLatestVersion(batch, version)

	batch.Write()

	return ankrcmm.CommitID{version, cmmInfo.AppHash}
}




