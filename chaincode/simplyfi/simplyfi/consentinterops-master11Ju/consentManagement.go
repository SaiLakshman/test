package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

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
	Purpose           string `json:"pur"`
	Creator           string `json:"crtr"`
	UpdateTs          string `json:"uts"`
	CreateTs          string `json:"cts"`
	UpdatedBy         string `json:"uby"`
	UpdatedOrg        string `json:"uorg"`
	CommunicationMode string `json:"cmode"`
}

//ErrorData holds only Error Consesnts
type ErrorData struct {
	ID  string `json:"id"`
	Msg string `json:"errormsg"`
}

//SuccessData holds only Success Consesnts
type SuccessData struct {
	TrxnID      string         `json:"trxnId"`
	ConsID      string         `json:"consId"`
	Message     string         `json:"message"`
	ConsentDets Consentdetails `json:"consentDets"`
}

//TotalResponse holds both Success and Error Consents
type TotalResponse struct {
	SuccesConsents []SuccessData `json:"successData"`
	FailedConsents []ErrorData   `json:"failedData"`
}

//ConsentStatus is strcuture of ConsentStatus
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

//commMode - valid values of Communication Mode
var commMode = map[string]bool{
	"0": true, //Migration
	"1": true, //WEB
	"2": true, //SMS
	"3": true, //IVR
	"4": true, //USSD
	"5": true, //APP
	"6": true, //Customer Support
}

//consentStatus - valid values of Consent Status
var consentStatus = map[string]bool{
	"1": true, //ConsentRaised
	"2": true, //Approved
	"3": true, //Revoked
	"4": true, //Churned
}

//purposeValues  - Valid values of Purpose Values
var purposeValues = map[string]bool{
	"1": true, //1- Promotional
	"2": true, //2- Service
	"3": true, //3 -Both
}

const _CreateEvent = "CREATE_CONSENT"
const _UpdateEvent = "UPDATE_CONSENT"

//Object type for Consent - do not change
const _ObjectType = "Consent"
const _ConsentChurnedStatus = "4"
const _ConsentRevokedStatus = "3"
const _ConsentApprovedStatus = "2"
const _ConsentRaisedStatus = "1"

//below are the error message format
const _Format0 = "Invalid json provided as input."
const _Format1 = "Invalid number of arguments provided for transaction."
const _Format2 = "Unable to save consent. "
const _Format3 = "Unable to modify the consent. "
const _Format4 = "Consent ID already registered. "
const _Format4A = "Consent already registered with given MSISDN and HEADER. "
const _Format5 = "Event not generated for event. "
const _Format6 = "No Matching Consent record found. "
const _Format7 = "More than one consent record found"
const _Format8 = "Error while unmarshaling the data"
const _Format9 = "Unable to save with consent id"

var _consentLogger = shim.NewLogger("ConsentManagementSmartContract")

//ConsentManager is ConsentManager
type ConsentManager struct {
}

//isValidConsent check the validity of the consent
func isValidConsent(c Consentdetails) (bool, string) {
	if len(c.ConsentID) == 0 {
		return false, "ConsentId is mandatory"
	}
	/* if len(c.ConsentTemplateID) == 0 {
		return false, "Consent TemplateId is mandatory"
	} */
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
	if !validEnumEntry(c.CommunicationMode, commMode) {
		return false, "Communication mode can be either (0)Migration, (1)WEB, (2)SMS, (3)IVR, (4)USSD, (5)APP or (6)Customer Support"
	}
	if !validEnumEntry(c.Purpose, purposeValues) {
		return false, "Purpose can be either (1) Both, (2)Promotional or (3)Service"
	}
	return true, ""
}

func isValidStatus(status string) (bool, string) {
	if !validEnumEntry(status, consentStatus) {
		return false, "Status can be either (1)Consent Raised, (2)Approved, (3)Revoked or (4)PD/Churned"
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
	// example of epock time 1551788124
	if date != "" {
		_, err := strconv.ParseInt(date, 10, 64)
		if err != nil {
			return false, "Date needs to be in Epoch format e.g. '1551788124'"
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

func isValidPurpose(purpose string) (bool, string) {
	if !validEnumEntry(purpose, purposeValues) {
		return false, "Purpose can be either 1(Promotional), 2(Service) or (3) Both"
	}
	return true, ""
}

func validEnumEntry(input string, enumMap map[string]bool) bool {
	if _, isEntryExists := enumMap[input]; !isEntryExists {
		return false
	}
	return true
}

//RecordConsent saves the consent in DLT with the given input
//Takes an array of consents from args[0] to be stored ( eg. []Consentdetails )
//Returned payload contains two blocks - 'failedData' and 'successData'. FailedData contains an array of map with Falure details and SuccessData contains an array of map with details of Successfully Saved Consent.
//--FailedData block is as below:
//"errormsg": <Reason for Failure>,
//"id"		: <URN>
//--SuccessData block is as below:
//"consId": <URN>,
//"consentDets": <saved consent>,
//"message": "Consent Creation Successful",
//"trxnId": <transactionid>
//Tips: Check length of the FailedData array to know, how many Consents failed to save.
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
	}

	//Success Consents Message
	sConsents := make([]SuccessData, 0)
	//Failed Consesnts Message
	fConsents := make([]ErrorData, 0)

	_, creater := cm.getInvokerIdentity(stub)

	for _, eachConsent := range consents {

		if recordBytes, _ := stub.GetState(eachConsent.ConsentID); len(recordBytes) > 0 {

			_consentLogger.Infof(_Format4)
			e := ErrorData{ID: eachConsent.ConsentID, Msg: _Format4}
			fConsents = append(fConsents, e)
			continue
		}

		consents := cm.getConsentsByMsisdnCli(stub, eachConsent.Msisdn, eachConsent.Cli)

		var flag bool
		if len(consents) > 0 {
			_consentLogger.Infof("More than one record found for given MSISDN + CLI")

			for _, cons := range consents {
				if cons.Status == _ConsentRaisedStatus || cons.Status == _ConsentApprovedStatus {
					flag = true
				}
			}
			if flag {
				_consentLogger.Infof(_Format4A)
				e := ErrorData{ID: eachConsent.ConsentID, Msg: _Format4A + "Status is either 1 or 2"}
				fConsents = append(fConsents, e)
				continue
			}

			/* if !flag {
				_consentLogger.Infof("Warning:Consent(s) with status either  3 / 4 is already there in Ledger")
			} */
		}

		//validation
		if isValid, errMsg := isValidConsent(eachConsent); !isValid {

			_consentLogger.Infof(_Format2)
			e := ErrorData{ID: eachConsent.ConsentID, Msg: errMsg}
			fConsents = append(fConsents, e)
			continue

		}

		if isValid, errMsg := isValidConsentToRecord(eachConsent); !isValid {
			_consentLogger.Infof(_Format2, ".Error :", eachConsent.ConsentID, errMsg)

			e := ErrorData{ID: eachConsent.ConsentID, Msg: errMsg}
			fConsents = append(fConsents, e)
			continue

		}

		//Update each Consent Object
		eachConsent.ObjectType = _ObjectType
		eachConsent.Status = _ConsentRaisedStatus

		eachConsent.Creator = creater
		eachConsent.UpdatedBy = creater

		consentJSON, _ := json.Marshal(eachConsent)

		_consentLogger.Info("Consent to Save :", eachConsent.ConsentID)

		err = stub.PutState(eachConsent.ConsentID, consentJSON)
		if err != nil {
			_consentLogger.Errorf(_Format2, eachConsent.ConsentID)
			e := ErrorData{ID: eachConsent.ConsentID, Msg: _Format2}
			fConsents = append(fConsents, e)
			continue
		}

		p := EventPayLoad{consent: eachConsent, txnID: stub.GetTxID()}

		payloadbytes, _ := json.Marshal(p)

		retErr := stub.SetEvent(_CreateEvent, payloadbytes)

		if retErr != nil {
			_consentLogger.Errorf(_Format5, _CreateEvent)
		}

		resultData := SuccessData{TrxnID: stub.GetTxID(), ConsID: eachConsent.ConsentID, Message: "Consent Creation Successful", ConsentDets: eachConsent}

		sConsents = append(sConsents, resultData)

	}

	totalResponse := TotalResponse{SuccesConsents: sConsents, FailedConsents: fConsents}

	respJSON, _ := json.Marshal(totalResponse)
	return shim.Success(respJSON)
}

//RecordConsentInBulk records the Consents in Bulk
//LESS Validation
//Dev In progress
func (cm *ConsentManager) RecordConsentInBulk(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()

	if len(args) < 1 {
		_consentLogger.Errorf("Invalid number of arguments provided for transaction.")
		return shim.Error(getErrorMsg(_Format1))

	}

	var consents []Consentdetails

	err := json.Unmarshal([]byte(args[0]), &consents)
	if err != nil {
		_consentLogger.Errorf(_Format0)
		return shim.Error(getErrorMsg(_Format0))
	}

	//Success Consents Message
	sConsents := make([]SuccessData, 0)
	//Failed Consesnts Message
	fConsents := make([]ErrorData, 0)

	_, creater := cm.getInvokerIdentity(stub)
	phoneNos := make([]string, 0)
	for _, eachConsent := range consents {

		if recordBytes, _ := stub.GetState(eachConsent.ConsentID); len(recordBytes) > 0 {

			_consentLogger.Infof(_Format4)
			e := ErrorData{ID: eachConsent.ConsentID, Msg: _Format4}
			fConsents = append(fConsents, e)
			continue
		}

		consents := cm.getConsentsByMsisdnCli(stub, eachConsent.Msisdn, eachConsent.Cli)

		var flag bool
		if len(consents) > 0 {
			_consentLogger.Infof("More than one record found for given MSISDN + CLI")

			for _, cons := range consents {
				if cons.Status == _ConsentApprovedStatus {
					flag = true
				}
			}
			if flag {
				_consentLogger.Infof(_Format4)
				e := ErrorData{ID: eachConsent.ConsentID, Msg: _Format4 + "Status is 2."}
				fConsents = append(fConsents, e)
			}
		}

		//validation
		/*
			if isValid, errMsg := isValidConsent(eachConsent); !isValid {
				_consentLogger.Infof(_Format2)
				e := ErrorData{ID: eachConsent.ConsentID, Msg: errMsg}
				fConsents = append(fConsents, e)
				continue
			}

			if isValid, errMsg := isValidConsentToRecord(eachConsent); !isValid {
				_consentLogger.Infof(_Format2, ".Error :", eachConsent.ConsentID, errMsg)

				e := ErrorData{ID: eachConsent.ConsentID, Msg: errMsg}
				fConsents = append(fConsents, e)
				continue
			} */

		//Update each Consent Object
		eachConsent.ObjectType = _ObjectType
		eachConsent.Status = _ConsentApprovedStatus

		eachConsent.Creator = creater
		eachConsent.UpdatedBy = creater

		consentJSON, _ := json.Marshal(eachConsent)

		_consentLogger.Info("Consent to Save :", eachConsent.ConsentID)

		err = stub.PutState(eachConsent.ConsentID, consentJSON)
		if err != nil {
			_consentLogger.Errorf(_Format2, eachConsent.ConsentID)
			e := ErrorData{ID: eachConsent.ConsentID, Msg: _Format2}
			fConsents = append(fConsents, e)
			continue
		}

		if !hasElem(phoneNos, eachConsent.Msisdn) {
			phoneNos = append(phoneNos, eachConsent.Msisdn)
		}

		resultData := SuccessData{TrxnID: stub.GetTxID(), ConsID: eachConsent.ConsentID, Message: "Bulk Consents Load Successful", ConsentDets: eachConsent}

		sConsents = append(sConsents, resultData)

	}

	totalResponse := TotalResponse{SuccesConsents: sConsents, FailedConsents: fConsents}

	respJSON, _ := json.Marshal(totalResponse)

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
//"message"		:   "Update Consent Successful",
//"consent"		:   Updated consent element,
func (cm *ConsentManager) UpdateConsentStatus(stub shim.ChaincodeStubInterface) pb.Response {
	_consentLogger.Info("Within UpdateConsentStatus")
	_, args := stub.GetFunctionAndParameters()

	if len(args) != 3 {
		jsonResp := "{\"error\":\"Invalid Number of argumnets provided for transaction\"}"
		return shim.Error(jsonResp)
	}

	searchConsentID := args[0]
	searchConsentID = strings.TrimSpace(searchConsentID)
	newStatus := args[1]
	newStatus = strings.TrimSpace(newStatus)
	newUpdatedTS := args[2]

	if isValid, errMsg := isValidDate(newUpdatedTS); !isValid {
		_consentLogger.Infof("Invalid Update Timestamp to modify the consent. :", errMsg)
		return shim.Error("{\"error\":\"Invalid Update Timestamp to modify the consent." + errMsg + "\"}")
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
		jsonResp := "{\"error\":\"" + _Format8 + "\"}"
		return shim.Error(jsonResp)
	}

	updatedStatusConsent.Status = newStatus
	updatedStatusConsent.UpdateTs = newUpdatedTS

	_, updatedBy := cm.getInvokerIdentity(stub)
	updatedStatusConsent.UpdatedBy = updatedBy

	marshalConsentJSON, _ := json.Marshal(updatedStatusConsent)

	finalErr := stub.PutState(updatedStatusConsent.ConsentID, marshalConsentJSON)
	if finalErr != nil {
		_consentLogger.Errorf(_Format9 + updatedStatusConsent.ConsentID)
		return shim.Error("{\"error\":\"" + _Format9 + updatedStatusConsent.ConsentID + ".\"}")
	}

	p := EventPayLoad{consent: updatedStatusConsent, txnID: stub.GetTxID()}

	payloadbytes, _ := json.Marshal(p)

	retErr := stub.SetEvent(_UpdateEvent, payloadbytes)

	if retErr != nil {
		_consentLogger.Errorf("Event not generated for event : UPDATE_CONSENT")
	}

	resultData := map[string]interface{}{
		"trxnID":    stub.GetTxID(),
		"consentID": updatedStatusConsent.ConsentID,
		"message":   "Update Consent Successful",
		"consent":   updatedStatusConsent,
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)

}

//UpdateConsentStatusByIDs changes the status of the Consents based upon the given ConsentIds as passed in the args[o]
//args[0] string array of URNs(consentIDs) for which the status need to be changed to Revoked (3)
//args[1] new status to be updated with
//args[2] updateTs is the time to update the consent records
//Returned payload contains two blocks - 'failedData' and 'successData'. FailedData contains an array of map with Falure details and SuccessData contains an array of map with details of Successfully Saved Consent.
//--FailedData block is as below:
//"errormsg": <Reason for Failure>,
//"id"		: <URN>
//--SuccessData block is as below:
//"consId": <URN>,
//"consentDets": <saved consent>,
//"message": "Consent Status Update Successful",
//"trxnId": <transactionid>
func (cm *ConsentManager) UpdateConsentStatusByIDs(stub shim.ChaincodeStubInterface) pb.Response {
	_consentLogger.Info("Within UpdateConsentStatusByIDs")
	_, args := stub.GetFunctionAndParameters()

	var jsonResp = ""
	if len(args) != 3 {
		jsonResp = "{\"error\":\"Invalid Number of argumnets provided for transaction\"}"
		return shim.Error(jsonResp)
	}

	arr := make([]string, 0)
	err := json.Unmarshal([]byte(args[0]), &arr)
	if err != nil {
		_consentLogger.Errorf("Invalid json provided as input.")
		jsonResp = "{\"error\":\"Invalid json provided as input\"}"
		return shim.Error(jsonResp)
	}

	newStatus := args[1]
	newStatus = strings.TrimSpace(newStatus)
	if isValid, errMsg := isValidStatus(newStatus); !isValid {
		_consentLogger.Infof("Unable to modify the consent. :", errMsg)
		jsonResp = "{\"error\":\"Invalid json provided as input\"}" + errMsg + "\"}"
		return shim.Error(jsonResp)
	}

	newUpdatedTS := args[2]
	if isValid, errMsg := isValidDate(newUpdatedTS); !isValid {
		_consentLogger.Infof("Invalid Update Timestamp to modify the consent. :", errMsg)
		jsonResp = "{\"error\":\"Invalid Update Timestamp to modify the consent.\"}" + errMsg + "\"}"
		return shim.Error(jsonResp)
	}

	_, updatedBy := cm.getInvokerIdentity(stub)

	//Success Consents Message
	sConsents := make([]SuccessData, 0)
	//Failed Consesnts Message
	fConsents := make([]ErrorData, 0)

	for _, searchConsentID := range arr {

		searchConsentID = strings.TrimSpace(searchConsentID)

		existingRec, err1 := stub.GetState(searchConsentID)

		if len(existingRec) == 0 {
			_consentLogger.Errorf("Consent does not exist with id ", searchConsentID)

			e := ErrorData{ID: searchConsentID, Msg: "Consent does not exist with id " + searchConsentID}
			fConsents = append(fConsents, e)
			continue
		}

		if err1 != nil {
			_consentLogger.Errorf("Error while fetching Consent with id ", searchConsentID)

			e := ErrorData{ID: searchConsentID, Msg: "Consent does not exist with id " + searchConsentID}
			fConsents = append(fConsents, e)
			continue
		}

		var updatedStatusConsent Consentdetails
		err := json.Unmarshal(existingRec, &updatedStatusConsent)
		if err != nil {
			e := ErrorData{ID: searchConsentID, Msg: _Format8}
			fConsents = append(fConsents, e)
			continue
		}

		updatedStatusConsent.Status = newStatus
		updatedStatusConsent.UpdateTs = newUpdatedTS
		updatedStatusConsent.UpdatedBy = updatedBy

		marshalConsentJSON, _ := json.Marshal(updatedStatusConsent)

		finalErr := stub.PutState(updatedStatusConsent.ConsentID, marshalConsentJSON)

		if finalErr != nil {
			_consentLogger.Errorf("Unable to save with consent id " + updatedStatusConsent.ConsentID)

			e := ErrorData{ID: searchConsentID, Msg: "Unable to save with consent id " + updatedStatusConsent.ConsentID}
			fConsents = append(fConsents, e)
			continue
		}

		p := EventPayLoad{consent: updatedStatusConsent, txnID: stub.GetTxID()}

		payloadbytes, _ := json.Marshal(p)

		retErr := stub.SetEvent(_UpdateEvent, payloadbytes)

		if retErr != nil {
			_consentLogger.Errorf("Event not generated for event : UPDATE_CONSENT")
		}

		resultData := SuccessData{TrxnID: stub.GetTxID(), ConsID: updatedStatusConsent.ConsentID, Message: "Consent Status Update Successful", ConsentDets: updatedStatusConsent}

		sConsents = append(sConsents, resultData)

	}
	totalResponse := TotalResponse{SuccesConsents: sConsents, FailedConsents: fConsents}

	respJSON, _ := json.Marshal(totalResponse)

	return shim.Success(respJSON)

}

//UpdateConsentStatusByHeader updates the consent status to either 'Approved' (2), 'Revoked'(3) or any other valid status
//Method can be used to for bulk update of records
//args[0] should be an array of JSON objects of below structure (Keys are important to follow)
//"msisdn" : msisdn for which record should be fetched
//"cli"	 : cli ; combination of cli and msisdn should return a unique record from ledger
//args[1] : should be the status of the updated record, can be either 2, 3 or any valid one
//args[2] : UpdateTimestamp
//Returned payload contains two blocks - 'failedData' and 'successData'. FailedData contains an array of map with Falure details and SuccessData contains an array of map with details of Successfully Saved Consent.
//--FailedData block is as below:
//"errormsg": <Reason for Failure>,
//"id"		: <URN>
//--SuccessData block is as below:
//"consId": <URN>,
//"consentDets": <saved consent>,
//"message": "Update Consent Successful",
//"trxnId": <transactionid>
func (cm *ConsentManager) UpdateConsentStatusByHeader(stub shim.ChaincodeStubInterface) pb.Response {
	_consentLogger.Info("Within UpdateConsentStatusByHeader")

	searchCriteriaArr := make([]map[string]string, 0)
	_, args := stub.GetFunctionAndParameters()

	if len(args) != 3 {
		jsonResp := "{\"error\":\"Invalid Number of argumnets provided for transaction\"}"
		return shim.Error(jsonResp)
	}
	err := json.Unmarshal([]byte(args[0]), &searchCriteriaArr)

	if err != nil {
		_consentLogger.Errorf("Invalid json provided as input.")
		return shim.Error("{\"error\":\"Invalid json provided as input\"}")
	}

	// stsT : target status for all the element for the query result
	stsT := args[1] // can be either 2,3, or any other valid state

	if isValid, errMsg := isValidStatus(stsT); !isValid {
		_consentLogger.Infof("Invalid Status to update the consent. :", errMsg)
		return shim.Error("{\"error\":\"Invalid Status to update the consent." + errMsg + "\"}")
	}

	newUpdatedTS := args[2]
	if isValid, errMsg := isValidDate(newUpdatedTS); !isValid {
		_consentLogger.Infof("Invalid Update Timestamp to modify the consent. :", errMsg)
		return shim.Error("{\"error\":\"Invalid Update Timestamp to modify the consent." + errMsg + "\"}")
	}

	//Success Consents Message
	sConsents := make([]SuccessData, 0)
	//Failed Consesnts Message
	fConsents := make([]ErrorData, 0)

	_, updatedBy := cm.getInvokerIdentity(stub)

	for _, searchCriteria := range searchCriteriaArr {

		msisdn := searchCriteria["msisdn"]
		msisdn = strings.TrimSpace(msisdn)
		cli := searchCriteria["cli"]
		cli = strings.TrimSpace(cli)

		consents := cm.getConsentsByMsisdnCli(stub, msisdn, cli)

		if len(consents) == 0 {
			_consentLogger.Infof(_Format6)
			e := ErrorData{ID: "msisdn:" + msisdn + ",cli:" + cli, Msg: _Format6}
			fConsents = append(fConsents, e)
			continue
		}

		if len(consents) > 1 {
			_consentLogger.Infof(_Format7)
			e := ErrorData{ID: "msisdn:" + msisdn + ",cli:" + cli, Msg: _Format7}
			fConsents = append(fConsents, e)
			continue
		}

		singleConsent := consents[0]
		singleConsent.Status = stsT
		singleConsent.UpdateTs = newUpdatedTS

		singleConsent.UpdatedBy = updatedBy

		marshalConsentJSON, _ := json.Marshal(singleConsent)

		finalErr := stub.PutState(singleConsent.ConsentID, marshalConsentJSON)

		if finalErr != nil {
			_consentLogger.Infof(_Format9+" : %v", finalErr)
			e := ErrorData{ID: singleConsent.ConsentID, Msg: _Format9}
			fConsents = append(fConsents, e)
		}

		p := EventPayLoad{consent: singleConsent, txnID: stub.GetTxID()}

		payloadbytes, _ := json.Marshal(p)

		retErr := stub.SetEvent(_UpdateEvent, payloadbytes)

		if retErr != nil {
			_consentLogger.Errorf("Event not generated for event : UPDATE_CONSENT")
		}

		resultData := SuccessData{TrxnID: stub.GetTxID(), ConsID: singleConsent.ConsentID, Message: "Update Consent Successful", ConsentDets: singleConsent}

		sConsents = append(sConsents, resultData)
	}

	totalResponse := TotalResponse{SuccesConsents: sConsents, FailedConsents: fConsents}

	respJSON, _ := json.Marshal(totalResponse)

	return shim.Success(respJSON)
}

//UpdateConsentExpiryDateByIDs changes the Expiry of the Consents based upon the given ConsentIds as passed in the args[o]
//args[0] string array of URN( consentIDs) for which the expiry need to be changed
//args[1] new exiry date in epoch format e.g. '1551788124'
//args[2] updateTs is the time to update the consent records
//Returned payload contains two blocks - 'failedData' and 'successData'. FailedData contains an array of map with Falure details and SuccessData contains an array of map with details of Successfully Saved Consent.
//--FailedData block is as below:
//"errormsg": <Reason for Failure>,
//"id"		: <URN>
//--SuccessData block is as below:
//"consId": <URN>,
//"consentDets": <saved consent>,
//"message": "Update Consent Expiry Successful",
//"trxnId": <transactionid>
func (cm *ConsentManager) UpdateConsentExpiryDateByIDs(stub shim.ChaincodeStubInterface) pb.Response {
	_consentLogger.Info("Within UpdateConsentExpiryDateByIDs")
	_, args := stub.GetFunctionAndParameters()

	if len(args) != 3 {
		jsonResp := "{\"error\":\"Invalid Number of argumnets provided for transaction\"}"
		return shim.Error(jsonResp)
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

	//Success Consents Message
	sConsents := make([]SuccessData, 0)
	//Failed Consesnts Message
	fConsents := make([]ErrorData, 0)

	_, updatedBy := cm.getInvokerIdentity(stub)

	for _, searchConsentID := range arr {

		existingRec, err1 := stub.GetState(searchConsentID)

		if len(existingRec) == 0 {
			_consentLogger.Errorf("Consent does not exist with id ", searchConsentID)
			e := ErrorData{ID: searchConsentID, Msg: _Format6}
			fConsents = append(fConsents, e)
			continue
		}

		if err1 != nil {
			_consentLogger.Errorf("Error while fetching Consent with id ", searchConsentID)

			e := ErrorData{ID: searchConsentID, Msg: "Error while fetching Consent with id"}
			fConsents = append(fConsents, e)
			continue
		}

		var updatedStatusConsent Consentdetails
		err := json.Unmarshal(existingRec, &updatedStatusConsent)
		if err != nil {
			e := ErrorData{ID: searchConsentID, Msg: _Format8}
			fConsents = append(fConsents, e)
			continue
		}

		updatedStatusConsent.ExpiryDate = newExpiryDate
		updatedStatusConsent.UpdateTs = newUpdatedTS
		updatedStatusConsent.UpdatedBy = updatedBy

		marshalConsentJSON, _ := json.Marshal(updatedStatusConsent)

		finalErr := stub.PutState(updatedStatusConsent.ConsentID, marshalConsentJSON)

		if finalErr != nil {
			_consentLogger.Errorf("Unable to save with consent id " + updatedStatusConsent.ConsentID)

			e := ErrorData{ID: updatedStatusConsent.ConsentID, Msg: "Unable to save with consent id"}
			fConsents = append(fConsents, e)
			continue
		}

		p := EventPayLoad{consent: updatedStatusConsent, txnID: stub.GetTxID()}

		payloadbytes, _ := json.Marshal(p)

		retErr := stub.SetEvent(_UpdateEvent, payloadbytes)

		if retErr != nil {
			_consentLogger.Errorf("Event not generated for event : UPDATE_CONSENT")
		}

		resultData := SuccessData{TrxnID: stub.GetTxID(), ConsID: updatedStatusConsent.ConsentID, Message: "Update Consent Expiry Successful", ConsentDets: updatedStatusConsent}

		sConsents = append(sConsents, resultData)
	}

	totalResponse := TotalResponse{SuccesConsents: sConsents, FailedConsents: fConsents}

	respJSON, _ := json.Marshal(totalResponse)

	return shim.Success(respJSON)

}

//UpdateConsentExpiryDateByHeader updates the consent Expiry Date as input only for the active Consents ( status = 2)
//Method can be used to for bulk update of records
//args[0] should be an array of JSON objects of below structure (Keys are important to follow)
//"msisdn" : msisdn for which record should be fetched
//"cli"	 : cli ; combination of cli and msisdn should return a unique record from ledger
//args[1] : should be the expiry date to be updated with, in Epoch format (e.g. '1551788124' )
//args[2] : UpdateTimeStamp
//Returned payload contains two blocks - 'failedData' and 'successData'. FailedData contains an array of map with Falure details and SuccessData contains an array of map with details of Successfully Saved Consent.
//--FailedData block is as below:
//"errormsg": <Reason for Failure>,
//"id"		: <URN>
//--SuccessData block is as below:
//"consId": <URN>,
//"consentDets": <saved consent>,
//"message": "Update Consent Expiry Date Successful",
//"trxnId": <transactionid>
func (cm *ConsentManager) UpdateConsentExpiryDateByHeader(stub shim.ChaincodeStubInterface) pb.Response {
	_consentLogger.Info("Within UpdateConsentExpiryDateByHeader")

	searchCriteriaArr := make([]map[string]string, 0)

	_, args := stub.GetFunctionAndParameters()

	if len(args) != 3 {
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

	//returnResult := make([]map[string]interface{}, 0)

	//Success Consents Message
	sConsents := make([]SuccessData, 0)
	//Failed Consesnts Message
	fConsents := make([]ErrorData, 0)
	_, updatedBy := cm.getInvokerIdentity(stub)

	for _, searchCriteria := range searchCriteriaArr {

		msisdn := searchCriteria["msisdn"]
		msisdn = strings.TrimSpace(msisdn)
		cli := searchCriteria["cli"]
		cli = strings.TrimSpace(cli)

		sts := _ConsentApprovedStatus

		consents := cm.getConsentsByMsisdnCliStatus(stub, msisdn, cli, sts)

		if len(consents) == 0 {
			_consentLogger.Infof("No consent record found for :" + cli + ", MSISDN:" + msisdn)

			e := ErrorData{ID: "msisdn:" + msisdn + ",cli:" + cli, Msg: _Format6}
			fConsents = append(fConsents, e)
			continue
		}

		if len(consents) > 1 {
			_consentLogger.Infof(_Format7)
			e := ErrorData{ID: "msisdn:" + msisdn + ",cli:" + cli, Msg: _Format7}
			fConsents = append(fConsents, e)
			continue
		}

		singleConsent := consents[0]
		singleConsent.ExpiryDate = expiryDate
		singleConsent.UpdateTs = newUpdatedTS

		singleConsent.UpdatedBy = updatedBy

		marshalConsentJSON, _ := json.Marshal(singleConsent)

		finalErr := stub.PutState(singleConsent.ConsentID, marshalConsentJSON)

		if finalErr != nil {
			_consentLogger.Infof("Unable to save with consent : %v", finalErr)

			e := ErrorData{ID: singleConsent.ConsentID, Msg: _Format2}
			fConsents = append(fConsents, e)
			continue
		}

		p := EventPayLoad{consent: singleConsent, txnID: stub.GetTxID()}

		payloadbytes, _ := json.Marshal(p)

		retErr := stub.SetEvent(_UpdateEvent, payloadbytes)

		if retErr != nil {
			_consentLogger.Errorf("Event not generated for event : UPDATE_CONSENT")
		}

		resultData := SuccessData{TrxnID: stub.GetTxID(), ConsID: singleConsent.ConsentID, Message: "Update Consent Expiry Date Successful", ConsentDets: singleConsent}

		sConsents = append(sConsents, resultData)
	}

	totalResponse := TotalResponse{SuccesConsents: sConsents, FailedConsents: fConsents}

	respJSON, _ := json.Marshal(totalResponse)

	return shim.Success(respJSON)

}

//GetActiveConsentsByMSISDN returns a array of Consents with status 'approved' (sts = 2) for the given MSISDN. Default is Approved. Different status can be given to fetch different set of Consent records
//args[0] - MSISDN
//args[1] - Status, default value 2 (approved), set any other valid value for required result
func (cm *ConsentManager) GetActiveConsentsByMSISDN(stub shim.ChaincodeStubInterface) pb.Response {
	_consentLogger.Info("Within GetActiveConsentsByMSISDN")

	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		_consentLogger.Errorf("Invalid number of minimum-arguments provided for transaction.")
		return shim.Error("{\"error\":\"Invalid number of minimum-arguments provided for transaction.\"}")
	}

	msisdn := args[0]
	msisdn = strings.TrimSpace(msisdn)

	//setting default value 2 (approved)
	sts := _ConsentApprovedStatus
	if len(args) == 2 && len(args[1]) > 0 {
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

	_consentLogger.Info("Within RevokeActiveConsentsByMsisdn")

	_, args := stub.GetFunctionAndParameters()
	if len(args) < 2 {
		_consentLogger.Errorf("Invalid number of minimum-arguments provided for transaction.")
		return shim.Error("{\"error\":\"Invalid number of minimum-arguments provided for transaction.\"}")
	}
	msisdn := args[0]
	msisdn = strings.TrimSpace(msisdn)

	var updateTs = ""
	if len(args[1]) > 1 {
		if isValid, errMsg := isValidDate(args[1]); !isValid {
			_consentLogger.Infof("Invalid Update Timestamp to modify the consent . :", errMsg)
			return shim.Error("{\"error\":\"Invalid Update Timestamp to modify the consent ." + errMsg + "\"}")
		}
		updateTs = args[1]
	}

	//setting default value 3 (Revoked)
	sts := _ConsentRevokedStatus
	if len(args) == 3 && len(args[2]) > 0 {
		if isValid, errMsg := isValidStatus(args[2]); !isValid {
			_consentLogger.Infof("Invalid status as Input. :", errMsg)
			return shim.Error("{\"error\":\"Invalid status as Input." + errMsg + "\"}")
		}
		sts = args[2]
	}

	//Active consents - status 2 ( Approved)
	activeSts := "2"

	consents := cm.getConsentsByPhoneNumber(stub, msisdn, activeSts)

	//Success Consents Message
	sConsents := make([]SuccessData, 0)
	//Failed Consesnts Message
	fConsents := make([]ErrorData, 0)

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
			e := ErrorData{ID: eachConsent.ConsentID, Msg: _Format3}
			fConsents = append(fConsents, e)
			continue
		}

		p := EventPayLoad{consent: eachConsent, txnID: stub.GetTxID()}

		payloadbytes, _ := json.Marshal(p)

		retErr := stub.SetEvent(_CreateEvent, payloadbytes)

		if retErr != nil {
			_consentLogger.Errorf(_Format5, _CreateEvent)
		}

		resultData := SuccessData{TrxnID: stub.GetTxID(), ConsID: eachConsent.ConsentID, Message: "Consent Revoke Successful", ConsentDets: eachConsent}

		sConsents = append(sConsents, resultData)

	}

	totalResponse := TotalResponse{SuccesConsents: sConsents, FailedConsents: fConsents}

	respJSON, _ := json.Marshal(totalResponse)

	return shim.Success(respJSON)
}

//UpdateConsentPurposeByIDs changes the Purpose of the Consents based upon the given ConsentIds as passed in the args[o]
//args[0] string array of URN( consentIDs) for which the purpose
//args[1] purpose to be changed with. Can be either 1/2/3
//args[2] updateTs is the time to update the consent records
//Returned payload contains two blocks - 'failedData' and 'successData'. FailedData contains an array of map with Falure details and SuccessData contains an array of map with details of Successfully Saved Consent.
//--FailedData block is as below:
//"errormsg": <Reason for Failure>,
//"id"		: <URN>
//--SuccessData block is as below:
//"consId": <URN>,
//"consentDets": <saved consent>,
//"message": "Update Consent Purpose Successful",
//"trxnId": <transactionid>
func (cm *ConsentManager) UpdateConsentPurposeByIDs(stub shim.ChaincodeStubInterface) pb.Response {

	_consentLogger.Info("Within UpdateConsentPurposeByIDs")
	_, args := stub.GetFunctionAndParameters()

	if len(args) < 3 {
		_consentLogger.Infof("Invalid No of arguments provided")
		return shim.Error("Invalid No of arguments provided")
	}

	arr := make([]string, 0)
	err := json.Unmarshal([]byte(args[0]), &arr)
	if err != nil {
		_consentLogger.Errorf("Invalid json provided as input.")
		return shim.Error("{\"error\":\"Invalid json provided as input\"}")
	}

	newPurpose := args[1]
	if isValid, errMsg := isValidPurpose(newPurpose); !isValid {
		_consentLogger.Infof("Invalid Purpose to modify the consent. :", errMsg)
		return shim.Error("{\"error\":\"Invalid Purpose to modify the consent." + errMsg + "\"}")
	}

	newUpdatedTS := args[2]
	if isValid, errMsg := isValidDate(newUpdatedTS); !isValid {
		_consentLogger.Infof("Invalid Update TS to modify the consent. :", errMsg)
		return shim.Error("{\"error\":\"Invalid Update TS to modify the consent." + errMsg + "\"}")
	}

	//returnStatus := make([]map[string]interface{}, 0)

	//Success Consents Message
	sConsents := make([]SuccessData, 0)
	//Failed Consesnts Message
	fConsents := make([]ErrorData, 0)

	_, updatedBy := cm.getInvokerIdentity(stub)

	for _, searchConsentID := range arr {

		existingRec, err1 := stub.GetState(searchConsentID)

		if err1 != nil {
			_consentLogger.Errorf("Error while fetching Consent with id ", searchConsentID)

			e := ErrorData{ID: searchConsentID, Msg: "Error while fetching Consent with id"}
			fConsents = append(fConsents, e)
			continue
		}

		if len(existingRec) == 0 {
			_consentLogger.Errorf("Consent does not exist with id ", searchConsentID)

			e := ErrorData{ID: searchConsentID, Msg: _Format6}
			fConsents = append(fConsents, e)
			continue
		}

		var updatedStatusConsent Consentdetails

		err := json.Unmarshal(existingRec, &updatedStatusConsent)
		if err != nil {
			e := ErrorData{ID: searchConsentID, Msg: _Format8}
			fConsents = append(fConsents, e)
			continue
		}

		updatedStatusConsent.Purpose = newPurpose
		updatedStatusConsent.UpdateTs = newUpdatedTS
		updatedStatusConsent.UpdatedBy = updatedBy

		marshalConsentJSON, _ := json.Marshal(updatedStatusConsent)

		finalErr := stub.PutState(updatedStatusConsent.ConsentID, marshalConsentJSON)

		if finalErr != nil {
			_consentLogger.Errorf("Unable to save with consent id " + updatedStatusConsent.ConsentID)

			e := ErrorData{ID: updatedStatusConsent.ConsentID, Msg: "Unable to save with consent id"}
			fConsents = append(fConsents, e)
		}

		p := EventPayLoad{consent: updatedStatusConsent, txnID: stub.GetTxID()}

		payloadbytes, _ := json.Marshal(p)

		retErr := stub.SetEvent(_UpdateEvent, payloadbytes)

		if retErr != nil {
			_consentLogger.Errorf("Event not generated for event : UPDATE_CONSENT")
		}

		resultData := SuccessData{TrxnID: stub.GetTxID(), ConsID: updatedStatusConsent.ConsentID, Message: "Update Consent Purpose Successful", ConsentDets: updatedStatusConsent}

		sConsents = append(sConsents, resultData)

	}

	totalResponse := TotalResponse{SuccesConsents: sConsents, FailedConsents: fConsents}

	respJSON, _ := json.Marshal(totalResponse)

	return shim.Success(respJSON)
}

//UpdateConsentPurposeByHeader updates the Purpose of the consents only for the active cosenst ( status = 2) based upon the given inputs
//Method can be used for bulk update of consent records
//args[0] should be an array of JSON objects of below structure (Keys are important to follow)
//"msisdn" : msisdn for which record should be fetched
//"cli"	 : Header (cli) ; combination of cli and msisdn should return a unique record from ledger
//args[1] : Purpose to be updated with
//args[2] : UpdateTimeStamp
//Returned payload contains two blocks - 'failedData' and 'successData'. FailedData contains an array of map with Falure details and SuccessData contains an array of map with details of Successfully Saved Consent.
//--FailedData block is as below:
//"errormsg": <Reason for Failure>,
//"id"		: <URN>
//--SuccessData block is as below:
//"consId": <URN>,
//"consentDets": <saved consent>,
//"message": "Update Consent Purpose Successful",
//"trxnId": <transactionid>
func (cm *ConsentManager) UpdateConsentPurposeByHeader(stub shim.ChaincodeStubInterface) pb.Response {

	_consentLogger.Info("Within UpdateConsentPurposeByHeader")

	searchCriteriaArr := make([]map[string]string, 0)
	_, args := stub.GetFunctionAndParameters()

	if len(args) < 3 {
		_consentLogger.Errorf("Invalid number of arguments provided for transaction.")
		return shim.Error("{\"error\":\"Invalid number of arguments provided for transaction\"}")
	}
	err := json.Unmarshal([]byte(args[0]), &searchCriteriaArr)

	if err != nil {
		_consentLogger.Errorf("Invalid json provided as input.")
		return shim.Error("{\"error\":\"Invalid json provided as input\"}")
	}

	// purpose : target purpose for all the element for the query result
	purpose := args[1] // can be either 1/2/3

	if isValid, errMsg := isValidPurpose(purpose); !isValid {
		_consentLogger.Infof("Invalid Purpose to update the consent. :", errMsg)
		return shim.Error("{\"error\":\"Invalid Purpose to update the consent." + errMsg + "\"}")
	}

	newUpdatedTS := args[2]
	if isValid, errMsg := isValidDate(newUpdatedTS); !isValid {
		_consentLogger.Infof("Invalid Update TS to modify the consent. :", errMsg)
		return shim.Error("{\"error\":\"Invalid Update TS to modify the consent." + errMsg + "\"}")
	}

	//returnResult := make([]map[string]interface{}, 0)
	//Success Consents Message
	sConsents := make([]SuccessData, 0)
	//Failed Consesnts Message
	fConsents := make([]ErrorData, 0)

	_, updatedBy := cm.getInvokerIdentity(stub)

	for _, searchCriteria := range searchCriteriaArr {

		msisdn := searchCriteria["msisdn"]
		msisdn = strings.TrimSpace(msisdn)
		cli := searchCriteria["cli"]
		cli = strings.TrimSpace(cli)

		sts := _ConsentApprovedStatus

		consents := cm.getConsentsByMsisdnCliStatus(stub, msisdn, cli, sts)

		if len(consents) == 0 {
			_consentLogger.Infof(_Format6)
			e := ErrorData{ID: "msisdn:" + msisdn + ",cli:" + cli, Msg: _Format6}
			fConsents = append(fConsents, e)
			continue
		}

		if len(consents) > 1 {
			_consentLogger.Infof(_Format7)
			e := ErrorData{ID: "msisdn:" + msisdn + ",cli:" + cli, Msg: _Format7}
			fConsents = append(fConsents, e)
			continue
		}

		singleConsent := consents[0]

		singleConsent.Purpose = purpose
		singleConsent.UpdateTs = newUpdatedTS

		singleConsent.UpdatedBy = updatedBy

		marshalConsentJSON, _ := json.Marshal(singleConsent)

		finalErr := stub.PutState(singleConsent.ConsentID, marshalConsentJSON)

		if finalErr != nil {
			_consentLogger.Infof("Unable to save with consent : %v", finalErr)
			e := ErrorData{ID: singleConsent.ConsentID, Msg: _Format2}
			fConsents = append(fConsents, e)
			continue
		}

		p := EventPayLoad{consent: singleConsent, txnID: stub.GetTxID()}

		payloadbytes, _ := json.Marshal(p)

		retErr := stub.SetEvent(_UpdateEvent, payloadbytes)

		if retErr != nil {
			_consentLogger.Errorf("Event not generated for event : UPDATE_CONSENT")
		}

		//make the payload to return and pass it through the shim.success
		resultData := SuccessData{TrxnID: stub.GetTxID(), ConsID: singleConsent.ConsentID, Message: "Update Consent Purpose Successful", ConsentDets: singleConsent}

		sConsents = append(sConsents, resultData)

	}

	totalResponse := TotalResponse{SuccesConsents: sConsents, FailedConsents: fConsents}

	respJSON, _ := json.Marshal(totalResponse)

	return shim.Success(respJSON)
}

//GetHistoryByKey queries the ledger using the given key.
//args[0] takes the key for search input
//It retrieve all the changes to the value happened over time as input given, across time
func (cm *ConsentManager) GetHistoryByKey(stub shim.ChaincodeStubInterface) peer.Response {

	_consentLogger.Debug("getHistoryByKey is being called.")

	_, args := stub.GetFunctionAndParameters()

	// Essential check to verify number of arguments
	if len(args) < 1 {
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

//QueryConsentsWithPagination uses a query string, page size and a bookmark to perform a query
//for Consesnts. Query string matching state database syntax is passed in and executed as-is.
//The number of fetched records would be equal to or lesser than the specified page size.
//Supports ad hoc queries that can be defined at runtime by the client.
//This supports state databases that support rich query (e.g. CouchDB)
//Paginated queries are only valid for read only transactions.
func (cm *ConsentManager) QueryConsentsWithPagination(stub shim.ChaincodeStubInterface) pb.Response {

	_, args := stub.GetFunctionAndParameters()

	var jsonResp string

	if len(args) != 3 {
		_consentLogger.Errorf("QueryConsentsWithPagination:Invalid number of arguments provided for transaction")
		jsonResp = "{\"error\":\"Invalid Number of argumnets provided for transaction\"}"
		return shim.Error(jsonResp)
	}
	var records []Consentdetails
	queryString := args[0]
	pageSize, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		_consentLogger.Errorf("QueryConsentsWithPagination:Error while ParseInt is :" + string(err.Error()))
		jsonResp = "{\"error\":\"PageSize ParseInt error- \"" + string(err.Error()) + "\"}"
		return shim.Error(jsonResp)
	}
	bookmark := args[2]
	resultsIterator, responseMetaData, err := stub.GetQueryResultWithPagination(queryString, int32(pageSize), bookmark)
	if err != nil {
		_consentLogger.Errorf("queryTemplateWithPagination:GetQueryResultWithPagination is Failed :" + string(err.Error()))
		jsonResp = "{\"error\":\"GetQueryResultWithPagination is Failed- \"" + string(err.Error()) + "\"}"
		return shim.Error(jsonResp)
	}

	for resultsIterator.HasNext() {
		record := Consentdetails{}
		recordBytes, _ := resultsIterator.Next()
		if string(recordBytes.Value) == "" {
			continue
		}
		err := json.Unmarshal(recordBytes.Value, &record)
		if err != nil {
			_consentLogger.Errorf("QueryConsentsWithPagination:Unable to unmarshal Consent retrieved :" + string(err.Error()))
			jsonResp = "{\"error\":\"Unable to unmarshal Consent retrieved- \"" + string(err.Error()) + "\"}"
			return shim.Error(jsonResp)
		}
		records = append(records, record)
	}

	resultData := map[string]interface{}{
		"status":       "true",
		"consents":     records,
		"recordscount": responseMetaData.FetchedRecordsCount,
		"bookmark":     responseMetaData.Bookmark,
	}
	respJSON, _ := json.Marshal(resultData)

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

	consents := cm.retrieveConsentRecords(stub, fmt.Sprintf(consentSearchCriteria, msisdn, cli), "consentSearchByHeaderMsisdn")

	return consents
}

//getConsentsByMsisdnCliStatus returns the consents upon the given MSISDN, Cli (header) and status
func (cm *ConsentManager) getConsentsByMsisdnCliStatus(stub shim.ChaincodeStubInterface, msisdn, cli, sts string) []Consentdetails {
	consentSearchCriteria := `{
		"obj":"Consent"	,
		"msisdn":"%s",
		"cli":"%s",
		"sts":"%s"	
	}`

	consents := cm.retrieveConsentRecords(stub, fmt.Sprintf(consentSearchCriteria, msisdn, cli, sts), "consentSearchByHeaderMsisdnSts")

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

func hasElem(s interface{}, elem interface{}) bool {
	arrV := reflect.ValueOf(s)

	if arrV.Kind() == reflect.Slice {
		for i := 0; i < arrV.Len(); i++ {
			if arrV.Index(i).Interface() == elem {
				return true
			}
		}
	}
	return false
}
