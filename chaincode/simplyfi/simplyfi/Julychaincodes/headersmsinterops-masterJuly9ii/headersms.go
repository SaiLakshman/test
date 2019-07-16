
/*

Copyright © 2019 by Pennybase Technologies Sol. Pvt. Ltd. | Blockcube 
All rights reserved. No part of this publication may be reproduced, distributed, or transmitted in any form 
or by any means, including photocopying, recording, or other electronic or mechanical methods, without the 
prior written permission of the Company, except in the case of brief quotations embodied in critical reviews 
and certain other noncommercial uses permitted by copyright law.


This Chaincode is written for Storing, Retrieving, querying the Header that are stored in DLT.
*/



// ====CHAINCODE EXECUTION SAMPLES (CLI) ================== START


// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["rh","{\"hid\":\"QT041111111111111101\",\"peid\":\"A11111111101\",\"htyp\":\"T\",\"cli\":\"BLOCKCUBE\",\"ctgr\":\"8\",\"cts\":\"456789\",\"uts\":\"456787678\"}"]}'

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["rbh","{\"hid\":\"QT041111111111111102\",\"peid\":\"A11111111101\",\"htyp\":\"T\",\"cli\":\"BLOCKCUBE2\",\"ctgr\":\"8\",\"cts\":\"456789\",\"uts\":\"456787678\"}","{\"hid\":\"QT041111111111111103\",\"peid\":\"A11111111102\",\"htyp\":\"T\",\"cli\":\"BLOCKCUBE3\",\"ctgr\":\"8\",\"cts\":\"456789\",\"uts\":\"456787678\"}"]}'

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["uhs","{\"cli\":\"BLOCKCUBE\",\"sts\":\"I\",\"uts\":\"2345678\"}"]}'

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["qh","BLOCKCUBE2","BLOCKCUBE3"]}'

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["hfh","{\"cli\":\"BLOCKCUBE\"}"]}'

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["qhbp","{\"typ\":\"cli\",\"cli\":\"BLOCKCUBE\"}"]}'

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["qhbp","{\"typ\":\"peid\",\"peid\":\"A11111111102\"}"]}'

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["qhwp","{\"selector\":{\"peid\":\"22\"}}","5",""]}'

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["bhe","22"]}'

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["bbh","BLOCKCUBE6","BLOCKCUBE7"]}'


// ====CHAINCODE EXECUTION SAMPLES (CLI) ================== END



package main


import (
	"strings"
 	"strconv"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/chaincode/lib/cid" 
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Logger for logging 
var logger = shim.NewLogger("HEADERSMS-CHAINCODE-INITIALIZED")

// Event Names
const EVTRegisterHeader = "EVT_RegisterHeaderSMS"
const EVTUpdateHeaderStatus = "EVT_UpdateHeaderStatusSMS"
const EVTBlacklistHeader = "EVT_BlacklistHeader"

// Smart contract structure
type HeaderChainCode struct {
}

//Header structure defines the ledger record for any Header
type Header struct {

    ObjType           string `json:"obj"`   // ObjType : is used to distinguish the various types of objects in state database
    Header_ID         string `json:"hid"`	// hid     : Unique id Autogenerated in Backend
    PrincipleEntityId string `json:"peid"`  // peid    : Unique id and A group of headers belongs to this particular entity
    Header_Type		  string `json:"htyp"`  // htyp    : Either T/SE/SI/P : Transactional / Service / Promotional  
    Header_Name		  string `json:"cli"`	// cli     : Unique name to be registeres in DLT
    Status 			  map[string]string`json:"sts"`	   // sts     : Either A/I : Active /  Inactive Operator wise
    Category          string `json:"ctgr"`  // ctgr    : In betwen 0-8 
    CreatedTs		  string `json:"cts"`   // cts     : Header creation time : Autogenerated in Backend
    UpdatedTs         string `json:"uts"`	// uts     : When the header last updated : Autogenerate in Backend
    Creator           string `json:"crtr"`	// crtr    : DLT  Creator node's name 
	UpdatedBy         string `json:"uby"`	// uby     : DLT Node's name
	TMID 			  string `json:"tmid"`  // tmid    : Details of the RTM who added this header on behalf of Entity
	Blacklisted       bool `json:"blklst"`   // blklst  : Header is blacklisted (or not) across TSP
}

// Header Type 
var validHeaderType = map[string]bool{
	"SE": true,
	"SI": true,
	"T": true,
	"P": true,
}

// Valid Category type
var validCategory = map[string]bool{
	"0": true,
	"1": true,
	"2": true,
	"3": true,
	"4": true,
	"5": true,
	"6": true,
	"7": true,
	"8": true,
}

var dltDomainNames = map [string]string {
    "airtel.com" : "AI",                     //Airtel
    "vil.com"    : "VO",                     //"VO" , "ID", "VI"
    "bsnl.com"   : "BL",                     //BSNL
    "mtnl.com"   : "ML",                     //MTNL
    "qtl.infotelconnect.com"  : "QL",        //QTL
    "tata.com"   : "TA",                     //TATA
    "jio.com"    : "JI",                     //JIO
}

func validHeaderEntry(input string, enumMap map[string]bool) bool {
        if _, isEntryExists := enumMap[input]; !isEntryExists {
               return false
        }
        return true
}

func isValidHeader(header Header) (bool, string) {
	
	if len(header.Header_ID) == 0 {
		return false, "Header_ID is mandatory"
	}

	if len(header.PrincipleEntityId) == 0 {
		return false, "PrincipleEntityId is mandatory"
	}

	if len(header.Header_Name) == 0 {
		return false, "Header_Name is mandatory"
	}

	if len(header.CreatedTs) == 0 {
		return false, "Created Timestamp is mandatory"
	}

	if len(header.UpdatedTs) == 0 {
		return false, "Updated Timestamp is mandatory"
	}

	if len(header.Header_Type) == 0 {
		return false, "Header_Type is mandatory"
	} 

	if len(header.Status) == 0 {
		return true, "" 
	} else {
		return false, "Do not pass status as It is being handled at chaincode level."
	}
 
	if header.Header_Type == "P" {
		if _,err:=strconv.Atoi(header.Header_Name); err!=nil{
    		return false, "CLI is not numeric"
		}
	}

    if !validHeaderEntry(header.Header_Type,validHeaderType){
        return false, "Invalid Header Type"
    }

	if !validHeaderEntry(header.Category,validCategory){
    return false, "Invalid Category Provided" 
    } 
   
    if len(header.Category) == 0 {
    	return false, "Category is mandatory"
    }

    if len(header.TMID) == 0 {
    	return true, ""
    } else if _,err:=strconv.Atoi(header.TMID); err!=nil{
    		return false, "TMID is not numeric"
    }
    
    return true,""
}



// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(HeaderChainCode))
	logger.SetLevel(shim.LogDebug)
	if err != nil {
		logger.Error("Error starting Header Chaincode: %s: ", err)
	} else {
		logger.Info("|| STARTING HEADER SMS CHAINCODE ||")
	}
}


// ===================================================================================
// Init initializes chaincode
// ===================================================================================
func (t *HeaderChainCode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	logger.Info("|| HEADER CHAINCODE IS INITIALIZED ||")
	return shim.Success(nil)
}

// ===================================================================================
// Invoke - Our entry point for Invocations
// ===================================================================================
func (t *HeaderChainCode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	function, args := stub.GetFunctionAndParameters()
	logger.Infof("Header Chaincode Invoked, Function name : " +string(function))

	
	// Handle different functions
	switch function {
		case "rh": 								
			return t.registerHeader(stub, args) 			// Register a new header
		case "rbh":	
			return t.registerBulkHeader(stub, args) 		// Register headers in Bulk
		case "uhs":
			return t.updateHeaderStatus(stub,args)			// Change status of a header
		case "qh":
			return t.queryHeader(stub,args)					// Query by Array of CLI : All Matching headers will be returned
		case "qhbp":
			return t.queryHeaderByParam(stub,args)			// Query header depending upon different parameters : CLI / PEID / HID
		case "hfh":
			return t.getHistoryForHeader(stub,args)			// Get history against header when Header name is passed
		case "qhwp":
			return t.queryHeaderWithPagination(stub,args)   // Uses a query string, page size and a bookmark to perform a query
		case "bhe":
			return t.blacklistHeaderByEntity(stub,args)       // Set status Blacklisted to "true" for headers  against Entity 
		case "bbh":
			return t.blacklistBulkHeaders(stub,args)          // Set all headers to blacklisted when cli array is passed
		default:
			logger.Errorf("Received Unknown Function invocation : Available Function : rh , rbh , uhs, qh, hfh, qhwp, bhe, bbh")
			return shim.Error("Received Unknown Function invocation : Available function : rh , rbh , uhs, qh, hfh, qhwp, bhe, bbh")
		}
}


// ========================================================================================
// registerHeader - register a header in chaincode state
// ========================================================================================
func (t *HeaderChainCode) registerHeader(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	var dltNode string
	certData, err := cid.GetX509Certificate(stub)
	if err != nil {
		logger.Errorf("setHeader : Getting certificate Details Error : " + string(err.Error()))
		return shim.Error("setHeader : Getting certificate Details Error : " + string(err.Error()))
	}

	Organizations := certData.Issuer.Organization
	if isExists, ok := dltDomainNames[Organizations[0]];!ok{
	return shim.Error("Unauthorized Node Access")
    } else { dltNode = isExists }

	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}

	var data Header
	err1 := json.Unmarshal([]byte(args[0]), &data)
	if err1 != nil {
		logger.Errorf("setHeader : Input arguments unmarhsaling Error : " + string(err.Error()))
		return shim.Error("setHeader : Input arguments unmarhsaling Error : " + string(err.Error()))
	}

	if isValid,errMsg:=isValidHeader(data);!isValid{
			logger.Errorf("setHeader:"+string(errMsg))
			return shim.Error(errMsg)
	}

	headerSearch := `{
		"obj":"HeaderSMS",
		"hid":"%s"
	}`
	hID := data.Header_ID
	headerData := t.retriveHeaderRecords(stub, fmt.Sprintf(headerSearch, hID), "headerSearchByID")
	if len(headerData) > 0  {
		logger.Infof("Header_ID already exist for : " + data.Header_Name + ", Please provide unique hid ")
        return shim.Error("Header_ID already exist for : " + data.Header_Name + ", Please provide unique hid ")
	}

	if recordBytes, _ := stub.GetState(data.Header_Name); len(recordBytes) > 0 {
		return shim.Error("Header already registered. Provide an unique header name")
	}
	
		data.ObjType = "HeaderSMS"
		var m = make(map[string]string)
		m[dltNode] = "A"
		data.Status = m
		data.Creator = Organizations[0]
		data.UpdatedBy = Organizations[0]
		data.Blacklisted = false
		logger.Infof("Header_ID is " + data.Header_ID)
		headerAsBytes, err := json.Marshal(data)
		if err != nil {
			logger.Errorf("setHeader : Marshalling Error : " + string(err.Error()))
			return shim.Error("setHeader : Marshalling Error : " + string(err.Error()))
		}
		//Inserting DataBlock to BlockChain
		err = stub.PutState(data.Header_Name, headerAsBytes)
		if err != nil {
			logger.Errorf("setHeader : PutState Failed Error : " + string(err.Error()))
			return shim.Error("setHeader : PutState Failed Error : " + string(err.Error()))
		}
		logger.Infof("setHeader : PutState Success : " + string(headerAsBytes))

		err2 := stub.SetEvent(EVTRegisterHeader, headerAsBytes)
		if err2 != nil {
			logger.Errorf("Event not generated for event : EVTRegisterHeader")
			return shim.Error("Event not generated for event : EVTRegisterHeader")
		}

		resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"headerRegistered": data.Header_Name,
		"message":  "single header registered successfully",
		"Header":   data,
		"status": "true",
		}
		respJSON, _ := json.Marshal(resultData)
	    return shim.Success(respJSON)
	
}



// ========================================================================================
// registerBulkHeader - Register Bulk header in chaincode state
// ========================================================================================
func (t *HeaderChainCode) registerBulkHeader(stub shim.ChaincodeStubInterface, args []string) sc.Response  {
	
	var recordcount int
	var dltNode string
	headerRejected := make([]map[string]interface{}, 0)
	headerRegistered := make([]string, 0)

	certData, err := cid.GetX509Certificate(stub)
	if err != nil {
		logger.Errorf("registerBulkHeader : Getting certificate Details Error : " + string(err.Error()))
		return shim.Error("registerBulkHeader : Getting certificate Details Error : " + string(err.Error()))
	}

	Organizations := certData.Issuer.Organization
	if isExists, ok := dltDomainNames[Organizations[0]];!ok{
	return shim.Error("Unauthorized Node Access")
    } else { dltNode = isExists }
    
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}

	recordcount = 0
	for i := 0; i < len(args); i++ {
		var data Header
		logger.Infof(args[i])
		err := json.Unmarshal([]byte(args[i]), &data)
		if err != nil {
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": data.Header_Name , "Value": "Input arguments unmarhsaling Error" })	
			continue
		}

		if isValid,errMsg:=isValidHeader(data);!isValid{
			logger.Errorf("registerBulkHeader:"+string(errMsg))
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": data.Header_Name, "Value": string(errMsg) })	
			continue
		}

		headerSearch := `{
			"obj":"HeaderSMS",
			"hid":"%s"
		}`
		hID := data.Header_ID
		headerData := t.retriveHeaderRecords(stub, fmt.Sprintf(headerSearch, hID), "headerSearchByID")
		if len(headerData) > 0  {
			logger.Errorf("Header_ID already exist for : " + data.Header_Name + ", Please provide unique hid ")
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": data.Header_Name , "Value": "Header ID already exists" })	
			continue
		}
			
		if recordBytes, _ := stub.GetState(data.Header_Name); len(recordBytes) > 0 {
			logger.Errorf("Header already registered. Provide an unique header name")
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": data.Header_Name , "Value": "Header already registered"})	
			continue
		}

		recordcount = recordcount + 1
		data.ObjType = "HeaderSMS"
		var m = make(map[string]string)
		m[dltNode] = "A"
		data.Status = m

		// data.Status[dltNode]="A"
		data.Creator= Organizations[0]
		data.UpdatedBy = Organizations[0]
		data.Blacklisted = false
		logger.Infof("Header_ID is " + data.Header_ID)
		headerAsBytes, err := json.Marshal(data)
		if err != nil {
			logger.Errorf("registerBulkHeader : Marshalling Error : " + string(err.Error()))
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": data.Header_Name , "Value": "Marshalling Error " })	
			continue
		}

		//Inserting DataBlock to BlockChain
		err = stub.PutState(data.Header_Name, headerAsBytes)
		if err != nil {
			logger.Errorf("registerBulkHeader : PutState Failed Error : " + string(err.Error()))
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": data.Header_Name , "Value": "PutState Failed Error" })	
			continue
		}

		err2 := stub.SetEvent(EVTRegisterHeader, headerAsBytes)
		if err2 != nil {
			logger.Errorf("Event not generated for event : EVTRegisterHeader")
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": data.Header_Name , "Value": "Event not generated for event : EVTRegisterHeader" })	
			continue
		}

		logger.Infof("registerBulkHeader : PutState Success : " + string(headerAsBytes))
		headerRegistered = append(headerRegistered, data.Header_Name)	 	
	}	

	resultData := map[string]interface{} {
	"trxnID":   stub.GetTxID(),
	"headerRegistered": headerRegistered,
	"headerRejected": headerRejected,
	"message" : "Bulk Header registered successfully",
	"countSuccess":  strconv.Itoa(recordcount),
	}

	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)	
}
	

// ========================================================================================
// updateHeaderStatus - Update header status to A / I
// It only updates Status to either Active or Inactive. TO delete header status 
// can be done using "dbh" function.
// ========================================================================================
func (t *HeaderChainCode) updateHeaderStatus(stub shim.ChaincodeStubInterface, args []string) sc.Response { 

	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}
	var dltNode string
	var data map[string]interface{}
	err := json.Unmarshal([]byte(args[0]), &data)
	if err != nil {
		logger.Errorf("updateHeaderStatus : Input arguments unmarhsaling Error : " + string(err.Error()))
		return shim.Error("updateHeaderStatus : Input arguments unmarhsaling Error : " + string(err.Error()))
	}

	certData, err := cid.GetX509Certificate(stub)
	if err != nil {
		logger.Errorf("registerBulkHeader : Getting certificate Details Error : " + string(err.Error()))
		return shim.Error("registerBulkHeader : Getting certificate Details Error : " + string(err.Error()))
	}

	Organizations := certData.Issuer.Organization
	if isExists, ok := dltDomainNames[Organizations[0]];!ok{
	return shim.Error("Unauthorized Node Access")
    } else { dltNode = isExists }

	if len(data) == 3 {

		RecordAsBytes, err1 := stub.GetState(data["cli"].(string))
		if err1 != nil {
			logger.Infof(" Failed to get Header Record : " + data["cli"].(string) + " Error : " + string(err.Error()))
			return shim.Error(" Failed to get Header Record " + data["cli"].(string) + " Error : " + string(err.Error()))
		} else if RecordAsBytes == nil {
			logger.Infof(" Failed to get Header Record : " + data["cli"].(string) + " Error : Record Does not exist ")
			return shim.Error(" Failed to get Header Record " + data["cli"].(string) + " Error : Record Does not exist ")
		}

		header := Header{}
		err := json.Unmarshal(RecordAsBytes, &header)
		if err != nil {
			logger.Errorf("updateHeaderStatus : Existing header data Unmarhsaling Error : " + string(err.Error()))
			return shim.Error("updateHeaderStatus : Existing header data Unmarhsaling Error : " + string(err.Error()))
		}

		var existingStatus = make(map[string]string)
		existingStatus = header.Status				

			switch data["sts"].(string) {
		case "A": 
			if existingStatus[dltNode] == "I" { existingStatus[dltNode] = "A" } else {
				logger.Errorf("Header is already Active")
			 	return shim.Error("Header is already Active")
			}
		case "I":
			if existingStatus[dltNode] == "A" { existingStatus[dltNode] = "I" } else {
			 	logger.Errorf("Header is already Inactive")
			 	return shim.Error("Header is already Inactive")
			 }
		default:
			logger.Errorf("Received Unknown Status type || Must provide either A or I ")
			return shim.Error("Received Unknown Status type || Must provide either A or I ")
		}

		header.Status = existingStatus
		header.UpdatedTs = data["uts"].(string)
		logger.Infof("Header_Name is " + header.Header_Name)
		headerAsBytes, err := json.Marshal(header)
		if err != nil {
			logger.Errorf("updateHeaderStatus : Marshalling Error : " + string(err.Error()))
			return shim.Error("updateHeaderStatus : Marshalling Error : " + string(err.Error()))
		}
		//Inserting DataBlock to BlockChain
		err = stub.PutState(header.Header_Name, headerAsBytes)
		if err != nil {
			logger.Errorf("updateHeaderStatus : PutState Failed Error : " + string(err.Error()))
			return shim.Error("updateHeaderStatus : PutState Failed Error : " + string(err.Error()))
		}
		logger.Infof("updateHeaderStatus : PutState Success : " + string(headerAsBytes))
		err2 := stub.SetEvent(EVTUpdateHeaderStatus, headerAsBytes)

		if err2 != nil {
			logger.Errorf("Event not generated for event : EVTUpdateHeaderStatus")
			return shim.Error("Event not generated for event : EVTUpdateHeaderStatus")
		}
	
	} else {
		logger.Errorf("updateHeaderStatus : Incorrect Number Of Arguments, i.e. CLI, Status, UpdatedTs expected")
	    return shim.Error("updateHeaderStatus : Incorrect Number Of Arguments i.e. CLI, Status, UpdatedTs expected")				
	}

	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"headerUpdated": data["cli"].(string),
		"message":  "Status is updated Successfully.",
		"Header":   data,
	}
	respJSON, _ := json.Marshal(resultData)
    return shim.Success(respJSON)
}


// ========================================================================================
// queryHeader - Query by Array of CLI : All Matching headers will be returned
// ========================================================================================
func (t *HeaderChainCode) queryHeader(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	var recordcount = 0
	headerNotExist := make([]map[string]interface{}, 0)
	headerExist := make([]string, 0)
	resp := make([]map[string]interface{}, 0)

	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}

	for i:=0; i<len(args); i++ {
		valAsBytes, err := stub.GetState(args[i]) //get the record from chaincode state
		if err != nil {
			logger.Infof("Failed to get state for Header_Name " + args[i] )
			headerNotExist = append(headerNotExist, map[string]interface{}{"Header_Name": args[i] , "Value": "Failed to get state for Header" })	
			continue
		} else if valAsBytes == nil {
			logger.Infof("Record does not exist for Header_Name " + args[i] )
			headerNotExist = append(headerNotExist, map[string]interface{}{"Header_Name": args[i] , "Value": "Record does not exist for Header" })	
			continue
		}

		recordcount = recordcount +1
		headerExist = append(headerExist, args[i])
		value := make(map[string]interface{})
		json.Unmarshal(valAsBytes, &value)
		// recordsJSON, _ := json.Marshal(valAsBytes)
		resp = append(resp, map[string]interface{}{"Header_Name": args[i], "Value": value })	
		logger.Info("Successfully submitted the result for " +args[i])
	}

		logger.Info("")
		resultData := map[string]interface{}{
			"headerNotExist": headerNotExist,
			"headerExist":   headerExist, 
			"dataOfHeader" : resp,
			"countSuccess":  strconv.Itoa(recordcount),
		}

		respJSON, _ := json.Marshal(resultData)
		return shim.Success(respJSON)
}



// ========================================================================================
// queryHeaderByParam - Query header depending upon different parameters : CLI / PEID / HID
// ========================================================================================
func (t *HeaderChainCode) queryHeaderByParam(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var response sc.Response
	searchCriteria := make(map[string]string)

	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}

	err := json.Unmarshal([]byte(args[0]), &searchCriteria)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}
	searchType, isOk := searchCriteria["typ"]
	if !isOk {
		return shim.Error("Search type not provided")
	}

	logger.Infof(args[0])
	logger.Infof("length is : ", len(searchCriteria[searchType]))

	switch searchType {
	case "cli":
		headerSearchCriteria := `{
			"obj":"HeaderSMS"	,
			"cli":"%s"
		}`
		headerName := searchCriteria[searchType]
		header := t.retriveHeaderRecords(stub, fmt.Sprintf(headerSearchCriteria, headerName), "headerSearchByName")
		recordsJSON, _ := json.Marshal(header)
		response = shim.Success(recordsJSON)
	case "hid":
		headerSearchCriteria := `{
			"obj":"HeaderSMS"	,
			"hid":"%s"
		}`
		headerID := searchCriteria[searchType]
		header := t.retriveHeaderRecords(stub, fmt.Sprintf(headerSearchCriteria, headerID), "headerSearchByID")
		recordsJSON, _ := json.Marshal(header)
		response = shim.Success(recordsJSON)
	case "peid":
		headerSearchCriteria := `{
			"obj":"HeaderSMS"	,
			"peid":"%s"
		}`
		headerID := searchCriteria[searchType]
		header := t.retriveHeaderRecords(stub, fmt.Sprintf(headerSearchCriteria, headerID), "headerSearchByPeid")
		recordsJSON, _ := json.Marshal(header)
		response = shim.Success(recordsJSON)
	default:
		response = shim.Error("Unsupported search type provided " + searchType)
	}
	return response
}


// ========================================================================================
// getHistoryForHeader - Get history of the key when header name is passed
// ========================================================================================
func (t *HeaderChainCode) getHistoryForHeader(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}

	var data map[string]interface{}
	err := json.Unmarshal([]byte(args[0]), &data)
	if err != nil {
		logger.Errorf("getHistoryForHeader : Input arguments unmarhsaling Error : " + string(err.Error()))
		return shim.Error("getHistoryForHeader : Input arguments unmarhsaling Error : " + string(err.Error()))
	}

	RecordAsBytes, err := stub.GetState(data["cli"].(string))
	if err != nil {
		logger.Infof("Failed to get Header Record : " + data["cli"].(string) + " Error : " + string(err.Error()))
		return shim.Error("Failed to get Header Record : " + data["cli"].(string) + " Error : " + string(err.Error()))
	} else if RecordAsBytes == nil {
		fmt.Println("This record does not exists : " + data["cli"].(string))
		return shim.Error("This record does not exists : " + data["cli"].(string))
	}

	historyIer, err := stub.GetHistoryForKey(data["cli"].(string))

	if err != nil {
	    fmt.Println(err.Error())
	    return shim.Error(err.Error())
	}

	historicResponse := make([]map[string]interface{}, 0)
	for historyIer.HasNext() {
	    modification, err := historyIer.Next()
	    if err != nil {
	        fmt.Println(err.Error())
	        return shim.Error(err.Error())
	}

		value := make(map[string]interface{})
		json.Unmarshal(modification.Value, &value)
		historicResponse = append(historicResponse, map[string]interface{}{"txId": modification.TxId, "value": value})
	} 

	respJSON, _ := json.Marshal(historicResponse)
	return shim.Success(respJSON)
}


// ===========================================================================================
// queryHeaderWithPagination - uses a query string, page size and a bookmark to perform a query
// for marbles. Query string matching state database syntax is passed in and executed as is.
// ===========================================================================================
func (t *HeaderChainCode) queryHeaderWithPagination(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 3 {
		return shim.Error("queryHeadersWithPagination : Incorrect number of arguments. Expecting 3 Args")
	}

	var records []Header
	queryString := args[0]
	pageSize, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		logger.Errorf("queryHeadersWithPagination : Unable to parse the input data: " + string(err.Error()))
		return shim.Error("queryHeadersWithPagination : Unable to parse the input data: " + string(err.Error()))
	}
	bookmark := args[2]
	resultsIterator,responseMetaData,err:=stub.GetQueryResultWithPagination(queryString,int32(pageSize),bookmark)
    if err!=nil{
        logger.Errorf("queryHeadersWithPagination:GetQueryResultWithPagination is Failed :"+string(err.Error()))
        return shim.Error("queryHeadersWithPagination:GetQueryResultWithPagination is Failed ")
    }
        

    for resultsIterator.HasNext() {
        record:=Header{}
        recordBytes,_:=resultsIterator.Next()
        if string(recordBytes.Value)==""{
                continue
        }
        err:=json.Unmarshal(recordBytes.Value,&record)
        if err!=nil{
                logger.Errorf("queryHeadersWithPagination:Unable to unmarshal Header retrieved :"+string(err.Error()))
                return shim.Error("queryHeadersWithPagination:Unable to unmarshal Header retrieved ")
        }
        records=append(records,record)
    }

    resultData:=map[string]interface{}{
            "status":"true",
            "HeaderReceived":records,
            "RecordsCount":responseMetaData.FetchedRecordsCount,
            "bookmark":responseMetaData.Bookmark,
    }
    respJson,_:=json.Marshal(resultData)
    return shim.Success(respJson)
}



// ===========================================================================================
// blacklistHeaderByEntity -  Blacklist all headers againsit entity ID.
// ===========================================================================================
func (t *HeaderChainCode) blacklistHeaderByEntity(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}

	var recordcount int
	headerRejected := make([]map[string]interface{}, 0)
	headerDeleted := make([]string, 0)

	recordcount = 0
	headerSearch := `{
		"obj":"HeaderSMS",
		"peid":"%s"
	}`
	peid := args[0]
	headerData := t.retriveHeaderRecords(stub, fmt.Sprintf(headerSearch, peid), "headerSearchByPeid")

	if(len(headerData) == 0) {
		logger.Errorf("No header exists for this Entity")
		shim.Error("No header exists for this Entity")
	}

	for i:=0; i<len(headerData); i++ {

		hName := headerData[i].Header_Name

		if headerData[i].Blacklisted == false {
			headerData[i].Blacklisted = true
		} else {
			logger.Errorf("blacklistHeaderByEntity : Already Blacklisted  ")
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": hName , "Value": "Already Blacklisted " })	
			continue
		}

		headerAsBytes, err := json.Marshal(headerData[i])
		if err != nil {
			logger.Errorf("blacklistHeaderByEntity : Marshalling Error : " + string(err.Error()))
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": hName , "Value": "Marshalling Error " })	
			continue
		}

		//Inserting DataBlock to BlockChain
		err = stub.PutState(hName, headerAsBytes)
		if err != nil {
			logger.Errorf("blacklistHeaderByEntity : PutState Failed Error : " + string(err.Error()))
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": hName , "Value": "PutState Failed Error" })	
			continue
		}

		err2 := stub.SetEvent(EVTBlacklistHeader, headerAsBytes)
		if err2 != nil {
			logger.Errorf("Event not generated for event : EVTBlacklistHeader")
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": hName , "Value": "Event not generated for event : EVTBlacklistHeader" })	
			continue
		}

		recordcount = recordcount + 1
		logger.Infof("blacklistHeaderByEntity : PutState Success : " + string(headerAsBytes))
		headerDeleted = append(headerDeleted, hName)	 	
	}	

	resultData := map[string]interface{} {
	"trxnID":   stub.GetTxID(),
	"headerBlacklisted": headerDeleted,
	"headerRejected": headerRejected,
	"message" : "Blacklisted all headers against PEID : " +peid ,
	"countSuccess":  strconv.Itoa(recordcount),
	}

	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)	
}



// ===========================================================================================
// blacklistBulkHeaders - Blacklist headers in bulk
// ===========================================================================================
func (t *HeaderChainCode)  blacklistBulkHeaders(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	var recordcount = 0
	headerRejected := make([]map[string]interface{}, 0)
	headerDeleted := make([]string, 0)

	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}

	certData, err := cid.GetX509Certificate(stub)
	if err != nil {
		logger.Errorf("setHeader : Getting certificate Details Error : " + string(err.Error()))
		return shim.Error("setHeader : Getting certificate Details Error : " + string(err.Error()))
	}

	Organizations := certData.Issuer.Organization

	for i:=0; i<len(args); i++ {
		valAsBytes, err := stub.GetState(args[i]) //get the record from chaincode state
		if err != nil {
			logger.Infof("Failed to get state for Header_Name " + args[i] )
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": args[i] , "Value": "Failed to get state for Header" })	
			continue
		} else if valAsBytes == nil {
			logger.Infof("Record does not exist for Header_Name " + args[i] )
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": args[i] , "Value": "Record does not exist for Header" })	
			continue
		}

		var data Header
		err1 := json.Unmarshal([]byte(valAsBytes), &data)
		if err1 != nil {
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": args[i] , "Value": "Input arguments unmarhsaling Error" })	
			continue
		}

		var creatr string
		var existingUpdatedBy string
		creatr = Organizations[0]
		existingUpdatedBy = data.UpdatedBy

		if strings.Compare(existingUpdatedBy,creatr)==0{

		if data.Blacklisted == false {
			data.Blacklisted = true
		} else {
			logger.Errorf("blacklistBulkHeaders : Already Blacklisted : " )
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": data.Header_Name , "Value": "Already Blacklisted " })	
			continue
		}

		headerAsBytes, err := json.Marshal(data)
		if err != nil {
			logger.Errorf("blacklistBulkHeaders : Marshalling Error : " + string(err.Error()))
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": data.Header_Name , "Value": "Marshalling Error " })	
			continue
		}

		//Inserting DataBlock to BlockChain
		err = stub.PutState(data.Header_Name, headerAsBytes)
		if err != nil {
			logger.Errorf("blacklistBulkHeaders : PutState Failed Error : " + string(err.Error()))
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": data.Header_Name , "Value": "PutState Failed Error" })	
			continue
		}

		err2 := stub.SetEvent(EVTBlacklistHeader, headerAsBytes)
		if err2 != nil {
			logger.Errorf("Event not generated for event : EVTBlacklistHeader")
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": data.Header_Name , "Value": "Event not generated for event : EVTBlacklistHeader" })	
			continue
		}

		recordcount = recordcount + 1
		logger.Infof("blacklistBulkHeaders : PutState Success : " + string(headerAsBytes))
		headerDeleted = append(headerDeleted, args[i])

		} else {
	    	logger.Errorf("Unauthorized access to blacklist headers created by other node")
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name": data.Header_Name , "Value": "Unauthorized access to blacklist headers created by other node" })	
			continue
	    }

	}

		logger.Info("")
		resultData := map[string]interface{}{
			"trxnID":   stub.GetTxID(),
			"headerRejected": headerRejected,
			"headerBlacklisted":   headerDeleted, 
			"message" : "All headers have been set to blacklisted",
			"countSuccess":  strconv.Itoa(recordcount),
		}

		respJSON, _ := json.Marshal(resultData)
		return shim.Success(respJSON)
}


func (t *HeaderChainCode) retriveHeaderRecords(stub shim.ChaincodeStubInterface, criteria string, indexs ...string) []Header {
    
	var finalSelector string
	records := make([]Header, 0)

	if len(indexs) == 0 {
		finalSelector = fmt.Sprintf("{\"selector\":%s }", criteria)

	} else {
		finalSelector = fmt.Sprintf("{\"selector\":%s , \"use_index\" :\"%s\" }", criteria, indexs[0])
	}

	logger.Infof("Query Selector : %s", finalSelector)
	resultsIterator, _ := stub.GetQueryResult(finalSelector)
	for resultsIterator.HasNext() {
		record := Header{}
		recordBytes, _ := resultsIterator.Next()
		err := json.Unmarshal(recordBytes.Value, &record)
		if err != nil {
			logger.Infof("Unable to unmarshal Header retrived:: %v", err)
		}
		records = append(records, record)
	}
	return records
}

