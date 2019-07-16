package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var _mainLogger = shim.NewLogger("HeaderSMSContract")

//SmartContract represents the main entart contract
type SmartContract struct {
	header *HeaderSMSManager
}

// Init initializes chaincode.
func (sc *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_mainLogger.Infof("Inside the init method ")
	sc.header = new(HeaderSMSManager)
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
	case "rh":
		response = sc.header.registerHeader(stub)
	case "qhid":
		response = sc.header.queryByHeaderId(stub)
	case "rbh":
		response = sc.header.registerBulkHeader(stub)
	case "us":
		response = sc.header.UpdateStatus(stub)
	case "qcli":
		response = sc.header.queryByHeader(stub)
	case "hhid":
		response = sc.header.getHistoryByHeaderId(stub)
	case "hcli":
		response = sc.header.getHistoryByHeader(stub)
	case "hakp":
		response = sc.header.getByAnyKeyWithPagination(stub)
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
