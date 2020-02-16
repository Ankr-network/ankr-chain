## ankr_cli admin

admin is used to do admin operations including set validator, set cert, set stake, set balance to target address,  remove validator and remove cert    
### Sub commands
```
PS D:\> ankr-cli admin --help
admin is used to do admin operations

Usage:
  ankr_cli admin [flags]
  ankr_cli admin [command]

Available Commands:
  removecert      remove cert from validator
  setcert         set metering cert
  validator       update validator with actions (set-name/set-val-addr/set-pub/set-stake-addr/set-val-height/set-stake-amount)

Flags:
      --chain-id string    block chain id (default "ankr-chain")
      --gas-limit string   gas limmit (default "20000000")
      --gas-price string   gas price (default "10000000000000")
  -h, --help               help for admin
      --memo string        transaction memo
      --nodeurl string     url of a validator
      --privkey string     operator private key
      --version string     block chain net version (default "1.0")

Use "ankr-cli admin [command] --help" for more information about a command.
```

### usage
    * Global Flags:
          --chain-id string    block chain id (default "ankr-chain")
          --gas-limit string   gas limmit (default "20000000")
          --gas-price string   gas price (default "10000000000000")
          --memo string        transaction memo
          --nodeurl string     url of a validator
          --privkey string     operator private key
          --version string     block chain net version (default "1.0") 
    * removecert, remove cert from validator.  
        options: 
            --dcname string      name of data center name
            -h, --help               help for removecert
            --namespace string   name space
    * setcert, set metering cert   
        options:
            --perm string     cert perm to be set
            --dcname string   data center name
    * setstake, set stake   
             options:
                 --amount string   set stake amount
                 --pubkey string   public key
    * validator, add a new validator    
            options:
                --action string     update validator action
                --address string    validator stake address
                --amount string     validator stake amount
                --flag string       flag of validator tansaction
                --gas-used string   gas used
                --height int        validator stake height
                --name string       update validator action
                --pubkey string     the public address of the added validator   
                
### example 
+ removecert 
    ```
    PS D:\> ankr-cli admin removecert --nodeurl http://localhost:26657 --privkey wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg== --dcname my-dcname
    Remove cert success. 
    ```
+ setcert 
    ```
    PS D:\> ankr-cli admin setcert --nodeurl http://localhost:26657 --privkey wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg== --dcname my-dcname --perm `signature perm`
    c2lnbmF0dXJl
    set_crt=my-dcname:c2lnbmF0dXJl:4:MKR/hOyrYKS85sjl1Je3t4DO358hx0i75NAsjV4ot/dXoo5nGDnUj4tS6KRYyEGiIk1kKL5Hf7fAqDdqb74aAQ==
    Set cert success. 
    ```    
+ validator
    ``` 
    PS D:\> ankr-cli admin validator set-name --nodeurl http://localhost:26657 --privkey Q5P4l16P+/Cyxq3BvavuWnQPkmeHNYPFkjfuWyQoNyK2vCvT1jyyoh2DYfu+EkWx/hoGjAHOqQw6PMAa7ZkXoQ== --pubkey FSyq/mTVPO/WdxMNCMEKiA5UVBFXVL8OAnDspO+buZY= --name val-name
    Set validator success.
    ```
  
