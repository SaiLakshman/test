package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim" // import for Chaincode Interface
	pb "github.com/hyperledger/fabric/protos/peer"      // import for peer response
)

type Training struct {
}

//Employee Data
type Employee struct {
	EmployeeId      string `json:"empId"`
	EmployeeName 	string `json:"empName"`
	Designation     string `json:"designation"`
	ContactNumber   string `json:"contactNumber"`
	Hobby        	string `json:"hobby"`
}

//Init function of the chaincode
func (c *Training) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

//Invoke function of the chaincode
func (c *Training) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "create":
		return c.createEmployee(stub, args)
	case "retrieve":
		return c.queryEmployee(stub, args)
	case "update":
		return c.updateEmployeeHobby(stub, args)
	case "queryByName":
		return c.queryEmployeeByName(stub, args)
	case "history":
		return c.getEmployeeHistory(stub, args)
	default:
		return shim.Error("Not a Valid Function.")
	}
}

//Creating Employee in Blockchain
/*
	fcnName: createEmployee
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [empId, empName, designation, contactNumber, hobby]
	create a employee object of the structure entity,
	set all the necessary data of that object
	marshall the object into jsonObject
	create a tx in blockchain using PutState
	return pb.Response= "Employee Created Successfully"
*/
func (c *Training) createEmployee(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 5 {
		return shim.Error("Incorrect Number of Arguments. Expecting 8(EmployeeId, EmployeeName, Designation, ContactNumber, Hobby))")
	}
	empId := args[0]
	empName := args[1]
	designation := args[2]
	contactNumber := args[3]
	hobby := args[4]

	empStruct := &Employee{}
	empStruct.EmployeeId = empId
	empStruct.EmployeeName = empName
	empStruct.Designation = designation
	empStruct.ContactNumber = contactNumber
	empStruct.Hobby = hobby
	
	empAsBytes, err := json.Marshal(empStruct)
	if err != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Entity. \"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutState(empId, empAsBytes)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed: Creating Employee Data\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Employee Created Successfully!!"))
}

//Retrieving Employee given EmployeeId from Blockchain
/*
	fcnName: queryEmployee
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [empId]
	retrieve the tx from blockchain using GetState
	return pb.Response= Payload of employees which was created, else error saying does not exist
*/
func (c *Training) queryEmployee(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of Arguments. Expecting 1(EmployeeId)")
	}
	empId := args[0]
	valueAsBytes, err := stub.GetState(empId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + empId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Employee does not exist: " + empId + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valueAsBytes)
}
//Modify & Update the Hobby of Employee in Blockchain
/*
	fcnName: updateEmployeeHobby
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [empId, hobby]
	retrieve the employee from blockchain given employeeId
	create a employee object from employee structure
	unmarshal the retrieved data
	set the status
	marshal the data and create the data in blockchain using PutState
	return pb.Response= "Employee Hobby Updated Successfully"
*/
func (c *Training) updateEmployeeHobby(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	if len(args) != 2 {
		return shim.Error("Incorrect number of Arguments. Expecting 2(EmpId, Hobby)")
	}
	empId := args[0]
	hobby := args[1]
	valueAsBytes, err := stub.GetState(empId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + empId + "\"}"
		return shim.Error(jsonResp)
	} else if valueAsBytes == nil {
		jsonResp = "{\"Error\" : \"Employee does not exist: " + empId + "\"}"
		return shim.Error(jsonResp)
	}
	empStruct := &Employee{}
	err = json.Unmarshal(valueAsBytes, &empStruct)
	empStruct.Hobby = hobby
	empAsBytes, err1 := json.Marshal(empStruct)
	if err1 != nil {
		jsonResp = "{\"Error\":\"JSON Marshalling Error for Entity Status Updation \"}"
		return shim.Error(jsonResp)
	}
	err1 = stub.PutState(empId, empAsBytes)
	if err1 != nil {
		jsonResp = "{\"Error\":\"Updating Employee Hobby Failed \"}"
		return shim.Error(jsonResp)
	}
	return shim.Success([]byte("Employee Hobby Updated Successfully!!"))
}
//Retrieving Employee By Name from Blockchain
/*
	fcnName: queryEmployeeByName
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [empId]
	retrieve the employee using empName from blockchain using complex query(selector query)
	return pb.Response= Payload of the employee under that empName
*/
func (c *Training) queryEmployeeByName(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonresp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. expecting 1(EmpName)")
	}
	empName := args[0]
	querystring := fmt.Sprintf("{\"selector\":{\"empName\":\"%s\"}}", empName)
	queryresults, err := getQueryResultForQueryString(stub, querystring)
	if err != nil {
		jsonresp = "{\"Error\":\"Failed to get state for " + empName + "\"}"
		return shim.Error(jsonresp)
	}
	return shim.Success(queryresults)
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

//Retrieving Employee History given EmpID from Blockchain
/*
	fcnName: getEmployeeHistory
	arguments: 2
	argument1: chaincode stub interface
	argument2: array consists of [empId]
	return pb.Response= Payload of the employee under that empId from Blockchain
*/
func (c *Training) getEmployeeHistory(stub shim.ChaincodeStubInterface, args []string)pb.Response {

        if len(args) != 1 {
            return shim.Error("Incorrect number of arguments. Expecting 1(EmpId)")
        }
        empId := args[0]
        resultsIterator, err := stub.GetHistoryForKey(empId)
        if err != nil {
            return shim.Error(err.Error())
        }
        defer resultsIterator.Close()
        // buffer is a JSON array containing historic values for the user
        var buffer bytes.Buffer
        buffer.WriteString("[")

        bArrayMemberAlreadyWritten := false
        for resultsIterator.HasNext() {
                response, err := resultsIterator.Next()
                if err != nil {
                        return shim.Error(err.Error())
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
        return shim.Success(buffer.Bytes())

}
//Main function for Training Chaincode
func main() {
	err := shim.Start(new(Training))
	if err != nil {
		fmt.Printf("Error Starting Training Chaincode : %s", err)
	}
}

