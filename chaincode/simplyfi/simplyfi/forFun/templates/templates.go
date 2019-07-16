/*
This Chaincode is written for storing,retrieving,deleting the Content/Consent Templates that are stored in DLT
*/

package main

import (
	"encoding/json" //reading and writing JSON
	"fmt"
	"strconv"
	"strings"
	"bytes"
	"github.com/hyperledger/fabric/core/chaincode/shim"             // import for Chaincode Interface
	id "github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
	pb "github.com/hyperledger/fabric/protos/peer"                  // import for peer response

)

//Logger for Logging
var _templateLogger = shim.NewLogger("Templates")
//Event Names
const _AddTemplate = "ADD_TEMPLATE"
const _UpdateTemplate = "UPDATE_TEMPLATE"
const _BulkTemplate = "BULK_TEMPLATE"

//template struct and variables to store in the dlt
type Template struct {
	ObjType             string   `json:"obj"`
	TemplateID          string   `json:"urn"`
	PEID                string   `json:"peid"`
	CLI                 []string `json:"cli"`
	TemplateName        string   `json:"tname"`
	TemplateType        string   `json:"ttyp"`
	CommunicationType   string   `json:"ctyp"`
	ConsentTemplateType string   `json:"csty"`
	Contenttype         string   `json:"coty"`
	NoOfVariables       string   `json:"vars"`
	Category            string   `json:"ctgr"`
	TempContent         string   `json:"tcont"`
	TMID                string   `json:"tmid"`
	Creator             string   `json:"crtr"`
	CreateTs            string   `json:"cts"`
	UpdatedBy           string   `json:"uby"`
	UpdateTs            string   `json:"uts"`
	Status              string   `json:"sts"`
}
//TemplateManager manages Template related transactions in the ledger
type TemplateManager struct {
}

var errorDetails, errKey, jsonResp, repError string
//Template Type
var tempType = map[string]bool{
	"CS": true,
	"CT": true,
}

//CommunicationType
var communicationType = map[string]bool{
	"P":  true,
	"T":  true,
	"SE": true,
	"SI": true,
}

//content Type
var contentType = map[string]bool{
	"T": true,
	"U": true,
}

//Status
var status = map[string]bool{
	"A": true,
	"I": true,
}

var category= map[string]bool {
	"1":true,
	"2":true,
	"3":true,
	"4":true,
	"5":true,
	"6":true,
	"7":true,
	"8":true,
}

//consentTemplate Type
var consentTemplateType = map[string]bool{
	"1": true,
	"2": true,
	"3": true,
}

//Output Structure for the output response
type ErrorDetails struct {
	Data         string `json:"data"`
	Details 	 string `json:"error"`
}

//checks whether field is present or not in map
func validEnumEntry(input string, enumMap map[string]bool) bool {
	if _, isEntryExists := enumMap[input]; !isEntryExists {
		return false
	}
	return true
}

//IsValidTemplatePresent checks for  validity of Template transaction before recording in the ledger
func IsValidTemplatePresent(s Template) (bool, string) {
	if len(s.TemplateID) == 0 {
		return false, "TemplateID is mandatory"
	}
	if len(s.PEID) == 0 {
		return false, "Principal EntityID is mandatory"
	}
	if len(s.CLI) == 0 {
		return false, "CLI is mandatory"
	}
	if len(s.TemplateName) == 0 {
		return false, "TemplateName is mandatory"
	}
	if !validEnumEntry(s.TemplateType, tempType){
		return false, "Template Type: Either CT or CS"
	}
	if s.TemplateType == "CT"{
		if !validEnumEntry(s.Contenttype, contentType){
			return false, "Content Type: Either T or U"
		}
		if len(s.NoOfVariables) == 0{
			return false, "NoOfVariables is mandatory"
		}else if _,err:= strconv.Atoi(s.NoOfVariables); err != nil {
			return false, "Vars should be numeric"
		}
		if !validEnumEntry(s.Category, category){
			return false, "Category: Either 1, 2, 3, 4, 5, 6, 7 or 8"
		}
		if !validEnumEntry(s.CommunicationType, communicationType){
			return false, "Communication Type: Either P, T, SE or SI"
		}

	}else if s.TemplateType == "CS"{
		if !validEnumEntry(s.ConsentTemplateType, consentTemplateType){
			return false, "Consent Template Type: 1, 2 or 3"
		}
		if s.CommunicationType != "SE" {
			return false, "Communication Type : SE"
		}
	}
	if len(s.TempContent) == 0 {
		return false, "Template Content is mandatory"
	}
	if len(s.Creator) == 0 {
		return false, "Creator is mandatory"
	}
	if len(s.CreateTs) == 0 {
		return false, "CreateTs is mandatory"
	}
	if len(s.UpdatedBy) == 0 {
		return false, "UpdatedBy is mandatory"
	}
	if len(s.UpdateTs) == 0 {
		return false, "UpdateTs is mandatory"
	}
	if !validEnumEntry(s.Status, status) {
		return false, "Status: Either A or I"
	}
	return true, ""
}
//setTemplate is a function to record the template into the ledger
func (s *TemplateManager) setTemplate(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	var templateToSave Template
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("setTemplate: " + jsonResp)
		return shim.Error(jsonResp)
	}
	err := json.Unmarshal([]byte(args[0]), &templateToSave)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON provided- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("setTemplate: " + jsonResp)
		return shim.Error(jsonResp)
	}
	if recordBytes, _ := stub.GetState(templateToSave.TemplateID); len(recordBytes) > 0 {
		errKey = templateToSave.TemplateID
		errorDetails = "Template with this URN already exist, provide unique URN"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("setTemplate: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the template with the details provided as input
	templateToSave.ObjType = "Template"
	_, creator := s.getInvokerIdentity(stub)
	templateToSave.Creator = creator
	templateToSave.UpdatedBy = creator
	templateToSave.UpdateTs = templateToSave.CreateTs
	if templateToSave.TemplateType == "CT"{
		templateToSave.ConsentTemplateType=""
	}else if templateToSave.TemplateType == "CS"{
		templateToSave.Category=""
		templateToSave.NoOfVariables=""
		templateToSave.Contenttype=""
	}
	//marshalling the data for storing into the ledger
	templateJSON, marshalErr := json.Marshal(templateToSave)
	if marshalErr != nil {
		repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("setTemplate: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//checking for the validity of the template values before storing into the ledger
	if isValid, errMsg := IsValidTemplatePresent(templateToSave); !isValid {
		errorDetails = errMsg
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("setTemplate: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//storing the header into the ledger
	_templateLogger.Info("Saving Template to the ledger with id----------", templateToSave.TemplateID)
	err = stub.PutState(templateToSave.TemplateID, templateJSON)
	if err != nil {
		errKey = string(templateJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save template with URN- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("setTemplate: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//setting event after storing into the ledger
	retErr := stub.SetEvent(_AddTemplate,templateJSON)
	if retErr != nil {
		errKey = string(templateJSON)
		repError = strings.Replace(retErr.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : ADD_TEMPLATE- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("setTemplate: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the reponse and returning to the app layer
	resultData := map[string]interface{}{		
		"trxnID"	  : stub.GetTxID(),
		"peid"		  : templateToSave.PEID,
		"templateID"  : templateToSave.TemplateID,
		"templateName": templateToSave.TemplateName,
		"message"	  : "Template created Successfully",
		"status"	  : "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}
//queryTemplate using selector query
func (s *TemplateManager)queryTemplates(stub shim.ChaincodeStubInterface) pb.Response {
	_, args:= stub.GetFunctionAndParameters()
	if len(args) != 1{
		errKey= strconv.Itoa(len(args))
		errorDetails= "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("queryTemplates: " + jsonResp)
		return shim.Error(jsonResp)
	}
	queryString := args[0]
	resultData := s.retrieveTemplateRecords(stub, fmt.Sprintf(queryString))
	recordsJSON, _ := json.Marshal(resultData)
	return shim.Success(recordsJSON)
	
}
//addbatchtemplates will add the templates to the ledger in bulk
func (s *TemplateManager) addBatchTemplates(stub shim.ChaincodeStubInterface) pb.Response {
	_,args:= stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("addBatchTemplates: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var listTemplates []Template
	err := json.Unmarshal([]byte(args[0]), &listTemplates)
	if err != nil {
		repError= strings.Replace(err.Error(),"\""," ", -1)
		errorDetails = "Invalid JSON provided- "+ repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("addBatchTemplates: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var approvedTemplates []string
	rejectedTemplates:=  make([]ErrorDetails, 0)
	_, creator := s.getInvokerIdentity(stub)
	for i := 0; i < len(listTemplates); i++ {
		var templateToSave Template
		templateToSave = listTemplates[i]
		if recordBytes, _ := stub.GetState(templateToSave.TemplateID); len(recordBytes) > 0 {
			msg := "Template with this URN already Exist, provide unique URN(templateId)"
			errMsg:= ErrorDetails{Data: templateToSave.TemplateID,Details:msg}
			rejectedTemplates = append(rejectedTemplates, errMsg)
			continue
		}
		templateToSave.ObjType = "Template"
		templateToSave.Creator = creator
		templateToSave.UpdatedBy = creator
		templateToSave.UpdateTs = templateToSave.CreateTs
		if templateToSave.TemplateType == "CT"{
			templateToSave.ConsentTemplateType=""
		}else if templateToSave.TemplateType == "CS"{
			templateToSave.Category=""
			templateToSave.NoOfVariables=""
			templateToSave.Contenttype=""
		}
		//marshalling the data for storing into the ledger
		templateJSON, marshalErr := json.Marshal(templateToSave)
		if marshalErr != nil {
			repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
			errorDetails = "Cannot Marshal the JSON- " + repError
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_templateLogger.Errorf("setTemplate: " + jsonResp)
			errMsg:= ErrorDetails{Data: templateToSave.TemplateID,Details:errorDetails}
			rejectedTemplates = append(rejectedTemplates, errMsg)
			continue
		}
		//checking for the validity of the template values before storing into the ledger
		if isValid, errMsg := IsValidTemplatePresent(templateToSave); !isValid {
			errorDetails = errMsg
			jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_templateLogger.Errorf("setTemplate: " + jsonResp)
			errMsg:= ErrorDetails{Data: templateToSave.TemplateID,Details:errorDetails}
			rejectedTemplates = append(rejectedTemplates, errMsg)
			continue
		}
		//storing the header into the ledger
		_templateLogger.Info("Saving Template to the ledger with id----------", templateToSave.TemplateID)
		err = stub.PutState(templateToSave.TemplateID, templateJSON)
		if err != nil {
			errKey = string(templateJSON)
			repError = strings.Replace(err.Error(), "\"", " ", -1)
			errorDetails = "Unable to save template with URN- " + repError
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_templateLogger.Errorf("setTemplate: " + jsonResp)
			errMsg:= ErrorDetails{Data: templateToSave.TemplateID,Details:errorDetails}
			rejectedTemplates = append(rejectedTemplates, errMsg)
			continue
		
		}
		//setting event after storing into the ledger
		retErr := stub.SetEvent(_AddTemplate,templateJSON)
		if retErr != nil {
			errKey = string(templateJSON)
			repError = strings.Replace(retErr.Error(), "\"", " ", -1)
			errorDetails = "Event not generated for event : ADD_TEMPLATE- " + repError
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_templateLogger.Errorf("setTemplate: " + jsonResp)
			errMsg:= ErrorDetails{Data: templateToSave.TemplateID,Details:errorDetails}
			rejectedTemplates= append(rejectedTemplates, errMsg)
			continue
		}
		approvedTemplates = append(approvedTemplates, templateToSave.TemplateID)
	}
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"failed_urn": rejectedTemplates,
		"message":  "Batch Templates created successfully",
		"status": "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}
//updateTemplateStatus for Updating Template status in DL based on PE and Template Name on successful PE check
func (s *TemplateManager) updateTemplateStatus(stub shim.ChaincodeStubInterface) pb.Response {
	_, args:= stub.GetFunctionAndParameters()
	var jsonResp string
	if len(args) != 3 {
		errKey= strconv.Itoa(len(args))
		errorDetails= "Invalid number of arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("updateTemplateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	urn:= args[0]
	status:= args[1]
	uts:= args[2]
	templateRecord, err := stub.GetState(urn)
	if err != nil {
		errKey = urn
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to fetch the Template Record- " + repError
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("updateTemplateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	} else if templateRecord == nil {
		errKey = urn
		errorDetails = "Template details does not exist with URN"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("updateTemplateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var existingTemplate Template
	err = json.Unmarshal([]byte(templateRecord), &existingTemplate)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON for storing- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("updateTemplateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	_, creatorUpdatedBy := s.getInvokerIdentity(stub)
	existingTemplate.UpdateTs = uts
	existingTemplate.UpdatedBy = creatorUpdatedBy
	existingTemplate.Status = status
	templateJSON, marshalErr  := json.Marshal(existingTemplate)
	if marshalErr != nil {
		repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("updateTemplateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	if isValid, errMsg := IsValidTemplatePresent(existingTemplate); !isValid {
		errorDetails = errMsg
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("updateTemplateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	err = stub.PutState(urn, templateJSON)
	if err != nil {
		errKey = string(templateJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save Templates Record with URN- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("updateTemplateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	retErr := stub.SetEvent(_UpdateTemplate, templateJSON)
	if retErr != nil {
		errKey = string(templateJSON)
		repError = strings.Replace(retErr.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : UPDATE_SCRUB- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("updateTemplateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	resultData := map[string]interface{}{
		"trxnID":     stub.GetTxID(),
		"templateID": urn,
		"message":    "Template updated successfully",
		"status":  "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}
//this function is used to fetch the records from the dlt 
//To-Do: add the indexes to the function and work with indexes if they pass it as index
func (s *TemplateManager) retrieveTemplateRecords(stub shim.ChaincodeStubInterface, criteria string) []Template {
	var finalSelector string
	records := make([]Template, 0)
	finalSelector= criteria
	_templateLogger.Infof("Query Selector : %s", finalSelector)
	resultsIterator, _ := stub.GetQueryResult(finalSelector)
	for resultsIterator.HasNext() {
		record := Template{}
		recordBytes, _ := resultsIterator.Next()
		err := json.Unmarshal(recordBytes.Value, &record)
		if err != nil {
			repError= strings.Replace(err.Error(), "\""," ", -1)
			errorDetails= "Invalid JSON to marshal- "+ repError
			jsonResp= "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_templateLogger.Errorf("retrieveTemplateRecords: "+ jsonResp)
		}
		records = append(records, record)
	}
	return records
}
//getTemplateByTemplateID for Getting template data based on templateid
func (s *TemplateManager) getTemplateByTemplateID(stub shim.ChaincodeStubInterface) pb.Response {
	_, args:= stub.GetFunctionAndParameters()
	if len(args) != 1 {
		errKey= strconv.Itoa(len(args))
		errorDetails= "Invalid number of arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("getTemplateByTemplateID: " + jsonResp)
		return shim.Error(jsonResp)
	}
	urn := args[0]
	templateRecord, err := stub.GetState(urn)
	if err != nil {
		errKey = urn
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to fetch Template Details- " + repError
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("getTemplateByTemplateID: " + jsonResp)
		return shim.Error(jsonResp)
	} else if templateRecord == nil {
		errKey = urn
		errorDetails = "Template record does not exist with URN"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("getTemplateByTemplateID: " + jsonResp)
		return shim.Error(jsonResp)
	}
	record := Template{}
	err = json.Unmarshal(templateRecord, &record)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("getTemplateByTemplateID: " + jsonResp)
		return shim.Error(jsonResp)
	}
	resultData := map[string]interface{}{
		"data": record,
		"status": "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}
//getHistoryQuery for Getting all history data for urn
func (s *TemplateManager) queryTemplatesHistory(stub shim.ChaincodeStubInterface) pb.Response {
	_,args:= stub.GetFunctionAndParameters()
	if len(args) != 1 {
		errKey= strconv.Itoa(len(args))
		errorDetails= "Invalid number of arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("queryTemplateHistory: " + jsonResp)
		return shim.Error(jsonResp)
	}
	var records []Template
	urn := args[0] 
 	resultsIterator, err := stub.GetHistoryForKey(urn)
 	if err != nil {
		errKey= urn
		repError= strings.Replace(err.Error(), "\""," ", -1)
		errorDetails= "Get History for the key failed- "+ repError
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("queryTemplateHistory: " + jsonResp)
		return shim.Error(jsonResp) 
 	}
 	for resultsIterator.HasNext() {
 		record := Template{}
 		recordBytes, _ := resultsIterator.Next()
 		if string(recordBytes.Value) == "" {
 			continue
 		}
 		err := json.Unmarshal(recordBytes.Value, &record)
 		if err != nil {
			repError= strings.Replace(err.Error(), "\""," ", -1)
			errorDetails= "Unmarshalling Error- "+ repError
			jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_templateLogger.Errorf("queryTemplateHistory: " + jsonResp)
 			return shim.Error(jsonResp)
 		}
 		records = append(records, record)
 	}
 	resultData := map[string]interface{}{
 		"status":    "true",
 		"templates": records,
 	}
 	respJson, _ := json.Marshal(resultData)
 	return shim.Success(respJson)
}
func (s *TemplateManager) queryTemplatesWithPagination(stub shim.ChaincodeStubInterface) pb.Response {
	_,args := stub.GetFunctionAndParameters()
 	if len(args) != 3 {
		errKey= strconv.Itoa(len(args))
		errorDetails= "Invalid number of arguments"
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_templateLogger.Errorf("queryTemplateWithPagination: " + jsonResp)
		return shim.Error(jsonResp)
	}
	queryString := args[0]
 	pageSize,err1 := strconv.ParseInt(args[1], 10, 32)
 	if err1 != nil {
 		errKey = string(args[1])
 		errorDetails = "PageSize should be a Number"
 		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
 		_templateLogger.Errorf("queryTemplateWithPagination: " + jsonResp)
 		return shim.Error(jsonResp)
 	}
 	bookMark := args[2]
 	paginationResults, err2 := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookMark)
 	if err2 != nil {
 		errKey = queryString + "," + string(pageSize) + "," + bookMark
 		errorDetails = "Could not fetch the data"
 		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
 		_templateLogger.Errorf("queryTemplateWithPagination: " + jsonResp)
 		return shim.Error(jsonResp)
	 }
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

func addPaginationMetadataToQueryResults(buffer *bytes.Buffer, responseMetadata *pb.QueryResponseMetadata) *bytes.Buffer {
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
//function used to get the identity of the invoker
func (s *TemplateManager) getInvokerIdentity(stub shim.ChaincodeStubInterface) (bool, string) {
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

func (s *TemplateManager) getID(stub shim.ChaincodeStubInterface) pb.Response {
	//Following id comes in the format X509::<Subject>::<Issuer>>
	enCert, err := id.GetID(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	resultData := map[string]interface{}{
		"status":    "true",
		"id": enCert,
	}
	respJson, _ := json.Marshal(resultData)
	return shim.Success(respJson)
}