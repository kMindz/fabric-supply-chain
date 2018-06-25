package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/satori/go.uuid"
	"strconv"
	"strings"
)

var logger = shim.NewLogger("SimpleChaincode")

const objectType = "product"

type CompositeKey struct {
	Org         string `json:"org"`
	ProductName string `json:"productName"`
	ID          string `json:"id"`
}

type Product struct {
	ID          string `json:"productID"`
	Name        string `json:"productName"`
	Desc        string `json:"productDesc"`
	State       int    `json:"productState"`
	Org         string `json:"productOrg"`
	DateUpdated int    `json:"productDateUpdated"`
}

type Org struct {
	Name        string `json:"orgName"`
	Products    [] Product `json:"products"`
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
	productState, _ := strconv.Atoi(args[2])
	productOrg := args[3]
	productDateUpdated, err := strconv.Atoi(args[4])

	productID := uuid.Must(uuid.NewV4())
	key := &CompositeKey{productOrg, productName, productID.String()}
	productData := &Product{productID.String(), productName, productDesc, productState, productOrg, productDateUpdated}

	dataKey, err := stub.CreateCompositeKey(objectType, []string{key.Org, key.ProductName, key.ID})
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
	productState, _ := strconv.Atoi(args[2])
	productOrg := args[3]
	productDateUpdated, err := strconv.Atoi(args[4])
	productID := args[5]

	key := &CompositeKey{productOrg, productName, productID}

	dataKey, err := stub.CreateCompositeKey(objectType, []string{key.Org, key.ProductName, key.ID})
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
	productData.State = productState
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

	productOrg := args[0]

	var Orgs []Org

	var Org Org

	var productsData []Product

	var productData Product

	var keyPart []string


	key := &CompositeKey{productOrg, "", ""}

	if len(args) != 1 {
		return pb.Response{Status: 403, Message: "Incorrect number of arguments"}
	}

	// Get the state from the ledger
	if args[0] != "all" {
		keyPart = append(keyPart, key.Org)
	}

	productResultsIterator, err := stub.GetStateByPartialCompositeKey(objectType, keyPart)

	if err != nil {
		return shim.Error(err.Error())
	}

	defer productResultsIterator.Close()

	// Iterate through result set and for each product found
	var i int
	for i = 0; productResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the marble name from the composite key
		responseRange, err := productResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		//productData := &Product{productID.String(), productName, productDesc, productStateInt, productOrg, productDateUpdated}

		productJSONasBytes := responseRange.Value

		if productJSONasBytes == nil {
			return pb.Response{Status: 404, Message: "Entity not found"}
		}

		err = json.Unmarshal(productJSONasBytes, &productData)
		if err != nil {
			return shim.Error(err.Error())
		}

		productsData = append(productsData, productData)

	}
	Org.Name = productOrg
	Org.Products = productsData

	Orgs = append(Orgs, Org)

	OrgProductsJSONasBytes, err := json.Marshal(Orgs)
	if err != nil {
		return shim.Error(err.Error())
	}

	if OrgProductsJSONasBytes == nil {
		return pb.Response{Status: 404, Message: "Entity not found"}
	}

	return shim.Success(OrgProductsJSONasBytes)
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
