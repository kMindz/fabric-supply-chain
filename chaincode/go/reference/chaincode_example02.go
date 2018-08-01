package main

import (
	"strconv"
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

var logger = shim.NewLogger("ProductChaincode")

const (
	productIndex = "product"
)

const (
	stateIndexName = "state~name"
)

const (
	stateUnknown = iota
	stateRegistered
	stateActive
	stateDecisionMaking
	stateInactive
)

// ProductChaincode example simple Chaincode implementation
type ProductChaincode struct {
}

type Product struct {
	Key   ProductKey   `json:"key"`
	Value ProductValue `json:"value"`
}

type ProductKey struct {
	Name string `json:"name"`
}

type ProductValue struct {
	ObjectType  string `json:"docType"`
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

// Init initializes chaincode
// ===========================
func (t *ProductChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *ProductChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")

	function, args := stub.GetFunctionAndParameters()
	logger.Debug("invoke is running " + function)

	// Handle different functions
	if function == "initProduct" { //create a new product
		return t.initProduct(stub, args)
	} else if function == "updateProduct" { //update an existing product
		return t.updateProduct(stub, args)
	} else if function == "updateOwner" { //update an owner of an existing product
		return t.updateOwner(stub, args)
	} else if function == "delete" { //delete a product
		return t.delete(stub, args)
	} else if function == "readProduct" { //read a product
		return t.readProduct(stub, args)
	} else if function == "queryProductsByOwner" { //find products for the owner X using rich query
		return t.queryProductsByOwner(stub, args)
	} else if function == "queryProducts" { //find products based on an ad hoc rich query
		return t.queryProducts(stub, args)
	} else if function == "getHistoryForProduct" { //get history of values for a product
		return t.getHistoryForProduct(stub, args)
	}

	logger.Debug("invoke did not find func: " + function) //error
	return pb.Response{Status: 403, Message: "Invalid invoke function name."}
}

// ============================================================
// initProduct - create a new product, store into chaincode state
// ============================================================
func (t *ProductChaincode) initProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	product := Product{
		Key:   ProductKey{productName},
		Value: ProductValue{productIndex,desc, state, lastUpdated, owner},
	}
	productJSONasBytes, err := json.Marshal(product.Value)
	if err != nil {
		return shim.Error(err.Error())
	}
	//Alternatively, build the product json string manually if you don't want to use struct marshalling
	//productJSONasString := `{"docType":"product",  "name": "` + productName + `", "desc": "` + desc + `", "state": ` + strconv.Itoa(state) + `, "lastUpdated": ` + strconv.Itoa(lastUpdated) + `, "owner": "` + owner + `"}`
	//productJSONasBytes := []byte(productJSONasString)

	// === Save product to state ===
	compositeKey, err := stub.CreateCompositeKey(productIndex, []string{product.Key.Name})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(compositeKey, productJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the product to enable state-based range queries, e.g. return all Active products ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on stateIndexName~state~name.
	//  This will enable very efficient state range queries based on composite keys matching stateIndexName~state~*
	stateIndexKey, err := stub.CreateCompositeKey(stateIndexName, []string{strconv.Itoa(product.Value.State), product.Key.Name})
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
func (t *ProductChaincode) updateProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	legalStateMachine := checkNewState(productStateMachine, productToUpdate.Value.State, newState)
	if !legalState || !legalStateMachine {
		return shim.Error("Not legal product state")
	}
	productToUpdate.Value.Desc = newDesc
	productToUpdate.Value.State = newState
	productToUpdate.Value.Owner = newOwner
	productToUpdate.Value.LastUpdated = lastUpdated

	productJSONasBytes, _ := json.Marshal(productToUpdate.Value)
	compositeKey, err := stub.CreateCompositeKey(productIndex, []string{productToUpdate.Key.Name})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(compositeKey, productJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// maintain the index
	if productToUpdate.Value.State != newState {
		//delete old index
		stateIndexKey, err := stub.CreateCompositeKey(stateIndexName,
			[]string{strconv.Itoa(productToUpdate.Value.State), productToUpdate.Key.Name})
		if err != nil {
			return shim.Error(err.Error())
		}

		//  Delete index entry to state.
		err = stub.DelState(stateIndexKey)
		if err != nil {
			return shim.Error("Failed to delete state:" + err.Error())
		}
		//create new index
		stateIndexKey, err = stub.CreateCompositeKey(stateIndexName,
			[]string{strconv.Itoa(newState), productToUpdate.Key.Name})
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
func (t *ProductChaincode) readProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name string
	var err error

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the product to query")
	}

	name = args[0]
	compositeKey, err := stub.CreateCompositeKey(productIndex, []string{name})
	if err != nil {
		return shim.Error(err.Error())
	}

	stateAsBytes, err := stub.GetState(compositeKey) //get the product from chaincode state
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to get state for product %s", name))
	} else if stateAsBytes == nil {
		return shim.Error(fmt.Sprintf("product %s doesn't exist", name))
	}

	product := Product{
		Key: ProductKey{name},
	}
	if err := json.Unmarshal(stateAsBytes, &product.Value); err != nil {
		return shim.Error(err.Error())
	}

	result, _ := json.Marshal(product)

	return shim.Success(result)
}

// ==================================================
// delete - remove a product key/value pair from state
// ==================================================
func (t *ProductChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//var jsonResp string
	//var productJSON Product
	//if len(args) < 1 {
	//	return shim.Error("Incorrect number of arguments. Expecting 1")
	//}
	//productName := args[0]
	//
	//// to maintain the state~name index, we need to read the product first and get its state
	//valAsbytes, err := stub.GetState(productName) //get the product from chaincode state
	//if err != nil {
	//	jsonResp = "{\"Error\":\"Failed to get state for " + productName + "\"}"
	//	return shim.Error(jsonResp)
	//} else if valAsbytes == nil {
	//	jsonResp = "{\"Error\":\"Product does not exist: " + productName + "\"}"
	//	return shim.Error(jsonResp)
	//}
	//
	//err = json.Unmarshal([]byte(valAsbytes), &productJSON)
	//if err != nil {
	//	jsonResp = "{\"Error\":\"Failed to decode JSON of: " + productName + "\"}"
	//	return shim.Error(jsonResp)
	//}
	//
	//err = stub.DelState(productName) //remove the product from chaincode state
	//if err != nil {
	//	return shim.Error("Failed to delete state:" + err.Error())
	//}
	//
	//// maintain the index
	//stateIndexKey, err := stub.CreateCompositeKey(stateIndexName,
	//	[]string{strconv.Itoa(productJSON.Value.State), productJSON.Key.Name})
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	//
	////  Delete index entry to state.
	//err = stub.DelState(stateIndexKey)
	//if err != nil {
	//	return shim.Error("Failed to delete state:" + err.Error())
	//}
	return shim.Success(nil)
}

// ===== Example: Parameterized rich query =================================================
// queryProductsByOwner queries for products based on a passed in owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *ProductChaincode) queryProductsByOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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

// ===== Example: Ad hoc rich query ========================================================
// queryProducts uses a query string to perform a query for products.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryProductsForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *ProductChaincode) queryProducts(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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

	entries := []Product{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		entry := Product{}

		if err := json.Unmarshal(response.Value, &entry.Value); err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(response.Key)
		if err != nil {
			return nil, err
		}

		entry.Key.Name = compositeKeyParts[0]

		entries = append(entries, entry)
	}

	result, err := json.Marshal(entries)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (t *ProductChaincode) getHistoryForProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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

	type productHistory struct {
		Value ProductValue `json:"value"`
		TxId string `json:"txId"`
		Timestamp string `json:"timestamp"`
		IsDelete bool `json:"isDelete"`
	}

	entries := []productHistory{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		entry := productHistory{}

		if err := json.Unmarshal(response.Value, &entry.Value); err != nil {
			return shim.Error(err.Error())
		}

		entry.TxId = response.TxId
		entry.Timestamp = time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String()
		entry.IsDelete = response.IsDelete

		entries = append(entries, entry)
	}

	result, err := json.Marshal(entries)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(result)
}

func (t *ProductChaincode) updateOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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

	if productToUpdate.Value.Owner != oldOwner {
		return shim.Error("The specified product doesn't belong to the specified owner.")
	}

	productToUpdate.Value.Owner = newOwner
	productToUpdate.Value.LastUpdated = lastUpdated

	productJSONasBytes, _ := json.Marshal(productToUpdate.Value)
	compositeKey, err := stub.CreateCompositeKey(productIndex, []string{productToUpdate.Key.Name})
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(compositeKey, productJSONasBytes)
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

func main() {
	err := shim.Start(new(ProductChaincode))
	if err != nil {
		logger.Error(err.Error())
	}
}
