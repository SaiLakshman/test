package main

import (
	"encoding/json"
	"fmt"
	"bytes"
	"strconv"
	"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	id "github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
	"github.com/hyperledger/fabric/protos/peer"
)

var _scrubVoiceLogger = shim.NewLogger("Scrubbing")

const _CreateEvent = "INITIATE_SCRUBBING"
const _UpdateEvent = "UPDATE_SCRUB"

//Scrub structure defines the ledger record for any scrubbing
type ScrubVoice struct {
	ObjType           string `json:"obj"`
	ScrubToken        string `json:"stok"`
	PEID  			  string `json:"peid"`
	TMID     		  string `json:"tmid"`
	CLI				  string `json:"cli"`
	CNAME             string `json:"cname"`
	TemplateID 		  string `json:"tid"`
	Category          string `json:"ctgr"`
	CommunicationMode string `json:"cmode"`
	DayTimeBand       string `json:"time"`
	CommunicationType string `json:"ctyp"`
	Creator           string `json:"crtr"`
	CreateTs          string `json:"cts"`
	ConsumedBy        string `json:"csby"`
	Status            string `json:"sts"`
	SourceFileName    string `json:"ifile"`
	SourceFileHash    string `json:"iHash"`
	ScrubbedFileName  string `json:"ofile"`
	ScrubbedFileHash  string `json:"ohash"`
	UpdatedBy         string `json:"uby"`
	UpdateTs          string `json:"uts"`
}

//Scrubbing manages scrubb related transactions
type ScrubbingVoice struct {
}

var errorDetails, errKey, jsonResp, repError string

var scrubStatus = map[string]bool{
	"A": true,
	"C": true,
	"X": true,
	"P": true,
}

var commType = map[string]bool{
	"P": true,
	"T": true,
	"SE": true,
	"SI": true,
}
var commMode = map[string]bool{
	"11": true,
	"12": true,
	"13": true,
	"14": true,
	"15": true,
}
var timeBand = map[string]bool{
	"21": true,
	"22": true,
	"23": true,
	"24": true,
	"25": true,
	"26": true,
	"27": true,
	"28": true,
	"29": true,
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

//IsValidScrubDataPresent checks for  validity of scrubbing for all the trxns
func IsValidScrubDataPresent(s ScrubVoice) (bool, string) {

	if len(s.ScrubToken) == 0 {
		return false, "Scrub token is mandatory"
	}
	if len(s.PEID) == 0 {
		return false, "Principal EntityID is mandatory"
	}
	if len(s.CLI) == 0 {
		return false, "CLI is mandatory"
	}
	if len(s.CNAME) == 0 {
		return false, "CNAME is mandatory"
	}
	if len(s.TemplateID) == 0 {
		return false, "TemplateID is mandatory"
	}
	if len(s.Creator) == 0 {
		return false, "Scrub Creator is mandatory"
	}
	if len(s.CreateTs) == 0 {
		return false, "Created Timestamp is mandatory"
	}
	if len(s.SourceFileName) == 0 {
		return false, "Source File name is mandatory"
	}
	if len(s.SourceFileHash) == 0 {
		return false, "Source File hash is mandatory"
	}
	if len(s.ScrubbedFileName) == 0 {
		return false, "Scrub file name is mandatory"
	}
	if len(s.ScrubbedFileHash) == 0 {
		return false, "Scrub file hash is mandatory"
	}
	if len(s.UpdatedBy) == 0 {
		return false, "UpdatedBy is mandatory"
	}
	if len(s.UpdateTs) == 0 {
		return false, "Updated Timestamp is mandatory"
	}
	if !validEnumEntry(s.Status, scrubStatus) {
		return false, "Enter either A, C, X or P"
	}
	if !validEnumEntry(s.CommunicationType, commType) {
		return false, "Enter either P, T, SE or SI"
	}
	if !validEnumEntry(s.Category, category) {
		return false, "Enter either 1, 2, 3, 4, 5, 6, 7 or 8"
	}
	if !validEnumEntry(s.CommunicationMode, commMode) {
		return false, "Enter either 11, 12, 13, 14 or 15"
	}
	if len(s.DayTimeBand) == 0 {
		retrun true, ""
	}else if !validEnumEntry(s.DayTimeBand, timeBand) {
		return false, "Enter either 21, 22, 23, 24, 25, 26, 27, 28 or 29"
	}
	return true, ""
}

//InitiateScrubbing creates a scrubbing record in the ledger
func (s *ScrubbingVoice) createScrubDetails(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	var scrubToSave ScrubVoice
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VcreateScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	err := json.Unmarshal([]byte(args[0]), &scrubToSave)
	if err != nil {
		repError= strings.Replace(err.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VcreateScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//Second Check if the scrub token is existing or not
	if recordBytes, _ := stub.GetState(scrubToSave.ScrubToken); len(recordBytes) > 0 {
		errKey = scrubToSave.ScrubToken
		errorDetails = "Scrub with this scrubToken already Exist, provide unique scrubToken"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VcreateScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	scrubToSave.ObjType = "VScrubbing"
	_, creator := s.getInvokerIdentity(stub)
	scrubToSave.Creator = creator
	if scrubToSave.Status == "" {
		scrubToSave.Status = "A"
	}
	scrubToSave.UpdatedBy = creator
	scrubToSave.UpdateTs = scrubToSave.CreateTs
	//Save the entry
	scrubJSON, marshalErr := json.Marshal(scrubToSave)
	if marshalErr != nil {
		repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VcreateScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	if isValid, errMsg := IsValidScrubDataPresent(scrubToSave); !isValid {
		errorDetails = errMsg
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VcreateScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	_scrubVoiceLogger.Info("Saving Scrub Details to the ledger with token----------", scrubToSave.ScrubToken)
	err = stub.PutState(scrubToSave.ScrubToken, scrubJSON)
	if err != nil {
		errKey = string(scrubJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save scrub with scrubToken- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VcreateScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	retErr := stub.SetEvent(_CreateEvent, scrubJSON)
	if retErr != nil {
		errKey = string(scrubJSON)
		repError = strings.Replace(retErr.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : INITIATE_SCRUBBING- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VcreateScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"stok": 	scrubToSave.ScrubToken,
		"message":  "Scrub data recorded successfully",
		"status": "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}


//InitiateBulkScrubbing creates a scrubbing record in the ledger
// To Do : If one scrub record is not validated, this function will skip all the next scrub record, 
// To keep track of each scrub record in the array we need to create two list of successful and unsuccessful
// scrub
func (s *ScrubbingVoice) createBulkScrubDetails(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VcreateBulkScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var listScrub []ScrubVoice
	err := json.Unmarshal([]byte(args[0]), &listScrub)
	if err != nil {
		repError= strings.Replace(err.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VcreateBulkScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var approvedStok []string
	var rejectedStok []string
	_, creator := s.getInvokerIdentity(stub)
	
	for i := 0; i < len(listScrub); i++ {
		var scrubToSave ScrubVoice
		scrubToSave = listScrub[i]
		if recordBytes, _ := stub.GetState(scrubToSave.ScrubToken); len(recordBytes) > 0 {
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		scrubToSave.Creator = creator	
		scrubToSave.ObjType = "VScrubbing"
		if scrubToSave.Status == "" {
			scrubToSave.Status = "A"
		}
		scrubToSave.UpdatedBy = creator
		scrubToSave.UpdateTs = scrubToSave.CreateTs
		//Save the entry
		scrubJSON, marshalErr := json.Marshal(scrubToSave)
		if marshalErr != nil {
			repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
			errorDetails = "Cannot Marshal the JSON- " + repError
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_scrubVoiceLogger.Errorf("VcreateBulkScrubDetails: " + jsonResp)
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		if isValid, err := IsValidScrubDataPresent(scrubToSave); !isValid {
			errorDetails = err
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_scrubVoiceLogger.Errorf("VcreateBulkScrubDetails: " + jsonResp)
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		_scrubVoiceLogger.Info("Saving Scrub Details to the ledger with token----------", scrubToSave.ScrubToken)
		err = stub.PutState(scrubToSave.ScrubToken, scrubJSON)
		if err != nil {
			errKey = string(scrubJSON)
			repError = strings.Replace(err.Error(), "\"", " ", -1)
			errorDetails = "Unable to save Scrub details with Token- " + repError
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_scrubVoiceLogger.Errorf("VcreateBulkScrubDetails: " + jsonResp)
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		retErr := stub.SetEvent(_CreateEvent, scrubJSON)
		if retErr != nil {
			errKey = string(scrubJSON)
			repError = strings.Replace(retErr.Error(), "\"", " ", -1)
			errorDetails = "Event not generated for event : BULK_CREATE- " + repError
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_scrubVoiceLogger.Errorf("VcreateBulkScrubDetails: " + jsonResp)
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		 approvedStok = append(approvedStok, scrubToSave.ScrubToken)
	}
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"stok_f": rejectedStok,
		"message":  "Scrub data recorded successfully",
		"status": "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//updateStatus updates status of an existing scrub record
func (s *ScrubbingVoice) updateScrubStatus(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VupdateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var updatedScrub ScrubVoice
	errScrub := json.Unmarshal([]byte(args[0]), &updatedScrub)
	if errScrub != nil {
		repError= strings.Replace(errScrub.Error(), "\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VupdateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	scrubRecord, err := stub.GetState(updatedScrub.ScrubToken)
	if err != nil {
		errKey = updatedScrub.ScrubToken
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to fetch the Scrub details- " + repError
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VupdateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	} else if scrubRecord == nil {
		errKey = updatedScrub.ScrubToken
		errorDetails = "Scrub details does not exist with Token"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VupdateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var existingScrub ScrubVoice
	err = json.Unmarshal([]byte(scrubRecord), &existingScrub)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON for storing- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VupdateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	_, creatorUpdatedBy := s.getInvokerIdentity(stub)
	// updatedScrub.ObjType = "Scrubbing"
	existingScrub.UpdateTs = updatedScrub.UpdateTs
	existingScrub.UpdatedBy = creatorUpdatedBy
	existingScrub.Status = updatedScrub.Status
	scrubJSON, marshalErr  := json.Marshal(existingScrub)
	if marshalErr != nil {
		repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VupdateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	if isValid, errMsg := IsValidScrubDataPresent(existingScrub); !isValid {
		errorDetails = errMsg
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VupdateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	err = stub.PutState(existingScrub.ScrubToken, scrubJSON)
	if err != nil {
		errKey = string(scrubJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save Scrub Details with Token- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VupdateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	retErr := stub.SetEvent(_UpdateEvent, scrubJSON)
	if retErr != nil {
		errKey = string(scrubJSON)
		repError = strings.Replace(retErr.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : UPDATE_SCRUB- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VupdateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"stok": updatedScrub.ScrubToken,
		"message":  "Scrub record status updated successfully",
		"status": "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

func (s *ScrubbingVoice) queryScrubDetails(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VqueryScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}

	var jsonResp string
	var qScrub ScrubVoice
	errScrub := json.Unmarshal([]byte(args[0]), &qScrub)
	if errScrub != nil {
		repError= strings.Replace(errScrub.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VqueryScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	scrubRecord, err := stub.GetState(qScrub.ScrubToken)
	if err != nil {
		errKey = qScrub.ScrubToken
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to fetch Scrub Details- " + repError
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VqueryScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	} else if scrubRecord == nil {
		errKey = qScrub.ScrubToken
		errorDetails = "Scrub details does not exist with Token"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VqueryScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	record := ScrubVoice{}
	err1 := json.Unmarshal(scrubRecord, &record)
	if err1 != nil {
		repError = strings.Replace(err1.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VqueryScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	resultData := map[string]interface{}{
		"data": record,
		"status": "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}
func (s *ScrubbingVoice) queryScrub(stub shim.ChaincodeStubInterface) peer.Response {
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
		_scrubVoiceLogger.Errorf("VqueryScrub: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var tempQuery Query
	err := json.Unmarshal([]byte(args[0]), &tempQuery)
	if err != nil {
		repError= strings.Replace(err.Error(), "\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VqueryScrub: " + jsonResp)
		return shim.Error(jsonResp)
	}
	queryString := tempQuery.SQuery
	pageSize,err1 := strconv.ParseInt(tempQuery.PageSize, 10, 32)
	if err1 != nil {
		errKey = string(tempQuery.PageSize)
		errorDetails = "PageSize should be a Number"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VqueryScrub: " + jsonResp)
		return shim.Error(jsonResp)
	}
	bookMark := tempQuery.Bookmark
	paginationResults, err2 := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookMark)
	if err2 != nil {
		errKey = queryString + "," + string(pageSize) + "," + bookMark
		errorDetails = "Could not fetch the data"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubVoiceLogger.Errorf("VqueryScrub: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//fmt.Println(paginationResults)
	return shim.Success([]byte(paginationResults))
}

// func (s *ScrubbingVoice) queryScrub(stub shim.ChaincodeStubInterface) peer.Response {
// 	_, args := stub.GetFunctionAndParameters()
// 	if len(args) < 3 {
// 		errKey = strconv.Itoa(len(args))
// 		errorDetails = "Invalid Number of Arguments"
// 		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
// 		_scrubVoiceLogger.Errorf("VqueryScrub: " + jsonResp)
// 		return shim.Error(jsonResp)
// 	}
// 	queryString := args[0]
// 	pageSize,err1 := strconv.ParseInt(args[1], 10, 32)
// 	if err1 != nil {
// 		errKey = string(args[1])
// 		errorDetails = "PageSize should be a Number"
// 		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
// 		_scrubVoiceLogger.Errorf("VqueryScrub: " + jsonResp)
// 		return shim.Error(jsonResp)
// 	}
// 	bookMark := args[2]
// 	paginationResults, err2 := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookMark)
// 	if err2 != nil {
// 		errKey = queryString + "," + string(pageSize) + "," + bookMark
// 		errorDetails = "Could not fetch the data"
// 		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
// 		_scrubVoiceLogger.Errorf("VqueryScrub: " + jsonResp)
// 		return shim.Error(jsonResp)
// 	}
// 	//fmt.Println(paginationResults)
// 	return shim.Success([]byte(paginationResults))
// }


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
func (s *ScrubbingVoice) getInvokerIdentity(stub shim.ChaincodeStubInterface) (bool, string) {
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

