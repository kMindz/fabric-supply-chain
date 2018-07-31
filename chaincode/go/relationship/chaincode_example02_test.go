package main

import (
	"testing"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"fmt"
)

func toByteArray(args []string) [][]byte {
	var res [][]byte
	for _, s := range args {
		res = append(res, []byte(s))
	}

	return res
}

func getInitializedStub(t *testing.T) *shim.MockStub{
	cc := new(OwnershipChaincode)
	stub := shim.NewMockStub("ownership", cc)
	stub.MockInit("1", toByteArray([]string{"init"}))
	return stub
}

func TestQuery(t *testing.T) {
	var response pb.Response
	stub := getInitializedStub(t)

	args := []string{"sendRequest", "product", "sender", "receiver", "message"}
	response = stub.MockInvoke("ownership", toByteArray(args))
	if response.Status < 400 {
		response = stub.MockInvoke("ownership", toByteArray([]string{"query"}))
		if response.Status < 400 {
			fmt.Print(string(response.Payload))
		} else {
			fmt.Print("Query error")
			t.FailNow()
		}
	} else {
		fmt.Print("Send request error")
		t.FailNow()
	}
}