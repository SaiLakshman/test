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
const _BulkCreateEvent = "BULK_CREATE"

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
	Purpose           string `json:"pur"`    // Purpose where user selects while giving consent

}

//ConsentManager manages Consent related transactions
type ConsentManager struct {
}

var errorDetails, errKey, jsonResp string
var consentStatus = map[string]bool{
	"1": true,
	"2": true,
	"3": true,
	"4": true,
}

var purpose = map[string]bool{
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
		return false, "ConsentId is mandatory"
	}
	if len(s.Phone) == 0 {
		return false, "Phone Number is mandatory"
	}
	if len(s.EntityId) == 0 {
		return false, "EntityId is mandatory"
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
	if len(s.UpdatedOrg) == 0 {
		return false, "UpdatedOrg is mandatory"
	}
	if !validEnumEntry(s.Status, consentStatus) {
		return false, "Status: Enter either 1, 2, 3, 4"
	}
	if !validEnumEntry(s.CommunicationMode, communicationMode) {
		return false, "Communication Mode: Enter either 0, 1, 2, 3, 4, 5, 6"
	}
	if !validEnumEntry(s.Purpose, purpose) {
		return false, "Purpose: Enter either 1, 2, 3"
	}
	return true, ""
}

//creating Consent record in the ledger
func (s *ConsentManager) createConsent(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = string(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createConsent: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var consentToSave Consent
	err := json.Unmarshal([]byte(args[0]), &consentToSave)
	if err != nil {
		errKey = args[0]
		errorDetails = "Invalid JSON provided"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createConsent: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//Second Check if the consent id is already existing or not
	if recordBytes, _ := stub.GetState(consentToSave.ConsentId); len(recordBytes) > 0 {
		errKey = string(consentToSave.ConsentId)
		errorDetails = "Consent with this ConsentId already exist, provide unique ConsentId"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createConsent: " + jsonResp)
		return shim.Error(jsonResp)
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
	//Save the entry
	consentJSON, _ := json.Marshal(consentToSave)
	if isValid, errMsg := IsValidConsentPresent(consentToSave); !isValid {
		errKey = string(consentJSON)
		errorDetails = string(errMsg)
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createConsent: " + jsonResp)
		return shim.Error(jsonResp)
	}
	_consentLogger.Info("Saving Consent to the ledger with id----------", consentToSave.ConsentId)
	err = stub.PutState(consentToSave.ConsentId, consentJSON)
	if err != nil {
		errKey = string(consentJSON)
		errorDetails = "Unable to save consent with ConsentId - " + string(err.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createConsent: " + jsonResp)
		return shim.Error(jsonResp)
	}
	retErr := stub.SetEvent(_CreateEvent, consentJSON)
	if retErr != nil {
		errKey = string(consentJSON)
		errorDetails = "Event not generated for event : CREATE_CONSENT- " + string(retErr.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createConsent: " + jsonResp)
		return shim.Error(jsonResp)
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
		errKey = string(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("queryConsentById: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var qConsent Consent
	err := json.Unmarshal([]byte(args[0]), &qConsent)
	if err != nil {
		errKey = args[0]
		errorDetails = "Invalid JSON provided"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("queryConsentById: " + jsonResp)
		return shim.Error(jsonResp)
	}
	consentRecord, retErr := stub.GetState(qConsent.ConsentId)
	if retErr != nil {
		errKey = string(qConsent.ConsentId)
		errorDetails = "Unable to fetch the consent - " + string(retErr.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("queryConsentById: " + jsonResp)
		return shim.Error(jsonResp)
	} else if consentRecord == nil {
		errKey = string(qConsent.ConsentId)
		errorDetails = "Consent does not exist with ConsentId"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("queryConsentById: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var record Consent
	err1 := json.Unmarshal(consentRecord, &record)
	if err1 != nil {
		errKey = string(consentRecord)
		errorDetails = "Invalid JSON - " + string(err1.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("queryConsentById: " + jsonResp)
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
		errKey = string(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createBulkConsent: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var listConsent []Consent
	err := json.Unmarshal([]byte(args[0]), &listConsent)
	if err != nil {
		errKey = args[0]
		errorDetails = "Invalid JSON provided"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createBulkConsent: " + jsonResp)
		return shim.Error(jsonResp)
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
		consentJSON, _ := json.Marshal(consentToSave)
		if isValid, err := IsValidConsentPresent(consentToSave); !isValid {
			errKey = string(consentJSON)
			errorDetails = string(err)
			jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_consentLogger.Errorf("createBulkConsent: " + jsonResp)
			rejectedConsents = append(rejectedConsents, consentToSave.ConsentId)
			continue
		}
		_consentLogger.Info("Saving Consent to the ledger with id----------", consentToSave.ConsentId)
		err = stub.PutState(consentToSave.ConsentId, consentJSON)
		if err != nil {
			errKey = string(consentJSON)
			errorDetails = "Unable to save with ConsentId -" + string(err.Error())
			jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_consentLogger.Errorf("createBulkConsent: " + jsonResp)
			rejectedConsents = append(rejectedConsents, consentToSave.ConsentId)
			continue
		}
		retErr := stub.SetEvent(_BulkCreateEvent, consentJSON)
		if retErr != nil {
			errKey = string(consentJSON)
			errorDetails = "Event not generated for event : BULK_CREATE- " + string(retErr.Error())
			jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_consentLogger.Errorf("createBulkConsent: " + jsonResp)
			rejectedConsents = append(rejectedConsents, consentToSave.ConsentId)
			continue
		}
	}
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
		errKey = string(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	}
	var updatedConsent Consent
	errConsent := json.Unmarshal([]byte(args[0]), &updatedConsent)
	if errConsent != nil {
		errKey = args[0]
		errorDetails = "Invalid JSON provided"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	}
	consentRecord, err := stub.GetState(updatedConsent.ConsentId)
	if err != nil {
		errKey = string(updatedConsent.ConsentId)
		errorDetails = "Could not fetch the details for the Consent- " + string(err.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	} else if consentRecord == nil {
		errKey = string(updatedConsent.ConsentId)
		errorDetails = "Consent does not exist with ConsentId"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	}
	var existingConsent Consent
	err = json.Unmarshal([]byte(consentRecord), &existingConsent)
	if err != nil {
		errKey = string(consentRecord)
		errorDetails = "Invalid JSON for storing" + string(err.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	}
	_, creatorUpdatedBy := s.getInvokerIdentity(stub)
	existingConsent.UpdateTs = updatedConsent.UpdateTs
	existingConsent.UpdatedBy = creatorUpdatedBy
	existingConsent.Status = updatedConsent.Status

	consentJSON, _ := json.Marshal(existingConsent)
	if isValid, errMsg := IsValidConsentPresent(existingConsent); !isValid {
		errKey = string(consentJSON)
		errorDetails = string(errMsg)
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	}
	err = stub.PutState(existingConsent.ConsentId, consentJSON)
	if err != nil {
		errKey = string(consentJSON)
		errorDetails = "Unable to save consent with ConsentId -" + string(err.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	}
	retErr := stub.SetEvent(_UpdateEvent, consentJSON)
	if retErr != nil {
		errKey = string(consentJSON)
		errorDetails = "Event not generated for event : UPDATE_CONSENT-  " + string(retErr.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	}
	resultData := map[string]interface{}{
		"trxnID":  stub.GetTxID(),
		"urn":     updatedConsent.ConsentId,
		"message": "Consent status updated successfully",
		"status":  "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//updating the purpose of the consent in the ledger by providing consentId and purpose
func (s *ConsentManager) updatePurpose(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = string(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	}
	var updatedConsent Consent
	errConsent := json.Unmarshal([]byte(args[0]), &updatedConsent)
	if errConsent != nil {
		errKey = args[0]
		errorDetails = "Invalid JSON provided"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	}
	consentRecord, err := stub.GetState(updatedConsent.ConsentId)
	if err != nil {
		errKey = string(updatedConsent.ConsentId)
		errorDetails = "Could not fetch the details for the Consent- " + string(err.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	} else if consentRecord == nil {
		errKey = string(updatedConsent.ConsentId)
		errorDetails = "Consent does not exist with ConsentId"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	}
	var existingConsent Consent
	err = json.Unmarshal([]byte(consentRecord), &existingConsent)
	if err != nil {
		errKey = string(consentRecord)
		errorDetails = "Invalid JSON for storing" + string(err.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	}
	_, creatorUpdatedBy := s.getInvokerIdentity(stub)
	existingConsent.UpdateTs = updatedConsent.UpdateTs
	existingConsent.UpdatedBy = creatorUpdatedBy
	existingConsent.Purpose = updatedConsent.Purpose
	consentJSON, _ := json.Marshal(existingConsent)

	if isValid, errMsg := IsValidConsentPresent(existingConsent); !isValid {
		errKey = string(consentJSON)
		errorDetails = string(errMsg)
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	}
	err = stub.PutState(existingConsent.ConsentId, consentJSON)
	if err != nil {
		errKey = string(consentJSON)
		errorDetails = "Unable to save consent with ConsentId -" + string(err.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	}
	retErr := stub.SetEvent(_UpdateEvent, consentJSON)
	if retErr != nil {
		errKey = string(consentJSON)
		errorDetails = "Event not generated for event : UPDATE_CONSENT- " + string(retErr.Error())
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	}
	resultData := map[string]interface{}{
		"trxnID":  stub.GetTxID(),
		"urn":     updatedConsent.ConsentId,
		"message": "Consent Purpose updated successfully",
		"status":  "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//get the history of a consent from the ledger providing consentId
func (s *ConsentManager) getHistoryByConsentId(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = string(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("getHistoryByConsentId: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//	var jsonResp string
	var qConsent Consent
	err := json.Unmarshal([]byte(args[0]), &qConsent)
	if err != nil {
		errKey = args[0]
		errorDetails = "Invalid JSON provided"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("getHistoryByConsentId: " + jsonResp)
		return shim.Error(jsonResp)
	}
	historyResults, _ := getHistoryResults(stub, qConsent.ConsentId)
	return shim.Success(historyResults)
}

//function used for getting the history of a transaction
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

//getDataByPagination will query the ledger on the selector input, and display using the pagination
func (s *ConsentManager) getDataByPagination(stub shim.ChaincodeStubInterface) peer.Response {
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
		_consentLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var tempQuery Query
	err := json.Unmarshal([]byte(args[0]), &tempQuery)
	if err != nil {
		errKey = args[0]
		errorDetails = "Invalid JSON provided"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	queryString := tempQuery.SQuery
	pageSize, err1 := strconv.ParseInt(tempQuery.PageSize, 10, 32)
	if err1 != nil {
		errKey = string(pageSize)
		errorDetails = "PageSize should be a Number"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	bookMark := tempQuery.Bookmark
	paginationResults, err2 := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookMark)
	if err2 != nil {
		errKey = queryString + "," + string(pageSize) + "," + bookMark
		errorDetails = "Could not fetch the data"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("getDataByPagination: " + jsonResp)
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
