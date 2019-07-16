package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"bytes"
	"crypto/x509"
    "encoding/pem"
	"github.com/hyperledger/fabric/core/chaincode/shim" // import for Chaincode Interface
	pb "github.com/hyperledger/fabric/protos/peer"      // import for peer response
	"github.com/hyperledger/fabric/protos/msp"
	"github.com/golang/protobuf/proto"
)

type Test struct {
}

//Item Data
type Item struct {
	MobileId    string `json:"phNum"`
	Description string `json:"desc"`
	Model       string `json:"model"`
}
type inputItem struct {
	MobileId    string `json:"phNum"`
	Description string `json:"desc"`
	Model       string `json:"model"`
	Price       string `json:"price"`
}

//Item With Price Data
type ItemPrice struct {
	MobileId    string `json:"phNum"`
	Description string `json:"desc"`
	Model       string `json:"model"`
	Price       string `json:"price"`
}

//test data for checking the updation of the status by creator
type dataToSave struct {
	MobileId	string `json:"phNum"`
	Owner		string `json:"owner"`
	Status	string `json:"sts"`
}
//Init function of the chaincode
func (c *Test) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

//Invoke function of the chaincode
func (c *Test) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "create":
		return c.createItem(stub, args)
	case "retrieveWithPrice":
		return c.retrieveWithPrice(stub, args)
	case "retrieve":
		return c.retrieve(stub, args)
	case "getItem":
		 return c.retrieveRange(stub, args)
	case "set":
		return c.setData(stub, args)
	case "get":
		return c.getData(stub, args)
	case "update":
		return c.updateData(stub, args)		
	default:
		return shim.Error("Not a Valid Function.")
	}
}

func (c *Test) createItem(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect Number of Arguments. Expecting 1")
	}
	var itemToSave inputItem
	err := json.Unmarshal([]byte(args[0]), &itemToSave)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}
	// ==== Check if marble already exists ====
	itemAsBytes, err := stub.GetPrivateData("PrivateCollection", itemToSave.MobileId)
	if err != nil {
		return shim.Error("Failed to get Item: " + err.Error())
	} else if itemAsBytes != nil {
		fmt.Println("This Item already exists: " + itemToSave.MobileId)
		return shim.Error("This Item already exists: " + itemToSave.MobileId)
	}

	itemPD := &Item{}
	itemPD.MobileId = itemToSave.MobileId
	itemPD.Description = itemToSave.Description
	itemPD.Model = itemToSave.Model

	itemJSON, err := json.Marshal(itemPD)
	if err != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Item. \"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutPrivateData("PrivateCollection", itemToSave.MobileId, itemJSON)
	if err != nil {
		return shim.Error(err.Error())
	}

	itemPPD := &ItemPrice{}
	itemPPD.MobileId = itemToSave.MobileId
	itemPPD.Description = itemToSave.Description
	itemPPD.Model = itemToSave.Model
	itemPPD.Price = itemToSave.Price

	itemPriceJSON, err := json.Marshal(itemPPD)
	if err != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Item. \"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutPrivateData("PrivatePriceCollection", itemToSave.MobileId, itemPriceJSON)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("Item Created Successfully!!"))
}

//setting data for testing the invalid certifcates
func (c *Test) setData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect Number of Arguments. Expecting 1")
	}
	var ip dataToSave
	creator, err := stub.GetCreator()
	if err != nil {
        jsonResp = "{\"Error\":\"Failed to get Creator\"}"
        return shim.Error(jsonResp)
    }
	sId := &msp.SerializedIdentity{}
	proto.Unmarshal(creator, sId)
	
	certificate := string(sId.IdBytes)
	name := getNameFromCertificate([] byte(certificate))
	
	err = json.Unmarshal([]byte(args[0]), &ip)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}
	// ==== Check if marble already exists ====
	itemAsBytes, err := stub.GetState(ip.MobileId)
	if err != nil {
		return shim.Error("Failed to get Item: " + err.Error())
	} else if itemAsBytes != nil {
		fmt.Println("This Item already exists: " + ip.MobileId)
		return shim.Error("This Item already exists: " + ip.MobileId)
	}

	itemPD := &dataToSave{}
	itemPD.MobileId = ip.MobileId
	itemPD.Owner = name
	itemPD.Status = ip.Status

	itemJSON, err := json.Marshal(itemPD)
	if err != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Item. \"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutState(itemPD.MobileId, itemJSON)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("Data Set Successfully!!"))
}

func (c *Test) updateData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect Number of Arguments. Expecting 1")
	}
	creator, err := stub.GetCreator()
	if err != nil {
        jsonResp = "{\"Error\":\"Failed to get Creator\"}"
        return shim.Error(jsonResp)
    }
	
	sId := &msp.SerializedIdentity{}
	proto.Unmarshal(creator, sId)
	
	certificate := string(sId.IdBytes)
	name := getNameFromCertificate([] byte(certificate))
	

	var ip dataToSave
	err = json.Unmarshal([]byte(args[0]), &ip)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}
	itemAsBytes, err := stub.GetState(ip.MobileId)
	if err != nil {
		return shim.Error("Failed to get Item: " + err.Error())
	} 
	var existingData dataToSave
	err = json.Unmarshal([]byte(itemAsBytes), &existingData)
	if err != nil {
		return shim.Error("Cannot Unmarshall the data")
	}
	if strings.Compare(name,existingData.Owner) == 0 {
		existingData.Status= ip.Status
		dataJSON, _ := json.Marshal(existingData)
		err = stub.PutState(existingData.MobileId, dataJSON)
		if err != nil {
			return shim.Error("Put State Error")
		}
	} else {
		return shim.Error("Cant Update the Status: Invalid Certificate")
	}
	return shim.Success([]byte("Update Status Successfull"))

}
func (c *Test) getData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of Arguments.")
	}
	var itemP inputItem
	err := json.Unmarshal([]byte(args[0]), &itemP)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}
	valAsbytes, err := stub.GetState(itemP.MobileId) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + itemP.MobileId + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + itemP.MobileId + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valAsbytes)
}
func (c *Test) retrieve(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of Arguments.")
	}
	var itemP inputItem
	err := json.Unmarshal([]byte(args[0]), &itemP)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}
	valAsbytes, err := stub.GetPrivateData("PrivateCollection", itemP.MobileId) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + itemP.MobileId + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + itemP.MobileId + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valAsbytes)
}

func (c *Test) retrieveRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	key := strings.ToLower(args[0])
	value := strings.ToLower(args[1])
	collection := args[2]
	queryString := fmt.Sprintf("{\"selector\":{\"%s\":\"%s\"}}",key,value)
	fmt.Println("QueryString: ", queryString)
	queryResults, err := getQueryResultForQueryString(stub, queryString, collection)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string, collection string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetPrivateDataQueryResult(collection, queryString)
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
func (c *Test) retrieveWithPrice(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of Arguments.")
	}
	var itemP inputItem
	err := json.Unmarshal([]byte(args[0]), &itemP)
	if err != nil {
		return shim.Error("Invalid json provided as input")
	}
	valAsbytes, err := stub.GetPrivateData("PrivatePriceCollection", itemP.MobileId) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + itemP.MobileId + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + itemP.MobileId + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valAsbytes)
}

func getNameFromCertificate(certificate []byte)(string) {
    block, _ := pem.Decode([]byte(certificate))
    cert, _ := x509.ParseCertificate(block.Bytes)
    name:= cert.Subject.CommonName
    return name
}

//Main function for Training Chaincode
func main() {
	err := shim.Start(new(Test))
	if err != nil {
		fmt.Printf("Error Starting Test Chaincode : %s", err)
	}
}
