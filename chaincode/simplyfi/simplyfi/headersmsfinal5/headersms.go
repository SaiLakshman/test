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

var _headerSMSLogger = shim.NewLogger("HeaderSMS")

const _CreateEvent = "CREATE_HEADER"
const _UpdateEvent = "UPDATE_HEADER"
const _BulkCreateEvent = "BULK_CREATE"

//HeaderSMS structure defines the ledger record for any HeaderSMS's
type HeaderSMS struct {
	ObjType    string `json:"obj"`  //DocType
	HeaderId   string `json:"hid"`  // HeaderId unique
	PEID       string `json:"peid"` //creator of the scrub
	HeaderType string `json:"htyp"` // Type of the header (P,T,SE,SI)
	CLI        string `json:"cli"`  //HeaderName
	Creator    string `json:"crtr"` // creator of the txn in the ledger
	CreateTS   string `json:"cts"`  // created timestamp of the txn
	UpdatedBy  string `json:"uby"`  //updater of the txn in the ledger
	UpdateTs   string `json:"uts"`  //updated timestamp of the txn
	Status     string `json:"sts"`  // Status of the header (A,I,B,D)
	Category   string `json:"ctgr"` // category of header
	TMID       string `json:"tmid"` //telemarketer id
}

//HeaderSMSManager manages HeaderSMS related transactions in the ledger
type HeaderSMSManager struct {
}

var errorDetails, errKey, jsonResp, repError string

//valid status for header
var headerStatus = map[string]bool{
	"A": true,
	"I": true,
	"B": true,
	"D": true,
}

//valid header types
var headerType = map[string]bool{
	"P":  true,
	"T":  true,
	"SE": true,
	"SI": true,
}

//valid categories
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

//IsValidHeaderIdPresent checks for  validity of Header transaction before recording in the ledger
func IsValidHeaderIdPresent(s HeaderSMS) (bool, string) {
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
	if len(s.CreateTS) == 0 {
		return false, "CreateTS is mandatory"
	}
	if len(s.UpdatedBy) == 0 {
		return false, "UpdatedBy is mandatory"
	}
	if len(s.UpdateTs) == 0 {
		return false, "UpdateTS is mandatory"
	}
	if !validEnumEntry(s.HeaderType, headerType) {
		return false, "Header Type: Enter either P, T, SE, SI"
	}
	if !validEnumEntry(s.Status, headerStatus) {
		return false, "Status: Enter either A, I, B, D"
	}
	if !validEnumEntry(s.Category, category) {
		return false, "Category: Enter either 1, 2, 3, 4, 5, 6, 7, 8"
	}
	return true, ""
}

//registering the header in the ledger
func (s *HeaderSMSManager) registerHeader(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	var headerToSave HeaderSMS
	//checking the length of the input
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the input to headerToSave object
	err := json.Unmarshal([]byte(args[0]), &headerToSave)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON provided- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//checking whether header already exists or not with cli
	if recordBytes, _ := stub.GetState(headerToSave.CLI); len(recordBytes) > 0 {
		errKey = headerToSave.CLI
		errorDetails = "Header with this CLI already exist, provide unique CLI"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//checking whether header already exists or not with header id using selector query
	headerSearch := `{"obj":"HeaderSMS","hid":"%s"}`
	hID := headerToSave.HeaderId
	headerData := s.retriveHeaderRecords(stub, fmt.Sprintf(headerSearch, hID), "headerSearchByID")
	if len(headerData) > 0 {
		errKey = headerToSave.HeaderId
		errorDetails = "Header with HeaderId already Exists, Provide unique HeaderId"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the header with the details provided as input
	headerToSave.ObjType = "HeaderSMS"
	_, creator := s.getInvokerIdentity(stub)
	headerToSave.Creator = creator
	headerToSave.UpdatedBy = creator
	headerToSave.UpdateTs = headerToSave.CreateTS
	//marshalling the data for storing into the ledger
	headerJSON, marshalErr := json.Marshal(headerToSave)
	if marshalErr != nil {
		repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//checking for the validity of the header values before storing into the ledger
	if isValid, errMsg := IsValidHeaderIdPresent(headerToSave); !isValid {
		errKey = string(headerJSON)
		errorDetails = errMsg
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//storing the header into the ledger
	_headerSMSLogger.Info("Saving Header to the ledger with id----------", headerToSave.HeaderId)
	err = stub.PutState(headerToSave.CLI, headerJSON)
	if err != nil {
		errKey = string(headerJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save header with CLI- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//setting event after storing into the ledger
	retErr := stub.SetEvent(_CreateEvent, headerJSON)
	if retErr != nil {
		errKey = string(headerJSON)
		repError = strings.Replace(retErr.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : CREATE_HEADER- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the reponse and returning to the app layer
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"headerID": headerToSave.HeaderId,
		"message":  "Header Registered Successfully",
		"header":   headerToSave,
		"status":   "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//registerBulkHeader will create headers from the input at once
func (s *HeaderSMSManager) registerBulkHeader(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	var recordcount int
	rejectedHeaders := make([]map[string]interface{}, 0)
	registeredHeaders := make([]string, 0)
	recordcount = 0
	var listHeader []HeaderSMS
	//checking the length of the input
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerBulkHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the input to headerToSave object
	err := json.Unmarshal([]byte(args[0]), &listHeader)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON provided- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerBulkHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//iterating through  the list of headers to save
	_, creator := s.getInvokerIdentity(stub)
	for i := 0; i < len(listHeader); i++ {
		var headerToSave HeaderSMS
		headerToSave = listHeader[i]
		//checking whether header already exists or not with cli(header name)
		if recordBytes, _ := stub.GetState(headerToSave.CLI); len(recordBytes) > 0 {
			errKey = headerToSave.CLI
			errorDetails = "Header With CLI already Exists, Provide unique CLI(HeaderName) "
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerSMSLogger.Errorf("registerBulkHeader: " + jsonResp)
			rejectedHeaders = append(rejectedHeaders, map[string]interface{}{"Header_Name: ": headerToSave.CLI, "Value": "Header Name Already Exists"})
			continue
		}
		//checking whether header already exists or not with header id using selector query
		headerSearch := `{"obj":"HeaderSMS","hid":"%s"}`
		hID := headerToSave.HeaderId
		headerData := s.retriveHeaderRecords(stub, fmt.Sprintf(headerSearch, hID), "headerSearchByID")
		if len(headerData) > 0 {
			errKey = headerToSave.HeaderId
			errorDetails = "Header with HeaderId already Exists, Provide unique HeaderId"
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerSMSLogger.Errorf("registerBulkHeader: " + jsonResp)
			rejectedHeaders = append(rejectedHeaders, map[string]interface{}{"Header_Name: ": headerToSave.CLI, "Value": "Header ID already exists"})
			continue
		}
		//packaging the header with the details provided as input
		headerToSave.ObjType = "HeaderSMS"
		headerToSave.Creator = creator
		headerToSave.UpdatedBy = creator
		headerToSave.UpdateTs = headerToSave.CreateTS
		//marshalling the data for storing into the ledger
		headerJSON, marshalErr := json.Marshal(headerToSave)
		if marshalErr != nil {
			repError = strings.Replace(err.Error(), "\"", " ", -1)
			errorDetails = "Cannot Marshal the JSON- " + repError
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerSMSLogger.Errorf("registerBulkHeader: " + jsonResp)
			rejectedHeaders = append(rejectedHeaders, map[string]interface{}{"Header_Name: ": headerToSave.CLI, "Value": "Marshalling Error "})
			continue
		}
		//checking for the validity of the header values before storing into the ledger
		if isValid, err := IsValidHeaderIdPresent(headerToSave); !isValid {
			errKey = string(headerJSON)
			errorDetails = err
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerSMSLogger.Errorf("registerBulkHeader: " + jsonResp)
			rejectedHeaders = append(rejectedHeaders, map[string]interface{}{"Header_Name: ": headerToSave.CLI, "Value": err})
			continue
		}
		//storing the header into the ledger
		_headerSMSLogger.Info("Saving Header to the ledger with id----------", headerToSave.HeaderId)
		err = stub.PutState(headerToSave.CLI, headerJSON)
		if err != nil {
			errKey = string(headerJSON)
			repError = strings.Replace(err.Error(), "\"", " ", -1)
			errorDetails = "Unable to save header with CLI- " + repError
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerSMSLogger.Errorf("registerBulkHeader: " + jsonResp)
			rejectedHeaders = append(rejectedHeaders, map[string]interface{}{"Header_Name: ": headerToSave.CLI, "Value": "PutState Failed"})
			continue
		} else if err == nil {
			recordcount = recordcount + 1
		}
		//setting event after storing into the ledger
		retErr := stub.SetEvent(_BulkCreateEvent, headerJSON)
		if retErr != nil {
			errKey = string(headerJSON)
			repError = strings.Replace(retErr.Error(), "\"", " ", -1)
			errorDetails = "Event not generated for event : BULK_CREATE- " + repError
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerSMSLogger.Errorf("registerBulkHeader: " + jsonResp)
			rejectedHeaders = append(rejectedHeaders, map[string]interface{}{"Header_Name: ": headerToSave.CLI, "Value": "Event not generated for EVTRegisterHeader"})
			continue
		}
		registeredHeaders = append(registeredHeaders, headerToSave.CLI)
	}
	//packaging the response and return to the app layer
	resultData := map[string]interface{}{
		"trxnID":       stub.GetTxID(),
		"header_f":     rejectedHeaders,
		"message":      "Bulk Header Registered Successfully",
		"status":       "true",
		"successCount": strconv.Itoa(recordcount),
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//updateStatus will update the status of the header given headerId in the ledger
func (s *HeaderSMSManager) updateStatus(stub shim.ChaincodeStubInterface) peer.Response {
	var jsonResp string
	var updatedHeader, existingHeader HeaderSMS
	_, args := stub.GetFunctionAndParameters()
	//checking the length of the input
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the input to the updatedHeader object
	errHeader := json.Unmarshal([]byte(args[0]), &updatedHeader)
	if errHeader != nil {
		repError = strings.Replace(errHeader.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON provided- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//getting the data from the ledger with headername as input
	existingHeaderResult, err := stub.GetState(updatedHeader.CLI)
	if err != nil {
		errKey = updatedHeader.CLI
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Could not fetch details for the Header- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	} else if existingHeaderResult == nil {
		errKey = updatedHeader.CLI
		errorDetails = "Header does not exist with CLI"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the data retrieved from the ledger to existingHeader object for update
	err = json.Unmarshal([]byte(existingHeaderResult), &existingHeader)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
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
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//checking for the validity of the values before storing into the ledger
	if isValid, errMsg := IsValidHeaderIdPresent(existingHeader); !isValid {
		errKey = string(headerJSON)
		errorDetails = errMsg
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//storing the updated values into the ledger against the existing cli
	err = stub.PutState(existingHeader.CLI, headerJSON)
	if err != nil {
		errKey = string(headerJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save header with CLI- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//setting an event after recording into the ledger
	retErr := stub.SetEvent(_UpdateEvent, headerJSON)
	if retErr != nil {
		errKey = string(headerJSON)
		repError = strings.Replace(retErr.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : UPDATE_HEADER- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the response and return to the app layer
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"headerID": existingHeader.HeaderId,
		"message":  "Update Successful",
		"header":   existingHeader,
		"status":   "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//queryByHeader will query the header by cli from the ledger
func (s *HeaderSMSManager) queryByHeader(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	//checking the length of the input
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("queryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var qHeader, existingHeader HeaderSMS
	//unmarshalling the input to the qHeader object
	errHeader := json.Unmarshal([]byte(args[0]), &qHeader)
	if errHeader != nil {
		repError= strings.Replace(errHeader.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("queryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//getting the values from the ledger against CLI(header name)
	existingHeaderResult, err := stub.GetState(qHeader.CLI)
	if err != nil {
		errKey = qHeader.CLI
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to fetch the Header- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("queryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	} else if existingHeaderResult == nil {
		errKey = qHeader.CLI
		errorDetails = "Header does not exist with CLI"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("queryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the existingresult from the ledger to existingHeader object
	err = json.Unmarshal([]byte(existingHeaderResult), &existingHeader)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("queryByHeader: " + jsonResp)
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

//get history of the header by providing cli from the ledger
func (s *HeaderSMSManager) getHistoryByHeader(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	//checking the length of the input
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("getHistoryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var qHeader HeaderSMS
	//unmarshalling the input to the qHeader object
	err := json.Unmarshal([]byte(args[0]), &qHeader)
	if err != nil {
		errKey = args[0]
		errorDetails = "Invalid JSON provided"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("getHistoryByHeader: " + jsonResp)
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
func (s *HeaderSMSManager) getDataByPagination(stub shim.ChaincodeStubInterface) peer.Response {
	type Query struct {
		SQuery   string `json:"sq"`
		PageSize string `json:"ps"`
		Bookmark string `json:"bm"`
	}
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var tempQuery Query
	errHeader := json.Unmarshal([]byte(args[0]), &tempQuery)
	if errHeader != nil {
		repError= strings.Replace(errHeader.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	queryString := tempQuery.SQuery
	pageSize, err1 := strconv.ParseInt(tempQuery.PageSize, 10, 32)
	if err1 != nil {
		errKey = strconv.Itoa(pageSize)
		errorDetails = "PageSize should be a Number"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	bookMark := tempQuery.Bookmark
	paginationResults, err := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookMark)
	if err != nil {
		errKey = queryString + "," + strconv.Itoa(pageSize) + "," + bookMark
		errorDetails = "Could not fetch the data"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("getDataByPagination: " + jsonResp)
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
func (s *HeaderSMSManager) getDataByDateRange(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 4 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("getDataByDateRange: " + jsonResp)
		return shim.Error(jsonResp)
	}
	startKey := args[0]
	endKey := args[1]
	pageSize := args[2]
	bookmark := args[3]
	querystring := fmt.Sprintf("{\"selector\":{\"$and\":[{\"cts\":{\"$gte\":\"%s\"}},{\"cts\":{\"$lte\":\"%s\"}}]}}", startKey, endKey)
	pageSizeInt, err := strconv.ParseInt(pageSize, 10, 32)
	if err != nil {
		errKey = string(pageSize)
		errorDetails = "PageSize should be a Number"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("getDataByDateRange: " + jsonResp)
		return shim.Error(jsonResp)
	}
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

//function used to get the identity of the invoker
func (s *HeaderSMSManager) getInvokerIdentity(stub shim.ChaincodeStubInterface) (bool, string) {
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
func (s *HeaderSMSManager) retriveHeaderRecords(stub shim.ChaincodeStubInterface, criteria string, indexs ...string) []HeaderSMS {
	var finalSelector string
	records := make([]HeaderSMS, 0)
	if len(indexs) == 0 {
		finalSelector = fmt.Sprintf("{\"selector\":%s }", criteria)
	} else {
		finalSelector = fmt.Sprintf("{\"selector\":%s , \"use_index\" :\"%s\" }", criteria, indexs[0])
	}
	_headerSMSLogger.Infof("Query Selector : %s", finalSelector)
	resultsIterator, _ := stub.GetQueryResult(finalSelector)
	for resultsIterator.HasNext() {
		record := HeaderSMS{}
		recordBytes, _ := resultsIterator.Next()
		err := json.Unmarshal(recordBytes.Value, &record)
		if err != nil {
			repError = strings.Replace(err.Error(), "\"", " ", -1)
			errorDetails = "Invalid JSON provided- " + repError
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerSMSLogger.Errorf("retrieveHeaderRecords: " + jsonResp)
		}
		records = append(records, record)
	}
	return records
}
