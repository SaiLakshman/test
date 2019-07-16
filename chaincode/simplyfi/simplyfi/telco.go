package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim" // import for Chaincode Interface
	pb "github.com/hyperledger/fabric/protos/peer"      // import for peer response
)

type Telco struct {
}

//Entity Data
type Entity struct {
	DocType        string `json:"entityDocType"`
	EntityId       string `json:"entityId"`
	TelemarketerId string `json:"entTmId"`
	EntityName     string `json:"entityName"`
	EntityAddress  string `json:"entityAddress"`
	EmailId        string `json:"emailId"`
	ContactNumber  string `json:"contactNumber"`
	Status         string `json:"entityStatus"`
	CIN            string `json:"entityCIN"`
}

//Registrar Data
type Registrar struct {
	RegistrarId   string `json:"registrarId"`
	TSPId         string `json:"regTspId"`
	RegistrarName string `json:"registrarName"`
	RegistrarType string `json:"registrarType"`
	EmailId       string `json:"registrarEmailId"`
	Status        string `json:"registrarStatus"`
}

//TSP Data
type TSP struct {
	TSPId   string `json:"tspId"`
	TSPName string `json:"tspName"`
}

//TeleMarketer Data
type TeleMarketer struct {
	TMId                 string `json:"tmId"`
	TMName               string `json:"tmName"`
	TMAddress            string `json:"tmAddress"`
	TMState              string `json:"tmState"`
	TMDistrict           string `json:"tmDistrict"`
	TMContactNo          string `json:"tmContactNo"`
	TMEmailId            string `json:"tmEmailId"`
	TMDateOfRegistration string `json:"tmDor"`
	TMStatus             string `json:"tmStatus"`
	TMCIN                string `json:"tmCIN"`
}

//TSP Onboarding TM
type TMWithTSP struct {
	TSPId          string `json:"tspId"`
	TMId           string `json:"tmId"`
	ConnectionId   string `json:"connecId"`
	ConnectionType string `json:"connecType"`
	Validity       string `json:"tspTmValidity"`
	Status         string `json:"TmTspStatus"`
}

//Template Data
type Template struct {
	TemplateId          string `json:"tempId"`
	TelemarketerId      string `json:"tempTmId"`
	TSPId               string `json:"tempTSPId"`
	TemplateName        string `json:"tempName"`
	TemplateType        string `json:"tempType"`
	TemplateCategory    string `json:"tempCategory"`
	TemplateEntityId    string `json:"tempEntityId"`
	TemplateBody        string `json:"tempBody"`
	TemplateCreatedDate string `json:"createDate"`
	TemplateStatus      string `json:"tempStatus"`
	TemplateValidity    string `json:"tempValidity"`
}

//Preferences Data
type Preference struct {
	DocType          string     `json:"preferenceDocType"`
	SubscriberNumber string     `json:"subscriberNum"`
	Category         Categories `json:"category"`
	Mode             Modes      `json:"mode"`
	TimeBand         TimeBands  `json:"timeBand"`
	Day              Days       `json:"day"`
}

//Categories Data
type Categories struct {
	All               string `json:"allCategories"`
	BlockPromo        string `json:"blockPromo"`
	FinancialServices string `json:"financialServices"`
	Education         string `json:"education"`
	RealEstate        string `json:"realEstate"`
	Health            string `json:"health"`
	Consumergoods     string `json:"consumerGoods"`
	Broadcasting      string `json:"broadcasting"`
	Tourism           string `json:"tourism"`
	Food              string `json:"food"`
}

//Modes Data
type Modes struct {
	AllModes            string `json:"allModes"`
	VoiceCall           string `json:"voiceCall"`
	SMS                 string `json:"sms"`
	ADCPreRecorded      string `json:"adcPreRecorded"`
	ADCWithConnectivity string `json:"adcWithConnectivity"`
	RoboCall            string `json:"roboCall"`
}

//Days Data
type Days struct {
	Monday    string `json:"monday"`
	Tuesday   string `json:"tuesday"`
	Wednesday string `json:"wednesday"`
	Thursday  string `json:"thursday"`
	Friday    string `json:"friday"`
	Saturday  string `json:"saturday"`
	Sunday    string `json:"sunday"`
	All       string `json:"allDays"`
	Holiday   string `json:"holidays"`
}

//TimeBands Data
type TimeBands struct {
	All   string `json:"allTimeBands"`
	Slot0 string `json:"slot0"`
	Slot1 string `json:"slot1"`
	Slot2 string `json:"slot2"`
	Slot3 string `json:"slot3"`
	Slot4 string `json:"slot4"`
	Slot5 string `json:"slot5"`
	Slot6 string `json:"slot6"`
	Slot7 string `json:"slot7"`
	Slot8 string `json:"slot8"`
}

//Consent Data
type Consent struct {
	DocType           string `json:"docType"`
	MasterConsentId   string `json:"masterConsentId"`
	EntityId          string `json:"consententityId"`
	SubscriberNo      string `json:"subscriberNumber"`
	CommunicationType string `json:"communicationType"`
	Type              string `json:"type"`
	Category          string `json:"category"`
	Purpose           string `json:"purpose"`
	ContentTemplateId string `json:"contentTemplateId"`
	Status            string `json:"consentStatus"`
}

//Header Data
type Header struct {
	HeaderName         string `json:"headerName"`
	HeaderEntityId     string `json:"headerEntityId"`
	TSPId              string `json:"headerTSPId"`
	TelemarketerId     string `json:"headerTmId"`
	HeaderType         string `json:"headerType"`
	HeaderCategory     string `json:"headerCategory"`
	HeaderStatus       string `json:"headerStatus"`
	HeaderValidity     string `json:"headerValidity"`
	HeaderCreatedDate  string `json:"headerCreatedDate"`
	HeaderModifiedDate string `json:"headerModifiedDate"`
}

//Campaign Data
type Scrubbing struct {
	CampaignId   string `json:"campaignId"`
	ConsentId    string `json:"scrubConsentId"`
	SubscriberNo string `json:"subscriberNoList"`
	EntityId     string `json:scrubEntityId`
	HeaderName   string `json:"scrubHeaderName"`
	DateTime     string `json:"dateTime"`
	TemplateId   string `json:"scrubTemplateId"`
	Status       string `json:"scrubStatus"` //scrubbing status for subscriber list
	InputHash    string `json:"inputHash"`
	OutputHash   string `json:"outputHash"`
}

//Complaints Data
type Complaint struct {
	DocType           string `json:"complaintsDocType"`
	SubscriberNo      string `json:"subscriberNo"`
	UniqueReferenceNo string `json:"uniqueReferenceNumber"`
	ComplaintDate     string `json:"complaintDate"`
	ComplaintTime     string `json:"complaintTime"`
	UCCDate           string `json:"uccDate"`
	OAP               string `json:"oap"`
	HeaderName        string `json:"headerName"`
	TAPCode           string `json:"tapCode"`
	UCCOAPDate        string `json"uccOapDate"`
	UCCType           string `json:"uccType"`
	UCCDescription    string `json:"uccDescription"`
	CDRStatus         string `json:"cdrStatus"`
	Status            string `json:"status"`
	ActionTaken       string `json:"actionTaken"`
	ComplaintType     string `json:"complaintType"`
	Remarks           string `json:"remarks"`
}

//Init function of the chaincode
func (c *Telco) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

//Invoke function of the chaincode
func (c *Telco) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "setEntity":
		return c.setEntity(stub, args)
	case "getEntity":
		return c.getEntity(stub, args)
	case "getAllEntities":
		return c.queryEntityByTM(stub, args)
	case "getNumberOfEntities":
		return c.getNumberOfEntities(stub, args)
	case "updateEntityStatus":
		return c.updateEntityStatus(stub, args)
	case "setRegistrar":
		return c.setRegistrar(stub, args)
	case "getRegistrar":
		return c.getRegistrar(stub, args)
	case "getAllRegistrars":
		return c.queryRegistrarByTSP(stub, args)
	case "updateRegistrarStatus":
		return c.updateRegistrarStatus(stub, args)
	case "setTsp":
		return c.setTelecomServiceProvider(stub, args)
	case "getTsp":
		return c.getTelecomServiceProvider(stub, args)
	case "setTM":
		return c.setTelemarketer(stub, args)
	case "getTM":
		return c.getTelemarketer(stub, args)
	case "getAllTelemarketers":
		return c.getAllTelemarketers(stub, args)
	case "updateTMStatus":
		return c.updateTelemarketerStatus(stub, args)
	case "setTMOnTSP":
		return c.setTMOnboardingWithTSP(stub, args)
	case "getTMOnTSP":
		return c.getTMOnboardingWithTSP(stub, args)
	case "getTMByTSP":
		return c.getAllTMByTSP(stub, args)
	case "setTemplate":
		return c.setTemplate(stub, args)
	case "getTemplate":
		return c.getTemplate(stub, args)
	case "modifyTemplateStatus":
		return c.modifyTemplateStatus(stub, args)
	case "getAllTemplates":
		return c.queryTemplateByEntity(stub, args)
	case "getAllTemplatesByStatus":
		return c.queryTemplateByStatus(stub, args)
	case "setPreference":
		return c.setPreference(stub, args)
	case "getPreference":
		return c.getPreference(stub, args)
	case "getNumberOfPreferences":
		return c.getNumberOfPreferences(stub, args)
	case "gnpwc":
		return c.getNumberOfPreferencesWithCategory(stub, args)
	case "getPreferencesWithPagination":
		return c.getPreferencesWithPagination(stub, args)
	case "setConsent":
		return c.setConsent(stub, args)
	case "getConsentWithEntity":
		return c.getConsentWithEntity(stub, args)
	case "getConsentWithSubscriber":
		return c.getConsentWithSubscriber(stub, args)
	case "approveConsentRequest":
		return c.createConsentForSubscriberNo(stub, args)
	case "updateConsentStatus":
		return c.updateConsentStatus(stub, args)
	case "getAllConsentsWithSubscriberNumber":
		return c.queryAllConsentsBySubscriberNo(stub, args)
	case "getAllConsentsWithEntityId":
		return c.queryAllConsentsByEntityId(stub, args)
	case "queryAllConsentsByDocTypeAndStatus":
		return c.queryAllConsentsByDocTypeAndStatus(stub, args)
	case "raiseComplaint":
		return c.raiseComplaint(stub, args)
	case "fetchComplaint":
		return c.fetchComplaint(stub, args)
	case "queryComplaintByOAP":
		return c.queryComplaintByOAP(stub, args)
	case "queryComplaintByTAP":
		return c.queryComplaintByTAP(stub, args)
	case "queryComplaintByDate":
		return c.queryComplaintByDate(stub, args)
	case "queryComplaintByStatus":
		return c.queryComplaintByStatus(stub, args)
	case "updateCDRStatus":
		return c.updateCDRStatus(stub, args)
	case "updateStatus":
		return c.updateComplaintStatus(stub, args)
	case "getComplaintsBySubscriberNumber":
		return c.getAllComplaintsBySubscriberNumber(stub, args)
	case "setHeader":
		return c.setHeader(stub, args)
	case "getHeader":
		return c.getHeader(stub, args)
	case "modifyHeaderStatus":
		return c.modifyHeaderStatus(stub, args)
	case "getAllHeaders":
		return c.queryHeaderByEntity(stub, args)
	case "getAllHeadersByStatus":
		return c.queryHeaderByStatus(stub, args)
	case "promotionalScrub":
		return c.promotionalScrubbing(stub, args)
	case "transactionalScrub":
		return c.transactionalScrubbing(stub, args)
	case "serviceScrub":
		return c.serviceScrubbing(stub, args)
	case "createCampaign":
		return c.createCampaign(stub, args)
	case "getScrubStatus":
		return c.getScrubbingStatus(stub, args)
	case "updateScrubHash":
		return c.updateScrubOutputHash(stub, args)
	case "getAllCampaigns":
		return c.getAllCampaigns(stub, args)
	default:
		return shim.Error("Not a Valid Function.")
	}
}

//Creating Entity in Blockchain
/*
	fcnName: setEntity
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [entityId, tmId, entityName, entityAddress, emailId, contactNum, status, cin]
	create a entity object of the structure entity,
	set all the necessary data of that object
	marshall the object into jsonObject
	create a tx in blockchain using PutState
	return pb.Response= "Entity Created Successfully"
*/
func (c *Telco) setEntity(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 8 {
		return shim.Error("Incorrect Number of Arguments. Expecting 8(EntityId, TelemarketerId, EntityName, EntityAddress, EmailId, ContactNumber, Status, CIN))")
	}
	docType := "Entity"
	entityId := args[0]
	tmId := args[1]
	entityName := args[2]
	entityAddress := args[3]
	emailId := args[4]
	contactNum := args[5]
	status := args[6]
	cin := args[7]

	entityStruct := &Entity{}
	entityStruct.DocType = docType
	entityStruct.EntityId = entityId
	entityStruct.TelemarketerId = tmId
	entityStruct.EntityName = entityName
	entityStruct.EntityAddress = entityAddress
	entityStruct.EmailId = emailId
	entityStruct.ContactNumber = contactNum
	entityStruct.Status = status
	entityStruct.CIN = cin

	entityAsBytes, err := json.Marshal(entityStruct)
	if err != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Entity. \"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutState(entityStruct.EntityId, entityAsBytes)
	if err != nil {
		jsonResp = "{\"Error\":\"Creating Entity Data Failed\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Entity Created Successfully!!"))
}

//Retrieving Entity given EntityId from Blockchain
/*
	fcnName: getEntity
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [entityId]
	retrieve the tx from blockchain using GetState
	return pb.Response= Payload of entity which was created, else error saying does not exist
*/
func (c *Telco) getEntity(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of Arguments. Expecting 1(EntityId)")
	}
	entityId := args[0]
	valueAsBytes, err := stub.GetState(entityId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + entityId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Entity does not exist: " + entityId + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valueAsBytes)
}

//Retrieving all the Entities given TelemarketerId from Blockchain
/*
	fcnName: queryEntityByTM
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [tmId]
	retrieve all the entities who are under the tmId from blockchain using range query(selector query)
	return pb.Response= Payload of all the entities under that tmId
*/
func (c *Telco) queryEntityByTM(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonresp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. expecting 1(TmId)")
	}
	entTmId := args[0]
	querystring := fmt.Sprintf("{\"selector\":{\"entTmId\":\"%s\"}}", entTmId)
	queryresults, err := getQueryResultForQueryString(stub, querystring)
	if err != nil {
		jsonresp = "{\"Error\":\"Failed to get state for " + entTmId + "\"}"
		return shim.Error(jsonresp)
	}
	return shim.Success(queryresults)
}

//Retrieving all the Entities from the blockchain based on docType
/*
	fcnName: getNumberOfEntities
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [empty]
	to know the total number of entities in blockchain, using range query(selector query)
	return pb.Response= Payload of all the entities
*/
func (c *Telco) getNumberOfEntities(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonresp string
	docType := "Entity"
	querystring := fmt.Sprintf("{\"selector\":{\"entityDocType\":\"%s\"}}", docType)
	queryresults, err := getQueryResultForQueryString(stub, querystring)
	if err != nil {
		jsonresp = "{\"Error\":\"Failed to get state.\"}"
		return shim.Error(jsonresp)
	}
	return shim.Success(queryresults)
}

//Modify & Update the status of Entity in Blockchain
/*
	fcnName: updateEntityStatus
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [entityId, status]
	retrieve the entity from blockchain given entityId
	create a entity object from entity structure
	unmarshal the retrieved data
	set the status
	marshal the data and create the data in blockchain using PutState
	return pb.Response= "Entity Status Updated Successfully"
*/
func (c *Telco) updateEntityStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 2 {
		return shim.Error("Incorrect number of Arguments. Expecting 2(EntityId, Status)")
	}
	entityId := args[0]
	status := args[1]
	valueAsBytes, err := stub.GetState(entityId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + entityId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Entity does not exist: " + entityId + "\"}"
		return shim.Error(jsonResp)
	}
	entityStruct := &Entity{}
	err = json.Unmarshal(valueAsBytes, &entityStruct)
	entityStruct.Status = status
	entityAsBytes, err1 := json.Marshal(entityStruct)
	if err1 != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Entity Status Updation \"}"
		return shim.Error(jsonResp)
	}
	err1 = stub.PutState(entityId, entityAsBytes)
	if err1 != nil {
		jsonResp = "{\"Error\":\"Updating Entity Status Failed \"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Entity Status Updated Successfully!!"))
}

//Creating Registrar in Blockchain
/*
	fcnName: setRegistrar
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [registrarId, tspId, registrarName, registrarType, emailId, status]
	create a registrar object of the structure registrar,
	set all the necessary data of that object
	marshall the object into jsonObject
	create a tx in blockchain using PutState
	return pb.Response= "Registrar Created Successfully"
*/
func (c *Telco) setRegistrar(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 6 {
		return shim.Error("Incorrect Number of Arguments. Expecting 6(RegistrarId, TSPId, RegistrarName, RegistrarType, EmailId, Status))")
	}
	registrarId := args[0]
	tspId := args[1]
	registrarName := args[2]
	registrarType := args[3]
	emailId := args[4]
	status := args[5]

	registrarStruct := &Registrar{}
	registrarStruct.RegistrarId = registrarId
	registrarStruct.TSPId = tspId
	registrarStruct.RegistrarName = registrarName
	registrarStruct.RegistrarType = registrarType
	registrarStruct.EmailId = emailId
	registrarStruct.Status = status

	registrarAsBytes, err := json.Marshal(registrarStruct)
	if err != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Entity. \"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutState(registrarStruct.RegistrarId, registrarAsBytes)
	if err != nil {
		jsonResp = "{\"Error\":\"Creating Registrar Data Failed\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Registrar Created Successfully!!"))
}

//Retrieving Registrar given RegistrarId from Blockchain
/*
	fcnName: getRegistrar
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [registrarId]
	retrieve the tx from blockchain using GetState
	return pb.Response= Payload of registrar which was created, else error saying does not exist
*/
func (c *Telco) getRegistrar(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of Arguments. Expecting 1(RegistrarId)")
	}
	registrarId := args[0]
	valueAsBytes, err := stub.GetState(registrarId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + registrarId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Registrar does not exist: " + registrarId + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valueAsBytes)
}

//Retrieve all the Registrars given TSPId from Blockchain
/*
	fcnName: queryRegistrarByTSP
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [tspId]
	retrieve all the registrars who are under the tspId from blockchain using range query(selector query)
	return pb.Response= Payload of all the registrars under that tspId
*/
func (c *Telco) queryRegistrarByTSP(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonresp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. expecting 1(TspId)")
	}
	regTspId := args[0]
	querystring := fmt.Sprintf("{\"selector\":{\"regTspId\":\"%s\"}}", regTspId)
	queryresults, err := getQueryResultForQueryString(stub, querystring)
	if err != nil {
		jsonresp = "{\"Error\":\"Failed to get state for " + regTspId + "\"}"
		return shim.Error(jsonresp)
	}
	return shim.Success(queryresults)
}

//Modify & Update the status of Registrar in Blockchain
/*
	fcnName: updateRegistrarStatus
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [regId, status]
	retrieve the registrar from blockchain given regId
	create a registrar object from registrar structure
	unmarshal the retrieved data
	set the status
	marshal the data and create the data in blockchain using PutState
	return pb.Response= "Registrar Status Updated Successfully"
*/
func (c *Telco) updateRegistrarStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 2 {
		return shim.Error("Incorrect number of Arguments. Expecting 2(RegistrarId, Status)")
	}
	regId := args[0]
	status := args[1]
	valueAsBytes, err := stub.GetState(regId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + regId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Regsitrar does not exist: " + regId + "\"}"
		return shim.Error(jsonResp)
	}
	registrarStruct := &Registrar{}
	err = json.Unmarshal(valueAsBytes, &registrarStruct)
	registrarStruct.Status = status
	regAsBytes, err1 := json.Marshal(registrarStruct)
	if err1 != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Registrar Status Updation \"}"
		return shim.Error(jsonResp)
	}
	err1 = stub.PutState(regId, regAsBytes)
	if err1 != nil {
		jsonResp = "{\"Error\":\"Updating Registrar Status Failed \"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Registrar Status Updated Successfully!!"))
}

//Creating TelecomServiceProvider in Blockchain
/*
	fcnName: setTelecomServiceProvider
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [tspId, tspName]
	create a telecomserviceprovider object of the structure telecomserviceprovider,
	set all the necessary data of that object
	marshall the object into jsonObject
	create a tx in blockchain using PutState
	return pb.Response= "Telecom Service Provider Registered Successfully"
*/
func (c *Telco) setTelecomServiceProvider(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 2 {
		return shim.Error("Incorrect Number of Arguments. Expecting 2(TspId, TSPName)")
	}

	tspId := args[0]
	tspName := args[1]

	tspStruct := &TSP{}
	tspStruct.TSPId = tspId
	tspStruct.TSPName = tspName
	tspAsBytes, err := json.Marshal(tspStruct)
	if err != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for TSP \"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutState(tspStruct.TSPId, tspAsBytes)
	if err != nil {
		jsonResp = "{\"Error\":\"Creating Telecom Service Providers Data Failed\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Telecom Service Provider Created Successfully!!"))
}

//Retrieving TelecomServiceProvider given TSPId from Blockchain
/*
	fcnName: getTelecomServiceProvider
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [tspId]
	retrieve the tx from blockchain using GetState
	return pb.Response= Payload of TelecomServiceProvider which was created, else error saying does not exist
*/
func (c *Telco) getTelecomServiceProvider(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of Arguments. Expecting 1(TspId)")
	}
	tspId := args[0]
	valueAsBytes, err := stub.GetState(tspId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + tspId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"TSP does not exist: " + tspId + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valueAsBytes)
}

//Creating TeleMarketer in Blockchain
/*
	fcnName: setTelemarketer
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [tmId, tmName, tmAddress, tmState, tmDist, tmContactNum, tmEmail, tmDateOfRegistration,
	tmStatus, tmCIN]
	create a telemarketer object of the structure telemarketer,
	set all the necessary data of that object
	marshall the object into jsonObject
	create a tx in blockchain using PutState
	return pb.Response= "Telemarketer Created Successfully"
*/
func (c *Telco) setTelemarketer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 9 {
		return shim.Error("Incorrect Number of Arguments. Expecting 9")
	}
	loc, _ := time.LoadLocation("Asia/Kolkata")
	newTime := time.Now().In(loc)
	dt := newTime.Format("2006-01-02 15:04:05")
	tmId := args[0]
	tmName := args[1]
	tmAdd := args[2]
	tmState := args[3]
	tmDist := args[4]
	tmCNo := args[5]
	tmEmail := args[6]
	tmDor := dt
	tmStatus := args[7]
	tmCIN := args[8]
	// ==== Check if telemarketer already exists ====
	tmAsBytes, err := stub.GetState(tmId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + tmId + "\"}"
		return shim.Error(jsonResp)
	} else if tmAsBytes != nil {
		jsonResp = "{\"Error\":\"true\"}"
		return shim.Error(jsonResp)
	}
	tmStruct := &TeleMarketer{}

	tmStruct.TMId = tmId
	tmStruct.TMName = tmName
	tmStruct.TMAddress = tmAdd
	tmStruct.TMState = tmState
	tmStruct.TMDistrict = tmDist
	tmStruct.TMContactNo = tmCNo
	tmStruct.TMEmailId = tmEmail
	tmStruct.TMDateOfRegistration = tmDor
	tmStruct.TMStatus = tmStatus
	tmStruct.TMCIN = tmCIN

	tmAsBytes, err1 := json.Marshal(tmStruct)
	if err1 != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for TM \"}"
		return shim.Error(jsonResp)
	}
	err1 = stub.PutState(tmStruct.TMId, tmAsBytes)
	if err1 != nil {
		jsonResp = "{\"Error\":\"Creating TeleMarketers Data Failed\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("TeleMarketer Created Successfully!!"))
}

//Retrieving TeleMarketers given TmId from Blockchain
/*
	fcnName: getTelemarketer
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [tmId]
	retrieve the tx from blockchain using GetState
	return pb.Response= Payload of telemarkter which was created, else error saying does not exist
*/

func (c *Telco) getTelemarketer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of Arguments. Expecting 1(TmId)")
	}
	tmId := args[0]
	valueAsBytes, err := stub.GetState(tmId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + tmId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"TM does not exist: " + tmId + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valueAsBytes)
}

//Retrieve all Telemarketers from Blockchain
/*
	fcnName: getAllTelemarketers
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [empty]
	to know the total number of telemarketers in blockchain, using range query
	return pb.Response= Payload of all the telemarketers
*/
func (c *Telco) getAllTelemarketers(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	startKey := "TM100000000"
	endKey := "TM999999999"
	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(",\"Record\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return shim.Success(buffer.Bytes())
}

//Modify & Update the status of Telemarketer in Blockchain
/*
	fcnName: updateTelemarketerStatus
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [tmId, status]
	retrieve the telemarketer from blockchain given tmId
	create a telemarketer object from telemarketer structure
	unmarshal the retrieved data
	set the status
	marshal the data and create the data in blockchain using PutState
	return pb.Response= "Telemarketer Status Updated Successfully"
*/
func (c *Telco) updateTelemarketerStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 2 {
		return shim.Error("Incorrect number of Arguments. Expecting 2(TMId, Status)")
	}
	tmId := args[0]
	status := args[1]
	valueAsBytes, err := stub.GetState(tmId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + tmId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Telemarketer does not exist: " + tmId + "\"}"
		return shim.Error(jsonResp)
	}
	tmStruct := &TeleMarketer{}
	err = json.Unmarshal(valueAsBytes, &tmStruct)
	tmStruct.TMStatus = status
	tmAsBytes, err1 := json.Marshal(tmStruct)
	if err1 != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Telemarketer Status Updation \"}"
		return shim.Error(jsonResp)
	}
	err1 = stub.PutState(tmId, tmAsBytes)
	if err1 != nil {
		jsonResp = "{\"Error\":\"Updating Telemarketer Status Failed \"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Telemarketer Status Updated Successfully!!"))
}

//TM onboarding with TSP in Blockchain
/*
	fcnName: setTMOnboardingWithTSP
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [tspId, tmId, connectionId, connectionType, status, validity]
	create a onbaording object of the structure TMWithTSP,
	set all the necessary data of that object
	marshall the object into jsonObject
	create a tx in blockchain using PutState using compositeKey(tspId, tmId)
	return pb.Response= "Onboarding Created Successfully" if the telemarketer status is Active, else error
*/
func (c *Telco) setTMOnboardingWithTSP(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 6 {
		return shim.Error("Incorrect Number of Arguments. Expecting 6(TSPId, TMId, ConnectionId, ConnectionType, Status, Validity))")
	}
	tspId := args[0]
	tmId := args[1]
	connId := args[2]
	connType := args[3]
	status := args[4]
	validity := args[5]

	valueAsBytes, err := stub.GetState(tmId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + tmId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"TM does not exist: " + tmId + "\"}"
		return shim.Error(jsonResp)
	}
	tmStruct := &TeleMarketer{}
	err = json.Unmarshal(valueAsBytes, &tmStruct)

	if strings.Compare(tmStruct.TMStatus, "Active") == 0 {
		compKey := tspId + "#$#" + tmId
		tspTmStruct := &TMWithTSP{}

		tspTmStruct.TSPId = tspId
		tspTmStruct.TMId = tmId
		tspTmStruct.ConnectionId = connId
		tspTmStruct.ConnectionType = connType
		tspTmStruct.Status = status
		tspTmStruct.Validity = validity

		tspTmAsBytes, err := json.Marshal(tspTmStruct)
		if err != nil {
			jsonResp = "{\"Error\":\"JSON Marshalling Error for TSP Onboarding with TM \"}"
			return shim.Error(jsonResp)
		}
		err = stub.PutState(compKey, tspTmAsBytes)
		if err != nil {
			jsonResp = "{\"Error\":\"Creating TSP Oboarding with TM Data Failed\"}"
			return shim.Error(jsonResp)
		}
		return shim.Success([]byte("TSP Onboarding with TM Created Successfully!!"))
	} else {
		jsonResp = "{\"Error\":\"Telemarketer cannot be Onboarded\"}"
		return shim.Error(jsonResp)
	}

}

//Retrieving TM onboarding with TSP data given TSPId, TMId in Blockchain
/*
	fcnName: getTMOnboardingWithTSP
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [tspId, tmId]
	retrieve the tx from blockchain using GetState using compositekey
	return pb.Response= Payload of TMOnboardingTSP which was created, else error saying does not exist
*/
func (c *Telco) getTMOnboardingWithTSP(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 2 {
		return shim.Error("Incorrect number of Arguments. Expecting 2(TSPId, TMId)")
	}
	tspId := args[0]
	tmId := args[1]
	compKey := tspId + "#$#" + tmId
	valueAsBytes, err := stub.GetState(compKey)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + compKey + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"TSP Onboarding with TM does not exist: " + compKey + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valueAsBytes)
}

//Retrieving all the Telemarketers under TSP given TSPId from Blockchain
/*
	fcnName: getAllTMByTSP
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [TspId]
	to know the total number of telemarketers under TSP in blockchain
	return pb.Response= Payload of all the telemarketers under particular TSP
*/
func (c *Telco) getAllTMByTSP(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of Arguments. Expecting 1(TSPId)")
	}
	startKey := args[0] + "#$#" + "TM100000000"
	endKey := args[0] + "#$#" + "TM999999999"
	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		jsonResp = "{\"Error\":\"Could not get Telemarketers by TSP\"}"
		return shim.Error(jsonResp)
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(",\"Record\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return shim.Success(buffer.Bytes())
}

//Creating Template in Blockchain
/*
	fcnName: setTemplate
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [tmpId, tspId, tmId, name, type, category, entityId, body, status, validity]
	create a template object of the structure template,
	set all the necessary data of that object
	marshall the object into jsonObject
	create a tx in blockchain using PutState
	return pb.Response= "Template Created Successfully" if template doesn't exist already
*/
func (c *Telco) setTemplate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 10 {
		return shim.Error("Incorrect Number of Arguments. Expecting 10(TmpId, TSPId, TelemarketerId, Name, Type, Category, EntityId, Body, Status, Validity)")
	}
	loc, _ := time.LoadLocation("Asia/Kolkata")
	newTime := time.Now().In(loc)
	dt := newTime.Format("2006-01-02 15:04:05")
	tmpId := args[0]
	tspId := args[1]
	tmId := args[2]
	tmpName := args[3]
	tmpType := args[4]
	tmpCategory := args[5]
	tmpEntityId := args[6]
	tmpBody := args[7]
	tmpcreatedDate := dt
	tmpStatus := args[8]
	tmpValidity := args[9]

	// ==== Check if template already exists ====
	tempAsBytes, err := stub.GetState(tmpId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + tmpId + "\"}"
		return shim.Error(jsonResp)
	} else if tempAsBytes != nil {
		jsonResp = "{\"Error\":\"true\"}"
		return shim.Error(jsonResp)
	}
	tmpStruct := &Template{}
	tmpStruct.TemplateId = tmpId
	tmpStruct.TSPId = tspId
	tmpStruct.TelemarketerId = tmId
	tmpStruct.TemplateName = tmpName
	tmpStruct.TemplateType = tmpType
	tmpStruct.TemplateCategory = tmpCategory
	tmpStruct.TemplateEntityId = tmpEntityId
	tmpStruct.TemplateBody = tmpBody
	tmpStruct.TemplateCreatedDate = tmpcreatedDate
	tmpStruct.TemplateStatus = tmpStatus
	tmpStruct.TemplateValidity = tmpValidity

	tmpAsBytes, err := json.Marshal(tmpStruct)
	if err != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Template \"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutState(tmpStruct.TemplateId, tmpAsBytes)
	if err != nil {
		jsonResp = "{\"Error\":\"Creating Template Data Failed\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Template Created Successfully!!"))
}

//Retrieving Template given TemplateId from Blockchain
/*
	fcnName: getTemplate
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [tmpId]
	retrieve the tx from blockchain using GetState
	return pb.Response= Payload of template which was created, else error saying does not exist
*/
func (c *Telco) getTemplate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of Arguments. Expecting 1(TmpId)")
	}
	tmpId := args[0]
	valueAsBytes, err := stub.GetState(tmpId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + tmpId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Template does not exist: " + tmpId + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valueAsBytes)
}

//Modify & Update the status of Template in Blockchain
/*
	fcnName: modifyTemplateStatus
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [tmpId, status]
	retrieve the template from blockchain given templateId
	create a template object from template structure
	unmarshal the retrieved data
	set the status
	marshal the data and create the data in blockchain using PutState
	return pb.Response= "Template Status Updated Successfully"
*/
func (c *Telco) modifyTemplateStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 2 {
		return shim.Error("Incorrect number of Arguments. Expecting 2(TmpId, Status)")
	}
	tmpId := args[0]
	status := args[1]
	valueAsBytes, err := stub.GetState(tmpId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + tmpId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Template does not exist: " + tmpId + "\"}"
		return shim.Error(jsonResp)
	}
	tempStruct := &Template{}
	err = json.Unmarshal(valueAsBytes, &tempStruct)
	tempStruct.TemplateStatus = status
	tempAsBytes, err1 := json.Marshal(tempStruct)
	if err1 != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for TemplateUpdate \"}"
		return shim.Error(jsonResp)
	}
	err1 = stub.PutState(tmpId, tempAsBytes)
	if err1 != nil {
		jsonResp = "{\"Error\":\"Updating Template Status Failed\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Template Updated Successfully!!"))
}

//Retrieving all the Templates given EntityId from Blockchain
/*
	fcnName: queryTemplateByEntity
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [entityId]
	to know the total number of templates in blockchain, using range query(selector query)
	return pb.Response= Payload of all the templates under the entityId
*/
func (c *Telco) queryTemplateByEntity(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonresp string
	if len(args) != 1 {
		return shim.Error("incorrect number of arguments. expecting 1(entityid)")
	}
	entityid := args[0]
	querystring := fmt.Sprintf("{\"selector\":{\"tempEntityId\":\"%s\"}}", entityid)
	queryresults, err := getQueryResultForQueryString(stub, querystring)
	if err != nil {
		jsonresp = "{\"error\":\"failed to get state for " + entityid + "\"}"
		return shim.Error(jsonresp)
	}
	return shim.Success(queryresults)
}

//Retrieving all the Templates given Status from Blockchain
/*
	fcnName: queryTemplateByStatus
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [status]
	to know the total number of templates in blockchain, based on the status, using range query(selector query)
	return pb.Response= Payload of all the templates under the status
*/
func (c *Telco) queryTemplateByStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1(Status)")
	}
	status := args[0]
	queryString := fmt.Sprintf("{\"selector\":{\"tempStatus\":\"%s\"}}", status)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + status + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(queryResults)
}

//Creating Preference for a Subscriber in Blockchain
/*
	fcnName: setPreference
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [subNum, categories(10),modes(6), days(9), timebands(10)]
	timebands are segregated into slots(0-8)
	create a preference object of the structure preference,
	set all the necessary data of that object
	marshall the object into jsonObject
	create a tx in blockchain using PutState
	return pb.Response= "Preference Created Successfully"
*/
func (c *Telco) setPreference(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 36 {
		return shim.Error("Incorrect Number of Arguments. Expecting 36")
	}
	docType := "Preferences"
	subNum := args[0]

	all := args[1]
	blockPromo := args[2]
	financialServices := args[3]
	education := args[4]
	realEstate := args[5]
	health := args[6]
	consumerGoods := args[7]
	broadcasting := args[8]
	tourism := args[9]
	food := args[10]

	allModes := args[11]
	voiceCall := args[12]
	sms := args[13]
	adcPreRec := args[14]
	adcWithConn := args[15]
	roboCall := args[16]

	mon := args[17]
	tue := args[18]
	wed := args[19]
	thu := args[20]
	fri := args[21]
	sat := args[22]
	sun := args[23]
	allDays := args[24]
	holiday := args[25]

	allTimeBands := args[26]
	sl0 := args[27]
	sl1 := args[28]
	sl2 := args[29]
	sl3 := args[30]
	sl4 := args[31]
	sl5 := args[32]
	sl6 := args[33]
	sl7 := args[34]
	sl8 := args[35]

	preferenceStruct := &Preference{}
	preferenceStruct.DocType = docType
	preferenceStruct.SubscriberNumber = subNum

	preferenceStruct.Category.All = all
	preferenceStruct.Category.BlockPromo = blockPromo
	preferenceStruct.Category.FinancialServices = financialServices
	preferenceStruct.Category.Education = education
	preferenceStruct.Category.RealEstate = realEstate
	preferenceStruct.Category.Health = health
	preferenceStruct.Category.Consumergoods = consumerGoods
	preferenceStruct.Category.Broadcasting = broadcasting
	preferenceStruct.Category.Tourism = tourism
	preferenceStruct.Category.Food = food

	preferenceStruct.Mode.AllModes = allModes
	preferenceStruct.Mode.VoiceCall = voiceCall
	preferenceStruct.Mode.SMS = sms
	preferenceStruct.Mode.ADCPreRecorded = adcPreRec
	preferenceStruct.Mode.ADCWithConnectivity = adcWithConn
	preferenceStruct.Mode.RoboCall = roboCall

	preferenceStruct.Day.Monday = mon
	preferenceStruct.Day.Tuesday = tue
	preferenceStruct.Day.Wednesday = wed
	preferenceStruct.Day.Thursday = thu
	preferenceStruct.Day.Friday = fri
	preferenceStruct.Day.Saturday = sat
	preferenceStruct.Day.Sunday = sun
	preferenceStruct.Day.All = allDays
	preferenceStruct.Day.Holiday = holiday

	preferenceStruct.TimeBand.All = allTimeBands
	preferenceStruct.TimeBand.Slot0 = sl0
	preferenceStruct.TimeBand.Slot1 = sl1
	preferenceStruct.TimeBand.Slot2 = sl2
	preferenceStruct.TimeBand.Slot3 = sl3
	preferenceStruct.TimeBand.Slot4 = sl4
	preferenceStruct.TimeBand.Slot5 = sl5
	preferenceStruct.TimeBand.Slot6 = sl6
	preferenceStruct.TimeBand.Slot7 = sl7
	preferenceStruct.TimeBand.Slot8 = sl8

	preferenceAsBytes, err1 := json.Marshal(preferenceStruct)
	if err1 != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Preference Creation \"}"
		return shim.Error(jsonResp)
	}
	err1 = stub.PutState(preferenceStruct.SubscriberNumber, preferenceAsBytes)
	if err1 != nil {
		jsonResp = "{\"Error\":\"Creating Preference Data Failed \"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Preference Created Successfully!!"))
}

//Retrieving Preference given Subscriber Number from Blockchain
/*
	fcnName: getPreference
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [subscriberNumber]
	retrieve the tx from blockchain using GetState
	return pb.Response= Payload of preference which was created, else error saying does not exist
*/
func (c *Telco) getPreference(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of Arguments. Expecting 1(SubscriberNumber)")
	}
	subNum := args[0]
	valueAsBytes, err := stub.GetState(subNum)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + subNum + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Preference does not exist: " + subNum + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valueAsBytes)
}

//Retrieving all the Preferences for dashboard
/*
	fcnName: getNumberOfPreferences
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [empty]
	to know the total number of preferences in blockchain based on docType, using range query(selector query)
	return pb.Response= Payload of all the preferences
*/
func (c *Telco) getNumberOfPreferences(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	docType := "Preferences"
	queryString := fmt.Sprintf("{\"selector\":{\"preferenceDocType\":\"%s\"}}", docType)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	//	num := len(queryResults)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + docType + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(queryResults)
}


func (c *Telco) getNumberOfPreferencesWithCategory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	docType := "Preferences"
	cat1 := "true"
	queryString := fmt.Sprintf("{\"selector\":{\"preferenceDocType\":\"%s\",\"category.food\":\"%s\"}}", docType,cat1)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	//	num := len(queryResults)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + docType + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(queryResults)
}

//GetAll Preferences based on Pagination Testing
func (c *Telco) getPreferencesWithPagination(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect Number of Arguments.Expecting 1(Page Size)")
	}
	//	var jsonResp string
	docType := "Preferences"
	queryString := fmt.Sprintf("{\"selector\":{\"preferenceDocType\":\"%s\"}}", docType)
	pageSize, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return shim.Error(err.Error())
	}
	bookmark := ""
	queryResults, err1 := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookmark)
	if err1 != nil {
		return shim.Error(err1.Error())
	}
	return shim.Success(queryResults)
}

//Creating Consent in Blockchain
/*
	fcnName: setConsent
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [consentId, entityId, subscriberNo, commType, type, category, purpose, contentTempId]
	create a consent object of the structure consent,
	set all the necessary data of that object
	marshall the object into jsonObject
	create a tx in blockchain, using compkey(consentId, entityId) using PutState
	return pb.Response= "Consent Created Successfully"
*/
func (c *Telco) setConsent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 8 {
		return shim.Error("Incorrect Number of Arguments. Expecting 8")
	}
	docType := "MasterConsent"
	masterConId := args[0]
	entityId := args[1]
	subscriberNo := args[2]
	commType := args[3]
	typ := args[4]
	category := args[5]
	purpose := args[6]
	contentTemplateId := args[7]
	status := "Pending"

	compKey := entityId + "#$#" + masterConId
	conseStruct := &Consent{}
	conseStruct.DocType = docType
	conseStruct.MasterConsentId = masterConId
	conseStruct.EntityId = entityId
	conseStruct.SubscriberNo = subscriberNo
	conseStruct.CommunicationType = commType
	conseStruct.Type = typ
	conseStruct.Category = category
	conseStruct.Purpose = purpose
	conseStruct.ContentTemplateId = contentTemplateId
	conseStruct.Status = status //active
	consentAsBytes, err := json.Marshal(conseStruct)
	if err != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Consent Creation \"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutState(compKey, consentAsBytes)
	if err != nil {
		jsonResp = "{\"Error\":\"Creating Consent Data Failed \"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Consent Created Successfully!!"))
}

//Retrieving Consent given ConsentId and EntityId from Blockchain
/*
	fcnName: getConsentWithEntity
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [consentId, entityId]
	retrieve the tx from blockchain using GetState using compkey
	return pb.Response= Payload of consent which was created , else error saying does not exist
*/
func (c *Telco) getConsentWithEntity(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 2 {
		return shim.Error("Incorrect number of Arguments. Expecting 2(ConsentId, EntityId)")
	}
	consentId := args[0]
	entityId := args[1]

	compKey := entityId + "#$#" + consentId
	valueAsBytes, err := stub.GetState(compKey)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + compKey + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Consent does not exist: " + compKey + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valueAsBytes)
}

//Retrieving Consent data given SubscriberNumber from Blockchain
/*
	fcnName: getConsentWithSubscriberNumber
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [consentId, subscriberNum]
	retrieve the tx from blockchain using GetState using compkey
	return pb.Response= Payload of consent which was created , else error saying does not exist
*/
func (c *Telco) getConsentWithSubscriber(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 2 {
		return shim.Error("Incorrect number of Arguments. Expecting 2(ConsentId, SubscriberNumber)")
	}
	consentId := args[0]
	sNo := args[1]

	compKey := consentId + "#$#" + sNo
	valueAsBytes, err := stub.GetState(compKey)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + compKey + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Consent does not exist: " + compKey + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valueAsBytes)
}

//Creating a consent for the subscriber and changing the status done by ConsentRegistrar in Blockchain
/*
	fcnName: createConsentForSubscriberNo
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [entityId, consentId, status]
	retrieve the consent for that entityId.
	create a consent object of the structure consent,
	check the status, if the status is valid, then create a consent for all the subscriber numbers
	and update the status to valid, else consent cannot be created
	set all the necessary data of that object
	marshall the object into jsonObject
	create a tx in blockchain using PutState
	return pb.Response= "Consent Created Successfully, Updated Successfully"
*/
func (c *Telco) createConsentForSubscriberNo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 3 {
		return shim.Error("Incorrect Number of Arguments. Expecting 3(EntityId, ConsentId, Status)")
	}
	entityId := args[0]
	consentId := args[1]
	status := args[2]

	compKey := entityId + "#$#" + consentId
	valueAsBytes, err := stub.GetState(compKey)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + compKey + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Consent does not exist: " + compKey + "\"}"
		return shim.Error(jsonResp)
	}
	conseStruct := &Consent{}
	err = json.Unmarshal(valueAsBytes, &conseStruct)

	if strings.Compare(status, "Valid") == 0 {
		subscriberNo := conseStruct.SubscriberNo
		commType := conseStruct.CommunicationType
		typ := conseStruct.Type
		category := conseStruct.Category
		purpose := conseStruct.Purpose
		docType := "SubscriberConsent"
		contentTemplateId := conseStruct.ContentTemplateId
		subscriberNoTrimmed := strings.Trim(subscriberNo, "[ ]")
		subscriberNoReplaced := strings.Replace(subscriberNoTrimmed, ",", " ", -1)
		subscribersArray := strings.Fields(subscriberNoReplaced)
		for i := 0; i < len(subscribersArray); i++ {
			conseStruct := &Consent{}
			conseStruct.DocType = docType
			conseStruct.MasterConsentId = consentId
			conseStruct.EntityId = entityId
			conseStruct.SubscriberNo = subscribersArray[i]
			conseStruct.CommunicationType = commType
			conseStruct.Type = typ
			conseStruct.Category = category
			conseStruct.Purpose = purpose
			conseStruct.ContentTemplateId = contentTemplateId
			conseStruct.Status = "Pending"
			consentAsBytes, err := json.Marshal(conseStruct)
			compKey := conseStruct.MasterConsentId + "#$#" + subscribersArray[i]
			if err != nil {
				jsonResp = "{\"Error\":\"JSON Marshalling Error for Consent Creation \"}"
				return shim.Error(jsonResp)
			}
			err = stub.PutState(compKey, consentAsBytes)
			if err != nil {
				jsonResp = "{\"Error\":\"Creating Consent Data Failed \"}"
				return shim.Error(jsonResp)
			}
		}
		conseStruct.Status = status
		consentAsBytes, err1 := json.Marshal(conseStruct)
		if err1 != nil {
			jsonResp = "{\"Error\":\"JSON Marshalling Error for Consent Status Updation \"}"
			return shim.Error(jsonResp)
		}
		err1 = stub.PutState(compKey, consentAsBytes)
		if err1 != nil {
			jsonResp = "{\"Error\":\"Updating Consent Status Failed \"}"
			return shim.Error(jsonResp)
		}
	} else {
		conseStruct.Status = status
		consentAsBytes, err1 := json.Marshal(conseStruct)
		if err1 != nil {
			jsonResp = "{\"Error\":\"JSON Marshalling Error for Consent Status Updation \"}"
			return shim.Error(jsonResp)
		}
		err1 = stub.PutState(compKey, consentAsBytes)
		if err1 != nil {
			jsonResp = "{\"Error\":\"Updating Consent Status Failed \"}"
			return shim.Error(jsonResp)
		}
	}
	return shim.Success([]byte("Consent Status Updated Successfully!!"))
}

//Modify and Update the status of Consent status given consentId and subscriber Number in Blockchain
/*
	fcnName: updateConsentStatus
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [consentId, subscriberNumber, status]
	retrieve the consent from blockchain given consentId and subscriberNumber
	create a consent object from consent structure
	unmarshal the retrieved data
	set the status
	marshal the data and create the data in blockchain using PutState
	return pb.Response= "Consent Status Updated Successfully"
*/
func (c *Telco) updateConsentStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 3 {
		return shim.Error("Incorrect number of Arguments. Expecting 3(ConsentId, SubscriberNumber,Status)")
	}
	consentId := args[0]
	subNo := args[1]
	status := args[2]
	compKey := consentId + "#$#" + subNo
	valueAsBytes, err := stub.GetState(compKey)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + compKey + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Complaint does not exist: " + compKey + "\"}"
		return shim.Error(jsonResp)
	}
	conseStruct := &Consent{}
	err = json.Unmarshal(valueAsBytes, &conseStruct)
	conseStruct.Status = status
	consentAsBytes, err1 := json.Marshal(conseStruct)
	if err1 != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Consent Status Updation \"}"
		return shim.Error(jsonResp)
	}
	err1 = stub.PutState(compKey, consentAsBytes)
	if err1 != nil {
		jsonResp = "{\"Error\":\"Updating Consent Status Failed \"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Consent Status Updated Successfully!!"))
}

//Retrieve all Consents given SubscriberNumber from Blockchain
/*
	fcnName: getAllConsentsBySubscriberNumber
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [subscirberNumber]
	to know the total number of consents for a subscriber in blockchain, using range query(selector query)
	return pb.Response= Payload of all the consents for a particular subscriber
*/
func (c *Telco) queryAllConsentsBySubscriberNo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1(SubscriberNumber)")
	}
	docType := "SubscriberConsent"
	sNo := args[0]
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"%s\",\"subscriberNumber\":\"%s\"}}", docType, sNo)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + sNo + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(queryResults)
}

//Retrieving all the Consents given DocType and Status from Blockchain
/*
	fcnName: queryAllConsentsByDocTypeAndStatus
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [DocType and Status]
	retrieve all the consents from blockchain given status and doctype using range query(selector query)
	return pb.Response= Payload of all the consents
*/
func (c *Telco) queryAllConsentsByDocTypeAndStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2(DocType, Status)")
	}
	docType := args[0]
	status := args[1]
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"%s\",\"consentStatus\":\"%s\"}}", docType, status)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + docType + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(queryResults)
}

//Retrieving all the consents given EntityId from Blockchain
/*
	fcnName: queryAllConsentsByEntityId
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [EntityId]
	retrieve all the consents from blockchain given entityId using range query(selector query)
	return pb.Response= Payload of all the consents
*/
func (c *Telco) queryAllConsentsByEntityId(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1(EntityId)")
	}
	startKey := args[0] + "#$#" + "CON100000000"
	endKey := args[0] + "#$#" + "CON999999999"
	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(",\"Record\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return shim.Success(buffer.Bytes())
}

//Creating Header in Blockchain
/*
	fcnName: setHeader
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [HeaderName, tspId, tmId, headerEntityId, type, category, status, createdDate,
	modifiedDate, validity]
	create a header object of the structure header,
	set all the necessary data of that object
	marshall the object into jsonObject
	create a tx in blockchain using PutState
	return pb.Response= "Header Created Successfully" if header doesn't exist already
*/
func (c *Telco) setHeader(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string

	if len(args) != 7 {
		return shim.Error("Incorrect Number of Arguments. Expecting 7(HeaderName, TSPId, TelemarketerId, EntityId, Type, Category, Validity)")
	}
	loc, _ := time.LoadLocation("Asia/Kolkata")
	newTime := time.Now().In(loc)
	dt := newTime.Format("2006-01-02 15:04:05")

	headerName := args[0]
	tspId := args[1]
	tmId := args[2]
	headerEntityId := args[3]
	headerType := args[4]
	headerCategory := args[5]
	headerStatus := "Pending"
	headerCreatedDate := dt
	headerModifiedDate := dt
	headerValidity := args[6]

	//==== Check if header already exists ====
	headerAsBytes, err := stub.GetState(headerName)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + headerName + "\"}"
		return shim.Error(jsonResp)
	} else if headerAsBytes != nil {
		jsonResp = "{\"Error\":\"true\"}"
		return shim.Error(jsonResp)
	}

	headerStruct := &Header{}

	headerStruct.HeaderName = headerName
	headerStruct.TSPId = tspId
	headerStruct.TelemarketerId = tmId
	headerStruct.HeaderEntityId = headerEntityId
	headerStruct.HeaderType = headerType
	headerStruct.HeaderCategory = headerCategory
	headerStruct.HeaderStatus = headerStatus
	headerStruct.HeaderCreatedDate = headerCreatedDate
	headerStruct.HeaderModifiedDate = headerModifiedDate
	headerStruct.HeaderValidity = headerValidity

	headerAsBytes, err1 := json.Marshal(headerStruct)
	if err1 != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Header \"}"
		return shim.Error(jsonResp)
	}
	err1 = stub.PutState(headerStruct.HeaderName, headerAsBytes)
	if err1 != nil {
		jsonResp = "{\"Error\":\"Creating Header Data Failed\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Header Created Successfully!!"))
}

//Retrieving Header given HeaderName from Blockchain
/*
	fcnName: getHeader
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [headerName]
	retrieve the tx from blockchain using GetState
	return pb.Response= Payload of header which was created, else error saying does not exist
*/
func (c *Telco) getHeader(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of Arguments. Expecting 1(HeaderName)")
	}
	headerName := args[0]
	valueAsBytes, err := stub.GetState(headerName)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + headerName + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Header does not exist: " + headerName + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valueAsBytes)
}

//Modify & Update the status of Header given HeaderName and Status in Blockchain
/*
	fcnName: modifyHeaderStatus
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [headerName, status]
	retrieve the header from blockchain given headerName
	create a header object from header structure
	unmarshal the retrieved data
	set the status
	marshal the data and create the data in blockchain using PutState
	return pb.Response= "Header Status Updated Successfully"
*/
func (c *Telco) modifyHeaderStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 2 {
		return shim.Error("Incorrect number of Arguments. Expecting 2(HeaderName, Status)")
	}
	headerName := args[0]
	status := args[1]
	loc, _ := time.LoadLocation("Asia/Kolkata")
	newTime := time.Now().In(loc)
	dt := newTime.Format("2006-01-02 15:04:05")
	valueAsBytes, err := stub.GetState(headerName)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + headerName + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Header does not exist: " + headerName + "\"}"
		return shim.Error(jsonResp)
	}
	headerStruct := &Header{}
	err = json.Unmarshal(valueAsBytes, &headerStruct)
	headerStruct.HeaderStatus = status
	headerStruct.HeaderModifiedDate = dt
	headerAsBytes, err1 := json.Marshal(headerStruct)
	if err1 != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for HeaderUpdate \"}"
		return shim.Error(jsonResp)
	}
	err1 = stub.PutState(headerName, headerAsBytes)
	if err1 != nil {
		jsonResp = "{\"Error\":\"Updating Header Status Failed\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Header Updated Successfully!!"))
}

//Retrieving all the Headers given EntityId from Blockchain
/*
	fcnName: queryHeaderByEntity
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [entityId]
	to know the total number of headers under entity in blockchain, using range query(selector query)
	return pb.Response= Payload of all the headers under entity
*/

func (c *Telco) queryHeaderByEntity(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1(EntityId)")
	}
	entityId := args[0]
	queryString := fmt.Sprintf("{\"selector\":{\"headerEntityId\":\"%s\"}}", entityId)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + entityId + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(queryResults)
}

//Retrieving all the Headers given Status from Blockchain
/*
	fcnName: queryHeaderByStatus
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [status]
	to know the total number of headers given status in blockchain, using range query(selector query)
	return pb.Response= Payload of all the headers based on the status
*/

func (c *Telco) queryHeaderByStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1(Status)")
	}
	status := args[0]
	queryString := fmt.Sprintf("{\"selector\":{\"headerStatus\":\"%s\"}}", status)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + status + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(queryResults)
}

//Used to find whether a given time is in between two times or not
func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

//Transactional Scrubbing
/*
	fcnName: transactionalScrubbing
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [subscribersList, consentId]
	for all the subscribers check for the consent, if exists then check for the status to be approved
	return pb.Response= Payload of all the unblocked numbers
*/
func (c *Telco) transactionalScrubbing(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect Number of Arguments. Expecting 2(SubscribersList, ConsentId)")
	}
	var unBlockedList []string
	var jsonResp string
	subscribersList := args[0]
	consentId := args[1]

	subscriberNoTrimmed := strings.Trim(subscribersList, "[ ]")
	subscriberNoReplaced := strings.Replace(subscriberNoTrimmed, ",", " ", -1)
	subscribersArray := strings.Fields(subscriberNoReplaced)

	for i := 0; i < len(subscribersArray); i++ {
		valueAsBytes, err := stub.GetState(consentId + "#$#" + subscribersArray[i])
		if err != nil {
			jsonResp = "{\"Error\":\"Failed to get state for " + subscribersArray[i] + "\"}"
			fmt.Println(jsonResp)
		} else if valueAsBytes == nil {
			jsonResp = "{\"Error\" : \"Consent does not exist: " + subscribersArray[i] + "\"}"
			fmt.Println(jsonResp)
		} else {
			consentStruct := &Consent{}
			err = json.Unmarshal(valueAsBytes, &consentStruct)
			statusConsent := consentStruct.Status
			if strings.Compare(statusConsent, "Approved") == 0 {
				unBlockedList = append(unBlockedList, subscribersArray[i])
			}
		}
	}
	uBl, _ := json.Marshal(unBlockedList)
	final := "[" + string(uBl) + "]"
	return shim.Success([]byte(final))
}

//Promotional Scrubbing
/*
	fcnName: promotionalScrubbing
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [subscribersList, templateId]
	for all the subscribers check for the preference, if exists then check for categories, days, timebands,
	if any one of the values are set to be true then add that subscriber to the blockList,
	else add that subscriber to the unblocked list, if preference does not exist then add that user to unblockedList
	return pb.Response= Payload of all the unblocked numbers
*/
func (c *Telco) promotionalScrubbing(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect Number of Arguments. Expecting 2(SubscribersList, TemplateId)")
	}
	var blockedList, unBlockedList []string
	var slot, jsonResp string
	loc, _ := time.LoadLocation("Asia/Kolkata")
	newTime := time.Now().In(loc)
	dt := newTime.Format("15:04")
	day := newTime.Weekday().String()

	subscribersList := args[0]
	templateId := args[1]

	currentTime, _ := time.Parse("15:04", dt)
	slot0Min, _ := time.Parse("15:04", "00:00")
	slot0Max, _ := time.Parse("15:04", "05:59")

	slot1Min, _ := time.Parse("15:04", "06:00")
	slot1Max, _ := time.Parse("15:04", "07:59")

	slot2Min, _ := time.Parse("15:04", "08:00")
	slot2Max, _ := time.Parse("15:04", "09:59")

	slot3Min, _ := time.Parse("15:04", "10:00")
	slot3Max, _ := time.Parse("15:04", "11:59")

	slot4Min, _ := time.Parse("15:04", "12:00")
	slot4Max, _ := time.Parse("15:04", "13:59")

	slot5Min, _ := time.Parse("15:04", "14:00")
	slot5Max, _ := time.Parse("15:04", "15:59")

	slot6Min, _ := time.Parse("15:04", "16:00")
	slot6Max, _ := time.Parse("15:04", "17:59")

	slot7Min, _ := time.Parse("15:04", "18:00")
	slot7Max, _ := time.Parse("15:04", "20:59")

	slot8Min, _ := time.Parse("15:04", "21:00")
	slot8Max, _ := time.Parse("15:04", "24:00")

	if inTimeSpan(slot0Min, slot0Max, currentTime) {
		slot = "Slot0"
	} else if inTimeSpan(slot1Min, slot1Max, currentTime) {
		slot = "Slot1"
	} else if inTimeSpan(slot2Min, slot2Max, currentTime) {
		slot = "Slot2"
	} else if inTimeSpan(slot3Min, slot3Max, currentTime) {
		slot = "Slot3"
	} else if inTimeSpan(slot4Min, slot4Max, currentTime) {
		slot = "Slot4"
	} else if inTimeSpan(slot5Min, slot5Max, currentTime) {
		slot = "Slot5"
	} else if inTimeSpan(slot6Min, slot6Max, currentTime) {
		slot = "Slot6"
	} else if inTimeSpan(slot7Min, slot7Max, currentTime) {
		slot = "Slot7"
	} else if inTimeSpan(slot8Min, slot8Max, currentTime) {
		slot = "Slot8"
	} else {
		slot = "All"
	}

	valueAsBytes, err := stub.GetState(templateId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + templateId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Template does not exist: " + templateId + "\"}"
		return shim.Error(jsonResp)
	}

	subscriberNoTrimmed := strings.Trim(subscribersList, "[ ]")
	subscriberNoReplaced := strings.Replace(subscriberNoTrimmed, ",", " ", -1)
	subscribersArray := strings.Fields(subscriberNoReplaced)

	template := &Template{}
	err = json.Unmarshal(valueAsBytes, &template)
	category := template.TemplateCategory
	for i := 0; i < len(subscribersArray); i++ {
		valueAsBytes, err := stub.GetState(subscribersArray[i])
		if err != nil {
			unBlockedList = append(unBlockedList, subscribersArray[i])
		} else if valueAsBytes == nil {
			unBlockedList = append(unBlockedList, subscribersArray[i])
		} else {
			preference := &Preference{}
			err = json.Unmarshal(valueAsBytes, &preference)
			categoryStruct := preference.Category
			catNum := reflect.ValueOf(categoryStruct)
			catValue := reflect.Indirect(catNum).FieldByName(category).String()
			if strings.Compare(catValue, "true") == 0 {
				blockedList = append(blockedList, subscribersArray[i])
			} else if strings.Compare(catValue, "true") != 0 {
				dayStruct := preference.Day
				dayNum := reflect.ValueOf(dayStruct)
				dayValue := reflect.Indirect(dayNum).FieldByName(day).String()
				if strings.Compare(dayValue, "true") == 0 {
					blockedList = append(blockedList, subscribersArray[i])
				} else if strings.Compare(dayValue, "true") != 0 {
					timeBandStruct := preference.TimeBand
					timeNum := reflect.ValueOf(timeBandStruct)
					timeValue := reflect.Indirect(timeNum).FieldByName(slot).String()
					if strings.Compare(timeValue, "true") == 0 {
						blockedList = append(blockedList, subscribersArray[i])
					} else if strings.Compare(timeValue, "true") != 0 {
						unBlockedList = append(unBlockedList, subscribersArray[i])
					}
				}
			}
		}
	}
	uBl, _ := json.Marshal(unBlockedList)
	final := "[" + string(uBl) + "]"
	return shim.Success([]byte(final))
}

//Service Scrubbing
/*
	fcnName: serviceScrubbing
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [subscribersList, templateId, consentId]
	for all the subscribers check for the consent, if consent exists then check for the status to be approved, if approved
	then add to the unblockedList, else check for the preference, if exists then check for categories, days, timebands,
	if any one of the values are set to be true then add that subscriber to the blockList,
	else add that subscriber to the unblocked list, if preference does not exist then add that user to unblockedList
	return pb.Response= Payload of all the unblocked numbers
*/
func (c *Telco) serviceScrubbing(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect Number of Arguments. Expecting 3(Subscribers List, TemplateId, ConsentId)")
	}
	var blockedList, unBlockedList []string
	var jsonResp, slot string
	subscribersList := args[0]
	templateId := args[1]
	consentId := args[2]

	loc, _ := time.LoadLocation("Asia/Kolkata")
	newTime := time.Now().In(loc)
	dt := newTime.Format("15:04")
	day := newTime.Weekday().String()

	currentTime, _ := time.Parse("15:04", dt)
	slot0Min, _ := time.Parse("15:04", "00:00")
	slot0Max, _ := time.Parse("15:04", "05:59")

	slot1Min, _ := time.Parse("15:04", "06:00")
	slot1Max, _ := time.Parse("15:04", "07:59")

	slot2Min, _ := time.Parse("15:04", "08:00")
	slot2Max, _ := time.Parse("15:04", "07:59")

	slot3Min, _ := time.Parse("15:04", "10:00")
	slot3Max, _ := time.Parse("15:04", "11:59")

	slot4Min, _ := time.Parse("15:04", "12:00")
	slot4Max, _ := time.Parse("15:04", "13:59")

	slot5Min, _ := time.Parse("15:04", "14:00")
	slot5Max, _ := time.Parse("15:04", "15:59")

	slot6Min, _ := time.Parse("15:04", "16:00")
	slot6Max, _ := time.Parse("15:04", "17:59")

	slot7Min, _ := time.Parse("15:04", "18:00")
	slot7Max, _ := time.Parse("15:04", "20:59")

	slot8Min, _ := time.Parse("15:04", "21:00")
	slot8Max, _ := time.Parse("15:04", "24:00")

	if inTimeSpan(slot0Min, slot0Max, currentTime) {
		slot = "Slot0"
	} else if inTimeSpan(slot1Min, slot1Max, currentTime) {
		slot = "Slot1"
	} else if inTimeSpan(slot2Min, slot2Max, currentTime) {
		slot = "Slot2"
	} else if inTimeSpan(slot3Min, slot3Max, currentTime) {
		slot = "Slot3"
	} else if inTimeSpan(slot4Min, slot4Max, currentTime) {
		slot = "Slot4"
	} else if inTimeSpan(slot5Min, slot5Max, currentTime) {
		slot = "Slot5"
	} else if inTimeSpan(slot6Min, slot6Max, currentTime) {
		slot = "Slot6"
	} else if inTimeSpan(slot7Min, slot7Max, currentTime) {
		slot = "Slot7"
	} else if inTimeSpan(slot8Min, slot8Max, currentTime) {
		slot = "Slot8"
	} else {
		slot = "All"
	}
	valueAsBytes, err := stub.GetState(templateId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + templateId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Template does not exist: " + templateId + "\"}"
		return shim.Error(jsonResp)
	}
	template := &Template{}
	err = json.Unmarshal(valueAsBytes, &template)
	category := template.TemplateCategory
	subscriberNoTrimmed := strings.Trim(subscribersList, "[ ]")
	subscriberNoReplaced := strings.Replace(subscriberNoTrimmed, ",", " ", -1)
	subscribersArray := strings.Fields(subscriberNoReplaced)

	for i := 0; i < len(subscribersArray); i++ {
		valueAsBytes, err := stub.GetState(consentId + "#$#" + subscribersArray[i])
		if err != nil {
			jsonResp = "{\"Error\":\"Failed to get state for " + subscribersArray[i] + "\"}"
			prefAsBytes, err1 := stub.GetState(subscribersArray[i])
			if err1 != nil {
				unBlockedList = append(unBlockedList, subscribersArray[i])
			} else if prefAsBytes == nil {
				unBlockedList = append(unBlockedList, subscribersArray[i])
			} else {
				preference := &Preference{}
				err = json.Unmarshal(valueAsBytes, &preference)
				categoryStruct := preference.Category
				catNum := reflect.ValueOf(categoryStruct)
				catValue := reflect.Indirect(catNum).FieldByName(category).String()
				if strings.Compare(catValue, "true") == 0 {
					blockedList = append(blockedList, subscribersArray[i])
				} else if strings.Compare(catValue, "true") != 0 {
					dayStruct := preference.Day
					dayNum := reflect.ValueOf(dayStruct)
					dayValue := reflect.Indirect(dayNum).FieldByName(day).String()
					if strings.Compare(dayValue, "true") == 0 {
						blockedList = append(blockedList, subscribersArray[i])
					} else if strings.Compare(dayValue, "true") != 0 {
						timeBandStruct := preference.TimeBand
						timeNum := reflect.ValueOf(timeBandStruct)
						timeValue := reflect.Indirect(timeNum).FieldByName(slot).String()
						if strings.Compare(timeValue, "true") == 0 {
							blockedList = append(blockedList, subscribersArray[i])
						} else if strings.Compare(timeValue, "true") != 0 {
							unBlockedList = append(unBlockedList, subscribersArray[i])
						}
					}
				}
			}
		} else if valueAsBytes == nil {
			jsonResp = "{\"Error\" : \"Consent does not exist: " + subscribersArray[i] + "\"}"
			prefAsBytes, err1 := stub.GetState(subscribersArray[i])
			if err1 != nil {
				unBlockedList = append(unBlockedList, subscribersArray[i])
			} else if prefAsBytes == nil {
				unBlockedList = append(unBlockedList, subscribersArray[i])
			} else {
				preference := &Preference{}
				err = json.Unmarshal(valueAsBytes, &preference)
				categoryStruct := preference.Category
				catNum := reflect.ValueOf(categoryStruct)
				catValue := reflect.Indirect(catNum).FieldByName(category).String()
				if strings.Compare(catValue, "true") == 0 {
					blockedList = append(blockedList, subscribersArray[i])
				} else if strings.Compare(catValue, "true") != 0 {
					dayStruct := preference.Day
					dayNum := reflect.ValueOf(dayStruct)
					dayValue := reflect.Indirect(dayNum).FieldByName(day).String()
					if strings.Compare(dayValue, "true") == 0 {
						blockedList = append(blockedList, subscribersArray[i])
					} else if strings.Compare(dayValue, "true") != 0 {
						timeBandStruct := preference.TimeBand
						timeNum := reflect.ValueOf(timeBandStruct)
						timeValue := reflect.Indirect(timeNum).FieldByName(slot).String()
						if strings.Compare(timeValue, "true") == 0 {
							blockedList = append(blockedList, subscribersArray[i])
						} else if strings.Compare(timeValue, "true") != 0 {
							unBlockedList = append(unBlockedList, subscribersArray[i])
						}
					}
				}
			}
		} else {
			consentStruct := &Consent{}
			err = json.Unmarshal(valueAsBytes, &consentStruct)
			statusConsent := consentStruct.Status
			if strings.Compare(statusConsent, "Approved") == 0 {
				unBlockedList = append(unBlockedList, subscribersArray[i])
			}
		}
	}
	uBl, _ := json.Marshal(unBlockedList)
	final := "[" + string(uBl) + "]"
	return shim.Success([]byte(final))
}

//Creating a Campaign in Blockchain
/*
	fcnName: createCampaign
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [campaignId, consentId, templateId, headerName, status, inputHash, outputHash]
	query for the template and header if exists check for the status of both template and header should be approved,then
	create a campaign in blockchain
	set all the necessary data of that object
	marshall the object into jsonObject
	create a tx in blockchain using PutState
	return pb.Response= "Campaign Created Successfully"
*/
func (c *Telco) createCampaign(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 7 {
		return shim.Error("Incorrect Number of Arguments. Expecting 7")
	}
	campaignId := args[0]
	consentId := args[1]
	templateId := args[2]
	headerName := args[3]
	status := args[4]
	inputHash := args[5]
	outputHash := args[6]

	loc, _ := time.LoadLocation("Asia/Kolkata")
	newTime := time.Now().In(loc)
	dateTime := newTime.Format("2006-01-02 15:04:05")
	valueAsBytes, err := stub.GetState(templateId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + templateId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Template does not exist: " + templateId + "\"}"
		return shim.Error(jsonResp)
	}

	template := &Template{}
	err = json.Unmarshal(valueAsBytes, &template)
	entityId := template.TemplateEntityId
	templateStatus := template.TemplateStatus

	headAsBytes, err := stub.GetState(headerName)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + headerName + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Header does not exist: " + headerName + "\"}"
		return shim.Error(jsonResp)
	}

	header := &Header{}
	err = json.Unmarshal(headAsBytes, &header)
	headerStatus := header.HeaderStatus

	if strings.Compare(headerStatus, "Approved") == 0 && strings.Compare(templateStatus, "Approved") == 0 {
		scrubStruct := &Scrubbing{}
		scrubStruct.CampaignId = campaignId
		scrubStruct.ConsentId = consentId
		scrubStruct.TemplateId = templateId
		scrubStruct.HeaderName = headerName
		scrubStruct.EntityId = entityId
		scrubStruct.DateTime = dateTime
		scrubStruct.Status = status
		scrubStruct.InputHash = inputHash
		scrubStruct.OutputHash = outputHash
		scrubAsBytes, err1 := json.Marshal(scrubStruct)
		if err1 != nil {
			jsonResp = "{\"Error\":\"JSON Marshalling Error for Campaign Creation Transactional \"}"
			return shim.Error(jsonResp)
		}
		err1 = stub.PutState(scrubStruct.CampaignId, scrubAsBytes)
		if err1 != nil {
			jsonResp = "{\"Error\":\"Creating Campaign Data Transactional Failed\"}"
			return shim.Error(jsonResp)
		}
		return shim.Success([]byte("Campaign Created Successfully."))
	} else {
		jsonResp := "{\"Error\":\"Scrubbing not Possible, Because Header or Template is Not Approved\"}"
		return shim.Error(jsonResp)
	}
}

//Retrieving the Campaign given CampaignId from Blockchain
/*
	fcnName: getScrubbingStatus
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [campaignId]
	retrieve the tx from blockchain using GetState
	return pb.Response= Payload of campaign which was created, else error saying does not exist
*/
func (c *Telco) getScrubbingStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of Arguments. Expecting 1(CampaignId)")
	}
	campaignId := args[0]
	valueAsBytes, err := stub.GetState(campaignId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + campaignId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Campaign does not exist: " + campaignId + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valueAsBytes)
}

//Modify & Update the Output Hash of Campaign in Blockchain
/*
	fcnName: updateScrubOutputHash
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [campaignId, outputshash]
	retrieve the campaign from blockchain given campaignId
	create a campaign object from campaign structure
	unmarshal the retrieved data
	set the outputHash
	marshal the data and create the data in blockchain using PutState
	return pb.Response= "Campaign Hash Updated Successfully"
*/
func (c *Telco) updateScrubOutputHash(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 2 {
		return shim.Error("Incorrect number of Arguments. Expecting 2(CampaignId, OutputHash)")
	}
	campaignId := args[0]
	outputHash := args[1]

	loc, _ := time.LoadLocation("Asia/Kolkata")
	newTime := time.Now().In(loc)
	dateTime := newTime.Format("2006-01-02 15:04:05")

	valueAsBytes, err := stub.GetState(campaignId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + campaignId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Campaign does not exist: " + campaignId + "\"}"
		return shim.Error(jsonResp)
	}
	scrubStruct := &Scrubbing{}
	err = json.Unmarshal(valueAsBytes, &scrubStruct)
	scrubStruct.OutputHash = outputHash
	scrubStruct.DateTime = dateTime
	scrubAsBytes, err1 := json.Marshal(scrubStruct)
	if err1 != nil {
		return shim.Error(err1.Error())
	}
	err1 = stub.PutState(campaignId, scrubAsBytes)
	if err1 != nil {
		jsonResp = "{\"Error\":\"Updating Campaign OutputHash Data Failed\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Campaign Hash Updated Successfully!!"))
}

//Retrieving all the campaigns from Blockchain
/*
	fcnName: getAllCampaigns
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [empty]
	retrieve all the campaigns who are under the from blockchain using range query(selector query)
	return pb.Response= Payload of all the campaigns.
*/
func (c *Telco) getAllCampaigns(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	startKey := "C100000000"
	endKey := "C999999999"
	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(",\"Record\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return shim.Success(buffer.Bytes())
}

//Finding the difference between the two dates in days.
func getDays(start, end time.Time) int {
	return int(start.Sub(end).Hours() / 24)
}

//Creating Complaint in blockchain
/*
	fcnName: raiseComplaint
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [subscribernum, URNNo, date, time, UCCDate, headerName, tap, UCCOAPdate,
	UCCType, UCCDesc, CDRStatus, ActionTake, Remarks]
	check for the difference between dates (complaintdate and UCC date), if the difference is <=3 then set the type as complaint
	else >3 then set type as Report and set the status to closed
	get tspId given headerId
	create complaint object from the complaint structure
	set all the necessary data of that object
	marshall the object into jsonObject
	create a tx in blockchain using PutState
	return pb.Response= "Complaint Created Successfully"
*/
func (c *Telco) raiseComplaint(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp, comStatus, comType string
	if len(args) != 12 {
		return shim.Error("Incorrect Number of Arguements.Expecting 12")
	}
	loc, _ := time.LoadLocation("Asia/Kolkata")
	newTime := time.Now().In(loc)
	date := newTime.Format("2006-01-02")
	timeStamp := newTime.Format("15:04")

	docType := "Complaints"
	comSubNo := args[0]
	comURNNo := args[1]
	comDate := date
	comTime := timeStamp
	comUCCDate := args[2]
	comHeaderName := args[3]
	comTAPCode := args[4]
	comUCCOAPDate := args[5]
	comUCCType := args[6]
	comUCCDesc := args[7]
	comCDRStatus := args[8]
	comActionTaken := args[10]
	comRemarks := args[11]

	//check the difference between complaint date and ucc date
	//if the difference is >= 3 set the type =report and status = closed
	compDate, _ := time.Parse("2006-01-02", comDate)
	uccDate, _ := time.Parse("2006-01-02", comUCCDate)

	if getDays(compDate, uccDate) <= 3 {
		comType = "Complaint"
		comStatus = args[9]
	} else if getDays(compDate, uccDate) > 3 {
		comType = "Report"
		comStatus = "Closed"
	}
	//get tspId from the Header given Header Id
	headerAsBytes, err := stub.GetState(comHeaderName)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + comHeaderName + "\"}"
		return shim.Error(jsonResp)
	} else if headerAsBytes == nil {
		jsonResp = "{\"Error\" : \"Header does not exist: " + comHeaderName + "\"}"
		return shim.Error(jsonResp)
	}
	headerStruct := &Header{}
	err = json.Unmarshal(headerAsBytes, &headerStruct)
	comOAP := headerStruct.TSPId

	compKey := comSubNo + "#$#" + comURNNo

	complStruct := &Complaint{}
	complStruct.DocType = docType
	complStruct.SubscriberNo = comSubNo
	complStruct.UniqueReferenceNo = comURNNo
	complStruct.ComplaintDate = comDate
	complStruct.ComplaintTime = comTime
	complStruct.UCCDate = comUCCDate
	complStruct.OAP = comOAP
	complStruct.HeaderName = comHeaderName
	complStruct.TAPCode = comTAPCode
	complStruct.UCCOAPDate = comUCCOAPDate
	complStruct.UCCType = comUCCType
	complStruct.UCCDescription = comUCCDesc
	complStruct.CDRStatus = comCDRStatus
	complStruct.Status = comStatus
	complStruct.ActionTaken = comActionTaken
	complStruct.Remarks = comRemarks
	complStruct.ComplaintType = comType

	complaintAsBytes, err := json.Marshal(complStruct)
	if err != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Raising Complaint \"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutState(compKey, complaintAsBytes)
	if err != nil {
		jsonResp = "{\"Error\":\"Creating Complaint Failed \"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Complaint Created Successfully!!"))
}

//Retrieving the Complaint given ComplaintId from Blockchain
/*
	fcnName: fetchComplaint
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [subscriberNum, URN number]
	retrieve the tx from blockchain using GetState
	return pb.Response= Payload of complaint which was created, else error saying does not exist
*/
func (c *Telco) fetchComplaint(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 2 {
		return shim.Error("Incorrect number of Arguments. Expecting 2(SubscriberNo, UniqueReferenceNumber)")
	}
	subscriberNo := args[0]
	urNumber := args[1]
	compKey := subscriberNo + "#$#" + urNumber
	valueAsBytes, err := stub.GetState(compKey)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + compKey + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Complaint does not exist: " + compKey + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valueAsBytes)
}

//Retrieving all the Complaint given SubscriberNumber from Blockchain
/*
	fcnName: getAllComplaintsBySunscriberNumber
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [subscriberNumber]
	to know the total number of complaints given subscriber number in blockchain
	return pb.Response= Payload of all the complaints given subscriberNumber
*/
func (c *Telco) getAllComplaintsBySubscriberNumber(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect Number of Arguments. Expecting 1(SubscriberNumber)")
	}
	startKey := args[0] + "#$#" + "COM100000000"
	endKey := args[0] + "#$#" + "COM999999999"
	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(",\"Record\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return shim.Success(buffer.Bytes())
}

//Modify & Update the status of CDR in Blockchain
/*
	fcnName: updateCDRStatus
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [subNo, URNno, cdrstatus, remarks]
	retrieve the complaint from blockchain given subNum and URNno
	create a complaint object from complaint structure
	unmarshal the retrieved data
	set the cdrstatus, remarks
	marshal the data and create the data in blockchain using PutState
	return pb.Response= "Complaints CDRStatus Updated Successfully"
*/
func (c *Telco) updateCDRStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 4 {
		return shim.Error("Incorrect number of Arguments. Expecting 4(SubscriberNo,UniqueReferenceNumber,CDRStatus , Remarks)")
	}
	subscriberNo := args[0]
	urNum := args[1]
	status := args[2]
	remarks := args[3]
	compKey := subscriberNo + "#$#" + urNum
	valueAsBytes, err := stub.GetState(compKey)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + compKey + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Complaint does not exist: " + compKey + "\"}"
		return shim.Error(jsonResp)
	}
	complStruct := &Complaint{}
	err = json.Unmarshal(valueAsBytes, &complStruct)
	complStruct.CDRStatus = status
	complStruct.Remarks = remarks
	complaintAsBytes, err1 := json.Marshal(complStruct)
	if err1 != nil {
		return shim.Error(err1.Error())
	}
	err1 = stub.PutState(compKey, complaintAsBytes)
	if err1 != nil {
		return shim.Error("Error while Updating Complaint Data.")
	}
	return shim.Success([]byte("Complaint CDR Status Updated Successfully!!"))
}

//Modify & Update the status of Complaint in Blockchain
/*
	fcnName: updateComplaintStatus
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [subNo, URNno, status, remarks, action]
	retrieve the complaint from blockchain given subNum and URNno
	create a complaint object from complaint structure
	unmarshal the retrieved data
	set the cdrstatus, remarks, actions
	marshal the data and create the data in blockchain using PutState
	return pb.Response= "Complaints Status Updated Successfully"
*/
func (c *Telco) updateComplaintStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 5 {
		return shim.Error("Incorrect number of Arguments. Expecting 5(SubscriberNo,UniqueReferenceNumber,Status, Remarks, Action)")
	}
	subscriberNo := args[0]
	urNum := args[1]
	status := args[2]
	remarks := args[3]
	action := args[4]
	compKey := subscriberNo + "#$#" + urNum
	valueAsBytes, err := stub.GetState(compKey)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + compKey + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Complaint does not exist: " + compKey + "\"}"
		return shim.Error(jsonResp)
	}
	complStruct := &Complaint{}
	err = json.Unmarshal(valueAsBytes, &complStruct)
	complStruct.Status = status
	complStruct.Remarks = remarks
	complStruct.ActionTaken = action
	complaintAsBytes, err1 := json.Marshal(complStruct)
	if err1 != nil {
		return shim.Error(err1.Error())
	}
	err1 = stub.PutState(compKey, complaintAsBytes)
	if err1 != nil {
		return shim.Error("Error while Updating Complaint Data.")
	}
	return shim.Success([]byte("Complaint Status Updated Successfully!!"))
}

//Retrieving all the complaints given OAP from Blockchain
/*
	fcnName: queryComplaintByOAP
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [Oap]
	to know the total number of complaints given oap in blockchain, using range query(selector query)
	return pb.Response= Payload of all the complaints given oap
*/
func (c *Telco) queryComplaintByOAP(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1(OAP)")
	}
	oap := args[0]
	queryString := fmt.Sprintf("{\"selector\":{\"oap\":\"%s\"}}", oap)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//Retrieving all the complaints given TAP from Blockchain
/*
	fcnName: queryComplaintByTAP
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [Tap]
	to know the total number of complaints given tap in blockchain, using range query(selector query)
	return pb.Response= Payload of all the complaints given tap
*/
func (c *Telco) queryComplaintByTAP(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1(TAPCode)")
	}
	tap := args[0]
	queryString := fmt.Sprintf("{\"selector\":{\"tapCode\":\"%s\"}}", tap)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//Retrieving all the complaints given Status from Blockchain
/*
	fcnName: queryComplaintByStatus
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [status]
	to know the total number of complaints given status in blockchain, using range query(selector query)
	return pb.Response= Payload of all the complaints given status
*/
func (c *Telco) queryComplaintByStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1(Status)")
	}
	status := args[0]
	queryString := fmt.Sprintf("{\"selector\":{\"status\":\"%s\"}}", status)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//Retrieving all the complaints given StartDate and EndDate from Blockchain
/*
	fcnName: queryComplaintByDate
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [StartDate, EndDate]
	to know the total number of complaints given oap in blockchain, using range query(selector query)
	return pb.Response= Payload of all the complaints.
*/
func (c *Telco) queryComplaintByDate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2(StartDate, EndDate)")
	}
	sDate := args[0]
	eDate := args[1]
	docType := "Complaints"
	queryString := fmt.Sprintf("{\"selector\":{\"complaintsDocType\":\"%s\",\"complaintDate\":\"%s\"}}", docType, sDate)
	queryString1 := fmt.Sprintf("{\"selector\":{\"complaintsDocType\":\"%s\",\"complaintDate\":\"%s\"}}", docType, eDate)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	queryResults1, err := getQueryResultForQueryString(stub, queryString1)
	if err != nil {
		return shim.Error(err.Error())
	}
	var buffer bytes.Buffer
	buffer.WriteString("[")
	buffer.WriteString(string(queryResults))
	buffer.WriteString(",")
	buffer.WriteString(string(queryResults1))
	buffer.WriteString("]")
	return shim.Success(buffer.Bytes())
}

//Function used for Complex & Rich Queries
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {
	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")
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
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

// Addition of Metadata
// Used for building metadata for the pagination query
func addPaginationMetadataToQueryResults(buffer *bytes.Buffer, responseMetadata *pb.QueryResponseMetadata) *bytes.Buffer {

	buffer.WriteString("[{\"ResponseMetadata\":{\"RecordsCount\":")
	buffer.WriteString("\"")
	buffer.WriteString(fmt.Sprintf("%v", responseMetadata.FetchedRecordsCount))
	buffer.WriteString("\"")
	buffer.WriteString(", \"Bookmark\":")
	buffer.WriteString("\"")
	buffer.WriteString(responseMetadata.Bookmark)
	buffer.WriteString("\"}}]")

	return buffer
}

// used for building json response result iterator for pagination query
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

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
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return &buffer, nil
}

// getQueryResultForQueryStringWithPagination executes the passed in query string with
// pagination info. Result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryStringWithPagination(stub shim.ChaincodeStubInterface, queryString string, pageSize int32, bookmark string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)
	resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(queryString, pageSize, bookmark)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}
	bufferWithPaginationInfo := addPaginationMetadataToQueryResults(buffer, responseMetadata)
	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", bufferWithPaginationInfo.String())
	return buffer.Bytes(), nil
}

//Main function for Telco Chaincode
func main() {
	err := shim.Start(new(Telco))
	if err != nil {
		fmt.Printf("Error Starting Telco Chaincode : %s", err)
	}
}
