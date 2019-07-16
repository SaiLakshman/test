/*
Copyright Blockcube | Pennybase Technologies Sol. Pvt. Ltd. 2019 All Rights Reserved.
This Chaincode is written for Storing, Retrieving, querying the Header that are stored in DLT.
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ================== START

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["rh","{\"hid\":\"QT041111111111111101\",\"peid\":\"A11111111101\",\"htyp\":\"T\",\"cli\":\"BLOCKCUBE\",\"ctgr\":\"8\",\"cts\":\"456789\",\"uts\":\"456787678\"}"]}'

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["rbh","{\"hid\":\"QT041111111111111102\",\"peid\":\"A11111111101\",\"htyp\":\"T\",\"cli\":\"BLOCKCUBE2\",\"ctgr\":\"8\",\"cts\":\"456789\",\"uts\":\"456787678\"}","{\"hid\":\"QT041111111111111103\",\"peid\":\"A11111111102\",\"htyp\":\"T\",\"cli\":\"BLOCKCUBE3\",\"ctgr\":\"8\",\"cts\":\"456789\",\"uts\":\"456787678\"}"]}'

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["uhs","{\"cli\":\"BLOCKCUBE\",\"sts\":\"I\",\"uts\":\"2345678\"}"]}'

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["qh","BLOCKCUBE2","BLOCKCUBE3"]}'

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["hfh","{\"cli\":\"BLOCKCUBE\"}"]}'

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["qhbp","{\"typ\":\"cli\",\"cli\":\"BLOCKCUBE\"}"]}'

// peer chaincode invoke -o <ORDERER_ENDPOINT> -n header -C chheader  -c '{"args":["qhbp","{\"typ\":\"peid\",\"peid\":\"A11111111102\"}"]}'

// ====CHAINCODE EXECUTION SAMPLES (CLI) ================== END

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("HEADERSMS-CHAINCODE-INITIALIZED")

const EVTRegisterHeader = "EVT_RegisterHeader"
const EVTUpdateHeaderStatus = "EVT_UpdateHeaderStatus"

type HeaderChainCode struct {
}

var jsonResp string

//Header structure defines the ledger record for any Header
type Header struct {
	ObjType           string `json:"obj"`  // ObjType : is used to distinguish the various types of objects in state database
	Header_ID         string `json:"hid"`  // hid     : Unique id Autogenerated in Backend
	PrincipleEntityId string `json:"peid"` // peid    : Unique id and A group of headers belongs to this particular entity
	Header_Type       string `json:"htyp"` // htyp    : Either T/S/P : Transactional / Service / Promotional
	Header_Name       string `json:"cli"`  // cli     : Unique name to be registeres in DLT
	Status            string `json:"sts"`  // sts     : Either A/B/I/D : Active / Block / Inactive / Delete
	Category          string `json:"ctgr"` // ctgr    : In betwen 1-8
	CreatedTs         string `json:"cts"`  // cts     : Header creation time : Autogenerated in Backend
	UpdatedTs         string `json:"uts"`  // uts     : When the header last updated : Autogenerate in Backend
	Creator           string `json:"crtr"` // crtr    : DLT  Creator node's name
	UpdatedBy         string `json:"uby"`  // uby     : DLT Node's name
	TMID              string `json:"tmid"`
}

var validHeaderType = map[string]bool{
	"SE": true,
	"SI": true,
	"T":  true,
	"P":  true,
}

var validCategory = map[string]bool{
	"1": true,
	"2": true,
	"3": true,
	"4": true,
	"5": true,
	"6": true,
	"7": true,
	"8": true,
}

func validHeaderEntry(input string, enumMap map[string]bool) bool {
	if _, isEntryExists := enumMap[input]; !isEntryExists {
		return false
	}
	return true
}

func isValidHeader(header Header) (bool, string) {

	if len(header.Header_Name) > 11 || len(header.Header_Name) < 6 {
		return false, "Invalid length for cli, shoud lie in between 6 to 11 digits"
	}

	if len(header.PrincipleEntityId) == 0 {
		return false, "EntityID is mandatory"
	}

	if len(header.Header_Name) == 0 {
		return false, "Header_Name is mandatory"
	}

	if len(header.CreatedTs) == 0 {
		return false, "Created Timestamp is mandatory"
	}

	if len(header.UpdatedTs) == 0 {
		return false, "Updated Timestamp is mandatory"
	}

	if !validHeaderEntry(header.Header_Type, validHeaderType) {
		return false, "Invalid Header Type"
	}

	if !validHeaderEntry(header.Category, validCategory) {
		return false, "Invalid Category Provided"
	}

	return true, ""
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(HeaderChainCode))
	logger.SetLevel(shim.LogDebug)
	if err != nil {
		logger.Error("Error starting Header Chaincode: %s: ", err)
	} else {
		logger.Info("##### || STARTING HEADER CHAINCODE || #####")
	}
}

// ===================================================================================
// Init initializes chaincode
// ===================================================================================
func (t *HeaderChainCode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	logger.Info("##### || HEADER CHAINCODE IS INITIALIZED || #####")
	return shim.Success(nil)
}

// ===================================================================================
// Invoke - Our entry point for Invocations
// ===================================================================================
func (t *HeaderChainCode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	function, args := stub.GetFunctionAndParameters()
	logger.Infof("Header Chaincode Invoked, Function name : " + string(function))

	// Handle different functions
	switch function {
	case "rh":
		return t.registerHeader(stub, args) // Register a new header
	case "rbh":
		return t.registerBulkHeader(stub, args) // Register headers in Bulk
	case "uhs":
		return t.updateHeaderStatus(stub, args) // Change status of a header
	case "qh":
		return t.queryHeader(stub, args) // Query by Array of CLI : All Matching headers will be returned
	case "qhbp":
		return t.queryHeaderByParam(stub, args) // Query header depending upon different parameters : CLI / PEID / HID
	case "hfh":
		return t.getHistoryForHeader(stub, args) // Get history against header when Header name is passed
	case "qhwp":
		return t.queryHeaderWithPagination(stub, args) // uses a query string, page size and a bookmark to perform a query
	default:
		logger.Errorf("##### || Received Unknown Function invocation || ##### : Available Function : rh , rbh , uhs, qh, hfh, qhwp")
		return shim.Error("##### || Received Unknown Function invocation || ##### Available function : rh , rbh , uhs, qh, hfh, qhwp")
	}
}

// ========================================================================================
// registerHeader - register a header in chaincode state
// ========================================================================================
func (t *HeaderChainCode) registerHeader(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	var data Header
	err := json.Unmarshal([]byte(args[0]), &data)
	if err != nil {
		logger.Errorf("setHeader : Input arguments unmarhsaling Error : " + string(err.Error()))
		return shim.Error("setHeader : Input arguments unmarhsaling Error : " + string(err.Error()))
	}

	if isValid, errMsg := isValidHeader(data); !isValid {
		logger.Errorf("setHeader:" + string(errMsg))
		return shim.Error(errMsg)
	}

	certData, err := cid.GetX509Certificate(stub)
	if err != nil {
		logger.Errorf("setHeader : Getting certificate Details Error : " + string(err.Error()))
		return shim.Error("setHeader : Getting certificate Details Error : " + string(err.Error()))
	}

	Organizations := certData.Issuer.Organization
	headerSearch := `{
		"obj":"HeaderSMS",
		"hid":"%s"
	}`
	hID := data.Header_ID
	headerData := t.retriveHeaderRecords(stub, fmt.Sprintf(headerSearch, hID), "headerSearchByID")
	if len(headerData) > 0 {
		logger.Infof("### Header_ID already exist for : " + data.Header_Name + ", Please provide unique hid ####")
		return shim.Error("### Header_ID already exist for : " + data.Header_Name + ", Please provide unique hid ####")
	}

	value, err := stub.GetState(data.Header_Name)
	if err != nil {
		logger.Infof("#### Failed to get Header Record ####: " + data.Header_Name + " Error : " + string(err.Error()))
		return shim.Error("#### Failed to get Header Record ####" + data.Header_Name + " Error : " + string(err.Error()))
	}

	if value == nil {
		data.ObjType = "HeaderSMS"
		data.Status = "A"
		data.Creator = Organizations[0]
		data.UpdatedBy = Organizations[0]
		logger.Infof("Header_ID is " + data.Header_ID)
		headerAsBytes, err := json.Marshal(data)
		if err != nil {
			logger.Errorf("setHeader : Marshalling Error : " + string(err.Error()))
			return shim.Error("setHeader : Marshalling Error : " + string(err.Error()))
		}
		//Inserting DataBlock to BlockChain
		err = stub.PutState(data.Header_Name, headerAsBytes)
		if err != nil {
			logger.Errorf("setHeader : PutState Failed Error : " + string(err.Error()))
			return shim.Error("setHeader : PutState Failed Error : " + string(err.Error()))
		}
		logger.Infof("setHeader : PutState Success : " + string(headerAsBytes))

		err2 := stub.SetEvent(EVTRegisterHeader, headerAsBytes)
		if err2 != nil {
			logger.Errorf("Event not generated for event : EVTRegisterHeader")
			return shim.Error("{\"error\":\"Event not generated for event - EVTRegisterHeader.\"}")
		}

		resultData := map[string]interface{}{
			"trxnID":           stub.GetTxID(),
			"headerRegistered": data.Header_Name,
			"message":          "success",
			"Header":           data,
		}
		respJSON, _ := json.Marshal(resultData)
		return shim.Success(respJSON)
	} else {
		logger.Errorf("setHeader : Header is Already Registered")
		return shim.Error("setHeader : Header is Already Registered")
	}

}

// ========================================================================================
// registerBulkHeader - Register Bulk header in chaincode state
// ========================================================================================
func (t *HeaderChainCode) registerBulkHeader(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	var recordcount int
	headerRejected := make([]map[string]interface{}, 0)
	headerRegistered := make([]string, 0)

	recordcount = 0
	for i := 0; i < len(args); i++ {
		var data Header
		logger.Infof(args[i])
		err := json.Unmarshal([]byte(args[i]), &data)
		if err != nil {
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name: ": data.Header_Name, "Value": "Input arguments unmarhsaling Error"})
			continue
		}

		certData, err := cid.GetX509Certificate(stub)
		if err != nil {
			logger.Errorf("registerBulkHeader : Getting certificate Details Error : " + string(err.Error()))
			return shim.Error("registerBulkHeader : Getting certificate Details Error : " + string(err.Error()))
		}

		Organizations := certData.Issuer.Organization
		if isValid, errMsg := isValidHeader(data); !isValid {
			logger.Errorf("batchPreferences:" + string(errMsg))
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name: ": data.Header_Name, "Value": string(errMsg)})
			continue
		}

		headerSearch := `{
			"obj":"HeaderSMS",
			"hid":"%s"
		}`
		hID := data.Header_ID
		headerData := t.retriveHeaderRecords(stub, fmt.Sprintf(headerSearch, hID), "headerSearchByID")
		if len(headerData) > 0 {
			logger.Errorf("### Header_ID already exist for : " + data.Header_Name + ", Please provide unique hid ####")
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name: ": data.Header_Name, "Value": "Header ID already exists"})
			continue
		}

		value, err := stub.GetState(data.Header_Name)
		if err != nil {
			logger.Errorf("registerBulkHeader : GetState Failed for Header_Name : " + data.Header_Name + " , Error : " + string(err.Error()))
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name: ": data.Header_Name, "Value": "Get State failed "})
			continue
		}

		if value == nil {
			recordcount = recordcount + 1
			data.ObjType = "HeaderSMS"
			data.Status = "A"
			data.Creator = Organizations[0]
			data.UpdatedBy = Organizations[0]
			logger.Infof("Header_ID is " + data.Header_ID)
			headerAsBytes, err := json.Marshal(data)
			if err != nil {
				logger.Errorf("registerBulkHeader : Marshalling Error : " + string(err.Error()))
				return shim.Error("registerBulkHeader : Marshalling Error : " + string(err.Error()))
			}
			//Inserting DataBlock to BlockChain
			err = stub.PutState(data.Header_Name, headerAsBytes)
			if err != nil {
				logger.Errorf("registerBulkHeader : PutState Failed Error : " + string(err.Error()))
				return shim.Error("registerBulkHeader : PutState Failed Error : " + string(err.Error()))
			}
			logger.Infof("registerBulkHeader : PutState Success : " + string(headerAsBytes))
			headerRegistered = append(headerRegistered, data.Header_Name)
		} else {
			headerRejected = append(headerRejected, map[string]interface{}{"Header_Name: ": data.Header_Name, "Value": "Header already registered"})
		}
	}

	resultData := map[string]interface{}{
		"headerRegistered": headerRegistered,
		"headerRejected":   headerRejected,
		"countSuccess":     strconv.Itoa(recordcount),
		"txid":             stub.GetTxID(),
	}

	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

// ========================================================================================
// updateHeaderStatus - Update header status to A / B / I / D
// ========================================================================================
func (t *HeaderChainCode) updateHeaderStatus(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	var count bool
	count = false
	var data map[string]interface{}
	err := json.Unmarshal([]byte(args[0]), &data)
	if err != nil {
		logger.Errorf("updateHeaderStatus : Input arguments unmarhsaling Error : " + string(err.Error()))
		return shim.Error("updateHeaderStatus : Input arguments unmarhsaling Error : " + string(err.Error()))
	}

	certData, err := cid.GetX509Certificate(stub)
	if err != nil {
		logger.Errorf("setHeader : Getting certificate Details Error : " + string(err.Error()))
		return shim.Error("setHeader : Getting certificate Details Error : " + string(err.Error()))
	}
	Organizations := certData.Issuer.Organization
	value, err := stub.GetState(data["cli"].(string))
	if err != nil {
		logger.Errorf("registerBulkHeader : GetState Failed for Header_Name : " + data["cli"].(string) + " , Error : " + string(err.Error()))
		return shim.Error("registerBulkHeader : GetState Failed for Header_Name : " + data["cli"].(string) + " , Error : " + string(err.Error()))
	}

	if value != nil {
		if len(data) == 3 {
			var organizationName string
			var orgName string
			header := Header{}
			err := json.Unmarshal(value, &header)
			if err != nil {
				logger.Errorf("registerBulkHeader : Existing header data Unmarhsaling Error : " + string(err.Error()))
				return shim.Error("registerBulkHeader : Existing header data Unmarhsaling Error : " + string(err.Error()))
			}

			orgName = header.Creator
			organizationName = Organizations[0]
			if strings.Compare(orgName, organizationName) == 0 {

				switch data["sts"].(string) {
				case "A":
					if header.Status == "I" || header.Status == "B" {
						header.Status = "A"
					} else {
						logger.Errorf("Header current status is not I or B")
						return shim.Error("Header current status is not I or B")
					}
				case "B":
					if header.Status == "I" || header.Status == "A" {
						header.Status = "B"
					} else {
						logger.Errorf("Header current status is not I or A")
						return shim.Error("Header current status is not I or A")
					}
				case "I":
					if header.Status == "A" {
						header.Status = "I"
					} else {
						logger.Errorf("Header current status is not A")
						return shim.Error("Header current status is not A")
					}
				case "D":
					if header.Status == "B" {
						header.Status = "D"
					} else {
						logger.Errorf("Header current status is not B")
						return shim.Error("Header current status is not B")
					}
				default:
					logger.Errorf("##### || Received Unknown Status type || Must provide either A or B or I or D ")
					return shim.Error("##### || Received Unknown Status type || Must provide either A or B or I or D ")
				}

				header.UpdatedTs = data["uts"].(string)
				logger.Infof("Header_Name is " + header.Header_Name)
				headerAsBytes, err := json.Marshal(header)
				if err != nil {
					logger.Errorf("updateHeaderStatus : Marshalling Error : " + string(err.Error()))
					return shim.Error("updateHeaderStatus : Marshalling Error : " + string(err.Error()))
				}
				//Inserting DataBlock to BlockChain
				err = stub.PutState(header.Header_Name, headerAsBytes)
				if err != nil {
					logger.Errorf("updateHeaderStatus : PutState Failed Error : " + string(err.Error()))
					return shim.Error("updateHeaderStatus : PutState Failed Error : " + string(err.Error()))
				}
				logger.Infof("updateHeaderStatus : PutState Success : " + string(headerAsBytes))

				err2 := stub.SetEvent(EVTUpdateHeaderStatus, headerAsBytes)

				if err2 != nil {
					logger.Errorf("Event not generated for event : EVTUpdateHeaderStatus")
					return shim.Error("{\"error\":\"Unable register the header.\"}")
				}

				count = true
			} else {
				logger.Errorf("Unauthorized Access")
				return shim.Error("Unauthorized Access")
			}

		} else {
			logger.Errorf("updateHeaderStatus : Incorrect Number Of Arguments, i.e. CLI, Status, UpdatedTs expected")
			return shim.Error("updateHeaderStatus : Incorrect Number Of Arguments i.e. CLI, Status, UpdatedTs expected")
		}
	} else {
		logger.Errorf("Header is not registered")
		return shim.Error("Header is not registered")
	}
	if count == true {
		logger.Infof("Header status is updated Successfully for Header_Name : " + data["cli"].(string))
		return shim.Success([]byte("Header status is upadted Successfully Header_Name : " + data["cli"].(string)))
	} else {
		return shim.Error("Error in header status update")
	}
}

// ========================================================================================
// queryHeader - Query by Array of CLI : All Matching headers will be returned
// ========================================================================================
func (t *HeaderChainCode) queryHeader(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	var jsonResp string
	resp := make([]map[string]interface{}, 0)
	notExist := make([]string, 0)
	exist := make([]string, 0)

	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}

	for i := 0; i < len(args); i++ {
		valAsBytes, err := stub.GetState(args[i]) //get the record from chaincode state

		if err != nil {
			jsonResp = "{\"Error\":\"Failed to get state for Header_Name " + args[i] + "\"}"
			fmt.Println(jsonResp)
			return shim.Error(jsonResp)
		} else if valAsBytes == nil {
			jsonResp = "{\"Error\":\"Record does not exist: " + args[i] + "\"}"
			logger.Infof("Record does not exist for header name : " + args[i])
			fmt.Println(jsonResp)
			notExist = append(notExist, args[i])
			// return shim.Error(jsonResp)
		}

		exist = append(exist, args[i])
		value := make(map[string]interface{})
		json.Unmarshal(valAsBytes, &value)
		// recordsJSON, _ := json.Marshal(valAsBytes)
		resp = append(resp, map[string]interface{}{"Header_Name: ": args[i], "Value": value})
		logger.Info("Successfully submitted the result for " + args[i])
	}

	logger.Info("")
	resultData := map[string]interface{}{
		"cliNotFound": notExist,
		"cliFound":    exist,
		"dataOfCli":   resp,
	}

	respJSON, _ := json.Marshal(resultData)
	return shim.Success(respJSON)
}

// ========================================================================================
// queryHeaderByParam - Query header depending upon different parameters : CLI / PEID / HID
// ========================================================================================
func (t *HeaderChainCode) queryHeaderByParam(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var response sc.Response
	searchCriteria := make(map[string]string)

	if len(args) < 1 {
		return shim.Error("Invalid number of arguments provided for transaction")
	}

	err := json.Unmarshal([]byte(args[0]), &searchCriteria)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}
	searchType, isOk := searchCriteria["typ"]
	if !isOk {
		return shim.Error("Search type not provided")
	}

	logger.Infof(args[0])
	logger.Infof("length is : ", len(searchCriteria[searchType]))

	switch searchType {
	case "cli":
		headerSearchCriteria := `{
			"obj":"HeaderSMS"	,
			"cli":"%s"
		}`
		headerName := searchCriteria[searchType]
		header := t.retriveHeaderRecords(stub, fmt.Sprintf(headerSearchCriteria, headerName), "headerSearchByName")
		recordsJSON, _ := json.Marshal(header)
		response = shim.Success(recordsJSON)
	case "hid":
		headerSearchCriteria := `{
			"obj":"HeaderSMS"	,
			"hid":"%s"
		}`
		headerID := searchCriteria[searchType]
		header := t.retriveHeaderRecords(stub, fmt.Sprintf(headerSearchCriteria, headerID), "headerSearchByID")
		recordsJSON, _ := json.Marshal(header)
		response = shim.Success(recordsJSON)
	case "peid":
		headerSearchCriteria := `{
			"obj":"HeaderSMS"	,
			"peid":"%s"
		}`
		headerID := searchCriteria[searchType]
		header := t.retriveHeaderRecords(stub, fmt.Sprintf(headerSearchCriteria, headerID), "headerSearchByPeid")
		recordsJSON, _ := json.Marshal(header)
		response = shim.Success(recordsJSON)
	default:
		response = shim.Error("Unsupported search type provided " + searchType)
	}
	return response
}

// ========================================================================================
// getHistoryForHeader - Get history of the key when header name is passed
// ========================================================================================
func (t *HeaderChainCode) getHistoryForHeader(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	var data map[string]interface{}
	err := json.Unmarshal([]byte(args[0]), &data)
	if err != nil {
		logger.Errorf("getHistoryForHeader : Input arguments unmarhsaling Error : " + string(err.Error()))
		return shim.Error("getHistoryForHeader : Input arguments unmarhsaling Error : " + string(err.Error()))
	}

	RecordAsBytes, err := stub.GetState(data["cli"].(string))
	if err != nil {
		logger.Infof("#### Failed to get Header Record ####: " + data["cli"].(string) + " Error : " + string(err.Error()))
		return shim.Error("#### Failed to get Header Record ####" + data["cli"].(string) + " Error : " + string(err.Error()))
	} else if RecordAsBytes == nil {
		fmt.Println("#### This record does not exists #### " + data["cli"].(string))
		return shim.Error("#### This record does not exists ####" + data["cli"].(string))
	}

	historyIer, err := stub.GetHistoryForKey(data["cli"].(string))

	if err != nil {
		fmt.Println(err.Error())
		return shim.Error(err.Error())
	}

	historicResponse := make([]map[string]interface{}, 0)
	for historyIer.HasNext() {
		modification, err := historyIer.Next()
		if err != nil {
			fmt.Println(err.Error())
			return shim.Error(err.Error())
		}

		value := make(map[string]interface{})
		json.Unmarshal(modification.Value, &value)
		historicResponse = append(historicResponse, map[string]interface{}{"txId": modification.TxId, "value": value})
	}

	respJSON, _ := json.Marshal(historicResponse)
	return shim.Success(respJSON)
}

// ===========================================================================================
// queryHeaderWithPagination - uses a query string, page size and a bookmark to perform a query
// for marbles. Query string matching state database syntax is passed in and executed as is.
// ===========================================================================================
func (t *HeaderChainCode) queryHeaderWithPagination(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 3 {
		return shim.Error("queryHeadersWithPagination : Incorrect number of arguments. Expecting 3 Args")
	}

	queryString := args[0]
	pageSize, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		logger.Errorf("queryHeadersWithPagination : Unable to parse the input data: " + string(err.Error()))
		return shim.Error("queryHeadersWithPagination : Unable to parse the input data: " + string(err.Error()))
	}
	bookmark := args[2]

	queryResults, err := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookmark)
	if err != nil {
		logger.Errorf("queryHeadersWithPagination : Error in pagination: " + string(err.Error()))
		return shim.Error("queryHeadersWithPagination : Error in pagination: " + string(err.Error()))
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryStringWithPagination executes the passed in query string with
// pagination info. Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryStringWithPagination(stub shim.ChaincodeStubInterface, queryString string, pageSize int32, bookmark string) ([]byte, error) {

	logger.Infof("- getQueryResultForQueryString queryString:\n%s\n", queryString)

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

	logger.Infof("### getQueryResultForQueryString queryResult ### \n%s\n", bufferWithPaginationInfo.String())

	return buffer.Bytes(), nil
}

// ===========================================================================================
// addPaginationMetadataToQueryResults adds QueryResponseMetadata, which contains pagination
// info, to the constructed query results
// ===========================================================================================
func addPaginationMetadataToQueryResults(buffer *bytes.Buffer, responseMetadata *sc.QueryResponseMetadata) *bytes.Buffer {

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

// ===========================================================================================
// constructQueryResponseFromIterator constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			logger.Errorf("constructQueryResponseFromIterator : unable to query : " + string(err.Error()))
			return nil, err
		}
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

func (t *HeaderChainCode) retriveHeaderRecords(stub shim.ChaincodeStubInterface, criteria string, indexs ...string) []Header {

	var finalSelector string
	records := make([]Header, 0)

	if len(indexs) == 0 {
		finalSelector = fmt.Sprintf("{\"selector\":%s }", criteria)

	} else {
		finalSelector = fmt.Sprintf("{\"selector\":%s , \"use_index\" :\"%s\" }", criteria, indexs[0])
	}

	logger.Infof("Query Selector : %s", finalSelector)
	resultsIterator, _ := stub.GetQueryResult(finalSelector)
	for resultsIterator.HasNext() {
		record := Header{}
		recordBytes, _ := resultsIterator.Next()
		err := json.Unmarshal(recordBytes.Value, &record)
		if err != nil {
			logger.Infof("Unable to unmarshal Header retrived:: %v", err)
		}
		records = append(records, record)
	}
	return records
}
