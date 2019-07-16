# Chaincode repository for UCC entity management 
## 08-May-2019
### Changelog
1. Added "getHistoryByKey" & "updateEntityStatus" methods
2. Removed updatedBy as mandatory input for entity update . 


This repository contains the chaincode for Entity management. The vendor folder contains all the necessaey dependent libraries. 


To  insert a record in ledger run the following commands from the CLI

```sh

peer chaincode invoke -o orderer0.ucccpr.com:7050  --tls --cafile $ORDERER_CA -C entitychannel -n entity -c '{"args":["createEntityRecord","{\"id\":\"1001103396725306\",\"reqid\":\"A005\",\"etype\":\"P\",\"poi\":\"ABC\",\"name\":\"Airtel001\",\"pid\":\"pid002\",\"eclass\":\"PE\",\"svcprv\":\"JI\",\"sts\":\"I\",\"appby\":\"abhijit007\",\"appon\":\"system1\"}"]}'


```

To fetch the inserted query run the following command from CLI

```sh
peer chaincode query --tls --cafile $ORDERER_CA -C entitychannel -n entity -c  {"args":["searchEntityRecord","{\"typ\":\"name\",\"name\":\"Airtel001\"}"]}'

```



### Dependencies

1. Hyperledger Fabric ( https://github.com/hyperledger/fabric )


### List of commands used for govendoring to populate the dependencies into the vendor folder
 

```sh
cd <chaincode path>
govendor init
govendor fetch github.com/hyperledger/fabric/core/chaincode/shim/ext/cid
govendor fetch github.com/op/go-logging

```



