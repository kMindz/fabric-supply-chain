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
)

var logger = shim.NewLogger("ProductChaincode")

const (
	stateIndexName = "state~name"
)

// ProductChaincode example simple Chaincode implementation
type ProductChaincode struct {
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
	var product Product
	if err := product.FillFromArguments(args); err != nil {
		return shim.Error(err.Error())
	}

	if product.ExistsIn(stub) {
		compositeKey, _ := product.ToCompositeKey(stub)
		return shim.Error(fmt.Sprintf("product with the key %s already exists", compositeKey))
	}

	// TODO: set owner from GetCreatorOrg
	product.Value.State = stateRegistered

	if err := product.UpdateOrInsertIn(stub); err != nil {
		return shim.Error(err.Error())
	}

	// TODO: think about index usability
	//  ==== Index the product to enable state-based range queries, e.g. return all Active products ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on stateIndexName~state~name.
	//  This will enable very efficient state range queries based on composite keys matching stateIndexName~state~*
	//stateIndexKey, err := stub.CreateCompositeKey(stateIndexName, []string{strconv.Itoa(product.Value.State), product.Key.Name})
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	////  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the product.
	////  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	//value := []byte{0x00}
	//stub.PutState(stateIndexKey, value)

	return shim.Success(nil)
}

// ============================================================
// updateProduct - update an existing product, store into chaincode state
// ============================================================
func (t *ProductChaincode) updateProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var product, productToUpdate Product
	if err := product.FillFromArguments(args); err != nil {
		return shim.Error(err.Error())
	}

	productToUpdate.Key = product.Key

	if !productToUpdate.ExistsIn(stub) {
		compositeKey, _ := productToUpdate.ToCompositeKey(stub)
		return shim.Error(fmt.Sprintf("product with the key %s doesn't exist", compositeKey))
	}

	if err := productToUpdate.LoadFrom(stub); err != nil {
		return shim.Error(err.Error())
	}

	if !checkStateValidity(productStateMachine, productToUpdate.Value.State, product.Value.State) {
		return shim.Error(fmt.Sprintf("product state cannot be updated from %d to %d",
			productToUpdate.Value.State, product.Value.State))
	}

	// TODO; check if creator == productToUpdate owner
	if productToUpdate.Value.Owner != product.Value.Owner {
		return shim.Error(fmt.Sprintf("ownership cannot be transferred via product updating (from %s to %s)",
			productToUpdate.Value.Owner, product.Value.Owner))
	}

	//oldState := productToUpdate.Value.State

	productToUpdate.Value.Desc = product.Value.Desc
	productToUpdate.Value.State = product.Value.State
	productToUpdate.Value.LastUpdated = product.Value.LastUpdated

	if err := productToUpdate.UpdateOrInsertIn(stub); err != nil {
		return shim.Error(err.Error())
	}

	//// maintain the index
	//if productToUpdate.Value.State != oldState {
	//	//delete old index
	//	stateIndexKey, err := stub.CreateCompositeKey(stateIndexName,
	//		[]string{strconv.Itoa(oldState), productToUpdate.Key.Name})
	//	if err != nil {
	//		return shim.Error(err.Error())
	//	}
	//
	//	//  Delete index entry to state.
	//	err = stub.DelState(stateIndexKey)
	//	if err != nil {
	//		return shim.Error("Failed to delete state:" + err.Error())
	//	}
	//	//create new index
	//	stateIndexKey, err = stub.CreateCompositeKey(stateIndexName,
	//		[]string{strconv.Itoa(productToUpdate.Value.State), productToUpdate.Key.Name})
	//	if err != nil {
	//		return shim.Error(err.Error())
	//	}
	//	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the product.
	//	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	//	value := []byte{0x00}
	//	stub.PutState(stateIndexKey, value)
	//}

	return shim.Success(nil)
}

// ===============================================
// readProduct - read a product from chaincode state
// ===============================================
func (t *ProductChaincode) readProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < keyFieldsNumber {
		return shim.Error(fmt.Sprintf("incorrect number of arguments: expected %d, got %d",
			keyFieldsNumber, len(args)))
	}

	var product Product
	if err := product.FillFromCompositeKeyParts(args); err != nil {
		return shim.Error(err.Error())
	}

	if !product.ExistsIn(stub) {
		compositeKey, _ := product.ToCompositeKey(stub)
		return shim.Error(fmt.Sprintf("product with the key %s doesn't exist", compositeKey))
	}

	if err := product.LoadFrom(stub); err != nil {
		return shim.Error(err.Error())
	}

	result, err := json.Marshal(product)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(result)
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
	it, err := stub.GetStateByPartialCompositeKey(productIndex, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	entries := []Product{}
	for it.HasNext() {
		response, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		entry := Product{}

		if err := entry.FillFromLedgerValue(response.Value); err != nil {
			return shim.Error(err.Error())
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(response.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		if err := entry.FillFromCompositeKeyParts(compositeKeyParts); err != nil {
			return shim.Error(err.Error())
		}

		entries = append(entries, entry)
	}

	result, err := json.Marshal(entries)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(result)
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

		if err := entry.FillFromCompositeKeyParts(compositeKeyParts); err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	result, err := json.Marshal(entries)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (t *ProductChaincode) getHistoryForProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var product Product
	if err := product.FillFromCompositeKeyParts(args); err != nil {
		return shim.Error(err.Error())
	}

	compositeKey, err := product.ToCompositeKey(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	resultsIterator, err := stub.GetHistoryForKey(compositeKey)
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
	//      0          1         2          3
	// productName, oldOwner, newOwner, timestamp
	const expectedArgumentsNumber = 4
	if len(args) < expectedArgumentsNumber {
		return shim.Error(fmt.Sprintf("incorrect number of arguments: expected %d, got %d",
			expectedArgumentsNumber, len(args)))
	}

	var product Product
	if err := product.FillFromCompositeKeyParts(args[:keyFieldsNumber]); err != nil {
		return shim.Error(err.Error())
	}

	// ==== Input sanitation ====
	for k, v := range args[1:] {
		if len(v) == 0 {
			return shim.Error(fmt.Sprintf("argument #%d must be a non-empty string", k + 1))
		}
	}

	oldOwner := args[keyFieldsNumber]
	newOwner := args[keyFieldsNumber + 1]
	lastUpdated, err := strconv.Atoi(args[keyFieldsNumber + 2])
	if err != nil {
		return shim.Error(fmt.Sprintf("product last change time is invalid: %s (must be int)",
			args[keyFieldsNumber + 2]))
	}

	// TODO: check if creator org and oldOwner are the same
	//if GetCreatorOrganization(stub) != oldOwner {
	//	return shim.Error(fmt.Sprintf("no privileges to send request from the side of %s", oldOwner))
	//}

	if !product.ExistsIn(stub) {
		compositeKey, _ := product.ToCompositeKey(stub)
		return shim.Error(fmt.Sprintf("product with the key %s doesn't exist", compositeKey))
	}

	if err := product.LoadFrom(stub); err != nil {
		return shim.Error(err.Error())
	}

	if product.Value.Owner != oldOwner {
		return shim.Error("the specified product doesn't belong to the specified owner")
	}

	product.Value.Owner = newOwner
	product.Value.LastUpdated = lastUpdated

	if err := product.UpdateOrInsertIn(stub); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func getOrganization(certificate []byte) string {
	data := certificate[strings.Index(string(certificate), "-----") : strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]
	return strings.Split(organization, ".")[0]
}

func GetCreatorOrganization(stub shim.ChaincodeStubInterface) string {
	certificate, _ := stub.GetCreator()
	return getOrganization(certificate)
}

func main() {
	err := shim.Start(new(ProductChaincode))
	if err != nil {
		logger.Error(err.Error())
	}
}
