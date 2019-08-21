package metering

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/common/code"
	apm "github.com/Ankr-network/ankr-chain/tx"
	"github.com/Ankr-network/ankr-chain/store/appstore"
	ankrtypes "github.com/Ankr-network/ankr-chain/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type SetCertMsg struct {
	apm.TxMsg
}

func (sc *SetCertMsg) GasWanted() int64 {
	return 0
}

func (sc *SetCertMsg) GasUsed() int64 {
	return 0
}

func (sc *SetCertMsg) ProcessTx(txMsg interface{}, appStore appstore.AppStore, isOnlyCheck bool) (uint32, string, []cmn.KVPair) {
	trxSetCertSlices, ok := txMsg.([]string)
	if !ok {
		return  code.CodeTypeEncodingError, fmt.Sprintf("invalid tx set cert msg"), nil
	}

	if len(trxSetCertSlices) != 4 {
		return code.CodeTypeEncodingError, fmt.Sprintf("Expected trx set cert. Got %v", trxSetCertSlices), nil
	}
	dcS     := trxSetCertSlices[0]
	pemB64S := trxSetCertSlices[1]
	nonceS  := trxSetCertSlices[2]
	sigS    := trxSetCertSlices[3]

	nonceInt, err_nonce := strconv.ParseInt(string(nonceS), 10, 64)
	if err_nonce != nil {
		return code.CodeTypeEncodingError, fmt.Sprintf("Unexpected cert nonce. Got %v, %v", nonceS, err_nonce), nil
	}

	nonceOldByte := appStore.Get([]byte(ankrtypes.SET_CRT_NONCE))
	nonceOld, err_nonce := strconv.ParseInt(string(nonceOldByte), 10, 64)
	if err_nonce != nil {
		if len(string(nonceOldByte)) == 0 {
			nonceOld = 0
		} else {
			return  code.CodeTypeEncodingError, fmt.Sprintf("Unexpected nonce5. Got %v", nonceOld), nil
		}
	}

	if nonceOld + 1 != nonceInt {
		return code.CodeTypeEncodingError, fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS), nil
	}

	var admin_pubkey_str string = ""
	admin_pubkey := appStore.Get([]byte(ankrtypes.ADMIN_OP_METERING_PUBKEY_NAME))
	if len(admin_pubkey) == 0 {
		fmt.Println("use default ADMIN_OP_METERING_PUBKEY_NAME")
		admin_pubkey_str = adminPubKeyOfMetering()
	} else {
		admin_pubkey_str = string(admin_pubkey)
	}

	pDec, _ := base64.StdEncoding.DecodeString(sigS)
	pubKeyObject, err := common.DeserilizePubKey(admin_pubkey_str) //set by super user
	if err != nil {
		return  code.CodeTypeEncodingError, fmt.Sprintf("Deserilize pubkey failure. Got %v", admin_pubkey_str), nil
	}

	s256 := common.ConvertBySha256([]byte(fmt.Sprintf("%s%s%s", dcS, pemB64S, nonceS)))
	bb := pubKeyObject.VerifyBytes(s256[:32], pDec)
	if !bb {
		fmt.Println("Bad signature, transaction failed.", sigS)
		return code.CodeTypeEncodingError, fmt.Sprintf("Bad signature. Got %v", sigS), nil
	}

	if isOnlyCheck {
		return code.CodeTypeOK, "", nil
	}

	appStore.Set(([]byte(ankrtypes.SET_CRT_NONCE)) ,[]byte(nonceS))
	appStore.Set(prefixCertKey([]byte(dcS)), []byte(pemB64S))

	tags := []cmn.KVPair{
		{Key: []byte("app.type"), Value: []byte("RemoveCert")},
	}

	return code.CodeTypeOK, "", tags
}

type RemoveCertMsg struct {
	apm.TxMsg
}

func (rc *RemoveCertMsg) GasWanted() int64 {
	return 0
}

func (rc *RemoveCertMsg) GasUsed() int64 {
	return 0
}

func (rc *RemoveCertMsg) ProcessTx(tx []byte, appStore appstore.AppStore, isOnlyCheck bool) (uint32, string, []cmn.KVPair) {
	tx = tx[len(ankrtypes.RemoveCertPrefix):]
	trxSetCertSlices := strings.SplitN(string(tx), ":", 3)
	if len(trxSetCertSlices) != 3 {
		return code.CodeTypeEncodingError, fmt.Sprintf("Expected trx remove cert. Got %v", trxSetCertSlices), nil
	}
	dcS    := trxSetCertSlices[0]
	nonceS := trxSetCertSlices[1]
	sigS   := trxSetCertSlices[2]

	nonceInt, err_nonce := strconv.ParseInt(string(nonceS), 10, 64)
	if err_nonce != nil {
		return code.CodeTypeEncodingError, fmt.Sprintf("Unexpected nonce6. Got %v", nonceS), nil
	}

	nonceOldByte := appStore.Get([]byte(ankrtypes.RMV_CRT_NONCE))
	nonceOld, err_nonce := strconv.ParseInt(string(nonceOldByte), 10, 64)
	if err_nonce != nil {
		if len(string(nonceOldByte)) == 0 {
			nonceOld = 0
		} else {
			return code.CodeTypeEncodingError, fmt.Sprintf("Unexpected nonce. Got %v", nonceOld), nil
		}
	}

	if nonceOld + 1 != nonceInt {
		return code.CodeTypeEncodingError, fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS), nil
	}

	// verify sig
	var admin_pubkey_str string = ""
	admin_pubkey := appStore.Get([]byte(ankrtypes.ADMIN_OP_METERING_PUBKEY_NAME))
	if len(admin_pubkey) == 0 {
		fmt.Println("use default ADMIN_OP_METERING_PUBKEY_NAME")
		admin_pubkey_str = adminPubKeyOfMetering()
	} else {
		admin_pubkey_str = string(admin_pubkey)
	}

	pDec, _ := base64.StdEncoding.DecodeString(sigS)
	pubKeyObject, err := common.DeserilizePubKey(admin_pubkey_str) //set by super user
	if err != nil {
		return code.CodeTypeEncodingError, fmt.Sprintf("Deserilize pubkey failure. Got %v", admin_pubkey_str), nil
	}

	s256 := common.ConvertBySha256([]byte(fmt.Sprintf("%s%s", dcS, nonceS)))
	bb := pubKeyObject.VerifyBytes(s256[:32], pDec)
	if !bb {
		fmt.Println("Bad signature, transaction failed.", sigS)
		return code.CodeTypeEncodingError, fmt.Sprintf("Bad signature. Got %v", sigS), nil
	}

	if isOnlyCheck {
		return code.CodeTypeOK, "", nil
	}

	appStore.Set(([]byte(ankrtypes.RMV_CRT_NONCE)), []byte(nonceS))
	appStore.Delete(prefixCertKey([]byte(dcS)))

	tags := []cmn.KVPair{
		{Key: []byte("app.type"), Value: []byte("RemoveCert")},
	}

	return code.CodeTypeOK, "", tags
}
