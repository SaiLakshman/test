package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

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
		return false, "Enter either 1, 2, 3, 4, 5, 6, 7, 8"
	}
	if !validEnumEntry(s.CommunicationType, communicationType) {
		return false, "Enter either A, C, X or P"
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
		return false, "Enter either A, C, X or P"
	}
	return true, ""
}

//InitiateScrubbing creates a scrubbing record in the ledger
func (s *ScrubbingSMS) CreateScrubDetails(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}
	var scrubToSave ScrubSMS
	err := json.Unmarshal([]byte(args[0]), &scrubToSave)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}
	//Second Check if the scrub token is existing or not
	if recordBytes, _ := stub.GetState(scrubToSave.ScrubToken); len(recordBytes) > 0 {
		return shim.Error("Scrub with this scrubToken already Exist, provide unique scrubToken")
	}
	scrubToSave.ObjType = "Scrubbing"
	_, creator := s.getInvokerIdentity(stub)
	scrubToSave.Creator = creator
	if scrubToSave.Status == "" {
		scrubToSave.Status = "A"
	}
	scrubToSave.UpdatedBy = creator
	scrubToSave.UpdateTimeStamp = scrubToSave.CreateTimeStamp
	//Save the entry
	scrubJSON, _ := json.Marshal(scrubToSave)
	if isValid, errMsg := IsValidScrubTokenPresent(scrubToSave); !isValid {
		return shim.Error(errMsg)
	}
	_scrubSMSLogger.Info("scrubToSave.ScrubToken----------", scrubToSave.ScrubToken)
	err = stub.PutState(scrubToSave.ScrubToken, scrubJSON)

	if err != nil {
		return shim.Error("Unable to save with scrub Token " + scrubToSave.ScrubToken)
	}
	retErr := stub.SetEvent(_CreateEvent, scrubJSON)

	if retErr != nil {
		_scrubSMSLogger.Errorf("Event not generated for event : INITIATE_SCRUBBING")
		return shim.Error("{\"Error\":\"Unable save scrubbing.\"}")
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
// To Do : If one scrub record is not validated, this function will skip all the next scrub record,
// To keep track of each scrub record in the array we need to create two list of successful and unsuccessful
// scrub
func (s *ScrubbingSMS) CreateBulkScrubDetails(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}
	var listScrub []ScrubSMS
	err := json.Unmarshal([]byte(args[0]), &listScrub)
	if err != nil {
		return shim.Error("Invalid json provided as input")
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
		scrubJSON, _ := json.Marshal(scrubToSave)
		if isValid, _ := IsValidScrubTokenPresent(scrubToSave); !isValid {
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		_scrubSMSLogger.Info("scrubToSave.ScrubToken----------", scrubToSave.ScrubToken)
		err = stub.PutState(scrubToSave.ScrubToken, scrubJSON)

		if err != nil {
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		retErr := stub.SetEvent(_BulkEvent, scrubJSON)

		if retErr != nil {
			_scrubSMSLogger.Errorf("Event not generated for event : BULK_CREATE")
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
func (s *ScrubbingSMS) UpdateScrubStatus(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid arguments provided")
	}
	var updatedScrub ScrubSMS
	errScrub := json.Unmarshal([]byte(args[0]), &updatedScrub)
	if errScrub != nil {
		return shim.Error(errScrub.Error())
	}
	scrubRecord, err := stub.GetState(updatedScrub.ScrubToken)
	if err != nil {
		return shim.Error(err.Error())
	}
	var existingScrub ScrubSMS
	errExistingScrub := json.Unmarshal([]byte(scrubRecord), &existingScrub)
	if errExistingScrub != nil {
		return shim.Error(errExistingScrub.Error())
	}
	_, creatorUpdatedBy := s.getInvokerIdentity(stub)
	existingScrub.UpdateTimeStamp = updatedScrub.UpdateTimeStamp
	existingScrub.UpdatedBy = creatorUpdatedBy
	existingScrub.Status = updatedScrub.Status
	existingScrub.ConsumedBy = updatedScrub.ConsumedBy
	if isValid, errMsg := IsValidScrubTokenPresent(existingScrub); !isValid {
		return shim.Error(errMsg)
	}
	marshalScrubJSON, _ := json.Marshal(existingScrub)
	finalErr := stub.PutState(existingScrub.ScrubToken, marshalScrubJSON)

	if finalErr != nil {
		return shim.Error("Unable to save with scrub with id " + updatedScrub.ScrubToken)
	}
	retErr := stub.SetEvent(_UpdateEvent, marshalScrubJSON)

	if retErr != nil {
		_scrubSMSLogger.Errorf("Event not generated for event : UPDATE_SCRUB")
		return shim.Error("{\"Error\":\"Unable to change status of scrub.\"}")
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
func (s *ScrubbingSMS) QueryScrubDetails(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid arguments provided")
	}
	var jsonResp string
	var qScrub ScrubSMS
	errScrub := json.Unmarshal([]byte(args[0]), &qScrub)
	if errScrub != nil {
		return shim.Error(errScrub.Error())
	}

	scrubRecord, err := stub.GetState(qScrub.ScrubToken)
	if err != nil {
		return shim.Error(err.Error())
	}
	record := ScrubSMS{}
	err1 := json.Unmarshal(scrubRecord, &record)
	if err1 != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + qScrub.ScrubToken + "\"}"
		return shim.Error(jsonResp)
	} else if scrubRecord == nil {
		jsonResp = "{\"Error\" : \"Scrub does not exist: " + qScrub.ScrubToken + "\"}"
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
