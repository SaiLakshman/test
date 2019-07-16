package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var _mainLogger = shim.NewLogger("ScrubbingSmartContract")

//SmartContract represents the main entart contract
type SmartContract struct {
	scrubbing *ScrubbingSMS
}

// Init initializes chaincode.
func (sc *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_mainLogger.Infof("Inside the init method ")
	sc.scrubbing = new(ScrubbingSMS)
	return shim.Success(nil)
}
// func (sc *SmartContract) probe(stub shim.ChaincodeStubInterface) pb.Response {
// 	ts := ""
// 	_mainLogger.Info("Inside probe method")
// 	tst, err := stub.GetTxTimestamp()
// 	if err == nil {
// 		ts = tst.String()
// 	}
// 	output := "{\"status\":\"Success\",\"ts\" : \"" + ts + "\" }"
// 	_mainLogger.Info("Retuning " + output)
// 	return shim.Success([]byte(output))
// }

//Invoke is the entry point for any transaction
func (sc *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var response pb.Response
	action, _ := stub.GetFunctionAndParameters()
	switch action {
	// case "probe":
	// 	response = sc.probe(stub)
	case "csc":
		response = sc.scrubbing.CreateScrubDetails(stub)
	case "qsc":
		response = sc.scrubbing.QueryScrubDetails(stub)
	case "uscs":
		response = sc.scrubbing.UpdateScrubStatus(stub)
	case "cbsc":
		response = sc.scrubbing.CreateBulkScrubDetails(stub)
	case "hakp":
		response = sc.scrubbing.getByAnyKeyWithPagination(stub)
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

