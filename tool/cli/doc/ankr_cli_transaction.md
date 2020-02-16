## ankr_cli transaction

transaction is used to send coins to specified address or send metering
### Sub commands

```
PS D:\> ankr-cli transaction --help
transaction is used to send coins to specified address or send metering

Usage:
  ankr-chain-cli transaction [flags]
  ankr-chain-cli transaction [command]

Available Commands:
  deploy      deploy smart contract
  invoke      invoke smart contract
  metering    send metering transaction
  transfer    send coins to another account

Flags:
      --chain-id string    block chain id (default "ankr-chain")
      --gas-limit string   gas limit (default "20000000")
      --gas-price string   gas price(should more than 10000000000000) (default "10000000000000")
  -h, --help               help for transaction
      --memo string        transaction memo
      --nodeurl string     the url of a validator
      --version string     block chain net version (default "1.0")
Use "ankr_cli transaction [command] --help" for more information about a command.
```

### usage

```
    global options 
        --chain-id string    block chain id (default "ankr-chain")
        --gas-limit string   gas limit (default "20000")
        --gas-price string   gas price(should more than 10000000000000) (default "10000000000000")
        --memo string        transaction memo
        --nodeurl string     the url of a validator
        --version string     block chain net version (default "1.0")
        --nodeurl string       url of a validator 
    * metering, send metering transaction.  
        options: 
            --dcname string      data center name
            --namespace string   namespace
            --privkey string     admin private key
            --value string       the value to be set
    * transfer, send coins to another account.   
        options: 
            --amount string     amount of ankr token to be transfered
            --keystore string   keystore of the transfer from account
            --to string         receive ankr token address
            
    * deploy deploy smart contract
        options: 
            --abi string        smart contract abi in json format
            -f, --file string       smart contract binary file name
            --keystore string   keystore file name
            --name string       smart contract name (default "contract")
            
```
### example    
+ metering  
    ```
    PS D:\> ankr-cli transaction metering --nodeurl http://localhost:26657 --dcname datacenter_name --namespace test --value test-value --privkey privkey --chain-id test-chain-FojX8z
    Set metering success.
    ```  
+ transfer    
    ```
    PS D:\> ankr-cli transaction transfer --to F4656949BD747057A59DDF90A218EC352E3916A096924D --amount 20000000000000000000 --keystore .\UTC--2019-08-01T02-25-01.685454800Z--E1403CA0DC201F377E820CFA62117A48D4D612400C20D3 --nodeurl http://localhost:26657 --chain-id test-chain-FojX8z
    Please input the keystore password:
    
    Transaction sent. Tx hash: 210AEB37AD654AE04CC7A5FC650C23CD4E03A12CC4D2A63A1288D534A8475C31
    ``` 
+ deploy 
    ```
    
    ```