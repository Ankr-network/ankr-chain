  Ankr chain is a blockchain which serves DCCN(Distributed Cloud Compute Network). The chain is based on the tendermint development framework.At present, the chain has the following features： 
  
  1. A extendable and flexible transaction structure model. So then varied transactions are supported;
  2. POS(Proof of staking) based on tendermint BFT(Byzantine Fault Tolerance) is supported;
  3. A sophisticated and extendable smart contracts system which is based on WebAssembly standard; At present the contracts development          language is C/C++;
  4. A multi-version and multi-zone data storage based on merkle IAVL+ Tree; The storage has efficient data accessing and convenient            transaction rolling mechanism;
  5. A secure and succinct resource data collections from data center of DCCN; The collected resource data can provide some data proofs to      the DCCN resource consumer.      
    
  However, if you want to try the ankr chain, you can operate according to the following steps(suppose your local os is Ubuntu 18.04 LTS): 
   
  1.clone the ankr-chain repo into you local pc: 
    git clone https://github.com/Ankr-network/ankr-chain.git
    
  2.Build ankr-chain image in the ankr-chain root dir:     
    make (please make sure that your os has been installed make tool)
    
  3.Under the subdirectory of root dir,  the image "ankrchain" is ankr chain node image. You can run the image according to the following steps: 
  
    a. Initialize the node configuration: 
       ./ankrchain init
    b. Start node:  
       ./ankrchain start
       
  4.Now you can generate a ankr chain account by ankr-chain-cli under the subdirectory "build/tool": 
   
     ./ankrchain-cli account genaccount
     generating accounts...

     Account_0
     private key:  u5lYdYLiwo+x4T+fSUTHUnw87ZP/BYkBG1i+uCrfyl4FMJwbXVX1hJSXesSg57eWE+jDUWaXraaw/N48U8kVUw==
     address:  B7665C86D4566627F3D9A2CF7E4518156ED139D80D216C
     
  5. Import the ankr chain testing account "B508ED0D54597D516A680E7951F18CAD24C7EC9FCFCD67"'s private key by the following ankr-chain-cli command: 
  
     ./ankrchain-cli account genkeystore --privkey "wmyZZoMedWlsPUDVCOy+TiVcrIBPcn3WJN8k5cPQgIvC8cbcR10FtdAdzIlqXQJL9hBw1i0RsVjF6Oep/06Ezg=="
     
  6. Query the ankr chain id by curl command:
  
     curl 127.0.0.1:26659/status
     
  7. Transfer some ANKR coins from testing account to "B7665C86D4566627F3D9A2CF7E4518156ED139D80D216C": 
  
      ./ankrchain-cli transaction  transfer --chain-id "test-chain-Hk17dM" --gas-limit 2000 --gas-price "10000000000000" --nodeurl "localhost:26657" --amount 6 ANKR --to "B7665C86D4566627F3D9A2CF7E4518156ED139D80D216C" --keystore  <key store file>
  
  You can also try ankr chain smart contracts by the following step:
  
1. Compile the example contracts by the tool contract-compiler under the subdirectory "build/tool":

   ./contract-compiler compile ../../contract/example/cpp/TestContract2.cpp  --gen-abi --output <your abi file path>
  
2. Deploy the example contracts by ankr-chain-cli under the subdirectory "build/tool": 

   ./ankrchain-cli transaction deploy --chain-id "test-chain-Hk17dM" --gas-limit 2000 --gas-price "10000000000000" --nodeurl "localhost:26657" --file "../../contract/example/cpp/TestContract2.wasm" --name "TestContract2" --keystore  --abi ""
   
3. Invoke the example contracts by ankr-chain-cli under the subdirectory "build/tool":

   ./ankrchain-cli transaction  invoke --method "testFuncWithString" --args "[{\"index\":1,\"Name\":\"args\",\"ParamType\":\"string\",\"Value\":\"testFuncWithInt arg\"}]" --rtn-type "string" --chain-id "test-chain-Hk17dM" --gas-limit 2000 --gas-price "10000000000000" --nodeurl "localhost:26657" ---keystore  <key store file> --address "BFB8206804DC410AAFB8828ABDD36B488DCFB7FA8EF984"
  
  more detailed instruction, you can refer to the related document.
