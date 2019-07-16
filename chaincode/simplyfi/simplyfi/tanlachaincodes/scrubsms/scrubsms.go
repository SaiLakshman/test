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

var _scrubSMSLogger = shim.NewLogger("Scrubbing")

const _CreateEvent = "INITIATE_SCRUBBING"
const _UpdateEvent = "UPDATE_SCRUB"
const _BulkEvent = "BULK_CREATE"

//Scrub structure defines the ledger record for any scrubbing
type ScrubSMS struct {
	ObjType           string `json:"obj"`   //DocType
	ScrubToken        string `json:"stok"`  // Scrub token unique -- search key
	PEID              string `json:"peid"`  //Enterprise Guideline
	TMID              string `json:"tmid"`  //Registered Telemarketer
	CLI               string `json:"cli"`   //Header Name
	TemplateID        string `json:"tid"`   //template unique id
	Category          string `json:"ctgr"`  //options (1-8)
	CommunicationType string `json:"ctyp"`  //options (P,SE,T,SI)
	ConsumedBy        string `json:"cby"`   //token consumedby
	Creator           string `json:"crtr"`  //creator of the scrub
	CreateTimeStamp   string `json:"cts"`   //scrub create time
	Status            string `json:"sts"`   // status of scrub - Default: A (Active)
	UpdatedBy         string `json:"uby"`   //updated by - Default: "Empty"
	UpdateTimeStamp   string `json:"uts"`   //updated time - Default: "Empty"
	ScrubbedFileName  string `json:"sFile"` //Scrubbed file name
	ScrubbedFileHash  string `json:"sHash"` //scrubbed file hash
}

//Scrubbing manages scrubb related transactions
type ScrubbingSMS struct {
}

var errorDetails, errKey, jsonResp, repError string

var scrubStatus = map[string]bool{
	"A": true,
	"C": true,
	"X": true,
	"P": true,
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

var communicationType = map[string]bool{
	"P":  true,
	"T":  true,
	"SE": true,
	"SI": true,
}

func validEnumEntry(input string, enumMap map[string]bool) bool {
	if _, isEntryExists := enumMap[input]; !isEntryExists {
		return false
	}
	return true
}

//IsValidScrubTokenPresent checks for  validity of scrubbing for update trxn
func IsValidScrubTokenPresent(s ScrubSMS) (bool, string) {

	if len(s.ScrubToken) == 0 {
		return false, "Scrub token is mandatory"
	}
	if len(s.PEID) == 0 {
		return false, "Principal Entity ID is mandatory"
	}
	if len(s.TMID) == 0 {
		return false, "Telemarketer ID is mandatory"
	}
	if len(s.CLI) == 0 {
		return false, "HeaderName(CLI) is mandatory"
	}
	if len(s.TemplateID) == 0 {
		return false, "Template ID is mandatory"
	}
	if !validEnumEntry(s.Category, category) {
		return false, "Category: Enter either 1, 2, 3, 4, 5, 6, 7, 8"
	}
	if !validEnumEntry(s.CommunicationType, communicationType) {
		return false, "Communication Type: Enter either A, C, X or P"
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
	if !validEnumEntry(s.Status, scrubStatus) {
		return false, "Status: Enter either A, C, X or P"
	}
	return true, ""
}

//InitiateScrubbing creates a scrubbing record in the ledger
func (s *ScrubbingSMS) createScrubDetails(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("createScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var scrubToSave ScrubSMS
	err := json.Unmarshal([]byte(args[0]), &scrubToSave)
	if err != nil {
		repError= strings.Replace(err.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("createScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//Second Check if the scrub token is existing or not
	if recordBytes, _ := stub.GetState(scrubToSave.ScrubToken); len(recordBytes) > 0 {
		errKey = scrubToSave.ScrubToken
		errorDetails = "Scrub with this scrubToken already Exist, provide unique scrubToken"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("createScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	scrubToSave.ObjType = "Scrubbing"
	_, creator := s.getInvokerIdentity(stub)
	scrubToSave.Creator = creator
	if scrubToSave.Status == "" {
		scrubToSave.Status = "A"
	}
	scrubToSave.UpdatedBy = creator
	scrubToSave.UpdateTimeStamp = scrubToSave.CreateTimeStamp
	scrubJSON, marshalErr := json.Marshal(scrubToSave)
	if marshalErr != nil {
		repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("createScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	if isValid, errMsg := IsValidScrubTokenPresent(scrubToSave); !isValid {
		errorDetails = errMsg
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("createScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	_scrubSMSLogger.Info("Saving Scrub Details to the ledger with token----------", scrubToSave.ScrubToken)
	err = stub.PutState(scrubToSave.ScrubToken, scrubJSON)
	if err != nil {
		errKey = string(scrubJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save scrub with scrubToken- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("createScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	retErr := stub.SetEvent(_CreateEvent, scrubJSON)
	if retErr != nil {
		errKey = string(scrubJSON)
		repError = strings.Replace(retErr.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : INITIATE_SCRUBBING- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("createScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}

	resultData := map[string]interface{}{
		"trxnID":  stub.GetTxID(),
		"stok":    scrubToSave.ScrubToken,
		"message": "Scrub data recorded successfully",
		"status":  "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//InitiateBulkScrubbing creates a scrubbing record in the ledger

func (s *ScrubbingSMS) createBulkScrubDetails(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("createBulkScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var listScrub []ScrubSMS
	err := json.Unmarshal([]byte(args[0]), &listScrub)
	if err != nil {
		repError= strings.Replace(err.Error(), "\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("createBulkScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var approvedStok []string
	var rejectedStok []string
	_, creator := s.getInvokerIdentity(stub)

	for i := 0; i < len(listScrub); i++ {
		var scrubToSave ScrubSMS
		scrubToSave = listScrub[i]
		if recordBytes, _ := stub.GetState(scrubToSave.ScrubToken); len(recordBytes) > 0 {
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		scrubToSave.Creator = creator
		scrubToSave.ObjType = "Scrubbing"
		if scrubToSave.Status == "" {
			scrubToSave.Status = "A"
		}
		scrubToSave.UpdatedBy = creator
		scrubToSave.UpdateTimeStamp = scrubToSave.CreateTimeStamp
		//Save the entry
		scrubJSON, marshalErr := json.Marshal(scrubToSave)
		if marshalErr != nil {
			repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
			errorDetails = "Cannot Marshal the JSON- " + repError
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_scrubSMSLogger.Errorf("createBulkScrubDetails: " + jsonResp)
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		if isValid, err := IsValidScrubTokenPresent(scrubToSave); !isValid {
			errorDetails = err
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_scrubSMSLogger.Errorf("createBulkScrubDetails: " + jsonResp)
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		_scrubSMSLogger.Info("Saving Scrub Details to the ledger with token----------", scrubToSave.ScrubToken)
		err = stub.PutState(scrubToSave.ScrubToken, scrubJSON)
		if err != nil {
			errKey = string(scrubJSON)
			repError = strings.Replace(err.Error(), "\"", " ", -1)
			errorDetails = "Unable to save Scrub details with Token- " + repError
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_scrubSMSLogger.Errorf("createBulkScrubDetails: " + jsonResp)
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		retErr := stub.SetEvent(_BulkEvent, scrubJSON)
		if retErr != nil {
			errKey = string(scrubJSON)
			repError = strings.Replace(retErr.Error(), "\"", " ", -1)
			errorDetails = "Event not generated for event : BULK_CREATE- " + repError
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_scrubSMSLogger.Errorf("createBulkScrubDetails: " + jsonResp)
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		approvedStok = append(approvedStok, scrubToSave.ScrubToken)
	}
	resultData := map[string]interface{}{
		"trxnID":  stub.GetTxID(),
		"stok_f":  rejectedStok,
		"message": "Scrub data recorded successfully",
		"status":  "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//updateStatus updates status of an existing scrub record
func (s *ScrubbingSMS) updateScrubStatus(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("updateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var updatedScrub ScrubSMS
	errScrub := json.Unmarshal([]byte(args[0]), &updatedScrub)
	if errScrub != nil {
		repError= strings.Replace(errScrub.Error(), "\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("updateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	scrubRecord, err := stub.GetState(updatedScrub.ScrubToken)
	if err != nil {
		errKey = updatedScrub.ScrubToken
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to fetch the Scrub details- " + repError
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("updateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	} else if scrubRecord == nil {
		errKey = updatedScrub.ScrubToken
		errorDetails = "Scrub details does not exist with Token"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("updateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var existingScrub ScrubSMS
	err = json.Unmarshal([]byte(scrubRecord), &existingScrub)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON for storing- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("updateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	_, creatorUpdatedBy := s.getInvokerIdentity(stub)
	existingScrub.UpdateTimeStamp = updatedScrub.UpdateTimeStamp
	existingScrub.UpdatedBy = creatorUpdatedBy
	existingScrub.Status = updatedScrub.Status
	existingScrub.ConsumedBy = updatedScrub.ConsumedBy

	scrubJSON, marshalErr := json.Marshal(existingScrub)
	if marshalErr != nil {
		repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("updateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	if isValid, errMsg := IsValidScrubTokenPresent(existingScrub); !isValid {
		errorDetails = errMsg
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("updateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	err = stub.PutState(existingScrub.ScrubToken, scrubJSON)
	if err != nil {
		errKey = string(scrubJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save Scrub Details with Token- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("updateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	retErr := stub.SetEvent(_UpdateEvent, scrubJSON)
	if retErr != nil {
		errKey = string(scrubJSON)
		repError = strings.Replace(retErr.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : UPDATE_SCRUB- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("updateScrubStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	resultData := map[string]interface{}{
		"trxnID":  stub.GetTxID(),
		"stok":    updatedScrub.ScrubToken,
		"message": "Scrub record status updated successfully",
		"status":  "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//queryScrubDetails function will fetch the scrub record from dlt given scrubtoken
func (s *ScrubbingSMS) queryScrubDetails(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("queryScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var qScrub ScrubSMS
	errScrub := json.Unmarshal([]byte(args[0]), &qScrub)
	if errScrub != nil {
		repError= strings.Replace(errScrub.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("queryScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	scrubRecord, err := stub.GetState(qScrub.ScrubToken)
	if err != nil {
		errKey = qScrub.ScrubToken
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to fetch Scrub Details- " + repError
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("queryScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	} else if scrubRecord == nil {
		errKey = qScrub.ScrubToken
		errorDetails = "Scrub details does not exist with Token"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("queryScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	record := ScrubSMS{}
	err = json.Unmarshal(scrubRecord, &record)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("queryScrubDetails: " + jsonResp)
		return shim.Error(jsonResp)
	}
	resultData := map[string]interface{}{
		"data":   record,
		"status": "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//queryScrub function queries the records from the ledger given any selector query using pagination
func (s *ScrubbingSMS) queryScrub(stub shim.ChaincodeStubInterface) peer.Response {
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
		_scrubSMSLogger.Errorf("queryScrub: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var tempQuery Query
	err := json.Unmarshal([]byte(args[0]), &tempQuery)
	if err != nil {
		repError= strings.Replace(err.Error(), "\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("queryScrub: " + jsonResp)
		return shim.Error(jsonResp)
	}
	queryString := tempQuery.SQuery
	pageSize, err1 := strconv.ParseInt(tempQuery.PageSize, 10, 32)
	if err1 != nil {
		errKey = string(tempQuery.PageSize)
		errorDetails = "PageSize should be a Number"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("queryScrub: " + jsonResp)
		return shim.Error(jsonResp)
	}
	bookMark := tempQuery.Bookmark
	paginationResults, err2 := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookMark)
	if err2 != nil {
		errKey = queryString + "," + string(pageSize) + "," + bookMark
		errorDetails = "Could not fetch the data"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_scrubSMSLogger.Errorf("queryScrub: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//fmt.Println(paginationResults)
	return shim.Success([]byte(paginationResults))
}

//anchor function for creating a response
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

//function to add the metadata for the response
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

//function to construct response to a json format
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
func (s *ScrubbingSMS) getInvokerIdentity(stub shim.ChaincodeStubInterface) (bool, string) {
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
