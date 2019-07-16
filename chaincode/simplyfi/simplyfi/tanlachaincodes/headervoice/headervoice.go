/**
*Header Voice chaincode written by @lucky
yet to do: deleteHeadersByEntityId, deleteBulkHeaders(by headername)
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	id "github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
	"github.com/hyperledger/fabric/protos/peer"
)

var _headerVoiceLogger = shim.NewLogger("HeaderVoice")

const _CreateEvent = "CREATE_HEADER"
const _UpdateEvent = "UPDATE_HEADER"
const _BulkCreateEvent = "BULK_CREATE"

//HeaderVoice structure defines the ledger record for any HeaderVoice's
type HeaderVoice struct {
	ObjType    string `json:"obj"`  //DocType
	HeaderId   string `json:"hid"`  // HeaderId unique
	PEID       string `json:"peid"` //creator of the scrub
	HeaderType string `json:"htyp"` // Type of the header (P,T,SE,SI)
	CLI        string `json:"cli"`  //HeaderName
	Cname	   string `json:"cname"` // cname   : May have value for voice
	Creator    string `json:"crtr"` // creator of the txn in the ledger
	CreateTs   string `json:"cts"`  // created timestamp of the txn
	UpdatedBy  string `json:"uby"`  //updater of the txn in the ledger
	UpdateTs   string `json:"uts"`  //updated timestamp of the txn
	Status     string `json:"sts"`  // Status of the header (A,I,B,D)
	Category   string `json:"ctgr"` // category of header
	TMID	   string `json:"tmid"` //telemarketer id
}

//HeaderVoiceManager manages HeaderVoice related transactions
type HeaderVoiceManager struct {
}

var errorDetails, errKey, jsonResp, repError string

var headerStatus = map[string]bool{
	"A": true,
	"I": true,
	"B": true,
	"D": true,
}

var headerType = map[string]bool{
	"P":  true,
	"T":  true,
	"SE": true,
	"SI": true,
}

var category = map[string]bool{
	"1": true,
	"2": true,
	"3": true,
	"4": true,
	"5": true,
	"6": true,
	"7": true,
	"8": true,
}

func validEnumEntry(input string, enumMap map[string]bool) bool {
	if _, isEntryExists := enumMap[input]; !isEntryExists {
		return false
	}
	return true
}

//IsValidHeaderIdPresent checks for  validity of Header
func IsValidHeaderIdPresent(s HeaderVoice) (bool, string) {
	if len(s.HeaderId) == 0 {
		return false, "HeaderId is mandatory"
	}
	if len(s.PEID) == 0 {
		return false, "Principal EntityID is mandatory"
	}
	if len(s.CLI) == 0 {
		return false, "CLI is mandatory"
	}
	if len(s.Creator) == 0 {
		return false, "Creator is mandatory"
	}
	if len(s.CreateTs) == 0 {
		return false, "CreateTs is mandatory"
	}
	if len(s.UpdatedBy) == 0 {
		return false, "UpdatedBy is mandatory"
	}
	if len(s.UpdateTs) == 0 {
		return false, "UpdateTs is mandatory"
	}
	if !validEnumEntry(s.HeaderType, headerType) {
		return false, "Header Type:Enter either P, T, SE, SI"
	}
	if !validEnumEntry(s.Status, headerStatus) {
		return false, "Status: Enter either A, I, B, D"
	}
	if !validEnumEntry(s.Category, category) {
		return false, "Category: Enter either 1, 2, 3, 4, 5, 6, 7, 8"
	}
	return true, ""
}

//registering the header voice data in the ledger
func (s *HeaderVoiceManager) registerHeader(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	var headerToSave HeaderVoice
	//checking for the length of the input
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	// unmarshalling the input to headerToSave object
	err := json.Unmarshal([]byte(args[0]), &headerToSave)
	if err != nil {
		repError= strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON provided- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	// checking whether header already exists or not with cli
	if recordBytes, _ := stub.GetState(headerToSave.CLI); len(recordBytes) > 0 {
		errKey = headerToSave.CLI
		errorDetails = "Header with this CLI already exist, provide unique CLI"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//checking whether header already exists or not with header id using selector query
	headerSearch := `{"obj":"HeaderVoice","hid":"%s"}`
	hID := headerToSave.HeaderId
	headerData := s.retrieveHeaderRecords(stub, fmt.Sprintf(headerSearch, hID), "headerSearchByID")
	if len(headerData) > 0 {
		errKey = headerToSave.HeaderId
		errorDetails = "Header with HeaderId already Exists, Provide unique HeaderId"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the header with the details provided as input
	headerToSave.ObjType = "HeaderVoice"
	_, creator := s.getInvokerIdentity(stub)
	headerToSave.Creator = creator
	headerToSave.UpdatedBy = creator
	headerToSave.UpdateTs = headerToSave.CreateTs
	//marshalling the data for storing into the ledger
	headerJSON, marshalErr := json.Marshal(headerToSave)
	if marshalErr != nil {
		repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//checking for the validity of the header values before storing into the ledger
	if isValid, errMsg := IsValidHeaderIdPresent(headerToSave); !isValid {
		errorDetails = errMsg
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//storing the header voice info  into the ledger
	_headerVoiceLogger.Info("Saving Header to the ledger with id----------", headerToSave.HeaderId)
	err = stub.PutState(headerToSave.CLI, headerJSON)
	if err != nil {
		errKey = string(headerJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save header with CLI- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//setting event after storing into the ledger
	retErr := stub.SetEvent(_CreateEvent, headerJSON)
	if retErr != nil {
		errKey = string(headerJSON)
		repError = strings.Replace(retErr.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : CREATE_HEADER- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the response and return to the application layer
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"headerID": headerToSave.HeaderId,
		"message":  "Header Registered Successful",
		"header":   headerToSave,
		"status": "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//registerBulkHeader will create headers from the input at once
func (s *HeaderVoiceManager) registerBulkHeader(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	var recordcount int
	rejectedHeaders := make([]map[string]interface{}, 0)
	registeredHeaders := make([]string, 0)
	recordcount = 0
	var listHeader []HeaderVoice
	//checking the length of the input
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("registerBulkHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the input to headerToSave object
	err := json.Unmarshal([]byte(args[0]), &listHeader)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON provided- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("registerBulkHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	_, creator := s.getInvokerIdentity(stub)
	//iterating through  the list of headers to save
	for i := 0; i < len(listHeader); i++ {
		var headerToSave HeaderVoice
		headerToSave = listHeader[i]
		//checking whether header already exists or not with cli(header name)
		if recordBytes, _ := stub.GetState(headerToSave.CLI); len(recordBytes) > 0 {
			errKey = headerToSave.CLI
			errorDetails = "Header With CLI already Exists, Provide unique CLI(HeaderName) "
			jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerVoiceLogger.Errorf("registerBulkHeader: " + jsonResp)
			rejectedHeaders = append(rejectedHeaders, map[string]interface{}{"Header_Name: ": headerToSave.CLI, "Value": "Header Name Already Exists"})
			continue
		}
		//checking whether header already exists or not with header id using selector query
		headerSearch := `{"obj":"HeaderVoice","hid":"%s"}`
		hID := headerToSave.HeaderId
		headerData := s.retrieveHeaderRecords(stub, fmt.Sprintf(headerSearch, hID), "headerSearchByID")
		if len(headerData) > 0 {
			errKey = headerToSave.HeaderId
			errorDetails = "Header with HeaderId already Exists, Provide unique HeaderId"
			jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerVoiceLogger.Errorf("registerBulkHeader: " + jsonResp)
			rejectedHeaders = append(rejectedHeaders, map[string]interface{}{"Header_Name: ": headerToSave.CLI, "Value": "Header ID already exists"})
			continue
		}
		//packaging the header with the details provided as input
		headerToSave.ObjType = "HeaderVoice"
		headerToSave.Creator = creator
		headerToSave.UpdatedBy = creator
		headerToSave.UpdateTs = headerToSave.CreateTs
		//marshalling the data for storing into the ledger
		headerJSON, marshalErr := json.Marshal(headerToSave)
		if marshalErr != nil {
			repError = strings.Replace(err.Error(), "\"", " ", -1)
			errorDetails = "Cannot Marshal the JSON- " + repError
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerVoiceLogger.Errorf("registerBulkHeader: " + jsonResp)
			rejectedHeaders = append(rejectedHeaders, map[string]interface{}{"Header_Name: ": headerToSave.CLI, "Value": "Marshalling Error "})
			continue
		}
		//checking for the validity of the header values before storing into the ledger
		if isValid, err := IsValidHeaderIdPresent(headerToSave); !isValid {
			errorDetails = err
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerVoiceLogger.Errorf("registerBulkHeader: " + jsonResp)
			rejectedHeaders = append(rejectedHeaders, map[string]interface{}{"Header_Name: ": headerToSave.CLI, "Value": err})
			continue
		}
		_headerVoiceLogger.Info("Saving Header to the ledger with id----------", headerToSave.HeaderId)
		//storing the header into the ledger
		err = stub.PutState(headerToSave.CLI, headerJSON)
		if err != nil {
			errKey = string(headerJSON)
			repError = strings.Replace(err.Error(), "\"", " ", -1)
			errorDetails = "Unable to save header with CLI- " + repError
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerVoiceLogger.Errorf("registerBulkHeader: " + jsonResp)
			rejectedHeaders = append(rejectedHeaders, map[string]interface{}{"Header_Name: ": headerToSave.CLI, "Value": "PutState Failed"})
			continue
		} else if err == nil {
			recordcount = recordcount + 1
		}
		//setting an event after storing the data into the ledger
		retErr := stub.SetEvent(_BulkCreateEvent, headerJSON)
		if retErr != nil {
			errKey = string(headerJSON)
			repError = strings.Replace(retErr.Error(), "\"", " ", -1)
			errorDetails = "Event not generated for event : BULK_CREATE- " + repError
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerVoiceLogger.Errorf("registerBulkHeader: " + jsonResp)
			rejectedHeaders = append(rejectedHeaders, map[string]interface{}{"Header_Name: ": headerToSave.CLI, "Value": "Event not generated for EVTRegisterHeader"})
			continue
		}
		registeredHeaders = append(registeredHeaders, headerToSave.CLI)
	}
	//packaging the response and return to the application layer
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"header_f": rejectedHeaders,
		"message":  "Header Registered Successfully",
		"status":   "true",
		"successCount": strconv.Itoa(recordcount),
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//updateStatus will update the status of the header given headerId in the ledger
func (s *HeaderVoiceManager) updateStatus(stub shim.ChaincodeStubInterface) peer.Response {
	var jsonResp string
	var updatedHeader, existingHeader HeaderVoice
	_, args := stub.GetFunctionAndParameters()
	//checking the lenght of the input
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the input data to the updatedHeader object
	errHeader := json.Unmarshal([]byte(args[0]), &updatedHeader)
	if errHeader != nil {
		repError = strings.Replace(errHeader.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON provided- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//getting the data from the ledger with headername as input
	existingHeaderResult, err := stub.GetState(updatedHeader.CLI)
	if err != nil {
		errKey = updatedHeader.CLI
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Could not fetch the details for the Header- " + repError
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	} else if existingHeaderResult == nil {
		errKey = updatedHeader.CLI
		errorDetails = "Header does not exist with CLI"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the data retrieved from the ledger to existingHeader object for update
	err = json.Unmarshal([]byte(existingHeaderResult), &existingHeader)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON for storing- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the updated values to store in the ledger
	_, creatorUpdatedBy := s.getInvokerIdentity(stub)
	existingHeader.UpdateTs = updatedHeader.UpdateTs
	existingHeader.UpdatedBy = creatorUpdatedBy
	existingHeader.Status = updatedHeader.Status
	//marshalling the data to store in the ledger
	headerJSON, marshalErr := json.Marshal(existingHeader)
	if marshalErr != nil {
		repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//checking for the validity of the values before storing into the ledger
	if isValid, errMsg := IsValidHeaderIdPresent(existingHeader); !isValid {
		errorDetails = errMsg
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//storing the updated values into the ledger against the existing cli
	err = stub.PutState(existingHeader.CLI, headerJSON)
	if err != nil {
		errKey = string(headerJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save header with CLI- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//setting an event after recording into the ledger
	retErr := stub.SetEvent(_UpdateEvent, headerJSON)
	if retErr != nil {
		errKey = string(headerJSON)
		repError = strings.Replace(retErr.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : UPDATE_HEADER- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the response and return to the app layer
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"headerID": existingHeader.HeaderId,
		"message":  "Update Status Successful",
		"header":   existingHeader,
		"status":"true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//queryByHeader will query the header by cli from the ledger
func (s *HeaderVoiceManager) queryByHeader(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	var qHeader, existingHeader HeaderVoice
	//checking the length of the input
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the input to the qHeader object
	errHeader := json.Unmarshal([]byte(args[0]), &qHeader)
	if errHeader != nil {
		repError= strings.Replace(errHeader.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("queryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//getting the values from the ledger against CLI(header name)
	existingHeaderResult, err := stub.GetState(qHeader.CLI)
	if err != nil {
		errKey = qHeader.CLI
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to fetch the Header- " + repError
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("queryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	} else if existingHeaderResult == nil {
		errKey = qHeader.CLI
		errorDetails = "Header does not exist with CLI"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("queryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the existingresult from the ledger to existingHeader object
	err = json.Unmarshal([]byte(existingHeaderResult), &existingHeader)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("queryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the response and return to the app layer
	resultData := map[string]interface{}{
		"data":   existingHeader,
		"status": "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}
//reassigning the header to different entity
func (s *HeaderVoiceManager) reassignHeader(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	var inputHeader,reassignedHeader, existingHeader HeaderVoice
	//checking the length of the input
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("reassignHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the input to the inputHeader object
	errHeader := json.Unmarshal([]byte(args[0]), &inputHeader)
	if errHeader != nil {
		repError= strings.Replace(errHeader.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("reassignHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//getting the header data from the ledger with cli
	existingHeaderResult, err := stub.GetState(inputHeader.CLI)
	if err != nil {
		errKey = inputHeader.CLI
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to fetch the Header- " + repError
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("reassignHeader: " + jsonResp)
		return shim.Error(jsonResp)
	} else if existingHeaderResult == nil {
		errKey = inputHeader.CLI
		errorDetails = "Header does not exist with CLI"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("reassignHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the data to the existingHeader object for update
	err = json.Unmarshal(existingHeaderResult, &existingHeader)
	if err != nil {
		repError= strings.Replace(err.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("reassignHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packing the data to store into the ledger
	reassignedHeader.ObjType = "HeaderVoice"
	reassignedHeader.HeaderId = existingHeader.HeaderId
	reassignedHeader.PEID = reassignedHeader.PEID
	reassignedHeader.CLI = existingHeader.CLI
	//checking for the status of the exisiting header
	if existingHeader.Status == "D" {
		reassignedHeader.Status = "A" // Active again after reassigning
	} else {
		errKey = existingHeader.PEID
		errorDetails = "Header is still active with peid, set header status to D(Delete) before reassigning"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("reassignHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packing the data to store into the ledger
	_, creatorUpdatedBy := s.getInvokerIdentity(stub)
	reassignedHeader.Category = inputHeader.Category
	reassignedHeader.CreateTs = existingHeader.CreateTs
	reassignedHeader.UpdateTs = inputHeader.UpdateTs
	reassignedHeader.Creator = existingHeader.Creator
	reassignedHeader.UpdatedBy = creatorUpdatedBy
	//marshalling the data to store into the ledger
	headerJSON, marshalErr := json.Marshal(reassignedHeader)
	if marshalErr != nil {
		repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("reassignHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//checking for the validity of the values before storing into the ledger
	if isValid, errMsg := IsValidHeaderIdPresent(reassignedHeader); !isValid {
		errorDetails = errMsg
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("reassignHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//storing the updated values into the ledger against the existing cli
	err = stub.PutState(reassignedHeader.CLI, headerJSON)
	if err != nil {
		errKey = string(headerJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save header with CLI- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("reassignHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//setting an event after recording into the ledger
	retErr := stub.SetEvent(_UpdateEvent, headerJSON)
	if retErr != nil {
		errKey = string(headerJSON)
		repError = strings.Replace(retErr.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : UPDATE_HEADER- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("reassignHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the response and return to the app layer
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"headerID": reassignedHeader.HeaderId,
		"message":  "Reassigning Header Successful",
		"header":   reassignedHeader,
		"status":"true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}
//reassignHeader is the function where one can assign the header to a entity once it is deactivated as per blockcube
// func (s *HeaderVoiceManager) reassignHeader(stub shim.ChaincodeStubInterface, args []string) sc.Response {
// 	var data map[string]interface{}
// 	HeaderStruct:=&HeaderVoice{}
// 	if len(args) != 1 {
// 		errKey = strconv.Itoa(len(args))
// 		errorDetails = "Invalid Number of Arguments"
// 		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
// 		_headerVoiceLogger.Errorf("reassignHeader: " + jsonResp)
// 		return shim.Error(jsonResp)
// 	}
// 	err := json.Unmarshal([]byte(args[0]), &data)
// 	if errHeader != nil {
// 		repError= strings.Replace(errHeader.Error(),"\""," ", -1)
// 		errorDetails = "Invalid JSON provided- "+ repError
// 		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
// 		_headerVoiceLogger.Errorf("queryByHeader: " + jsonResp)
// 		return shim.Error(jsonResp)
// 	}
// 	cat := map[string]int {"1":1, "2":1, "3":1, "4":1, "5":1, "6":1, "7":1, "8":1, "9":1}
// 	_, er := cat[data["ctgr"].(string)]
// 	if !er {
// 		logger.Info("Invalid Category ID. Must lie in between 1 to 8")
// 		return shim.Error("Invalid Category ID. Must lie in between 1 to 8") 
// 	}

// 	if len(data["cli"].(string)) > 11  || len(data["cli"].(string)) < 6 {
// 		logger.Infof("Cli : Header Name, is not a valid length i.e. Length shoud lie in between 6 to 11 digits ")
// 		return shim.Error("Cli : Header Name , is not a valid length i.e. Length shoud lie in between 6 to 11 digits")
// 	} 

// 	RecordAsBytes, err := stub.GetState(data["cli"].(string))
// 	if err != nil {
// 		logger.Infof(" Failed to get Header Record : " + data["cli"].(string) + " Error : " + string(err.Error()))
// 		return shim.Error(" Failed to get Header Record " + data["cli"].(string) + " Error : " + string(err.Error()))
// 	} else if RecordAsBytes == nil {
// 		logger.Infof(" Failed to get Header Record : " + data["cli"].(string) + " Error : Record Does not exist ")
// 		return shim.Error(" Failed to get Header Record " + data["cli"].(string) + " Error : Record Does not exist ")
// 	}

// 	if len(data) == 5 {

// 			header := Header{}
// 			err := json.Unmarshal(RecordAsBytes, &header)
// 			if err != nil {
// 				logger.Errorf("reassignHeader : Existing header data Unmarhsaling Error : " + string(err.Error()))
// 				return shim.Error("reassignHeader : Existing header data Unmarhsaling Error : " + string(err.Error()))
// 			}

// 			HeaderStruct.ObjType = "HeaderVoice"
// 			HeaderStruct.Header_ID = header.Header_ID
// 			HeaderStruct.PrincipleEntityId = data["peid"].(string)
// 			HeaderStruct.Cname = data["cname"].(string)
// 			HeaderStruct.Header_Name = header.Header_Name

// 			if header.Status == "D" {
// 					HeaderStruct.Status = "A" // Active again after reassigning
// 			} else {
// 				logger.Errorf("Header is still active with peid : " + header.PrincipleEntityId + "  Please set header status to Delete before reassigning")
// 				return shim.Error("Header is still active with peid : " + header.PrincipleEntityId + " Please set header status to Delete before reassigning")
// 			}

// 			HeaderStruct.Category = data["ctgr"].(string)
// 			HeaderStruct.CreatedTs = header.CreatedTs
// 			HeaderStruct.UpdatedTs = data["uts"].(string)
// 			HeaderStruct.Creator = header.Creator
// 			HeaderStruct.UpdatedBy = header.UpdatedBy
//       		HeaderStruct.Comment= "Header Reassigned with PrincipleEntityId : " +data["peid"].(string)
// 			logger.Infof("Header Name is " + HeaderStruct.Header_Name)
// 			headerAsBytes, err := json.Marshal(HeaderStruct)
// 			if err != nil {
// 				logger.Errorf("reassignHeader : Marshalling Error : " + string(err.Error()))
// 				return shim.Error("reassignHeader : Marshalling Error : " + string(err.Error()))
// 			}

// 			err = stub.PutState(HeaderStruct.Header_Name, headerAsBytes)
// 			if err != nil {
// 				logger.Errorf("reassignHeader : PutState Failed Error : " + string(err.Error()))
// 				return shim.Error("reassignHeader : PutState Failed Error : " + string(err.Error()))
// 			}
// 			logger.Info("reassignHeader : PutState Success : " + string(headerAsBytes))

// 			txid := stub.GetTxID()
// 			logger.Info("reassignHeader : Header is successfully reassigned for Header_Name : " + HeaderStruct.Header_Name + " , TransactionID      " + txid)
// 	} else {
// 		logger.Errorf("reassignHeader : Incorrect Number Of Arguments, i.e. 5 expected")
// 		return shim.Error("reassignHeader : Incorrect Number Of Arguments i.e. 5 expected")
// 	}

// 	resultData := map[string]interface{}{
// 		"trxnID":   stub.GetTxID(),
// 		"headerReassigned": data["cli"].(string),
// 		"message":  "Header is Successfully reassigned.",
// 		"Header":   HeaderStruct,
// 	}
// 	respJSON, _ := json.Marshal(resultData)
//     return shim.Success(respJSON)
// }
//get history of the header by providing cli from the ledger
func (s *HeaderVoiceManager) getHistoryByHeader(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	var qHeader HeaderVoice
	//checking the lenght of the input
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("getHistoryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the input to qHeader object
	errHeader := json.Unmarshal([]byte(args[0]), &qHeader)
	if errHeader != nil {
		repError= strings.Replace(errHeader.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("getHistoryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//calling the history results function and returning the data to the app layer
	historyResults, _ := getHistoryResults(stub, qHeader.CLI)
	return shim.Success(historyResults)
}

//function used for getting the history of a transaction
func getHistoryResults(stub shim.ChaincodeStubInterface, headerId string) ([]byte, error) {
	resultsIterator, err := stub.GetHistoryForKey(headerId)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	// buffer is a JSON array containing historic values for the user
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")
		buffer.WriteString(",\"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}
		buffer.WriteString(",\"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(",\"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return buffer.Bytes(), nil
}

//getDataByPagination will query the ledger on the selector input, and display using the pagination
func (s *HeaderVoiceManager) getDataByPagination(stub shim.ChaincodeStubInterface) peer.Response {
	type Query struct {
		SQuery   string `json:"sq"`
		PageSize string `json:"ps"`
		Bookmark string `json:"bm"`
	}
	_, args := stub.GetFunctionAndParameters()
	//checking the length of the input
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var tempQuery Query
	errHeader := json.Unmarshal([]byte(args[0]), &tempQuery)
	if errHeader != nil {
		repError= strings.Replace(errHeader.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	queryString := tempQuery.SQuery
	pageSize, _ := strconv.ParseInt(tempQuery.PageSize, 10, 32)
	bookMark := tempQuery.Bookmark
	paginationResults, err := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookMark)
	if err != nil {
		errKey = queryString + "," + tempQuery.PageSize + "," + bookMark
		errorDetails = "Could not fetch the data"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	fmt.Println(paginationResults)
	return shim.Success([]byte(paginationResults))
}

//function used for fetching the data from ledger using pagination
func getQueryResultForQueryStringWithPagination(stub shim.ChaincodeStubInterface, queryString string, pageSize int32, bookmark string) (string, error) {
	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)
	resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(queryString, pageSize, bookmark)
	if err != nil {
		return "", err
	}
	defer resultsIterator.Close()
	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return "", err
	}
	bufferWithPaginationInfo := addPaginationMetadataToQueryResults(buffer, responseMetadata)
	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", bufferWithPaginationInfo.String())
	return bufferWithPaginationInfo.String(), nil
}

//adding pagination metadata to results
func addPaginationMetadataToQueryResults(buffer *bytes.Buffer, responseMetadata *peer.QueryResponseMetadata) *bytes.Buffer {
	buffer.WriteString(",\"ResponseMetadata\":{\"RecordsCount\":")
	buffer.WriteString("\"")
	buffer.WriteString(fmt.Sprintf("%v", responseMetadata.FetchedRecordsCount))
	buffer.WriteString("\"")
	buffer.WriteString(", \"Bookmark\":")
	buffer.WriteString("\"")
	buffer.WriteString(responseMetadata.Bookmark)
	buffer.WriteString("\"}}")
	return buffer
}

//constructing result to the pagination query
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("{\"Records\":[")
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
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return &buffer, nil
}

//queryByHeader will query the header by cli from the ledger
func (s *HeaderVoiceManager) getDataByDateRange(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 4 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerVoiceLogger.Errorf("getDataByDateRange: " + jsonResp)
		return shim.Error(jsonResp)
	}
	startKey := args[0]
	endKey := args[1]
	pageSize := args[2]
	bookmark := args[3]
	querystring := fmt.Sprintf("{\"selector\":{\"$and\":[{\"cts\":{\"$gte\":\"%s\"}},{\"cts\":{\"$lte\":\"%s\"}}]}}", startKey, endKey)
	pageSizeInt, _ := strconv.ParseInt(pageSize, 10, 32)
	paginationResults, _ := getQueryResultForQueryStringWithPagination(stub, querystring, int32(pageSizeInt), bookmark)
	fmt.Println(paginationResults)
	return shim.Success([]byte(paginationResults))
}

//Function used for Complex & Rich Queries
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {
	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	// buffer is a JSON array containing QueryRecords
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
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())
	return buffer.Bytes(), nil
}

//used to get the identity of the invoker
func (s *HeaderVoiceManager) getInvokerIdentity(stub shim.ChaincodeStubInterface) (bool, string) {
	//Following id comes in the format X509::<Subject>::<Issuer>>
	enCert, err := id.GetX509Certificate(stub)
	if err != nil {
		return false, "Unknown."
	}
	issuersOrgs := enCert.Issuer.Organization
	if len(issuersOrgs) == 0 {
		return false, "Unknown.."
	}
	return true, fmt.Sprintf("%s", issuersOrgs[0])
}

//function to work with selector queries using index
func (s *HeaderVoiceManager) retrieveHeaderRecords(stub shim.ChaincodeStubInterface, criteria string, indexs ...string) []HeaderVoice {
	var finalSelector string                                   
	records := make([]HeaderVoice, 0)                                           
	if len(indexs) == 0 {                                        
		finalSelector = fmt.Sprintf("{\"selector\":%s }", criteria)
	} else {
		finalSelector = fmt.Sprintf("{\"selector\":%s , \"use_index\" :\"%s\" }", criteria, indexs[0])
	}
	_headerVoiceLogger.Infof("Query Selector : %s", finalSelector)
	resultsIterator, _ := stub.GetQueryResult(finalSelector)
	for resultsIterator.HasNext() {
		record := HeaderVoice{}
		recordBytes, _ := resultsIterator.Next()
		err := json.Unmarshal(recordBytes.Value, &record)
		if err != nil {
			repError = strings.Replace(err.Error(), "\"", " ", -1)
			errorDetails = "Invalid JSON provided- " + repError
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerVoiceLogger.Errorf("retrieveHeaderRecords: " + jsonResp)
		}
		records = append(records, record)
	}
	return records
}
