package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"text/tabwriter"
	"time"
)

// accountCmd represents the account command
var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "account is used to generate new accounts, encrypt privatekey or decrypt privatekey from keystore",
}

//names of sub command bind in viper, which is used to bind flags
//naming notions subCmdNameKey. eg. "account" have a flag named "url" shall named as accountUrl,
var (
	genAccountName = "genAccountName"
	genAccOutput   = "agenAccOutput"
	genkeyPrivkey  = "genkeyPrivkey"
	genkeyOutput   = "genkeyOutput"
	genkeyName     = "genkeyName"
	exportKeystore = "exportKeystore"
	resetKeystore  = "resetKeystore"
	importFileName = "importFileName"
	importName     = "importName"
)

func init() {
	appendSubCmd(accountCmd, "generate", "generate new account.", generateAccounts, addGenAccountFlags)
	appendSubCmd(accountCmd, "export", "export keystore file based on private key and user input password.", genKeystore, addGenkeystoreFlags)
	appendSubCmd(accountCmd, "recover", "recover private key from keystore.", exportPrivatekey, addExportFlags)
	appendSubCmd(accountCmd, "reset", "reset keystore password.", resetPwd, addResetPWDFlags)
	appendSubCmd(accountCmd, "list", "list ankr account store in local", listAccount, nil)
	appendSubCmd(accountCmd, "import", "import <file>, import ankr account from keystore", importAccount, addImportAccountFlags)
}

type ExeCmd struct {
	Name     string
	Short    string
	Long     string
	Exec     func(cmd *cobra.Command, args []string)
	FlagFunc func(cmd *cobra.Command)
}

type Account struct {
	PrivateKey string `json:"private_key"`
	Address    string `json:"address"`
	Name       string `json:"name,omitempty"`
}

//account genaccount functions
func addGenAccountFlags(cmd *cobra.Command) {
	err := addStringFlag(cmd, genAccOutput, outputParam, "o", "", "output account to file", notRequired)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, genAccountName, nameParam, "", "", "keystore alias name", required)
	if err != nil {
		panic(err)
	}
}

//generate new account, encrypt private key to keystore base on user input password
func generateAccounts(cmd *cobra.Command, args []string) {
	accName := viper.GetString(genAccountName)
	if isKeyExists(accName) {
		fmt.Println("KeyName already exist, try other key name.")
		return
	}
	fmt.Println(`please record and backup keystore once it is generated, we don’t store your private key!`)
	fmt.Println("\ngenerating accounts...")
	acc := generateAccount()
	acc.Name = accName
	fmt.Println("private key: ", acc.PrivateKey, "\naddress: ", acc.Address)
	path := viper.GetString(genAccOutput)
	if path == "" {
		path = configHome()
	}
	err := generateKeystore(acc, path)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// check if a keyName is already exist in local files
func isKeyExists(keyName string) bool {
	localAccList := getKeyList()
	if len(localAccList) == 0 {
		return false
	}

	for _, key := range localAccList {
		if key.Name == keyName {
			return true
		}
	}
	return false
}

//generate keystore based account and password
func generateKeystore(acc Account, path string) error {
	fmt.Println("\nabout to export to keystore.. ")

InputPassword:
	fmt.Print("please input the keystore encryption password:")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil
	}

	fmt.Print("\nplease input password again: ")
	confirmPassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}

	if string(password) != string(confirmPassword) {
		fmt.Println("\nError:password and confirm password not match!")
		goto InputPassword
	}

	cryptoStruct, err := EncryptDataV3([]byte(acc.PrivateKey), []byte(password), StandardScryptN, StandardScryptP)
	if err != nil {
		return err
	}
	//_ := cryptoStruct

	encryptedKeyJSONV3 := EncryptedKeyJSONV3{
		Address:        acc.Address,
		Name:           acc.Name,
		Crypto:         cryptoStruct,
		KeyJSONVersion: keyJSONVersion,
	}
	jsonKey, err := json.Marshal(encryptedKeyJSONV3)
	if err != nil {
		return err
	}

	fmt.Println("\n\nexporting to keystore...")
	ts := time.Now().UTC()
	fileName := fmt.Sprintf("UTC--%s--%s", toISO8601(ts), acc.Address)
	//writePrivateKey()
	fileName = filepath.Join(path, fileName)

	err = WriteToFile(fileName, jsonKey)
	if err != nil {
		return err
	}
	fmt.Println("\ncreated keystore:", fileName)
	return nil
}

//genkeystore functions
func addGenkeystoreFlags(cmd *cobra.Command) {
	err := addStringFlag(cmd, genkeyPrivkey, privkeyParam, "p", "", "private key of an account.", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, genkeyOutput, outputParam, "o", "", "output file path.", notRequired)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, genkeyName, nameParam, "", "", "keystore alias name.", required)
	if err != nil {
		panic(err)
	}
}

func genKeystore(cmd *cobra.Command, args []string) {
	aliasName := viper.GetString(genkeyName)
	if isKeyExists(aliasName){
		fmt.Println("KeyName already exist, try other key name.")
		return
	}
	privateKey := viper.GetString(genkeyPrivkey)
	if len(privateKey) == 0 {
		fmt.Println("invalid private key")
		return
	}
	acc, err := getAccountFromPrivatekey(privateKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	acc.Name = aliasName
	path := viper.GetString(genkeyOutput)
	if path == "" {
		path = configHome()
	}
	err = generateKeystore(acc, path)
	if err != nil {
		fmt.Println(err)
	}
}

func addExportFlags(cmd *cobra.Command) {
	err := addStringFlag(cmd, exportKeystore, keystoreParam, "", "", "the path where keystore file is located or the keystore alias name.", required)
	if err != nil {
		panic(err)
	}
}

//generate private key from keystore and password
func exportPrivatekey(cmd *cobra.Command, args []string) {
	ksf := viper.GetString(exportKeystore)
	ksf, err := getKeystoreFile(ksf)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}

	privateKey := decryptPrivatekey(ksf)

	if privateKey == "" {
		fmt.Println("Empty privateKey!!")
		return
	}
	fmt.Println("\nPrivate key exported:", privateKey)
}

//decrypt private key from keystore and user input password
func decryptPrivatekey(file string) string {
	ks, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	var key EncryptedKeyJSONV3

	err = json.Unmarshal(ks, &key)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	fmt.Print("\nPlease input the keystore password: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Print("\n")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	privateKeyDecrypt, err := DecryptDataV3(key.Crypto, string(password))
	if err != nil {
		fmt.Println(err)
		return ""
	}
	privateKey := string(privateKeyDecrypt)
	return privateKey
}

func resetPwd(cmd *cobra.Command, args []string) {
	ksf := viper.GetString(resetKeystore)
	privateKey := decryptPrivatekey(ksf)

	if privateKey == "" {
		fmt.Println("Empty privateKey!!")
		return
	}

	acc, err := getAccountFromPrivatekey(privateKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	pwd := readPassword()

	cryptoStruct, err := EncryptDataV3([]byte(acc.PrivateKey), []byte(pwd), StandardScryptN, StandardScryptP)
	if err != nil {
		panic(err)
	}
	//_ := cryptoStruct

	encryptedKeyJSONV3 := EncryptedKeyJSONV3{
		Address:        acc.Address,
		Name:           acc.Name,
		Crypto:         cryptoStruct,
		KeyJSONVersion: keyJSONVersion,
	}
	jsonKey, err := json.Marshal(encryptedKeyJSONV3)
	if err != nil {
		panic(err)
	}

	err = WriteToFile(ksf, jsonKey)
	if err != nil {
		fmt.Println("Failed to reset password:", err.Error())
		return
	}
	fmt.Println("\nPassword reset success.")
}

func addResetPWDFlags(cmd *cobra.Command) {
	err := addStringFlag(cmd, resetKeystore, fileParam, "f", "", "the path where keystore file is located.", required)
	if err != nil {
		panic(err)
	}
}

//read password from terminal
func readPassword() []byte {
InputPassword:
	fmt.Print("please input the keystore encryption password:")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}

	fmt.Print("\nplease input password again: ")
	confirmPassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}

	if string(password) != string(confirmPassword) {
		fmt.Println("\nError:password and confirm password not match!")
		goto InputPassword
		//return errors.New("\npassword and confirm password not match")
	}
	return password
}

func listAccount(cmd *cobra.Command, args []string) {
	keyList := getKeyList()
	displayKeyList(keyList)
}

// iterator all the files stored in config home, read account name and address
func getKeyList() []*KeyStore {
	var keyList []*KeyStore
	homeDir := configHome()
	files, err := ioutil.ReadDir(homeDir)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return keyList
	}
	for _, file := range files {
		keyFile := filepath.Join(homeDir, file.Name())
		fileByte, err := ioutil.ReadFile(keyFile)
		if err != nil {
			fmt.Println(err)
			continue
		}
		var keyStore EncryptedKeyJSONV3
		err = json.Unmarshal(fileByte, &keyStore)
		if err != nil {
			continue
		}
		keyList = append(keyList, &KeyStore{
			Name:    keyStore.Name,
			Address: keyStore.Address,
			FileName:file.Name(),
		})
	}
	return keyList
}

func displayKeyList(keyList []*KeyStore) {
	if len(keyList) == 0 {
		fmt.Println("can not find keystore in path")
		return
	}
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 0, 4, ' ', 0)
	homeDir := configHome()
	fmt.Fprintf(w, "\n Name \tAddress\tFile\n")
	for _, key := range keyList {
		fileName := filepath.Join(homeDir, key.FileName)
		fmt.Fprintf(w, "%s\t%s\t%s\n", key.Name, key.Address,fileName)
	}
	w.Flush()
}

func importAccount(cmd *cobra.Command, args []string) {
	aliasName := viper.GetString(importName)
	if isKeyExists(aliasName){
		fmt.Println("KeyName already exist, try other key name.")
		return
	}
	keyFile := viper.GetString(importFileName)
	keyByte, err := ioutil.ReadFile(keyFile)
	var encryptedKeyJSONV3 EncryptedKeyJSONV3
	err = json.Unmarshal(keyByte, &encryptedKeyJSONV3)
	if err != nil {
		fmt.Println("Failed to read encrypted key from keystore.")
		fmt.Println("Error:", err.Error())
		return
	}
	encryptedKeyJSONV3.Name = aliasName
	writeByte, err := json.Marshal(encryptedKeyJSONV3)
	if err != nil {
		fmt.Println("Failed to marshal keystore:", err.Error())
		return
	}
	ts := time.Now().UTC()
	fileName := fmt.Sprintf("UTC--%s--%s", toISO8601(ts), encryptedKeyJSONV3.Address)
	fileName = filepath.Join(configHome(), fileName)
	err = WriteToFile(fileName, writeByte)
	if err != nil {
		fmt.Println("Import keystore failed:", err.Error())
		return
	}
	fmt.Println("Import keystore success.")
}

func addImportAccountFlags(cmd *cobra.Command) {
	err := addStringFlag(cmd, importFileName, fileParam, "f", "", "the path where keystore file is located.", required)
	if err != nil {
		panic(err)
	}
	err = addStringFlag(cmd, importName, nameParam, "", "", "keystore alias name", required)
	if err != nil {
		panic(err)
	}
}
