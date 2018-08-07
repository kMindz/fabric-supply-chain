
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
	logger.Info("OwnershipChaincode.Invoke is running")
	logger.Debug("OwnershipChaincode.Invoke")

	function, args := stub.GetFunctionAndParameters()
	logger.Debug("Function: " + function + ", arguments: " + strings.Join(args, ","))

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

	message := "invalid invoke function name. " +
		"Expected one of {sendRequest, transferAccepted, transferRejected, query, history}, but got " + function

	logger.Error(message)
	return pb.Response{Status:400, Message: message}
}

func (t *OwnershipChaincode) sendRequest(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("OwnershipChaincode.sendRequest is running")
	logger.Debug("OwnershipChaincode.sendRequest")

	const expectedArgumentsNumber = basicArgumentsNumber + 1

	if len(args) < expectedArgumentsNumber {
		message := fmt.Sprintf("insufficient number of arguments: expected %d, got %d",
			expectedArgumentsNumber, len(args))
		logger.Error(message)
		return shim.Error(message)
	}

	request := TransferDetails{}
	if err := request.FillFromArguments(args); err != nil {
		message := fmt.Sprintf("cannot read transfer details from arguments: %s", err.Error())
		logger.Error(message)
		return shim.Error(message)
	}

	if bytes, err := json.Marshal(request); err == nil {
		logger.Debug("Request: " + string(bytes))
	}

	if GetCreatorOrganization(stub) != request.Key.RequestSender {
		message := fmt.Sprintf(
			"no privileges to send request from the side of organization %s (caller is from organization %s)",
			request.Key.RequestSender, GetCreatorOrganization(stub))
		logger.Error(message)
		return pb.Response{Status: 403, Message: message}
	}

	logger.Debug("RequestSender: " + request.Key.RequestSender)

	if request.ExistsIn(stub) {
		if err := request.LoadFrom(stub); err != nil {
			message := fmt.Sprintf("cannot load existing request: %s", err.Error())
			logger.Error(message)
			return pb.Response{Status: 404, Message: message}
		}

		if request.Value.Status == statusInitiated {
			message := "ownership transfer is already initiated"
			logger.Error(message)
			return shim.Error(message)
		}
	}

	if err := checkProductExistenceAndOwnership(stub, request.Key.ProductKey, request.Key.RequestReceiver); err != nil {
		message := err.Error()
		logger.Error(message)
		return shim.Error(message)
	}

	request.Value.Status = statusInitiated
	request.Value.Message = args[keyFieldsNumber]

	if err := request.UpdateOrInsertIn(stub); err != nil {
		message := fmt.Sprintf("persistence error: %s", err.Error())
		logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	logger.Info("OwnershipChaincode.sendRequest exited without errors")
	logger.Debug("Success: OwnershipChaincode.sendRequest")
	return shim.Success(nil)
}

func (t *OwnershipChaincode) transferAccepted(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("OwnershipChaincode.transferAccepted is running")
	logger.Debug("OwnershipChaincode.transferAccepted")

	if len(args) < basicArgumentsNumber {
		message := fmt.Sprintf("insufficient number of arguments: expected %d, got %d",
			basicArgumentsNumber, len(args))
		logger.Error(message)
		return shim.Error(message)
	}

	details := TransferDetails{}
	if err := details.FillFromArguments(args); err != nil {
		message := fmt.Sprintf("cannot read transfer details from arguments: %s", err.Error())
		logger.Error(message)
		return shim.Error(message)
	}

	if bytes, err := json.Marshal(details); err == nil {
		logger.Debug("Details: " + string(bytes))
	}

	if GetCreatorOrganization(stub) != details.Key.RequestReceiver {
		message := fmt.Sprintf(
			"no privileges to accept transfer from the side of organization %s (caller is from organization %s)",
			details.Key.RequestReceiver, GetCreatorOrganization(stub))
		logger.Error(message)
		return pb.Response{Status: 403, Message: message}
	}

	if !details.ExistsIn(stub) {
		message := "ownership transfer wasn't initiated"
		logger.Error(message)
		return shim.Error(message)
	}

	if err := details.LoadFrom(stub); err != nil {
		message := fmt.Sprintf("cannot load existing transfer details: %s", err.Error())
		logger.Error(message)
		return pb.Response{Status: 404, Message: message}
	}

	if details.Value.Status != statusInitiated {
		message := "ownership transfer wasn't initiated"
		logger.Error(message)
		return shim.Error(message)
	}

	if err := checkProductExistenceAndOwnership(stub, details.Key.ProductKey, details.Key.RequestReceiver); err != nil {
		// TODO: think about request deletion
		message := err.Error()
		logger.Error(message)
		return shim.Error(message)
	}

	details.Value.Status = statusAccepted

	if err := details.UpdateOrInsertIn(stub); err != nil {
		message := fmt.Sprintf("persistence error: %s", err.Error())
		logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	if err := details.EmitState(stub); err != nil {
		message := fmt.Sprintf("unable to emit outgoing event: %s", err.Error())
		logger.Error(message)
		return shim.Error(err.Error())
	}

	logger.Info("OwnershipChaincode.transferAccepted exited without errors")
	logger.Debug("Success: OwnershipChaincode.transferAccepted")
	return shim.Success(nil)
}

func (t *OwnershipChaincode) transferRejected(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("OwnershipChaincode.transferRejected is running")
	logger.Debug("OwnershipChaincode.transferRejected")

	if len(args) < basicArgumentsNumber {
		message := fmt.Sprintf("insufficient number of arguments: expected %d, got %d",
			basicArgumentsNumber, len(args))
		logger.Error(message)
		return shim.Error(message)
	}

	details := TransferDetails{}
	if err := details.FillFromArguments(args); err != nil {
		message := fmt.Sprintf("cannot read transfer details from arguments: %s", err.Error())
		logger.Error(message)
		return shim.Error(message)
	}

	if bytes, err := json.Marshal(details); err == nil {
		logger.Debug("Details: " + string(bytes))
	}

	creatorIsReceiver := GetCreatorOrganization(stub) == details.Key.RequestReceiver
	creatorIsSender := GetCreatorOrganization(stub) == details.Key.RequestSender

	if !creatorIsReceiver && !creatorIsSender {
		message := fmt.Sprintf(
			"no privileges to reject transfer from the side of organization %s", GetCreatorOrganization(stub))
		logger.Error(message)
		return pb.Response{Status: 403, Message: message}
	}

	if !details.ExistsIn(stub) {
		message := "ownership transfer wasn't initiated"
		logger.Error(message)
		return shim.Error(message)
	}

	if err := details.LoadFrom(stub); err != nil {
		message := fmt.Sprintf("cannot load existing transfer details: %s", err.Error())
		logger.Error(message)
		return pb.Response{Status: 404, Message: message}
	}

	if details.Value.Status != statusInitiated {
		message := "ownership transfer wasn't initiated"
		logger.Error(message)
		return shim.Error(message)
	}

	if creatorIsReceiver {
		logger.Debug("Rejected by receiver")
		details.Value.Status = statusRejected
	} else if creatorIsSender {
		logger.Debug("Rejected by sender")
		details.Value.Status = statusCancelled
	}

	if err := details.UpdateOrInsertIn(stub); err != nil {
		message := fmt.Sprintf("persistence error: %s", err.Error())
		logger.Error(message)
		return pb.Response{Status: 500, Message: message}
	}

	logger.Info("OwnershipChaincode.transferRejected exited without errors")
	logger.Debug("Success: OwnershipChaincode.transferRejected")
	return shim.Success(nil)
}

func (t *OwnershipChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("OwnershipChaincode.query is running")
	logger.Debug("OwnershipChaincode.query")

	it, err := stub.GetStateByPartialCompositeKey(transferIndex, []string{})
	if err != nil {
		message := fmt.Sprintf("unable to get state by partial composite key %s: %s", transferIndex, err.Error())
		logger.Error(message)
		return shim.Error(message)
	}
	defer it.Close()

	entries := []TransferDetails{}
	for it.HasNext() {
		response, err := it.Next()
		if err != nil {
			message := fmt.Sprintf("unable to get an element next to a query iterator: %s", err.Error())
			logger.Error(message)
			return shim.Error(message)
		}

		logger.Debug(fmt.Sprintf("Response: {%s, %s}", response.Key, string(response.Value)))

		entry := TransferDetails{}

		if err := entry.FillFromLedgerValue(response.Value); err != nil {
			message := fmt.Sprintf("cannot fill transfer details value from response value: %s", err.Error())
			logger.Error(message)
			return shim.Error(message)
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(response.Key)
		if err != nil {
			message := fmt.Sprintf("cannot split response key into composite key parts slice: %s", err.Error())
			logger.Error(message)
			return shim.Error(message)
		}

		if err := entry.FillFromCompositeKeyParts(compositeKeyParts); err != nil {
			message := fmt.Sprintf("cannot fill transfer details key from composite key parts: %s", err.Error())
			logger.Error(message)
			return shim.Error(message)
		}

		if bytes, err := json.Marshal(entry); err == nil {
			logger.Debug("Entry: " + string(bytes))
		}

		entries = append(entries, entry)
	}

	result, err := json.Marshal(entries)
	if err != nil {
		return shim.Error(err.Error())
	}
	logger.Debug("Result: " + string(result))

	logger.Info("OwnershipChaincode.query exited without errors")
	logger.Debug("Success: OwnershipChaincode.query")
	return shim.Success(result)
}

func (t *OwnershipChaincode) history(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("OwnershipChaincode.history is running")
	logger.Debug("OwnershipChaincode.history")

	const expectedArgumentsNumber = 1

	if len(args) < expectedArgumentsNumber {
		message := fmt.Sprintf("insufficient number of arguments: expected %d, got %d",
			expectedArgumentsNumber, len(args))
		logger.Error(message)
		return shim.Error(message)
	}

	queryIterator, err := stub.GetStateByPartialCompositeKey(transferIndex, []string{args[0]})
	if err != nil {
		var message string
		if compositeKey, e := stub.CreateCompositeKey(transferIndex, args[:1]); e == nil {
			message = fmt.Sprintf("unable to get state by partial composite key %s: %s",
				compositeKey, err.Error())
		} else {
			message = fmt.Sprintf("unable to get state by partial composite key: %s", err.Error())
		}
		logger.Error(message)
		return shim.Error(message)
	}
	defer queryIterator.Close()

	entries := []TransferDetails{}
	for queryIterator.HasNext() {
		queryResponse, err := queryIterator.Next()
		if err != nil {
			message := fmt.Sprintf("unable to get an element next to a query iterator: %s", err.Error())
			logger.Error(message)
			return shim.Error(message)
		}

		logger.Debug("Query response key: " + queryResponse.Key)

		historyIterator, err := stub.GetHistoryForKey(queryResponse.Key)
		if err != nil {
			message := fmt.Sprintf("unable to get history for key %s: %s", queryResponse.Key, err.Error())
			logger.Error(message)
			return shim.Error(message)
		}

		for historyIterator.HasNext() {
			historyResponse, err := historyIterator.Next()
			if err != nil {
				message := fmt.Sprintf("unable to get an element next to a history iterator: %s", err.Error())
				logger.Error(message)
				return shim.Error(message)
			}

			entry := TransferDetails{}

			if err := entry.FillFromLedgerValue(historyResponse.Value); err != nil {
				message := fmt.Sprintf("cannot fill transfer details value from response value: %s", err.Error())
				logger.Error(message)
				return shim.Error(message)
			}

			_, compositeKeyParts, err := stub.SplitCompositeKey(queryResponse.Key)
			if err != nil {
				message := fmt.Sprintf("cannot split response key into composite key parts slice: %s",
					err.Error())
				logger.Error(message)
				return shim.Error(message)
			}

			if err := entry.FillFromCompositeKeyParts(compositeKeyParts); err != nil {
				message := fmt.Sprintf("cannot fill transfer details key from composite key parts: %s",
					err.Error())
				logger.Error(message)
				return shim.Error(message)
			}

			if bytes, err := json.Marshal(entry); err == nil {
				logger.Debug("Entry: " + string(bytes))
			}

			entries = append(entries, entry)
		}
		historyIterator.Close()
	}

	result, err := json.Marshal(entries)
	if err != nil {
		return shim.Error(err.Error())
	}
	logger.Debug("Result: " + string(result))

	logger.Info("OwnershipChaincode.history exited without errors")
	logger.Debug("Success: OwnershipChaincode.history")
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
