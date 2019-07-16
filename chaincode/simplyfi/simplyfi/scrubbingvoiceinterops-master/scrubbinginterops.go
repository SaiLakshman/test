/*
Copyright Tech Mahindra. 2019 All Rights Reserved.

*/

package main

import (
	"bytes"
	"encoding/json"                                        //reading and writing JSON
	id "github.com/hyperledger/fabric/core/chaincode/lib/cid" // import for Client Identity
	pb "github.com/hyperledger/fabric/protos/peer"         // import for peer response
	"github.com/hyperledger/fabric/core/chaincode/shim"    // import for Chaincode Interface
	"strconv"                                              //import for msisdn validation
	"fmt"
)

//Logger for Logging
var logger = shim.NewLogger("SCRUBBING_VOICE")

//Event Names
const EVTADDSCRUBVOICE = "ADD-SCRUBVOICE"
const EVTUPDATESCRUBVOICE = "UPDATE-SCRUBVOICE"


//Output Structure for the output response
type Output struct {
	Data         string `json:"data"`
	ErrorDetails string `json:"error"`
}

//Event Payload Structure
type Event struct {
	Data string `json:"data"`
	Txid string `json:"txid"`
}

//Smart Contract structure
type ScrubbingVoice struct {
}

//=========================================================================================================
// scrubbing data structure.  Structure tags are used by encoding/json library
//=========================================================================================================
type ScrubVoice struct {
	ObjType           string `json:"obj"`
	ScrubToken        string `json:"stok"`
	PEID  			  string `json:"peid"`
	TMID     		  string `json:"tmid"`
	CLI				  string `json:"cli"`
	CNAME             string `json:"cname"`
	TemplateID 		  string `json:"tid"`
	category          string `json:"ctgr"`
	communicationMode string `json:"cmode"`
	dayTimeBand       string `json:"time"`
	communicationType string `json:"ctyp"`
	Creator           string `json:"crtr"`
	CreateTs          string `json:"cts"`
	ConsumedBy        string `json:"csby"`
	Status            string `json:"sts"`
	SourceFileName    string `json:"ifile"`
	SourceFileHash    string `json:"iHash"`
	ScrubbedFileName  string `json:"ofile"`
	ScrubbedFileHash  string `json:"ohash"`
	UpdatedBy         string `json:"uby"`
	UpdateTs          string `json:"uts"`
}

//=========================================================================================================
// Init Chaincode
// The Init method is called when the Smart Contract "scrubbinginterops" is instantiated by the blockchain network
//=========================================================================================================

func (sv *ScrubbingVoice) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("###### Scrubbing Voice-Chaincode is Initialized #######")
	return shim.Success(nil)
}

// ========================================
// Invoke - Entry point for Invocations
// ========================================
func (sv *ScrubbingVoice) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	logger.Infof("ScrubbingVoice ChainCode Invoked, Function Name: " + string(function))
	switch function {
	case "rs": // register scrubbing details
		return sv.registerScrub(stub, args)
	case "uss": //update scrubbing voice details
		return sv.updateScrubStatus(stub, args)
	case "qs": //query for scrubbing details
		return sv.queryScrub(stub, args)
	case "qsp": //query for scrubbing details
		return sv.queryScrubbingWithPagination(stub, args)	
	default:
		logger.Errorf("Unknown Function Invoked, Available Function argument shall be any one of : rs,rbs,uss,qs,qsp")
		return shim.Error("Available Functions: rs,rbs,uss,qs,qsp")
	}
}


// ========================================
// Scrubbing Registration method
// ========================================
func (sv *ScrubbingVoice) registerScrub(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	logger.Info("---- Inside registerScrub function ----")

	jsonResp := ""
	status := "Scrubbing registered successfully"
	jsonBlob := args[0]
	var scrubObj ScrubVoice

	err := json.Unmarshal([]byte(jsonBlob), &scrubObj)
	if err != nil {
		logger.Errorf("registerScrub : unable to Unmarshal data : " + string(err.Error()))
		return shim.Error(err.Error())
	}

	scrubObj.ObjType = "ScrubVoice"

	// === Get the creator Id ===

	_, Crtr := sv.getInvokerIdentity(stub)
	scrubObj.Creator = Crtr

	scrubAsBytes, err := stub.GetState(scrubObj.ScrubToken)

	if err != nil {
		logger.Errorf("registerScrub : unable to getstate  : " + string(err.Error()))
		jsonResp = "Failed to get scrub details: " + err.Error()
		return shim.Error(jsonResp)
	} else if scrubAsBytes != nil {
		logger.Errorf(" == This scrub token already exists ===" + scrubObj.ScrubToken)
		jsonResp = "This scrub token already exists: " + scrubObj.ScrubToken
		return shim.Error(jsonResp)
	}

	// === Convert the structure into byte array ===
	scrubJSONasBytes, err := json.Marshal(scrubObj)
	if err != nil {
		logger.Errorf("registerScrub : unable to marshal data : " + string(err.Error()))
		jsonResp = err.Error()
		return shim.Error(jsonResp)
	}
	err = stub.PutState(scrubObj.ScrubToken, scrubJSONasBytes)
	if err != nil {
		logger.Errorf("registerScrub : unable to putstate data : " + string(err.Error()))
		jsonResp = err.Error()
		return shim.Error(jsonResp)
	}
	
	logger.Infof("registerScrub : PutState Success : " + string(scrubJSONasBytes))
	eventbytes := Event{Data: string(scrubJSONasBytes), Txid: stub.GetTxID()}
	payload, err := json.Marshal(eventbytes)
	if err != nil {
		logger.Errorf("registerScrub : Event Payload marshaling Error : " + string(err.Error()))
		return shim.Error("registerScrub : Event Payload marshaling Error : " + string(err.Error()))
	}
	err2 := stub.SetEvent(EVTADDSCRUBVOICE, []byte(payload))
	if err2 != nil {
		logger.Errorf("registerScrub : Event Creation Error for EventID : " + string(EVTADDSCRUBVOICE))
		return shim.Error("registerScrub : Event Creation Error for EventID : " + string(EVTADDSCRUBVOICE))
	}
	logger.Infof("Event published data: " + string(payload))
	jsonResp = "{\"result\":\"success\",\"stok\":\"" + scrubObj.ScrubToken + "\",\"status\":\"" + status + "\"}"

	logger.Infof("Scrubbing registration status " + status)

	return shim.Success([]byte(jsonResp))
}


//=====================================================
//Update the registered scrubbing details 
//=====================================================

func (sv *ScrubbingVoice) updateScrubStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	logger.Info("---- Inside updateScrubStatus ----")
	jsonResp := ""
	status := "Scrubbing details updated successfully"
	jsonBlob := args[0]
	var scrubObj ScrubVoice
	err := json.Unmarshal([]byte(jsonBlob), &scrubObj)
	if err != nil {
		logger.Errorf("updateScrubStatus : unable to unmarshal data : " + string(err.Error()))
		return shim.Error(err.Error())
	}

	// === Get the updater Id ===

	_, Uby := sv.getInvokerIdentity(stub)
	scrubObj.UpdatedBy = Uby

	// === Get the DCMP ===

	scrubToken := scrubObj.ScrubToken
	scrubAsBytes, err := stub.GetState(scrubToken)
	if err != nil {
		logger.Errorf("updateScrubStatus : unable to getstate data : " + string(err.Error()))
		jsonResp = "Failed to get scrub details: " + err.Error()
		return shim.Error(jsonResp)
	} else if scrubAsBytes == nil {
		jsonResp = "scrub token does not exist"
		return shim.Error(jsonResp)
	}

	var scrubBCObj ScrubVoice
	err = json.Unmarshal(scrubAsBytes, &scrubBCObj)
	if err != nil {
		logger.Errorf("updateScrubStatus : unable to unmarshal data : " + string(err.Error()))
		return shim.Error(err.Error())
	}

	scrubBCObj.Status = scrubObj.Status
	scrubBCObj.ConsumedBy = scrubObj.ConsumedBy
	scrubBCObj.UpdateTs = scrubObj.UpdateTs
	
	scrubJSONasBytes, err := json.Marshal(scrubBCObj)

	if err != nil {
		jsonResp = err.Error()
		return shim.Error(jsonResp)
	}
	// === Save scrub to state ===
	err = stub.PutState(scrubToken, scrubJSONasBytes)
	if err != nil {
		logger.Errorf("updateScrubStatus : " + string(err.Error()))
		jsonResp = err.Error()
		return shim.Error(jsonResp)
	}
	logger.Infof("updateScrubStatus : PutState Success : " + string(scrubJSONasBytes))
	eventbytes := Event{Data: string(scrubJSONasBytes), Txid: stub.GetTxID()}
	payload, err := json.Marshal(eventbytes)
	if err != nil {
		logger.Errorf("updateScrubStatus : Event Payload marshaling Error : " + string(err.Error()))
		return shim.Error("updateScrubStatus : Event Payload marshaling Error : " + string(err.Error()))
	}
	err2 := stub.SetEvent(EVTUPDATESCRUBVOICE, []byte(payload))
	if err2 != nil {
		logger.Errorf("updateScrubStatus : Event Creation Error for EventID : " + string(EVTUPDATESCRUBVOICE))
		return shim.Error("updateScrubStatus : Event Creation Error for EventID : " + string(EVTUPDATESCRUBVOICE))
	}
	logger.Infof("Event published data: " + string(payload))	

	jsonResp = "{\"result\":\"success\",\"stok\":\"" + scrubToken + "\",\"status\":\"" + status + "\"}"

	logger.Infof("Scrubbing details modification status " + status)

	return shim.Success([]byte(jsonResp))

}


//======================================================================================
//queryScrub RichQuery for Obtaining scrubbing voice data
//======================================================================================

func (sv *ScrubbingVoice) queryScrub(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("queryScrub : Incorrect number of arguments, Expected 1 [Query String]")
	}
	queryString := args[0]
	logger.Info(args[0])
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		logger.Errorf("queryScrub : getQueryResultForQueryString Failed Error : " + string(err.Error()))
		return shim.Error("queryScrub : getQueryResultForQueryString Failed Error : " + string(err.Error()))
	}
	return shim.Success(queryResults)
}

func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}


//========================================================================================================
//queryScrubbingWithPagination RichQuery for Obtaining scrub voice data with pagination for more records
//========================================================================================================
func (sv *ScrubbingVoice) queryScrubbingWithPagination(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 3 {
		return shim.Error("queryScrubbingWithPagination : Incorrect number of arguments, Expected 3 arg")
	}

	queryString := args[0]
	pageSize, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		logger.Errorf("queryScrubbingWithPagination : Parsing error : " + string(err.Error()))
		return shim.Error("queryScrubbingWithPagination : Parsing error : " + string(err.Error()))
	}
	bookmark := args[2]

	queryResults, err := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookmark)
	if err != nil {
		logger.Errorf("queryScrubbingWithPagination : pagination process : " + string(err.Error()))
		return shim.Error("queryScrubbingWithPagination : pagination process : " + string(err.Error()))
	}
	return shim.Success(queryResults)
}


//====================================================================================================================
//getQueryResultForQueryStringWithPagination RichQuery for Obtaining scrub voice data with pagination for more records
//====================================================================================================================
func getQueryResultForQueryStringWithPagination(stub shim.ChaincodeStubInterface, queryString string, pageSize int32, bookmark string) ([]byte, error) {

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
	
	logger.Infof("-- getQueryResultForQueryString queryResult -- : " + bufferWithPaginationInfo.String())

	return buffer.Bytes(), nil
}

// ===========================================================================================
// addPaginationMetadataToQueryResults constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================
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


// ===========================================================================================
// constructQueryResponseFromIterator constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================
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

//to get invoker identity informations
func (em *ScrubbingVoice) getInvokerIdentity(stub shim.ChaincodeStubInterface) (bool, string) {

	//Following id comes in the format X509::<Subject>::<Issuer>>

	enCert, err := id.GetX509Certificate(stub)
	if err != nil {
		logger.Info("getInvokerIdentity : Unable to get certificate details --  ")
		return false, "Unknown."
	}

	issuersOrgs := enCert.Issuer.Organization
	if len(issuersOrgs) == 0 {
		return false, "Unknown.."
	}
	return true, fmt.Sprintf("%s", issuersOrgs[0])

}


// ===================================================================================
//main function for the scrubbing voice ChainCode
// ===================================================================================
func main() {
	err := shim.Start(new(ScrubbingVoice))
	logger.SetLevel(shim.LogDebug)
	if err != nil {
		logger.Error("Error Starting ScrubbingVoice Chaincode is " + string(err.Error()))
	} else {
		logger.Info("Starting ScrubbingVoice Chaincode")
	}
}

