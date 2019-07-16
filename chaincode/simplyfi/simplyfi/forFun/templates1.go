/*
This Chaincode is written for storing,retrieving,deleting the Content/Consent Templates that are stored in DLT
*/

package main

import (
	"encoding/json" //reading and writing JSON
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"             // import for Chaincode Interface
	cid "github.com/hyperledger/fabric/core/chaincode/shim/ext/cid" // import for Client Identity
	pb "github.com/hyperledger/fabric/protos/peer"                  // import for peer response
)

//Logger for Logging
var logger = shim.NewLogger("TEMPLATES")
//Event Names
const _AddTemplate = "ADD_TEMPLATE"
const _UpdateTemplate = "UPDATE_TEMPLATE"
const _BulkTemplate = "BULK_TEMPLATE"


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
var communicationTypeForCT = map[string]bool{
	"P":  true,
	"T":  true,
	"SE": true,
	"SI": true,
}
var communicationTypeForCS = map[string]bool{
	"SE": true,
}

//Status
var status = map[string]bool{
	"A": true,
	"I": true,
}

//content Type
var contentType = map[string]bool{
	"T": true,
	"U": true,
}

//content Type
var consentTemplateType = map[string]bool{
	"1": true,
	"2": true,
	"3": true,
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
	ConsentTemplateType string   `json:"csty"`
	Contenttype         string   `json:"coty"`
	NoOfVariables       string   `json:"vars"`
	Category            string   `json:"ctgr"`
	TempContent         string   `json:"tcont"`
	TMID                string   `json:"tmid"`
	Creator             string   `json:"crtr"`
	CreateTs            string   `json:"cts"`
	UpdatedBy           string   `json:"uby"`
	UpdateTs            string   `json:"uts"`
	Status              string   `json:"sts"`
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
		return dlt.queryTemplatesHistory(stub, args)
	case "qtp": //Rich Query to retrieve the Templates with pagination from DL
		return dlt.queryTemplatesWithPagination(stub, args)
	case "gt": //get Template data based on TemplateID
		return dlt.getTemplateByTemplateID(stub, args)
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
	//madatory field validation
	if _, flag := data["urn"].(string); !flag {
		jsonResp = "{\"Error\":\"urn is empty\"}"
		logger.Errorf("batchTemplates:" + string(jsonResp))
		return shim.Error(jsonResp)
	}
	if _, flag := data["peid"].(string); !flag {
		jsonResp = "{\"Error\":\"peid is empty\"}"
		logger.Errorf("batchTemplates:" + string(jsonResp))
		return shim.Error(jsonResp)
	}
	if _, flag := data["tname"].(string); !flag {
		jsonResp = "{\"Error\":\"tname is empty\"}"
		logger.Errorf("batchTemplates:" + string(jsonResp))
		return shim.Error(jsonResp)
	}
	if _, flag := data["ttyp"].(string); !flag {
		jsonResp = "{\"Error\":\"ttyp is empty\"}"
		logger.Errorf("batchTemplates:" + string(jsonResp))
		return shim.Error(jsonResp)
	}
	if _, flag := data["tcont"].(string); !flag {
		jsonResp = "{\"Error\":\"tcont is empty\"}"
		logger.Errorf("batchTemplates:" + string(jsonResp))
		return shim.Error(jsonResp)
	}
	if _, flag := data["cts"].(string); !flag {
		jsonResp = "{\"Error\":\"cts is empty\"}"
		logger.Errorf("batchTemplates:" + string(jsonResp))
		return shim.Error(jsonResp)
	}
	if _, flag := data["uts"].(string); !flag {
		jsonResp = "{\"Error\":\"uts is empty\"}"
		logger.Errorf("batchTemplates:" + string(jsonResp))
		return shim.Error(jsonResp)
	}
	if _, flag := data["sts"].(string); !flag {
		jsonResp = "{\"Error\":\"sts is empty\"}"
		logger.Errorf("batchTemplates:" + string(jsonResp))
		return shim.Error(jsonResp)
	}
	if _, flag := data["cli"].([]interface{}); !flag {
		jsonResp = "{\"Error\":\"cli is empty\"}"
		logger.Errorf("batchTemplates:" + string(jsonResp))
		return shim.Error(jsonResp)
	}
	cliList := data["cli"].([]interface{})
	cli := make([]string, len(cliList))
	for i, v := range cliList {
		if len(v.(string)) == 0 {
			jsonResp = "{\"Error\":\"Header data should be string\"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			return shim.Error(jsonResp)
		}
		cli[i] = fmt.Sprint(v)
	}
	if len(cli) == 0 {
		jsonResp = "{\"Error\":\"cli is empty\"}"
		logger.Errorf("batchTemplates:" + string(jsonResp))
		return shim.Error(jsonResp)
	}
	if !validEnumEntry(data["ttyp"].(string), tempType) {
		jsonResp = "{\"Error\":\"Please enter one of these value for TemplateType 'CS' or 'CT' \"}"
		logger.Errorf("batchTemplates:" + string(jsonResp))
		return shim.Error(jsonResp)
	}

	if !validEnumEntry(data["sts"].(string), status) {
		jsonResp = "{\"Error\":\"Please enter one of these value for Status 'A' or 'I' \"}"
		logger.Errorf("batchTemplates:" + string(jsonResp))
		return shim.Error(jsonResp)
	}

	if data["ttyp"].(string) == "CT" {
		if _, flag := data["ctgr"].(string); !flag {
			jsonResp = "{\"Error\":\"ctgr is empty\"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			return shim.Error(jsonResp)
		}
		if len(data["ctgr"].(string)) == 0 {
			jsonResp = "{\"Error\":\" Input Data: Category is empty; for Template Type 'Content Template' category is mandatory\"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			return shim.Error(jsonResp)
		}
		if _, flag := data["ctyp"].(string); !flag {
			jsonResp = "{\"Error\":\"ctyp is empty\"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			return shim.Error(jsonResp)
		}
		if !validEnumEntry(data["ctyp"].(string), communicationTypeForCT) {
			jsonResp = "{\"Error\":\"Please enter one of these value for communicationType 'p','T','SE' or 'SI' \"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			return shim.Error(jsonResp)
		}
		if _, err := strconv.Atoi(data["ctgr"].(string)); err != nil {
			jsonResp = "{\"Error\":\"category is not numeric \"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			return shim.Error(jsonResp)
		}
	}
	if data["ttyp"].(string) == "CS" {
		if _, flag := data["csty"].(string); !flag {
			jsonResp = "{\"Error\":\"csty is empty\"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			return shim.Error(jsonResp)
		}
		if !validEnumEntry(data["csty"].(string), consentTemplateType) {
			jsonResp = "{\"Error\":\"Please enter one of these value for consent template type '1' or '2' or '3' \"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			return shim.Error(jsonResp)
		}
		if _, flag := data["ctyp"].(string); !flag {
			jsonResp = "{\"Error\":\"ctyp is empty\"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			return shim.Error(jsonResp)
		}
		if !validEnumEntry(data["ctyp"].(string), communicationTypeForCS) {
			jsonResp = "{\"Error\":\"Please enter one of these value for communicationType 'SE' \"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			return shim.Error(jsonResp)
		}
	}
	if _, err := strconv.Atoi(data["urn"].(string)); err != nil {
		jsonResp = "{\"Error\":\"URN is not numeric \"}"
		logger.Errorf("batchTemplates:" + string(jsonResp))
		return shim.Error(jsonResp)
	}

	//--------
	Organizations := certData.Issuer.Organization
	//check template is already exist with same templateid
	value, err := stub.GetState(data["urn"].(string))

	if err != nil {
		logger.Errorf("setTemplate : GetState Failed for TemplateID : " + data["urn"].(string) + " Error : " + string(err.Error()))
		return shim.Error("setTemplate : GetState Failed for TemplateID : " + data["urn"].(string) + " Error : " + string(err.Error()))
	}

	if value != nil {
		logger.Errorf("setTemplate : Template already exists. Please choose unique TemplateID")
		return shim.Error("setTemplate : Template already exists. Please choose unique TemplateID")
	} else {
		TemplateStruct := &Template{}
		TemplateStruct.ObjType = "Templates"
		TemplateStruct.TemplateID = data["urn"].(string)
		TemplateStruct.PEID = data["peid"].(string)
		TemplateStruct.CLI = cli
		TemplateStruct.TemplateName = data["tname"].(string)
		TemplateStruct.TemplateType = data["ttyp"].(string)
		TemplateStruct.CommunicationType = data["ctyp"].(string)
		if _, flag := data["ctgr"].(string); flag {
			TemplateStruct.Category = data["ctgr"].(string)
		}
		if _, flag := data["coty"].(string); flag {
			TemplateStruct.Contenttype = data["coty"].(string)
		}
		if _, flag := data["csty"].(string); flag {
			TemplateStruct.ConsentTemplateType = data["csty"].(string)
		}
		if _, flag := data["vars"].(string); flag {
			TemplateStruct.NoOfVariables = data["vars"].(string)
		}
		TemplateStruct.TempContent = data["tcont"].(string)
		TemplateStruct.Creator = Organizations[0]
		if _, flag := data["tmid"].(string); flag {
			TemplateStruct.TMID = data["tmid"].(string)
		}
		TemplateStruct.CreateTs = data["cts"].(string)
		TemplateStruct.UpdatedBy = Organizations[0]
		TemplateStruct.UpdateTs = data["uts"].(string)
		TemplateStruct.Status = data["sts"].(string)
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
	}
}

//======================================================
//batchTemplates for Uploading Bulk Templates into DL
//=======================================================

func (dlt *TemplateMgmtChaincode) addBatchTemplates(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var failed_urn []string
	if len(args) == 0 {
		logger.Errorf("batchTemplates : Input Argument should not be empty")
		return shim.Error("batchTemplates : Input Argument should not be empty")
	}
	certData, err := cid.GetX509Certificate(stub)
	if err != nil {
		logger.Errorf("batchTemplates : Getting certificate Details Error : " + string(err.Error()))
		return shim.Error("batchTemplates : Getting certificate Details Error : " + string(err.Error()))
	}
	for i := 0; i < len(args); i++ {
		var data map[string]interface{}
		logger.Infof(args[i])
		err := json.Unmarshal([]byte(args[i]), &data)
		if err != nil {
			logger.Errorf("batchTemplates : Input arguments unmarhsaling Error : " + string(err.Error()))
			//return shim.Error("batchTemplates : Input arguments unmarhsaling Error : " + string(err.Error()))
			failed_urn = append(failed_urn, "Input argument unmarshaling error")
			continue
		}
		Organizations := certData.Issuer.Organization
		//madatory field validation
		if _, flag := data["urn"].(string); !flag {
			jsonResp = "{\"Error\":\"urn is empty\"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			failed_urn = append(failed_urn, "urn is empty")
			continue
		}
		if _, flag := data["peid"].(string); !flag {
			jsonResp = "{\"Error\":\"peid is empty\"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			failed_urn = append(failed_urn, data["urn"].(string))
			continue
		}
		if _, flag := data["tname"].(string); !flag {
			jsonResp = "{\"Error\":\"tname is empty\"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			failed_urn = append(failed_urn, data["urn"].(string))
			continue
		}
		if _, flag := data["ttyp"].(string); !flag {
			jsonResp = "{\"Error\":\"ttyp is empty\"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			failed_urn = append(failed_urn, data["urn"].(string))
			continue
		}
		if _, flag := data["tcont"].(string); !flag {
			jsonResp = "{\"Error\":\"tcont is empty\"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			failed_urn = append(failed_urn, data["urn"].(string))
			continue
		}
		if _, flag := data["cts"].(string); !flag {
			jsonResp = "{\"Error\":\"cts is empty\"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			failed_urn = append(failed_urn, data["urn"].(string))
			continue
		}
		if _, flag := data["uts"].(string); !flag {
			jsonResp = "{\"Error\":\"uts is empty\"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			failed_urn = append(failed_urn, data["urn"].(string))
			continue
		}
		if _, flag := data["sts"].(string); !flag {
			jsonResp = "{\"Error\":\"sts is empty\"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			failed_urn = append(failed_urn, data["urn"].(string))
			continue
		}
		if _, flag := data["cli"].([]interface{}); !flag {
			jsonResp = "{\"Error\":\"cli is empty\"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			failed_urn = append(failed_urn, data["urn"].(string))
			continue
		}
		cliList := data["cli"].([]interface{})
		cli := make([]string, len(cliList))
		for i, v := range cliList {
			if len(v.(string)) == 0 {
				jsonResp = "{\"Error\":\"Header data should be string\"}"
				logger.Errorf("batchTemplates:" + string(jsonResp))
			}
			cli[i] = fmt.Sprint(v)
		}
		if len(cli) == 0 {
			failed_urn = append(failed_urn, data["urn"].(string))
			continue
		}
		if !validEnumEntry(data["ttyp"].(string), tempType) {
			jsonResp = "{\"Error\":\"Please enter one of these value for TemplateType 'CS' or 'CT' \"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			failed_urn = append(failed_urn, data["urn"].(string))
			continue
		}

		if !validEnumEntry(data["sts"].(string), status) {
			jsonResp = "{\"Error\":\"Please enter one of these value for Status 'A' or 'I' \"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			failed_urn = append(failed_urn, data["urn"].(string))
			continue
		}

		if data["ttyp"].(string) == "CT" {
			if _, flag := data["ctgr"].(string); !flag {
				jsonResp = "{\"Error\":\"ctgr is empty\"}"
				logger.Errorf("batchTemplates:" + string(jsonResp))
				failed_urn = append(failed_urn, data["urn"].(string))
				continue
			}
			if len(data["ctgr"].(string)) == 0 {
				jsonResp = "{\"Error\":\" Input Data: Category is empty; for Template Type 'Content Template' category is mandatory\"}"
				logger.Errorf("batchTemplates:" + string(jsonResp))
				failed_urn = append(failed_urn, data["urn"].(string))
				continue
			}
			if _, flag := data["ctyp"].(string); !flag {
				jsonResp = "{\"Error\":\"ctyp is empty\"}"
				logger.Errorf("batchTemplates:" + string(jsonResp))
				failed_urn = append(failed_urn, data["urn"].(string))
				continue
			}
			if !validEnumEntry(data["ctyp"].(string), communicationTypeForCT) {
				jsonResp = "{\"Error\":\"Please enter one of these value for communicationType 'p','T','SE' or 'SI' \"}"
				logger.Errorf("batchTemplates:" + string(jsonResp))
				failed_urn = append(failed_urn, data["urn"].(string))
				continue
			}
			if _, err := strconv.Atoi(data["ctgr"].(string)); err != nil {
				jsonResp = "{\"Error\":\"category is not numeric \"}"
				logger.Errorf("batchTemplates:" + string(jsonResp))
				failed_urn = append(failed_urn, data["urn"].(string))
				continue
			}
		}
		if data["ttyp"].(string) == "CS" {
			if _, flag := data["csty"].(string); !flag {
				jsonResp = "{\"Error\":\"csty is empty\"}"
				logger.Errorf("batchTemplates:" + string(jsonResp))
				failed_urn = append(failed_urn, data["urn"].(string))
				continue
			}
			if !validEnumEntry(data["csty"].(string), consentTemplateType) {
				jsonResp = "{\"Error\":\"Please enter one of these value for consent template type '1' or '2' or '3' \"}"
				logger.Errorf("batchTemplates:" + string(jsonResp))
				failed_urn = append(failed_urn, data["urn"].(string))
				continue
			}
			if _, flag := data["ctyp"].(string); !flag {
				jsonResp = "{\"Error\":\"ctyp is empty\"}"
				logger.Errorf("batchTemplates:" + string(jsonResp))
				failed_urn = append(failed_urn, data["urn"].(string))
				continue
			}
			if !validEnumEntry(data["ctyp"].(string), communicationTypeForCS) {
				jsonResp = "{\"Error\":\"Please enter one of these value for communicationType 'SE' \"}"
				logger.Errorf("batchTemplates:" + string(jsonResp))
				failed_urn = append(failed_urn, data["urn"].(string))
				continue
			}
		}

		if _, err := strconv.Atoi(data["urn"].(string)); err != nil {
			jsonResp = "{\"Error\":\"URN is not numeric \"}"
			logger.Errorf("batchTemplates:" + string(jsonResp))
			failed_urn = append(failed_urn, data["urn"].(string))
			continue
		}

		value, err := stub.GetState(data["urn"].(string))
		if err != nil {
			logger.Errorf("batchTemplates : GetState Failed for Template : " + data["urn"].(string) + " , Error : " + string(err.Error()))
			failed_urn = append(failed_urn, data["urn"].(string))
			continue
		}

		if value != nil {
			logger.Errorf("batchTemplates : Template already exists. Please choose unique TemplateID")
			failed_urn = append(failed_urn, data["urn"].(string))
			continue
		} else {
			TemplateStruct := &Template{}
			TemplateStruct.ObjType = "Templates"
			TemplateStruct.TemplateID = data["urn"].(string)
			TemplateStruct.PEID = data["peid"].(string)
			TemplateStruct.CLI = cli
			TemplateStruct.TemplateName = data["tname"].(string)
			TemplateStruct.TemplateType = data["ttyp"].(string)
			TemplateStruct.CommunicationType = data["ctyp"].(string)
			if _, flag := data["ctgr"].(string); flag {
				TemplateStruct.Category = data["ctgr"].(string)
			}

			if _, flag := data["coty"].(string); flag {
				TemplateStruct.Contenttype = data["coty"].(string)
			}

			if _, flag := data["csty"].(string); flag {
				TemplateStruct.ConsentTemplateType = data["csty"].(string)
			}
			if _, flag := data["vars"].(string); flag {
				TemplateStruct.NoOfVariables = data["vars"].(string)
			}
			TemplateStruct.TempContent = data["tcont"].(string)
			TemplateStruct.Creator = Organizations[0]
			if _, flag := data["tmid"].(string); flag {
				TemplateStruct.TMID = data["tmid"].(string)
			}
			TemplateStruct.CreateTs = data["cts"].(string)
			TemplateStruct.UpdatedBy = Organizations[0]
			TemplateStruct.UpdateTs = data["uts"].(string)
			TemplateStruct.Status = data["sts"].(string)

			logger.Infof("Template is " + TemplateStruct.PEID + "-" + TemplateStruct.TemplateName)
			TemplateAsBytes, err := json.Marshal(TemplateStruct)
			if err != nil {
				logger.Errorf("batchTemplates : Marshalling Error : " + string(err.Error()))
				failed_urn = append(failed_urn, data["urn"].(string))
				continue
			}
			//Inserting DataBlock to BlockChain
			err = stub.PutState(TemplateStruct.TemplateID, TemplateAsBytes)
			if err != nil {
				logger.Errorf("batchTemplates : PutState Failed Error : " + string(err.Error()))
				failed_urn = append(failed_urn, data["urn"].(string))
				continue
			}
			logger.Infof("batchTemplates : PutState Success : " + string(TemplateAsBytes))
			eventbytes := Event{Data: string(TemplateAsBytes), Txid: stub.GetTxID()}
			payload, err := json.Marshal(eventbytes)
			if err != nil {
				logger.Errorf("batchTemplates : Event Payload Marshalling Error :" + string(err.Error()))
				failed_urn = append(failed_urn, data["urn"].(string))
				continue
			}
			err2 := stub.SetEvent(EVTADDTEMPLATE, []byte(payload))
			if err2 != nil {
				logger.Errorf("batchTemplates : Event Creation Error for EventID : " + string(EVTADDTEMPLATE))
				failed_urn = append(failed_urn, data["urn"].(string))
				continue
			}
			logger.Infof("batchTemplates : Event Payload data : " + string(payload))
		}
	}
	resultData := map[string]interface{}{
		"trxnid":     stub.GetTxID(),
		"failed_urn": failed_urn,
		"message":    "Batch Template Success",
	}
	respJson, _ := json.Marshal(resultData)
	return shim.Success(respJson)
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

	if _, err := strconv.Atoi(args[0]); err != nil {
		jsonResp = "{\"Error\":\"URN should contains Only Numeric Characters,It is Not Numeric\"}"
		logger.Errorf("updateTemplateStatus : " + string(jsonResp))
		return shim.Error(jsonResp)
	}

	if !validEnumEntry(args[1], status) {
		jsonResp = "{\"Error\":\"Please enter one of these value for Status 'A' or 'I' \"}"
		logger.Errorf("updateTemplateStatus : " + string(jsonResp))
		return shim.Error(jsonResp)
	}

	if _, err := strconv.Atoi(args[2]); err != nil {
		jsonResp = "{\"Error\":\"UpdatedTs should contains Only Numeric Characters,It is Not Numeric\"}"
		logger.Errorf("updateTemplateStatus : " + string(jsonResp))
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
//queryTemplate RichQuery for Obtaining Template Obj
//======================================================================================
func (dlt *TemplateMgmtChaincode) queryTemplates(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		logger.Errorf("queryTemplate:Invalid number of arguments are provided for transaction")
		jsonResp = "{\"Error\":\"Invalid number of arguments are provided for transaction\"}"
		return shim.Error(jsonResp)
	}
	var records []Template
	queryString := args[0]
	logger.Infof("Query Selector : " + string(queryString))
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		logger.Errorf("queryTemplate:GetQueryResult is Failed with error :" + string(err.Error()))
		jsonResp = "{\"Error\":\"GetQueryResult is Failed with error- \"" + string(err.Error()) + "\"}"
		return shim.Error(jsonResp)
	}
	for resultsIterator.HasNext() {
		record := Template{}
		recordBytes, _ := resultsIterator.Next()
		if (string(recordBytes.Value)) == "" {
			continue
		}
		err = json.Unmarshal(recordBytes.Value, &record)
		if err != nil {
			logger.Errorf("queryTemplate:Unable to unmarshal Template retrieved :" + string(err.Error()))
			jsonResp = "{\"Error\":\"GetQueryResult is Failed with error- \"" + string(err.Error()) + "\"}"
			return shim.Error(jsonResp)
		}
		records = append(records, record)
	}
	resultData := map[string]interface{}{
		"status":     "true",
		"templates:": records,
	}
	respJson, _ := json.Marshal(resultData)
	return shim.Success(respJson)
}

//========================================================================================
//getTemplateByTemplateID for Getting template data based on templateid
//=======================================================================================-
func (dlt *TemplateMgmtChaincode) getTemplateByTemplateID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		logger.Errorf("getTemplateByTemplateID:Invalid Number of arguments are provided for transaction")
		jsonResp = "{\"Error\":\"Invalid number of arguments are provided for transaction\"}"
		return shim.Error(jsonResp)
	}
	var records []Template
	TemplateExist, err := stub.GetState(args[0])
	if err != nil {
		jsonResp = "{\"Error\":\"GetState is Failed with error- \"" + string(err.Error()) + "\"}"
		return shim.Error(jsonResp)
	}
	if TemplateExist == nil {
		logger.Errorf("getTemplateByTemplateID:No Existing Template for TemplateID-" + string(args[0]))
		jsonResp = "{\"Error\":\"No Existing Template for TemplateID- \"" + string(args[0]) + "\"}"
		return shim.Error(jsonResp)
	} else {
		template := Template{}
		err := json.Unmarshal(TemplateExist, &template)
		if err != nil {
			logger.Errorf("getTemplateByTemplateID::Existing Template unmarshalling Error" + string(err.Error()))
			jsonResp = "{\"Error\":\"Existing Template unmarshalling Error-\"" + string(err.Error()) + "\"}"
			return shim.Error(jsonResp)
		}
		records = append(records, template)
		resultData := map[string]interface{}{
			"status":    "true",
			"templates": records[0],
		}
		respJson, _ := json.Marshal(resultData)
		return shim.Success(respJson)
	}
}

//========================================================================================
//getHistoryQuery for Getting all history data for urn
//=======================================================================================-

func (dlt *TemplateMgmtChaincode) queryTemplatesHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		logger.Errorf("queryTemplatesHistory:Invalid number of arguments are provided for transaction")
		jsonResp = "{\"Error\":\"Invalid number of arguments are provided for transaction\"}"
		return shim.Error(jsonResp)
	}
	var records []Template
	resultsIterator, err := stub.GetHistoryForKey(args[0])
	if err != nil {
		logger.Errorf("queryTemplatesHistory:GetHistoryForKey is Failed" + string(err.Error()))
		jsonResp = "{\"Error\":\"getHistoryTemplate is Failed with error- \"" + string(err.Error()) + "\"}"
		return shim.Error(jsonResp)
	}
	for resultsIterator.HasNext() {
		record := Template{}
		recordBytes, _ := resultsIterator.Next()
		if string(recordBytes.Value) == "" {
			continue
		}
		err := json.Unmarshal(recordBytes.Value, &record)
		if err != nil {
			logger.Errorf("queryTemplatesHistory:Unable to unmarshal Template retrieved- " + string(err.Error()))
			jsonResp = "{\"Error\":\"queryTemplatesHistory is Failed with error- \"" + string(err.Error()) + "\"}"
			return shim.Error(jsonResp)
		}
		records = append(records, record)
	}
	resultData := map[string]interface{}{
		"status":    "true",
		"templates": records,
	}
	respJson, _ := json.Marshal(resultData)
	return shim.Success(respJson)
}

// ===== Example: Pagination with Ad hoc Rich Query ========================================================
// queryTemplatesWithPagination uses a query string, page size and a bookmark to perform a query
// for Template. Query string matching state database syntax is passed in and executed as is.
// The number of fetched records would be equal to or lesser than the specified page size.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryMarblesForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// Paginated queries are only valid for read only transactions.
// =========================================================================================
func (dlt *TemplateMgmtChaincode) queryTemplatesWithPagination(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 3 {
		logger.Errorf("queryTemplatesWithPagination:Invalid number of arguments provided for transaction")
		jsonResp = "{\"Error\":\"Invalid Number of argumnets provided for transaction\"}"
		return shim.Error(jsonResp)
	}
	var records []Template
	queryString := args[0]
	pageSize, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		logger.Errorf("queryTemplateWithPagination:Error while ParseInt is :" + string(err.Error()))
		jsonResp = "{\"Error\":\"PageSize ParseInt error- \"" + string(err.Error()) + "\"}"
		return shim.Error(jsonResp)
	}
	bookmark := args[2]
	resultsIterator, responseMetaData, err := stub.GetQueryResultWithPagination(queryString, int32(pageSize), bookmark)
	if err != nil {
		logger.Errorf("queryTemplateWithPagination:GetQueryResultWithPagination is Failed :" + string(err.Error()))
		jsonResp = "{\"Error\":\"GetQueryResultWithPagination is Failed- \"" + string(err.Error()) + "\"}"
		return shim.Error(jsonResp)
	}
	for resultsIterator.HasNext() {
		record := Template{}
		recordBytes, _ := resultsIterator.Next()
		if string(recordBytes.Value) == "" {
			continue
		}
		err := json.Unmarshal(recordBytes.Value, &record)
		if err != nil {
			logger.Errorf("getHistoryTemplate:Unable to unmarshal Template retrieved :" + string(err.Error()))
			jsonResp = "{\"Error\":\"Unable to unmarshal Template retrieved- \"" + string(err.Error()) + "\"}"
			return shim.Error(jsonResp)
		}
		records = append(records, record)
	}
	resultData := map[string]interface{}{
		"status":       "true",
		"templates":    records,
		"recordscount": responseMetaData.FetchedRecordsCount,
		"bookmark":     responseMetaData.Bookmark,
	}
	respJson, _ := json.Marshal(resultData)
	return shim.Success(respJson)
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
