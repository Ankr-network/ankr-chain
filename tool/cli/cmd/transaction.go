package cmd

import (
	"errors"
	"fmt"
	client2 "github.com/Ankr-network/ankr-chain/client"
	"github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/crypto"
	"github.com/Ankr-network/ankr-chain/tx/contract"
	"github.com/Ankr-network/ankr-chain/tx/metering"
	"github.com/Ankr-network/ankr-chain/tx/serializer"
	"github.com/Ankr-network/ankr-chain/tx/token"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"math/big"
	"os"
)

var (
	validatorUrl string
	transferUrl    = "transferUrl"
	transferChainId = "transferChainId"
	transferGasPrice = "transferGasPrice"
	transferGasLimit = "transferGasLimit"

	//names of flags used in viper to bind keys
	transferTo      = "transferTo"
	transferMemo    = "transferMemo"
	transferAmount  = "transferAmount"
	transferKeyfile = "transferKeyfile"
	meteringDc      = "meteringDc"
	meteringNs      = "meteringNs"
	meteringValue   = "meteringValue"
	meteringPriv    = "meteringPriv"
	transferVersion = "transferVersion"
	transferSymbol = "ANKR"
	deployPriv = "deployPriv"
	deployContractName = "deployContractName"
	deployBin = "deployBin"
	deployAbi = "deployAbi"

	invokeAddr = "invokeAddr"
	invokeName = "invokeName"
	invokeArgs = "invokeArgs"
	invokeReturn = "invokeReturn"
	invokeKeyStore = "invokeKeyStore"
	getContractAddr = "getContractAddr"

)

// transactionCmd represents the transaction command
var transactionCmd = &cobra.Command{
	Use:   "transaction",
	Short: "transaction is used to send coins to specified address or send metering",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("transaction called")
	},
}

func init() {
	err := addPersistentString(transactionCmd, transferUrl, urlParam, "", "", "the url of a validator", required)
	if err != nil {
		panic(err)
	}
	err = addPersistentString(transactionCmd, transferChainId, chainIDParam, "", "ankr-chain", "block chain id", notRequired)
	if err != nil {
		panic(err)
	}
	err = addPersistentString(transactionCmd, transferGasPrice, gasPriceParam, "", "10000000000000000", "gas price(should more than 10000000000000000)", notRequired)
	if err != nil {
		panic(err)
	}

	err = addPersistentString(transactionCmd, transferMemo, memoParam, "", "", "transaction memo", notRequired)
	if err != nil {
		panic(err)
	}
	err = addPersistentString(transactionCmd, transferGasLimit, gasLimitParam, "", "20000", "gas limit", notRequired)
	if err != nil {
		panic(err)
	}

	err = addPersistentString(transactionCmd, transferVersion, versionParam, "", "1.0", "block chain net version", notRequired)
	if err != nil {
		panic(err)
	}
	appendSubCmd(transactionCmd, "transfer", "send coins to another account", transfer, addTransferFlag)
	appendSubCmd(transactionCmd, "metering", "send metering transaction", sendMetering, addMeteringFlags)
	appendSubCmd(transactionCmd, "deploy", "deploy smart contract", runDeploy, addDeployFlags)
	appendSubCmd(transactionCmd, "invoke", "invoke smart contract", runInvoke, addInvokeFlags)
}

//transaction transfer functions
func transfer(cmd *cobra.Command, args []string) {
	if len(args) > 1 {
		fmt.Println("Too much arguments received.")
		return
	}
	keystorePath := viper.GetString(transferKeyfile)
	_, err := os.Stat(keystorePath)
	if err != nil {
		fmt.Println("Error: Keystore does not exist!")
		return
	}
	privateKey := decryptPrivatekey(keystorePath)
	if privateKey == "" {
		fmt.Println("Error: Wrong keystore or password!")
		return
	}
	acc, err := getAccountFromPrivatekey(privateKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	validatorUrl = viper.GetString(transferUrl)

	client := newAnkrHttpClient(validatorUrl)

	//gather inputs
	if len(args) != 0 {
		transferSymbol = args[0]
	}
	amount := viper.GetString(transferAmount)
	amountInt, ok:= new(big.Int).SetString(amount, 10)
	if !ok {
		fmt.Println("Invalid Transfer Amount Received.")
		return
	}
	currency := new(common.Currency)
	currency.Symbol = transferSymbol
	txAmount := common.Amount{*currency, amountInt.Bytes()}

	//transaction msg header
	txMsgheader, err := getTxmsgHeader()
	if err != nil {
		fmt.Println(err)
		return
	}

	//transfer msg
	transferMsg := new(token.TransferMsg)
	transferMsg.FromAddr = acc.Address
	transferMsg.ToAddr = viper.GetString(transferTo)
	transferMsg.Amounts = append(transferMsg.Amounts, txAmount)

	//transaction builder
	key := crypto.NewSecretKeyEd25519(acc.PrivateKey)
	keyAddr, _ := key.Address()
	transferMsg.FromAddr = fmt.Sprintf("%X", keyAddr)
	builder := client2.NewTxMsgBuilder(*txMsgheader, transferMsg, serializer.NewTxSerializerCDC(), key)
	fmt.Println("Start Sending transaction...")
	txHash, txHeight, _, err := builder.BuildAndCommit(client)
	if err != nil {
		fmt.Println("Transaction commit failed.")
		fmt.Println(err)
		return
	}
	fmt.Println("\nTransaction commit successful.")
	fmt.Println("Transaction hash", txHash)
	fmt.Println("Transaction height", txHeight)
}

func addTransferFlag(cmd *cobra.Command) {
	err := addStringFlag(cmd, transferTo, toParam, "", "", "transaction receiver", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, transferAmount, amountParam, "", "0", "transfer amount", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, transferKeyfile, keystoreParam, "", "", "keystore to unlock account", required)
	if err != nil {
		panic(err)
	}
}

//transaction metering function
func sendMetering(cmd *cobra.Command, args []string) {
	privPem := viper.GetString(meteringPriv)

	client := newAnkrHttpClient(viper.GetString(transferUrl))
	//transaction msg header
	txMsgheader, err := getTxmsgHeader()
	if err != nil {
		fmt.Println(err)
		return
	}
	//metering msg
	meteringMsg := new(metering.MeteringMsg)
	dc := viper.GetString(meteringDc)
	meteringMsg.DCName = dc
	meteringMsg.NSName = viper.GetString(meteringNs)
	meteringMsg.Value = viper.GetString(meteringValue)

	resp := new(common.CertKeyQueryResp)

	err = client.Query("/store/certkey",&common.CertKeyQueryReq{dc}, resp)
	if err != nil {
		fmt.Println(err)
		return
	}

	key := crypto.NewSecretKeyPem(privPem, resp.PEMBase64, "@mert:"+"dc1_"+"ns1")

	builder := client2.NewTxMsgBuilder(*txMsgheader, meteringMsg, serializer.NewTxSerializerCDC(), key)
	fmt.Println("Start Sending transaction...")
	txHash, cHeight, _, err := builder.BuildAndCommit(client)
	if err != nil {
		fmt.Println("Send CertMsg failed.")
		fmt.Println(err)
		return
	}
	fmt.Println("Send CertMsg successful.")
	fmt.Println("transaction hash:", txHash)
	fmt.Println("transaction height:", cHeight)
}

func addMeteringFlags(cmd *cobra.Command) {
	//cmd.Flags().StringVarP(&privateKey, "privkey", "p", "", "admin private key")
	err := addStringFlag(cmd, meteringDc, dcnameParam, "", "", "data center name", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, meteringNs, nameSpaceParam, "", "", "namespace", required)
	if err != nil {
		panic(err)
	}

	err = addStringFlag(cmd, meteringValue, valueParam, "", "", "the value to be set", required)
	if err != nil {
		panic(err)
	}

	err = addStringFlag(cmd, meteringPriv, privkeyParam, "", "", "admin private key", required)
	if err != nil {
		panic(err)
	}
}

func runDeploy(cmd *cobra.Command, args []string){
	client := newAnkrHttpClient(viper.GetString(transferUrl))
	header, err := getTxmsgHeader()
	if err != nil {
		fmt.Println(err)
		return
	}

	contractFile := viper.GetString(deployBin)
	wasmBin, err := ioutil.ReadFile(contractFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	keyStore := viper.GetString(deployPriv)
	privKey := decryptPrivatekey(keyStore)
	if privKey == "" {
		fmt.Println("Error: Wrong keystore or password!")
		return
	}
	//acc, _ := getAccountFromPrivatekey(privateKey)
	
	contractMsg := new(contract.ContractDeployMsg)
	contractMsg.Name = viper.GetString(deployContractName)
	//contractMsg.FromAddr = acc.Address
	contractMsg.Codes = wasmBin
	contractMsg.CodesDesc = viper.GetString(abiParam)
	key := crypto.NewSecretKeyEd25519(privKey)
	keyAddr, err := key.Address()
	if err != nil {
	    fmt.Println("Error: Wrong Privekey!")
		fmt.Println(err)
		return
	}
	contractMsg.FromAddr = fmt.Sprintf("%X", keyAddr)
	builder := client2.NewTxMsgBuilder(*header, contractMsg, serializer.NewTxSerializerCDC(), key)
	fmt.Println("Start Sending transaction...")
	txHash, cHeight, contractAddr, err := builder.BuildAndCommit(client)
	if err != nil {
		fmt.Println("Deploy smart contract failed!")
		fmt.Println(err)
		return
	}

	fmt.Println("Contract deployed successful.")
	fmt.Println("transaction hash:", txHash)
	fmt.Println("block height:", cHeight)
	fmt.Println("contract address:", contractAddr)
}

func addDeployFlags(cmd *cobra.Command)  {
	err := addStringFlag(cmd, deployBin, fileParam, "f","", "smart contract binary file name", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, deployAbi, abiParam, "", "", "smart contract abi in json format", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, deployContractName, nameParam, "", "contract", "smart contract name", notRequired)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, deployPriv, keystoreParam, "", "", "keystore file name", required)
	if err != nil {
		panic(err)
	}
}

func runInvoke(cmd *cobra.Command, args []string)  {
	client := newAnkrHttpClient(viper.GetString(transferUrl))
	header, err := getTxmsgHeader()
	if err != nil {
		fmt.Println(err)
		return
	}

	keyFile := viper.GetString(invokeKeyStore)
	privKey := decryptPrivatekey(keyFile)
	if privKey == ""{
		fmt.Println("Error: Wrong keystore or password!")
		return
	}
	acc, _ := getAccountFromPrivatekey(privKey)

	invokeMsg := new(contract.ContractInvokeMsg)
	invokeMsg.FromAddr = acc.Address
	invokeMsg.Args = viper.GetString(invokeArgs)
	invokeMsg.Method = viper.GetString(invokeName)
	invokeMsg.ContractAddr = viper.GetString(invokeAddr)
	invokeMsg.RtnType = viper.GetString(invokeReturn)
	key := crypto.NewSecretKeyEd25519(privKey)
	keyAddr, _ := key.Address()
	invokeMsg.FromAddr = fmt.Sprintf("%X",keyAddr)
	builder := client2.NewTxMsgBuilder(*header, invokeMsg, serializer.NewTxSerializerCDC(), key)
	fmt.Println("Start Sending transaction...")
	txHash, cHeight, contractResultJson, err := builder.BuildAndCommit(client)
	if err != nil {
		fmt.Println("Invoke contract failed.")
		fmt.Println(err)
		return
	}
	fmt.Println("Invoke smart contract successful.")
	fmt.Println("transaction hash:", txHash)
	fmt.Println("block height:", cHeight)
	fmt.Println("contract address:", contractResultJson)

}

func addInvokeFlags(cmd *cobra.Command)  {
	err := addStringFlag(cmd, invokeAddr, addressParam, "", "", "contract address", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, invokeName, methodParam, "", "", "method name", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, invokeArgs, argsParam, "", "", "method input arguments",notRequired)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, invokeReturn, returnParam, "", "", "return type", notRequired)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, invokeKeyStore, keystoreParam, "", "", "keystore file name ", required)
	if err != nil {
		panic(err)
	}
}

func runGetContract(cmd *cobra.Command, args []string)  {
	client := newAnkrHttpClient(viper.GetString(queryUrl))
	resp := new(common.ContractQueryResp)
	req := new(common.ContractQueryReq)
	req.Address = viper.GetString(getContractAddr)
	err := client.Query("/store/contract", req, resp)
	if err != nil {
		fmt.Println("Query contract failed.")
		fmt.Println(err)
		return
	}
	decodeAndDisplay(resp)
}

func addGetContractFlags(cmd *cobra.Command)  {
	err := addStringFlag(cmd, getContractAddr, addressParam, "", "", "contract address", required)
	if err != nil {
		panic(err)
	}
}

//get transaction message header
func getTxmsgHeader() (*client2.TxMsgHeader, error)  {
	header := new(client2.TxMsgHeader)
	chainId := viper.GetString(transferChainId)
	gasLimit := viper.GetString(transferGasLimit)
	limitInt, ok:= new(big.Int).SetString(gasLimit, 10)
	if !ok {
		fmt.Println("Invalid Transfer Amount Received.")
		return nil, errors.New("Invalid Gas Limmit received. ")
	}
	gasPrice := viper.GetString(transferGasPrice)
	priceInt, ok:= new(big.Int).SetString(gasPrice, 10)
	if !ok {
		fmt.Println("Invalid Transfer Amount Received.")
		return nil, errors.New("Invalid Gas Price received. ")
	}
	//transaction msg header
	header.Version = viper.GetString(transferVersion)
	header.ChID = common.ChainID(chainId)
	header.GasLimit = limitInt.Bytes()
	header.GasPrice.Cur = ankrCurrency
	header.GasPrice.Value = priceInt.Bytes()
	header.Memo = viper.GetString(transferMemo)
	return header, nil
}
