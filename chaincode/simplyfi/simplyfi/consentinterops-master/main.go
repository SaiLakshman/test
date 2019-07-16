package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

//SmartContract is a structure
type SmartContract struct {
	consentManager *ConsentManager
}

var _mainLogger = shim.NewLogger("ConsentManagementSmartContract")

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		_mainLogger.Criticalf("Error starting  chaincode: %v", err)
	}
}

//Init initializes chaincode.
func (sc *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_mainLogger.Infof("Inside the init method ")
	sc.consentManager = new(ConsentManager)
	return shim.Success(nil)
}

func (sc *SmartContract) probe(stub shim.ChaincodeStubInterface) pb.Response {
	ts := ""
	_mainLogger.Info("Inside probe method")
	tst, err := stub.GetTxTimestamp()
	if err == nil {
		ts = tst.String()
	}

	output := "{\"status\":\"Success\",\"ts\" : \"" + ts + "\" }"
	_mainLogger.Info("Retuning " + output)
	return shim.Success([]byte(output))
}

//Invoke is the entry point for any transaction in Consent Module
func (sc *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	action, _ := stub.GetFunctionAndParameters()
	_mainLogger.Infof("Inside1 the invoke method with %s", action)

	switch action {
	case "probe":
		return sc.probe(stub)
	case "recordConsent":
		return sc.consentManager.RecordConsent(stub)
	case "getConsent":
		return sc.consentManager.GetConsent(stub)
	case "getHistory":
		return sc.consentManager.GetHistoryByKey(stub)
	case "updateConsentStatus":
		return sc.consentManager.UpdateConsentStatus(stub)
	case "updateConsentStatusByHeaderAndMsisdn":
		return sc.consentManager.UpdateConsentStatusByHeader(stub)
	case "updateConsentStatusByIDs":
		return sc.consentManager.UpdateConsentStatusByIDs(stub)
	case "updateConsentExpiryByIDs":
		return sc.consentManager.UpdateConsentExpiryDateByIDs(stub)
	case "updateConsentExpiryByHeaderAndMsisdn":
		return sc.consentManager.UpdateConsentExpiryDateByHeader(stub)
	case "getActiveConsentsByMSISDN":
		return sc.consentManager.GetActiveConsentsByMSISDN(stub)
	case "revokeActiveConsentsByMsisdn":
		return sc.consentManager.RevokeActiveConsentsByMsisdn(stub)

	default:
		return shim.Error("Invalid action provoided")
	}

}
