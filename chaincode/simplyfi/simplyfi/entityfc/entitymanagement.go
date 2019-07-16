package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	id "github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
	"github.com/hyperledger/fabric/protos/peer"
)

var _entityLogger = shim.NewLogger("EntityManager")

const _CreateEvent = "CREATE_ENTITY"
const _ModifyEvent = "MODIFY_ENTITY"

//Entity structure defines the ledger record for any entity
type Entity struct {
	ObjType              string `json:"obj"`    //DocType  -- search key
	RegReqestID          string `json:"reqid"`  //
	EntityID             string `json:"id"`     //EntityID -- Key field - autogenerated in backend
	EntityType           string `json:"etype"`  // GOVT OR PRIVATE OR SEBI OR KNOWN BRAND
	POI                  string `json:"poi"`    // TAN OR PAN NO OF ENTITY Mandatory for Private
	EntityName           string `json:"name"`   //EntityName -- search Key
	EntityClassification string `json:"eclass"` //PE or TM
	ServiceProvider      string `json:"svcprv"` //AccessProvidedID
	Status               string `json:"sts"`    //Status
	ApprovedOn           string `json:"appon"`
	ApprovedBy           string `json:"appby"`
	Creator              string `json:"crtr"` //CreatedBy
	UpdateTs             string `json:"uts"`  //UpdatedTs - autogenerated in backend
	CreateTs             string `json:"cts"`  //CreatedTs - autogenerated in backend
	UpdatedBy            string `json:"uby"`  //UpdatedBy
}

//EntityManager manages entity transactions
type EntityManager struct {
}

var errKey, errorDetails, jsonResp, repError string
var entityType = map[string]bool{
	"P": true,
	"G": true,
	"S": true,
	"K": true,
	"U": true,
	"O": true,
}

var entityStatus = map[string]bool{
	"A": true,
	"I": true,
	"B": true,
}

var serviceProvider = map[string]bool{
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
var validCategoryMap = map[string]bool{
	"PE": true,
	"TM": true,
}

func validEnumEntry(input string, enumMap map[string]bool) bool {
	if _, isEntryExists := enumMap[input]; !isEntryExists {
		return false
	}
	return true
}

//IsValidEntityIDPresent checks for  validity of entity for modify trxn
func IsValidEntityIDPresent(e Entity) (bool, string) {

	if len(e.EntityID) == 0 {
		return false, "Entity Id should be present there"
	}
	if len(e.EntityName) == 0 {
		return false, "Entity name is mandatory"
	}
	if len(e.UpdateTs) == 0 {
		return false, "UpdatedTS is mandatory"
	}
	if len(e.CreateTs) == 0 {
		return false, "CreateTS is mandatory"
	}
	if len(e.POI) == 0 {
		return false, "PAN No is mandatory"
	}
	if len(e.ApprovedOn) == 0 {
		return false, "ApprovedOn is mandatory"
	}
	if len(e.ApprovedBy) == 0 {
		return false, "ApprovedBy is mandatory"
	}
	if !validEnumEntry(e.EntityType, entityType) {
		return false, "Entity Type: Enter either P, G, S, K, U, O"
	}
	if !validEnumEntry(e.ServiceProvider, serviceProvider) {
		return false, "Service Provider: Enter either AI, VO, ID, BL, ML, QL, TA, JI, VI"
	}
	if !validEnumEntry(e.EntityClassification, validCategoryMap) {
		return false, "Entity Classification: Enter either PE, TM"
	}
	if !validEnumEntry(e.Status, entityStatus) {
		return false, "Status : Enter either A, I, B"
	}
	return true, ""
}

//SearchEntity searchs for entity based on the input parameters
func (em *EntityManager) searchEntity(stub shim.ChaincodeStubInterface) peer.Response {
	var response peer.Response
	searchCriteria := make(map[string]string)
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("searchEntity: " + jsonResp)
		return shim.Error(jsonResp)
	}
	err := json.Unmarshal([]byte(args[0]), &searchCriteria)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON provided- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("searchEntity: " + jsonResp)
		return shim.Error(jsonResp)
	}
	searchType, isOk := searchCriteria["typ"]
	if !isOk {
		return shim.Error("Search type not provided")
	}
	switch searchType {
	case "name":
		entitySearchCriteria := `{
			"obj":"Entity"	,
			"name":"%s"
		}`
		entityName := searchCriteria[searchType]
		entities := em.retriveEntityRecords(stub, fmt.Sprintf(entitySearchCriteria, entityName), "entitySearchByName")
		recordsJSON, marshalErr := json.Marshal(entities)
		if marshalErr != nil {
			repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
			errorDetails = "Cannot Marshal the JSON- " + repError
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_entityLogger.Errorf("searchEntity by Name: " + jsonResp)
			response = shim.Error(jsonResp)
		}
		response = shim.Success(recordsJSON)
	case "id":
		entitySearchCriteria := `{
			"obj":"Entity"	,
			"id":"%s"
		}`
		entityID := searchCriteria[searchType]
		entities := em.retriveEntityRecords(stub, fmt.Sprintf(entitySearchCriteria, entityID), "entitySearchByID")
		recordsJSON, marshalErr := json.Marshal(entities)
		if marshalErr != nil {
			repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
			errorDetails = "Cannot Marshal the JSON- " + repError
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_entityLogger.Errorf("searchEntity by ID: " + jsonResp)
			response = shim.Error(jsonResp)
		}
		response = shim.Success(recordsJSON)
	case "svcprv":
		entitySearchCriteria := `{
			"obj":"Entity"	,
			"svcprv":"%s"
		}`
		entityID := searchCriteria[searchType]
		entities := em.retriveEntityRecords(stub, fmt.Sprintf(entitySearchCriteria, entityID), "entitySearchByAP")
		recordsJSON, marshalErr := json.Marshal(entities)
		if marshalErr != nil {
			repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
			errorDetails = "Cannot Marshal the JSON- " + repError
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_entityLogger.Errorf("searchEntity by ServiceProvider: " + jsonResp)
			response = shim.Error(jsonResp)
		}
		response = shim.Success(recordsJSON)
	case "poi":
		entitySearchCriteria := `{
			"obj":"Entity"	,
			"poi":"%s"
		}`
		entityID := searchCriteria[searchType]
		entities := em.retriveEntityRecords(stub, fmt.Sprintf(entitySearchCriteria, entityID), "entitySearchByPoi")
		recordsJSON, marshalErr := json.Marshal(entities)
		if marshalErr != nil {
			repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
			errorDetails = "Cannot Marshal the JSON- " + repError
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_entityLogger.Errorf("searchEntity by POI: " + jsonResp)
			response = shim.Error(jsonResp)
		}
		response = shim.Success(recordsJSON)
	default:
		errKey = searchType
		errorDetails = "Unsupported Search type Provided"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("searchEntity: " + jsonResp)
		response = shim.Error(jsonResp)
	}
	return response
}

//CreateEntity creates an entity in the ledger
func (em *EntityManager) createEntity(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	var entityToSave Entity
	//checking for the length of the input
	if len(args) < 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("createEntity: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the input to the entityToSave object
	err := json.Unmarshal([]byte(args[0]), &entityToSave)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON provided- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("createEntity: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//checking whether the entity already exists or not against entityId
	if recordBytes, _ := stub.GetState(entityToSave.EntityID); len(recordBytes) > 0 {
		errKey = entityToSave.EntityID
		errorDetails = "Entityid already registered. Provide an unique entity id"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("createEntity: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the entity with the details provided as input
	entityToSave.ObjType = "Entity"
	_, creator := em.getInvokerIdentity(stub)
	entityToSave.Creator = creator
	entityToSave.UpdatedBy = creator
	//marshalling the data to store into the ledger
	entityJSON, marshalErr := json.Marshal(entityToSave)
	if marshalErr != nil {
		repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("createEntity: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//checking for the validity of the entity values before storing into the ledger
	if isValid, errMsg := IsValidEntityIDPresent(entityToSave); !isValid {
		errKey = string(entityJSON)
		errorDetails = errMsg
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("createEntity: " + jsonResp)
		return shim.Error(jsonResp)
	}
	_entityLogger.Info("Saving Entity to the ledger with entityID--------", entityToSave.EntityID)
	//storing the entity into the ledger
	err = stub.PutState(entityToSave.EntityID, entityJSON)
	if err != nil {
		errKey = entityToSave.EntityID
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save entity with entity id- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("createEntity: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//setting an event after writing into the ledger
	retErr := stub.SetEvent(_CreateEvent, entityJSON)
	if retErr != nil {
		errKey = string(entityJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : CREATE_ENTITY- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("createEvent: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the response and return to the app layer
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"entityID": entityToSave.EntityID,
		"message":  "Entity Created Successfully",
		"entity":   entityToSave,
		"status":   "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//function used to retrieve entities on entity id using rich queries
func (em *EntityManager) retriveEntityRecords(stub shim.ChaincodeStubInterface, criteria string, indexs ...string) []Entity {
	var finalSelector string
	records := make([]Entity, 0)
	if len(indexs) == 0 {
		finalSelector = fmt.Sprintf("{\"selector\":%s }", criteria)
	} else {
		finalSelector = fmt.Sprintf("{\"selector\":%s , \"use_index\" :\"%s\" }", criteria, indexs[0])
	}
	_entityLogger.Infof("Query Selector : %s", finalSelector)
	resultsIterator, _ := stub.GetQueryResult(finalSelector)
	for resultsIterator.HasNext() {
		record := Entity{}
		recordBytes, _ := resultsIterator.Next()
		err := json.Unmarshal(recordBytes.Value, &record)
		if err != nil {
			errKey = string(recordBytes.Value)
			repError = strings.Replace(err.Error(), "\"", " ", -1)
			errorDetails = "Invalid JSON for unMarshalling- " + repError
			jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
			_entityLogger.Errorf("modifyEntity: " + jsonResp)
		}
		records = append(records, record)
	}
	return records
}

//function used to get the history of the entity against entityId from the ledger
func (em *EntityManager) getHistoryByEntity(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	// checking the length of the input
	if len(args) != 1 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("getHistoryByEntity: " + jsonResp)
		return shim.Error(jsonResp)
	}
	key := args[0]
	resultsIterator, err := stub.GetHistoryForKey(key)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Could not fetch history- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("getHistoryByEntity: " + jsonResp)
		return shim.Error(jsonResp)
	}
	defer resultsIterator.Close()
	historicResponse := make([]map[string]interface{}, 0)
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			repError = strings.Replace(err.Error(), "\"", " ", -1)
			errorDetails = "Iteration error- " + repError
			jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
			_entityLogger.Errorf("getHistoryByEntity: " + jsonResp)
			return shim.Error(jsonResp)
		}
		value := make(map[string]interface{})
		json.Unmarshal(response.Value, &value)
		historicResponse = append(historicResponse, map[string]interface{}{"txId": response.TxId, "value": value, "status": "true"})
	}
	respJSON, _ := json.Marshal(historicResponse)
	return shim.Success(respJSON)
}

//UpdateEntityStatus updates the status of the entity against entityId
func (em *EntityManager) updateEntityStatus(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	var existingEntity Entity
	// checking the length of the input
	if len(args) < 3 {
		errKey = strconv.Itoa(len(args))
		errorDetails = "Invalid Number of Arguments"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("updateEntityStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	searchEntityID := args[0]
	newStatus := args[1]
	newUpdatedTS := args[2]
	//getting the entity data from the ledger against entityId
	existingEntityResult, err := stub.GetState(searchEntityID)
	if err != nil {
		errKey = searchEntityID
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Could not fetch details for the Entity- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("updateEntityStatus: " + jsonResp)
		return shim.Error(jsonResp)
	} else if existingEntityResult == nil {
		errKey = searchEntityID
		errorDetails = "Entity does not exist with EntityId"
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("updateEntityStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//unmarshalling the data retrieved from the ledger to existingEntity object for update
	err = json.Unmarshal([]byte(existingEntityResult), &existingEntity)
	if err != nil {
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Invalid JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("updateEntityStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the data to store in the ledger
	existingEntity.Status = newStatus
	existingEntity.UpdateTs = newUpdatedTS
	_, updatedBy := em.getInvokerIdentity(stub)
	existingEntity.UpdatedBy = updatedBy
	//marshalling the data to store in the ledger
	entityJSON, marshalErr := json.Marshal(existingEntity)
	if marshalErr != nil {
		repError = strings.Replace(marshalErr.Error(), "\"", " ", -1)
		errorDetails = "Cannot Marshal the JSON- " + repError
		jsonResp = "{\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("updateEntityStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//checking for the validity of the values before storing into the ledger
	if isValid, errMsg := IsValidEntityIDPresent(existingEntity); !isValid {
		errKey = string(entityJSON)
		errorDetails = errMsg
		jsonResp = "{\"Data\":\"" + errKey + "\",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("updateStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//storing the entity into the ledger
	err = stub.PutState(existingEntity.EntityID, entityJSON)
	if err != nil {
		errKey = string(entityJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Unable to save entity with entityId- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("updateEntityStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//setting an event after storing into the ledger
	retErr := stub.SetEvent(_ModifyEvent, entityJSON)
	if retErr != nil {
		errKey = string(entityJSON)
		repError = strings.Replace(err.Error(), "\"", " ", -1)
		errorDetails = "Event not generated for event : MODIFY_EVENT- " + repError
		jsonResp = "{\"Data\":" + errKey + ",\"ErrorDetails\":\"" + errorDetails + "\"}"
		_entityLogger.Errorf("updateEntityStatus: " + jsonResp)
		return shim.Error(jsonResp)
	}
	//packaging the response and return to the app layer
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"entityID": existingEntity.EntityID,
		"message":  "Update status successful",
		"entity":   existingEntity,
		"status":   "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//function used for getting identity of the invoker
func (em *EntityManager) getInvokerIdentity(stub shim.ChaincodeStubInterface) (bool, string) {
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