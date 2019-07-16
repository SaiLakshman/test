# Chaincode repository for UCC entity management 
## 09-July-2019
### Changelog
 1. updateStatus operator update fix

## 04-July-2019
### Changelog
 1. appby, reqid field removed
 2. Known Brand(K) type removed in entitytype
 3. blacklisted field added
 4. sts field structure change, now status stored operator wise
 5. Domain based validation
 6. Methods Added: UpdateBlacklistedValue, SearchEntityIDArray, EntityQueryWithPagination
 7. Method signature Updated:UpdateEntityStatus

## 03-June-2019
### Changelog
 1. validation of uts, cts in entity create
 2. status:true added in response for CreateEntity and ModifyEntity
 3. Error message updated
 4.  Spellings fixed

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



