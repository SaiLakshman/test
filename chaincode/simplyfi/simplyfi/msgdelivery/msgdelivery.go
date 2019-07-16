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
		return false, "Enter either AI, VO, ID, BL, ML, QL, TA, JI or VI"
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
		return false, "Enter either AI, VO, ID, BL, ML, QL, TA, JI or VI"
	}
	return true, ""
}

//createMSGDelivery creates a MSGDelivery record in the ledger
func (s *MSGDeliveryManager) createMSGDelivery(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}
	var msgToSave MSGDelivery
	err := json.Unmarshal([]byte(args[0]), &msgToSave)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}

	//Second Check if the scrub token is existing or not
	if recordBytes, _ := stub.GetState(msgToSave.ScrubToken); len(recordBytes) > 0 {
		return shim.Error("Scrub with this scrubToken already Exist, provide unique scrubToken")
	}

	msgToSave.ObjType = "msgDelivery"
	_, creator := s.getInvokerIdentity(stub)
	msgToSave.Creator = creator
	//Save the entry
	scrubJSON, _ := json.Marshal(msgToSave)
	if isValid, errMsg := IsValid(msgToSave); !isValid {
		return shim.Error(errMsg)
	}
	_msgSMSLogger.Info("msgToSave.ScrubToken----------", msgToSave.ScrubToken)
	err = stub.PutState(msgToSave.ScrubToken, scrubJSON)

	if err != nil {
		return shim.Error("Unable to save with scrub Token " + msgToSave.ScrubToken)
	}
	retErr := stub.SetEvent(_CreateEvent, scrubJSON)

	if retErr != nil {
		_msgSMSLogger.Errorf("Event not generated for event : INITIATE_MSGDELIVERY")
		return shim.Error("{\"error\":\"Unable save scrubbing.\"}")
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
		return shim.Error("Invalid number of arguments provided for transaction")
	}
	var listScrub []MSGDelivery
	err := json.Unmarshal([]byte(args[0]), &listScrub)
	if err != nil {
		return shim.Error("Invalid json provided as input")
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
		//Save the entry
		scrubJSON, _ := json.Marshal(scrubToSave)
		if isValid, _ := IsValid(scrubToSave); !isValid {
			// _scrubSMSLogger. ("Error "+errMsg)
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		_msgSMSLogger.Info("scrubToSave.ScrubToken----------", scrubToSave.ScrubToken)
		err = stub.PutState(scrubToSave.ScrubToken, scrubJSON)
		if err != nil {
			rejectedStok = append(rejectedStok, scrubToSave.ScrubToken)
			continue
		}
		retErr := stub.SetEvent(_BulkCreateEvent, scrubJSON)
		if retErr != nil {
			_msgSMSLogger.Errorf("Event not generated for event : BULK_CREATE")
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
		return shim.Error("Invalid arguments provided")
	}
	var jsonResp string
	var qScrub MSGDelivery
	errScrub := json.Unmarshal([]byte(args[0]), &qScrub)
	if errScrub != nil {
		return shim.Error(errScrub.Error())
	}
	scrubRecord, err := stub.GetState(qScrub.ScrubToken)
	if err != nil {
		return shim.Error(err.Error())
	}
	record := MSGDelivery{}
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

//getDataByPagination will query the ledger on the selector input, and display using the pagination
func (s *MSGDeliveryManager) getDataByPagination(stub shim.ChaincodeStubInterface) peer.Response {
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
