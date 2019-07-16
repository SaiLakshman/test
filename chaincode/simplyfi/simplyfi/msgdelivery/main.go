package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var _mainLogger = shim.NewLogger("MSGDeliveryContract")

//SmartContract represents the main entart contract
type SmartContract struct {
	message *MSGDeliveryManager
}

// Init initializes chaincode.
func (sc *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_mainLogger.Infof("Inside the init method ")
	sc.message = new(MSGDeliveryManager)
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
	case "cmd":
		response = sc.message.createMSGDelivery(stub)
	case "qmd":
		response = sc.message.queryMSGDelivery(stub)
	case "cbmd":
		response = sc.message.createBulkMSGDelivery(stub)
	case "qpg":
		response = sc.message.getDataByPagination(stub)
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
