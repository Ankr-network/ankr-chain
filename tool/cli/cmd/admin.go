package cmd

import (
	"errors"
	"fmt"
	"math/big"
	client2 "github.com/Ankr-network/ankr-chain/client"
	"github.com/Ankr-network/ankr-chain/common"
	"github.com/Ankr-network/ankr-chain/crypto"
	"github.com/Ankr-network/ankr-chain/tx/metering"
	"github.com/Ankr-network/ankr-chain/tx/serializer"
	"github.com/Ankr-network/ankr-chain/tx/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

)

// adminCmd represents the admin command
var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "admin is used to do admin operations ",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var (
	//names of sub command bind in viper, which is used to bind flags
	// naming notions subCmdNameKey. eg. "account" have a flag named "url" shall named as accountUrl,
	//persistent flags
	adminUrl        = "adminUrl"
	adminPrivateKey = "adminPrivateKey"
	adminChId = "adminChId"
	adminGasLimt = "adminGasLimt"
	adminGasPrice = "adminGasPrice"
	adminMemo = "adminMemo"
	adminVersion = "adminVersion"


	//sub cmd flags
	setCertDc           = "setCertDc"
	setCertPerm         = "setCertPerm"
	setCertPub          = "setCertPub"
	setValidPub         = "setValidPub"
	setValidAction      = "setValidAction"
	setValidName        = "setValidName"
	setValidStakeAddr   = "setValidStakeAddr"
	setValidStakeAmount = "setValidStakeAmount"
	setValidStakeHeight = "setValidStakeHeight"
	setValidFlag        = "setValidFlag"
	setValidGasUsed     = "setValidGasUsed"
	removeCertDc        = "removeCertDc"
	removeCertNs = "removeCertNs"
	removeCertPub = "removeCertPub"
)

func init() {
	//init persistent flags and append sub commands
	err := addPersistentString(adminCmd, adminUrl, urlParam, "", "", "url of a validator", required)
	if err != nil {
		panic(err)
	}

	err = addPersistentString(adminCmd, adminPrivateKey, privkeyParam, "", "", "operator private key", required)
	if err != nil {
		panic(err)
	}
	err = addPersistentString(adminCmd, adminChId, chainIDParam, "", "ankr-chain", "block chain id", notRequired)
	if err != nil {
		panic(err)
	}
	err = addPersistentString(adminCmd, adminGasPrice, gasPriceParam, "", "10000000000000", "gas price", notRequired)
	if err != nil {
		panic(err)
	}

	err = addPersistentString(adminCmd, adminMemo, memoParam, "", "", "transaction memo", notRequired)
	if err != nil {
		panic(err)
	}
	err = addPersistentString(adminCmd, adminGasLimt, gasLimitParam, "", "20000", "gas limmit", notRequired)
	if err != nil {
		panic(err)
	}

	err = addPersistentString(adminCmd, adminVersion, versionParam, "", "1.0", "block chain net version", notRequired)
	if err != nil {
		panic(err)
	}

	//add sub cmd to adminCmd
	appendSubCmd(adminCmd, "setcert", "set metering cert", setCert, addCertFlags)
	appendSubCmd(adminCmd, "validator", "add a new validator", setValidator, addSetValidatorFlags)
	appendSubCmd(adminCmd, "removecert", "remove cert from validator", removeCert, addRemoveCertFlags)
}

//admin setcert --dcname dataCenterName --certPerm certString --url https://validator-url:port
func setCert(cmd *cobra.Command, args []string) {
	client := newAnkrHttpClient(viper.GetString(adminUrl))
	opPrivateKey := viper.GetString(adminPrivateKey)
	if len(opPrivateKey) < 1 {
		fmt.Println("Invalid operator private key!")
		return
	}
	txMsg := new(metering.SetCertMsg)
	txMsg.DCName = viper.GetString(setCertDc)
	txMsg.PemBase64 = viper.GetString(setCertPerm)
	key := crypto.NewSecretKeyEd25519(viper.GetString(adminPrivateKey))

	pubBS64 := viper.GetString(setCertPub)
	txMsg.FromAddr = crypto.CreateCertAddress(pubBS64,txMsg.DCName, crypto.CertAddrTypeSet)
	header,err := getAdminMsgHeader()
	if err != nil {
		fmt.Println(err)
		return
	}
	//build and send transaction
	builder :=client2.NewTxMsgBuilder(*header, txMsg, serializer.NewTxSerializerCDC(), key)
	fmt.Println("Start Sending transaction...")
	txHash, cHeight, _, err := builder.BuildAndCommit(client)
	if err != nil {
		fmt.Println("Set Cert failed.")
		fmt.Println(err)
		return
	}

	fmt.Println("Set Cert success.")
	fmt.Println("Transaction hash:",txHash)
	fmt.Println("Block Height:", cHeight)
}

func addCertFlags(cmd *cobra.Command) {
	err := addStringFlag(cmd, setCertDc, dcnameParam, "", "", "data center name", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, setCertPerm, permParam, "", "", "cert perm to be set", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, setCertPub, pubkeyParam, "", "", "public key of opreator address", required)
	if err != nil {
		panic(err)
	}
}

// setvalidator --action --name --pubkey --address --amount --height --flag --gas-used
func setValidator(cmd *cobra.Command, args []string) {
	client := newAnkrHttpClient(viper.GetString(adminUrl))
	opPrivateKey := viper.GetString(adminPrivateKey)
	if len(opPrivateKey) < 1 {
		fmt.Println("Invalid operator private key!")
		return
	}
	header,err := getAdminMsgHeader()
	if err != nil {
		fmt.Println(err)
		return
	}
	validatorMsg := new(validator.ValidatorMsg)
	validatorMsg.Name = viper.GetString(setValidName)
	validatorMsg.Action = getAction(viper.GetString(setValidAction))
	validatorMsg.StakeAddress = viper.GetString(setValidStakeAddr)
	validatorMsg.StakeAmount.Cur = ankrCurrency
	amount, ok := new(big.Int).SetString(viper.GetString(setValidStakeAmount), 10)
	if !ok {
		fmt.Println("Invalid amount.")
		return
	}
	validatorMsg.StakeAmount.Value = amount.Bytes()
	validatorMsg.SetFlag = getFlagInfo(viper.GetString(setValidFlag))
	validatorMsg.ValidHeight = uint64(viper.GetInt(setValidStakeHeight))
	key := crypto.NewSecretKeyEd25519(opPrivateKey)
	keyAddr, _ := key.Address()
	validatorMsg.FromAddress = fmt.Sprintf("%X", keyAddr)
	builder := client2.NewTxMsgBuilder(*header, validatorMsg,serializer.NewTxSerializerCDC(), key)
	fmt.Println("Start Sending transaction...")
	txHash, cHeight, _, err := builder.BuildAndCommit(client)
	if err != nil {
		fmt.Println("Set Validator failed.")
		fmt.Println(err)
		return
	}

	fmt.Println("Set Validator success.")
	fmt.Println("Transaction hash:",txHash)
	fmt.Println("Block Height:", cHeight)
}

func getFlagInfo(flag string) common.ValidatorInfoSetFlag {
	switch flag {
	case "set-name":
		return common.ValidatorInfoSetName
	case "set-val-addr":
		return common.ValidatorInfoSetValAddress
	case "set-pub":
		return common.ValidatorInfoSetPubKey
	case "set-stake-addr":
		return common.ValidatorInfoSetStakeAddress
	case "set-val-height":
		return common.ValidatorInfoSetValidHeight
	case "set-stake-amount":
		return common.ValidatorInfoSetStakeAmount
	}
	return common.ValidatorInfoSetFlag(0)
}

func addSetValidatorFlags(cmd *cobra.Command) {
	err := addStringFlag(cmd, setValidPub, pubkeyParam, "", "", "the public address of the added validator", required)
	if err != nil {
		panic(err)
	}

	err = addStringFlag(cmd, setValidAction, actionParam, "", "", "update validator action", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, setValidName, nameParam, "", "", "update validator action", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, setValidFlag, flagParam, "", "", "flag of validator tansaction", notRequired)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, setValidStakeAddr, addressParam, "", "", "validator stake address", notRequired)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, setValidStakeAmount, amountParam, "", "", "validator stake amount", notRequired)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, setValidGasUsed, gasUsedParam, "", "", "gas used", notRequired)
	if err != nil {
		panic(err)
	}
	err = addIntFlag(cmd, setValidStakeHeight, heightParam, "", 0, "validator stake height", notRequired)
	if err != nil {
		panic(err)
	}
}

//transform action into uint type
func getAction(action string) uint8 {
	switch action {
	case "create":
		return 1
	case "update":
		return 2
	case "remove":
		return 3
	default:
		return 0
	}
}

// removecert --dcname --namespace
func removeCert(cmd *cobra.Command, args []string) {
	validatorUrl = viper.GetString(adminUrl)
	client := newAnkrHttpClient(viper.GetString(adminUrl))
	amdinPriv := viper.GetString(adminPrivateKey)
	header, err := getAdminMsgHeader()
	if err != nil {
		fmt.Println(err)
		return
	}
	txMsg := new(metering.RemoveCertMsg)
	txMsg.DCName = viper.GetString(removeCertDc)
	txMsg.NSName = viper.GetString(removeCertNs)
	key := crypto.NewSecretKeyEd25519(amdinPriv)
	pubBS64 := viper.GetString(removeCertPub)
	txMsg.FromAddr = crypto.CreateCertAddress(pubBS64,txMsg.DCName, crypto.CertAddrTypeRemove)
	
	builder := client2.NewTxMsgBuilder(*header, txMsg, serializer.NewTxSerializerCDC(), key)
	fmt.Println("Start Sending transaction...")
	txHash, cHeight, _, err := builder.BuildAndCommit(client)
	if err != nil {
		fmt.Println("Remove cert failed.")
		fmt.Println(err)
		return
	}

	fmt.Println("Remove cert success.")
	fmt.Println("Transaction hash:",txHash)
	fmt.Println("Block Height:", cHeight)

}

func addRemoveCertFlags(cmd *cobra.Command) {
	err := addStringFlag(cmd, removeCertDc, dcnameParam, "", "", "name of data center name", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, removeCertNs, nameSpaceParam, "", "", "name space", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, removeCertPub, pubkeyParam, "", "", "public key of opreator address", required)
	if err != nil {
		panic(err)
	}
}

// get transaction header .
func getAdminMsgHeader() (*client2.TxMsgHeader, error) {
	chainId := viper.GetString(adminChId)
	gasLimit := viper.GetString(adminGasLimt)
	limitInt, ok := new(big.Int).SetString(gasLimit, 10)
	if !ok {
		return nil, errors.New("Invalid Gas Limit received. ")
	}
	gasPrice := viper.GetString(adminGasPrice)
	priceInt, ok := new(big.Int).SetString(gasPrice, 10)
	if !ok {
		return nil, errors.New("Invalid Gas Price received. ")
	}
	header := new(client2.TxMsgHeader)
	header.Memo = viper.GetString(adminMemo)
	header.Version = viper.GetString(adminVersion)
	header.GasLimit = limitInt.Bytes()
	header.GasPrice.Cur = ankrCurrency
	header.GasPrice.Value = priceInt.Bytes()
	header.ChID = common.ChainID(chainId)
	return header, nil
}
