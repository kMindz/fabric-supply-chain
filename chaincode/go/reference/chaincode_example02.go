package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/satori/go.uuid"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/pem"
	"crypto/x509"
	"strings"
	"encoding/json"
	"fmt"
	"strconv"
)

var logger = shim.NewLogger("SimpleChaincode")

const objectType = "product"

var productStateMap = map[int]string{
	1: "Registered",
	2: "Active",
	3: "Decision-making",
	4: "Inactive",
}

type CompositeKey struct {
	ID          string `json:"id"`
	Org         string `json:"org"`
	ProductName string `json:"productName"`
}

type Product struct {
	ID          string `json:"productID"`
	Name        string `json:"productName"`
	Desc        string `json:"productDesc"`
	State       int    `json:"productState"`
	Org         string `json:"productOrg"`
	DateUpdated int    `json:"productDateUpdated"`
}

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")

	_, args := stub.GetFunctionAndParameters()
	var a, b string    // Entities
	var aVal,  bVal int // Asset holdings
	var err error

	if len(args) != 4 {
		return pb.Response{Status:403, Message:"Incorrect number of arguments. Expecting 4"}
	}

	// Initialize the chaincode
	a = args[0]
	aVal, err = strconv.Atoi(args[1])
	if err != nil {
		return pb.Response{Status:403, Message:"Expecting integer value for asset holding"}
	}
	b = args[2]
	bVal, err = strconv.Atoi(args[3])
	if err != nil {
		return pb.Response{Status:403, Message:"Expecting integer value for asset holding"}
	}
	logger.Debugf("aVal, bVal = %d", aVal, bVal)

	// Write the state to the ledger
	err = stub.PutState(a, []byte(strconv.Itoa(aVal)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(b, []byte(strconv.Itoa(bVal)))
	if err != nil {
		return shim.Error(err.Error())
	}

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
	} else if function == "move" {
		// Updates product from organisation
		return t.move(stub, args)
	} else if function == "update" {
		// Updates product from organisation
		return t.update(stub, args)
	} else if function == "delete" {
		// Deletes product from organisation
		return t.delete(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemented in invoke
		return t.query(stub, args)
	}

	return pb.Response{Status: 403, Message: "Invalid invoke function name."}
}

// Transaction makes payment of x units from a to b
func (t *SimpleChaincode) move(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	return shim.Success(nil)
}

// Transaction makes adding product
func (t *SimpleChaincode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	fmt.Println("Add product")

	for k, v := range args {
		if len(v) <= 0 {
			return shim.Error(strconv.Itoa(k+1) + "st argument must be a non-empty string")
		}
	}

	productName := args[0]
	productDesc := args[1]
	productState := args[2]
	productOrg := args[3]
	productDateUpdated, err := strconv.Atoi(args[4])

	productStateInt, legalState := mapKey(productStateMap, productState)
	if !legalState {
		return shim.Error("Not legal product state")
	}

	productID := uuid.Must(uuid.NewV4())
	key := &CompositeKey{productID.String(), productOrg, productName}
	productData := &Product{productID.String(), productName, productDesc, productStateInt, productOrg, productDateUpdated}

	dataKey, err := stub.CreateCompositeKey(objectType, []string{key.ID, key.Org, key.ProductName})
	if err != nil {
		return shim.Error("Couldn't create composite key " + err.Error())
	}

	productJSONasBytes, err := json.Marshal(productData)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(dataKey, productJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("End add product")

	return shim.Success(productJSONasBytes)
}

// Transaction makes updating product
func (t *SimpleChaincode) update(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var productData Product

	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}

	fmt.Println("Update product")

	for k, v := range args {
		if len(v) <= 0 {
			return shim.Error(strconv.Itoa(k+1) + "st argument must be a non-empty string")
		}
	}

	productName := args[0]
	productDesc := args[1]
	productState := args[2]
	productOrg := args[3]
	productDateUpdated, err := strconv.Atoi(args[4])
	productID := args[5]

	productStateInt, legalState := mapKey(productStateMap, productState)
	if !legalState {
		return shim.Error("Not legal product state")
	}

	key := &CompositeKey{productID, productOrg, productName}

	dataKey, err := stub.CreateCompositeKey(objectType, []string{key.ID, key.Org, key.ProductName})
	if err != nil {
		return shim.Error("Couldn't create composite key " + err.Error())
	}

	// Get the state from the ledger
	productJSONasBytes, err := stub.GetState(dataKey)
	if err != nil {
		return shim.Error(err.Error())
	}

	if productJSONasBytes == nil {
		return pb.Response{Status: 404, Message: "Entity not found"}
	}

	err = json.Unmarshal(productJSONasBytes, &productData)
	if err != nil {
		return shim.Error(err.Error())
	}
	productData.Name = productName
	productData.Desc = productDesc
	productData.State = productStateInt
	productData.DateUpdated = productDateUpdated

	productJSONasBytes, err = json.Marshal(productData)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(dataKey, productJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("End add product")

	return shim.Success(productJSONasBytes)
}

// deletes product
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return pb.Response{Status: 403, Message: "Incorrect number of arguments"}
	}

	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	productID := args[0]

	//Delete the key from the state in ledger
	err := stub.DelState(productID)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// read value
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//var a string // Entities
	//var err error
	//
	//if len(args) != 1 {
	//	return pb.Response{Status: 403, Message: "Incorrect number of arguments"}
	//}
	//
	//productName = args[0]
	//
	//// Get the state from the ledger
	//productBytes, err := stub.GetState(productName)
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	//
	//if productBytes == nil {
	//	return pb.Response{Status: 404, Message: "Entity not found"}
	//}
	//
	//return shim.Success(productBytes)
	return shim.Success(nil)
}

var mapKey = func(m map[int]string, value string) (key int, ok bool) {
	for k, v := range m {
		if v == value {
			key = k
			ok = true
			return
		}
	}
	return
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
