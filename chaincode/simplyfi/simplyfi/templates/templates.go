/*
This Chaincode is written for storing,retrieving,deleting the Content/Consent Templates that are stored in DLT
*/

package main

import (
	"bytes"
	"encoding/json" //reading and writing JSON
	"fmt"
	"strconv" //import for msisdn validation
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"             // import for Chaincode Interface
	cid "github.com/hyperledger/fabric/core/chaincode/shim/ext/cid" // import for Client Identity
	pb "github.com/hyperledger/fabric/protos/peer"                  // import for peer response
)

//Logger for Logging
var logger = shim.NewLogger("BATCH-TEMPLATES")

//Event Names
const EVTADDTEMPLATE = "ADD-TEMPLATE"
const EVTUPDTEMPLATE = "UPDATE-TEMPLATE"

//Output Structure for the output response
type Output struct {
	Data         string `json:"data"`
	ErrorDetails string `json:"error"`
}

//Event Payload Structure
type Event struct {
	Data string `json:"data"`
	Txid string `json:"txid"`
}

//Template Type
var tempType = map[string]bool{
	"CS": true,
	"CT": true,
}

//CommunicationType
var communicationType = map[string]bool{
	"P":  true,
	"T":  true,
	"SE": true,
	"SI": true,
}

//Status
var status = map[string]bool{
	"A": true,
	"I": true,
}

func validEnumEntry(input string, enumMap map[string]bool) bool {
	if _, isEntryExists := enumMap[input]; !isEntryExists {
		return false
	}
	return true
}

//Smart Contract structure
type TemplateMgmtChaincode struct {
}

//=========================================================================================================
// Template structure, with 13 properties.  Structure tags are used by encoding/json library
//=========================================================================================================
type Template struct {
	ObjType             string   `json:"obj"`
	TemplateID          string   `json:"urn"`
	PEID                string   `json:"peid"`
	CLI                 []string `json:"cli"`
	TemplateName        string   `json:"tname"`
	TemplateType        string   `json:"ttyp"`
	CommunicationType   string   `json:"ctyp"`
	Category            string   `json:"ctgr"`
	TempContent         string   `json:"tcont"`
	Creator             string   `json:"crtr"`
	CreateTs            string   `json:"cts"`
	UpdatedBy           string   `json:"uby"`
	UpdateTs            string   `json:"uts"`
	Status              string   `json:"sts"`
	ConsentTemplateType string   `json:"csty"`
	ContentType         string   `json:"coty"`
	NoOfVariables       string   `json:"vars"`
	TMID                string   `json:"tmid"`
}

//=========================================================================================================
// Init Chaincode
// The Init method is called when the Smart Contract "Templates" is instantiated by the blockchain network
//=========================================================================================================

func (c *TemplateMgmtChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("###### Templates-Chaincode is Initialized #######")
	return shim.Success(nil)
}

// ========================================
// Invoke - Entry point for Invocations
// ========================================
func (dlt *TemplateMgmtChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	logger.Infof("Templates ChainCode Invoked, Function Name: " + string(function))
	switch function {
	case "st": // add Template
		return dlt.setTemplate(stub, args)
	case "abt": //add batch Templates
		return dlt.addBatchTemplates(stub, args)
	case "uts": //update Template status
		return dlt.updateTemplateStatus(stub, args)
	case "qt": //Rich Query to retrieve the Templates from DL
		return dlt.queryTemplates(stub, args)
	case "th": //Rich Query to retrieve the Templates History from DL
		return dlt.queryTemplateHistory(stub, args)
	case "qtp": //Rich Query to retrieve the Templates with pagination from DL
		return dlt.queryTemplateWithPagination(stub, args)
	default:
		logger.Errorf("Unknown Function Invoked, Available Function argument shall be any one of : st,abt,dt,qt,th,qtp")
		return shim.Error("Available Functions: st,abt,dt,qt,th,qtp")
	}
}

//setTemplate - Setting new Template
// ==============================================================================
func (dlt *TemplateMgmtChaincode) setTemplate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var data map[string]interface{}
	err := json.Unmarshal([]byte(args[0]), &data)
	if err != nil {
		logger.Errorf("setTemplate : Input arguments unmarhsaling Error : " + string(err.Error()))
		return shim.Error("setTemplate : Input arguments unmarhsaling Error : " + string(err.Error()))
	}
	certData, err := cid.GetX509Certificate(stub)
	if err != nil {
		logger.Errorf("setTemplate : Getting certificate Details Error : " + string(err.Error()))
		return shim.Error("setTemplate : Getting certificate Details Error : " + string(err.Error()))
	}
	if _, flag := data["urn"].(string); !flag {
		jsonResp = "{\"Error\":\"urn is empty\"}"
		return shim.Error(jsonResp)
	}
	if _, flag := data["peid"].(string); !flag {
		jsonResp = "{\"Error\":\"peid is empty\"}"
		return shim.Error(jsonResp)
	}
	if _, flag := data["tname"].(string); !flag {
		jsonResp = "{\"Error\":\"tname is empty\"}"
		return shim.Error(jsonResp)
	}
	if _, flag := data["ttyp"].(string); !flag {
		jsonResp = "{\"Error\":\"ttyp is empty\"}"
		return shim.Error(jsonResp)
	}
	if _, flag := data["ctyp"].(string); !flag {
		jsonResp = "{\"Error\":\"ctyp is empty\"}"
		return shim.Error(jsonResp)
	}
	if _, flag := data["tcont"].(string); !flag {
		jsonResp = "{\"Error\":\"tcont is empty\"}"
		return shim.Error(jsonResp)
	}
	if _, flag := data["uts"].(string); !flag {
		jsonResp = "{\"Error\":\"uts is empty\"}"
		return shim.Error(jsonResp)
	}
	if _, flag := data["sts"].(string); !flag {
		jsonResp = "{\"Error\":\"sts is empty\"}"
		return shim.Error(jsonResp)
	}
	if _, flag := data["cli"].([]interface{}); !flag {
		jsonResp = "{\"Error\":\"cli is empty\"}"
		return shim.Error(jsonResp)
	}
	cliList := data["cli"].([]interface{})
	cli := make([]string, len(cliList))
	for i, v := range cliList {
		if len(v.(string)) == 0 {
			jsonResp = "{\"Error\":\"Header data should be string\"}"
			return shim.Error(jsonResp)
		}
		cli[i] = fmt.Sprint(v)
	}
	if !validEnumEntry(data["ttyp"].(string), tempType) {
		jsonResp = "{\"Error\":\"Please enter one of these value for TemplateType 'CS' or 'CT' \"}"
		return shim.Error(jsonResp)
	}

	if !validEnumEntry(data["ctyp"].(string), communicationType) {
		jsonResp = "{\"Error\":\"Please enter one of these value for communicationType 'p' or 'T' or 'SE', or 'SI' \"}"
		return shim.Error(jsonResp)
	}

	if !validEnumEntry(data["sts"].(string), status) {
		jsonResp = "{\"Error\":\"Please enter one of these value for Status 'A' or 'I' \"}"
		return shim.Error(jsonResp)
	}

	if data["ttyp"].(string) == "CT" {
		if len(data["ctgr"].(string)) == 0 {
			jsonResp = "{\"Error\":\" Input Data: Category is empty; for Template Type 'Consent Template' category is mandatory\"}"
			return shim.Error(jsonResp)
		}
	}

	if _, err := strconv.Atoi(data["urn"].(string)); err != nil {
		jsonResp = "{\"Error\":\"URN is not numeric \"}"
		return shim.Error(jsonResp)
	}

	Organizations := certData.Issuer.Organization
	value, err := stub.GetState(data["urn"].(string))

	if err != nil {
		logger.Errorf("setTemplate : GetState Failed for TemplateID : " + data["urn"].(string) + " Error : " + string(err.Error()))
		return shim.Error("setTemplate : GetState Failed for TemplateID : " + data["urn"].(string) + " Error : " + string(err.Error()))
	}

	if value != nil {
		logger.Errorf("setTemplate : Template already exists. Please choose unique TemplateID")
		return shim.Error("setTemplate : Template already exists. Please choose unique TemplateID")
	} else {
		if len(data) == 12 || len(data) == 14 {
			TemplateStruct := &Template{}
			TemplateStruct.ObjType = "Templates"
			TemplateStruct.TemplateID = data["urn"].(string)
			TemplateStruct.PEID = data["peid"].(string)
			TemplateStruct.CLI = cli
			TemplateStruct.TemplateName = data["tname"].(string)
			TemplateStruct.TemplateType = data["ttyp"].(string)
			TemplateStruct.CommunicationType = data["ctyp"].(string)
			TemplateStruct.TempContent = data["tcont"].(string)
			TemplateStruct.Creator = Organizations[0]
			TemplateStruct.CreateTs = data["cts"].(string)
			TemplateStruct.UpdatedBy = Organizations[0]
			TemplateStruct.UpdateTs = data["uts"].(string)
			TemplateStruct.Status = data["sts"].(string)
			TemplateStruct.TMID = data["tmid"].(string)
			if data["ttyp"].(string) == "CS" {
				TemplateStruct.ConsentTemplateType = data["csty"].(string)
			}
			if data["ttyp"].(string) == "CT" {
				TemplateStruct.Category = data["ctgr"].(string)
				TemplateStruct.ContentType = data["coty"].(string)
				TemplateStruct.NoOfVariables = data["vars"].(string)
			}
			logger.Infof("TemplateID " + TemplateStruct.TemplateID + "Template peid " + TemplateStruct.PEID + "Template Name" + TemplateStruct.TemplateName)
			TemplateAsBytes, err := json.Marshal(TemplateStruct)
			if err != nil {
				logger.Errorf("setTemplate : Marshalling Error : " + string(err.Error()))
				return shim.Error("setTemplate : Marshalling Error : " + string(err.Error()))
			}
			//Inserting DataBlock to BlockChain
			err = stub.PutState(TemplateStruct.TemplateID, TemplateAsBytes)
			if err != nil {
				logger.Errorf("setTemplate : PutState Failed Error : " + string(err.Error()))
				return shim.Error("setTemplate : PutState Failed Error : " + string(err.Error()))
			}
			logger.Infof("setTemplate : PutState Success : " + string(TemplateAsBytes))
			//Txid := stub.GetTxID()
			eventbytes := Event{Data: string(TemplateAsBytes), Txid: stub.GetTxID()}
			payload, err := json.Marshal(eventbytes)
			if err != nil {
				logger.Errorf("setTemplate : Event Payload Marshalling Error : " + string(err.Error()))
				return shim.Error("setTemplate : Event Payload Marshalling Error : " + string(err.Error()))
			}
			err2 := stub.SetEvent(EVTADDTEMPLATE, payload)
			if err2 != nil {
				logger.Errorf("setTemplate : Event Creation Error for EventID : " + string(EVTADDTEMPLATE))
				return shim.Error("setTemplate : Event Creation Error for EventID : " + string(EVTADDTEMPLATE))
			}
			logger.Infof("setTemplate : Event Payload data : " + string(payload))
			txid := stub.GetTxID()

			resultData := map[string]interface{}{
				"trxnID":       txid,
				"PEID":         data["peid"].(string),
				"TemplateID":   data["urn"].(string),
				"TemplateName": data["tname"].(string),
				"message":      "Template created successfully",
				"TxnStatus":    "true"}
			respJSON, _ := json.Marshal(resultData)
			return shim.Success(respJSON)
		} else {
			jsonResp = "{\"urn\":\"value\",\"peid\":\"value\",\"cli\":\"value\",\"tname\":\"value\",\"ttyp\":\"value\",\"ctyp\":\"value\",\"ctgr\":\"value\",\"tcont\":\"value\",\"cts\":\"value\",\"uts\":\"value\",\"sts\":\"value\"}"
			logger.Errorf("setTemplate : Incorrect Number Of Arguments, Expected json structure : " + string(jsonResp))
			return shim.Error("setTemplate : Incorrect Number Of Arguments, Expected json structure : " + string(jsonResp))
		}
	}
}

//======================================================
//batchTemplates for Uploading Bulk Templates into DL
//=======================================================
func (dlt *TemplateMgmtChaincode) addBatchTemplates(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var result []string
	var errorCount int
	errorCount = 0
	if len(args) == 0 {
		logger.Errorf("batchTemplates : Input Argument should not be empty ")
		return shim.Error("batchTemplates : Input Argument should not be empty")
	}
	for i := 0; i < len(args); i++ {
		var errorCheck int
		var data map[string]interface{}
		errorCheck = 0
		logger.Infof(args[i])
		err := json.Unmarshal([]byte(args[i]), &data)
		if err != nil {
			logger.Errorf("batchTemplates : Input arguments unmarhsaling Error : " + string(err.Error()))
			return shim.Error("batchTemplates : Input arguments unmarhsaling Error : " + string(err.Error()))
		}
		certData, err := cid.GetX509Certificate(stub)
		if err != nil {
			logger.Errorf("batchTemplates : Getting certificate Details Error : " + string(err.Error()))
			return shim.Error("batchTemplates : Getting certificate Details Error : " + string(err.Error()))
		}
		Organizations := certData.Issuer.Organization

		if _, flag := data["urn"].(string); !flag {
			jsonResp = "{\"Error\":\"urn is empty\"}"
			return shim.Error(jsonResp)
		}
		if _, flag := data["peid"].(string); !flag {
			jsonResp = "{\"Error\":\"peid is empty\"}"
			return shim.Error(jsonResp)
		}
		if _, flag := data["tname"].(string); !flag {
			jsonResp = "{\"Error\":\"tname is empty\"}"
			return shim.Error(jsonResp)
		}
		if _, flag := data["ttyp"].(string); !flag {
			jsonResp = "{\"Error\":\"ttyp is empty\"}"
			return shim.Error(jsonResp)
		}
		if _, flag := data["ctyp"].(string); !flag {
			jsonResp = "{\"Error\":\"ctyp is empty\"}"
			return shim.Error(jsonResp)
		}
		if _, flag := data["tcont"].(string); !flag {
			jsonResp = "{\"Error\":\"tcont is empty\"}"
			return shim.Error(jsonResp)
		}
		if _, flag := data["uts"].(string); !flag {
			jsonResp = "{\"Error\":\"uts is empty\"}"
			return shim.Error(jsonResp)
		}
		if _, flag := data["sts"].(string); !flag {
			jsonResp = "{\"Error\":\"sts is empty\"}"
			return shim.Error(jsonResp)
		}
		if _, flag := data["cli"].([]interface{}); !flag {
			jsonResp = "{\"Error\":\"cli is empty\"}"
			return shim.Error(jsonResp)
		}
		cliList := data["cli"].([]interface{})
		cli := make([]string, len(cliList))
		for i, v := range cliList {
			if _, flag := v.(string); !flag {
				jsonResp = "{\"Error\":\"Header data should be string\"}"
				return shim.Error(jsonResp)
			}
			cli[i] = fmt.Sprint(v)
		}

		if !validEnumEntry(data["ttyp"].(string), tempType) {
			out := Output{Data: data["ttyp"].(string), ErrorDetails: "Please enter one of these value for TemplateType 'CS' or 'CT' "}
			edata, err := json.Marshal(out)
			if err != nil {
				logger.Errorf("batchTemplates : Marshalling Error : " + string(err.Error()))
				return shim.Error("batchTemplates : Marshalling Error : " + string(err.Error()))
			}
			result = append(result, string(edata))
			errorCount = errorCount + 1
			errorCheck = errorCheck + 1

		}

		if !validEnumEntry(data["ctyp"].(string), communicationType) {
			out := Output{Data: data["ctyp"].(string), ErrorDetails: "Please enter one of these value for communicationType 'p' or 'T' or 'SE', or 'SI' "}
			edata, err := json.Marshal(out)
			if err != nil {
				logger.Errorf("batchTemplates : Marshalling Error : " + string(err.Error()))
				return shim.Error("batchTemplates : Marshalling Error : " + string(err.Error()))
			}
			result = append(result, string(edata))
			errorCount = errorCount + 1
			errorCheck = errorCheck + 1
		}

		if !validEnumEntry(data["sts"].(string), status) {

			out := Output{Data: data["sts"].(string), ErrorDetails: "Please enter one of these value for Status 'A' or 'I' "}
			edata, err := json.Marshal(out)
			if err != nil {
				logger.Errorf("batchTemplates : Marshalling Error : " + string(err.Error()))
				return shim.Error("batchTemplates : Marshalling Error : " + string(err.Error()))
			}
			result = append(result, string(edata))
			errorCount = errorCount + 1
			errorCheck = errorCheck + 1
		}

		if _, err := strconv.Atoi(data["urn"].(string)); err != nil {
			out := Output{Data: data["urn"].(string), ErrorDetails: "URN should contains Only Numeric Characters,It is Not Numeric"}
			edata, err := json.Marshal(out)
			if err != nil {
				logger.Errorf("batchTemplates : Marshalling Error : " + string(err.Error()))
				return shim.Error("batchTemplates : Marshalling Error : " + string(err.Error()))
			}
			result = append(result, string(edata))
			errorCount = errorCount + 1
			errorCheck = errorCheck + 1
		}

		if errorCheck != 0 {
			continue
		}

		value, err := stub.GetState(data["urn"].(string))
		if err != nil {
			logger.Errorf("batchTemplates : GetState Failed for Template : " + data["urn"].(string) + " , Error : " + string(err.Error()))
			return shim.Error("batchTemplates : GetState Failed for Template : " + data["urn"].(string) + " , Error : " + string(err.Error()))
		}

		if value != nil {
			logger.Errorf("batchTemplates : Template already exists. Please choose unique TemplateID")
			out := Output{Data: data["urn"].(string), ErrorDetails: "Template already exists. Please choose unique TemplateID"}
			edata, err := json.Marshal(out)
			if err != nil {
				logger.Errorf("batchTemplates : Marshalling Error : " + string(err.Error()))
				return shim.Error("batchTemplates : Marshalling Error : " + string(err.Error()))
			}
			result = append(result, string(edata))
			errorCount = errorCount + 1
			continue
		}

		if value == nil {
			if len(data) == 10 || len(data) == 11 {
				TemplateStruct := &Template{}
				TemplateStruct.ObjType = "Templates"
				TemplateStruct.TemplateID = data["urn"].(string)
				TemplateStruct.PEID = data["peid"].(string)
				TemplateStruct.CLI = cli
				TemplateStruct.TemplateName = data["tname"].(string)
				TemplateStruct.TemplateType = data["ttyp"].(string)
				TemplateStruct.CommunicationType = data["ctyp"].(string)
				if data["ttyp"].(string) == "CT" {
					TemplateStruct.Category = data["ctgr"].(string)
				}
				TemplateStruct.TempContent = data["tcont"].(string)
				TemplateStruct.Creator = Organizations[0]
				TemplateStruct.CreateTs = data["cts"].(string)
				TemplateStruct.UpdatedBy = Organizations[0]
				TemplateStruct.UpdateTs = data["uts"].(string)
				TemplateStruct.Status = data["sts"].(string)
				if data["ttyp"].(string) == "CS" {
					TemplateStruct.ConsentTemplateType = data["csty"].(string)
				}
				if data["ttyp"].(string) == "CT" {
					TemplateStruct.ContentType = data["coty"].(string)
					TemplateStruct.NoOfVariables = data["vars"].(string)
				}
				logger.Infof("Template is " + TemplateStruct.PEID + "-" + TemplateStruct.TemplateName)
				TemplateAsBytes, err := json.Marshal(TemplateStruct)
				if err != nil {
					logger.Errorf("batchTemplates : Marshalling Error : " + string(err.Error()))
					return shim.Error("batchTemplates : Marshalling Error : " + string(err.Error()))
				}
				//Inserting DataBlock to BlockChain
				err = stub.PutState(TemplateStruct.TemplateID, TemplateAsBytes)
				if err != nil {
					logger.Errorf("batchTemplates : PutState Failed Error : " + string(err.Error()))
					return shim.Error("batchTemplates : PutState Failed Error : " + string(err.Error()))
				}
				logger.Infof("batchTemplates : PutState Success : " + string(TemplateAsBytes))
				eventbytes := Event{Data: string(TemplateAsBytes), Txid: stub.GetTxID()}
				payload, err := json.Marshal(eventbytes)
				if err != nil {
					logger.Errorf("batchTemplates : Event Payload Marshalling Error :" + string(err.Error()))
					return shim.Error("batchTemplates : Event Payload Marshalling Error :" + string(err.Error()))
				}
				err2 := stub.SetEvent(EVTADDTEMPLATE, []byte(payload))
				if err2 != nil {
					logger.Errorf("batchTemplates : Event Creation Error for EventID : " + string(EVTADDTEMPLATE))
					return shim.Error("batchTemplates : Event Creation Error for EventID : " + string(EVTADDTEMPLATE))
				}
				logger.Infof("batchTemplates : Event Payload data : " + string(payload))
			} else {
				jsonResp = "{\"urn\":\"value\",\"peid\":\"value\",\"cli\":\"value\",\"tname\":\"value\",\"ttyp\":\"value\",\"ctyp\":\"value\",\"ctgr\":\"value\",\"tcont\":\"value\",\"cts\":\"value\",\"uts\":\"value\",\"sts\":\"value\"}"
				logger.Errorf("setTemplate : Incorrect Number Of Arguments, Expected json structure : " + string(jsonResp))
				out := Output{Data: args[i], ErrorDetails: "IncorrectNumberOfArGumentsExceptin[10 or 11 keys]"}
				edata, err := json.Marshal(out)
				if err != nil {
					logger.Errorf("batchTemplates : Marshalling Error : " + string(err.Error()))
					return shim.Error("batchTemplates : Marshalling Error : " + string(err.Error()))
				}
				result = append(result, string(edata))
				errorCount = errorCount + 1
				continue
			}
		}
	}
	logger.Infof("ErrorCount is " + string(errorCount))
	if errorCount == 0 {
		txid := stub.GetTxID()
		resultData := map[string]interface{}{
			"trxnID":    txid,
			"message":   "batchPreferences : Batch Preferences data added Successfully",
			"TxnStatus": "true"}
		respJSON, _ := json.Marshal(resultData)
		return shim.Success(respJSON)
	} else {
		response := strings.Join(result, "|")
		return shim.Success([]byte("batchTemplates : Updating batch Error : " + string(response)))
	}

}

//=============================================================================================================
//updateTemplateStatus for Updating Template status in DL based on PE and Template Name on successful PE check
//==============================================================================================================

func (dlt *TemplateMgmtChaincode) updateTemplateStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 3 {
		logger.Errorf("updateTemplateStatus : Incorrect Number Of Arguments: TemplateID, Status and update timestamp are Expected.")
		return shim.Error("updateTemplateStatus : Incorrect Number Of Arguments: TemplateID, Status and update timestamp are Expected.")
	}

	if !validEnumEntry(args[1], status) {
		jsonResp = "{\"Error\":\"Please enter one of these value for Status 'A' or 'I' \"}"
		return shim.Error(jsonResp)
	}

	if _, err := strconv.Atoi(args[0]); err != nil {
		jsonResp = "{\"Error\":\"URN should contains Only Numeric Characters,It is Not Numeric\"}"
		return shim.Error(jsonResp)
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		logger.Errorf("updateTemplateStatus : GetState Failed for TemplateID : " + string(args[0]) + " , Error : " + string(err.Error()))
		return shim.Error("updateTemplateStatus : GetState Failed for TemplateID :" + string(args[0]) + " , Error : " + string(err.Error()))
	}
	if value == nil {
		logger.Info("updateTemplateStatus : No Existing Templates for TemplateID : " + string(args[0]))
		return shim.Error("updateTemplateStatus : No Existing Templates for TemplateID : " + string(args[0]))
	} else {
		var organizationName string
		var orgName string
		Template := Template{}
		err := json.Unmarshal(value, &Template)
		if err != nil {
			logger.Errorf("updateTemplateStatus : Unmarshaling Error : " + string(err.Error()))
			return shim.Error("updateTemplateStatus : Unmarshaling Error : " + string(err.Error()))
		}
		certData, err := cid.GetX509Certificate(stub)
		if err != nil {
			logger.Errorf("updateTemplateStatus : Getting certificate Details Error : " + string(err.Error()))
			return shim.Error("updateTemplateStatus : Getting certificate Details Error : " + string(err.Error()))
		}
		Organizations := certData.Issuer.Organization
		orgName = Template.Creator
		organizationName = Organizations[0]

		if strings.Compare(orgName, organizationName) == 0 {

			Template.Status = args[1]
			Template.UpdatedBy = organizationName
			Template.UpdateTs = args[2]

			TempAsBytes, err := json.Marshal(Template)
			if err != nil {
				logger.Errorf("batchTemplates : Marshalling Error : " + string(err.Error()))
				return shim.Error("batchTemplates : Marshalling Error : " + string(err.Error()))
			}
			//Inserting DataBlock to BlockChain
			err = stub.PutState(Template.TemplateID, TempAsBytes)

			if err != nil {
				logger.Error("updateTemplateStatus : Update Template error for TemplateID " + string(args[0]) + " , Error : " + string(err.Error()))
				return shim.Error("updateTemplateStatus : Update Template error for TemplateID " + string(args[0]) + " , Error : " + string(err.Error()))
			}
			eventbytes := Event{Data: string(args[0]) + "-" + string(args[1]+"-"+string(args[2])), Txid: stub.GetTxID()}
			payload, err := json.Marshal(eventbytes)
			if err != nil {
				logger.Errorf("updateTemplateStatus : Event Payload Marshalling Error : " + string(err.Error()))
				return shim.Error("updateTemplateStatus : Event Payload Marshalling Error : " + string(err.Error()))
			}
			err2 := stub.SetEvent(EVTUPDTEMPLATE, []byte(payload))
			if err2 != nil {
				logger.Errorf("updateTemplateStatus : Event Creation Error for EventID : " + string(EVTUPDTEMPLATE))
				return shim.Error("updateTemplateStatus : Event Creation Error for EventID : " + string(EVTUPDTEMPLATE))
			}
			logger.Infof("updateTemplateStatus : Event Payload Data : " + string(args[0]) + "-" + string(args[1]+"-"+string(args[2])))

			resultData := map[string]interface{}{
				"trxnID":     stub.GetTxID(),
				"TemplateID": args[0],
				"Status":     args[1],
				"UpdateTs":   args[2],
				"message":    "Template updated successfully",
				"TxnStatus":  "true"}
			respJSON, _ := json.Marshal(resultData)
			return shim.Success(respJSON)

		} else {
			logger.Errorf("Unauthorized access")
			return shim.Error("Unauthorized access")
		}
	}
}

//======================================================================================
//queryTemplates RichQuery for Obtaining Template data
//======================================================================================

func (dlt *TemplateMgmtChaincode) queryTemplates(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("queryTemplates : Incorrect number of arguments, Expected 1 [Query String]")
	}
	queryString := args[0]
	logger.Info(args[0])
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		logger.Errorf("queryTemplates : getQueryResultForQueryString Failed Error : " + string(err.Error()))
		return shim.Error("queryTemplates : getQueryResultForQueryString Failed Error : " + string(err.Error()))
	}
	return shim.Success(queryResults)
}

func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

//======================================================================================
//queryTemplates RichQuery for Obtaining Template data
//======================================================================================

func (dlt *TemplateMgmtChaincode) queryTemplateHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var err error
	var TemplateID = args[0]
	var TemplateHistoryData []Template
	var TemplateHis Template

	if len(args) != 1 {
		return shim.Error("TemplateHistory : Incorrect number of arguments, Expected 1 arg")
	}

	historyResponse, err := stub.GetHistoryForKey(TemplateID)
	if err != nil {
		logger.Errorf("TemplateHistory : query Template history - Failed Error : " + string(err.Error()))
		return shim.Error("TemplateHistory : query Template history - Failed Error : " + string(err.Error()))
	}

	for historyResponse.HasNext() {
		iterratorResp, err := historyResponse.Next()
		if err != nil {
			logger.Errorf("Error on iterating : " + string(err.Error()))
			return shim.Error("Error on iterating : " + string(err.Error()))
		}
		value := iterratorResp.GetValue()

		iErr := json.Unmarshal(value, &TemplateHis)
		if iErr != nil {
			logger.Errorf("TemplateHistory : Error on unmarshalling : " + string(iErr.Error()))
			return shim.Error("TemplateHistory : Error on unmarshalling - Failed Error : " + string(iErr.Error()))
		}

		TemplateHistoryData = append(TemplateHistoryData, TemplateHis)

	}

	JSONBytes, err := json.Marshal(TemplateHistoryData)

	return shim.Success(JSONBytes)
}

//========================================================================================================
//queryTemplateWithPagination RichQuery for Obtaining Template data with pagination for more records
//========================================================================================================
func (dlt *TemplateMgmtChaincode) queryTemplateWithPagination(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 3 {
		return shim.Error("queryTemplateWithPagination : Incorrect number of arguments, Expected 3 arg")
	}

	queryString := args[0]
	pageSize, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		logger.Errorf("queryTemplateWithPagination : Parsing error : " + string(err.Error()))
		return shim.Error("queryTemplateWithPagination : Parsing error : " + string(err.Error()))
	}
	bookmark := args[2]

	queryResults, err := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookmark)
	if err != nil {
		logger.Errorf("queryTemplateWithPagination : pagination process : " + string(err.Error()))
		return shim.Error("queryTemplateWithPagination : pagination process : " + string(err.Error()))
	}
	return shim.Success(queryResults)
}

//====================================================================================================================
//getQueryResultForQueryStringWithPagination RichQuery for Obtaining Template data with pagination for more records
//====================================================================================================================
func getQueryResultForQueryStringWithPagination(stub shim.ChaincodeStubInterface, queryString string, pageSize int32, bookmark string) ([]byte, error) {

	resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(queryString, pageSize, bookmark)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	bufferWithPaginationInfo := addPaginationMetadataToQueryResults(buffer, responseMetadata)

	logger.Infof("-- getQueryResultForQueryString queryResult -- : " + bufferWithPaginationInfo.String())

	return buffer.Bytes(), nil
}

// ===========================================================================================
// addPaginationMetadataToQueryResults constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================
func addPaginationMetadataToQueryResults(buffer *bytes.Buffer, responseMetadata *pb.QueryResponseMetadata) *bytes.Buffer {

	buffer.WriteString("[{\"ResponseMetadata\":{\"RecordsCount\":")
	buffer.WriteString("\"")
	buffer.WriteString(string(responseMetadata.FetchedRecordsCount))
	buffer.WriteString("\"")
	buffer.WriteString(", \"Bookmark\":")
	buffer.WriteString("\"")
	buffer.WriteString(responseMetadata.Bookmark)
	buffer.WriteString("\"}}]")

	return buffer
}

// ===========================================================================================
// constructQueryResponseFromIterator constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return &buffer, nil
}

// ===================================================================================
//main function for the Template ChainCode
// ===================================================================================
func main() {
	err := shim.Start(new(TemplateMgmtChaincode))
	logger.SetLevel(shim.LogDebug)
	if err != nil {
		logger.Error("Error Starting TemplateMgmtChaincode Chaincode is " + string(err.Error()))
	} else {
		logger.Info("Starting TemplateMgmtChaincode Chaincode")
	}
}
