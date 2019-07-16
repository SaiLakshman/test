### ChangeLog dt:03/06/2019
added methods:
1. UpdateConsentExpiryDateByIDs - New method to update the ExpiryDate by the input URNs. 
2. UpdateConsentExpiryDateByHeader - New method to update ExpiryDate for given set of (Cli and Msisdn)s
3. GetActiveConsentsByMSISDN - New Method to get active Consents for a given MSISDN. Any consent with status -2 is considered to be active.
4. RevokeActiveConsentsByMsisdn - New Method to revoke the active consents for a given MSISDN. 
5. New status 4 ( churned ) has been taken care of. 

### ChangeLog dt:06/06/2019
added Attribjute:
1. Purpose ( string type) ( json: pur) has been added in data model. Not a madantory Field ( possible values - 1/2/3)

added method:
1. UpdateConsentPurposeByIDs - New method to update Purpose fby the input URNs. 
2. UpdateConsentPurposeByHeader - New method to update Purpose for given set of (Cli and Msisdn)s

### ChangeLog dt:10/06/2019
1. Validation on ConsentTeamplate Id kept off
2. Validation on communication Mode in place
3. Partial saving for Expiry date, Purpose, Status in place
4. GetActiveConsentsByMSISDN - fixed

In progress - Records Consents in Bulk

### ChangeLog dt:21/06/2019
1. Added new method to address the Bulk Consent Records
2. Changed the order of Purpose (Comments section only). Revised comments - are 1(Both), 2(Promotional) or (3)Service
3. Validation of input MSISDN
4. Method introduced to get pagination-based rault on raw input rich query

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




