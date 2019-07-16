package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var _mainLogger = shim.NewLogger("ConsentsContract")

//SmartContract represents the main entart contract
type SmartContract struct {
	consent *ConsentManager
}

// Init initializes chaincode.
func (sc *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_mainLogger.Infof("Inside the init method ")
	sc.consent = new(ConsentManager)
	return shim.Success(nil)
}

//Invoke is the entry point for any transaction
func (sc *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var response pb.Response
	action, _ := stub.GetFunctionAndParameters()
	switch action {
	case "cc":
		response = sc.consent.createConsent(stub)
	case "qcid":
		response = sc.consent.queryConsentById(stub)
	case "cbc":
		response = sc.consent.createBulkConsent(stub)
	case "us":
		response = sc.consent.updateStatus(stub)
	case "up":
		response = sc.consent.updatePurpose(stub)
	case "hcid":
		response = sc.consent.getHistoryByConsentId(stub)
	case "qpg":
		response = sc.consent.getDataByPagination(stub)
	default:
		response = shim.Error("Invalid action provided")
	}
	return response
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		_mainLogger.Criticalf("Error starting  chaincode: %v", err)
	}
}
