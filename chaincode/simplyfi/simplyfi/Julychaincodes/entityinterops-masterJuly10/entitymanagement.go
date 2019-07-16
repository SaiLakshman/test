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
	pb "github.com/hyperledger/fabric/protos/peer"
)

var _entityLogger = shim.NewLogger("EntityManager")

const _CreateEvent = "CREATE_ENTITY"
const _ModifyEvent = "MODIFY_ENTITY"
const _BlacklistEntity = "BLACKLIST_ENTITY"

//Status filed structure
type Status struct {
	Svcprv string `json:"svcprv"`
	Sts    string `json:"sts"`
}

var searchReturnObj string

//Entity structure defines the ledger record for any entity
type Entity struct {
	ObjType              string            `json:"obj"`    //DocType  -- search key
	EntityID             string            `json:"id"`     //EntityID -- Key field - autogenerated in backend
	EntityType           string            `json:"etype"`  // GOVT OR PRIVATE OR SEBI OR KNOWN BRAND
	POI                  string            `json:"poi"`    // TAN OR PAN NO OF ENTITY Mandatory for Private
	EntityName           string            `json:"name"`   //EntityName -- search Key
	EntityClassification string            `json:"eclass"` //PE or TM
	ServiceProvider      string            `json:"svcprv"` //AccessProvidedID
	Status               map[string]string `json:"sts"`    //status operator wise {"AI":"A","VO":"I"}
	ApprovedOn           string            `json:"appon"`
	Creator              string            `json:"crtr"` //CreatedBy
	UpdateTs             string            `json:"uts"`  //UpdatedTs - autogenerated in backend
	CreateTs             string            `json:"cts"`  //CreatedTs - autogenerated in backend
	UpdatedBy            string            `json:"uby"`  //UpdatedBy
	Blacklisted          bool              `json:"blacklisted"`
}

//EntityManager manages entity transactions
type EntityManager struct {
}

//dltDomainNames - list of valid domain names
var dltDomainNames = map[string]bool{
	"airtel.com":             true, //Airtel
	"vil.com":                true, //"VO" , "ID", "VI"
	"bsnl.com":               true, //BSNL
	"mtnl.com":               true, //MTNL
	"qtl.infotelconnect.com": true, //QTL
	"tata.com":               true, //TATA
	"jio.com":                true, //JIO
	"org1": true,//for testing purpose in local
	"org2":true,//for testing purpose in local
}

var svcprvDomain = map[string]string{
	"AI": "airtel.com",
	"VI": "vil.com",
	"BL": "bsnl.com",
	"ML": "mtnl.com",
	"QL": "qtl.infotelconnect.com",
	"TA": "tata.com",
	"JI": "jio.com",
	"VO": "vil.com",
	"ID": "vil.com",
	"Org1":"org1",
	"Org2":"org2",
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
	"Org1":true,
	"Org2":true,
}

var orgType = map[string]bool{
	"P": true,
	"G": true,
	"S": true,
	"U": true,
	"O": true,
}

var entityStatus = map[string]bool{
	"A": true,
	"I": true,
}

var blacklistValue = map[string]bool{
	"true":  true,
	"false": true,
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

//CheckValidityForBlacklistField checks for  validity of entity for status trxn
func CheckValidityForBlacklistField(searchEntityID, newBlacklistVal, newUpdatedTS, updatedBy string) (bool, string) {

	if searchEntityID == "" {
		return false, "Entity Id should be present there"
	}

	if !validEnumEntry(newBlacklistVal, blacklistValue) {
		return false, "Status: Enter Blacklist value either true or false"
	}
	if newUpdatedTS == "" {
		return false, "Update timeStamp should be present there"
	}

	return true, ""
}

//CheckValidityForStatus checks for  validity of entity for status trxn
func CheckValidityForStatus(searchEntityID, newStatus, newUpdatedTS, srvcProvider, updatedBy string) (bool, string) {

	if !validEnumEntry(srvcProvider, serviceProvider) {
		return false, "Invalid ServiceProvider"
	}

	if svcprvDomain[srvcProvider] != updatedBy {
		return false, "Sevice Provider and domain operator do not match"
	}

	if searchEntityID == "" {
		return false, "Entity Id should be present there"
	}

	if !validEnumEntry(newStatus, entityStatus) {
		return false, "Status: Enter either A, I"
	}
	if newUpdatedTS == "" {
		return false, "Update timeStamp should be present there"
	}

	return true, ""
}

func isValidDomainName(domainName string) (bool, string) {
	if !validEnumEntry(domainName, dltDomainNames) {
		return false, "WARNING: Needs to be a valid Domain name"
	}
	return true, ""
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

	if !validEnumEntry(e.EntityType, orgType) {
		return false, "Entity Type: Enter value P or G or S or U or O"
	}

	if e.EntityType == "P" {
		if len(e.POI) == 0 {
			return false, "PAN No is required"
		}
	}

	if !validEnumEntry(e.ServiceProvider, serviceProvider) {
		return false, "Invalid ServiceProvider"
	}

	if len(e.Status) != 1 {
		return false, "Status is mandatory and should be of length 1"
	}

	for p, val := range e.Status {
		if p != e.ServiceProvider {
			return false, "Invalid status update by operator"
		}
		if !validEnumEntry(val, entityStatus) {
			return false, "Status: Enter either A, I"
		}
	}

	if !validEnumEntry(e.EntityClassification, validCategoryMap) {
		return false, "Entity Classification : Must provide either PE or TM"
	}
	return true, ""
}

//IsValid checks if the entity fields are valid of not
func IsValid(e Entity, creator string) (bool, string) {

	if !validEnumEntry(e.ServiceProvider, serviceProvider) {
		return false, "Invalid ServiceProvider"
	}

	svcprv := e.ServiceProvider

	if svcprvDomain[svcprv] != creator {
		return false, "Svcprv domain and creator is not matched"
	}

	if len(e.EntityName) == 0 {
		return false, "Entity name is mandatory"
	}
	if !validEnumEntry(e.EntityType, orgType) {
		return false, "Entity Type: Enter value P or G or S or U or O"
	}

	if e.EntityType == "P" {
		if len(e.POI) == 0 {
			return false, "PAN No is required"
		}
	}

	if len(e.CreateTs) == 0 {
		return false, "CreateTS is mandatory"
	}

	if len(e.UpdateTs) == 0 {
		return false, "UpdateTs is mandatory"
	}

	if len(e.Status) != 1 {
		return false, "Status is mandatory and should be of length 1"
	}

	for p, val := range e.Status {
		if p != e.ServiceProvider {
			return false, "Invalid status update by operator"
		}
		if !validEnumEntry(val, entityStatus) {
			return false, "Status: Enter either A, I"
		}
	}

	if !validEnumEntry(e.EntityClassification, validCategoryMap) {
		return false, "Entity Classification : Must provide either PE or TM"
	}

	return true, ""
}

//SearchEntity searchs for entity based on the input parameters
func (em *EntityManager) SearchEntity(stub shim.ChaincodeStubInterface) peer.Response {
	var response peer.Response
	searchCriteria := make(map[string]string)
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}
	err := json.Unmarshal([]byte(args[0]), &searchCriteria)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}

	authorize, _ := em.getInvokerIdentity(stub)
	if authorize == false {
		return shim.Error("Unauthorized access")
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
		isOK, entities := em.retriveEntityRecords(stub, fmt.Sprintf(entitySearchCriteria, entityName), "entitySearchByName")

		if !isOK || entities == nil {
			return shim.Error("queryEntity:GetQueryResult is Failed")
		}

		recordsJSON, marshalErr := json.Marshal(entities)
		if marshalErr != nil {
			return shim.Error("Error marshalling Query response")
		}

		response = shim.Success(recordsJSON)
	case "id":
		entitySearchCriteria := `{
			"obj":"Entity"	,
			"id":"%s"
		}`
		entityID := searchCriteria[searchType]
		isOK, entities := em.retriveEntityRecords(stub, fmt.Sprintf(entitySearchCriteria, entityID), "entitySearchByID")
		if !isOK || entities == nil {
			return shim.Error("queryEntity:GetQueryResult is Failed")
		}

		recordsJSON, marshalErr := json.Marshal(entities)
		if marshalErr != nil {
			return shim.Error("Error marshalling Query response")
		}
		response = shim.Success(recordsJSON)

	case "svcprv":
		entitySearchCriteria := `{
			"obj":"Entity"	,
			"svcprv":"%s"
		}`
		entityID := searchCriteria[searchType]
		isOK, entities := em.retriveEntityRecords(stub, fmt.Sprintf(entitySearchCriteria, entityID), "entitySearchByAP")

		if !isOK || entities == nil {
			return shim.Error("queryEntity:GetQueryResult is Failed")
		}

		recordsJSON, marshalErr := json.Marshal(entities)
		if marshalErr != nil {
			return shim.Error("Error marshalling Query response")
		}
		response = shim.Success(recordsJSON)

	case "poi":
		entitySearchCriteria := `{
			"obj":"Entity"	,
			"poi":"%s"
		}`
		entityID := searchCriteria[searchType]
		isOK, entities := em.retriveEntityRecords(stub, fmt.Sprintf(entitySearchCriteria, entityID), "entitySearchByPoi")
		if !isOK || entities == nil {
			return shim.Error("queryEntity:GetQueryResult is Failed")
		}

		recordsJSON, marshalErr := json.Marshal(entities)
		if marshalErr != nil {
			return shim.Error("Error marshalling Query response")
		}
		response = shim.Success(recordsJSON)

	default:
		response = shim.Error("Unsupported search type provided " + searchType)
	}
	return response
}

//CreateEntity creates an entity in the ledger
func (em *EntityManager) CreateEntity(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}
	var entityToSave Entity
	err := json.Unmarshal([]byte(args[0]), &entityToSave)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}

	authorize, creator := em.getInvokerIdentity(stub)
	if authorize == false {
		return shim.Error("Unauthorized access")
	}

	//Second Check if the entity id is existing or not
	if recordBytes, _ := stub.GetState(entityToSave.EntityID); len(recordBytes) > 0 {
		return shim.Error("Entityid already registered. Provide an unique entity id")
	}

	entityToSave.ObjType = "Entity"
	entityToSave.Creator = creator
	entityToSave.UpdatedBy = creator
	entityToSave.Blacklisted = false //default to false

	//mandatory field validation
	entityJSON, _ := json.Marshal(entityToSave)
	if isValid, errMsg := IsValid(entityToSave, creator); !isValid {
		return shim.Error(errMsg)
	}

	//svcprv with domain name validation

	//Save the entry
	_entityLogger.Info("entityToSave.EntityID----------", entityToSave.EntityID)
	err = stub.PutState(entityToSave.EntityID, entityJSON)

	if err != nil {
		return shim.Error("Unable to save with entity id " + entityToSave.EntityID)
	}
	retErr := stub.SetEvent(_CreateEvent, entityJSON)

	if retErr != nil {
		_entityLogger.Errorf("Event not generated for event : CREATEENTITY")
		return shim.Error("{\"error\":\"Unable to generate Create Entity Event.\"}")
	}
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"entityID": entityToSave.EntityID,
		"message":  "Save successful",
		"entity":   entityToSave,
		"status":   "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

//ModifyEntity modifies an existing entry
func (em *EntityManager) ModifyEntity(stub shim.ChaincodeStubInterface) peer.Response {

	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid arguments provided")
	}

	authorize, creatorUpdateBy := em.getInvokerIdentity(stub)
	if authorize == false {
		return shim.Error("Unauthorized access")
	}

	var modifiedEntity Entity
	errEntity := json.Unmarshal([]byte(args[0]), &modifiedEntity)
	if errEntity != nil {
		return shim.Error(errEntity.Error())
	}

	entityRecords, err := stub.GetState(modifiedEntity.EntityID)
	if err != nil {
		return shim.Error(err.Error())
	}else if entityRecords == nil {
		return shim.Error("Entity does not exist with this EntityId")
	}

	var existingEntity Entity
	errExistingEntity := json.Unmarshal([]byte(entityRecords), &existingEntity)
	if errExistingEntity != nil {
		return shim.Error(errExistingEntity.Error())
	}

	modifiedEntity.ObjType = existingEntity.ObjType
	modifiedEntity.CreateTs = existingEntity.CreateTs
	modifiedEntity.Creator = existingEntity.Creator

	modifiedEntity.UpdatedBy = creatorUpdateBy

	if isValid, errMsg := IsValidEntityIDPresent(modifiedEntity); !isValid {
		return shim.Error(errMsg)
	}

	marshalEntryJSON, _ := json.Marshal(modifiedEntity)
	finalErr := stub.PutState(modifiedEntity.EntityID, marshalEntryJSON)

	if finalErr != nil {
		return shim.Error("Unable to save with entity id " + modifiedEntity.EntityID)
	}
	retErr := stub.SetEvent(_ModifyEvent, marshalEntryJSON)

	if retErr != nil {
		_entityLogger.Errorf("Event not generated for event : MODIFY_ENTITY")
		return shim.Error("{\"error\":\"Unable to generate event for Modify Entity.\"}")
	}
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"entityID": modifiedEntity.EntityID,
		"message":  "Save successful",
		"entity":   modifiedEntity,
		"status":   "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

func (em *EntityManager) retriveEntityRecords(stub shim.ChaincodeStubInterface, criteria string, indexs ...string) (bool, []Entity) {
	var finalSelector string

	records := make([]Entity, 0)

	if len(indexs) == 0 {
		finalSelector = fmt.Sprintf("{\"selector\":%s }", criteria)

	} else {
		finalSelector = fmt.Sprintf("{\"selector\":%s , \"use_index\" :\"%s\" }", criteria, indexs[0])
	}

	_entityLogger.Infof("Query Selector : %s", finalSelector)
	resultsIterator, err := stub.GetQueryResult(finalSelector)
	if err != nil {
		_entityLogger.Errorf("queryEntity:GetQueryResult is Failed with error :" + string(err.Error()))
		return false, nil
	}

	for resultsIterator.HasNext() {
		record := Entity{}
		recordBytes, iteratorErr := resultsIterator.Next()
		if iteratorErr != nil {
			_entityLogger.Errorf("queryEntity:GetQueryResult is Failed with error :" + string(iteratorErr.Error()))
			return false, nil
		}
		err := json.Unmarshal(recordBytes.Value, &record)
		if err != nil {
			_entityLogger.Infof("Unable to unmarshal entity retived:: %v", err)
			return false, nil
		}
		records = append(records, record)
	}

	return true, records
}

// GetHistoryByKey queries the ledger using key.
// It retrieve all the changes to the value happened over time.
func (em *EntityManager) GetHistoryByKey(stub shim.ChaincodeStubInterface) peer.Response {
	_entityLogger.Debug("GetHistoryByKey called.")

	_, args := stub.GetFunctionAndParameters()

	// Essential check to verify number of arguments
	if len(args) != 1 {
		_entityLogger.Error("Incorrect number of arguments passed in GetHistoryByKey.")
		resp := shim.Error("Incorrect number of arguments. Expecting 1 arguments: " + strconv.Itoa(len(args)) + " given.")
		return resp
	}

	authorize, _ := em.getInvokerIdentity(stub)
	if authorize == false {
		return shim.Error("Unauthorized access")
	}

	key := args[0]
	resultsIterator, err := stub.GetHistoryForKey(key)

	if err != nil {
		_entityLogger.Error("Error occured while calling GetHistoryForKey(): ", err)
		return shim.Error("Error occured while calling GetHistoryForKey: " + err.Error())
	}

	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the event
	historicResponse := make([]map[string]interface{}, 0)
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			_entityLogger.Error("Error occured while calling resultsIterator.Next(): ", err)
			return shim.Error("Error occured while calling GetHistoryByKey (resultsIterator): " + err.Error())
		}
		value := make(map[string]interface{})
		err1 := json.Unmarshal(response.Value, &value)
		if err1 != nil {
			return shim.Error("Error occured while Unmarhslaling GetHistoryByKey (resultsIterator): " + err1.Error())
		}
		historicResponse = append(historicResponse, map[string]interface{}{"txId": response.TxId, "value": value})

	}

	respJSON, _ := json.Marshal(historicResponse)
	return shim.Success(respJSON)
}

//UpdateBlacklistedValue updates the status of the entity ledger entry
func (em *EntityManager) UpdateBlacklistedValue(stub shim.ChaincodeStubInterface) peer.Response {
	_entityLogger.Info("within UpdateBlacklistedValue")
	_, args := stub.GetFunctionAndParameters()

	if len(args) < 3 {
		return shim.Error("Invalid No of arguments provided")
	}

	authorize, updatedBy := em.getInvokerIdentity(stub)
	if authorize == false {
		return shim.Error("Unauthorized access")
	}

	searchEntityID := args[0]
	blacklistSts := args[1]
	newUpdatedTS := args[2]

	if isValid, errMsg := CheckValidityForBlacklistField(searchEntityID, blacklistSts, newUpdatedTS, updatedBy); !isValid {
		return shim.Error(errMsg)
	}
	blackListStsBool, parseErr := strconv.ParseBool(blacklistSts)
	if parseErr != nil {
		return shim.Error("Invalid blacklist values provided")
	}

	entityRecords, errR := stub.GetState(searchEntityID)
	if errR != nil {
		return shim.Error(errR.Error())
	}

	if entityRecords == nil {
		return shim.Error("{\"error\":\"The EntityID doesn't exists\"}")
	}

	var updatedBlacklistEntity Entity
	err := json.Unmarshal(entityRecords, &updatedBlacklistEntity)
	if err != nil {
		return shim.Error("{\"error\":\"Error when unmarshaling the data\"}")
	}

	currentBlackListField := updatedBlacklistEntity.Blacklisted

	if currentBlackListField == blackListStsBool {
		if blackListStsBool == true {
			return shim.Error("{\"error\":\"Entity already blacklisted.\"}")
		} else if blackListStsBool == false {
			return shim.Error("{\"error\":\"Entity already not blacklisted.\"}")
		}

	}

	updatedBlacklistEntity.UpdateTs = newUpdatedTS

	updatedBlacklistEntity.UpdatedBy = updatedBy
	updatedBlacklistEntity.Blacklisted = blackListStsBool

	_entityLogger.Info("updated Entity blacklisted------------", updatedBlacklistEntity)

	marshalEntryJSON, err := json.Marshal(updatedBlacklistEntity)
	if err != nil {
		return shim.Error("{\"error\":\"Error at the time of Marshaling\"}")
	}

	finalErr := stub.PutState(updatedBlacklistEntity.EntityID, marshalEntryJSON)

	if finalErr != nil {
		return shim.Error("Unable to save with entity id " + updatedBlacklistEntity.EntityID)
	}
	retErr := stub.SetEvent(_BlacklistEntity, marshalEntryJSON)

	if retErr != nil {
		_entityLogger.Errorf("Event not generated for event : BLACKLIST_ENTITY")
		return shim.Error("{\"error\":\"Unable to update entity blacklise information.\"}")
	}
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"entityID": updatedBlacklistEntity.EntityID,
		"message":  "Save successful",
		"entity":   updatedBlacklistEntity,
		"status":   "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)

}

//UpdateEntityStatus updates the status of the entity ledger entry
func (em *EntityManager) UpdateEntityStatus(stub shim.ChaincodeStubInterface) peer.Response {
	_entityLogger.Info("within UpdateEntityStatus")
	_, args := stub.GetFunctionAndParameters()

	if len(args) < 4 {
		return shim.Error("Invalid No of arguments provided")
	}

	authorize, updatedBy := em.getInvokerIdentity(stub)
	if authorize == false {
		return shim.Error("Unauthorized access")
	}

	searchEntityID := args[0]
	newStatus := args[1]
	newUpdatedTS := args[2]
	serviceProvider := args[3]

	if isValid, errMsg := CheckValidityForStatus(searchEntityID, newStatus, newUpdatedTS, serviceProvider, updatedBy); !isValid {
		return shim.Error(errMsg)
	}

	entityRecords, errR := stub.GetState(searchEntityID)

	if errR != nil {
		return shim.Error(errR.Error())
	}

	if entityRecords == nil {
		return shim.Error("The EntityID doesn't exists")
	}

	var updatedStatusEntity Entity
	err := json.Unmarshal(entityRecords, &updatedStatusEntity)
	if err != nil {
		return shim.Error("Error when unmarshaling the data")
	}

	_entityLogger.Info("serviceProvider------------", serviceProvider)
	updatedStatusEntity.Status[serviceProvider] = newStatus
	updatedStatusEntity.UpdateTs = newUpdatedTS

	updatedStatusEntity.UpdatedBy = updatedBy

	_entityLogger.Info("updatedStatusEntity------------", updatedStatusEntity)

	marshalEntryJSON, marErr := json.Marshal(updatedStatusEntity)
	if marErr != nil {
		return shim.Error("Failed to marshal data")
	}
	finalErr := stub.PutState(updatedStatusEntity.EntityID, marshalEntryJSON)

	if finalErr != nil {
		return shim.Error("Unable to save with entity id " + updatedStatusEntity.EntityID)
	}
	retErr := stub.SetEvent(_ModifyEvent, marshalEntryJSON)

	if retErr != nil {
		_entityLogger.Errorf("Event not generated for event : MODIFY_ENTITY")
		return shim.Error("{\"error\":\"Unable to update entity status.\"}")
	}
	resultData := map[string]interface{}{
		"trxnID":   stub.GetTxID(),
		"entityID": updatedStatusEntity.EntityID,
		"message":  "Save successful",
		"entity":   updatedStatusEntity,
		"status":   "true",
	}
	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)

}

//SearchEntityIDArray -> it will search all the entityID Array field
func (em *EntityManager) SearchEntityIDArray(stub shim.ChaincodeStubInterface) peer.Response {
	_entityLogger.Info("within searchEntityIDArray")

	_, args := stub.GetFunctionAndParameters()

	searchEntityIDArray := args

	searchEntityIDArrayField := fmt.Sprintf("%+q", searchEntityIDArray)
	searchEntityIDArrayField = strings.Replace(searchEntityIDArrayField, " ", ",", -1)

	query := make([]string, 0)
	queryString := ""

	if len(args) > 0 {
		entityIDQuery := `{
			"id": {
			   "$in":%v
			}
		}`
		entityIDQuery = fmt.Sprintf(entityIDQuery, searchEntityIDArrayField)
		query = append(query, entityIDQuery)
	}

	if len(query) == 0 {
		return shim.Error("Error found at the time of query formation")
	} else {
		queryString = strings.Join(query, ",")
	}

	finalSelector := fmt.Sprintf("{\"selector\":%s }", queryString)
	_entityLogger.Infof("Query Selector : %s", finalSelector)

	isOk, records := ExecuteRichQuery(stub, finalSelector)
	if !isOk {
		return shim.Error("{\"error\":\"Unable to retrieve entity details.\"}")
	}

	marshalJSON, err := json.MarshalIndent(records, "", "    ")
	if err != nil {
		_entityLogger.Errorf("Error marshall data %s", err)
		return shim.Error("{\"error\":\"Unable to retrieve entity details.\"}")
	}

	return shim.Success(marshalJSON)
}

//ExecuteRichQuery for arrayfield EntityID search
func ExecuteRichQuery(stub shim.ChaincodeStubInterface, selectorString string) (bool, []Entity) {

	var records []Entity
	resultsIterator, errResult := stub.GetQueryResult(selectorString)

	if errResult != nil {
		_entityLogger.Errorf("Error found in GetQueryResult: %s", errResult)
		return false, nil
	}

	for resultsIterator.HasNext() {
		var data Entity
		commonByteArray, iterErr := resultsIterator.Next()
		if iterErr != nil {
			_entityLogger.Errorf("Error found in Iterator: %s", iterErr)
			return false, nil
		}
		err := json.Unmarshal([]byte(commonByteArray.Value), &data)
		if err != nil {
			_entityLogger.Errorf("Error found in Unmarshall data: %s", err)
			return false, nil
		}
		records = append(records, data)
	}

	return true, records
}

//EntityQueryWithPagination ====  Query With Pagination ===
func (em *EntityManager) EntityQueryWithPagination(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 3 {
		_entityLogger.Errorf("Incorrect number of arguments. Expecting 3")
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	queryString := args[0]
	//return type of ParseInt is int64
	pageSize, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		_entityLogger.Errorf(err.Error())
		return shim.Error(err.Error())
	}
	bookmark := args[2]
	_entityLogger.Infof("Query String %s", queryString)
	queryResults, err := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookmark)

	if err != nil {
		_entityLogger.Errorf(err.Error())
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryStringWithPagination executes the passed in query string with
// pagination info. Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================

func getQueryResultForQueryStringWithPagination(stub shim.ChaincodeStubInterface, queryString string, pageSize int32, bookmark string) ([]byte, error) {
	resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(queryString, pageSize, bookmark)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	buffer, err := constructQueryResponseWithPaginationFromIterator(resultsIterator, responseMetadata)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// ===========================================================================================
// constructQueryResponseFromIterator constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================

func constructQueryResponseWithPaginationFromIterator(resultsIterator shim.StateQueryIteratorInterface, responseMetadata *pb.QueryResponseMetadata) (*bytes.Buffer, error) {

	// buffer is a JSON array containing QueryResults

	var buffer bytes.Buffer
	buffer.WriteString("{\"RecordsCount\":")
	buffer.WriteString("\"")
	buffer.WriteString(fmt.Sprintf("%v", responseMetadata.FetchedRecordsCount))
	buffer.WriteString("\"")
	buffer.WriteString(", \"Bookmark\":")
	buffer.WriteString("\"")
	buffer.WriteString(responseMetadata.Bookmark)
	buffer.WriteString("\"")
	buffer.WriteString(",\"Result\":[")
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

		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true
	}

	buffer.WriteString("]}")
	return &buffer, nil

}

//Returns the complete identity in the format
//Certitificate issuer orgs's domain name
//Returns string Unkown if not able parse the invoker certificate
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
	isOK, msg := isValidDomainName(issuersOrgs[0])
	if !isOK {
		return false, msg
	}
	return true, fmt.Sprintf("%s", issuersOrgs[0])

}
