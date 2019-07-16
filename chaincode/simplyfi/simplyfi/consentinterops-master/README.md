### ChangeLog dt:03/06/2019
added methods:
1. UpdateConsentExpiryDateByIDs - New method to update the ExpiryDate by the input URNs. 
2. UpdateConsentExpiryDateByHeader - New method to update for given set of (Cli and Msisdn)s
3. GetActiveConsentsByMSISDN - New Method to get active Consents for a given MSISDN. Any consent with status -2 is considered to be active.
4. RevokeActiveConsentsByMsisdn - New Method to revoke the active consents for a given MSISDN. 
5. New status 4 ( churned ) has been taken care of. 




# Chaincode repository for UCC consent management 

### Setup instructions

Please make the directory names as specified below. ***DO NOT MISS THE DOT(.) AT THE END OF git clone command***

```sh

cd $GOPATH
mkdir -p $GOPATH/src/ibm.com/ucc/consentmgmt-cc
cd $GOPATH/src/ibm.com/ucc/consentmgmt-cc
git clone git@github.ibm.com:TRAIUCCFabricSolution/consentmgmt-cc.git .
go get -u github.com/hyperledger/fabric
go build   

```

### Dependencies
1. Hyperledger Fabric ( https://github.com/hyperledger/fabric )


### Dependency injection ( Not required for setup)

```sh
cd $GOPATH/src/ibm.com/ucc/consentmgmt-cc
govendor init
govendor fetch github.com/hyperledger/fabric/core/chaincode/shim/ext/cid
govendor fetch github.com/op/go-logging

```





