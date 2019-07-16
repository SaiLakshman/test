package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	id "github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
	"github.com/hyperledger/fabric/protos/peer"
)

var _msgSMSLogger = shim.NewLogger("MessageDelivery")

const _CreateEvent = "INITIATE_MSGDELIVERY"
const _BulkCreateEvent = "BULK_CREATE"

//MSGDelivery structure defines the ledger record.
type MSGDelivery struct {
	ObjType          string `json:"obj"`    //DocType
	ScrubToken       string `json:"stok"`   // Scrub token unique -- search key
	Creator          string `json:"crtr"`   //creator of the scrub
	CreateTimeStamp  string `json:"cts"`    //scrub create time
	ScrubbedFileName string `json:"sFile"`  //Scrubbed file name
	ScrubbedFileHash string `json:"sHash"`  //scrubbed file hash
	ServiceProvider  string `json:"svcprv"` // service provider who created this scrubbing
}

//MSGDeliveryManages manages MSGDelivery related transactions
type MSGDeliveryManager struct {
}

var errorDetails, errKey, jsonResp, repError string
var svcProvider = map[string]bool{
	"AI": true,
	"VO": true,
	"ID": true,
	"BL": true,
	"ML": true,
	"QL": true,
	"TA": true,
	"JI": true,
	"VI": true,
}

func validEnumEntry(input string, enumMap map[string]bool) bool {
	if _, isEntryExists := enumMap[input]; !isEntryExists {
		return false
	}
	return true
}

//IsValidScrubTokenPresent checks for  validity of scrubbing
func IsValidScrubTokenPresent(s MSGDelivery) (bool, string) {

	if len(s.ScrubToken) == 0 {
		return false, "Scrub token should be present there"
	}
	if len(s.Creator) == 0 {
		return false, "Scrub Creator is mandatory"
	}
	if len(s.ScrubbedFileName) == 0 {
		return false, "Scrub file name is mandatory"
	}
	if len(s.ScrubbedFileHash) == 0 {
		return false, "Scrub file hash is mandatory"
	}
	if !validEnumEntry(s.ServiceProvider, svcProvider) {
		return false, "ServiceProvider: Enter either AI, VO, ID, BL, ML, QL, TA, JI or VI"
	}
	return true, ""
}

//IsValid checks if the scrub fields are valid or not
func IsValid(s MSGDelivery) (bool, string) {

	if len(s.ScrubToken) == 0 {
		return false, "Scrub token should be present there"
	}
	if len(s.Creator) == 0 {
		return false, "Scrub Creator is mandatory"
	}
	if len(s.ScrubbedFileName) == 0 {
		return false, "Scrub file name is mandatory"
	}
	if len(s.ScrubbedFileHash) == 0 {
		return false, "Scrub file hash is mandatory"
	}
	if len(s.CreateTimeStamp) == 0 {
		return false, "Create Time Stamp is mandatory"
	}
	if !validEnumEntry(s.ServiceProvider, svcProvider) {
		return false, "ServiceProvider: Enter either AI, VO, ID, BL, ML, QL, TA, JI or VI"
	}
	return true, ""
}

//createMSGDelivery creates a MSGDelivery record in the ledger
func (s *MSGDeliveryManager) createMSGDelivery(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("createMSGDelivery: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var msgToSave MSGDelivery
	err := json.Unmarshal([]byte(args[0]), &msgToSave)
	if err != nil {
		repError= strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("createMSGDelivery: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//Second Check if the scrub token is existing or not
	if recordBytes, _ := stub.GetState(msgToSave.ScrubToken); len(recordBytes) > 0 {
		errKey = msgToSave.ScrubToken
		errorDetails = "Scrub with this scrubToken already Exist, provide unique scrubToken"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("createMSGDelivery: " + jsonResp)
		return shim.Error(jsonResp)
	}
	msgToSave.ObjType = "msgDelivery"
	_, creator := s.getInvokerIdentity(stub)
	msgToSave.Creator = creator
	scrubJSON, marshalErr:= json.Marshal(msgToSave)
	if marshalErr != nil {
		repError= strings.Replace(marshalErr.Error(), "\""," ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("createMSGDelivery: "+ jsonResp)
		return shim.Error(jsonResp)
	}
	if isValid, errMsg := IsValid(msgToSave); !isValid {
		errorDetails = errMsg
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("createMSGDelivery: " + jsonResp)
		return shim.Error(jsonResp)
	}
	_msgSMSLogger.Info("Saving MSGDelivery Details to the ledger with token----------", msgToSave.ScrubToken)
	err = stub.PutState(msgToSave.ScrubToken, scrubJSON)
	if err != nil {
		errKey = string(scrubJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save scrub with scrubToken- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("createMSGDelivery: " + jsonResp)
		return shim.Error(jsonResp)
	}
	retErr := stub.SetEvent(_CreateEvent, scrubJSON)
	if retErr != nil {
		errKey = string(scrubJSON)
		repError = strings.Replace(retErr.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : INITIATE_MSGDELIVERY- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("createMSGDelivery: " + jsonResp)
		return shim.Error(jsonResp)
	}
	resultData := map[string]interface{}{
		"trxnID":  stub.GetTxID(),
		"stok":    msgToSave.ScrubToken,
		"uby":     msgToSave.Creator,
		"uts":     msgToSave.CreateTimeStamp,
		"message": "MSGDelivery data recorded successfully",
		"status":  "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//createBulkMSGDelivery will create headers from the input at once
func (s *MSGDeliveryManager) createBulkMSGDelivery(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("createBulkMSGDelivery: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var listScrub []MSGDelivery
	err := json.Unmarshal([]byte(args[0]), &listScrub)
	if err != nil {
		repError= strings.Replace(err.Error(), "\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("createBulkMSGDelivery: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var rejectedStok []string
	_, creator := s.getInvokerIdentity(stub)
	for i := 0; i < len(listScrub); i++ {
		var scrubToSave MSGDelivery
		scrubToSave = listScrub[i]
		if recordBytes, _ := stub.GetState(scrubToSave.ScrubToken); len(recordBytes) > 0 {
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		scrubToSave.Creator = creator
		scrubToSave.ObjType = "msgDelivery"
		scrubJSON, marshalErr := json.Marshal(scrubToSave)
		if marshalErr != nil {
			repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
			errorDetails = "Cannot Marshal the JSON- " + repError
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_msgSMSLogger.Errorf("createBulkMSGDelivery: " + jsonResp)
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		if isValid, errMsg := IsValid(scrubToSave); !isValid {
			errorDetails = errMsg
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_msgSMSLogger.Errorf("createBulkMSGDelivery: " + jsonResp)
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		_msgSMSLogger.Info("scrubToSave.ScrubToken----------", scrubToSave.ScrubToken)
		err = stub.PutState(scrubToSave.ScrubToken, scrubJSON)
		if err != nil {
			errKey = string(scrubJSON)
			repError = strings.Replace(err.Error(), "\"", " ", -1)
			errorDetails = "Unable to save scrub with scrubToken- " + repError
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_msgSMSLogger.Errorf("createBulkMSGDelivery: " + jsonResp)
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		retErr := stub.SetEvent(_BulkCreateEvent, scrubJSON)
		if retErr != nil {
			errKey = string(scrubJSON)
			repError = strings.Replace(retErr.Error(), "\"", " ", -1)
			errorDetails = "Event not generated for event : BULK_CREATE- " + repError
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_msgSMSLogger.Errorf("createBulkMSGDelivery: " + jsonResp)
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
	}
	//Second Check if the scrub token is existing or not
	resultData := map[string]interface{}{
		"trxnID":  stub.GetTxID(),
		"stok_f":  rejectedStok,
		"message": "Scrub data recorded successfully",
		"status":  "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//queryMSGDelivery will query the messageDelivery by scrubToken from the ledger
func (s *MSGDeliveryManager) queryMSGDelivery(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("queryMSGDelivery: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var jsonResp string
	var qScrub MSGDelivery
	errScrub := json.Unmarshal([]byte(args[0]), &qScrub)
	if errScrub != nil {
		repError= strings.Replace(errScrub.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("queryMSGDelivery: " + jsonResp)
		return shim.Error(jsonResp)
	}
	scrubRecord, err := stub.GetState(qScrub.ScrubToken)
	if err != nil {
		errKey = qScrub.ScrubToken
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to fetch MSGDelivery Details- " + repError
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("queryMSGDelivery: " + jsonResp)
		return shim.Error(jsonResp)
	} else if scrubRecord == nil {
		errKey = qScrub.ScrubToken
		errorDetails = "MSGDelivery details does not exist with Token"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("queryMSGDelivery: " + jsonResp)
		return shim.Error(jsonResp)
	}
	record := MSGDelivery{}
	err = json.Unmarshal(scrubRecord, &record)
	if err != nil {
		errKey = string(scrubRecord)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON- " + repError
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("queryMSGDelivery: " + jsonResp)
		return shim.Error(jsonResp)
	}
	resultData := map[string]interface{}{
		"data":   record,
		"status": "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//getDataByPagination will query the ledger on the selector input, and display using the pagination
func (s *MSGDeliveryManager) getDataByPagination(stub shim.ChaincodeStubInterface) peer.Response {
	type Query struct {
		SQuery   string `json:"sq"`
		PageSize string `json:"ps"`
		Bookmark string `json:"bm"`
	}
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("getDataByPagination " + jsonResp)
		return shim.Error(jsonResp)
	}
	var tempQuery Query
	err := json.Unmarshal([]byte(args[0]), &tempQuery)
	if err != nil {
		repError= strings.Replace(err.Error(), "\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	queryString := tempQuery.SQuery
	pageSize, err1 := strconv.ParseInt(tempQuery.PageSize, 10, 32)
	if err1 != nil {
		errKey = tempQuery.PageSize
		errorDetails = "PageSize should be a Number"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	bookMark := tempQuery.Bookmark
	paginationResults, err2 := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookMark)
	if err2 != nil {
		errKey = queryString + "," + strconv.Itoa(pageSize) + "," + bookMark
		repError= strings.Replace(err2.Error(), "\""," ", -1)
		errorDetails = "Could not fetch the data- "+ repError
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_msgSMSLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	fmt.Println(paginationResults)
	return shim.Success([]byte(paginationResults))
}

// //getDataByPagination will query the ledger on the selector input, and display using the pagination
// func (s *MSGDeliveryManager) getDataByPagination(stub shim.ChaincodeStubInterface) peer.Response {
// 		_, args := stub.GetFunctionAndParameters()
// 	if len(args) < 3 {
// 		errKey = strconv.Itoa(len(args))
// 		errorDetails = "Invalid Number of Arguments"
// 		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
// 		_msgSMSLogger.Errorf("getDataByPagination " + jsonResp)
// 		return shim.Error(jsonResp)
// 	}
// 	queryString := args[0]
// 	pageSize, err1 := strconv.ParseInt(args[1], 10, 32)
// 	if err1 != nil {
// 		errKey = args[1]
// 		errorDetails = "PageSize should be a Number"
// 		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
// 		_msgSMSLogger.Errorf("getDataByPagination: " + jsonResp)
// 		return shim.Error(jsonResp)
// 	}
// 	bookMark := args[2]
// 	paginationResults, err2 := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookMark)
// 	if err2 != nil {
// 		errKey = queryString + "," + string(pageSize) + "," + bookMark
// 		repError= strings.Replace(err2.Error(), "\""," ", -1)
// 		errorDetails = "Could not fetch the data- "+ repError
// 		jsonResp = "{\"Data\":\{"Data":"{"obj":"Template","urn":"120756782783","peid":"120145678912","cli":["Header1","Header2"],"tname":"Template1","ttyp":"CS","ctyp":"P","csty":"1","coty":"U","vars":"1","ctgr":"1","tcont":"Dear Subscriber, This is Template Content","tmid":"12345","crtr":"org1","cts":"1511478902","uby":"org1","uts":"1511478902","sts":"A"}","ErrorDetails":"Communication Type: SE"}"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
// 		_msgSMSLogger.Errorf("g{"Data":"{"obj":"Template","urn":"120756782783","peid":"120145678912","cli":["Header1","Header2"],"tname":"Template1","ttyp":"CS","ctyp":"P","csty":"1","coty":"U","vars":"1","ctgr":"1","tcont":"Dear Subscriber, This is Template Content","tmid":"12345","crtr":"org1","cts":"1511478902","uby":"org1","uts":"1511478902","sts":"A"}","ErrorDetails":"Communication Type: SE"}etDataByPagination: " + jsonResp)
// 		return shim.Error(jsonR{"Data":"{"obj":"Template","urn":"120756782783","peid":"120145678912","cli":["Header1","Header2"],"tname":"Template1","ttyp":"CS","ctyp":"P","csty":"1","coty":"U","vars":"1","ctgr":"1","tcont":"Dear Subscriber, This is Template Content","tmid":"12345","crtr":"org1","cts":"1511478902","uby":"org1","uts":"1511478902","sts":"A"}","ErrorDetails":"Communication Type: SE"}esp)
// 	}
// 	fmt.Println(paginationResults)
// 	return shim.Success([]byte(paginationResults))
// }

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
func (s *MSGDeliveryManager) getInvokerIdentity(stub shim.ChaincodeStubInterface) (bool, string) {
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
