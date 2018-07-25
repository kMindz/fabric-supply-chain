
package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/pem"
	"crypto/x509"
	"strings"
	"fmt"
	"encoding/json"
	"errors"
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
)

var logger = shim.NewLogger("OwnershipChaincode")

type TransferDetailsKey struct {
	ProductKey      string `json:"product_key"`
	RequestSender   string `json:"request_sender"`
	RequestReceiver string `json:"request_receiver"`
}

type TransferDetailsValue struct {
	Status  string `json:"status"`
	Message string `json:"message"`
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

// OwnershipChaincode example simple Chaincode implementation
type OwnershipChaincode struct {
}

func (t *OwnershipChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *OwnershipChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")

	function, args := stub.GetFunctionAndParameters()

	if function == "sendRequest" {
		return t.sendRequest(stub, args)
	} else if function == "transferAccepted" {
		return t.transferAccepted(stub, args)
	} else if function == "transferRejected" {
		return t.transferRejected(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	}

	return pb.Response{Status:403, Message:"Invalid invoke function name."}
}

func (t *OwnershipChaincode) sendRequest(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	const expectedArgumentsNumber = basicArgumentsNumber + 1

	if len(args) < expectedArgumentsNumber {
		return shim.Error(fmt.Sprintf("insufficient number of arguments: expected %d, got %d",
			expectedArgumentsNumber, len(args)))
	}

	// TODO: check product existence in common channel
	// TODO: check ownership
	// TODO: check if request sender and creator are the same

	request := TransferDetails{}
	if err := request.FillFromArguments(args); err != nil {
		return shim.Error(err.Error())
	}

	if request.ExistsIn(stub) {
		if err := request.LoadFrom(stub); err != nil {
			return shim.Error(err.Error())
		}

		if request.Value.Status == statusInitiated {
			return shim.Error("ownership transfer is already initiated")
		}
	}

	request.Value.Status = statusInitiated
	request.Value.Message = args[keyFieldsNumber]

	if err := request.UpdateOrInsertIn(stub); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *OwnershipChaincode) transferAccepted(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < basicArgumentsNumber {
		return shim.Error(fmt.Sprintf("insufficient number of arguments: expected %d, got %d",
			basicArgumentsNumber, len(args)))
	}

	// TODO: check product existence in common channel
	// TODO: check ownership
	// TODO: check if request receiver and creator are the same

	details := TransferDetails{}
	if err := details.FillFromArguments(args); err != nil {
		return shim.Error(err.Error())
	}

	if !details.ExistsIn(stub) {
		return shim.Error("ownership transfer wasn't initiated")
	}

	if err := details.LoadFrom(stub); err != nil {
		return shim.Error(err.Error())
	}

	if details.Value.Status != statusInitiated {
		return shim.Error("ownership transfer wasn't initiated")
	}

	details.Value.Status = statusAccepted

	if err := details.UpdateOrInsertIn(stub); err != nil {
		return shim.Error(err.Error())
	}

	if err := details.EmitState(stub); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *OwnershipChaincode) transferRejected(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < basicArgumentsNumber {
		return shim.Error(fmt.Sprintf("insufficient number of arguments: expected %d, got %d",
			basicArgumentsNumber, len(args)))
	}

	// TODO: check if details receiver and creator are the same

	details := TransferDetails{}
	if err := details.FillFromArguments(args); err != nil {
		return shim.Error(err.Error())
	}

	if !details.ExistsIn(stub) {
		return shim.Error("ownership transfer wasn't initiated")
	}

	if err := details.LoadFrom(stub); err != nil {
		return shim.Error(err.Error())
	}

	if details.Value.Status != statusInitiated {
		return shim.Error("ownership transfer wasn't initiated")
	}

	details.Value.Status = statusRejected

	if err := details.UpdateOrInsertIn(stub); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}
func (t *OwnershipChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return shim.Success(nil)
}

var getCreator = func (certificate []byte) (string, string) {
	data := certificate[strings.Index(string(certificate), "-----"): strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]
	commonName := cert.Subject.CommonName
	logger.Debug("commonName: " + commonName + ", organization: " + organization)

	organizationShort := strings.Split(organization, ".")[0]

	return commonName, organizationShort
}

func main() {
	err := shim.Start(new(OwnershipChaincode))
	if err != nil {
		logger.Error(err.Error())
	}
}
