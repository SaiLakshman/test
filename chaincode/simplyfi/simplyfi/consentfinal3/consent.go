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

var errorDetails, errKey, jsonResp, repError string
var consentStatus = map[string]bool{
	"1": true, //ConsentRaised
	"2": true, //Approved
	"3": true, //Revoked
	"4": true, //Churned
}

var purpose = map[string]bool{
	"1": true, //Both
	"2": true, //Promotional
	"3": true, //Service
}

var communicationMode = map[string]bool{
	"0": true, //Migration
	"1": true, //WEB
	"2": true, //SMS
	"3": true, //IVR
	"4": true, //USSD
	"5": true, //APP
	"6": true, //Customer Support
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
		return false, "Purpose: Enter either 1(Both), 2(Promotional), 3(Service)"
	}
	return true, ""
}

//creating Consent record in the ledger
func (s *ConsentManager) createConsent(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	var consentToSave Consent
	//checking the length of the input
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createConsent: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the input to consentToSave Object
	err := json.Unmarshal([]byte(args[0]), &consentToSave)
	if err != nil {
		repError= strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON provided: "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createConsent: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//check whether consent already exists or not with consentId as input
	if recordBytes, _ := stub.GetState(consentToSave.ConsentId); len(recordBytes) > 0 {
		errKey = consentToSave.ConsentId
		errorDetails = "Consent with this ConsentId already exist, provide unique ConsentId"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createConsent: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the consent with the details provided as input
	consentToSave.ObjType = "Consent"
	_, creator := s.getInvokerIdentity(stub)
	consentToSave.Creator = creator
	consentToSave.UpdatedBy = creator
	consentToSave.UpdateTs = consentToSave.CreateTS
	if consentToSave.ExpiryDate == "" {
		consentToSave.ExpiryDate = "null"
	}
	//marshalling the data to store into the ledger
	consentJSON, marshalErr:= json.Marshal(consentToSave)
	if marshalErr != nil {
		repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createConsent: " + jsonResp)
		return shim.Error(jsonResp)
	}
	// checking the validity og the data before storing into the ledger
	if isValid, errMsg := IsValidConsentPresent(consentToSave); !isValid {
		errKey = string(consentJSON)
		errorDetails = errMsg
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createConsent: " + jsonResp)
		return shim.Error(jsonResp)
	}
	_consentLogger.Info("Saving Consent to the ledger with id----------", consentToSave.ConsentId)
	//storing the consent data to the ledger
	err = stub.PutState(consentToSave.ConsentId, consentJSON)
	if err != nil {
		errKey = string(consentJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save consent with ConsentId- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createConsent: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//setting an event after storing consent in the leger
	retErr := stub.SetEvent(_CreateEvent, consentJSON)
	if retErr != nil {
		errKey = string(consentJSON)
		repError = strings.Replace(retErr.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : CREATE_CONSENT- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createConsent: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the response and returning to the application layer
	resultData := map[string]interface{}{
		"trxnID":  stub.GetTxID(),
		"urn":     consentToSave.ConsentId,
		"message": "Consent Created Successfully",
		"status":  "true",
		"consent": consentToSave,
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//querying the consent record from the ledger given consentId
func (s *ConsentManager) queryConsentById(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	var qConsent Consent
	//checking the length of the input
	if len(args) != 1 {
		errKey = string.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("queryConsentById: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the data to qConsent object
	err := json.Unmarshal([]byte(args[0]), &qConsent)
	if err != nil {
		repError= strings.Replace(err.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("queryConsentById: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//getting the state of the consent from the ledger with consentId
	consentRecord, retErr := stub.GetState(qConsent.ConsentId)
	if retErr != nil {
		errKey = qConsent.ConsentId
		repError = strings.Replace(retErr.Error(), "\"", " ", -1)
		errorDetails = "Unable to fetch the consent- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("queryConsentById: " + jsonResp)
		return shim.Error(jsonResp)
	} else if consentRecord == nil {
		errKey = qConsent.ConsentId
		errorDetails = "Consent does not exist with ConsentId"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("queryConsentById: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the data retrieved from ledger to the response
	var record Consent
	err1 := json.Unmarshal(consentRecord, &record)
	if err1 != nil {
		repError = strings.Replace(err1.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("queryConsentById: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the result and returning to the application layer
	resultData := map[string]interface{}{
		"data":   record,
		"status": "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//createBulkConsent creates a consents in the ledger
func (s *ConsentManager) createBulkConsent(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	var listConsent []Consent
	var rejectedConsents []string
	//checking the length of the consents
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createBulkConsent: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the data to the listConsent object
	err := json.Unmarshal([]byte(args[0]), &listConsent)
	if err != nil {
		repError= strings.Replace(err.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("createBulkConsent: " + jsonResp)
		return shim.Error(jsonResp)
	}
	_, creator := s.getInvokerIdentity(stub)
	//iterating through the list of consents provided in the input
	for i := 0; i < len(listConsent); i++ {
		var consentToSave Consent
		consentToSave = listConsent[i]
		//checking whether the consent already exist or not with consentId
		if recordBytes, _ := stub.GetState(consentToSave.ConsentId); len(recordBytes) > 0 {
			errKey = listConsent[i].ConsentId
			errorDetails = "Consent with ConsentId already Exists, Provide unique ConsentId "
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_consentLogger.Errorf("createBulkConsent: " + jsonResp)
			rejectedConsents = append(rejectedConsents, consentToSave.ConsentId)
			continue
		}
		//packaging the consent with the details provided as input
		consentToSave.ObjType = "Consent"
		consentToSave.Creator = creator
		consentToSave.UpdatedBy = creator
		consentToSave.UpdateTs = consentToSave.CreateTS
		//marshalling the data to store into the ledger
		consentJSON, marshalErr := json.Marshal(consentToSave)
		if marshalErr != nil {
			repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
			errorDetails = "Cannot Marshal the JSON- " + repError
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_consentLogger.Errorf("createBulkConsent: " + jsonResp)
			rejectedConsents= append(rejectedConsents, consentToSave.ConsentId)
			continue
		}
		//checking the validity of the consent before storing into the ledger 
		if isValid, err := IsValidConsentPresent(consentToSave); !isValid {
			errKey = string(consentJSON)
			errorDetails = err
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_consentLogger.Errorf("createBulkConsent: " + jsonResp)
			rejectedConsents = append(rejectedConsents, consentToSave.ConsentId)
			continue
		}
		_consentLogger.Info("Saving Consent to the ledger with id----------", consentToSave.ConsentId)
		//storing the consent data into the ledger
		err = stub.PutState(consentToSave.ConsentId, consentJSON)
		if err != nil {
			errKey = string(consentJSON)
			repError = strings.Replace(err.Error(), "\"", " ", -1)
			errorDetails = "Unable to save with ConsentId- " + repError
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_consentLogger.Errorf("createBulkConsent: " + jsonResp)
			rejectedConsents = append(rejectedConsents, consentToSave.ConsentId)
			continue
		}
		//setting an event after storing into the ledger
		retErr := stub.SetEvent(_BulkCreateEvent, consentJSON)
		if retErr != nil {
			errKey = string(consentJSON)
			repError = strings.Replace(retErr.Error(), "\"", " ", -1)
			errorDetails = "Event not generated for event : BULK_CREATE- " + repError
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_consentLogger.Errorf("createBulkConsent: " + jsonResp)
			rejectedConsents = append(rejectedConsents, consentToSave.ConsentId)
			continue
		}
	}
	//packaging the response and returning to the application layer
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
	var updatedConsent Consent
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the input to updateConsent object
	errConsent := json.Unmarshal([]byte(args[0]), &updatedConsent)
	if errConsent != nil {
		repError= strings.Replace(errConsent.Error(),"\"", " ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	}
	//getting the state of the consent to perform update using ConsentId
	consentRecord, err := stub.GetState(updatedConsent.ConsentId)
	if err != nil {
		errKey = updatedConsent.ConsentId
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Could not fetch the details for the Consent- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	} else if consentRecord == nil {
		errKey = updatedConsent.ConsentId
		errorDetails = "Consent does not exist with ConsentId"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	}
	var existingConsent Consent
	//unmarshalling the retrieved consent data 
	err = json.Unmarshal([]byte(consentRecord), &existingConsent)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON for storing" + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the consent with the details provided as input
	_, creatorUpdatedBy := s.getInvokerIdentity(stub)
	existingConsent.UpdateTs = updatedConsent.UpdateTs
	existingConsent.UpdatedBy = creatorUpdatedBy
	existingConsent.Status = updatedConsent.Status
	//marshalling the data to store into the ledger
	consentJSON, marshalErr := json.Marshal(existingConsent)
	if marshalErr != nil {
		repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	}
	//checking the validity of the consent before storing into the blockchain
	if isValid, errMsg := IsValidConsentPresent(existingConsent); !isValid {
		errKey = string(consentJSON)
		errorDetails = errMsg
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	}
	//storing the data into the ledger
	err = stub.PutState(existingConsent.ConsentId, consentJSON)
	if err != nil {
		errKey = string(consentJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save consent with ConsentId- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	}
	//setting an event after storing into the ledger
	retErr := stub.SetEvent(_UpdateEvent, consentJSON)
	if retErr != nil {
		errKey = string(consentJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : UPDATE_CONSENT-  " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(status): " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the response and returning to the application layer
	resultData := map[string]interface{}{
		"trxnID":  stub.GetTxID(),
		"urn":     updatedConsent.ConsentId,
		"message": "Consent status updated successfully",
		"status":  "true",
		"consent": updatedConsent
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//updating the purpose of the consent in the ledger by providing consentId and purpose
func (s *ConsentManager) updatePurpose(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	var updatedConsent Consent
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the input to updatedConsent object
	errConsent := json.Unmarshal([]byte(args[0]), &updatedConsent)
	if errConsent != nil {
		repError= strings.Replace(errConsent.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	}
	//getting the state of the consent from the ledger to update with consentId
	consentRecord, err := stub.GetState(updatedConsent.ConsentId)
	if err != nil {
		errKey = updatedConsent.ConsentId
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Could not fetch the details for the Consent- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	} else if consentRecord == nil {
		errKey = updatedConsent.ConsentId
		errorDetails = "Consent does not exist with ConsentId"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	}
	var existingConsent Consent
	//unmarshalling the retrieved data to existingConsent object
	err = json.Unmarshal([]byte(consentRecord), &existingConsent)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON for storing" + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the consent with the details provided as input
	_, creatorUpdatedBy := s.getInvokerIdentity(stub)
	existingConsent.UpdateTs = updatedConsent.UpdateTs
	existingConsent.UpdatedBy = creatorUpdatedBy
	existingConsent.Purpose = updatedConsent.Purpose
	//marshalling the data to store into the ledger
	consentJSON, marshalErr:= json.Marshal(existingConsent) 
	if marshalErr != nil {
		repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	}
	//checking the validity of the consent before storing into the ledger
	if isValid, errMsg := IsValidConsentPresent(existingConsent); !isValid {
		errKey = string(consentJSON)
		errorDetails = errMsg
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	}
	//storing the data into the ledger
	err = stub.PutState(existingConsent.ConsentId, consentJSON)
	if err != nil {
		errKey = string(consentJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save consent with ConsentId -" + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	}
	//setting an event after storing into the ledger
	retErr := stub.SetEvent(_UpdateEvent, consentJSON)
	if retErr != nil {
		errKey = string(consentJSON)
		repError = strings.Replace(retErr.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : UPDATE_CONSENT- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("updateConsent(purpose): " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the response and returning to the application layer
	resultData := map[string]interface{}{
		"trxnID":  stub.GetTxID(),
		"urn":     updatedConsent.ConsentId,
		"message": "Consent Purpose updated successfully",
		"status":  "true",
		"consent": existingConsent,
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//get the history of a consent from the ledger providing consentId
func (s *ConsentManager) getHistoryByConsentId(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	//checking the length of the input
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("getHistoryByConsentId: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var qConsent Consent
	//unmarshalling the input to qConsent object
	err := json.Unmarshal([]byte(args[0]), &qConsent)
	if err != nil {
		repError= strings.Replace(err.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
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
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var tempQuery Query
	err := json.Unmarshal([]byte(args[0]), &tempQuery)
	if err != nil {
		repError= strings.Replace(err.Error(),"\"", " ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	queryString := tempQuery.SQuery
	pageSize, err1 := strconv.ParseInt(tempQuery.PageSize, 10, 32)
	if err1 != nil {
		errKey = string(pageSize)
		errorDetails = "PageSize should be a Number"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_consentLogger.Errorf("getDataByPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	bookMark := tempQuery.Bookmark
	paginationResults, err2 := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookMark)
	if err2 != nil {
		errKey = queryString + "," + string(pageSize) + "," + bookMark
		errorDetails = "Could not fetch the data"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
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
