package ankrchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/Ankr-network/ankr-chain/consensus/code"
    "github.com/Ankr-network/ankr-chain/store/appstore"
	ankrtypes "github.com/Ankr-network/ankr-chain/types"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	tmCoreTypes "github.com/tendermint/tendermint/types"

	//signmanager "github.com/Ankr-network/dccn-common/cert/sign"
)

const (
	MIN_TOKEN_SEND = "5000000000000000000" // 5 tokens
/*
test account:
JXEcRZfkjeMZ+W1zOP9KFEc7Py9noko6i6KLe3UzPX0jF82cs7uIVjybtyGPHRyFw3wk708QnPvPP85ALZeK5Q==
IxfNnLO7iFY8m7chjx0chcN8JO9PEJz7zz/OQC2XiuU=
64BC85F08C03F42B17EAAF5AFFAF9BFAF96CFCB85CA2F3
*/
        // for test
	//INIT_ADDRESS = "B508ED0D54597D516A680E7951F18CAD24C7EC9FCFCD67"
	//GAS_ADDRESS = "64BC85F08C03F42B17EAAF5AFFAF9BFAF96CFCB85CA2F3"

	// for prod
	INIT_ADDRESS = "52E90523B5262E3AC2582F08A23068EE898D445EDF4D18"
	GAS_ADDRESS = "47A65FBF3FADD12B81959AA3D8DF5E300E8C9FBFF98770"
)

const PubKeyEd25519Size = 32
/* for test
Q5P4l16P+/Cyxq3BvavuWnQPkmeHNYPFkjfuWyQoNyK2vCvT1jyyoh2DYfu+EkWx/hoGjAHOqQw6PMAa7ZkXoQ==
trwr09Y8sqIdg2H7vhJFsf4aBowBzqkMOjzAGu2ZF6E=
2BA1EA778B0760C83B71DB34B686F797C2E61444E4D85E
*/
/* for test
0mqsOtVueE7uq/I5J/dAhesumWXTu619xXuRgtj4l0d0ELMH6X9ZjGqT6Lnhrhp13LVeGIgrm3QgBnk4q16BZg==
dBCzB+l/WYxqk+i54a4addy1XhiIK5t0IAZ5OKtegWY=
17FE9087DF26BF874B40820201911E37D9BE1FEEFA1A19
*/

//for test
//const ADMIN_OP_VAL_PUBKEY = "trwr09Y8sqIdg2H7vhJFsf4aBowBzqkMOjzAGu2ZF6E="
//const ADMIN_OP_FUND_PUBKEY = "dBCzB+l/WYxqk+i54a4addy1XhiIK5t0IAZ5OKtegWY="
//const ADMIN_OP_METERING_PUBKEY = "wvHG3EddBbXQHcyJal0CS/YQcNYtEbFYxejnqf9OhM4="
//const ADMIN_PUBKEY = "sxhP4F6OLZKNPQ2lG13WcHzitNX9++h56cppBDhwMlI="

//for prod
const ADMIN_OP_VAL_PUBKEY = "cGSgVIfAsXWbuWImGxJlNzfqruzuGA+4JXv5gfB0FyY="
const ADMIN_OP_FUND_PUBKEY = "sasRoTNPFzpJIHkTILaJaBnhcoC78zJk1Jy3s1/xvAE="
const ADMIN_OP_METERING_PUBKEY = "cOKct2+weTftBpTvhvFKqzg9tBkN7gG/gtFVuoE53e0="
const ADMIN_PUBKEY = "j90knB4tx3d6xi9KefyCl2FwS/hd/jpEj+cbHdzFcqM="

func prefixCertKey(key []byte) []byte {
        return append([]byte(ankrtypes.CertPrefix), key...)
}

func prefixBalanceKey(key []byte) []byte {
        return append([]byte(ankrtypes.AccountBlancePrefix), key...)
}

func prefixStakeKey(key []byte) []byte {
        return append([]byte(ankrtypes.AccountStakePrefix), key...)
}

func prefixSetMeteringKey(key []byte) []byte {
        return append([]byte(ankrtypes.MeteringPrefix), key...)
}

//-----------------------------------------

var _ types.Application = (*AnkrChainApplication)(nil)

type AnkrChainApplication struct {
	app appstore.AppStore

	// validator set
	ValUpdates []types.ValidatorUpdate

	logger log.Logger
}

func NewAnkrChainApplication(dbDir string) *AnkrChainApplication {
	appStore := appstore.NewAppStore(dbDir)

	value := []byte("")
	value = appStore.Get([]byte(ankrtypes.AccountStakePrefix))
	if value == nil || string(value) == "" {
		appStore.Set(prefixStakeKey([]byte("")), []byte("0:1"))
	}

	return &AnkrChainApplication{
		app:    appStore,
		logger: log.NewNopLogger(),
	}
}

func (app *AnkrChainApplication) SetLogger(l log.Logger) {
	app.logger = l
}

func (app *AnkrChainApplication) Info(req types.RequestInfo) types.ResponseInfo {
	res := app.app.Info(req)
	res.LastBlockHeight = app.app.Height()
	res.LastBlockAppHash = app.app.APPHash()
	return res
}

func (app *AnkrChainApplication) SetOption(req types.RequestSetOption) types.ResponseSetOption {
	return app.app.SetOption(req)
}

// tx is either "val:pubkey/power" or "key=value" or just arbitrary bytes
func (app *AnkrChainApplication) DeliverTx(tx []byte) types.ResponseDeliverTx {
	// if it starts with "val:", update the validator set
	// format is "val:pubkey/power"
	if isValidatorTx(tx) {
		// update validators in the merkle tree
		// and in app.ValUpdates
		return app.execValidatorTx(tx)
	}

	if isSetCertTx(tx) {
		return app.execSetCertTx(tx)
	}

	if isRemoveCertTx(tx) {
		return app.execRemoveCertTx(tx)
	}

	if isSetMeteringTx(tx) {
		return app.execSetMeteringTx(tx)
	}

	if isSetCertTx(tx) {
		return app.execSetCertTx(tx)
	}

	if isRemoveCertTx(tx) {
		return app.execRemoveCertTx(tx)
	}

	if isTrxSendTx(tx) {
		return app.execTrxSendTx(tx)
	}

	if isSetBalanceTx(tx) {
		return app.execSetBalanceTx(tx)
	}

	if isSetOpTx(tx) {
		return app.execSetOpTx(tx)
	}

	if isSetStakeTx(tx) {
		return app.execSetStakeTx(tx)
	}

	return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected command. Got %v", tx)}
}

func (app *AnkrChainApplication) CheckTx(tx []byte) types.ResponseCheckTx {
	if isSetCertTx(tx) {
	    tx = tx[len(ankrtypes.SetCertPrefix):]
            trxSetCertSlices := strings.SplitN(string(tx), ":", 4)
            if len(trxSetCertSlices) != 4 {
                return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Expected trx set cert. Got %v", trxSetCertSlices)}
            }
            dcS := trxSetCertSlices[0]
            pemB64S := trxSetCertSlices[1]
            nonceS := trxSetCertSlices[2]
            sigS := trxSetCertSlices[3]

            nonceInt, err_nonce := strconv.ParseInt(string(nonceS), 10, 64)
            if err_nonce != nil {
                return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected cert nonce. Got %v, %v", nonceS, err_nonce)}
            }

            nonceOldByte := app.app.Get([]byte(ankrtypes.SET_CRT_NONCE))
            nonceOld, err_nonce := strconv.ParseInt(string(nonceOldByte), 10, 64)
            if err_nonce != nil {
                if len(string(nonceOldByte)) == 0 {
                        nonceOld = 0
                } else {
                    return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected nonce5. Got %v", nonceOld)}
                }
            }

            if nonceOld + 1 != nonceInt {
                return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS)}
            }

	    var admin_pubkey_str string = ""
	    admin_pubkey := app.app.Get([]byte(ankrtypes.ADMIN_OP_METERING_PUBKEY_NAME))
	    if len(admin_pubkey) == 0 {
		    fmt.Println("use default ADMIN_OP_METERING_PUBKEY_NAME")
		    admin_pubkey_str = ADMIN_OP_METERING_PUBKEY
	    } else {
		    admin_pubkey_str = string(admin_pubkey)
	    }

            pDec, _ := base64.StdEncoding.DecodeString(sigS)
            pubKeyObject, err := deserilizePubKey(admin_pubkey_str) //set by super user
            if err != nil {
                return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Deserilize pubkey failure. Got %v", admin_pubkey_str)}
            }

            s256 := do_sha256([]byte(fmt.Sprintf("%s%s%s", dcS, pemB64S, nonceS)))
	        bb := pubKeyObject.VerifyBytes(s256[:32], pDec)
            if !bb {
                return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Bad signature. Got %v", sigS)}
            }
	    return types.ResponseCheckTx{Code: code.CodeTypeOK, GasWanted: 1}
        }

	if isRemoveCertTx(tx) {
            tx = tx[len(ankrtypes.RemoveCertPrefix):]
            trxSetCertSlices := strings.SplitN(string(tx), ":", 3)
            if len(trxSetCertSlices) != 3 {
                return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Expected trx remove cert. Got %v", trxSetCertSlices)}
            }
            dcS := trxSetCertSlices[0]
            nonceS := trxSetCertSlices[1]
            sigS := trxSetCertSlices[2]

            nonceInt, err_nonce := strconv.ParseInt(string(nonceS), 10, 64)
            if err_nonce != nil {
                return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected nonce6. Got %v", nonceS)}
            }

            nonceOldByte := app.app.Get([]byte(ankrtypes.RMV_CRT_NONCE))
            nonceOld, err_nonce := strconv.ParseInt(string(nonceOldByte), 10, 64)
            if err_nonce != nil {
                if len(string(nonceOldByte)) == 0 {
                        nonceOld = 0
                } else {
                    return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected nonce. Got %v", nonceOld)}
                }
            }

            if nonceOld + 1 != nonceInt {
                return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS)}
            }

	    var admin_pubkey_str string = ""
	    admin_pubkey := app.app.Get([]byte(ankrtypes.ADMIN_OP_METERING_PUBKEY_NAME))
	    if len(admin_pubkey) == 0 {
		    fmt.Println("use default ADMIN_OP_METERING_PUBKEY_NAME")
		    admin_pubkey_str = ADMIN_OP_METERING_PUBKEY
	    } else {
		    admin_pubkey_str = string(admin_pubkey)
	    }

            pDec, _ := base64.StdEncoding.DecodeString(sigS)
            pubKeyObject, err := deserilizePubKey(admin_pubkey_str) //set by super user
            if err != nil {
                return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Deserilize pubkey failure. Got %v", admin_pubkey_str)}
            }

            s256 := do_sha256([]byte(fmt.Sprintf("%s%s", dcS, nonceS)))
            bb := pubKeyObject.VerifyBytes(s256[:32], pDec)
            if !bb {
                return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Bad signature. Got %v", sigS)}
            }
	    return types.ResponseCheckTx{Code: code.CodeTypeOK, GasWanted: 1}
	}

	if isSetMeteringTx(tx) {
		tx = tx[len(ankrtypes.SetMeteringPrefix):]
		trxSetMeteringSlices := strings.SplitN(string(tx), ":", 6)
                if len(trxSetMeteringSlices) != 6 {
		    return types.ResponseCheckTx{
                                Code: code.CodeTypeEncodingError,
                                Log:  fmt.Sprintf("Metering incorrect format, got %s", string(tx))}
		}

                dcS := trxSetMeteringSlices[0]
                nsS := trxSetMeteringSlices[1]
		sigxS := trxSetMeteringSlices[2]
		sigyS := trxSetMeteringSlices[3]
                nonceS := trxSetMeteringSlices[4]
                valueS := trxSetMeteringSlices[5]

		nonceInt, err_nonce := strconv.ParseInt(string(nonceS), 10, 64)
                if err_nonce != nil {
                    return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected nonce. Got %v", nonceS)}
                }

                /* verify nonce */
                var nonceOld int64 = 0
                meteringRec := app.app.Get(prefixSetMeteringKey([]byte(dcS + ":" + nsS)))
                if meteringRec == nil || string(meteringRec) == "" {
                    nonceOld = 0
                } else {
                    trxSetMeteringSlices := strings.SplitN(string(meteringRec), ":", 4)
                    if len(trxSetMeteringSlices) != 4 {
                        return types.ResponseCheckTx{
                            Code: code.CodeTypeEncodingError,
                            Log:  fmt.Sprintf("Expected trx set metering check. Got %v", trxSetMeteringSlices)}
                    }

                    nonceOld, err_nonce = strconv.ParseInt(string(trxSetMeteringSlices[3]), 10, 64)
                    if err_nonce != nil {
                        return types.ResponseCheckTx{
                            Code: code.CodeTypeEncodingError,
                            Log:  fmt.Sprintf("Unexpected nonce. Got %v", nonceS)}
                    }
                }

                if nonceOld + 1 != nonceInt {
                    return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS)}
                }

		pemB64Byte := app.app.Get(prefixCertKey([]byte(dcS)))
                if len(pemB64Byte) == 0 {
                    return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("can not find cert file of %s", dcS)}
                }

                pemByte, err := base64.StdEncoding.DecodeString(string(pemB64Byte))
                if err != nil {
                    return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("pem file decoding error. Got %v", string(pemByte))}
                }
                pem := string(pemByte)

                bResult := EcdsaVerify(pem, dcS+nsS+valueS+nonceS, sigxS, sigyS)
                if !bResult {
                    return types.ResponseCheckTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("metering signature wrong. Got %v,%v", sigxS, sigyS)}
                }

		return types.ResponseCheckTx{Code: code.CodeTypeOK, GasWanted: 1}
        }

	if isValidatorTx(tx) {
		//val:public_key:power:nonce:admin_pub:sig
		tx = tx[len(ankrtypes.ValidatorSetChangePrefix):]
		pubKeyAndPower := strings.Split(string(tx), ":")
                if len(pubKeyAndPower) != 5 {
		    return types.ResponseCheckTx{
                                Code: code.CodeTypeEncodingError,
                                Log:  fmt.Sprintf("Validator incorrect format, got %s", string(tx))}
		}

		pubkeyS, powerS := pubKeyAndPower[0], pubKeyAndPower[1]
		nonceS := pubKeyAndPower[2]
		adminPubS := pubKeyAndPower[3]
		sigS := pubKeyAndPower[4]


	        var admin_pubkey_str string = ""
	        admin_pubkey := app.app.Get([]byte(ankrtypes.ADMIN_OP_VAL_PUBKEY_NAME))
	        if len(admin_pubkey) == 0 {
		    fmt.Println("use default ADMIN_OP_VAL_PUBKEY_NAME")
		    admin_pubkey_str = ADMIN_OP_VAL_PUBKEY
	        } else {
		    admin_pubkey_str = string(admin_pubkey)
	        }

		if adminPubS != admin_pubkey_str {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected pubkey. Got %v", adminPubS)}
		}

		powerInt, err := strconv.ParseInt(string(powerS), 10, 64)
		if err != nil {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected power. Got %v", powerS)}
		} else { // power < 0
			if powerInt < 0 {
				return types.ResponseCheckTx{
					Code: code.CodeTypeEncodingError,
					Log:  fmt.Sprintf("Unexpected amount, negative num. Got %v", powerS)}
			}
		}

		curValidatorCount := app.app.TotalValidatorPowers(isValidatorTx)
		if (curValidatorCount + int64(powerInt)) > tmCoreTypes.MaxTotalVotingPower {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Total powers %d will reach with the power %d", tmCoreTypes.MaxTotalVotingPower, powerInt)}
		}

		nonceInt, err_n := strconv.ParseInt(string(nonceS), 10, 64)
		if err_n != nil {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected nonce. Got %v", nonceS)}
		}

		pDec, _ := base64.StdEncoding.DecodeString(sigS)
		pubKeyObject, err_d := deserilizePubKey(adminPubS)
		if err_d != nil {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Deserilize pubkey failure. Got %v", adminPubS)}
		}

		s256 := do_sha256([]byte(fmt.Sprintf("%s%s%s", pubkeyS, powerS, nonceS)))
		bb := pubKeyObject.VerifyBytes(s256[:32], pDec)
		if !bb {
			fmt.Println("Bad signature, transaction failed.", sigS)
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Bad signature. Got %v", sigS)}
		}

		var inNonceInt int64 = 0
		inNonce := app.app.Get(([]byte(ankrtypes.SET_VAL_NONCE)))
		if len(inNonce) == 0 {
			inNonceInt = 0
		} else {
			inNonceIntValue, err_p := strconv.ParseInt(string(inNonce), 10, 64)
			if err_p != nil || inNonceInt < 0 {
				return types.ResponseCheckTx{
					Code: code.CodeTypeEncodingError,
					Log:  fmt.Sprintf("Unexpected from nonce. Got %v", inNonce)}
			}
			inNonceInt = inNonceIntValue
		}

		if (inNonceInt + 1) != nonceInt {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS)}
		}

		_, err = hex.DecodeString(pubkeyS)
		if err != nil {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Pubkey (%s) is invalid hex", pubkeyS)}
		}

		_, err = strconv.ParseInt(powerS, 10, 64)
		if err != nil {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Power (%s) is not an int", powerS)}
		}

		return types.ResponseCheckTx{Code: code.CodeTypeOK, GasWanted: 1}
	}

	if isTrxSendTx(tx) {
		tx = tx[len(ankrtypes.TrxSendPrefix):]
		trxSendSlices := strings.Split(string(tx), ":")
		if len(trxSendSlices) < 6 {
		    return types.ResponseCheckTx{
                                Code: code.CodeTypeEncodingError,
                                Log:  fmt.Sprintf("Transaction send incorecct format, got %s", string(tx))}
		}

		fromS := trxSendSlices[0]
		toS := trxSendSlices[1]
		amountS := trxSendSlices[2]
		nonceS := trxSendSlices[3]

		if len(fromS) != ankrtypes.KeyAddressLen {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected from address. Got %v", fromS)}
		}

		if len(toS) != ankrtypes.KeyAddressLen {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected to address. Got %v", toS)}
		}

		amountSend, err := new(big.Int).SetString(string(amountS), 10)
		if !err {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected amount. Got %v", amountS)}
		} else { // amountSend < 0 or less than min
			zeroN, _ := new(big.Int).SetString("0", 10)
			if amountSend.Cmp(zeroN) == -1 {
				return types.ResponseCheckTx{
					Code: code.CodeTypeEncodingError,
					Log:  fmt.Sprintf("Unexpected amount, negative num. Got %v", amountS)}
			}

			minN, _ := new(big.Int).SetString(MIN_TOKEN_SEND, 10)
			if amountSend.Cmp(minN) == -1 || amountSend.Cmp(minN) == 0 {
				return types.ResponseCheckTx{
					Code: code.CodeTypeEncodingError,
					Log:  fmt.Sprintf("Unexpected amount, not enough amount. Got %v", amountS)}
			}
		}

		nonceInt, err_nonce := strconv.ParseInt(string(nonceS), 10, 64)
		if err_nonce != nil {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected nonce. Got %v", nonceS)}
		}


		fromBalanceNonce := app.app.Get(prefixBalanceKey([]byte(fromS)))
		balanceNonceSlices := strings.Split(string(fromBalanceNonce), ":")
		var fromBalance string
		var fromNonce string
		if len(balanceNonceSlices) == 1 {
			fromBalance = balanceNonceSlices[0]
			fromNonce = "1"
		} else if len(balanceNonceSlices) == 2 {
			fromBalance = balanceNonceSlices[0]
			fromNonce = balanceNonceSlices[1]
		} else {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Expected from balance and nonce. Got %v", balanceNonceSlices)}
		}

		fromBalanceInt, err := new(big.Int).SetString(string(fromBalance), 10)
		if !err {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected amount. Got %v", fromBalance)}
		} else { // fromBalanceInt < 0
			zeroN, _ := new(big.Int).SetString("0", 10)
			if fromBalanceInt.Cmp(zeroN) == -1 {
				return types.ResponseCheckTx{
					Code: code.CodeTypeEncodingError,
					Log:  fmt.Sprintf("Unexpected amount, negative num. Got %v", fromBalance)}
			}
		}

		if fromBalanceInt.Cmp(amountSend) == -1 { // bignumber comparison
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Not enough balance to send. Balance %v, send %v", fromBalanceInt, amountSend)}
		}

		// check stake here. If from balance is less than stake, let it fail.
		cstake, _ := new(big.Int).SetString("0", 10)
		value := app.app.Get([]byte(ankrtypes.AccountStakePrefix))
		if value == nil || string(value) == "" {
			// do nothing for now
		} else {
			stakeNonceSlices := strings.Split(string(value), ":")
			cstake, err = new(big.Int).SetString(string(stakeNonceSlices[0]), 10)
			if !err {
				return types.ResponseCheckTx{
					Code: code.CodeTypeEncodingError,
					Log:  fmt.Sprintf("stake format error, %v", stakeNonceSlices[0])}
			}
		}

		if fromBalanceInt.Cmp(cstake) == -1 {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Balance <= stake. Balance %v, stake %v", fromBalanceInt, cstake)}
		}

		// check stake again.
		fromBalanceInt.Sub(fromBalanceInt, amountSend)
		if fromBalanceInt.Cmp(cstake) == -1 {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Sub Balance <= stake. Balance %v, stake %v", fromBalanceInt, cstake)}
		}


		/* check nonce */
		fromNonceInt, err_from := strconv.ParseInt(string(fromNonce), 10, 64)
		if err_from != nil || fromNonceInt < 0 {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected from nonce. Got %v", fromNonce)}
		}

		if (len(balanceNonceSlices) == 2) && (fromNonceInt + 1) != nonceInt {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS)}
		}

		return types.ResponseCheckTx{Code: code.CodeTypeOK, GasWanted: 1}
	}

	if isSetOpTx(tx) {
                tx = tx[len(ankrtypes.SetOpPrefix):]
		trxSetOpSlices := strings.Split(string(tx), ":")
		if len(trxSetOpSlices) != 5{
		    return types.ResponseCheckTx{
                                Code: code.CodeTypeEncodingError,
                                Log:  fmt.Sprintf("Set Balance incorrect format, got %d", len(tx))}
		}

		keynameS := trxSetOpSlices[0]
		valueS := trxSetOpSlices[1]
		nonceS := trxSetOpSlices[2]
		adminPubS := trxSetOpSlices[3]
		sigS := trxSetOpSlices[4]

		if adminPubS != ADMIN_PUBKEY {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected pubkey. Got %v", adminPubS)}
		}

		if keynameS != ankrtypes.ADMIN_OP_VAL_PUBKEY_NAME && keynameS != ankrtypes.ADMIN_OP_FUND_PUBKEY_NAME &&
				  keynameS != ankrtypes.ADMIN_OP_METERING_PUBKEY_NAME {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected keyname. Got %v", keynameS)}
		}

		nonceInt, err_n := strconv.ParseInt(string(nonceS), 10, 64)
		if err_n != nil {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected nonce. Got %v", nonceS)}
		}

		pDec, _ := base64.StdEncoding.DecodeString(sigS)
		pubKeyObject, err_d := deserilizePubKey(adminPubS)
		if err_d != nil {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Deserilize pubkey failure. Got %v", adminPubS)}
		}

		s256 := do_sha256([]byte(fmt.Sprintf("%s%s%s", keynameS, valueS, nonceS)))
		bb := pubKeyObject.VerifyBytes(s256[:32], pDec)
		if !bb {
			fmt.Println("Bad signature, transaction failed.", sigS)
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Bad signature. Got %v", sigS)}
			}

		var inNonce string = "0"
		inNonceByte := app.app.Get([]byte(ankrtypes.SET_OP_NONCE))
		if len(inNonceByte) != 0 {
			inNonce = string(inNonceByte)
		}

		inNonceInt, err_p:= strconv.ParseInt(string(inNonce), 10, 64)
		if err_p != nil || inNonceInt < 0 {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected from nonce. Got %v", inNonce)}
		}

		if (inNonceInt + 1) != nonceInt {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS)}
		}

		return types.ResponseCheckTx{Code: code.CodeTypeOK, GasWanted: 1}

	}

	if isSetBalanceTx(tx) {
		tx = tx[len(ankrtypes.SetBalancePrefix):]
		trxSetBalanceSlices := strings.Split(string(tx), ":")
		if len(trxSetBalanceSlices) != 5{
		    return types.ResponseCheckTx{
                                Code: code.CodeTypeEncodingError,
                                Log:  fmt.Sprintf("Set Balance incorrect format, got %d", len(tx))}
		}

		addressS := trxSetBalanceSlices[0]
		amountS := trxSetBalanceSlices[1]
		nonceS := trxSetBalanceSlices[2]
		adminPubS := trxSetBalanceSlices[3]
		sigS := trxSetBalanceSlices[4]


	        var admin_pubkey_str string = ""
	        admin_pubkey := app.app.Get([]byte(ankrtypes.ADMIN_OP_FUND_PUBKEY_NAME))
	        if len(admin_pubkey) == 0 {
		    fmt.Println("use default ADMIN_OP_FUND_PUBKEY_NAME")
		    admin_pubkey_str = ADMIN_OP_FUND_PUBKEY
	        } else {
		    admin_pubkey_str = string(admin_pubkey)
	        }

		if adminPubS != admin_pubkey_str {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected pubkey. Got %v", adminPubS)}
		}

		if len(addressS) != ankrtypes.KeyAddressLen {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected address. Got %v", addressS)}
		}

		amountSet, err := new(big.Int).SetString(string(amountS), 10)
		if !err {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected amount. Got %v", amountS)}
		} else { // amountSet < 0
			zeroN, _ := new(big.Int).SetString("0", 10)
			if amountSet.Cmp(zeroN) == -1 {
				return types.ResponseCheckTx{
					Code: code.CodeTypeEncodingError,
					Log:  fmt.Sprintf("Unexpected amount, negative num. Got %v", amountS)}
			}
		}

		nonceInt, err_n := strconv.ParseInt(string(nonceS), 10, 64)
		if err_n != nil {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected nonce. Got %v", nonceS)}
		}

		pDec, _ := base64.StdEncoding.DecodeString(sigS)
		pubKeyObject, err_d := deserilizePubKey(adminPubS)
		if err_d != nil {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Deserilize pubkey failure. Got %v", adminPubS)}
		}

		s256 := do_sha256([]byte(fmt.Sprintf("%s%s%s", addressS, amountS, nonceS)))
		bb := pubKeyObject.VerifyBytes(s256[:32], pDec)
		if !bb {
			fmt.Println("Bad signature, transaction failed.", sigS)
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Bad signature. Got %v", sigS)}
			}

		inBalanceAndNonce := app.app.Get(prefixBalanceKey([]byte(addressS)))
		balanceNonceSlices := strings.Split(string(inBalanceAndNonce), ":")
		var inBalance string
		var inNonce string
		if len(balanceNonceSlices) == 1 {
			inBalance = balanceNonceSlices[0]
			inNonce = "0"
		} else if len(balanceNonceSlices) == 2 {
			inBalance = balanceNonceSlices[0]
			inNonce = balanceNonceSlices[1]
		} else {
			inBalance = "0"
			inNonce = "0"
		}
		_ = inBalance

		inNonceInt, err_p:= strconv.ParseInt(string(inNonce), 10, 64)
		if err_p != nil || inNonceInt < 0 {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected from nonce. Got %v", inNonce)}
		}

		if (len(balanceNonceSlices) == 2) && (inNonceInt + 1) != nonceInt {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS)}
		}

		return types.ResponseCheckTx{Code: code.CodeTypeOK, GasWanted: 1}
	}

	if isSetStakeTx(tx) {
		tx = tx[len(ankrtypes.SetStakePrefix):]
		trxSetStakeSlices := strings.Split(string(tx), ":")
		if len(trxSetStakeSlices) != 4 {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Expected trx set stake. Got %v", trxSetStakeSlices)}
		}

		return types.ResponseCheckTx{Code: code.CodeTypeOK, GasWanted: 1}
	}

	return types.ResponseCheckTx{
		Code: code.CodeTypeEncodingError,
		Log:  fmt.Sprintf("Unexpected. Got %v", tx)}
}

// Commit will panic if InitChain was not called
func (app *AnkrChainApplication) Commit() types.ResponseCommit {
	return app.app.Commit()
}

func (app *AnkrChainApplication) Query(reqQuery types.RequestQuery) types.ResponseQuery {
	return app.app.Query(reqQuery)
}

// Save the validators in the merkle tree
func (app *AnkrChainApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	var initTotalPowers int64
	for _, v := range req.Validators {
		r := app.updateValidator(v)
		if r.IsErr() {
			app.logger.Error("Error updating validators", "r", r)
		}

		initTotalPowers += v.Power
	}

	if initTotalPowers > tmCoreTypes.MaxTotalVotingPower {
		app.logger.Error("The init total validator powers reach max %d", "maxtotalvalidatorpower", tmCoreTypes.MaxTotalVotingPower)
		return types.ResponseInitChain{}
	}

	sbytes := string(req.AppStateBytes)
	if len(sbytes) > 0 {
		sbytes = sbytes[1 : len(sbytes)-1]
		addressAndBalance := strings.Split(sbytes, ":")
		if len(addressAndBalance) != 2 {
			app.logger.Error("Error read app states", "appstate", sbytes)
			return types.ResponseInitChain{}
		}
		addressS, balanceS := addressAndBalance[0], addressAndBalance[1]
		fmt.Println(addressS)
		fmt.Println(balanceS)
		//app.app.state.db.Set(prefixBalanceKey([]byte(addressS)), []byte(balanceS+":1"))
		//app.app.state.Size += 1
		//app.app.Commit()
	}

	/*
		1 token = 1,000,000,000,000,000,000 
			which is 1e18 wei, as ETH.
	*/
	app.app.Set(prefixBalanceKey([]byte(INIT_ADDRESS)),
			[]byte("10000000000000000000000000000:1")) // 10000000000,000,000,000,000,000,000, 10 billion tokens

	app.app.Set([]byte(ankrtypes.SET_VAL_NONCE), []byte("0"))
	app.app.Set(prefixStakeKey([]byte("")), []byte("0:1"))
	return types.ResponseInitChain{}
}

// Track the block hash and header information
func (app *AnkrChainApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	// reset valset changes
	app.ValUpdates = make([]types.ValidatorUpdate, 0)
	return types.ResponseBeginBlock{}
}

// Update the validator set
func (app *AnkrChainApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	return types.ResponseEndBlock{ValidatorUpdates: app.ValUpdates}
}

func MakeValSetChangeTx(pubkey types.PubKey, power int64) []byte {
	return []byte(fmt.Sprintf("val:%X/%d", pubkey.Data, power))
}

func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ankrtypes.ValidatorSetChangePrefix)
}

func isTrxSendTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ankrtypes.TrxSendPrefix)
}

func isSetMeteringTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ankrtypes.SetMeteringPrefix)
}

func isSetCertTx(tx []byte) bool {
        return strings.HasPrefix(string(tx), ankrtypes.SetCertPrefix)
}

func isRemoveCertTx(tx []byte) bool {
        return strings.HasPrefix(string(tx), ankrtypes.RemoveCertPrefix)
}

/* for super-user*/
func isSetBalanceTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ankrtypes.SetBalancePrefix)
}

func isSetOpTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ankrtypes.SetOpPrefix)
}

/* for super-user*/
func isSetStakeTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ankrtypes.SetStakePrefix)
}

// format is "trx_send=from:to:amount:nonce:pubkey:sig"
// nonce should be stored in from account.
// will add signature verification when wallet code is ready 
func (app *AnkrChainApplication) execTrxSendTx(tx []byte) types.ResponseDeliverTx {
	tx = tx[len(ankrtypes.TrxSendPrefix):]
	trxSendSlices := strings.Split(string(tx), ":")
	if len(trxSendSlices) < 6 {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Expected trx send. Got %v", trxSendSlices)}
	}

	fromS := trxSendSlices[0]
	toS := trxSendSlices[1]
	amountS := trxSendSlices[2]
	nonceS := trxSendSlices[3]
	pubkeyS := trxSendSlices[4]
	sigS := trxSendSlices[5]
	//fmt.Println(fromS, toS, amountS, nonceS, pubkeyS,  sigS)

	if len(fromS) != ankrtypes.KeyAddressLen {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Unexpected from address. Got %v", fromS)}
	}

	if len(toS) != ankrtypes.KeyAddressLen {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Unexpected to address. Got %v", toS)}
	}

	amountSend, ret := new(big.Int).SetString(string(amountS), 10)
	if !ret {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected amount. Got %v", amountS)}
	} else { // amountSend < 0 or less than MIN_TOKEN_SEND
	       minN, _ := new(big.Int).SetString(MIN_TOKEN_SEND, 10)
               if amountSend.Cmp(minN) == -1 || amountSend.Cmp(minN) == 0 {
                   return types.ResponseDeliverTx{
                       Code: code.CodeTypeEncodingError,
                       Log:  fmt.Sprintf("Unexpected amount, not enough amount. Got %v", amountS)}
               }

               zeroN, _ := new(big.Int).SetString("0", 10)
               if amountSend.Cmp(zeroN) == -1 {
                       return types.ResponseDeliverTx{
                               Code: code.CodeTypeEncodingError,
                               Log:  fmt.Sprintf("Unexpected amount, negative num. Got %v", amountS)}
               }
        }

	nonceInt, err_nonce := strconv.ParseInt(string(nonceS), 10, 64)
	if err_nonce != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Unexpected nonce4. Got %v", nonceS)}
	}

	if len(pubkeyS) == 0 {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Unexpected public key. Got %v", pubkeyS)}
	}

	if len(sigS) == 0 {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Unexpected signature. Got %v", sigS)}
	}

	// ensure pubkey match fromaddress, you can't send other person's money.
	addr, err := GetAddressByPublicKey(string(pubkeyS))
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Parse address error. Got %v", pubkeyS)}
	}

	if string(fromS) != string(addr) {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("FromAddress no match with pubkey. Got %v", pubkeyS)}
	}

	pDec, _ := base64.StdEncoding.DecodeString(sigS)

	pubKeyObject, err := deserilizePubKey(pubkeyS)
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Deserilize pubkey failure. Got %v", pubkeyS)}
        }

	s256 := do_sha256([]byte(fmt.Sprintf("%s%s%s%s", fromS, toS, amountS, nonceS)))
	bb := pubKeyObject.VerifyBytes(s256[:32], pDec)
	if !bb {
		fmt.Println("Bad signature, transaction failed.", sigS)
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Bad signature. Got %v", sigS)}
	}

	fromBalanceNonce := app.app.Get(prefixBalanceKey([]byte(fromS)))
	balanceNonceSlices := strings.Split(string(fromBalanceNonce), ":")
	var fromBalance string
	var fromNonce string
	if len(balanceNonceSlices) == 1 {
		fromBalance = balanceNonceSlices[0]
		fromNonce = "1"
	} else if len(balanceNonceSlices) == 2 {
		fromBalance = balanceNonceSlices[0]
		fromNonce = balanceNonceSlices[1]
	} else {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Expected from balance and nonce. Got %v", balanceNonceSlices)}
	}

	fromBalanceInt, ret := new(big.Int).SetString(string(fromBalance), 10)
	if !ret {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected amount. Got %v", fromBalance)}
	} else { // fromBalanceInt < 0
               zeroN, _ := new(big.Int).SetString("0", 10)
               if fromBalanceInt.Cmp(zeroN) == -1 {
                       return types.ResponseDeliverTx{
                               Code: code.CodeTypeEncodingError,
                               Log:  fmt.Sprintf("Unexpected amount, negative num. Got %v", fromBalance)}
               }
        }

	if fromBalanceInt.Cmp(amountSend) == -1 { // bignumber comparison
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Not enough balance to send. Balance %v, send %v", fromBalanceInt, amountSend)}
	}
	
	/* check nonce */
	fromNonceInt, err_from := strconv.ParseInt(string(fromNonce), 10, 64)
	if err_from != nil || fromNonceInt < 0 {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Unexpected from nonce. Got %v", fromNonce)}
	}

	if (len(balanceNonceSlices) == 2) && (fromNonceInt + 1) != nonceInt {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS)}
	}

	fundRealBalance, _ := new(big.Int).SetString("0", 10)
	fundBalanceNonce := app.app.Get(prefixBalanceKey([]byte(GAS_ADDRESS)))
	var fundBalance string
	var fundNonce string = "1"
	if  fundBalanceNonce != nil {
		balanceNonceSlices = strings.Split(string(fundBalanceNonce), ":")
		if (len(balanceNonceSlices) == 1) {
			fundBalance = balanceNonceSlices[0]
			fundNonce = "1"
		} else if len(balanceNonceSlices) == 2 {
			fundBalance = balanceNonceSlices[0]
			fundNonce = balanceNonceSlices[1]
			if fundNonce == "" {
				fundNonce = "1"
			}
		} else {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Expected to balance and nonce of fund. Got %v", balanceNonceSlices)}
		}
        }

	if fundBalanceNonce != nil {
		fundBalanceInt, err := new(big.Int).SetString(string(fundBalance), 10)
		if !err {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected amount. Got %v", fundBalance)}
		} else { // toBalanceInt < 0
			zeroN, _ := new(big.Int).SetString("0", 10)
			if fundBalanceInt.Cmp(zeroN) == -1 {
				return types.ResponseDeliverTx{
					Code: code.CodeTypeEncodingError,
					Log:  fmt.Sprintf("Unexpected amount, negative num. Got %v", fundBalance)}
			}
		}

		fundRealBalance = fundBalanceInt
	}

	toRealBalance, _ := new(big.Int).SetString("0", 10)
	toBalanceNonce := app.app.Get(prefixBalanceKey([]byte(toS)))
	var toBalance string
	var toNonce string = "1"
	if  toBalanceNonce != nil {
		balanceNonceSlices = strings.Split(string(toBalanceNonce), ":")
		if (len(balanceNonceSlices) == 1) {
			toBalance = balanceNonceSlices[0]
			toNonce = "1"
		} else if len(balanceNonceSlices) == 2 {
			toBalance = balanceNonceSlices[0]
			toNonce = balanceNonceSlices[1]
			if toNonce == "" {
				toNonce = "1"
			}
		} else {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Expected to balance and nonce. Got %v", balanceNonceSlices)}
		}
        }

	if toBalanceNonce != nil {
		toBalanceInt, err := new(big.Int).SetString(string(toBalance), 10)
		if !err {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected amount. Got %v", toBalance)}
		} else { // toBalanceInt < 0
			zeroN, _ := new(big.Int).SetString("0", 10)
			if toBalanceInt.Cmp(zeroN) == -1 {
				return types.ResponseDeliverTx{
					Code: code.CodeTypeEncodingError,
					Log:  fmt.Sprintf("Unexpected amount, negative num. Got %v", toBalance)}
			}
		}

		toRealBalance = toBalanceInt
	}

	//fmt.Println(toRealBalance)

	gas, _ := new(big.Int).SetString(MIN_TOKEN_SEND, 10)
	fromBalanceInt.Sub(fromBalanceInt, amountSend)
	// 1. calculate gas based on amountSend
	// 2. actualAmountSend = (amountSend - gas)
	toRealBalance.Add(toRealBalance, amountSend.Sub(amountSend, gas))
	fundRealBalance.Add(fundRealBalance, gas)

	app.app.Set(prefixBalanceKey([]byte(fromS)), []byte(fromBalanceInt.String()+":"+nonceS))
	app.app.Set(prefixBalanceKey([]byte(toS)), []byte(toRealBalance.String()+":"+toNonce)) // use original nonce
	app.app.Set(prefixBalanceKey([]byte(GAS_ADDRESS)), []byte(fundRealBalance.String()+":"+fundNonce)) // use original nonce
	app.app.IncSize()

	tvalue := time.Now().UnixNano()
	tags := []cmn.KVPair{
		{Key: []byte("app.fromaddress"), Value: []byte(fromS)},
		{Key: []byte("app.toaddress"), Value: []byte(toS)},
		{Key: []byte("app.timestamp"), Value: []byte(strconv.FormatInt(tvalue, 10))},
		{Key: []byte("app.type"), Value: []byte("Send")},
        }

	gasUsed, _ := strconv.ParseInt(MIN_TOKEN_SEND, 0, 64)
	return types.ResponseDeliverTx{Code: code.CodeTypeOK, GasUsed: gasUsed, Tags: tags}
}


/* this function is disabled for now */
func (app *AnkrChainApplication) execSetStakeTx(tx []byte) types.ResponseDeliverTx {
        tx = tx[len(ankrtypes.SetStakePrefix):]
	trxSetStakeSlices := strings.Split(string(tx), ":")
	if len(trxSetStakeSlices) != 4 {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Expected trx set balance. Got %v", trxSetStakeSlices)}
	}

	amountS := trxSetStakeSlices[0]
	//nonceS := trxSetStakeSlices[1]

	amountSet, err := new(big.Int).SetString(string(amountS), 10)
	if !err {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Unexpected amount. Got %v", amountS)}
	} else { // amountSet < 0
		zeroN, _ := new(big.Int).SetString("0", 10)
		if amountSet.Cmp(zeroN) == -1 {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected amount, negative num. Got %v", amountS)}
		}
	}

	//app.app.state.db.Set(prefixStakeKey([]byte("")), []byte(amountS +":"+ nonceS))
	//app.app.state.Size += 1

	tags := []cmn.KVPair{
		{Key: []byte("app.type"), Value: []byte("SetStake")},
	}
	return types.ResponseDeliverTx{Code: code.CodeTypeOK,  GasUsed: 0, Tags: tags}
}

func (app *AnkrChainApplication) execSetCertTx(tx []byte) types.ResponseDeliverTx {
        tx = tx[len(ankrtypes.SetCertPrefix):]
        trxSetCertSlices := strings.SplitN(string(tx), ":", 4)
        if len(trxSetCertSlices) != 4 {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Expected trx set cert. Got %v", trxSetCertSlices)}
        }
        dcS := trxSetCertSlices[0]
        pemB64S := trxSetCertSlices[1]
        nonceS := trxSetCertSlices[2]
        sigS := trxSetCertSlices[3]

	nonceInt, err_nonce := strconv.ParseInt(string(nonceS), 10, 64)
        if err_nonce != nil {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected cert nonce. Got %v, %v", nonceS, err_nonce)}
        }

	nonceOldByte := app.app.Get([]byte(ankrtypes.SET_CRT_NONCE))
	nonceOld, err_nonce := strconv.ParseInt(string(nonceOldByte), 10, 64)
        if err_nonce != nil {
		if len(string(nonceOldByte)) == 0 {
			nonceOld = 0
		} else {
                    return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected nonce5. Got %v", nonceOld)}
		}
        }

	if nonceOld + 1 != nonceInt {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS)}
        }

	var admin_pubkey_str string = ""
	admin_pubkey := app.app.Get([]byte(ankrtypes.ADMIN_OP_METERING_PUBKEY_NAME))
	if len(admin_pubkey) == 0 {
		fmt.Println("use default ADMIN_OP_METERING_PUBKEY_NAME")
		admin_pubkey_str = ADMIN_OP_METERING_PUBKEY
	} else {
		admin_pubkey_str = string(admin_pubkey)
	}

	pDec, _ := base64.StdEncoding.DecodeString(sigS)
	pubKeyObject, err := deserilizePubKey(admin_pubkey_str) //set by super user
        if err != nil {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Deserilize pubkey failure. Got %v", admin_pubkey_str)}
        }

        s256 := do_sha256([]byte(fmt.Sprintf("%s%s%s", dcS, pemB64S, nonceS)))
        bb := pubKeyObject.VerifyBytes(s256[:32], pDec)
        if !bb {
                fmt.Println("Bad signature, transaction failed.", sigS)
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Bad signature. Got %v", sigS)}
        }

	app.app.Set(([]byte(ankrtypes.SET_CRT_NONCE)) ,[]byte(nonceS))
	app.app.Set(prefixCertKey([]byte(dcS)), []byte(pemB64S))
        app.app.IncSize()

	return types.ResponseDeliverTx{Code: code.CodeTypeOK,  GasUsed: 0}
}

/* will add signature verification when wallet code is ready */
func (app *AnkrChainApplication) execSetMeteringTx(tx []byte) types.ResponseDeliverTx {
	tx = tx[len(ankrtypes.SetMeteringPrefix):]
        trxSetMeteringSlices := strings.SplitN(string(tx), ":", 6)
        if len(trxSetMeteringSlices) != 6 {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Expected trx set metering. Got %v", trxSetMeteringSlices)}
        }
        dcS := trxSetMeteringSlices[0]
        nsS := trxSetMeteringSlices[1]
        sigxS := trxSetMeteringSlices[2]
        sigyS := trxSetMeteringSlices[3]
        nonceS := trxSetMeteringSlices[4]
        valueS := trxSetMeteringSlices[5]

        nonceInt, err_nonce := strconv.ParseInt(string(nonceS), 10, 64)
        if err_nonce != nil {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected nonce8. Got %v", nonceS)}
        }

        /* verify nonce */
        var nonceOld int64 = 0
        meteringRec := app.app.Get(prefixSetMeteringKey([]byte(dcS + ":" + nsS)))
        if meteringRec == nil || string(meteringRec) == "" {
                nonceOld = 0
        } else {
                trxSetMeteringSlices := strings.SplitN(string(meteringRec), ":", 4)
                if len(trxSetMeteringSlices) != 4 {
                    return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Expected trx set metering. Got %v", trxSetMeteringSlices)}
                }

                nonceOld, err_nonce = strconv.ParseInt(string(trxSetMeteringSlices[3]), 10, 64)
                if err_nonce != nil {
                    return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected nonce9. Got %v", nonceS)}
                }
        }

	if nonceOld + 1 != nonceInt {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS)}
        }

        /* verify sig */
	pemB64Byte := app.app.Get(prefixCertKey([]byte(dcS)))
	if len(pemB64Byte) == 0 {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("can not find cert file of %s", dcS)}
	}

	pemByte, err := base64.StdEncoding.DecodeString(string(pemB64Byte))
        if err != nil {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("pem file decoding error. Got %v", string(pemByte))}
        }
	pem := string(pemByte)

        bResult := EcdsaVerify(pem, dcS+nsS+valueS+nonceS, sigxS, sigyS)
        if !bResult {
              return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("metering signature wrong. Got %v,%v", sigxS, sigyS)}
        }

        app.app.Set(prefixSetMeteringKey([]byte(dcS + ":" + nsS)),
                []byte(valueS + ":" + sigxS + ":" + sigyS + ":" + nonceS))
        //fmt.Println(string((prefixSetMeteringKey([]byte(dcS + ":" + nsS)))))
        //fmt.Println(string([]byte(valueS + ":" + sigxS + ":" + sigyS + ":" + sigaS + ":" + sigbS + ":" + nonceS)))
        app.app.IncSize()

        tvalue := time.Now().UnixNano()
        tags := []cmn.KVPair{
                {Key: []byte("app.metering"), Value: []byte(dcS + ":" + nsS)},
                {Key: []byte("app.timestamp"), Value: []byte(strconv.FormatInt(tvalue, 10))},
                {Key: []byte("app.type"), Value: []byte("SetMetering")},
        }
        return types.ResponseDeliverTx{Code: code.CodeTypeOK,  GasUsed: 0, Tags: tags}
}

func (app *AnkrChainApplication) execRemoveCertTx(tx []byte) types.ResponseDeliverTx {
	tx = tx[len(ankrtypes.RemoveCertPrefix):]
        trxSetCertSlices := strings.SplitN(string(tx), ":", 3)
        if len(trxSetCertSlices) != 3 {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Expected trx remove cert. Got %v", trxSetCertSlices)}
        }
        dcS := trxSetCertSlices[0]
        nonceS := trxSetCertSlices[1]
        sigS := trxSetCertSlices[2]

        nonceInt, err_nonce := strconv.ParseInt(string(nonceS), 10, 64)
        if err_nonce != nil {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected nonce6. Got %v", nonceS)}
        }

        nonceOldByte := app.app.Get([]byte(ankrtypes.RMV_CRT_NONCE))
        nonceOld, err_nonce := strconv.ParseInt(string(nonceOldByte), 10, 64)
        if err_nonce != nil {
                if len(string(nonceOldByte)) == 0 {
                        nonceOld = 0
                } else {
                    return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected nonce. Got %v", nonceOld)}
                }
        }

        if nonceOld + 1 != nonceInt {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS)}
        }

        // verify sig	
	var admin_pubkey_str string = ""
	admin_pubkey := app.app.Get([]byte(ankrtypes.ADMIN_OP_METERING_PUBKEY_NAME))
	if len(admin_pubkey) == 0 {
		fmt.Println("use default ADMIN_OP_METERING_PUBKEY_NAME")
		admin_pubkey_str = ADMIN_OP_METERING_PUBKEY
	} else {
		admin_pubkey_str = string(admin_pubkey)
	}

	pDec, _ := base64.StdEncoding.DecodeString(sigS)
	pubKeyObject, err := deserilizePubKey(admin_pubkey_str) //set by super user
        if err != nil {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Deserilize pubkey failure. Got %v", admin_pubkey_str)}
        }

        s256 := do_sha256([]byte(fmt.Sprintf("%s%s", dcS, nonceS)))
        bb := pubKeyObject.VerifyBytes(s256[:32], pDec)
        if !bb {
                fmt.Println("Bad signature, transaction failed.", sigS)
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Bad signature. Got %v", sigS)}
        }

	app.app.Set(([]byte(ankrtypes.RMV_CRT_NONCE)), []byte(nonceS))
	app.app.Delete(prefixCertKey([]byte(dcS)))
	app.app.IncSize()

	return types.ResponseDeliverTx{Code: code.CodeTypeOK, GasUsed: 0}
}

func (app *AnkrChainApplication) execSetOpTx(tx []byte) types.ResponseDeliverTx {
        tx = tx[len(ankrtypes.SetOpPrefix):]
	trxSetOpSlices := strings.Split(string(tx), ":")
	if len(trxSetOpSlices) != 5{
	    return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Set Balance incorrect format, got %d", len(tx))}
	}

	keynameS := trxSetOpSlices[0]
	valueS := trxSetOpSlices[1]
	nonceS := trxSetOpSlices[2]
	adminPubS := trxSetOpSlices[3]
	sigS := trxSetOpSlices[4]

	if adminPubS != ADMIN_PUBKEY {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Unexpected pubkey. Got %v", adminPubS)}
	}

	if keynameS != ankrtypes.ADMIN_OP_VAL_PUBKEY_NAME && keynameS != ankrtypes.ADMIN_OP_FUND_PUBKEY_NAME &&
			  keynameS != ankrtypes.ADMIN_OP_METERING_PUBKEY_NAME {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Unexpected keyname. Got %v", keynameS)}
	}

	nonceInt, err_n := strconv.ParseInt(string(nonceS), 10, 64)
	if err_n != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Unexpected nonce. Got %v", nonceS)}
	}

	pDec, _ := base64.StdEncoding.DecodeString(sigS)
	pubKeyObject, err_d := deserilizePubKey(adminPubS)
	if err_d != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Deserilize pubkey failure. Got %v", adminPubS)}
	}

	s256 := do_sha256([]byte(fmt.Sprintf("%s%s%s", keynameS, valueS, nonceS)))
	bb := pubKeyObject.VerifyBytes(s256[:32], pDec)
	if !bb {
		fmt.Println("Bad signature, transaction failed.", sigS)
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Bad signature. Got %v", sigS)}
	}

	var inNonce string = "0"
	inNonceByte := app.app.Get([]byte(ankrtypes.SET_OP_NONCE))
	if len(inNonceByte) != 0 {
		inNonce = string(inNonceByte)
	}

	inNonceInt, err_p:= strconv.ParseInt(string(inNonce), 10, 64)
	if err_p != nil || inNonceInt < 0 {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Unexpected from nonce. Got %v", inNonce)}
	}

	if (inNonceInt + 1) != nonceInt {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS)}
	}

	app.app.Set([]byte(keynameS), []byte(valueS))
	app.app.Set([]byte(ankrtypes.SET_OP_NONCE), []byte(nonceS))
    app.app.IncSize()

	return types.ResponseDeliverTx{Code: code.CodeTypeOK, GasWanted: 1, GasUsed: 0}
}
/* will add signature verification when wallet code is ready */
func (app *AnkrChainApplication) execSetBalanceTx(tx []byte) types.ResponseDeliverTx {
	tx = tx[len(ankrtypes.SetBalancePrefix):]
	trxSetBalanceSlices := strings.Split(string(tx), ":")
	if len(trxSetBalanceSlices) != 5 {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Expected trx set balance. Got %v", trxSetBalanceSlices)}
	}
	addressS := trxSetBalanceSlices[0]
	amountS := trxSetBalanceSlices[1]
	nonceS := trxSetBalanceSlices[2]
	adminPubS := trxSetBalanceSlices[3]
	sigS := trxSetBalanceSlices[4]

	var admin_pubkey_str string = ""
	admin_pubkey := app.app.Get([]byte(ankrtypes.ADMIN_OP_FUND_PUBKEY_NAME))
	if len(admin_pubkey) == 0 {
		fmt.Println("use default ADMIN_OP_FUND_PUBKEY_NAME")
		admin_pubkey_str = ADMIN_OP_FUND_PUBKEY
	} else {
		admin_pubkey_str = string(admin_pubkey)
	}

	if adminPubS != admin_pubkey_str {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Unexpected pubkey. Got %v", adminPubS)}
	}

	if len(addressS) != ankrtypes.KeyAddressLen {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Unexpected address. Got %v", addressS)}
	}

	amountSet, err := new(big.Int).SetString(string(amountS), 10)
	if !err {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Unexpected amount. Got %v", amountS)}
	} else { // amountSet < 0
		zeroN, _ := new(big.Int).SetString("0", 10)
		if amountSet.Cmp(zeroN) == -1 {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected amount, negative num. Got %v", amountS)}
		}
	}

	nonceInt, err_n := strconv.ParseInt(string(nonceS), 10, 64)
        if err_n != nil {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected nonce. Got %v", nonceS)}
        }

        pDec, _ := base64.StdEncoding.DecodeString(sigS)
        pubKeyObject, err_d := deserilizePubKey(adminPubS)
        if err_d != nil {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Deserilize pubkey failure. Got %v", adminPubS)}
        }

	s256 := do_sha256([]byte(fmt.Sprintf("%s%s%s", addressS, amountS, nonceS)))
        bb := pubKeyObject.VerifyBytes(s256[:32], pDec)
        if !bb {
                fmt.Println("Bad signature, transaction failed.", sigS)
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Bad signature. Got %v", sigS)}
        }

        inBalanceAndNonce := app.app.Get(prefixBalanceKey([]byte(addressS)))
        balanceNonceSlices := strings.Split(string(inBalanceAndNonce), ":")
        var inBalance string
        var inNonce string
        if len(balanceNonceSlices) == 1 {
                inBalance = balanceNonceSlices[0]
                inNonce = "0"
        } else if len(balanceNonceSlices) == 2 {
                inBalance = balanceNonceSlices[0]
                inNonce = balanceNonceSlices[1]
        } else {
                inBalance = "0"
                inNonce = "0"
        }
	_ = inBalance

	inNonceInt, err_p:= strconv.ParseInt(string(inNonce), 10, 64)
        if err_p != nil || inNonceInt < 0 {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected from nonce. Got %v", inNonce)}
        }

        if (len(balanceNonceSlices) == 2) && (inNonceInt + 1) != nonceInt {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS)}
        }

	app.app.Set(prefixBalanceKey([]byte(addressS)), []byte(amountS + ":" + nonceS))
	app.app.IncSize()

	tags := []cmn.KVPair{
		{Key: []byte("app.type"), Value: []byte("SetBalance")},
	}
	return types.ResponseDeliverTx{Code: code.CodeTypeOK,  GasUsed: 0, Tags: tags}
}

func (app *AnkrChainApplication) execValidatorTx(tx []byte) types.ResponseDeliverTx {
	tx = tx[len(ankrtypes.ValidatorSetChangePrefix):]

	//get the pubkey and power
	pubKeyAndPower := strings.Split(string(tx), ":")
	if len(pubKeyAndPower) != 5 {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Expected 'pubkey/power'. Got %v", pubKeyAndPower)}
	}
	pubkeyS, powerS := pubKeyAndPower[0], pubKeyAndPower[1]
	nonceS := pubKeyAndPower[2]
        adminPubS := pubKeyAndPower[3]
        sigS := pubKeyAndPower[4]

	var admin_pubkey_str string = ""
	admin_pubkey := app.app.Get([]byte(ankrtypes.ADMIN_OP_VAL_PUBKEY_NAME))
	if len(admin_pubkey) == 0 {
		fmt.Println("use default ADMIN_OP_VAL_PUBKEY_NAME")
		admin_pubkey_str = ADMIN_OP_VAL_PUBKEY
	} else {
		admin_pubkey_str = string(admin_pubkey)
	}

        if adminPubS != admin_pubkey_str {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected pubkey. Got %v", adminPubS)}
        }

        powerInt, err := strconv.ParseInt(string(powerS), 10, 64)
        if err != nil {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected power. Got %v", powerS)}
        } else { // power < 0
                if powerInt < 0 {
                        return types.ResponseDeliverTx{
                                Code: code.CodeTypeEncodingError,
                                Log:  fmt.Sprintf("Unexpected amount, negative num. Got %v", powerS)}
                }
        }

        nonceInt, err_n := strconv.ParseInt(string(nonceS), 10, 64)
        if err_n != nil {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Unexpected nonce. Got %v", nonceS)}
        }

        pDec, _ := base64.StdEncoding.DecodeString(sigS)
        pubKeyObject, err_d := deserilizePubKey(adminPubS)
        if err_d != nil {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("Deserilize pubkey failure. Got %v", adminPubS)}
        }

        s256 := do_sha256([]byte(fmt.Sprintf("%s%s%s", pubkeyS, powerS, nonceS)))
        bb := pubKeyObject.VerifyBytes(s256[:32], pDec)
        if !bb {
                fmt.Println("Bad signature, transaction failed.", sigS)
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("DeliverTx, Bad signature. Got %v", sigS)}
        }

	var inNonceInt int64 = 0
	inNonce := app.app.Get(([]byte(ankrtypes.SET_VAL_NONCE)))
	if len(inNonce) == 0 {
		inNonceInt = 0
	} else {
		inNonceIntValue, err_p:= strconv.ParseInt(string(inNonce), 10, 64)
		if err_p != nil || inNonceInt < 0 {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Unexpected from nonce. Got %v", inNonce)}
		}
		inNonceInt = inNonceIntValue
	}

	if (inNonceInt + 1) != nonceInt {
                return types.ResponseDeliverTx{
                        Code: code.CodeTypeEncodingError,
                        Log:  fmt.Sprintf("nonce should be one more than last nonce. Got %v", nonceS)}
        }

	// decode the pubkey
	pubkey, err := hex.DecodeString(pubkeyS)
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Pubkey (%s) is invalid hex", pubkeyS)}
	}

	// decode the power
	power, err := strconv.ParseInt(powerS, 10, 64)
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Power (%s) is not an int", powerS)}
	}

	// update
	app.app.Set([]byte(ankrtypes.SET_VAL_NONCE), []byte(nonceS))
	return app.updateValidator(types.Ed25519ValidatorUpdate(pubkey, int64(power)))
}

// add, update, or remove a validator
func (app *AnkrChainApplication) updateValidator(v types.ValidatorUpdate) types.ResponseDeliverTx {
	key := []byte("val:" + string(v.PubKey.Data))
	if v.Power == 0 {
		// remove validator
		if !app.app.Has(key) {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeUnauthorized,
				Log:  fmt.Sprintf("Cannot remove non-existent validator %X", key)}
		}
		app.app.Delete(key)
	} else {
		// add or update validator
		value := bytes.NewBuffer(make([]byte, 0))
		if err := types.WriteMessage(&v, value); err != nil {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Error encoding validator: %v", err)}
		}
		app.app.Set(key, value.Bytes())
	}

	// we only update the changes array if we successfully updated the tree
	app.ValUpdates = append(app.ValUpdates, v)

	tags := []cmn.KVPair{
		{Key: []byte("app.type"), Value: []byte("UpdateValidator")},
	}
	return types.ResponseDeliverTx{Code: code.CodeTypeOK, GasUsed: 0, Tags: tags}
}

func deserilizePubKey(pub_key_b64 string) (ed25519.PubKeyEd25519, error) {
        pDec, err := base64.StdEncoding.DecodeString(pub_key_b64)
        if err != nil {
                return ed25519.PubKeyEd25519{}, err
        }

        pk := []byte(pDec)
        var pubObject ed25519.PubKeyEd25519 = ed25519.PubKeyEd25519{pk[0], pk[1], pk[2], pk[3],pk[4], pk[5],pk[6],
                pk[7],pk[8], pk[9], pk[10], pk[11], pk[12], pk[13], pk[14], pk[15], pk[16], pk[17], pk[18], pk[19],
                pk[20], pk[21],pk[22], pk[23],pk[24], pk[25],pk[26], pk[27],pk[28], pk[29],pk[30], pk[PubKeyEd25519Size - 1]}

        return pubObject, nil
}

func GetAddressByPublicKey(pub_key string) (string, error) {
        pubKeyObject, err := deserilizePubKey(pub_key)
        if err != nil {
                return "", err
        }

        address := fmt.Sprintf("%s", pubKeyObject.Address())
        return address, nil
}

func do_sha256(input []byte) [32]byte{
        sum := sha256.Sum256([]byte(input))
        return sum
}

/*
use pem format string ECDSA public_key to verify the input's signature.
*/
func EcdsaVerify(pubpem, input, signature1, signature2 string) bool{

        rSig := new(big.Int)
        rSig, ok := rSig.SetString(signature1, 10)
        if !ok {
                fmt.Println("SetString: error")
                return false
        }

        sSig := new(big.Int)
        sSig, ok = sSig.SetString(signature2, 10)
        if !ok {
                fmt.Println("SetString: error")
                return false
        }

        ecPublicKey, err := parseEcdsaPublicKeyFromPemStr(pubpem)
        if (err != nil) {
                fmt.Println(err)
                return false
        }

        sum := sha256.Sum256([]byte(input))
        valid := ecdsa.Verify(ecPublicKey, sum[:32], rSig, sSig)

        return valid
}

func parseEcdsaPublicKeyFromPemStr(pubPEM string) (*ecdsa.PublicKey, error) {
        block, _ := pem.Decode([]byte(pubPEM))
        if block == nil {
                fmt.Println("failed to parse certificate PEM")
                return nil, errors.New("failed to parse PEM block containing the cert")
        }
        cert, err := x509.ParseCertificate(block.Bytes)
        if err != nil {
                fmt.Println("failed to parse certificate: ", err.Error())
                return nil, err
        }

        pub := cert.PublicKey.(*ecdsa.PublicKey)

        return pub, nil
}
