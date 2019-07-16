package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var _mainLogger = shim.NewLogger("TemplateContract")

//SmartContract represents the main contract
type SmartContract struct {
	template *TemplateManager
}

// Init initializes chaincode.
func (sc *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_mainLogger.Infof("Inside the init method ")
	sc.template = new(TemplateManager)
	return shim.Success(nil)
}

//Invoke is the entry point for any transaction
func (sc *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var response pb.Response
	action, _ := stub.GetFunctionAndParameters()
	switch action {
		case "st": // add Template
			response= sc.template.setTemplate(stub)
		case "abt": //add batch Templates
		 	response= sc.template.addBatchTemplates(stub)
		case "uts": //update Template status
			response= sc.template.updateTemplateStatus(stub)
		case "qt": //Rich Query to retrieve the Templates from DL
			 response= sc.template.queryTemplates(stub)
		case "gt": //get Template data based on TemplateID
		 	response= sc.template.getTemplateByTemplateID(stub)
		case "th": //Rich Query to retrieve the Templates History from DL
		 	response= sc.template.queryTemplatesHistory(stub)
		case "qtp": //Rich Query to retrieve the Templates with pagination from DL
			 response= sc.template.queryTemplatesWithPagination(stub)
		case "gid": //Rich Query to retrieve the Templates with pagination from DL
		 	response= sc.template.getID(stub)
		default:
			response= shim.Error("Invalid Action Provided, Available Functions: st,abt,uts,qt,th,qtp,gt")
		}
	return response
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		_mainLogger.Criticalf("Error starting chaincode: %v", err)
	}
}
