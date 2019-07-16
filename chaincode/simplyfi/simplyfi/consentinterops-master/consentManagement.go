package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	id "github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
	"github.com/hyperledger/fabric/protos/peer"
	pb "github.com/hyperledger/fabric/protos/peer"
)

//Consentdetails consent details
type Consentdetails struct {
	ObjectType        string `json:"obj"`
	ConsentID         string `json:"urn"`
	Msisdn            string `json:"msisdn"`
	ConsentTemplateID string `json:"cstid"`
	EntityID          string `json:"eid"`
	ExpiryDate        string `json:"exdt"`
	Cli               string `json:"cli"`
	Status            string `json:"sts"`
	Creator           string `json:"crtr"`
	UpdateTs          string `json:"uts"`
	CreateTs          string `json:"cts"`
	UpdatedBy         string `json:"uby"`
	UpdatedOrg        string `json:"uorg"`
	CommunicationMode string `json:"cmode"`
}

//ConsentStatus is structure of ConsentStatus
type ConsentStatus struct {
	ConsentID string `json:"urn"`
	Status    string `json:"sts"`
	UpdateTs  string `json:"uts"`
}

//EventPayLoad is strcuture for EventPayload
type EventPayLoad struct {
	consent Consentdetails
	txnID   string
}

const _CreateEvent = "CREATE_CONSENT"
const _UpdateEvent = "UPDATE_CONSENT"

//Object type for Consent - do not change
const _ObjectType = "Consent"

//below are the error message format
const _Format0 = "Invalid json provided as input."
const _Format1 = "Invalid number of arguments provided for transaction."
const _Format2 = "Unable to save consent with id:"
const _Format3 = "Unable to modify the consent. :"
const _Format4 = "Consent ID already registered for ConsentId:"
const _Format5 = "Event not generated for event :"

var _consentLogger = shim.NewLogger("ConsentManagementSmartContract")

//ConsentManager is ConsentManager
type ConsentManager struct {
}

//isValidConsent check the validity of the consent
func isValidConsent(c Consentdetails) (bool, string) {
	if len(c.ConsentID) == 0 {
		return false, "ConsentId is mandatory"
	}
	if len(c.ConsentTemplateID) == 0 {
		return false, "Consent TemplateId is mandatory"
	}
	if len(c.EntityID) == 0 {
		return false, "EntityId is mandatory"
	}
	if len(c.Msisdn) == 0 {
		return false, "MSISDN is mandatory"
	}
	if len(c.Cli) == 0 {
		return false, "Cli/Header is mandatory"
	}
	if len(c.UpdatedOrg) == 0 {
		return false, "UpdatedOrg is mandatory"
	}
	if !validEnumEntry(c.Status, consentStatus) {
		return false, "Status can be either (1)Consent Raised, (2)Approved, (3)Revoked) or (4)PD/Churned"
	}
	if !validEnumEntry(c.Status, commMode) {
		return false, "Communication mode can be either (0)Migration, (1)WEB, (2)SMS, (3)IVR, (4)USSD, (5)APP or, (6)Customer Support"
	}
	return true, ""
}

func isValidStatus(status string) (bool, string) {
	if !validEnumEntry(status, consentStatus) {
		return false, "Status can be either (1)Consent Raised, (2)Approved, (3)Revoked) or (4)PD/Churned"
	}
	return true, ""
}

func isValidConsentToModify(c Consentdetails) (bool, string) {
	if len(c.UpdateTs) == 0 {
		return false, "UpdateTs is mandatory"
	}
	return true, ""
}

func isValidConsentToRecord(c Consentdetails) (bool, string) {
	if len(c.CreateTs) == 0 {
		return false, "CreateTs is mandatory"
	}
	if len(c.UpdateTs) == 0 {
		return false, "UpdateTs is mandatory"
	}
	return true, ""
}

//isValidDate validates the format of date, only when there is date given
func isValidDate(date string) (bool, string) {
	// example of epoch time 1551788124
	if date != "" {
		_, err := strconv.ParseInt(date, 10, 64)
		if err != nil {
			return false, "Date requires proper format e.g. '1551788124'"
		}
	}
	return true, ""
}

func isValidMsisdn(Msisdn string) (bool, string) {
	regex := regexp.MustCompile("^[0-9]{10,13}$")
	if !regex.MatchString(Msisdn) {
		return false, "Invalid MSISDN. It will be numeric number between 10 to 13 without Country Code."
	}
	return true, ""
}

var commMode = map[string]bool{
	"0": true, //Migration
	"1": true, //WEB
	"2": true, //SMS
	"3": true, //IVR
	"4": true, //USSD
	"5": true, //APP
	"6": true, //Customer Support
}

var consentStatus = map[string]bool{
	"1": true, //ConsentRaised
	"2": true, //Approved
	"3": true, //Revoked
	"4": true, //Churned
}

func validEnumEntry(input string, enumMap map[string]bool) bool {
	if _, isEntryExists := enumMap[input]; !isEntryExists {
		return false
	}
	return true
}

//RecordConsent saves the consent in DLT with the given inputTelcoCommon
//Takes an array of consents from args[0] to be updated ( eg. []Consentdetails )
//Returns an array of map, each element of that array contains the below information:
//"trxnID"		:  	transaction Id,
//"consentID"	:  	URN,
//"message"		:  	"Consent Creation Successful"
//"consent" 	: 	Saved Consent object
func (cm *ConsentManager) RecordConsent(stub shim.ChaincodeStubInterface) pb.Response {

	_, args := stub.GetFunctionAndParameters()

	if len(args) < 1 {
		_consentLogger.Errorf("Invalid number of arguments provided for transaction.")
		//return shim.Error("{\"error\":\"Invalid number of arguments provided for transaction\"}")
		return shim.Error(getErrorMsg(_Format1))

	}

	var consents []Consentdetails

	err := json.Unmarshal([]byte(args[0]), &consents)
	if err != nil {
		_consentLogger.Errorf(_Format0)
		return shim.Error(getErrorMsg(_Format0))
		//return shim.Error("{\"error\":\"Invalid json provided as input\"}")

	}

	result := make([]map[string]interface{}, 0)

	_, creater := cm.getInvokerIdentity(stub)

	for _, eachConsent := range consents {

		if recordBytes, _ := stub.GetState(eachConsent.ConsentID); len(recordBytes) > 0 {

			//return shim.Error("{\"error\":\"Consent ID already registered for ConsentId:" + eachConsent.ConsentID + "\"}")
			_consentLogger.Infof(_Format4)
			return shim.Error(getErrorMsg(_Format4, eachConsent.ConsentID))
		}

		consents := cm.getConsentsByMsisdnCli(stub, eachConsent.Msisdn, eachConsent.Cli)

		var flag bool
		if len(consents) > 0 {
			_consentLogger.Infof("More than one record found for given MSISDN + CLI")

			for _, cons := range consents {
				if cons.Status == "1" {
					flag = true
				}
			}
			if flag {
				_consentLogger.Infof(_Format4)
				return shim.Error(getErrorMsg(_Format4, eachConsent.ConsentID, " msisdn > "+eachConsent.Msisdn, " header > "+eachConsent.Cli))
			}

			if !flag {
				_consentLogger.Infof("Consent(s) with status either 2 / 3 / 4 is already there in Ledger")
			}
		}

		//validation
		if isValid, errMsg := isValidConsent(eachConsent); !isValid {
			//_consentLogger.Infof("Unable to save consent with id:"+eachConsent.ConsentID, errMsg)
			_consentLogger.Infof(_Format2)
			return shim.Error(getErrorMsg(_Format2, eachConsent.ConsentID, ". Error :", errMsg))
			//return shim.Error("{\"error\":\"Unable to save consent with id :" + eachConsent.ConsentID + ". Error :" + errMsg + ".\"}")
		}

		if isValid, errMsg := isValidConsentToRecord(eachConsent); !isValid {
			_consentLogger.Infof(_Format2, ".Error :", eachConsent.ConsentID, errMsg)
			//_consentLogger.Infof("Unable to save consent with id:"+eachConsent.ConsentID, errMsg)
			//return shim.Error("{\"error\":\"Unable to save consent with id :" + eachConsent.ConsentID + ". Error :" + errMsg + ".\"}")

			return shim.Error(getErrorMsg(_Format2, eachConsent.ConsentID, ". Error :", errMsg))
		}

		//Update each Consent Object
		eachConsent.ObjectType = _ObjectType

		eachConsent.Creator = creater
		eachConsent.UpdatedBy = creater

		consentJSON, _ := json.Marshal(eachConsent)

		_consentLogger.Info("Consent to Save :", eachConsent.ConsentID)

		err = stub.PutState(eachConsent.ConsentID, consentJSON)
		if err != nil {
			_consentLogger.Errorf(_Format2, eachConsent.ConsentID)
			//_consentLogger.Errorf("Unable to save with consent id" + eachConsent.ConsentID)
			//return shim.Error("{\"error\":\"Unable to save with consent id" + eachConsent.ConsentID + "\"}")
			return shim.Error(getErrorMsg(_Format2, eachConsent.ConsentID))
		}

		p := EventPayLoad{consent: eachConsent, txnID: stub.GetTxID()}

		payloadbytes, _ := json.Marshal(p)
   
		retErr := stub.SetEvent(_CreateEvent, payloadbytes)

		if retErr != nil {

			_consentLogger.Errorf(_Format5, _CreateEvent)
		}

		resultData := map[string]interface{}{
			"trxnID":    stub.GetTxID(),
			"consentID": eachConsent.ConsentID,
			"message":   "Consent Creation Successful",
			"consent":   eachConsent,
		}

		result = append(result, resultData)

	}
	respJSON, _ := json.Marshal(result)
	return shim.Success(respJSON)
}

//GetConsent gets the consent for the given subscribers
//Takes json of either MSISDN, ConsentID, EntityID, HeaderID in spcific format in args[0]
//Returns an array of consent objects for the given input and seacrch criteria
func (cm *ConsentManager) GetConsent(stub shim.ChaincodeStubInterface) pb.Response {
	var response peer.Response
	searchCriteria := make(map[string]string)
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		_consentLogger.Errorf("Invalid number of arguments provided for transaction.")
		return shim.Error("{\"error\":\"Invalid number of arguments provided for transaction\"}")
	}
	err := json.Unmarshal([]byte(args[0]), &searchCriteria)
	if err != nil {
		_consentLogger.Errorf("Invalid json provided as input.")
		return shim.Error("{\"error\":\"Invalid json provided as input\"}")
	}
	searchType, isOk := searchCriteria["type"]
	if !isOk {
		_consentLogger.Errorf("Search type not provided.")
		return shim.Error("{\"error\":\"Search type not provided\"}")
	}

	switch searchType {
	case "msisdn":
		consentSearchCriteria := `{
			"obj":"Consent"	,
			"msisdn":"%s"
		}`
		msisdn := searchCriteria[searchType]
		consents := cm.retrieveConsentRecords(stub, fmt.Sprintf(consentSearchCriteria, msisdn), "consentSearchByMsisdn")
		recordsJSON, _ := json.Marshal(consents)
		response = shim.Success(recordsJSON)

	case "urn":
		consentSearchCriteria := `{
				"obj":"Consent"	,
				"urn":"%s"
			}`
		urn := searchCriteria[searchType]
		consents := cm.retrieveConsentRecords(stub, fmt.Sprintf(consentSearchCriteria, urn), "consentSearchByUrn")
		recordsJSON, _ := json.Marshal(consents)
		response = shim.Success(recordsJSON)

	case "entity":
		consentSearchCriteria := `{
				"obj":"Consent"	,
				"eid":"%s"
			}`
		entity := searchCriteria[searchType]
		consents := cm.retrieveConsentRecords(stub, fmt.Sprintf(consentSearchCriteria, entity), "consentSearchByEntity")
		recordsJSON, _ := json.Marshal(consents)
		response = shim.Success(recordsJSON)

	case "header":
		consentSearchCriteria := `{
				"obj":"Consent"	,
				"cli":"%s"
			}`
		cli := searchCriteria[searchType]
		consents := cm.retrieveConsentRecords(stub, fmt.Sprintf(consentSearchCriteria, cli), "consentSearchByCli")
		recordsJSON, _ := json.Marshal(consents)
		response = shim.Success(recordsJSON)

	default:
		_consentLogger.Errorf("Unsupported search type provided." + searchType)
		return shim.Error("{\"error\":\"Unsupported search type provided " + searchType + "provided\"}")
	}
	return response
}

//UpdateConsentStatus updates only the status of a single Consent for the given consentId
//args[0] consentId(URN) for searching consent
//args[1] new status to be updated
//args[2] new updatedTs to be updated
//Returns an array of map, each element of that array contains the below information:
//"trxnID"		:   transaction Id,
//"consentID"	: 	updatedStatusConsent.ConsentId/URN,
//"message"		:   "Update Successful",
//"consent"		:   Updated consent element,
func (cm *ConsentManager) UpdateConsentStatus(stub shim.ChaincodeStubInterface) pb.Response {
	_consentLogger.Info("Within UpdateConsentStatus")
	_, args := stub.GetFunctionAndParameters()

	if len(args) < 3 {
		return shim.Error("Invalid No of arguments provided")
	}

	searchConsentID := args[0]
	newStatus := args[1]
	newUpdatedTS := args[2]

	if isValid, errMsg := checkValidityForStatus(searchConsentID, newStatus, newUpdatedTS); !isValid {
		_consentLogger.Infof("Unable to modify the status.Provide the proper data :", errMsg)
		return shim.Error("{\"error\":\"Unable to modify the consent." + errMsg + "\"}")
	}

	if isValid, errMsg := isValidStatus(newStatus); !isValid {
		_consentLogger.Infof("Invalid Status to update the consent.", errMsg)
		return shim.Error("{\"error\":\"Invalid Status to update the consent." + errMsg + "\"}")
	}

	existingRec, err1 := stub.GetState(searchConsentID)

	if len(existingRec) == 0 {
		_consentLogger.Errorf("Consent does not exist with id ", searchConsentID)
		return shim.Error("{\"error\":\"Consent does not exist with id " + searchConsentID + ".\"}")
	}

	if err1 != nil {
		_consentLogger.Errorf("Error while fetching Consent with id ", searchConsentID)
		return shim.Error("{\"error\":\"Error while fetching Consent with id " + searchConsentID + ".\"}")
	}

	var updatedStatusConsent Consentdetails
	err := json.Unmarshal(existingRec, &updatedStatusConsent)
	if err != nil {
		return shim.Error("{\"error\":\"Error while unmarshaling the data.\"}")
	}

	updatedStatusConsent.Status = newStatus
	updatedStatusConsent.UpdateTs = newUpdatedTS

	_, updatedBy := cm.getInvokerIdentity(stub)
	updatedStatusConsent.UpdatedBy = updatedBy

	marshalConsentJSON, _ := json.Marshal(updatedStatusConsent)

	finalErr := stub.PutState(updatedStatusConsent.ConsentID, marshalConsentJSON)
	if finalErr != nil {
		_consentLogger.Errorf("Unable to save with consent id " + updatedStatusConsent.ConsentID)
		return shim.Error("{\"error\":\"Unable to save with consent id " + updatedStatusConsent.ConsentID + ".\"}")
	}

	p := EventPayLoad{consent: updatedStatusConsent, txnID: stub.GetTxID()}

	payloadbytes, _ := json.Marshal(p)

	retErr := stub.SetEvent(_UpdateEvent, payloadbytes)

	if retErr != nil {
		_consentLogger.Errorf("Event not generated for event : MODIFY_CONSENT")
	}

	resultData := map[string]interface{}{
		"trxnID":    stub.GetTxID(),
		"consentID": updatedStatusConsent.ConsentID,
		"message":   "Update Successful",
		"consent":   updatedStatusConsent,
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)

}

//UpdateConsentStatusByIDs changes the status of the Consents based upon the given ConsentIds as passed in the args[o]
//args[0] string array of URNs(consentIDs) for which the status need to be changed to Revoked (3)
//args[1] new status to be updated with
//args[2] updateTs is the time to update the consent records
//Returns an array of map, each element of that array contains the below information:
//"trxnID"		:  transaction Id
//"consentID"	:  Consent Id for which data has been modified
//"status"		:  "Update Consent Status Successful"
//"consent"		:  Updated consent element
func (cm *ConsentManager) UpdateConsentStatusByIDs(stub shim.ChaincodeStubInterface) pb.Response {
	_consentLogger.Info("Within RevokeConsentInBulkByIDs")
	_, args := stub.GetFunctionAndParameters()

	if len(args) == 0 {
		return shim.Error("Invalid No of arguments provided")
	}

	arr := make([]string, 0)
	err := json.Unmarshal([]byte(args[0]), &arr)
	if err != nil {
		_consentLogger.Errorf("Invalid json provided as input.")
		return shim.Error("{\"error\":\"Invalid json provided as input\"}")
	}

	newStatus := args[1]
	if isValid, errMsg := isValidStatus(newStatus); !isValid {
		_consentLogger.Infof("Unable to modify the consent. :", errMsg)
		return shim.Error("{\"error\":\"Unable to modify the consent." + errMsg + "\"}")
	}

	newUpdatedTS := args[2]

	returnStatus := make([]map[string]interface{}, 0)

	_, updatedBy := cm.getInvokerIdentity(stub)

	for _, searchConsentID := range arr {

		existingRec, err1 := stub.GetState(searchConsentID)

		if len(existingRec) == 0 {
			_consentLogger.Errorf("Consent does not exist with id ", searchConsentID)
			return shim.Error("{\"error\":\"Consent does not exist with id " + searchConsentID + ".\"}")
		}

		if err1 != nil {
			_consentLogger.Errorf("Error while fetching Consent with id ", searchConsentID)
			return shim.Error("{\"error\":\"Error while fetching Consent with id " + searchConsentID + ".\"}")
		}

		var updatedStatusConsent Consentdetails
		err := json.Unmarshal(existingRec, &updatedStatusConsent)
		if err != nil {
			return shim.Error("{\"error\":\"Error while unmarshaling the data.\"}")
		}

		updatedStatusConsent.Status = newStatus
		updatedStatusConsent.UpdateTs = newUpdatedTS
		updatedStatusConsent.UpdatedBy = updatedBy

		marshalConsentJSON, _ := json.Marshal(updatedStatusConsent)

		finalErr := stub.PutState(updatedStatusConsent.ConsentID, marshalConsentJSON)

		if finalErr != nil {
			_consentLogger.Errorf("Unable to save with consent id " + updatedStatusConsent.ConsentID)
			return shim.Error("{\"error\":\"Unable to save with consent id " + updatedStatusConsent.ConsentID + ".\"}")
		}

		p := EventPayLoad{consent: updatedStatusConsent, txnID: stub.GetTxID()}

		payloadbytes, _ := json.Marshal(p)

		retErr := stub.SetEvent(_UpdateEvent, payloadbytes)

		if retErr != nil {
			_consentLogger.Errorf("Event not generated for event : MODIFY_CONSENT")
		}

		eachReturnData := map[string]interface{}{
			"trxnID":    stub.GetTxID(),
			"consentID": updatedStatusConsent.ConsentID,
			"status":    "Update Successful",
			"consent":   updatedStatusConsent,
		}

		returnStatus = append(returnStatus, eachReturnData)
	}

	respJSON, _ := json.Marshal(returnStatus)

	return shim.Success(respJSON)

}

//UpdateConsentStatusByHeader updates the consent status to either 'Approved' (2), 'Revoked'(3) or any other valid status
//Method can be used to for bulk update of records
//args[0] should be an array of JSON objects of below structure (Keys are important to follow)
//"msisdn" : msisdn for which record should be fetched
//"cli"	 : cli ; combination of cli and msisdn should return a unique record from ledger
//args[1] : should be the status of the updated record, can be either 2, 3 or any valid one
//args[2] : UpdateTimestamp
//Returns an array of map, each element of that array contains the below information:
//"trxnID"		:  Transaction Id
//"consentID"	:  ConsentId
//"message"		:  "Update Successful"
//"consent"		:  Updated consent element
func (cm *ConsentManager) UpdateConsentStatusByHeader(stub shim.ChaincodeStubInterface) pb.Response {
	_consentLogger.Info("Within UpdateConsentStatusByHeader")
	searchCriteriaArr := make([]map[string]string, 0)
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		_consentLogger.Errorf("Invalid number of arguments provided for transaction.")
		return shim.Error("{\"error\":\"Invalid number of arguments provided for transaction\"}")
	}
	err := json.Unmarshal([]byte(args[0]), &searchCriteriaArr)

	if err != nil {
		_consentLogger.Errorf("Invalid json provided as input.")
		return shim.Error("{\"error\":\"Invalid json provided as input\"}")
	}

	// stsT : target status for all the element for the query result
	stsT := args[1] // can be either 2,3, or anyther state

	if isValid, errMsg := isValidStatus(stsT); !isValid {
		_consentLogger.Infof("Invalid Status to update the consent. :", errMsg)
		return shim.Error("{\"error\":\"Invalid Status to update the consent." + errMsg + "\"}")
	}
	newUpdatedTS := args[2]

	returnResult := make([]map[string]interface{}, 0)

	for _, searchCriteria := range searchCriteriaArr {

		consentSearchCriteria := `{
			"obj":"Consent"	,
			"msisdn":"%s",
			"cli":"%s"	
		}`

		msisdn := searchCriteria["msisdn"]
		cli := searchCriteria["cli"]

		consents := cm.retrieveConsentRecords(stub, fmt.Sprintf(consentSearchCriteria, msisdn, cli), "consentSearchByHeaderCliSts")

		if len(consents) == 0 {
			_consentLogger.Infof("No consent record found ")
			return shim.Error("{\"error\":\"No consent record found \"}")
		}

		if len(consents) > 1 {
			_consentLogger.Infof("More than one consent record found ")
			return shim.Error("{\"error\":\"More than one consent record found \"}")
		}

		singleConsent := consents[0]
		singleConsent.Status = stsT
		singleConsent.UpdateTs = newUpdatedTS

		_, updatedBy := cm.getInvokerIdentity(stub)
		singleConsent.UpdatedBy = updatedBy

		marshalConsentJSON, _ := json.Marshal(singleConsent)

		finalErr := stub.PutState(singleConsent.ConsentID, marshalConsentJSON)

		if finalErr != nil {
			_consentLogger.Infof("Unable to save with consent : %v", finalErr)
			return shim.Error("{\"error\":\"Unable to save with consent id " + singleConsent.ConsentID + "\"}")
		}

		p := EventPayLoad{consent: singleConsent, txnID: stub.GetTxID()}

		payloadbytes, _ := json.Marshal(p)

		retErr := stub.SetEvent(_UpdateEvent, payloadbytes)

		if retErr != nil {
			_consentLogger.Errorf("Event not generated for event : MODIFY_CONSENT")
		}

		//make the payload to return and pass it through the shim.success
		returnStatus := map[string]interface{}{
			"trxnID":    stub.GetTxID(),
			"consentID": singleConsent.ConsentID,
			"message":   "Update Successful",
			"consent":   singleConsent,
		}
		returnResult = append(returnResult, returnStatus)

	}

	respJSON, _ := json.Marshal(returnResult)

	return shim.Success(respJSON)
}

//UpdateConsentExpiryDateByIDs changes the Expiry of the Consents based upon the given ConsentIds as passed in the args[o]
//args[0] string array of URN( consentIDs) for which the expiry need to be changed
//args[1] new exiry date in epoch format e.g. '1551788124'
//args[2] updateTs is the time to update the consent records
//Returns an array of map, each element of that array contains the below information:
//"trxnID"		:  transaction Id
//"consentID"	:  Consent Id for which data has been modified
//"status"		:  "Update Consent Expiry Successful"
//"consent"		:  Updated consent element
func (cm *ConsentManager) UpdateConsentExpiryDateByIDs(stub shim.ChaincodeStubInterface) pb.Response {
	_consentLogger.Info("Within UpdateConsentExpiryDateByIDs")
	_, args := stub.GetFunctionAndParameters()

	if len(args) == 0 {
		return shim.Error("Invalid No of arguments provided")
	}

	arr := make([]string, 0)
	err := json.Unmarshal([]byte(args[0]), &arr)
	if err != nil {
		_consentLogger.Errorf("Invalid json provided as input.")
		return shim.Error("{\"error\":\"Invalid json provided as input\"}")
	}

	newExpiryDate := args[1]
	if isValid, errMsg := isValidDate(newExpiryDate); !isValid {
		_consentLogger.Infof("Invalid Expiry Date to modify the consent. :", errMsg)
		return shim.Error("{\"error\":\"Invalid Expiry Date to modify the consent." + errMsg + "\"}")
	}

	newUpdatedTS := args[2]
	if isValid, errMsg := isValidDate(newUpdatedTS); !isValid {
		_consentLogger.Infof("Invalid Update Timestamp to modify the consent. :", errMsg)
		return shim.Error("{\"error\":\"Invalid Update Timestamp to modify the consent." + errMsg + "\"}")
	}

	returnStatus := make([]map[string]interface{}, 0)

	_, updatedBy := cm.getInvokerIdentity(stub)

	for _, searchConsentID := range arr {

		existingRec, err1 := stub.GetState(searchConsentID)

		if len(existingRec) == 0 {
			_consentLogger.Errorf("Consent does not exist with id ", searchConsentID)
			return shim.Error("{\"error\":\"Consent does not exist with id " + searchConsentID + ".\"}")
		}

		if err1 != nil {
			_consentLogger.Errorf("Error while fetching Consent with id ", searchConsentID)
			return shim.Error("{\"error\":\"Error while fetching Consent with id " + searchConsentID + ".\"}")
		}

		var updatedStatusConsent Consentdetails
		err := json.Unmarshal(existingRec, &updatedStatusConsent)
		if err != nil {
			return shim.Error("{\"error\":\"Error while unmarshaling the data.\"}")
		}

		updatedStatusConsent.ExpiryDate = newExpiryDate
		updatedStatusConsent.UpdateTs = newUpdatedTS
		updatedStatusConsent.UpdatedBy = updatedBy

		marshalConsentJSON, _ := json.Marshal(updatedStatusConsent)

		finalErr := stub.PutState(updatedStatusConsent.ConsentID, marshalConsentJSON)

		if finalErr != nil {
			_consentLogger.Errorf("Unable to save with consent id " + updatedStatusConsent.ConsentID)
			return shim.Error("{\"error\":\"Unable to save with consent id " + updatedStatusConsent.ConsentID + ".\"}")
		}

		p := EventPayLoad{consent: updatedStatusConsent, txnID: stub.GetTxID()}

		payloadbytes, _ := json.Marshal(p)

		retErr := stub.SetEvent(_UpdateEvent, payloadbytes)

		if retErr != nil {
			_consentLogger.Errorf("Event not generated for event : MODIFY_CONSENT")
		}

		eachReturnData := map[string]interface{}{
			"trxnID":    stub.GetTxID(),
			"consentID": updatedStatusConsent.ConsentID,
			"status":    "Update Consent Expiry Successful",
			"consent":   updatedStatusConsent,
		}

		returnStatus = append(returnStatus, eachReturnData)
	}

	respJSON, _ := json.Marshal(returnStatus)

	return shim.Success(respJSON)

}

//UpdateConsentExpiryDateByHeader updates the consent Expiry Date as input
//Method can be used to for bulk update of records
//args[0] should be an array of JSON objects of below structure (Keys are important to follow)
//"msisdn" : msisdn for which record should be fetched
//"cli"	 : cli ; combination of cli and msisdn should return a unique record from ledger
//args[1] : should be the expiry date to be updated with, in Epoch format (e.g. '1551788124' )
//args[2] : UpdateTimeStamp
//Returns an array of map, each element of that array contains the below information:
//"trxnID"		:  Transaction Id
//"consentID"	:  ConsentId
//"message"		:  "Update Consent Expiry Successful"
//"consent"		:  Updated consent element
func (cm *ConsentManager) UpdateConsentExpiryDateByHeader(stub shim.ChaincodeStubInterface) pb.Response {
	_consentLogger.Info("Within UpdateConsentExpiryDateByHeader")
	searchCriteriaArr := make([]map[string]string, 0)
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		_consentLogger.Errorf("Invalid number of arguments provided for transaction.")
		return shim.Error("{\"error\":\"Invalid number of arguments provided for transaction\"}")
	}
	err := json.Unmarshal([]byte(args[0]), &searchCriteriaArr)

	if err != nil {
		_consentLogger.Errorf("Invalid json provided as input.")
		return shim.Error("{\"error\":\"Invalid json provided as input\"}")
	}

	// expiryDate : target Expiry Date for all the element for the query result
	expiryDate := args[1] // should be date in Epoch format e.g. '1551788124'

	if isValid, errMsg := isValidDate(expiryDate); !isValid {
		_consentLogger.Infof("Invalid Expiry Date to update the consent. :", errMsg)
		return shim.Error("{\"error\":\"Invalid Expiry Date to update the consent." + errMsg + "\"}")
	}
	newUpdatedTS := args[2]
	if isValid, errMsg := isValidDate(newUpdatedTS); !isValid {
		_consentLogger.Infof("Invalid Update Date to update the consent. :", errMsg)
		return shim.Error("{\"error\":\"Invalid Update Date to update the consent." + errMsg + "\"}")
	}

	returnResult := make([]map[string]interface{}, 0)

	for _, searchCriteria := range searchCriteriaArr {

		consentSearchCriteria := `{
			"obj":"Consent"	,
			"msisdn":"%s",
			"cli":"%s"	
		}`

		msisdn := searchCriteria["msisdn"]
		cli := searchCriteria["cli"]

		consents := cm.retrieveConsentRecords(stub, fmt.Sprintf(consentSearchCriteria, msisdn, cli), "consentSearchByHeaderCliSts")

		if len(consents) == 0 {
			_consentLogger.Infof("No consent record found for :" + cli + ", MSISDN:" + msisdn)
			return shim.Error("{\"error\":\"No consent record found" + cli + ", MSISDN:" + msisdn + "\"}")
		}

		if len(consents) > 1 {
			_consentLogger.Infof("More than one consent record found ")
			return shim.Error("{\"error\":\"More than one consent record found \"}")
		}

		singleConsent := consents[0]
		singleConsent.ExpiryDate = expiryDate
		singleConsent.UpdateTs = newUpdatedTS

		_, updatedBy := cm.getInvokerIdentity(stub)
		singleConsent.UpdatedBy = updatedBy

		marshalConsentJSON, _ := json.Marshal(singleConsent)

		finalErr := stub.PutState(singleConsent.ConsentID, marshalConsentJSON)

		if finalErr != nil {
			_consentLogger.Infof("Unable to save with consent : %v", finalErr)
			return shim.Error("{\"error\":\"Unable to save with consent id " + singleConsent.ConsentID + "\"}")
		}

		p := EventPayLoad{consent: singleConsent, txnID: stub.GetTxID()}

		payloadbytes, _ := json.Marshal(p)

		retErr := stub.SetEvent(_UpdateEvent, payloadbytes)

		if retErr != nil {
			_consentLogger.Errorf("Event not generated for event : MODIFY_CONSENT")
		}

		//make the payload to return and pass it through the shim.success
		returnStatus := map[string]interface{}{
			"trxnID":    stub.GetTxID(),
			"consentID": singleConsent.ConsentID,
			"message":   "Update Consent Expiry Successful",
			"consent":   singleConsent,
		}
		returnResult = append(returnResult, returnStatus)

	}

	respJSON, _ := json.Marshal(returnResult)

	return shim.Success(respJSON)

}

//GetActiveConsentsByMSISDN returns a array of Consents with status 'approved' (sts = 2) - dafault, else pass a valid status as args[1]
//args[0] - MSISDN
//args[1] - Status, default value 2 (approved), set any other valid value for required result
func (cm *ConsentManager) GetActiveConsentsByMSISDN(stub shim.ChaincodeStubInterface) pb.Response {

	_, args := stub.GetFunctionAndParameters()
	msisdn := args[0]

	//setting default value 2 (approved)
	sts := "2"
	if len(args[1]) > 0 {
		if isValid, errMsg := isValidStatus(args[1]); !isValid {
			_consentLogger.Infof("Invalid status as Input. :", errMsg)
			return shim.Error("{\"error\":\"Invalid status as Input." + errMsg + "\"}")
		}
		sts = args[1]
	}
	consents := cm.getConsentsByPhoneNumber(stub, msisdn, sts)

	recordsJSON, _ := json.Marshal(consents)
	response := shim.Success(recordsJSON)

	return response
}

//RevokeActiveConsentsByMsisdn updates the status of the Active Consents (consents with status 2) for the matching msisdn. It updates the status by the input status. If no status is given, default value will be 3 (revoke).
//args[0] - MSISDN
//args[1] - UpdateTs
//args[2] - Status, default value 3 (revoke). Set any other valid status value for required result
func (cm *ConsentManager) RevokeActiveConsentsByMsisdn(stub shim.ChaincodeStubInterface) pb.Response {

	_consentLogger.Debug("RevokeActiveConsentsByMsisdn is being called.")

	_, args := stub.GetFunctionAndParameters()
	msisdn := args[0]

	var updateTs = ""
	if len(args[1]) > 0 {
		if isValid, errMsg := isValidDate(args[1]); !isValid {
			_consentLogger.Infof("Invalid Update Timestamp to modify the consent . :", errMsg)
			return shim.Error("{\"error\":\"Invalid Update Timestamp to modify the consent ." + errMsg + "\"}")
		}
		updateTs = args[1]
	}

	//setting default value 3 (Revoked)
	sts := "3"
	if len(args[2]) > 0 {
		if isValid, errMsg := isValidStatus(args[2]); !isValid {
			_consentLogger.Infof("Invalid status as Input. :", errMsg)
			return shim.Error("{\"error\":\"Invalid status as Input." + errMsg + "\"}")
		}
		sts = args[2]
	}

	//Active consents - status 2 ( Approved)
	activeSts := "2"

	consents := cm.getConsentsByPhoneNumber(stub, msisdn, activeSts)

	returnStatus := make([]map[string]interface{}, 0)

	_, updatedBy := cm.getInvokerIdentity(stub)

	for _, eachConsent := range consents {

		eachConsent.Status = sts
		eachConsent.UpdateTs = updateTs
		eachConsent.UpdatedBy = updatedBy

		consentJSON, _ := json.Marshal(eachConsent)

		_consentLogger.Info("Consent to Update :", eachConsent.ConsentID)

		err := stub.PutState(eachConsent.ConsentID, consentJSON)

		if err != nil {
			_consentLogger.Errorf(_Format3, eachConsent.ConsentID)

			return shim.Error(getErrorMsg(_Format3, eachConsent.ConsentID))
		}

		p := EventPayLoad{consent: eachConsent, txnID: stub.GetTxID()}

		payloadbytes, _ := json.Marshal(p)

		retErr := stub.SetEvent(_CreateEvent, payloadbytes)

		if retErr != nil {
			_consentLogger.Errorf(_Format5, _CreateEvent)
		}

		resultData := map[string]interface{}{
			"trxnID":    stub.GetTxID(),
			"consentID": eachConsent.ConsentID,
			"message":   "Consent Revoke Successful",
			"consent":   eachConsent,
		}

		returnStatus = append(returnStatus, resultData)

	}
	respJSON, _ := json.Marshal(returnStatus)
	return shim.Success(respJSON)
}

//GetHistoryByKey queries the ledger using the given key.
//args[0] takes the key for search input
//It retrieve all the changes to the value happened over time as input given, across time
func (cm *ConsentManager) GetHistoryByKey(stub shim.ChaincodeStubInterface) peer.Response {

	_consentLogger.Debug("getHistoryByKey is being called.")

	_, args := stub.GetFunctionAndParameters()

	// Essential check to verify number of arguments
	if len(args) != 1 {
		_consentLogger.Error("Incorrect number of arguments passed in getHistoryByKey.")
		return shim.Error("{\"error\":\"Incorrect number of arguments. Expecting 1 arguments:" + strconv.Itoa(len(args)) + "given.\"}")
	}

	key := args[0]
	resultsIterator, err := stub.GetHistoryForKey(key)

	if err != nil {
		_consentLogger.Error("Error occured while calling getHistoryByKey(): ", err)
		return shim.Error("{\"error\":\"Error occured while calling getHistoryByKey(): " + err.Error() + ".\"}")
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the event
	historicResponse := make([]map[string]interface{}, 0)
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			_consentLogger.Error("Error occured while calling resultsIterator.Next(): ", err)
			return shim.Error("{\"error\":\"Error occured while calling GetHistoryByKey (resultsIterator): " + err.Error() + ".\"}")
		}
		value := make(map[string]interface{})
		json.Unmarshal(response.Value, &value)
		historicResponse = append(historicResponse, map[string]interface{}{"txId": response.TxId, "value": value})

	}

	respJSON, _ := json.Marshal(historicResponse)
	return shim.Success(respJSON)
}

//getConsentsByPhoneNumber returns the consents upon the given MSISDN and Status
func (cm *ConsentManager) getConsentsByPhoneNumber(stub shim.ChaincodeStubInterface, msisdn, sts string) []Consentdetails {
	consentSearchCriteria := `{
		"obj":"Consent"	,
		"msisdn":"%s",
		"sts":"%s"	
	}`

	consents := cm.retrieveConsentRecords(stub, fmt.Sprintf(consentSearchCriteria, msisdn, sts), "consentSearchByMsisdnSts")

	return consents
}

//getConsentsByMsisdnCli returns the consents upon the given MSISDN and Cli (header)
func (cm *ConsentManager) getConsentsByMsisdnCli(stub shim.ChaincodeStubInterface, msisdn, cli string) []Consentdetails {
	consentSearchCriteria := `{
		"obj":"Consent"	,
		"msisdn":"%s",
		"cli":"%s"	
	}`

	consents := cm.retrieveConsentRecords(stub, fmt.Sprintf(consentSearchCriteria, msisdn, cli), "consentSearchByHeaderCliSts")

	return consents
}

//CheckValidityForStatus checks for  validity of Consent for status trxn
func checkValidityForStatus(searchConsentID, newStatus, newUpdatedTS string) (bool, string) {
	if searchConsentID == "" {
		return false, "Consent Id should be present"
	}
	if newStatus == "" {
		return false, "Status should be present"
	}
	if newUpdatedTS == "" {
		return false, "Update timeStamp should be present"
	}
	return true, ""
}

//getInvokerIdentity returns complete identity in the format <MSPID>/<ISSUERID>/<SUBJECTNAME>
//Returns string Unknown if not able parse the invoker certificate
func (cm *ConsentManager) getInvokerIdentity(stub shim.ChaincodeStubInterface) (bool, string) {
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

//getErrorMsg generates the Error message based upon the error message format as input
func getErrorMsg(format string, msg ...string) string {
	//err := "{\"error\":\"Invalid json provided as input :\"}"
	var buffer bytes.Buffer
	buffer.WriteString("{\"error\":\"")

	buffer.WriteString(format)

	for i, ms := range msg {
		if i != 0 {
			buffer.WriteString(" ")
		}

		buffer.WriteString(ms)
		if i != len(msg)-1 {

			buffer.WriteString(" ")
		}
	}

	buffer.WriteString("\"}")

	return buffer.String()
}

//retrieveConsentRecords fetches the consent record for trhe given sea4rch criteria
func (cm *ConsentManager) retrieveConsentRecords(stub shim.ChaincodeStubInterface, criteria string, indexs ...string) []Consentdetails {
	var finalSelector string
	records := make([]Consentdetails, 0)

	if len(indexs) == 0 {
		finalSelector = fmt.Sprintf("{\"selector\":%s }", criteria)

	} else {
		finalSelector = fmt.Sprintf("{\"selector\":%s , \"use_index\" :\"%s\" }", criteria, indexs[0])
	}

	_consentLogger.Infof("Query Selector : %s", finalSelector)
	resultsIterator, _ := stub.GetQueryResult(finalSelector)
	for resultsIterator.HasNext() {
		record := Consentdetails{}
		recordBytes, _ := resultsIterator.Next()
		err := json.Unmarshal(recordBytes.Value, &record)
		if err != nil {
			_consentLogger.Infof("Unable to unmarshal consent retrieves:: %v", err)
		}
		records = append(records, record)
	}
	return records
}
