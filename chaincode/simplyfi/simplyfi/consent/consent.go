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

var _consentLogger = shim.NewLogger("Consent")

const _CreateEvent = "CREATE_CONSENT"
const _UpdateEvent = "UPDATE_CONSENT"

//Consent structure defines the ledger record
type Consent struct {
	ObjType           string `json:"obj"`    // DocType
	ConsentId         string `json:"urn"`    // Consent Id unique -- search key
	Phone             string `json:"msisdn"` // Phone number against that consent
	ConsentTemplateId string `json:"cstid"`  // Id of registered consent template
	EntityId          string `json:"eid"`    // Content Provider
	ExpiryDate        string `json:"exdt"`   // Expiry Date of consent (forever if null)
	CLI               string `json:"cli"`    // cli
	Creator           string `json:"crtr"`   // Consent creator
	CreateTS          string `json:"cts"`    // Consent create time
	UpdatedBy         string `json:"uby"`    // Consent Updated by
	UpdateTs          string `json:"uts"`    // Consent update time
	UpdatedOrg        string `json:"uorg"`   // Consent updated by which org(service provider)
	Status            string `json:"sts"`    // Status of the consent
	CommunicationMode string `json:"cmode"`  // Source of Consent Acquisition

}

//ConsentManager manages Consent related transactions
type ConsentManager struct {
}

var consentStatus = map[string]bool{
	"1": true,
	"2": true,
	"3": true,
}

var communicationMode = map[string]bool{
	"0": true,
	"1": true,
	"2": true,
	"3": true,
	"4": true,
	"5": true,
	"6": true,
}

func validEnumEntry(input string, enumMap map[string]bool) bool {
	if _, isEntryExists := enumMap[input]; !isEntryExists {
		return false
	}
	return true
}

//IsValidConsent checks for  validity of Consent for update trxn
func IsValidConsentPresent(s Consent) (bool, string) {

	if len(s.ConsentId) == 0 {
		return false, "ConsentId should be present there"
	}
	if len(s.Phone) == 0 {
		return false, "Phone Number is mandatory"
	}
	if len(s.EntityId) == 0 {
		return false, "Entity Id is mandatory"
	}
	// if len(s.HeaderType) != 0 {
	// 	if !validEnumEntry(s.HeaderType, headerType) {
	// 		return false, "Enter either P, T, SE, SI"
	// 	}
	// }
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
	if len(s.UpdatedOrg) == 0 {
		return false, "UpdatedOrg is mandatory"
	}
	if !validEnumEntry(s.Status, consentStatus) {
		return false, "Enter either 1, 2, 3"
	}
	if !validEnumEntry(s.CommunicationMode, communicationMode) {
		return false, "Enter either 0, 1, 2, 3, 4, 5, 6"
	}
	return true, ""
}

//creating Consent record in the ledger
func (s *ConsentManager) createConsent(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}
	var consentToSave Consent
	err := json.Unmarshal([]byte(args[0]), &consentToSave)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}

	//Second Check if the consent id is already existing or not
	if recordBytes, _ := stub.GetState(consentToSave.ConsentId); len(recordBytes) > 0 {
		return shim.Error("Consent with this ConsentId already Exist, provide unique ConsentId")
	}

	consentToSave.ObjType = "Consent"
	_, creator := s.getInvokerIdentity(stub)
	consentToSave.Creator = creator
	consentToSave.UpdatedBy = creator
	consentToSave.UpdateTs = consentToSave.CreateTS
	//checking whether the key is present
	if consentToSave.ExpiryDate == "" {
		consentToSave.ExpiryDate = "null"
	}
	if consentToSave.ConsentTemplateId == "" {
		consentToSave.ConsentTemplateId = ""
	}
	//Save the entry
	consentJSON, _ := json.Marshal(consentToSave)
	if isValid, errMsg := IsValidConsentPresent(consentToSave); !isValid {
		return shim.Error(errMsg)
	}
	_consentLogger.Info("consentToSave.ConsentId----------", consentToSave.ConsentId)
	err = stub.PutState(consentToSave.ConsentId, consentJSON)
	if err != nil {
		return shim.Error("Unable to save with ConsentId " + consentToSave.ConsentId)
	}
	retErr := stub.SetEvent(_CreateEvent, consentJSON)

	if retErr != nil {
		_consentLogger.Errorf("Event not generated for event : CREATE_CONSENT")
		return shim.Error("{\"error\":\"Unable to save Consent.\"}")
	}
	resultData := map[string]interface{}{
		"trxnID":  stub.GetTxID(),
		"urn":     consentToSave.ConsentId,
		"message": "Consent Created Successfully",
		"status":  "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//querying the consent record from the ledger given consentId
func (s *ConsentManager) queryConsentById(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid arguments provided")
	}
	var jsonResp string
	var qConsent Consent
	errConsent := json.Unmarshal([]byte(args[0]), &qConsent)
	if errConsent != nil {
		return shim.Error(errConsent.Error())
	}
	consentRecord, err := stub.GetState(qConsent.ConsentId)
	if err != nil {
		return shim.Error(err.Error())
	}
	record := Consent{}
	err1 := json.Unmarshal(consentRecord, &record)
	if err1 != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + qConsent.ConsentId + "\"}"
		return shim.Error(jsonResp)
	} else if consentRecord == nil {
		jsonResp = "{\"Error\" : \"Consent does not exist: " + qConsent.ConsentId + "\"}"
		return shim.Error(jsonResp)
	}
	resultData := map[string]interface{}{
		"data":   record,
		"status": "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//InitiateBulkConsents creates a consent record in the ledger
func (s *ConsentManager) createBulkConsent(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}
	var listConsent []Consent
	err := json.Unmarshal([]byte(args[0]), &listConsent)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}
	var rejectedConsents []string
	_, creator := s.getInvokerIdentity(stub)
	for i := 0; i < len(listConsent); i++ {
		var consentToSave Consent
		consentToSave = listConsent[i]
		if recordBytes, _ := stub.GetState(consentToSave.ConsentId); len(recordBytes) > 0 {
			rejectedConsents = append(rejectedConsents, consentToSave.ConsentId)
			continue
		}
		consentToSave.ObjType = "Consent"
		consentToSave.Creator = creator
		consentToSave.UpdatedBy = creator
		consentToSave.UpdateTs = consentToSave.CreateTS
		//Save the entry
		consentJSON, _ := json.Marshal(consentToSave)
		if isValid, _ := IsValidConsentPresent(consentToSave); !isValid {
			rejectedConsents = append(rejectedConsents, consentToSave.ConsentId)
			continue
		}
		_consentLogger.Info("consentToSave.ConsentId----------", consentToSave.ConsentId)
		err = stub.PutState(consentToSave.ConsentId, consentJSON)

		if err != nil {
			rejectedConsents = append(rejectedConsents, consentToSave.ConsentId)
			continue
		}
		retErr := stub.SetEvent(_CreateEvent, consentJSON)

		if retErr != nil {
			_consentLogger.Errorf("Event not generated for event : CREATE_CONSENT")
			rejectedConsents = append(rejectedConsents, consentToSave.ConsentId)
			continue
		}
	}
	//Second Check if the scrub token is existing or not
	resultData := map[string]interface{}{
		"trxnID":     stub.GetTxID(),
		"consents_f": rejectedConsents,
		"message":    "Consents Registered Successfully",
		"status":     "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//updating the status of the consent in the ledger by providing consentId and status
func (s *ConsentManager) updateStatus(stub shim.ChaincodeStubInterface) peer.Response {

	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid arguments provided")
	}

	var updatedConsent Consent
	errConsent := json.Unmarshal([]byte(args[0]), &updatedConsent)
	if errConsent != nil {
		return shim.Error(errConsent.Error())
	}

	consentRecord, err := stub.GetState(updatedConsent.ConsentId)
	if err != nil {
		return shim.Error(err.Error())
	}
	var existingConsent Consent
	errExistingConsent := json.Unmarshal([]byte(consentRecord), &existingConsent)
	if errExistingConsent != nil {
		return shim.Error(errExistingConsent.Error())
	}
	_, creatorUpdatedBy := s.getInvokerIdentity(stub)
	existingConsent.UpdateTs = updatedConsent.UpdateTs
	existingConsent.UpdatedBy = creatorUpdatedBy
	existingConsent.Status = updatedConsent.Status

	if isValid, errMsg := IsValidConsentPresent(existingConsent); !isValid {
		return shim.Error(errMsg)
	}

	marshalConsentJSON, _ := json.Marshal(existingConsent)
	finalErr := stub.PutState(existingConsent.ConsentId, marshalConsentJSON)

	if finalErr != nil {
		return shim.Error("Unable to save Consent with id " + updatedConsent.ConsentId)
	}
	retErr := stub.SetEvent(_UpdateEvent, marshalConsentJSON)

	if retErr != nil {
		_consentLogger.Errorf("Event not generated for event : UPDATE_CONSENT")
		return shim.Error("{\"error\":\"Unable to change status of Consent.\"}")
	}
	resultData := map[string]interface{}{
		"trxnID":  stub.GetTxID(),
		"urn":     updatedConsent.ConsentId,
		"message": "Header status updated successfully",
		"status":  "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//get the history of a consent from the ledger providing consentId
func (s *ConsentManager) getHistoryByConsentId(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid arguments provided")
	}
	//	var jsonResp string
	var qConsent Consent
	errConsent := json.Unmarshal([]byte(args[0]), &qConsent)
	if errConsent != nil {
		return shim.Error(errConsent.Error())
	}
	historyResults, _ := getHistoryResults(stub, qConsent.ConsentId)
	return shim.Success(historyResults)
}

func getHistoryResults(stub shim.ChaincodeStubInterface, consentId string) ([]byte, error) {

	resultsIterator, err := stub.GetHistoryForKey(consentId)
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

func (s *ConsentManager) getByAnyKeyWithPagination(stub shim.ChaincodeStubInterface) peer.Response {
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
	errConsent := json.Unmarshal([]byte(args[0]), &tempQuery)
	if errConsent != nil {
		return shim.Error(errConsent.Error())
	}
	queryString := tempQuery.SQuery
	pageSize, _ := strconv.ParseInt(tempQuery.PageSize, 10, 32)
	bookMark := tempQuery.Bookmark
	paginationResults, _ := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookMark)
	fmt.Println(paginationResults)
	return shim.Success([]byte(paginationResults))
}
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
func (s *ConsentManager) getInvokerIdentity(stub shim.ChaincodeStubInterface) (bool, string) {
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
