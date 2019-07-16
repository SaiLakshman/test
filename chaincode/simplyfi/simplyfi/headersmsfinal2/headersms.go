package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
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

}

//HeaderSMSManager manages HeaderSMS related transactions
type HeaderSMSManager struct {
}

var errorDetails, errKey, jsonResp string
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
func IsValidHeaderIdPresent(s HeaderSMS) (bool, string) {
	if len(s.HeaderId) == 0 {
		return false, "HeaderId is mandatory"
	}
	if len(s.PEID) == 0 {
		return false, "Principal EntityID is mandatory"
	}
	if len(s.HeaderType) != 0 {
		if !validEnumEntry(s.HeaderType, headerType) {
			return false, "Header Type: Enter either P, T, SE, SI"
		}
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
	if len(args) < 1 {
		errKey = string(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var headerToSave HeaderSMS
	err := json.Unmarshal([]byte(args[0]), &headerToSave)
	if err != nil {
		errKey = args[0]
		errorDetails = "Invalid JSON provided"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//Second Check if the header already exists or not
	if recordBytes, _ := stub.GetState(headerToSave.CLI); len(recordBytes) > 0 {
		errKey = string(headerToSave.CLI)
		errorDetails = "Header with this CLI already exist, provide unique CLI"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	queryString := fmt.Sprintf("{\"selector\":{\"hid\":\"%s\"}}", headerToSave.HeaderId)
	queryresults, _ := getQueryResultForQueryString(stub, queryString)
	if len(queryresults) > 2 {
		errKey = string(headerToSave.HeaderId)
		errorDetails = "Header with HeaderId already Exists, Provide unique HeaderId"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	headerToSave.ObjType = "HeaderSMS"
	_, creator := s.getInvokerIdentity(stub)
	headerToSave.Creator = creator
	headerToSave.UpdatedBy = creator
	headerToSave.UpdateTs = headerToSave.CreateTS
	headerJSON, _ := json.Marshal(headerToSave)
	if isValid, errMsg := IsValidHeaderIdPresent(headerToSave); !isValid {
		errKey = string(headerJSON)
		errorDetails = string(errMsg)
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	_headerSMSLogger.Info("Saving Header to the ledger with id----------", headerToSave.HeaderId)
	err = stub.PutState(headerToSave.CLI, headerJSON)
	if err != nil {
		errKey = string(headerJSON)
		errorDetails = "Unable to save header with CLI - " + string(err.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	retErr := stub.SetEvent(_CreateEvent, headerJSON)
	if retErr != nil {
		errKey = string(headerJSON)
		errorDetails = "Event not generated for event : CREATE_HEADER- " + string(retErr.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"headerID": headerToSave.HeaderId,
		"message":  "Save Successful",
		"header":   headerToSave,
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//registerBulkHeader will create headers from the input at once
func (s *HeaderSMSManager) registerBulkHeader(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = string(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerBulkHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var listHeader []HeaderSMS
	err := json.Unmarshal([]byte(args[0]), &listHeader)
	if err != nil {
		errKey = args[0]
		errorDetails = "Invalid JSON provided"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("registerBulkHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var rejectedHeaders []string
	_, creator := s.getInvokerIdentity(stub)
	for i := 0; i < len(listHeader); i++ {
		var headerToSave HeaderSMS
		headerToSave = listHeader[i]
		if recordBytes, _ := stub.GetState(headerToSave.CLI); len(recordBytes) > 0 {
			rejectedHeaders = append(rejectedHeaders, headerToSave.CLI)
			continue
		}
		queryString := fmt.Sprintf("{\"selector\":{\"hid\":\"%s\"}}", headerToSave.HeaderId)
		queryresults, _ := getQueryResultForQueryString(stub, queryString)
		if len(queryresults) > 2 {
			rejectedHeaders = append(rejectedHeaders, headerToSave.CLI)
			continue
		}
		headerToSave.ObjType = "HeaderSMS"
		headerToSave.Creator = creator
		headerToSave.UpdatedBy = creator
		headerToSave.UpdateTs = headerToSave.CreateTS
		//Save the entry
		headerJSON, _ := json.Marshal(headerToSave)
		if isValid, err := IsValidHeaderIdPresent(headerToSave); !isValid {
			errKey = string(headerJSON)
			errorDetails = string(err)
			jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerSMSLogger.Errorf("registerBulkHeader: " + jsonResp)
			rejectedHeaders = append(rejectedHeaders, headerToSave.CLI)
			continue
		}
		_headerSMSLogger.Info("Saving Header to the ledger with id----------", headerToSave.HeaderId)
		err = stub.PutState(headerToSave.CLI, headerJSON)
		if err != nil {
			errKey = string(headerJSON)
			errorDetails = "Unable to save header with CLI - " + string(err.Error())
			jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerSMSLogger.Errorf("registerBulkHeader: " + jsonResp)
			rejectedHeaders = append(rejectedHeaders, headerToSave.CLI)
			continue
		}
		retErr := stub.SetEvent(_BulkCreateEvent, headerJSON)
		if retErr != nil {
			errKey = string(headerJSON)
			errorDetails = "Event not generated for event : BULK_CREATE- " + string(retErr.Error())
			jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_headerSMSLogger.Errorf("registerBulkHeader: " + jsonResp)
			rejectedHeaders = append(rejectedHeaders, headerToSave.CLI)
			continue
		}
	}
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"header_f": rejectedHeaders,
		"message":  "Header Registered Successfully",
		"status":   "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//updateStatus will update the status of the header given headerId in the ledger
func (s *HeaderSMSManager) updateStatus(stub shim.ChaincodeStubInterface) peer.Response {
	var jsonResp string
	var updatedHeader, existingHeader HeaderSMS
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = string(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	errHeader := json.Unmarshal([]byte(args[0]), &updatedHeader)
	if errHeader != nil {
		errKey = args[0]
		errorDetails = "Invalid JSON provided"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	existingHeaderResult, err := stub.GetState(updatedHeader.CLI)
	if err != nil {
		errKey = string(updatedHeader.CLI)
		errorDetails = "Could not fetch details for the Header- " + string(err.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	} else if existingHeaderResult == nil {
		errKey = string(updatedHeader.CLI)
		errorDetails = "Header does not exist with CLI"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	err = json.Unmarshal([]byte(existingHeaderResult), &existingHeader)
	if err != nil {
		errKey = string(existingHeaderResult)
		errorDetails = "Invalid JSON for storing" + string(err.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	_, creatorUpdatedBy := s.getInvokerIdentity(stub)
	existingHeader.UpdateTs = updatedHeader.UpdateTs
	existingHeader.UpdatedBy = creatorUpdatedBy
	existingHeader.Status = updatedHeader.Status
	//validity check
	headerJSON, _ := json.Marshal(existingHeader)
	if isValid, errMsg := IsValidHeaderIdPresent(existingHeader); !isValid {
		errKey = string(headerJSON)
		errorDetails = string(errMsg)
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	err = stub.PutState(existingHeader.CLI, headerJSON)
	if err != nil {
		errKey = string(headerJSON)
		errorDetails = "Unable to save header with CLI - " + string(err.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	retErr := stub.SetEvent(_UpdateEvent, headerJSON)
	if retErr != nil {
		errKey = string(headerJSON)
		errorDetails = "Event not generated for event : UPDATE_HEADER- " + string(retErr.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"headerID": existingHeader.HeaderId,
		"message":  "Update Successful",
		"header":   existingHeader,
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//queryByHeader will query the header by cli from the ledger
func (s *HeaderSMSManager) queryByHeader(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = string(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("queryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var qHeader, existingHeader HeaderSMS
	errHeader := json.Unmarshal([]byte(args[0]), &qHeader)
	if errHeader != nil {
		errKey = args[0]
		errorDetails = "Invalid JSON provided"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("queryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	existingHeaderResult, err := stub.GetState(qHeader.CLI)
	if err != nil {
		errKey = string(qHeader.CLI)
		errorDetails = "Unable to fetch the Header - " + string(err.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("queryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	} else if existingHeaderResult == nil {
		errKey = string(qHeader.CLI)
		errorDetails = "Header does not exist with CLI"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("queryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	err = json.Unmarshal([]byte(existingHeaderResult), &existingHeader)
	if err != nil {
		errKey = string(existingHeaderResult)
		errorDetails = "Invalid JSON - " + string(err.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("queryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
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
	if len(args) < 1 {
		errKey = string(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("getHistoryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var qHeader HeaderSMS
	err := json.Unmarshal([]byte(args[0]), &qHeader)
	if err != nil {
		errKey = args[0]
		errorDetails = "Invalid JSON provided"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("getHistoryByHeader: " + jsonResp)
		return shim.Error(jsonResp)
	}
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
		errKey = string(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var tempQuery Query
	errHeader := json.Unmarshal([]byte(args[0]), &tempQuery)
	if errHeader != nil {
		errKey = args[0]
		errorDetails = "Invalid JSON provided"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	queryString := tempQuery.SQuery
	pageSize, err1 := strconv.ParseInt(tempQuery.PageSize, 10, 32)
	if err1 != nil {
		errKey = string(pageSize)
		errorDetails = "PageSize should be a Number"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_headerSMSLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	bookMark := tempQuery.Bookmark
	paginationResults, err := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookMark)
	if err != nil {
		errKey = queryString + "," + string(pageSize) + "," + bookMark
		errorDetails = "Could not fetch the data"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
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
		errKey = string(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
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
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
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

//Returns the complete identity in the format
//Certitificate issuer orgs's domain name
//Returns string Unkown if not able parse the invoker certificate
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
