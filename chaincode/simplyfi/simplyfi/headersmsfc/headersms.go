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
			return false, "Enter either P, T, SE, SI"
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
		return false, "Enter either A, I, B, D"
	}
	if !validEnumEntry(s.Category, category) {
		return false, "Enter either 1, 2, 3, 4, 5, 6, 7, 8"
	}
	return true, ""
}

//registering the header in the ledger
func (s *HeaderSMSManager) registerHeader(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}
	var headerToSave HeaderSMS
	err := json.Unmarshal([]byte(args[0]), &headerToSave)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}
	//Second Check if the headerId is existing or not
	if recordBytes, _ := stub.GetState(headerToSave.HeaderId); len(recordBytes) > 0 {
		return shim.Error("Header with this HeaderId already exist, Provide unique HeaderId")
	}
	headerToSave.ObjType = "HeaderSMS"
	_, creator := s.getInvokerIdentity(stub)
	headerToSave.Creator = creator
	headerToSave.UpdatedBy = creator
	headerToSave.UpdateTs = headerToSave.CreateTS
	//Save the entry
	headerJSON, _ := json.Marshal(headerToSave)
	if isValid, errMsg := IsValidHeaderIdPresent(headerToSave); !isValid {
		return shim.Error(errMsg)
	}
	_headerSMSLogger.Info("headerToSave.HeaderId----------", headerToSave.HeaderId)
	err = stub.PutState(headerToSave.HeaderId, headerJSON)

	if err != nil {
		return shim.Error("Unable to save with HeaderId " + headerToSave.HeaderId)
	}
	retErr := stub.SetEvent(_CreateEvent, headerJSON)
	if retErr != nil {
		_headerSMSLogger.Errorf("Event not generated for event : CREATE_HEADER")
		return shim.Error("{\"error\":\"Unable to save Header.\"}")
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
		return shim.Error("Invalid number of arguments provided for transaction")
	}
	var listHeader []HeaderSMS
	err := json.Unmarshal([]byte(args[0]), &listHeader)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}
	var rejectedHeaders []string
	_, creator := s.getInvokerIdentity(stub)
	for i := 0; i < len(listHeader); i++ {
		var headerToSave HeaderSMS
		headerToSave = listHeader[i]
		if recordBytes, _ := stub.GetState(headerToSave.HeaderId); len(recordBytes) > 0 {
			rejectedHeaders = append(rejectedHeaders, headerToSave.HeaderId)
			continue
		}
		headerToSave.ObjType = "HeaderSMS"
		headerToSave.Creator = creator
		headerToSave.UpdatedBy = creator
		headerToSave.UpdateTs = headerToSave.CreateTS
		//Save the entry
		headerJSON, _ := json.Marshal(headerToSave)
		if isValid, _ := IsValidHeaderIdPresent(headerToSave); !isValid {
			rejectedHeaders = append(rejectedHeaders, headerToSave.HeaderId)
			continue
		}
		_headerSMSLogger.Info("headerToSave.HeaderId----------", headerToSave.HeaderId)
		err = stub.PutState(headerToSave.HeaderId, headerJSON)

		if err != nil {
			rejectedHeaders = append(rejectedHeaders, headerToSave.HeaderId)
			continue
		}
		retErr := stub.SetEvent(_BulkCreateEvent, headerJSON)
		if retErr != nil {
			_headerSMSLogger.Errorf("Event not generated for event : BULK_CREATE")
			rejectedHeaders = append(rejectedHeaders, headerToSave.HeaderId)
			continue
		}
	}
	//Second Check if the scrub token is existing or not
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
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid arguments provided")
	}
	var updatedHeader HeaderSMS
	errHeader := json.Unmarshal([]byte(args[0]), &updatedHeader)
	if errHeader != nil {
		return shim.Error(errHeader.Error())
	}
	querystring := fmt.Sprintf("{\"selector\":{\"cli\":\"%s\"}}", updatedHeader.CLI)
	queryresults, _ := getQueryResultForQueryString(stub, querystring)
	records := []HeaderSMS{}
	err1 := json.Unmarshal(queryresults, &records)
	if err1 != nil {
		jsonResp = "Failed to get state for " + records[0].CLI
		resultData := map[string]interface{}{
			"message": jsonResp,
			"status":  "false",
		}
		respJson, _ := json.Marshal(resultData)
		return shim.Success(respJson)
	} else if queryresults == nil {
		jsonResp = "Header does not exist: " + records[0].CLI
		resultData := map[string]interface{}{
			"message": jsonResp,
			"status":  "false",
		}
		respJson, _ := json.Marshal(resultData)
		return shim.Success(respJson)
	}
	_, creatorUpdatedBy := s.getInvokerIdentity(stub)
	records[0].UpdateTs = updatedHeader.UpdateTs
	records[0].UpdatedBy = creatorUpdatedBy
	records[0].Status = updatedHeader.Status
	if isValid, errMsg := IsValidHeaderIdPresent(records[0]); !isValid {
		return shim.Error(errMsg)
	}
	marshalHeaderJSON, _ := json.Marshal(records[0])
	finalErr := stub.PutState(records[0].CLI, marshalHeaderJSON)
	if finalErr != nil {
		jsonResp = "{\"Error\":\"Unable to save the Header with Id " + records[0].HeaderId + "\"}"
		return shim.Success([]byte(jsonResp))
	}
	retErr := stub.SetEvent(_UpdateEvent, marshalHeaderJSON)
	if retErr != nil {
		_headerSMSLogger.Errorf("Event not generated for event : UPDATE_HEADER")
		return shim.Error("{\"error\":\"Unable to change status of Header.\"}")
	}

	// headerRecord, err := stub.GetState(updatedHeader.HeaderId)
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }
	// var existingHeader HeaderSMS
	// errExistingHeader := json.Unmarshal([]byte(headerRecord), &existingHeader)
	// if errExistingHeader != nil {
	// 	return shim.Error(errExistingHeader.Error())
	// }
	// _, creatorUpdatedBy := s.getInvokerIdentity(stub)
	// existingHeader.UpdateTs = updatedHeader.UpdateTs
	// existingHeader.UpdatedBy = creatorUpdatedBy
	// existingHeader.Status = updatedHeader.Status

	// if isValid, errMsg := IsValidHeaderIdPresent(existingHeader); !isValid {
	// 	return shim.Error(errMsg)
	// }
	// marshalHeaderJSON, _ := json.Marshal(existingHeader)
	// finalErr := stub.PutState(existingHeader.HeaderId, marshalHeaderJSON)
	// if finalErr != nil {
	// 	return shim.Error("Unable to save with Header with id " + updatedHeader.HeaderId)
	// }
	// retErr := stub.SetEvent(_UpdateEvent, marshalHeaderJSON)
	// if retErr != nil {
	// 	_headerSMSLogger.Errorf("Event not generated for event : UPDATE_HEADER")
	// 	return shim.Error("{\"error\":\"Unable to change status of Header.\"}")
	// }
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"headerID": records[0].HeaderId,
		"message":  "Save Successful",
		"header":   records[0],
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//queryByHeader will query the header by cli from the ledger
func (s *HeaderSMSManager) queryByHeader(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid arguments provided")
	}
	var jsonResp string
	var qHeader HeaderSMS
	errHeader := json.Unmarshal([]byte(args[0]), &qHeader)
	if errHeader != nil {
		return shim.Error(errHeader.Error())
	}
	querystring := fmt.Sprintf("{\"selector\":{\"cli\":\"%s\"}}", qHeader.CLI)
	queryresults, _ := getQueryResultForQueryString(stub, querystring)
	records := []HeaderSMS{}
	err1 := json.Unmarshal(queryresults, &records)
	if err1 != nil {
		jsonResp = "Failed to get state for " + qHeader.CLI
		resultData := map[string]interface{}{
			"message": jsonResp,
			"status":  "false",
		}
		respJson, _ := json.Marshal(resultData)
		return shim.Success(respJson)
	} else if queryresults == nil {
		jsonResp = "Header does not exist: " + qHeader.CLI
		resultData := map[string]interface{}{
			"message": jsonResp,
			"status":  "false",
		}
		respJson, _ := json.Marshal(resultData)
		return shim.Success(respJson)
	}
	resultData := map[string]interface{}{
		"data":   records,
		"status": "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//queryByHeader will query the header by cli from the ledger
func (s *HeaderSMSManager) getByDateRange(stub shim.ChaincodeStubInterface) peer.Response {
	type input struct {
		startKey string `json:"sk"`
		endKey   string `json:"ek"`
	}
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid arguments provided")
	}
	//var jsonResp string
	var qHeader input
	errHeader := json.Unmarshal([]byte(args[0]), &qHeader)
	if errHeader != nil {
		return shim.Error(errHeader.Error())
	}
	querystring := fmt.Sprintf("{\"selector\":{\"$and\":[{\"cts\":{\"$gte\":\"%s\"}},{\"cts\":{\"$lte\":\"%s\"}}]}}", qHeader.startKey, qHeader.endKey)
	//	queryresults, _ := getQueryResultForQueryString(stub, querystring)
	fmt.Println(querystring)
	return shim.Success([]byte(qHeader.startKey))
	// records := []HeaderSMS{}
	// err1 := json.Unmarshal(queryresults, &records)
	// if err1 != nil {
	// 	jsonResp = "Failed to get state for " + qHeader.startKey
	// 	resultData := map[string]interface{}{
	// 		"message": jsonResp,
	// 		"status":  "false",
	// 	}
	// 	respJson, _ := json.Marshal(resultData)
	// 	return shim.Success(respJson)
	// } else if queryresults == nil {
	// 	jsonResp = "Header does not exist: " + qHeader.startKey
	// 	resultData := map[string]interface{}{
	// 		"message": jsonResp,
	// 		"status":  "false",
	// 	}
	// 	respJson, _ := json.Marshal(resultData)
	// 	return shim.Success(respJson)
	// }
	// resultData := map[string]interface{}{
	// 	"data":   records,
	// 	"status": "true",
	// }
	// respJSON, _ := json.Marshal(resultData)
	// return shim.Success(respJSON)
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

//get history of the header by providing cli from the ledger
func (s *HeaderSMSManager) getHistoryByHeader(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid arguments provided")
	}
	var jsonResp string
	var qHeader HeaderSMS
	errHeader := json.Unmarshal([]byte(args[0]), &qHeader)
	if errHeader != nil {
		jsonResp = "{\"Error\":\"Unmarshalling \"}"
		return shim.Success([]byte(jsonResp))
	}
	querystring := fmt.Sprintf("{\"selector\":{\"cli\":\"%s\"}}", qHeader.CLI)
	queryresults, _ := getQueryResultForQueryString(stub, querystring)
	records := []HeaderSMS{}
	err1 := json.Unmarshal(queryresults, &records)
	if err1 != nil {
		jsonResp = "Failed to get state for " + qHeader.CLI
		resultData := map[string]interface{}{
			"message": jsonResp,
			"status":  "false",
		}
		respJson, _ := json.Marshal(resultData)
		return shim.Success(respJson)
	} else if queryresults == nil {
		jsonResp = "Header does not exist: " + qHeader.CLI
		resultData := map[string]interface{}{
			"message": jsonResp,
			"status":  "false",
		}
		respJson, _ := json.Marshal(resultData)
		return shim.Success(respJson)
	}
	hid := records[0].HeaderId
	historyResults, _ := getHistoryResults(stub, hid)
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
		return shim.Error("Invalid arguments provided")
	}
	var tempQuery Query
	errHeader := json.Unmarshal([]byte(args[0]), &tempQuery)
	if errHeader != nil {
		return shim.Error(errHeader.Error())
	}
	queryString := tempQuery.SQuery
	pageSize, _ := strconv.ParseInt(tempQuery.PageSize, 10, 32)
	bookMark := tempQuery.Bookmark
	paginationResults, _ := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookMark)
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
