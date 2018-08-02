package main

import (
	"strconv"
	"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"errors"
	"fmt"
	"encoding/json"
)

const (
	productIndex = "product"
)

const (
	basicArgumentsNumber = 5
	keyFieldsNumber = 1
)

const (
	stateUnknown = iota
	stateRegistered
	stateActive
	stateDecisionMaking
	stateInactive
)

var productStateMachine = map[int][]int{
	stateUnknown: {},
	stateRegistered: {stateRegistered, stateActive},
	stateActive: {stateActive, stateDecisionMaking},
	stateDecisionMaking: {stateActive, stateDecisionMaking, stateInactive},
	stateInactive: {stateInactive},
}

func contains(m map[int][]int, key int) bool {
	_, ok := m[key]
	if !ok {
		return false
	}

	return true
}

func checkStateValidity(statesAutomaton map[int][]int, oldState, newState int) bool {
	possibleStates, ok := statesAutomaton[oldState]
	if ok {
		for _, state := range possibleStates {
			if state == newState {
				return true
			}
		}
	}
	
	return false
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

func (product *Product) FillFromArguments(args []string) error {
	//      0            1         2       3        4
	// productName, description, status, owner, timestamp
	if len(args) < basicArgumentsNumber {
		return errors.New(fmt.Sprintf("incorrect number of arguments: expected %d, got %d",
			basicArgumentsNumber, len(args)))
	}

	// ==== Input sanitation ====
	for k, v := range args {
		if k != 1 && len(v) == 0 {
			return errors.New(fmt.Sprintf("argument #%d must be a non-empty string", k + 1))
		}
	}

	if err := product.FillFromCompositeKeyParts(args[:keyFieldsNumber]); err != nil {
		return err
	}

	desc := args[keyFieldsNumber]
	state, err := strconv.Atoi(args[keyFieldsNumber + 1])
	if err != nil {
		return errors.New(fmt.Sprintf("product state is invalid: %s (must be int)", args[keyFieldsNumber + 1]))
	}
	owner := strings.ToLower(args[keyFieldsNumber + 2])
	lastUpdated, err := strconv.Atoi(args[keyFieldsNumber + 3])
	if err != nil {
		return errors.New(fmt.Sprintf("product last change time is invalid: %s (must be int)",
			args[keyFieldsNumber + 3]))
	}
	
	if !contains(productStateMachine, state) {
		return errors.New(fmt.Sprintf("product is invalid: %d (must be from 0 to 4)", state))
	}

	product.Value.Desc = desc
	product.Value.State = state
	product.Value.Owner = owner
	product.Value.LastUpdated = lastUpdated

	return nil
}

func (product *Product) FillFromCompositeKeyParts(compositeKeyParts []string) error {
	if len(compositeKeyParts) < keyFieldsNumber {
		return errors.New(fmt.Sprintf("composite key parts array must contain at least %d item(s)",
			keyFieldsNumber))
	}

	for k, v := range compositeKeyParts {
		if len(v) == 0 {
			return errors.New(fmt.Sprintf("key part #%d must be a non-empty string", k + 1))
		}
	}

	product.Key.Name = compositeKeyParts[0]

	return nil
}

func (product *Product) FillFromLedgerValue(ledgerValue []byte) error {
	if err := json.Unmarshal(ledgerValue, &product.Value); err != nil {
		return err
	} else {
		return nil
	}
}

func (product *Product) ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	compositeKeyParts := []string {
		product.Key.Name,
	}

	return stub.CreateCompositeKey(productIndex, compositeKeyParts)
}

func (product *Product) ToLedgerValue() ([]byte, error) {
	return json.Marshal(product.Value)
}

func (product *Product) ExistsIn(stub shim.ChaincodeStubInterface) bool {
	compositeKey, err := product.ToCompositeKey(stub)
	if err != nil {
		return false
	}

	if data, err := stub.GetState(compositeKey); err != nil || data == nil {
		return false
	}

	return true
}

func (product *Product) LoadFrom(stub shim.ChaincodeStubInterface) error {
	compositeKey, err := product.ToCompositeKey(stub)
	if err != nil {
		return err
	}

	data, err := stub.GetState(compositeKey)
	if err != nil {
		return err
	}

	return product.FillFromLedgerValue(data)
}

func (product *Product) UpdateOrInsertIn(stub shim.ChaincodeStubInterface) error {
	compositeKey, err := product.ToCompositeKey(stub)
	if err != nil {
		return err
	}

	value, err := product.ToLedgerValue()
	if err != nil {
		return err
	}

	if err = stub.PutState(compositeKey, value); err != nil {
		return err
	}

	return nil
}