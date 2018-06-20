package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/pem"
	"crypto/x509"
	"strings"
	"time"
	"encoding/json"
	"fmt"
)

var logger = shim.NewLogger("SimpleChaincode")

type OrgData [] Product

type Product struct {
	ID          string `json:"productID"`
	ObjectType  string `json:"productObjectType"`
	Name        string `json:"productName"`
	Desc        string `json:"productDesc"`
	State       string `json:"productState"`
	Org         string `json:"productOrg"`
	DateCreated time.Time
	DateUpdated time.Time
}

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")

	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error(err.Error())
	}

	name, org := getCreator(creatorBytes)

	logger.Debug("transaction creator " + name + "@" + org)

	function, args := stub.GetFunctionAndParameters()
	if function == "add" {
		// Add product to organisation
		return t.add(stub, args)
	} else if function == "deleteOrg" {
		// Deletes organisation
		return t.deleteOrg(stub, args)
	} else if function == "deleteProduct" {
		// Deletes product from organisation
		return t.deleteProduct(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemented in invoke
		return t.query(stub, args)
	}

	return pb.Response{Status: 403, Message: "Invalid invoke function name."}
}

// Transaction makes adding product to org
func (t *SimpleChaincode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var OrgDataJSON OrgData
	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	fmt.Println("Add product")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}

	productName := args[0]
	productDesc := args[1]
	productState := args[2]
	productOrg := args[3]

	// Get Organisation data
	OrgAsBytes, err := stub.GetState(productOrg)
	if err != nil {
		return shim.Error("Failed to get Org: " + err.Error())
	}

	err = json.Unmarshal([]byte(OrgAsBytes), &OrgDataJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of Org: " + productOrg + "\"}"
		return shim.Error(jsonResp)
	}

	// Create product object and marshal to JSON
	productID := uuid.Must(uuid.NewV4())
	DateCreated := time.Now()
	DateUpdated := time.Now()
	objectType := "product"
	product := Product{productID, objectType, productName, productDesc, productState,
		productOrg, DateCreated, DateUpdated}
	OrgDataJSON = append(OrgDataJSON, product)
	productJSONasBytes, err := json.Marshal(OrgDataJSON)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Write the product to the ledger
	err = stub.PutState(productOrg, productJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("End add product")

	return shim.Success(nil)
}

// deletes an entity from state
func (t *SimpleChaincode) deleteOrg(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return pb.Response{Status: 403, Message: "Incorrect number of arguments"}
	}

	Org := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(Org)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// deletes product from from organisation
func (t *SimpleChaincode) deleteProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var OrgDataJSON OrgData

	if len(args) != 2 {
		return pb.Response{Status: 403, Message: "Incorrect number of arguments"}
	}

	productOrg := args[0]
	//productID := args[1]

	// Get Organisation data
	OrgAsBytes, err := stub.GetState(productOrg)
	if err != nil {
		return shim.Error("Failed to get Org: " + err.Error())
	}

	err = json.Unmarshal([]byte(OrgAsBytes), &OrgDataJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of Org: " + productOrg + "\"}"
		return shim.Error(jsonResp)
	}

	// Delete the key from the state in ledger
	//err := stub.DelState(Org)
	//if err != nil {
	//	return shim.Error(err.Error())
	//}

	return shim.Success(nil)
}

// read value
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var a string // Entities
	var err error

	if len(args) != 1 {
		return pb.Response{Status: 403, Message: "Incorrect number of arguments"}
	}

	productName = args[0]

	// Get the state from the ledger
	productBytes, err := stub.GetState(productName)
	if err != nil {
		return shim.Error(err.Error())
	}

	if productBytes == nil {
		return pb.Response{Status: 404, Message: "Entity not found"}
	}

	return shim.Success(productBytes)
}

var getCreator = func(certificate []byte) (string, string) {
	data := certificate[strings.Index(string(certificate), "-----") : strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]
	commonName := cert.Subject.CommonName
	logger.Debug("commonName: " + commonName + ", organization: " + organization)

	organizationShort := strings.Split(organization, ".")[0]

	return commonName, organizationShort
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		logger.Error(err.Error())
	}
}
