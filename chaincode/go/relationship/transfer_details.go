package main

import (
	"errors"
	"fmt"
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

const (
	transferIndex = "TransferDetails"
)

const (
	basicArgumentsNumber = 3
	keyFieldsNumber = 3
)

const (
	statusInitiated = "Initiated"
	statusAccepted = "Accepted"
	statusRejected = "Rejected"
	statusCancelled = "Cancelled"
)

type TransferDetailsKey struct {
	ProductKey      string `json:"productKey"`
	RequestSender   string `json:"requestSender"`
	RequestReceiver string `json:"requestReceiver"`
}

type TransferDetailsValue struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type TransferDetails struct {
	Key   TransferDetailsKey   `json:"key"`
	Value TransferDetailsValue `json:"value"`
}

func (details *TransferDetails) FillFromArguments(args []string) error {
	if len(args) < basicArgumentsNumber {
		return errors.New(fmt.Sprintf("arguments array must contain at least %d items", basicArgumentsNumber))
	}

	if err := details.FillFromCompositeKeyParts(args[:keyFieldsNumber]); err != nil {
		return err
	}

	return nil
}

func (details *TransferDetails) FillFromCompositeKeyParts(compositeKeyParts []string) error {
	if len(compositeKeyParts) < keyFieldsNumber {
		return errors.New(fmt.Sprintf("composite key parts array must contain at least %d items", keyFieldsNumber))
	}

	details.Key.ProductKey = compositeKeyParts[0]
	details.Key.RequestSender = compositeKeyParts[1]
	details.Key.RequestReceiver = compositeKeyParts[2]

	return nil
}

func (details *TransferDetails) FillFromLedgerValue(ledgerValue []byte) error {
	if err := json.Unmarshal(ledgerValue, &details.Value); err != nil {
		return err
	} else {
		return nil
	}
}

func (details *TransferDetails) ToCompositeKey(stub shim.ChaincodeStubInterface) (string, error) {
	compositeKeyParts := []string {
		details.Key.ProductKey,
		details.Key.RequestSender,
		details.Key.RequestReceiver,
	}

	return stub.CreateCompositeKey(transferIndex, compositeKeyParts)
}

func (details *TransferDetails) ToLedgerValue() ([]byte, error) {
	return json.Marshal(details.Value)
}

func (details *TransferDetails) ExistsIn(stub shim.ChaincodeStubInterface) bool {
	compositeKey, err := details.ToCompositeKey(stub)
	if err != nil {
		return false
	}

	if data, err := stub.GetState(compositeKey); err != nil || data == nil {
		return false
	}

	return true
}

func (details *TransferDetails) LoadFrom(stub shim.ChaincodeStubInterface) error {
	compositeKey, err := details.ToCompositeKey(stub)
	if err != nil {
		return err
	}

	data, err := stub.GetState(compositeKey)
	if err != nil {
		return err
	}

	return details.FillFromLedgerValue(data)
}

func (details *TransferDetails) UpdateOrInsertIn(stub shim.ChaincodeStubInterface) error {
	compositeKey, err := details.ToCompositeKey(stub)
	if err != nil {
		return err
	}

	value, err := details.ToLedgerValue()
	if err != nil {
		return err
	}

	if err = stub.PutState(compositeKey, value); err != nil {
		return err
	}

	return nil
}

func (details *TransferDetails) EmitState(stub shim.ChaincodeStubInterface) error {
	type eventDetails struct {
		ProductKey string `json:"product_key"`
		OldOwner   string `json:"old_owner"`
		NewOwner   string `json:"new_owner"`
	}

	ed := eventDetails{
		ProductKey: details.Key.ProductKey,
		OldOwner: details.Key.RequestReceiver,
		NewOwner: details.Key.RequestSender,
	}

	bytes, err := json.Marshal(ed)
	if err != nil {
		return err
	}

	if err = stub.SetEvent(transferIndex + "." + details.Value.Status, bytes); err != nil {
		return err
	}

	return nil
}

