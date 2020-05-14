package iavl

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/common/code"
	ankrapscmm "github.com/Ankr-network/ankr-chain/store/appstore/common"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/iavl"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	ChainIDKey = "chainidkey"
	TotalTxKey = "totaltxkey"
)

const (
	StoreCertKeyPrefix = "certkey:"
	StoreMeteringPrefix = "merting:"
	StoreContractInfoPrefix = "continfo:"
	StoreContractCurrencyPrefix = "contcur:"
	StoreCurrencyPrefix = "cur:"
)

//type storeQueryHandler func(store *IavlStoreApp, reqData []byte) (resQuery types.ResponseQuery)

type IavlStoreApp struct {
	iavlSM          *IavlStoreMulti
	lastCommitID    ankrcmm.CommitID
	totalTx         int64
	storeLog        log.Logger
	cdc             *amino.Codec
	kvState         ankrapscmm.State
	queryHandleMap  map[string]*storeQueryHandler
	accStoreLocker  sync.RWMutex
	certStoreLocker sync.RWMutex
	totalTxLocker   sync.RWMutex
}

func containCertKeyPrefix(dcnsName string) string {
	return containPrefix(dcnsName, StoreCertKeyPrefix)
}

func containMeteringPrefix(dcnsName string) string {
	return containPrefix(dcnsName, StoreMeteringPrefix)
}

func containContractInfoPrefix(cAddr string) string {
	return containPrefix(cAddr, StoreContractInfoPrefix)
}

func containContractCurrencyPrefix(symbol string) string {
	return containPrefix(symbol, StoreContractCurrencyPrefix)
}

func containCurrencyPrefix(symbol string) string {
	return containPrefix(symbol, StoreCurrencyPrefix)
}

func stripCertKeyPrefix(key string) (string, error) {
	return stripKeyPrefix(key, StoreCertKeyPrefix)
}

type storeQueryHandler struct {
	storeName string
	req interface{}
	callFunc interface{}
}

func NewIavlStoreApp(dbDir string, storeLog log.Logger) *IavlStoreApp {
	kvPath := filepath.Join(dbDir, "kvstore.db")
	isKVPathExist, err := ankrcmm.PathExists(kvPath)
	if err != nil {
		panic(err)
	}

	var kvState ankrapscmm.State
	if isKVPathExist {
		kvDB, err := dbm.NewGoLevelDB("kvstore", dbDir)
		if err != nil {
			panic(err)
		}
		kvState = ankrapscmm.LoadState(kvDB)

		os.RemoveAll(kvPath)
	}

	//appStorePath := filepath.Join(dbDir, "appstore.db")
	//os.RemoveAll(appStorePath)

	var lcmmID ankrcmm.CommitID

	db, err := dbm.NewGoLevelDB("appstore", dbDir)
	if err != nil {
		panic(err)
	}

	iavlSM := NewIavlStoreMulti(db, storeLog)

	lcmmID = iavlSM.lastCommit()

	fmt.Printf("lcmmID.version=%d, lcmmID.hash=%X, kvState=%v\n", lcmmID.Version, lcmmID.Hash, kvState)

	if lcmmID.Version  > 0 {
		fmt.Printf("theLastVersion'hash=%X\n", iavlSM.commitInfo(lcmmID.Version - 1).AppHash)
	}

	iavlSM.Load()

	iavlSApp := &IavlStoreApp{iavlSM: iavlSM, lastCommitID: lcmmID, storeLog: storeLog, cdc: amino.NewCodec(), kvState: kvState}

	iavlSApp.totalTxLocker.Lock()

	if !iavlSM.storeMap[IAvlStoreMainKey].Has([]byte(TotalTxKey)) {
		buf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(buf, int64(0))
		iavlSM.storeMap[IAvlStoreMainKey].Set([]byte(TotalTxKey), buf[:n])
	} else {
		totalTxBytes, err := iavlSM.storeMap[IAvlStoreMainKey].Get([]byte(TotalTxKey))
		if err == nil {
			iavlSApp.totalTx, _ = binary.Varint(totalTxBytes)
		}else {
			storeLog.Error("load txtal tx error", "err", err)
		}
	}

	lastHash := make([]byte, 8)
	binary.PutVarint(lastHash, iavlSApp.totalTx)

	iavlSApp.totalTxLocker.Unlock()

	if lcmmID.Hash != nil {
		lcmmID.Hash = lastHash
	}

	fmt.Printf("lcmmID.version=%d, lcmmID.hash=%X, totalTx=%d\n", lcmmID.Version, lcmmID.Hash, iavlSApp.totalTx)

	iavlSApp.lastCommitID = lcmmID

	iavlSApp.queryHandleMap = make(map[string]*storeQueryHandler)

	iavlSApp.queryHandleMap["nonce"]            = &storeQueryHandler{IavlStoreAccountKey, &ankrcmm.NonceQueryReq{},    iavlSApp.NonceQuery}
	iavlSApp.queryHandleMap["balance"]          = &storeQueryHandler{IavlStoreAccountKey, &ankrcmm.BalanceQueryReq{},  iavlSApp.BalanceQuery}
	iavlSApp.queryHandleMap["certkey"]          = &storeQueryHandler{IAvlStoreMainKey, &ankrcmm.CertKeyQueryReq{},  iavlSApp.CertKeyQuery}
	iavlSApp.queryHandleMap["metering"]         = &storeQueryHandler{IAvlStoreMainKey, &ankrcmm.MeteringQueryReq{}, iavlSApp.MeteringQuery}
	iavlSApp.queryHandleMap["validator"]        = &storeQueryHandler{IAvlStoreMainKey, &ankrcmm.ValidatorQueryReq{},iavlSApp.ValidatorQuery}
	iavlSApp.queryHandleMap["contract"]         = &storeQueryHandler{IAvlStoreContractKey, &ankrcmm.ContractQueryReq{}, iavlSApp.LoadContractQuery}
	iavlSApp.queryHandleMap["account"]          = &storeQueryHandler{IavlStoreAccountKey, &ankrcmm.AccountQueryReq{}, iavlSApp.AccountQuery}
	iavlSApp.queryHandleMap["currency"]         = &storeQueryHandler{IAvlStoreContractKey, &ankrcmm.CurrencyQueryReq{}, iavlSApp.CurrencyInfoQuery}
	iavlSApp.queryHandleMap["statisticalinfo"] = &storeQueryHandler{IAvlStoreMainKey, &ankrcmm.StatisticalInfoReq{}, iavlSApp.StatisticalInfoQuery}

	return iavlSApp
}

func NewMockIavlStoreApp() *IavlStoreApp {
	db := dbm.NewMemDB()
	storeLog := log.NewNopLogger()

	iavlSM := NewIavlStoreMulti(db, storeLog)
	lcmmID := iavlSM.lastCommit()

	iavlSApp := &IavlStoreApp{iavlSM: iavlSM, lastCommitID: lcmmID, storeLog: storeLog, cdc: amino.NewCodec()}

	iavlSApp.queryHandleMap = make(map[string]*storeQueryHandler)

	iavlSApp.queryHandleMap["nonce"]            = &storeQueryHandler{IavlStoreAccountKey, &ankrcmm.NonceQueryReq{},    iavlSApp.NonceQuery}
	iavlSApp.queryHandleMap["balance"]          = &storeQueryHandler{IavlStoreAccountKey, &ankrcmm.BalanceQueryReq{},  iavlSApp.BalanceQuery}
	iavlSApp.queryHandleMap["certkey"]          = &storeQueryHandler{IAvlStoreMainKey, &ankrcmm.CertKeyQueryReq{},  iavlSApp.CertKeyQuery}
	iavlSApp.queryHandleMap["metering"]         = &storeQueryHandler{IAvlStoreMainKey, &ankrcmm.MeteringQueryReq{}, iavlSApp.MeteringQuery}
	iavlSApp.queryHandleMap["validator"]        = &storeQueryHandler{IAvlStoreMainKey, &ankrcmm.ValidatorQueryReq{},iavlSApp.ValidatorQuery}
	iavlSApp.queryHandleMap["contract"]         = &storeQueryHandler{IAvlStoreContractKey, &ankrcmm.ContractQueryReq{}, iavlSApp.LoadContractQuery}
	iavlSApp.queryHandleMap["account"]          = &storeQueryHandler{IavlStoreAccountKey, &ankrcmm.AccountQueryReq{}, iavlSApp.AccountQuery}
	iavlSApp.queryHandleMap["currency"]         = &storeQueryHandler{IAvlStoreContractKey, &ankrcmm.CurrencyQueryReq{}, iavlSApp.CurrencyInfoQuery}
	iavlSApp.queryHandleMap["statisticalinfo"] = &storeQueryHandler{IAvlStoreMainKey, &ankrcmm.StatisticalInfoReq{}, iavlSApp.StatisticalInfoQuery}

	return  &IavlStoreApp{iavlSM: iavlSM, lastCommitID: lcmmID, storeLog: storeLog, cdc: amino.NewCodec()}
}

func (sp* IavlStoreApp) queryHandlerWapper(queryKey string, reqData []byte, height int64, prove bool) (resQuery types.ResponseQuery, storeKey string, proof *iavl.RangeProof) {
	defer func() {
		if rErr := recover(); rErr != nil {
			resQuery.Code = code.CodeTypeQueryInvalidQueryReqData
			resQuery.Log  = fmt.Sprintf("excetion catched, invalid %s query req data, err=%v", queryKey, rErr)
		}
	}()

	req      := sp.queryHandleMap[queryKey].req
	callFunc := sp.queryHandleMap[queryKey].callFunc
	err := sp.cdc.UnmarshalJSON(reqData, req)
	if err != nil {
		resQuery.Code = code.CodeTypeQueryInvalidQueryReqData
		resQuery.Log  = fmt.Sprintf("invalid %s query req data, err=%s", queryKey, err.Error())
		return
	}

	v := reflect.ValueOf(callFunc)
	var paramVals [] reflect.Value
	reqVals := reflect.ValueOf(req)
	reqVals = reflect.Indirect(reqVals)
	for i := 0; i < reqVals.NumField(); i++ {
		paramVals = append(paramVals, reqVals.Field(i))
	}
	paramVals = append(paramVals, reflect.ValueOf(height))
	paramVals = append(paramVals, reflect.ValueOf(prove))
	respVals := v.Call(paramVals)
	if respVals[3].Interface() != nil && respVals[3].Type().Name() == "error" {
		err :=  respVals[3].Interface().(error)
		resQuery.Code = code.CodeTypeLoadBalError
		resQuery.Log  = fmt.Sprintf("load %s query err, err=%s", queryKey, err.Error())

		storeKey = respVals[1].Interface().(string)
		proof    = respVals[2].Interface().(*iavl.RangeProof)

		return
	}

	if respVals[3].Interface() == nil {
		resQDataBytes, _ := sp.cdc.MarshalJSON(respVals[0].Interface())
		resQuery.Code    = code.CodeTypeOK
		resQuery.Value   = resQDataBytes

		storeKey = respVals[1].Interface().(string)
		proof    = respVals[2].Interface().(*iavl.RangeProof)

		return
	}

	resQuery.Code  = code.CodeTypeUnknownError
	resQuery.Log   = fmt.Sprintf("load %s query unknown err", queryKey)

	return
}

func (sp *IavlStoreApp) SetChainID(chainID string) {
	sp.iavlSM.IavlStore(IAvlStoreMainKey).Set([]byte(ChainIDKey), []byte(chainID))
}

func (sp *IavlStoreApp) ChainID() string {
	chainIDBytes, err := sp.iavlSM.IavlStore(IAvlStoreMainKey).Get([]byte(ChainIDKey))
	if err != nil || chainIDBytes == nil{
		return ""
	}

	return string(chainIDBytes)
}

func (sp *IavlStoreApp) LastCommit() *ankrcmm.CommitID{
	return &sp.lastCommitID
}

func (sp *IavlStoreApp) Commit() types.ResponseCommit {
	sp.totalTxLocker.RLock()
	curTotalTx := sp.totalTx
	sp.totalTxLocker.RUnlock()

    commitID := sp.iavlSM.Commit(sp.lastCommitID.Version, curTotalTx)

	sp.lastCommitID.Hash = sp.lastCommitID.Hash[0:0]

    sp.lastCommitID.Version = commitID.Version
	sp.lastCommitID.Hash    = append(sp.lastCommitID.Hash, commitID.Hash...)

	sp.storeLog.Info("IavlStoreApp Commit", "totalTx", curTotalTx, "appHash", fmt.Sprintf("%X", commitID.Hash))

	return types.ResponseCommit{Data: commitID.Hash}
}

func (sp *IavlStoreApp) parsePath(path string)(storeName string, subPath string) {
	if path == "" || !strings.HasPrefix(path, "/"){
		sp.storeLog.Error("invalid path", "path", path)
		return "", ""
	}

	pathSegs := strings.Split(path[1:], "/")
	storeName = pathSegs[0]
	if len(pathSegs) == 2 {
		subPath = "/" + pathSegs[1]
	}

	return
}

func (sp *IavlStoreApp) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	resQuery.Log = "exists"
	if reqQuery.Path == "" {
		resQuery.Log = "blank store name"
		resQuery.Code = code.CodeTypeQueryInvalidStoreName
		return
	}

	resQuery, storeKey, proof := sp.queryHandlerWapper(reqQuery.Path, reqQuery.Data, reqQuery.Height, reqQuery.Prove)

	if reqQuery.Height == 0 {
		resQuery.Height = sp.Height()
	}else {
		resQuery.Height = reqQuery.Height
	}

	resQuery.Key = []byte(storeKey)

	if resQuery.Code == code.CodeTypeOK && reqQuery.Prove {
		if resQuery.Value != nil {
			resQuery.Proof = &merkle.Proof{Ops: []merkle.ProofOp{iavl.NewIAVLValueOp([]byte(storeKey), proof).ProofOp()}}
		} else {
			resQuery.Proof = &merkle.Proof{Ops: []merkle.ProofOp{iavl.NewIAVLAbsenceOp([]byte(storeKey), proof).ProofOp()}}
		}

		qVer := reqQuery.Height
		if qVer == 0 {
			qVer = sp.lastCommitID.Version
		}
		commInfo := sp.iavlSM.commitInfo(qVer)

		if commInfo != nil {
			pOP := NewIavlStoreMultiOp(
				[]byte(sp.queryHandleMap[reqQuery.Path].storeName),
				&IavlStoreMultiProof{*commInfo},
				).ProofOp()

			resQuery.Proof.Ops = append(resQuery.Proof.Ops, pOP)
		}
	}

    return
}

func (sp *IavlStoreApp) SetCertKey(dcName string, pemBase64 string)  {
	key := []byte(containCertKeyPrefix(dcName))
	sp.iavlSM.IavlStore(IAvlStoreMainKey).Set(key, []byte(pemBase64))
}

func (sp *IavlStoreApp) CertKey(dcName string, height int64, prove bool)(string, string, *iavl.RangeProof, []byte) {
	if dcName == "" {
		sp.storeLog.Error("CertKey, blank dcName")
		return "", "", nil, nil
	}

	key := []byte(containCertKeyPrefix(dcName))
	valBytes, proof, err :=  sp.iavlSM.IavlStore(IAvlStoreMainKey).GetWithVersionProve(key, height, prove)
	if err != nil {
		sp.storeLog.Error("can't get the key's value", "dcName", dcName)
		return "", containCertKeyPrefix(dcName), nil, nil
	}

	return string(valBytes), containCertKeyPrefix(dcName), proof, valBytes
}

func (sp *IavlStoreApp) CertKeyQuery(dcName string, height int64, prove bool) (*ankrcmm.QueryResp, string, *iavl.RangeProof, error) {
	pemBase64, storeKey, proof, proofVal := sp.CertKey(dcName, height, prove)
	respData, err := sp.cdc.MarshalJSON(&ankrcmm.CertKeyQueryResp{pemBase64})
	if err != nil {
		return nil, storeKey, proof, err
	}

	return &ankrcmm.QueryResp{respData, proofVal}, storeKey, proof, nil
}

func (sp *IavlStoreApp) DeleteCertKey(dcName string) {
	key := []byte(containCertKeyPrefix(dcName))
	sp.iavlSM.IavlStore(IAvlStoreMainKey).Remove(key)
}

func (sp *IavlStoreApp) CertKeyList() ([]byte, uint64) {
	certKeyCount := uint64(0)
	certKeyList  := ""

	endBytes := prefixEndBytes([]byte(StoreCertKeyPrefix))

	sp.iavlSM.storeMap[IAvlStoreMainKey].tree.IterateRange([]byte(StoreCertKeyPrefix), endBytes, true, func(key []byte, value []byte) bool{
		if len(key) >= len(StoreAccountPrefix) && string(key[0:len(StoreCertKeyPrefix)]) == StoreCertKeyPrefix {
			dcnsName, err := stripCertKeyPrefix(string(key))
			if err != nil {
				sp.storeLog.Error("stripCertKeyPrefix error", "err", err)
			}else {
				certKeyCount++
				certKeyList += certKeyList + ";" + dcnsName + ":" + string(value)
			}
		}

		return false
	})

	if certKeyCount > 0 {
		certKeyList = certKeyList[1:]
		return []byte(certKeyList), certKeyCount
	}else {
		return nil, certKeyCount
	}
}

func (sp *IavlStoreApp) SetMetering(dcName string, nsName string, value string) {
	key := []byte(containMeteringPrefix(dcName+"_"+nsName))
	sp.iavlSM.IavlStore(IAvlStoreMainKey).Set(key, []byte(value))
}

func (sp *IavlStoreApp) Metering(dcName string, nsName string, height int64, prove bool) (string, string, *iavl.RangeProof, []byte) {
	if dcName == "" || nsName == "" {
		sp.storeLog.Error("Metering, blank dcName or nsName", "dcName", dcName, "nsName", nsName)
		return "", "", nil, nil
	}

	key := []byte(containMeteringPrefix(dcName+"_"+nsName))
	valueBytes, proof, err := sp.iavlSM.IavlStore(IAvlStoreMainKey).GetWithVersionProve(key, height, prove)
	if err != nil {
		sp.storeLog.Error("can't get the responding metering value", "dcName", dcName, "nsName", nsName, "err", err)
		return "", containMeteringPrefix(dcName+"_"+nsName), nil, nil
	}

	return string(valueBytes),  containMeteringPrefix(dcName+"_"+nsName), proof, valueBytes
}

func (sp *IavlStoreApp) MeteringQuery(dcName string, nsName string, height int64, prove bool) (*ankrcmm.QueryResp, string, *iavl.RangeProof, error) {
	val, storeKey, proof, proofVal := sp.Metering(dcName, nsName, height, prove)
	respData, err := sp.cdc.MarshalJSON(&ankrcmm.MeteringQueryResp{val})
	if err != nil {
		return nil, storeKey, proof, err
	}

	return &ankrcmm.QueryResp{respData, proofVal}, storeKey, proof, nil
}

func (sp *IavlStoreApp) SetValidator(valInfo *ankrcmm.ValidatorInfo) {
	valBytes := ankrcmm.EncodeValidatorInfo(sp.cdc, valInfo)

	sp.iavlSM.IavlStore(IAvlStoreMainKey).Set([]byte(containValidatorPrefix(valInfo.ValAddress)), valBytes)
}

func (sp *IavlStoreApp) Validator(valAddr string, height int64, prove bool) (*ankrcmm.ValidatorInfo, string, *iavl.RangeProof, []byte, error) {
	if valAddr == "" {
		return nil, "", nil, nil, errors.New("Validator, blank valAddr")
	}

	valBytes, proof, err := sp.iavlSM.IavlStore(IAvlStoreMainKey).GetWithVersionProve([]byte(containValidatorPrefix(valAddr)), height, prove)
	if err != nil {
		return nil, containValidatorPrefix(valAddr), nil, nil, fmt.Errorf("can't get the responding validator info: valAddr=%s", valAddr)
	}

	valInfo := ankrcmm.DecodeValidatorInfo(sp.cdc, valBytes)

	return  &valInfo, containValidatorPrefix(valAddr), proof, valBytes, nil
}

func (sp *IavlStoreApp) ValidatorQuery(valAddr string, height int64, prove bool) (*ankrcmm.QueryResp, string, *iavl.RangeProof, error) {
	valInfo, storeKey, proof, proofVal, err := sp.Validator(valAddr, height, prove)
	if err != nil {
		return nil, storeKey, proof, err
	}

	valQResp := &ankrcmm.ValidatorQueryResp{
		Name:         valInfo.Name,
		ValAddress:   valInfo.ValAddress,
		PubKey:       valInfo.PubKey,
		Power:        valInfo.Power,
		StakeAddress: valInfo.StakeAddress,
		StakeAmount:  valInfo.StakeAmount,
		ValidHeight:  valInfo.ValidHeight,
	}

	respData, err := sp.cdc.MarshalJSON(valQResp)
	if err != nil {
		return nil, storeKey, proof, err
	}

	return &ankrcmm.QueryResp{respData, proofVal}, storeKey, proof, nil
}

func (sp *IavlStoreApp) RemoveValidator(valAddr string) {
	sp.iavlSM.IavlStore(IAvlStoreMainKey).Remove([]byte(containValidatorPrefix(valAddr)))
}

func (sp *IavlStoreApp) TotalValidatorPowers() int64 {
	valPower := int64(0)

	endBytes := prefixEndBytes([]byte(StoreValidatorPrefix))

	sp.iavlSM.storeMap[IAvlStoreMainKey].tree.IterateRange([]byte(StoreValidatorPrefix), endBytes, true, func(key []byte, value []byte) bool{
		if len(key) >= len(StoreValidatorPrefix) && string(key[0:len(StoreValidatorPrefix)]) == StoreValidatorPrefix {
			valInfo := ankrcmm.DecodeValidatorInfo(sp.cdc, value)

			valPower += valInfo.Power
		}

		return false
	})

	return valPower
}

func (sp *IavlStoreApp) Get(key []byte) []byte {
	val, err := sp.iavlSM.IavlStore(IAvlStoreMainKey).Get(key)
	if err != nil {
		sp.storeLog.Error("can't get the key value", "key", string(key))
		val = nil
	}

	return val
}

func (sp *IavlStoreApp) Set(key []byte, val []byte) {
	sp.iavlSM.IavlStore(IAvlStoreMainKey).Set(key, val)
}

func (sp *IavlStoreApp) Delete(key []byte) {
	sp.iavlSM.IavlStore(IAvlStoreMainKey).Remove(key)
}

func (sp *IavlStoreApp) Has(key []byte) bool {
	return sp.iavlSM.IavlStore(IAvlStoreMainKey).Has(key)
}

func (sp *IavlStoreApp) Height() int64 {
	return sp.lastCommitID.Version
}

func (sp *IavlStoreApp) totalTxWithoutlock(height int64, prove bool) (int64, string, *iavl.RangeProof, []byte, error) {
	if height <= 0 {
		height = sp.Height()
	}

	if sp.iavlSM.storeMap[IAvlStoreMainKey].Has([]byte(TotalTxKey)) {
		val, proof, err := sp.iavlSM.storeMap[IAvlStoreMainKey].GetWithVersionProve([]byte(TotalTxKey), height, prove)
		if err != nil {
			return 0, TotalTxKey, nil, nil, err
		}

		totalTx, _ := binary.Varint(val)

		sp.storeLog.Info("IavlStoreApp totalTxWithoutlock", "sp.Height", sp.Height, "sp.totalTx", sp.totalTx, "store.totalTx", totalTx)

		return totalTx, TotalTxKey, proof, val, nil
	}

	return 0, TotalTxKey, nil, nil, fmt.Errorf("Not exist TotalTxKey(%s), height=%d", TotalTxKey, height)
}

func (sp *IavlStoreApp) TotalTx(height int64, prove bool) (int64, string, *iavl.RangeProof, []byte, error) {
	sp.totalTxLocker.RLock()
	defer  sp.totalTxLocker.RUnlock()

	sp.storeLog.Info("IavlStoreApp TotalTx", "sp.Height", sp.Height, "height", height, "prove", prove)

	return sp.totalTxWithoutlock(height, prove)
}

func (sp *IavlStoreApp) totalTxLatestWithoutlock() (int64, error) {
	if sp.iavlSM.storeMap[IAvlStoreMainKey].Has([]byte(TotalTxKey)) {
		val, err := sp.iavlSM.storeMap[IAvlStoreMainKey].Get([]byte(TotalTxKey))
		if err != nil {
			return 0, err
		}

		totalTx, _ := binary.Varint(val)

		sp.storeLog.Info("IavlStoreApp totalTxWithoutlock", "sp.Height", sp.Height(), "sp.totalTx", sp.totalTx, "store.totalTx", totalTx)

		return totalTx, nil
	}

	return 0, fmt.Errorf("Not exist TotalTxKey(%s), height=%d", TotalTxKey, sp.Height())
}

func (sp *IavlStoreApp) setTotalTxWithoutlock(totalTx int64) {
	sp.totalTx = totalTx

	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, sp.totalTx)
	sp.iavlSM.storeMap[IAvlStoreMainKey].Set([]byte(TotalTxKey), buf[:n])

	sp.storeLog.Info("IavlStoreApp setTotalTxWithoutlock", "sp.Height", sp.Height(), "sp.totalTx", sp.totalTx)
}

func (sp *IavlStoreApp) SetTotalTx(totalTx int64) {
	sp.totalTxLocker.Lock()
	defer sp.totalTxLocker.Unlock()

	sp.storeLog.Info("IavlStoreApp SetTotalTx", "sp.Height", sp.Height(), "sp.totalTx", sp.totalTx, "totalTx", totalTx)

	sp.setTotalTxWithoutlock(totalTx)
}

func (sp *IavlStoreApp) IncTotalTx() int64 {
	sp.totalTxLocker.Lock()
	defer sp.totalTxLocker.Unlock()

	sp.storeLog.Info("IavlStoreApp IncTotalTx", "sp.Height", sp.Height(), "sp.totalTx", sp.totalTx)

	sp.totalTx++
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, sp.totalTx)
	sp.iavlSM.storeMap[IAvlStoreMainKey].Set([]byte(TotalTxKey), buf[:n])

	sp.storeLog.Info("IavlStoreApp IncTotalTx", "sp.Height", sp.Height(), "store.TotalTx", sp.totalTx)

	return sp.totalTx
}

func (sp *IavlStoreApp) APPHash() []byte {
	return sp.lastCommitID.Hash
}

func (sp *IavlStoreApp) APPHashByHeight(height int64) []byte {
	return sp.iavlSM.commitInfo(height).AppHash
}

func (sp *IavlStoreApp) KVState() ankrapscmm.State {
	return sp.kvState
}

func (sp *IavlStoreApp) ResetKVState() {
	sp.kvState = ankrapscmm.State{}
}

func (sp *IavlStoreApp) Rollback() {
	sp.totalTxLocker.Lock()
	defer sp.totalTxLocker.Unlock()

	curTotalTx := sp.totalTx//sp.totalTxLatestWithoutlock

	for _, iavlS := range sp.iavlSM.storeMap {
		iavlS.Rollback()
	}

	sp.setTotalTxWithoutlock(curTotalTx)
}

func (sp *IavlStoreApp) DB() dbm.DB {
	return sp.iavlSM.db
}

func (sp *IavlStoreApp) StatisticalInfoQuery(height int64, prove bool)(*ankrcmm.QueryResp, string, *iavl.RangeProof, error) {
	addrArry, _ := sp.AccountList(height)
	totalTx, storeKey, proof, proofVal, err := sp.TotalTx(height, prove)
	if err != nil {
		totalTx = 0
	}

	sInfoResp := &ankrcmm.StatisticalInfoResp{
		addrArry,
		totalTx,
	}

	respData, err := sp.cdc.MarshalJSON(sInfoResp)
	if err != nil {
		return nil, storeKey, proof, err
	}

	return &ankrcmm.QueryResp{respData,proofVal }, storeKey, proof, err
}

func (sp *IavlStoreApp) IsExist(cAddr string) bool {
	return sp.iavlSM.IavlStore(IAvlStoreContractKey).Has([]byte(cAddr))
}

func (sp *IavlStoreApp) CreateCurrency(symbol string, currency *ankrcmm.CurrencyInfo) error {
	if sp.iavlSM.IavlStore(IAvlStoreContractKey).Has([]byte(containCurrencyPrefix(symbol))) {
		 sp.storeLog.Info("CreateCurrency, currency has existed and its info will be updated, symbol=%s", symbol)
	}

	curBytes, _ := sp.cdc.MarshalJSON(currency)

	isSucess := sp.iavlSM.IavlStore(IAvlStoreContractKey).Set([]byte(containCurrencyPrefix(symbol)), curBytes)
	if !isSucess {
		return fmt.Errorf("create currency error, symbol=%s", symbol)
	}

	return nil
}

func (sp *IavlStoreApp) CurrencyInfo(symbol string, height int64, prove bool) (*ankrcmm.CurrencyInfo, string, *iavl.RangeProof, []byte, error) {
	if symbol == "" {
		return nil, "", nil, nil, errors.New("CurrencyInfo, blank symbol name")
	}

	curBytes, proof, err := sp.iavlSM.IavlStore(IAvlStoreContractKey).GetWithVersionProve([]byte(containCurrencyPrefix(symbol)), height, prove)
	if err != nil || len(curBytes) == 0 {
		sp.storeLog.Error("can't get the currency", "symbol", symbol)
		return nil, containCurrencyPrefix(symbol), nil, nil, err
	}

	var curInfo ankrcmm.CurrencyInfo

	err = sp.cdc.UnmarshalJSON(curBytes, &curInfo)
	if err != nil {
		return nil, containCurrencyPrefix(symbol), proof, curBytes, err
	}

	return &curInfo, containCurrencyPrefix(symbol), proof, curBytes, nil
}

func (sp *IavlStoreApp) CurrencyInfoQuery(symbol string, height int64, prove bool) (*ankrcmm.QueryResp, string, *iavl.RangeProof, error) {
	cInfo, storeKey, proof, proofVal, err := sp.CurrencyInfo(symbol, height, prove)
	if err != nil {
		return nil, storeKey, proof, err
	}

	cInfoQResp := &ankrcmm.CurrencyQueryResp{}
	cInfoQResp.Symbol      = cInfo.Symbol
	cInfoQResp.Decimal     = cInfo.Decimal
	cInfoQResp.TotalSupply = cInfo.TotalSupply

	respData, err := sp.cdc.MarshalJSON(cInfoQResp)
	if err != nil {
		return nil, storeKey, proof, err
	}

	return &ankrcmm.QueryResp{respData, proofVal},  storeKey, proof, nil
}

func (sp *IavlStoreApp) BuildCurrencyCAddrMap(symbol string, cAddr string) error {
	if sp.iavlSM.IavlStore(IAvlStoreContractKey).Has([]byte(containContractCurrencyPrefix(symbol))) {
		return errors.New("the contract name has existed")
	}

	sp.iavlSM.IavlStore(IAvlStoreContractKey).Set([]byte(containContractCurrencyPrefix(symbol)), []byte(cAddr))

	return nil
}

func (sp *IavlStoreApp) ContractAddrBySymbol(symbol string) (string, error) {
	cAddrBytes, err := sp.iavlSM.IavlStore(IAvlStoreContractKey).Get([]byte(containContractCurrencyPrefix(symbol)))
	if err != nil {
		sp.storeLog.Error("can't get the contract addr", "symbol", symbol)
		return "", err
	}

	if cAddrBytes != nil {
		return string(cAddrBytes), nil
	}

	return "", nil
}

func (sp *IavlStoreApp) SaveContract(cAddr string, cInfo *ankrcmm.ContractInfo) error{
	if sp.iavlSM.IavlStore(IAvlStoreContractKey).Has([]byte(containContractInfoPrefix(cAddr))) {
		return errors.New("the contract name has existed")
	}

	cInfoBytes := ankrcmm.EncodeContractInfo(sp.cdc, cInfo)

	sp.iavlSM.IavlStore(IAvlStoreContractKey).Set([]byte(containContractInfoPrefix(cAddr)), cInfoBytes)

	return nil
}

func (sp *IavlStoreApp) LoadContract(cAddr string, height int64, prove bool) (*ankrcmm.ContractInfo, string, *iavl.RangeProof, []byte, error) {
	if cAddr == "" {
		return nil, "", nil, nil, errors.New("LoadContract, blank cAddr")
	}

	cInfoBytes, proof, err := sp.iavlSM.IavlStore(IAvlStoreContractKey).GetWithVersionProve([]byte(containContractInfoPrefix(cAddr)), height, prove)
	if err != nil || len(cInfoBytes) == 0 {
		sp.storeLog.Error("can't get the contract", "addr", cAddr)
		return nil, containContractInfoPrefix(cAddr), nil, nil, err
	}

	cInfo := ankrcmm.DecodeContractInfo(sp.cdc, cInfoBytes)

	return &cInfo, containContractInfoPrefix(cAddr), proof, cInfoBytes, nil
}

func (sp *IavlStoreApp) UpdateContractState(cAddr string, state ankrcmm.ContractState) error {
	if cAddr == "" {
		return errors.New("UpdateContractState, blank cAddr")
	}
	cInfoBytes, err := sp.iavlSM.IavlStore(IAvlStoreContractKey).Get([]byte(containContractInfoPrefix(cAddr)))
	if err != nil || len(cInfoBytes) == 0 {
		sp.storeLog.Error("UpdateContractState can't get the contract", "addr", cAddr)
		return err
	}

	cInfo := ankrcmm.DecodeContractInfo(sp.cdc, cInfoBytes)
	cInfo.State = state


	cInfoBytes = ankrcmm.EncodeContractInfo(sp.cdc, &cInfo)
	sp.iavlSM.IavlStore(IAvlStoreContractKey).Set([]byte(containContractInfoPrefix(cAddr)), cInfoBytes)

	return nil
}

func (sp *IavlStoreApp) ChangeContractOwner(cAddr string, ownerAddr string) error {
	if cAddr == "" {
		return errors.New("ChangeContractOwner, blank cAddr")
	}
	cInfoBytes, err := sp.iavlSM.IavlStore(IAvlStoreContractKey).Get([]byte(containContractInfoPrefix(cAddr)))
	if err != nil || len(cInfoBytes) == 0 {
		sp.storeLog.Error("ChangeContractOwner can't get the contract", "addr", cAddr)
		return err
	}

	cInfo := ankrcmm.DecodeContractInfo(sp.cdc, cInfoBytes)
	cInfo.Owner = ownerAddr


	cInfoBytes = ankrcmm.EncodeContractInfo(sp.cdc, &cInfo)
	sp.iavlSM.IavlStore(IAvlStoreContractKey).Set([]byte(containContractInfoPrefix(cAddr)), cInfoBytes)

	return nil
}

func (sp *IavlStoreApp) IsContractNormal(cAddr string) bool {
	if cAddr == "" {
		return false
	}
	cInfoBytes, err := sp.iavlSM.IavlStore(IAvlStoreContractKey).Get([]byte(containContractInfoPrefix(cAddr)))
	if err != nil || len(cInfoBytes) == 0 {
		sp.storeLog.Error("IsContractNormal can't get the contract", "addr", cAddr)
		return false
	}

	cInfo := ankrcmm.DecodeContractInfo(sp.cdc, cInfoBytes)
	if cInfo.State != ankrcmm.ContractNormal {
		return false
	}

	return true
}

func (sp *IavlStoreApp) LoadContractQuery(cAddr string, height int64, prove bool) (*ankrcmm.QueryResp, string, *iavl.RangeProof, error) {
	cInfo, storeKey, proof, proofVal, err := sp.LoadContract(cAddr, height, prove)
	if err != nil {
		return nil, storeKey, proof, err
	}

	cInfoQResp := &ankrcmm.ContractQueryResp{}
	cInfoQResp.Addr  = cInfo.Addr
	cInfoQResp.Name  = cInfo.Name
	cInfoQResp.Owner = cInfo.Owner

	cInfoQResp.Codes = make([]byte, len(cInfo.Codes))
	copy(cInfoQResp.Codes[:], cInfo.Codes)

	cInfoQResp.CodesDesc = cInfo.CodesDesc

	respData, err := sp.cdc.MarshalJSON(cInfoQResp)
	if err != nil {
		return nil, storeKey, proof, err
	}

	return &ankrcmm.QueryResp{respData, proofVal},  storeKey, proof, nil
}

func (sp *IavlStoreApp) AddContractRelatedObject(cAddr string, key string, jsonObject string) error{
	if cAddr == "" {
		return errors.New("AddContractRelatedObject, blank cAddr")
	}
	cInfoBytes, err := sp.iavlSM.IavlStore(IAvlStoreContractKey).Get([]byte(containContractInfoPrefix(cAddr)))
	if err != nil || len(cInfoBytes) == 0 {
		sp.storeLog.Error("AddContractRelatedObject can't get the contract", "addr", cAddr)
		return err
	}

	cInfo := ankrcmm.DecodeContractInfo(sp.cdc, cInfoBytes)

	if _, ok := cInfo.RelatedInfos[key]; !ok {
		cInfo.RelatedInfos[key] = jsonObject
	}

	cInfoBytes = ankrcmm.EncodeContractInfo(sp.cdc, &cInfo)
	sp.iavlSM.IavlStore(IAvlStoreContractKey).Set([]byte(containContractInfoPrefix(cAddr)), cInfoBytes)

	return nil
}

func (sp *IavlStoreApp) LoadContractRelatedObject(cAddr string, key string)(jsonObject string, err error) {
	if cAddr == "" {
		return "", errors.New("LoadContractRelatedObject, blank cAddr")
	}
	cInfoBytes, err := sp.iavlSM.IavlStore(IAvlStoreContractKey).Get([]byte(containContractInfoPrefix(cAddr)))
	if err != nil || len(cInfoBytes) == 0 {
		sp.storeLog.Error("LoadContractRelatedObject can't get the contract", "addr", cAddr)
		return "", err
	}

	cInfo := ankrcmm.DecodeContractInfo(sp.cdc, cInfoBytes)

	if val, ok := cInfo.RelatedInfos[key]; ok {
		return val, nil
	}

	return "",  fmt.Errorf("LoadContractRelatedObject: the contract(%s) hasn't the related key(%s) info", cAddr, key)
}


