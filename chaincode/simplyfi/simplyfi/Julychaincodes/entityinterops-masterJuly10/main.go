package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var _mainLogger = shim.NewLogger("EntityManagementSmartContract")

//SmartContract represents the main entart contract
type SmartContract struct {
	entityMgr *EntityManager
}

// Init initializes chaincode.
func (sc *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_mainLogger.Infof("Inside the init method ")
	sc.entityMgr = new(EntityManager)
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

//Invoke is the entry point for any transaction
func (sc *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var response pb.Response
	action, args := stub.GetFunctionAndParameters()
	switch action {
	case "probe":
		response = sc.probe(stub)
	case "createEntityRecord":
		response = sc.entityMgr.CreateEntity(stub)
	case "searchEntityRecord":
		response = sc.entityMgr.SearchEntity(stub)
	case "modifyEntityRecord":
		response = sc.entityMgr.ModifyEntity(stub)
	case "getHistoryByKey":
		response = sc.entityMgr.GetHistoryByKey(stub)
	case "updateEntityStatus":
		response = sc.entityMgr.UpdateEntityStatus(stub)
	case "searchEntityIDArray":
		response = sc.entityMgr.SearchEntityIDArray(stub)
	case "entityQueryWithPagination":
		response = sc.entityMgr.EntityQueryWithPagination(stub, args)
	case "updateBlacklistedValue":
		response = sc.entityMgr.UpdateBlacklistedValue(stub)
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
