/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"strconv"
        "strings"

	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("example_cc0")

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### example_cc0 from Init ###########")

	_, args := stub.GetFunctionAndParameters()
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var err error

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	logger.Info("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	logger.Info("##############   Testing CID API as an admin user ################")

	id, err := cid.GetID(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	logger.Infof("id = %s\n", id)

	mspid, err := cid.GetMSPID(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	logger.Infof("mspid = %s\n", mspid)

	//err = cid.AssertAttributeValue(stub, "ibm", "New York")
	//if err != nil {
	//	return shim.Error(err.Error())
	//}else{
	// logger.Info("successfuly asserted attribute ibm\n")
	//}
	val, ok, err := cid.GetAttributeValue(stub, "myattrib")
	if err != nil {
		return shim.Error(err.Error())
	}
	if !ok {
		logger.Info("Value for 'myattrib' not found")
	} else {
		logger.Infof("Value for attribute 'myattrib' :%s", val)
	}

	cert, err := cid.GetX509Certificate(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	logger.Infof("cert = %+v\n", cert)

	logger.Info("##############  DONE Testing CID API as an admin user in INIT  ################")

	return shim.Success(nil)

}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### example_cc0 Invoke ###########")

	function, args := stub.GetFunctionAndParameters()

	if function == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	}

	if function == "query" {
		// queries an entity state
		return t.query(stub, args)
	}
	if function == "move" {
		// Deletes an entity from its state
		return t.move(stub, args)
	}

	logger.Errorf("Unknown action, check the first argument, must be one of 'delete', 'query', or 'move'. But got: %v", args[0])
	return shim.Error(fmt.Sprintf("Unknown action, check the first argument, must be one of 'delete', 'query', or 'move'. But got: %v", args[0]))
}

func (t *SimpleChaincode) move(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Printf("\n\n ====================== M O V E ========================")
	// must be an invoke
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 4, function followed by 2 names and 1 value")
	}

	A = args[0]
	B = args[1]

	fmt.Printf("\n\n  A --- > %s , B ---> %s", A, B)
	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	fmt.Printf("\n\n  X --- > %s ", X)
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	logger.Infof("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var A string // Entities
	var err error

	logger.Info("##############   Testing CID API as a registered user ################")

	id, err := cid.GetID(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	logger.Infof("id = %s\n", id)

	mspid, err := cid.GetMSPID(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	logger.Infof("mspid = %s\n", mspid)

  
        //if (strings.Compare(id, "Jim") == 0 || strings.Compare(id, "Barry") == 0 || strings.Compare(id, "Cathy") == 0) {
 
        logger.Info("ID: after comparing", id)
	err = cid.AssertAttributeValue(stub, "ibm", "raleigh")
	if err != nil {
		return shim.Error(err.Error())
	} else {
		logger.Info("successfuly asserted attribute ibm")
	}

	val, ok, err := cid.GetAttributeValue(stub, "test1attr")
	if err != nil {
		return shim.Error(err.Error())
	}
	if !ok {
		logger.Info("Value for 'test1attr' not found")
	} else {
		logger.Infof("Value for attribute 'test1attr' :%s", val)
	}
        if (strings.Compare(id, "Cathy")==0) {
	   _, ok, _ := cid.GetAttributeValue(stub, "test1attr")
	   if !ok {
		logger.Info("Value for 'test1attr' not found")
	   }
        }
        if (strings.Compare(id, "Jen")==0) {
	   _, ok, _ := cid.GetAttributeValue(stub, "test1attr")
	   if !ok {
		logger.Info("Value for 'test1attr' not found")
	   }
        }

	cert, err := cid.GetX509Certificate(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	logger.Infof("cert = %+v\n", cert)

	logger.Info("##############   DONE Testing CID API as a registered user ################")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	logger.Infof("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
}
