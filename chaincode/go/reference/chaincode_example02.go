package main

import (
	"strconv"
	"bytes"
	"encoding/json"
	"fmt"
	"time"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/pem"
	"crypto/x509"
	"strings"
	"net/url"
	)

var logger = shim.NewLogger("SimpleChaincode")

const (
	stateUnknown = iota
	stateRegistered
	stateActive
	stateDecisionMaking
	stateInactive
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type Product struct {
	ObjectType  string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Name        string `json:"name"`    //the fieldtags are needed to keep case from bouncing around
	Desc        string `json:"desc"`
	State       int    `json:"state"`
	LastUpdated int    `json:"lastUpdated"`
	Owner       string `json:"owner"`
}

var productStateMachine = map[int][]int{
	stateRegistered: {stateRegistered, stateActive},
	stateActive: {stateActive, stateDecisionMaking},
	stateDecisionMaking: {stateActive, stateDecisionMaking, stateInactive},
	stateInactive: {stateInactive},
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		logger.Error(err.Error())
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")
	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error(err.Error())
	}

	name, org := getCreator(creatorBytes)

	logger.Debug("transaction creator " + name + "@" + org)

	function, args := stub.GetFunctionAndParameters()
	logger.Debug("invoke is running " + function)

	// Handle different functions
	if function == "initProduct" { //create a new product
		return t.initProduct(stub, args)
	} else if function == "updateProduct" { //update a existing product
		return t.updateProduct(stub, args)
	} else if function == "updateOwner" { //update a existing product
		return t.updateOwner(stub, args)
	} else if function == "transferProduct" { //change owner of a specific product
		return t.transferProduct(stub, args)
	} else if function == "transferProductsBasedOnState" { //transfer all products of a certain state
		return t.transferProductsBasedOnState(stub, args)
	} else if function == "delete" { //delete a product
		return t.delete(stub, args)
	} else if function == "readProduct" { //read a product
		return t.readProduct(stub, args)
	} else if function == "queryProductsByOwner" { //find products for owner X using rich query
		return t.queryProductsByOwner(stub, args)
	} else if function == "queryProducts" { //find products based on an ad hoc rich query
		return t.queryProducts(stub, args)
	} else if function == "getHistoryForProduct" { //get history of values for a product
		return t.getHistoryForProduct(stub, args)
	} else if function == "getProductsByRange" { //get products based on range query
		return t.getProductsByRange(stub, args)
	}

	logger.Debug("invoke did not find func: " + function) //error
	return pb.Response{Status: 403, Message: "Invalid invoke function name."}
}

// ============================================================
// initProduct - create a new product, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0      1               2             3       4
	// "Book", "Go Tutorial", "Registered", "OrgA", "1530532962"
	if len(args) < 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	// ==== Input sanitation ====
	for k, v := range args {
		if len(v) <= 0 {
			return shim.Error(strconv.Itoa(k+1) + "st argument must be a non-empty string")
		}
	}

	productName := args[0]
	desc := args[1]
	state, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Product state must be int. Error: " + err.Error())
	}
	owner := strings.ToLower(args[3])
	lastUpdated, err := strconv.Atoi(args[4])
	if err != nil {
		return shim.Error("Product date updated must be timestamp. Error: " + err.Error())
	}
	// ==== Check if state not legal ====
	legalState := mapKey(productStateMachine, state)
	if !legalState {
		return shim.Error("Not legal product state") // only "Registered" state when add new product
	}
	// ==== Check if product already exists ====
	productAsBytes, err := stub.GetState(productName)
	if err != nil {
		return shim.Error("Failed to get product: " + err.Error())
	} else if productAsBytes != nil {
		logger.Debug("This product already exists: " + productName)
		return shim.Error("This product already exists: " + productName)
	}

	// ==== Create product object and marshal to JSON ====
	objectType := "product"
	product := &Product{objectType, productName, desc, state, lastUpdated, owner}
	productJSONasBytes, err := json.Marshal(product)
	if err != nil {
		return shim.Error(err.Error())
	}
	//Alternatively, build the product json string manually if you don't want to use struct marshalling
	//productJSONasString := `{"docType":"product",  "name": "` + productName + `", "desc": "` + desc + `", "state": ` + strconv.Itoa(state) + `, "lastUpdated": ` + strconv.Itoa(lastUpdated) + `, "owner": "` + owner + `"}`
	//productJSONasBytes := []byte(productJSONasString)

	// === Save product to state ===
	err = stub.PutState(productName, productJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the product to enable state-based range queries, e.g. return all Active products ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~state~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~state~*
	indexName := "state~name"
	stateIndexKey, err := stub.CreateCompositeKey(indexName, []string{strconv.Itoa(product.State), product.Name})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the product.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(stateIndexKey, value)

	// ==== Product saved and indexed. Return success ====
	logger.Debug("- end init product")
	return shim.Success(nil)
}

// ============================================================
// updateProduct - update a existing product, store into chaincode state
// ============================================================
func (t *SimpleChaincode) updateProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0      1               2             3       4
	// "Book", "Go Tutorial", "Registered", "OrgA", "1530532962"
	if len(args) < 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	// ==== Input sanitation ====
	for k, v := range args {
		if len(v) <= 0 {
			return shim.Error(strconv.Itoa(k+1) + "st argument must be a non-empty string")
		}
	}

	productName := args[0]
	newDesc := args[1]
	newState, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Product state must be int. Error: " + err.Error())
	}
	newOwner := strings.ToLower(args[3])
	lastUpdated, err := strconv.Atoi(args[4])
	if err != nil {
		return shim.Error("Product date updated must be timestamp. Error: " + err.Error())
	}

	productAsBytes, err := stub.GetState(productName)
	if err != nil {
		return shim.Error("Failed to get product:" + err.Error())
	} else if productAsBytes == nil {
		return shim.Error("Product does not exist")
	}

	productToUpdate := Product{}
	err = json.Unmarshal(productAsBytes, &productToUpdate) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Check if state not legal ====
	legalState := mapKey(productStateMachine, newState)
	legalStateMachine := checkNewState(productStateMachine, productToUpdate.State, newState)
	if !legalState || !legalStateMachine {
		return shim.Error("Not legal product state")
	}
	productToUpdate.Desc = newDesc
	productToUpdate.State = newState
	productToUpdate.Owner = newOwner
	productToUpdate.LastUpdated = lastUpdated

	productJSONasBytes, _ := json.Marshal(productToUpdate)
	err = stub.PutState(productName, productJSONasBytes) //rewrite the product
	if err != nil {
		return shim.Error(err.Error())
	}

	// maintain the index
	if productToUpdate.State != newState {
		//delete old index
		indexName := "state~name"
		stateIndexKey, err := stub.CreateCompositeKey(indexName, []string{strconv.Itoa(productToUpdate.State), productToUpdate.Name})
		if err != nil {
			return shim.Error(err.Error())
		}

		//  Delete index entry to state.
		err = stub.DelState(stateIndexKey)
		if err != nil {
			return shim.Error("Failed to delete state:" + err.Error())
		}
		//create new index
		stateIndexKey, err = stub.CreateCompositeKey(indexName, []string{strconv.Itoa(newState), productToUpdate.Name})
		if err != nil {
			return shim.Error(err.Error())
		}
		//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the product.
		//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
		value := []byte{0x00}
		stub.PutState(stateIndexKey, value)
	}
	// ==== Product updated and indexed. Return success ====
	logger.Debug("- end update product")
	return shim.Success(nil)
}

// ===============================================
// readProduct - read a product from chaincode state
// ===============================================
func (t *SimpleChaincode) readProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the product to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name) //get the product from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Prodcut does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ==================================================
// delete - remove a product key/value pair from state
// ==================================================
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var productJSON Product
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	productName := args[0]

	// to maintain the state~name index, we need to read the product first and get its state
	valAsbytes, err := stub.GetState(productName) //get the product from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + productName + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Product does not exist: " + productName + "\"}"
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal([]byte(valAsbytes), &productJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + productName + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(productName) //remove the product from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// maintain the index
	indexName := "state~name"
	stateIndexKey, err := stub.CreateCompositeKey(indexName, []string{strconv.Itoa(productJSON.State), productJSON.Name})
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Delete index entry to state.
	err = stub.DelState(stateIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	return shim.Success(nil)
}

// ===========================================================
// transfer a product by setting a new owner name on the product
// ===========================================================
func (t *SimpleChaincode) transferProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       1
	// "book", "OrgB"
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	productName := args[0]
	newOwner := strings.ToLower(args[1])
	logger.Debug("- start transferProduct ", productName, newOwner)

	productAsBytes, err := stub.GetState(productName)
	if err != nil {
		return shim.Error("Failed to get product:" + err.Error())
	} else if productAsBytes == nil {
		return shim.Error("Product does not exist")
	}

	productToTransfer := Product{}
	err = json.Unmarshal(productAsBytes, &productToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	productToTransfer.Owner = newOwner //change the owner

	productJSONasBytes, _ := json.Marshal(productToTransfer)
	err = stub.PutState(productName, productJSONasBytes) //rewrite the product
	if err != nil {
		return shim.Error(err.Error())
	}

	logger.Debug("- end transferProduct (success)")
	return shim.Success(nil)
}

// ===========================================================================================
// getProductsByRange performs a range query based on the start and end keys provided.

// Read-only function results are not typically submitted to ordering. If the read-only
// results are submitted to ordering, or if the query is used in an update transaction
// and submitted to ordering, then the committing peers will re-execute to guarantee that
// result sets are stable between endorsement time and commit time. The transaction is
// invalidated by the committing peers if the result set has changed between endorsement
// time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) getProductsByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
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

	logger.Debug("- getProductsByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// ==== Example: GetStateByPartialCompositeKey/RangeQuery =========================================
// transferProductsBasedOnState will transfer products of a given state to a certain new owner.
// Uses a GetStateByPartialCompositeKey (range query) against state~name 'index'.
// Committing peers will re-execute range queries to guarantee that result sets are stable
// between endorsement time and commit time. The transaction is invalidated by the
// committing peers if the result set has changed between endorsement time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) transferProductsBasedOnState(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0        1
	// "Active", "OrgB"
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	state := args[0]
	newOwner := strings.ToLower(args[1])
	logger.Debug("- start transferProductsBasedOnState ", state, newOwner)

	// Query the state~name index by state
	// This will execute a key range query on all keys starting with 'state'
	statedProductResultsIterator, err := stub.GetStateByPartialCompositeKey("state~name", []string{state})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer statedProductResultsIterator.Close()

	// Iterate through result set and for each product found, transfer to newOwner
	var i int
	for i = 0; statedProductResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the product name from the composite key
		responseRange, err := statedProductResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// get the state and name from state~name composite key
		objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		returnedState := compositeKeyParts[0]
		returnedProductName := compositeKeyParts[1]
		logger.Debug("- found a product from index:%s state:%s name:%s\n", objectType, returnedState, returnedProductName)

		// Now call the transfer function for the found product.
		// Re-use the same function that is used to transfer individual products
		response := t.transferProduct(stub, []string{returnedProductName, newOwner})
		// if the transfer failed break out of loop and return error
		if response.Status != shim.OK {
			return shim.Error("Transfer failed: " + response.Message)
		}
	}

	responsePayload := fmt.Sprintf("Transferred %d %s products to %s", i, state, newOwner)
	fmt.Println("- end transferProductsBasedOnState: " + responsePayload)
	return shim.Success([]byte(responsePayload))
}

// =======Rich queries =========================================================================
// Two examples of rich queries are provided below (parameterized query and ad hoc query).
// Rich queries pass a query string to the state database.
// Rich queries are only supported by state database implementations
//  that support rich query (e.g. CouchDB).
// The query string is in the syntax of the underlying state database.
// With rich queries there is no guarantee that the result set hasn't changed between
//  endorsement time and commit time, aka 'phantom reads'.
// Therefore, rich queries should not be used in update transactions, unless the
// application handles the possibility of result set changes between endorsement and commit time.
// Rich queries can be used for point-in-time queries against a peer.
// ============================================================================================

// ===== Example: Parameterized rich query =================================================
// queryProductsByOwner queries for products based on a passed in owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryProductsByOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "bob"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	owner := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"product\",\"owner\":\"%s\"}}", owner)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// ===== Example: Ad hoc rich query ========================================================
// queryProducts uses a query string to perform a query for products.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryProductsForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryProducts(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	in, err := url.QueryUnescape(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	queryString := fmt.Sprintf(in);

	fmt.Printf("- Show QueryString:\n%s\n", queryString)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)

}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
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

func (t *SimpleChaincode) getHistoryForProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	productName := args[0]

	logger.Debug("- start getHistoryForProduct: %s\n", productName)

	resultsIterator, err := stub.GetHistoryForKey(productName)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the product
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

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON product)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	logger.Debug("- getHistoryForProduct returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (t *SimpleChaincode) updateOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//      0          1         2          3
	// productName, oldOwner, newOwner, timestamp
	if len(args) < 4 {
		return shim.Error(fmt.Sprintf("Incorrect number of arguments. Expecting %d", 4))
	}

	// ==== Input sanitation ====
	for k, v := range args {
		if len(v) <= 0 {
			return shim.Error(fmt.Sprintf("Argument #%d must be a non-empty string", k + 1))
		}
	}

	productName := args[0]
	oldOwner := args[1]
	newOwner := args[2]
	lastUpdated, err := strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Product date updated must be timestamp. Error: " + err.Error())
	}

	// TODO: check if creator org and oldOwner are the same
	//if GetCreatorOrganization(stub) != oldOwner {
	//	return shim.Error(fmt.Sprintf("no privileges to send request from the side of %s", oldOwner))
	//}

	productAsBytes, err := stub.GetState(productName)
	if err != nil {
		return shim.Error("Failed to get product:" + err.Error())
	} else if productAsBytes == nil {
		return shim.Error("Product does not exist")
	}

	productToUpdate := Product{}
	err = json.Unmarshal(productAsBytes, &productToUpdate) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}

	if productToUpdate.Owner != oldOwner {
		return shim.Error("The specified product doesn't belong to the specified owner.")
	}

	productToUpdate.Owner = newOwner
	productToUpdate.LastUpdated = lastUpdated

	productJSONasBytes, _ := json.Marshal(productToUpdate)
	err = stub.PutState(productName, productJSONasBytes) //rewrite the product
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Product updated and indexed. Return success ====
	logger.Debug("- end update owner")
	return shim.Success(nil)
}

func getOrganization(certificate []byte) string {
	data := certificate[strings.Index(string(certificate), "-----") : strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]
	return organization
}

func GetCreatorOrganization(stub shim.ChaincodeStubInterface) string {
	certificate, _ := stub.GetCreator()
	return getOrganization(certificate)
}

var mapKey = func(m map[int][]int, value int) (check bool) {
	_, ok := m[value]
	if ok {
		check = true
	}

	return
}

var checkNewState = func(states map[int][]int, stateOld int, stateNew int) (check bool) {
	newStates, ok := states[stateOld]
	if ok {
		for _, v := range newStates {
			if v == stateNew {
				check = true
			}
		}
	}
	return
}
