package cmd

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	client2 "github.com/Ankr-network/ankr-chain/client"
	common2 "github.com/Ankr-network/ankr-chain/common"
	common3 "github.com/Ankr-network/ankr-chain/tx/common"
	"github.com/Ankr-network/ankr-chain/tx/serializer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/common"
	core_types "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	"regexp"
	"strconv"
	"strings"
)

const (
	defaultPerPage = 20
	maxPerPage     = 50
)

var(
	//flags used by query sub commands
	//persistent flag
	queryUrl = "queryUrl"

	//bind transaction flags
	trxTxid = "trxTxid"
	trxApprove = "trxApprove"
	trxPage = "trxPage"
	trxPerPage = "trxPerPage"
	trxMetering = "trxMetering"
	trxTimeStamp = "trxTimeStamp"
	trxType = "trxType"
	trxHeight = "trxHeight"
	trxCreator = "trxCreator"
	trxFrom = "trxFrom"
	trxTo = "trxTo"
	trxDetail = "trxDetail"

	//bind block flags
	blockHeight = "blockHeight"
	blockPage = "blockPage"
	blockPerPage ="blockPerPage"
	blockTxTransferOnly = "blockTxTransferOnly"
	validatorHeight = "validatorHeight"
	unconfirmedTxLimit = "unconfirmedTxLimit"


	//transaction prefix
    TxPrefix = "trx_send="
    setMeteringPrefix = "set_mtr="
    setBalancePrefix = "set_bal="
    setStakePrefix = "set_stk="
    setCertPrefix = "set_crt="
    removeCertPrefix = "rmv_crt="
    setValidatorPrefix = "val:"

	querySymbol     = "querySymbol"
	queryAddress    = "queryAddress"
	queryNonceAddr = "queryNonceAddr"
	queryCurrencySymbol = "queryCurrencySymbol"
	queryAccAddress = "queryAccAddress"
)

var (
	txSearchFlags = []string{meteringParam, timeStampParam, typeParam, fromParam, toParam, heightParam, creatorParam}
	periodRegexp = `((\(|\[)\d\:\d+(\]|\()|\d+)`
	reg, _ = regexp.Compile(periodRegexp)
)
// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "query information from ankr chain",
}

func init() {
	// init functions
	serializer.NewTxSerializerCDC()
	err := addPersistentString(queryCmd, queryUrl, urlParam, "", "", "validator url", required)
	if err != nil {
		panic(err)
	}
	appendSubCmd(queryCmd, "transaction","transaction allows you to query the transaction results with multiple conditions.", transactionInfo, addTransactionInfoFlags)
	appendSubCmd(queryCmd, "block", "Get block at a given height. If no height is provided, it will fetch the latest block. And you can use \"detail\" to show more information about transactions contained in block",
		queryBlock, addQueryBlockFlags)
	appendSubCmd(queryCmd, "validators", "Get the validator set at the given block height. If no height is provided, it will fetch the current validator set.",
		queryValidator, addQueryValidatorFlags)
	appendSubCmd(queryCmd, "status", "Get Ankr status including node info, pubkey, latest block hash, app hash, block height and time.",
		queryStatus, nil)
	appendSubCmd(queryCmd, "genesis", "Get genesis file.", queryGenesis, nil)
	appendSubCmd(queryCmd, "consensusstate", "ConsensusState returns a concise summary of the consensus state", queryConsensusState, nil)
	appendSubCmd(queryCmd, "dumpconsensusstate", "dumps consensus state", queryDumpConsensusState, nil)
	appendSubCmd(queryCmd, "unconfirmedtxs", "Get unconfirmed transactions (maximum ?limit entries) including their number",
		queryUnconfirmedTxs, addQueryUncofirmedTxsFlags)
	appendSubCmd(queryCmd, "numunconfirmedtxs","Get number of unconfirmed transactions.", queryNumUnconfiredTxs, nil)
	appendSubCmd(queryCmd, "contract", "get smart contract data", runGetContract, addGetContractFlags)
	appendSubCmd(queryCmd, "account", "query account info",queryAccount, addQueryAccountFlags)
	appendSubCmd(queryCmd, "balance", "get the balance of an address.", getBalance, addGetBalanceFlags)
	appendSubCmd(queryCmd, "nonce", "get the nonce of an address.", getNonce, addGetNonceFlags)
	appendSubCmd(queryCmd, "currency", "get currency info of a contract.", getCurrency, addQueryCurrencyFlags)
}

func transactionInfo(cmd *cobra.Command, args []string)  {
	client := newAnkrHttpClient(viper.GetString(queryUrl))
	prove := viper.GetBool(trxApprove)

	// query --txid 0xxxxxx --nodeurl url
	if cmd.Flag(txidParam).Changed {
		txid := viper.GetString(trxTxid)
		txid = strings.TrimLeft(txid,"0x")
		txidByte, err := hex.DecodeString(txid)
		if err != nil {
			fmt.Println("Invalid txid.")
			return
		}
		resp, err := client.Tx(txidByte, prove)
		detail := viper.GetBool(trxDetail)
		displayTxMsg(resp, detail)
		return
	}

	//collectedFlags := make([]string, 0, numFlags)
	collectedFlags := make(map[string] string)
	for _, flag := range txSearchFlags {
		if cmd.Flag(flag).Changed {
			collectedFlags[flag] = cmd.Flag(flag).Value.String()
		}
	}
	//query transaction --txid hash --nodeurl https://xx:xx --prove bool

	//query transaction --type/from/to/metering/timestamp
	query := formatQueryContent(collectedFlags)
	page := viper.GetInt(trxPage)
	perPage := viper.GetInt(trxPerPage)
	resp, err := client.TxSearch(query, prove, page, perPage)
	if err != nil {
		fmt.Println("Transaction search failed.")
		fmt.Println(err)
		return
	}
	detail := viper.GetBool(trxDetail)
	fmt.Println("Total Tx Count:", resp.TotalCount)
	fmt.Println("Transactions search result:")
	for _, tx := range resp.Txs {
		displayTxMsg(tx, detail)
	}
}

func displayTxMsg(txMsg *core_types.ResultTx, detail bool)  {
	if detail{
		displayTx(txMsg.Tx)
	}else {
		displayStruct(txMsg)
	}
}

func displayTx(data []byte)  {
	decoder := client2.NewTxDecoder()
	tx, err  := decoder.Decode(data)
	if err != nil {
		fmt.Println("Decode transaction error!")
		fmt.Println(err)
		return
	}
	if viper.GetBool(blockTxTransferOnly) && tx.Type() != common3.TxMsgTypeTransfer {
		return
	}
	decodeAndDisplay(tx)
}

func formatQueryContent(parameters map[string]string) string {
	result := make([]string, 0, len(parameters))
	var query string
	for key, value := range parameters {
		switch key {
		case meteringParam:
			query = fmt.Sprintf("app.metering='%s'",value)
		case timeStampParam:
			valueSlice := strings.Split(value, ":")
			if len(valueSlice) == 1 {
				//if only one digit is received in interval, trim bracket
				value = strings.TrimPrefix(value,"[")
				value = strings.TrimPrefix(value,"(")
				value = strings.TrimRight(value,"]")
				value = strings.TrimRight(value,")")
				query = fmt.Sprintf("app.timestamp=%s",value)
				break
			}
			interval := formatInterval(value)
			if len(interval) != 2{
				query = fmt.Sprintf("app.timestamp%s", interval[0])
				break
			}
			query = fmt.Sprintf("app.timestamp%s and app.timestamp%s", interval[0], interval[1])
		case typeParam:
			query = fmt.Sprintf("app.type='%s'",value)
		case fromParam:
			query = fmt.Sprintf("app.fromaddress='%s'",value)
		case toParam:
			query = fmt.Sprintf("app.toaddress='%s'",value)
		case creatorParam:
			query = fmt.Sprintf("app.creator='%s'",value)
		case heightParam:
			valueSlice := strings.Split(value, ":")
			if len(valueSlice) == 1 {
				value = strings.TrimPrefix(value,"[")
				value = strings.TrimPrefix(value,"(")
				value = strings.TrimRight(value,"]")
				value = strings.TrimRight(value,")")
				query = fmt.Sprintf("tx.height=%s",value)
				break
			}
			interval := formatInterval(value)
			if len(interval) != 2{
				query = fmt.Sprintf("tx.height%s", interval[0])
				break
			}
			query = fmt.Sprintf("tx.height%s and tx.height%s", interval[0], interval[1])
		}
		result =append(result, query)
	}
	return strings.Join(result, " and ")
}
func formatInterval(period string) []string {
	periodSlice := strings.Split(period, ":")
	leftOp := []rune(periodSlice[0])[0]
	length := len(periodSlice[1])
	rightOp := []rune(periodSlice[1])[length-1]
	var leftValue, rightValue string
	switch leftOp {
	case '(':
		leftValue = strings.TrimLeft(periodSlice[0],"(")
		if len(leftValue) > 0 {
			leftValue = fmt.Sprintf(">%s",string(leftValue))
		}
	case '[':
		leftValue = strings.TrimLeft(periodSlice[0],"[")
		if len(leftValue) > 0 {
			leftValue = fmt.Sprintf(">%s",string(leftValue))
		}
	}

	switch rightOp {
	case ')':
		rightValue = strings.TrimRight(periodSlice[1],")")
		if len(rightValue) > 0{
			rightValue = fmt.Sprintf("<%s",rightValue)
		}
	case ']':
		rightValue = strings.TrimRight(periodSlice[1],"]")
		if len(rightValue) > 0{
			rightValue = fmt.Sprintf("<=%s",rightValue)
		}
	}
	result := make([]string, 0 , 2)
	if leftValue != ""{
		result = append(result, leftValue)
	}
	if rightValue != ""{
		result = append(result, rightValue)
	}
	return result
}

func addTransactionInfoFlags(cmd *cobra.Command)  {
	err := addStringFlag(cmd, trxTxid, txidParam, "", "", "The transaction hash", notRequired)
	if err != nil {
		panic(err)
	}
	err = addBoolFlag(cmd, trxApprove, approveParam, "", false, "Include a proof of the transaction inclusion in the block", notRequired)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, trxFrom, fromParam, "", "", "the from address contained in a transaction", notRequired)
	if err != nil {
		panic(err)
	}

	err = addStringFlag(cmd, trxTo, toParam, "", "", "the to address contained in a transaction", notRequired)
	if err != nil {
		panic(err)
	}

	err = addStringFlag(cmd, trxTimeStamp, timeStampParam, "", "",
		"transaction executed timestamp. Input can be an exactly unix timestamp  or a time interval separate by \":\", and time interval should be enclosed with \"[]\" or \"()\" which is mathematically open interval and close interval." ,
		notRequired)
	if err != nil {
		panic(err)
	}

	err = addStringFlag(cmd, trxMetering, meteringParam, "", "", "query metering transaction, both datacenter name and namespace should be  provided and separated  by \":\"", notRequired)
	if err != nil {
		panic(err)
	}

	err = addIntFlag(cmd, trxPage, pageParam, "", 1, "Page number (1 based)", notRequired)
	if err != nil {
		panic(err)
	}

	err = addIntFlag(cmd, trxPerPage, perPageParam, "", 30, "Number of entries per page(max: 100)", notRequired)
	if err != nil {
		panic(err)
	}

	err = addStringFlag(cmd, trxType, typeParam, "", "", "Ankr chain predefined types, SetMetering, SetBalance, UpdatValidator, SetStake, Send", notRequired)
	if err != nil {
		panic(err)
	}

	err = addStringFlag(cmd, trxHeight, heightParam, "", "",
		"block height. Input can be an exactly block height  or a height interval separate by \":\", and height interval should be enclosed with \"[]\" or \"()\" which is mathematically open interval and close interval.", notRequired)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, trxCreator, creatorParam, "", "", "app creator", notRequired)
	if err != nil {
		panic(err)
	}
	err = addBoolFlag(cmd, trxDetail, detailParam, "", false, "display transaction detail", notRequired)
	if err != nil {
		panic(err)
	}
}

//query block
func queryBlock(cmd *cobra.Command, args []string)  {
	validatorUrl = viper.GetString(queryUrl)
	if len(validatorUrl) < 1 {
		fmt.Println("Illegal url is received!")
		return
	}
	cl := newAnkrHttpClient(viper.GetString(queryUrl))
	from, to, err := getBlockInterval()
	if err != nil {
		fmt.Println(err)
		return
	}

	resps := make([]*core_types.ResultBlock,0, to - from + 1)
	for iter := from; iter <= to; iter ++ {
		heightInt := int64(iter)
		heightP := &heightInt
		if heightInt == -1 {
			heightP = nil
		}
		resp, err :=cl.Block(heightP)
		if err != nil {
			fmt.Println("Query block failed.", err)
			return
		}
		resps = append(resps, resp)
	}

	detail := false
	if len(args) > 0 && args[0] == "detail"{
		detail = true
	}
	page := viper.GetInt(blockPage)
	perPage := viper.GetInt(trxPerPage)
	outPutBlockResp(resps, page, perPage, detail)
}

func validatePage(page, perPage, totalCount int) int {
	if perPage < 1 {
		return 1
	}

	pages := ((totalCount - 1) / perPage) + 1
	if page < 1 {
		page = 1
	} else if page > pages {
		page = pages
	}

	return page
}

func validatePerPage(perPage int) int {
	if perPage < 1 {
		return defaultPerPage
	} else if perPage > maxPerPage {
		return maxPerPage
	}
	return perPage
}

func validateSkipCount(page, perPage int) int {
	skipCount := (page - 1) * perPage
	if skipCount < 0 {
		return 0
	}

	return skipCount
}

func getBlockInterval() (from int, to int,err  error) {
	from = -1
	to = -1
	heightStr := viper.GetString(blockHeight)

	//height flag is not set, get the latest block
	if heightStr == ""{
		return from, to, nil
	}

	//if height flags is not set properly, return error
	if matched := reg.MatchString(heightStr); !matched {
		return from, to , errors.New("Invalid Height format, should be \"[from:to]\". ")
	}

	//strictly flow the rule [from:to]
	heightStr = strings.TrimLeft(heightStr, "[")
	heightStr = strings.TrimRight(heightStr, "]")
	height, err := strconv.Atoi(heightStr)
	if err == nil {
		from = height
		to = height
		return from, to, nil
	}
	heightSlice := strings.Split(heightStr, ":")
	if len(heightSlice) != 2 {
		return from, to, errors.New("input both from and to separated with \":\". ")
	}
	fromStr, toStr := heightSlice[0],heightSlice[1]
	from, err = strconv.Atoi(fromStr)
	if err != nil {
		return from, to, errors.New("from is not an integer. ")
	}
	to, err = strconv.Atoi(toStr)
	if err != nil {
		return from, to, errors.New("to is not an integer. ")
	}
	if from >to {
		return from, to, errors.New("from should be less or equal than to")
	}
	return from, to, nil
}

func outPutBlockResp(resps []*core_types.ResultBlock,page int, perPage int, detail bool)  {
	totalCount := len(resps)
	page = validatePage(page, perPage, totalCount)
	perPage = validatePerPage(perPage)
	skipCount := validateSkipCount(page, perPage)
	resultLength := common.MinInt(perPage, totalCount - skipCount)

	fmt.Println("\nTotal ount:", totalCount)
	for i := 0; i < resultLength; i ++{
		resp := resps[i + skipCount]
		fmt.Println( "\nBlock info:")
		outPutHeader( resp.Block.Header)
		fmt.Println( "\nTransactions contained in block:")
		if resp.Block.Txs == nil || len(resp.Block.Txs) == 0 {
			fmt.Println( "[]")
		}else{
			for _, tx := range resp.Block.Txs {
				displayTx(tx)
			}
		}
	}
}

func addQueryBlockFlags(cmd *cobra.Command)  {
	err := addStringFlag(cmd, blockHeight, heightParam, "", "", "height interval of the blocks to query. integer or block interval formatted as [from:to] are accepted ", notRequired )
	if err != nil {
		panic(err)
	}
	err = addIntFlag(cmd, blockPage, pageParam, "", 1, "Page number (1 based)", notRequired)
	if err != nil {
		panic(err)
	}
	err = addIntFlag(cmd, blockPerPage, perPageParam, "", 20, "Page number (1 based)", notRequired)
	if err != nil {
		panic(err)
	}
	err = addBoolFlag(cmd, blockTxTransferOnly, transferOnlyParam, "", false, "display transfer type transaction only", notRequired)
	if err != nil {
		panic(err)
	}
}

//query validator
func queryValidator(cmd *cobra.Command, args []string)  {
	validatorUrl = viper.GetString(queryUrl)
	if len(validatorUrl) < 1 {
		fmt.Println("Illegal url is received!")
		return
	}
	cl := newAnkrHttpClient(validatorUrl)
	height := viper.GetInt64(validatorHeight)
	heightP := &height
	if height <= 0 {
		heightP = nil
	}
	resp, err := cl.Validators(heightP)
	if err != nil {
		fmt.Println("Query validators failed.", err)
		return
	}

	// decode response and display
	decodeAndDisplay(resp)
}
func addQueryValidatorFlags(cmd *cobra.Command)  {
	err := addInt64Flag(cmd, validatorHeight, heightParam, "", -1, "block height", notRequired)
	if err != nil {
		panic(err)
	}
}

//query status
func queryStatus(cmd *cobra.Command, args []string)  {
	validatorUrl = viper.GetString(queryUrl)
	if len(validatorUrl) < 1 {
		fmt.Println("Illegal url is received!")
		return
	}
	cl := newAnkrHttpClient(validatorUrl)
	resp, err := cl.Status()
	if err != nil{
		fmt.Println("Query status failed.",err)
		return
	}
	decodeAndDisplay(resp)
	//displayStruct(resp)
}

// display messages that needs decode
func decodeAndDisplay(resp interface{})  {
	jByte, err := TxSerializer.MarshalJSON(resp)
	if err != nil {
		fmt.Println(err)
		return
	}

	var distByte bytes.Buffer
	err = json.Indent(&distByte, jByte, "", "\t")
	//jsonByte, err := json.MarshalIndent(tx, "", "\t")
	if err != nil {
		fmt.Println("Marshal Error.")
		fmt.Println(err)
		return
	}
	fmt.Println(string(distByte.Bytes()))
}

//query genesis
func queryGenesis(cmd *cobra.Command, args []string)  {
	validatorUrl = viper.GetString(queryUrl)
	if len(validatorUrl) < 1 {
		fmt.Println("Illegal url is received!")
		return
	}
	cl := newAnkrHttpClient(validatorUrl)
	resp, err := cl.Genesis()
	if err != nil {
		fmt.Println("Query genesis failed.", err)
		return
	}
	decodeAndDisplay(resp)
}

//query consensus state
func queryConsensusState(cmd *cobra.Command, args []string)  {
	validatorUrl = viper.GetString(queryUrl)
	if len(validatorUrl) < 1 {
		fmt.Println("Illegal url is received!")
		return
	}
	cl := newAnkrHttpClient(validatorUrl)
	resp, err := cl.ConsensusState()
	if err != nil {
		fmt.Println("Query consensus state failed.", err)
		return
	}
	decodeAndDisplay(resp)
}

//query dump consensus state
func queryDumpConsensusState(cmd *cobra.Command, args []string)  {
	validatorUrl = viper.GetString(queryUrl)
	if len(validatorUrl) < 1 {
		fmt.Println("Illegal url is received!")
		return
	}
	cl := newAnkrHttpClient(validatorUrl)
	resp, err := cl.DumpConsensusState()
	if err != nil {
		fmt.Println("Query dump consensus state failed.", err)
		return
	}
	decodeAndDisplay(resp)
}

//query unconfirmed transactions
func queryUnconfirmedTxs(cmd *cobra.Command, args []string)  {
	cl := newAnkrHttpClient(viper.GetString(queryUrl))
	limmit := viper.GetInt(unconfirmedTxLimit)
	resp, err := cl.UnconfirmedTxs(limmit)
	if err != nil {
		fmt.Println("Query unconfirmed transactions failed.", err)
		return
	}
	outputTxResult(resp)
}

func outputTxResult(txResult *core_types.ResultUnconfirmedTxs)  {
	fmt.Println( "n_tx: ", txResult.Count)
	fmt.Println( "total:", txResult.Total)
	fmt.Println("total_bytes:", txResult.TotalBytes)
	fmt.Println("transactions:")
	if len(txResult.Txs) == 0 {
		fmt.Println("[]")
	}else {
		for _, tx := range txResult.Txs {
			displayTx(tx)
		}
	}
}

func addQueryUncofirmedTxsFlags(cmd *cobra.Command)  {
	err := addIntFlag(cmd, unconfirmedTxLimit, limitParam, "",30, "number of entries", notRequired)
	if err != nil {
		panic(err)
	}
}

//query number of unconfirmed transactions
func queryNumUnconfiredTxs(cmd *cobra.Command, args []string)  {
	validatorUrl = viper.GetString(queryUrl)
	if len(validatorUrl) < 1 {
		fmt.Println("Illegal url is received!")
		return
	}
	cl := newAnkrHttpClient(validatorUrl)
	resp, err := cl.NumUnconfirmedTxs()
	if err != nil {
		fmt.Println("Query number of unconfirmed transactions failed.", err)
		return
	}
	decodeAndDisplay(resp)
}

//transaction data structure
type Transaction struct {
	Type string
	Hash string
	From string
	To string
	Nonce string
	Amount string
}


//transaction data structure used in parsing all kinds of transactions
type ResultTx struct {
	Type string `json:"type"`
	Hash     string   `json:"hash"`//common.HexBytes           `json:"hash"`
	Height   int64                  `json:"height"` //block height
	Index    uint32                 `json:"index"` //transaction index in block
	Data map[string] string `json:"data"` //used to store different type of transaction data
}

func outPutHeader(header types.Header)  {
	//information to be displayed in the window
	fmt.Println("Version: ", header.Version)
	fmt.Println("Chain-Id:", header.ChainID)
	fmt.Println("Height: ", header.Height)
	fmt.Println("Time:", header.Time)
	fmt.Println("Number-Txs: ", header.NumTxs)
	fmt.Println("Total-Txs:", header.TotalTxs)
	fmt.Println("Last-block-id: ", header.LastBlockID)
	fmt.Println( "Last-commit-hash:",header.LastCommitHash)
	fmt.Println("Data-hash: ", header.DataHash)
	fmt.Println("Validator:", header.ValidatorsHash)
	fmt.Println("Consensus: ", header.ConsensusHash)
	fmt.Println("Version: ", header.Version)
	fmt.Println("App-hash:", header.AppHash)
	fmt.Println("Proposer-Address:", header.ProposerAddress)
}

func queryAccount(cmd *cobra.Command, args []string){
	client := newAnkrHttpClient(viper.GetString(queryUrl))
	req := new(common2.AccountQueryReq)
	req.Addr = viper.GetString(queryAccAddress)
	resp := new(common2.AccountQueryResp)
	err := client.Query("/store/account", req, resp)
	if err != nil {
		fmt.Println("Query Account Failed.")
		fmt.Println(err)
		return
	}
	decodeAndDisplay(resp)
}

func addQueryAccountFlags(cmd *cobra.Command)  {
	err := addStringFlag(cmd, queryAccAddress, addressParam, "", "", "account address", required)
	if err != nil {
		panic(err)
	}
}

func getBalance(cmd *cobra.Command, args []string) {
	client := newAnkrHttpClient(viper.GetString(queryUrl))
	req := new(common2.BalanceQueryReq)
	req.Address = viper.GetString(queryAddress)
	req.Symbol = viper.GetString(querySymbol)
	balanceResp := new(common2.BalanceQueryResp)
	err := client.Query("/store/balance", req, balanceResp)
	if err != nil {
		fmt.Println(err)
		return
	}
	displayStruct(balanceResp)
}

//get balance functions
func addGetBalanceFlags(cmd *cobra.Command) {
	err := addStringFlag(cmd, queryAddress, addressParam, "a", "", "the address of an account.", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, querySymbol, symbolParam, "", "ANKR", "token symbol", notRequired)
	if err != nil {
		panic(err)
	}
}

func getNonce(cmd *cobra.Command, args []string) {
	client := newAnkrHttpClient(viper.GetString(queryUrl))
	req := new(common2.NonceQueryReq)
	req.Address = viper.GetString(queryNonceAddr)
	nonceResp := new(common2.NonceQueryResp)
	err := client.Query("/store/nonce", req, nonceResp)
	if err != nil {
		fmt.Println(err)
		return
	}
	displayStruct(nonceResp)
}

func addGetNonceFlags(cmd *cobra.Command) {
	err := addStringFlag(cmd, queryNonceAddr, addressParam, "a", "", "the address of an account.", required)
	if err != nil {
		panic(err)
	}
}

func getCurrency(cmd *cobra.Command, args []string) {
	client := newAnkrHttpClient(viper.GetString(queryUrl))
	req := new(common2.CurrencyQueryReq)
	req.Symbol = viper.GetString(queryCurrencySymbol)
	currencyResp := new(common2.CurrencyQueryResp)
	err := client.Query("/store/currency", req, currencyResp)
	if err != nil {
		fmt.Println(err)
		return
	}
	displayStruct(currencyResp)
}

func addQueryCurrencyFlags(cmd *cobra.Command) {
	err := addStringFlag(cmd, queryCurrencySymbol, symbolParam, "", "", "currency symbol.", required)
	if err != nil {
		panic(err)
	}
}