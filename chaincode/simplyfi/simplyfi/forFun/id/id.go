package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"crypto/x509"
    "encoding/pem"
	"github.com/hyperledger/fabric/core/chaincode/shim" // import for Chaincode Interface
	pb "github.com/hyperledger/fabric/protos/peer"      // import for peer response
	id "github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
)

type Test struct {
}

//Init function of the chaincode
func (c *Test) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

//Invoke function of the chaincode
func (c *Test) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, _ := stub.GetFunctionAndParameters()
	switch function {
	case "getID":
		return c.getID(stub)
	case "getName":
		return c.getName(stub)	
	default:
		return shim.Error("Not a Valid Function.")
	}
}
func (c *Test) getID(stub shim.ChaincodeStubInterface) pb.Response {
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

func (c *Test) getName(stub shim.ChaincodeStubInterface) pb.Response {
	enCert, err := id.GetX509Certificate(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	cert :=  enCert.Raw
	typ := reflect.TypeOf(cert).String()
	//name:= getNameFromCertificate(cert)
	fmt.Println(typ)
	//block, _ := pem.Decode([]byte(cert))
    cert1, _ := x509.ParseCertificate(cert.Bytes)
	//name:= cert1.Subject.CommonName
	fmt.Println(typ)
	resultData := map[string]interface{}{
		"status":    "true",
		"name": []byte(cert),
		"type":typ,
	//	"actualName":name,
	}
	respJson, _ := json.Marshal(resultData)
	return shim.Success(respJson)
	//name:= getNameFromCertificate([]byte(cert))
	//resultData := map[string]interface{}{
	//	"status":    "true",
	//	"name": name,
	//}
	//respJson, _ := json.Marshal(resultData)
	//return shim.Success(respJson)
	// issuersOrgs := enCert.Issuer.Organization
	// if len(issuersOrgs) == 0 {
	// 	return false, "Unknown.."
	// }
	// return true, fmt.Sprintf("%s", issuersOrgs[0])
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