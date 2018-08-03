
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

var logger = shim.NewLogger("OwnershipChaincode")

const (
	commonChannelName = "common"
	commonChaincodeName = "reference"
)

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
	} else if function == "history" {
		return t.history(stub, args)
	}

	return pb.Response{Status:403, Message:"Invalid invoke function name."}
}

func (t *OwnershipChaincode) sendRequest(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	const expectedArgumentsNumber = basicArgumentsNumber + 1

	if len(args) < expectedArgumentsNumber {
		return shim.Error(fmt.Sprintf("insufficient number of arguments: expected %d, got %d",
			expectedArgumentsNumber, len(args)))
	}


	request := TransferDetails{}
	if err := request.FillFromArguments(args); err != nil {
		return shim.Error(err.Error())
	}

	if GetCreatorOrganization(stub) != request.Key.RequestSender {
		return shim.Error(fmt.Sprintf(
			"no privileges to send request from the side of organization %s (caller is from organization %s)",
			request.Key.RequestSender, GetCreatorOrganization(stub)))
	}

	if request.ExistsIn(stub) {
		if err := request.LoadFrom(stub); err != nil {
			return shim.Error(err.Error())
		}

		if request.Value.Status == statusInitiated {
			return shim.Error("ownership transfer is already initiated")
		}
	}

	if err := checkProductExistenceAndOwnership(stub, request.Key.ProductKey, request.Key.RequestReceiver); err != nil {
		return shim.Error(err.Error())
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

	details := TransferDetails{}
	if err := details.FillFromArguments(args); err != nil {
		return shim.Error(err.Error())
	}

	if GetCreatorOrganization(stub) != details.Key.RequestReceiver {
		return shim.Error(fmt.Sprintf(
			"no privileges to accept transfer from the side of organization %s (caller is from organization %s)",
			details.Key.RequestReceiver, GetCreatorOrganization(stub)))
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

	if err := checkProductExistenceAndOwnership(stub, details.Key.ProductKey, details.Key.RequestReceiver); err != nil {
		// TODO: think about request deletion
		return shim.Error(err.Error())
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

	details := TransferDetails{}
	if err := details.FillFromArguments(args); err != nil {
		return shim.Error(err.Error())
	}

	if GetCreatorOrganization(stub) != details.Key.RequestReceiver {
		return shim.Error(fmt.Sprintf(
			"no privileges to reject transfer from the side of organization %s (caller is from organization %s)",
			details.Key.RequestReceiver, GetCreatorOrganization(stub)))
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
	it, err := stub.GetStateByPartialCompositeKey(transferIndex, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	entries := []TransferDetails{}
	for it.HasNext() {
		response, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		entry := TransferDetails{}

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

func (t *OwnershipChaincode) history(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	const expectedArgumentsNumber = 1

	if len(args) < expectedArgumentsNumber {
		return shim.Error(fmt.Sprintf("insufficient number of arguments: expected %d, got %d",
			expectedArgumentsNumber, len(args)))
	}

	queryIterator, err := stub.GetStateByPartialCompositeKey(transferIndex, []string{args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer queryIterator.Close()

	entries := []TransferDetails{}
	for queryIterator.HasNext() {
		queryResponse, err := queryIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		historyIterator, err := stub.GetHistoryForKey(queryResponse.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		for historyIterator.HasNext() {
			historyResponse, err := historyIterator.Next()
			if err != nil {
				return shim.Error(err.Error())
			}

			entry := TransferDetails{}

			if err := entry.FillFromLedgerValue(historyResponse.Value); err != nil {
				return shim.Error(err.Error())
			}

			_, compositeKeyParts, err := stub.SplitCompositeKey(queryResponse.Key)
			if err != nil {
				return shim.Error(err.Error())
			}

			if err := entry.FillFromCompositeKeyParts(compositeKeyParts); err != nil {
				return shim.Error(err.Error())
			}
			
			entries = append(entries, entry)
		}
		historyIterator.Close()
	}

	result, err := json.Marshal(entries)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(result)
}

func checkProductExistenceAndOwnership(stub shim.ChaincodeStubInterface, productKey, requiredOwner string) error {
	type simplifiedProduct struct {
		Value struct {
			Owner string `json:"owner"`
		} `json:"value"`
	}

	const queryFunctionName = "readProduct"

	response := stub.InvokeChaincode(commonChaincodeName,
		[][]byte{[]byte(queryFunctionName), []byte(productKey)}, commonChannelName)
	if response.Status >= 400 {
		return errors.New(
			fmt.Sprintf("unable to read product %s from common channel: %s", productKey, response.Message))
	} else {
		var p simplifiedProduct
		if err := json.Unmarshal(response.Payload, &p); err != nil {
			return errors.New(
				fmt.Sprintf("unable to unmarshal response on product %s from common channel", productKey))
		}

		if p.Value.Owner != requiredOwner {
			return errors.New(
				fmt.Sprintf("product %s don't belong to organization %s", productKey, requiredOwner))
		}
	}

	return nil
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
	err := shim.Start(new(OwnershipChaincode))
	if err != nil {
		logger.Error(err.Error())
	}
}
