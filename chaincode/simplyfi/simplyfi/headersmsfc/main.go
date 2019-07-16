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

//Invoke is the entry point for any transaction
func (sc *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var response pb.Response
	action, _ := stub.GetFunctionAndParameters()
	switch action {
	case "rh":
		response = sc.header.registerHeader(stub)
	case "rbh":
		response = sc.header.registerBulkHeader(stub)
	case "us":
		response = sc.header.updateStatus(stub)
	case "qcli":
		response = sc.header.queryByHeader(stub)
	case "hcli":
		response = sc.header.getHistoryByHeader(stub)
	case "qpg":
		response = sc.header.getDataByPagination(stub)
	case "qdp":
		response = sc.header.getByDateRange(stub)
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
