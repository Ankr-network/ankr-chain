package handler

import (
	"encoding/base64"
	_ "encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/Ankr-network/ankr-chain/client"
	ankrcmm "github.com/Ankr-network/ankr-chain/common"
	lascmm "github.com/Ankr-network/ankr-chain/service/las/common"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	_ "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func QueryAccountInfoHandler(c *client.Client) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		reqVars := mux.Vars(req)
		addr := reqVars["address"]
		queryData := &ankrcmm.AccountQueryReq{addr}
		respData := &ankrcmm.AccountQueryResp{}

		isNeedVerify := viper.GetBool("proof-verify")
		err := c.QueryWithOption("/store/account", 0, isNeedVerify, viper.GetString(lascmm.FlagHome), queryData, respData)
		if err != nil {
			WriteErrorResponse(resp, http.StatusInternalServerError, err.Error())
			return
		}

		respJson, err := Cdc.MarshalJSON(respData)
		if err != nil {
			WriteErrorResponse(resp, http.StatusInternalServerError, err.Error())
			return
		}
		resp.Header().Set("Content-Type", "application/json")
		resp.Write(respJson)
		return
	}
}

func QueryAccountNonceHandler(c *client.Client) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		reqVars := mux.Vars(req)
		addr := reqVars["address"]
		queryData := &ankrcmm.NonceQueryReq{addr}
		respData := &ankrcmm.NonceQueryResp{}

		isNeedVerify := viper.GetBool("proof-verify")
		err := c.QueryWithOption("/store/nonce", 0, isNeedVerify, viper.GetString(lascmm.FlagHome), queryData, respData)
		if err != nil {
			WriteErrorResponse(resp, http.StatusInternalServerError, err.Error())
			return
		}

		respJson, err := Cdc.MarshalJSON(respData)
		if err != nil {
			WriteErrorResponse(resp, http.StatusInternalServerError, err.Error())
			return
		}
		resp.Header().Set("Content-Type", "application/json")
		resp.Write(respJson)
		return
	}
}

func QueryAccountBalanceHandler(c *client.Client) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			WriteErrorResponse(resp, http.StatusBadRequest, err.Error())
			return
		}

		if req.Form == nil || len(req.Form) == 0 {
			WriteErrorResponse(resp, http.StatusBadRequest, "QueryAccountBalanceHandler, form val's len is 0 or nil")
			return
		}

		addr   := req.FormValue("address")
		symbol := req.FormValue("symbol")
		if addr == "" {
			WriteErrorResponse(resp, http.StatusBadRequest, fmt.Sprintf("blank request param address"))
			return
		}
		if symbol == "" {
			WriteErrorResponse(resp, http.StatusBadRequest, fmt.Sprintf("blank request param symbol"))
			return
		}

		queryData := &ankrcmm.BalanceQueryReq{addr, symbol}
		respData := &ankrcmm.BalanceQueryResp{}
		isNeedVerify := viper.GetBool("proof-verify")
		err = c.QueryWithOption("/store/balance", 0, isNeedVerify, viper.GetString(lascmm.FlagHome), queryData, respData)
		if err != nil {
			WriteErrorResponse(resp, http.StatusInternalServerError, err.Error())
			return
		}

		balBig, isSuccess := new(big.Int).SetString(respData.Amount, 10)
		if !isSuccess {
			WriteErrorResponse(resp, http.StatusInternalServerError, fmt.Sprintf("invalid balance value, addr=%s, symbol=%s", addr, symbol))
			return
		}

		respData.Amount = convertToFloat64strFromDevimalVal(balBig)

		respJson, err := Cdc.MarshalJSON(respData)
		if err != nil {
			WriteErrorResponse(resp, http.StatusInternalServerError, err.Error())
			return
		}
		resp.Header().Set("Content-Type", "application/json")
		resp.Write(respJson)
		return
	}
}

func GenerateAccounts(c *client.Client) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		cnt := int64(1)
		reqVars := mux.Vars(req)
		if reqVars != nil {
			if cntString, ok := reqVars["count"]; ok {
				cntT, err := strconv.ParseInt(cntString, 10, 64)
				if err != nil {
					WriteErrorResponse(resp, http.StatusBadRequest, err.Error())
					return
				}
				cnt = cntT
			}
		}

		var respData []*AccountInfoItem
		for i:= int64(0); i < cnt; i++ {
			priKeyEd25519 := ed25519.GenPrivKey()

			priKeyArray := [64]byte(priKeyEd25519)
			priKeyBytes := priKeyArray[:]
			priKey      := base64.StdEncoding.EncodeToString(priKeyBytes)

			pubKeyBytes := priKeyEd25519.PubKey().Bytes()
			pubKey      := fmt.Sprintf("%X", pubKeyBytes[:])

			address := priKeyEd25519.PubKey().Address().String()

			accountInfo := &AccountInfoItem{
				priKey,
				pubKey,
				address,
			}

			respData = append(respData, accountInfo)
		}

		respJson, err := Cdc.MarshalJSON(respData)
		if err != nil {
			WriteErrorResponse(resp, http.StatusInternalServerError, err.Error())
			return
		}
		resp.Header().Set("Content-Type", "application/json")
		resp.Write(respJson)

		return
	}
}