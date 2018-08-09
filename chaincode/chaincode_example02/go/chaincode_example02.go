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

//WARNING - this chaincode's ID is hard-coded in chaincode_example04 to illustrate one way of
//calling chaincode from a chaincode. If this example is modified, chaincode_example04.go has
//to be modified as well with the new ID of chaincode_example02.
//chaincode_example05 show's how chaincode ID can be passed in as a parameter instead of
//hard-coding.

import (
	"fmt"
	"strconv"
	"math/big"
	"crypto/rand"
	"encoding/json"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type ECPointStr struct {
	X, Y string
}

type pubkey struct {
	Name       string 					`json:"name"`    //the fieldtags are needed to keep case from bouncing around
	PubKey     ECPointStr   		`json:"publickey"`
}

type pubkeys struct {
	ObjectType string 					`json:"docType"` //docType is used to distinguish the various types of objects in state database
	Object     []pubkey          `json:"keys"`
}

type randomNum struct {
	Name       string 					`json:"name"`    //the fieldtags are needed to keep case from bouncing around
	RandNum    string   				`json:"randomnumber"`
}

type randomNums struct {
	ObjectType string 					`json:"docType"` //docType is used to distinguish the various types of objects in state database
	Object     []randomNum      `json:"rnumbers"`
	Ralpha     string						`json:"rAlpha"`
	Rrho       string						`json:"rRho"`
	ObjectV1   []string					`json:"rVector1"`
	ObjectV2   []string      		`json:"rVector2"`
	Rtau1			 string						`json:"rTau1"`
	Rtau2      string						`json:"rTau2"`
}

type InnerProdArgStr struct {
	L []ECPointStr
	R []ECPointStr
	A string
	B string
	Challenges []string
}

type RangeProofStr struct {
	Comm ECPointStr
	A    ECPointStr
	S    ECPointStr
	T1   ECPointStr
	T2   ECPointStr
	Tau  string
	Th   string
	Mu   string
	IPP  InnerProdArgStr
	// challenges
	Cy string
	Cz string
	Cx string
}

type ZKElementStr struct {
	Commitment, Token ECPointStr
	RP                RangeProofStr
	CommitPAll        ECPointStr
}


type zkelement struct {
	Name       string 					`json:"name"`    //the fieldtags are needed to keep case from bouncing around
	ZKElement  ZKElementStr 		`json:"zkelement"`
}

type zkrow struct {
	ObjectType string 					`json:"docType"` //docType is used to distinguish the various types of objects in state database
	Object     []zkelement      `json:"elements"`
}

func convertBIntStrArry(pStr []string) []*big.Int {
	PArr := make([]*big.Int, len(pStr))
	for i := 0; i < len(pStr); i++ {
		PCX := new(big.Int)
		PCX, _ = PCX.SetString(pStr[i], 10) //10 is base
		PArr[i] = PCX
	}
	return PArr
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("========chaincode_example Init=========")

	_, args := stub.GetFunctionAndParameters()
	var A, B, C, D string    // Entities
	var Aval, Bval, Cval, Dval int // Asset holdings
	var err error

	if len(args) != 8 {
		return shim.Error("Incorrect number of arguments. Expecting 8")
	}

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
	C = args[4]
	Cval, err = strconv.Atoi(args[5])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	D = args[6]
	Dval, err = strconv.Atoi(args[7])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	bankname := []string{A, B, C, D,}
	//fmt.Printf("Aval = %d, Bval = %d, Cval = %d, Dval = %d\n", Aval, Bval, Cval, Dval)
	fmt.Println("0)\tBefore tx, each bank's asset: Aval = ", Aval, ", Bval = ", Bval, ", Cval = ", Cval, ", Dval = ", Dval)

	fmt.Println("1)\tInitialize Elliptic Curve parameters...")
	dimension := 64//just need to guarantee that they all use 64
	EC := pb.NewECPrimeGroupKey(dimension)
	fmt.Println("2)\tAssigning private keys...")
	banknum := 4
	SK := make([]*big.Int, banknum)//row
	for j := range SK {
		r, err := rand.Int(rand.Reader, EC.N)
		if err != nil {
			return shim.Error(err.Error())
		}
		SK[j] = r
	}
	fmt.Println("SK: ", SK)
	fmt.Println("3)\tComputing public keys...")
	PK := make([]pb.ECPoint, banknum)
	for j := range PK {
		pkey := pb.PublicKeyPC(SK[j])
		PK[j] = pkey
	}
	fmt.Println("PK: ", PK)
	fmt.Println("4)\tWriting Public keys into the ledger...")

	//pubkeys := &pubkeys{}
	var pubkeys pubkeys
  pubkeys.ObjectType = "PubKey"
	pubkeys.Object = make([]pubkey, 0)
	// ==== Create EC object and marshal to JSON ====
	for j := range PK {
		//objectType := "PubKey"
		ecPointStr := ECPointStr{PK[j].X.String(), PK[j].Y.String()}
		pubkey := pubkey{bankname[j], ecPointStr}//save ECPoint as ECPointStr
		pubkeys.Object = append(pubkeys.Object, pubkey)
	}
	pkJSONasBytes, err := json.Marshal(pubkeys)
	if err != nil {
		return shim.Error(err.Error())
	}
	// === Save pubkey to state ===
	//fmt.Println("pkJSONasBytes = ", pkJSONasBytes)
	err = stub.PutState("PubKey", pkJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("5)\tGenerating random numbers...")
	r := pb.GetR(banknum) //generate r for each tx
	var randomNums, randomNumsNext randomNums
	randomNums.ObjectType = "RandNum"
	randomNums.Object = make([]randomNum, 0)
	for j := range r {
		rbigint := *r[j]
		randomNum := randomNum{bankname[j], rbigint.String()}//save *big.Int as string
		randomNums.Object = append(randomNums.Object, randomNum)
	}
	rJSONasBytes, err := json.Marshal(randomNums)
	if err != nil {
		return shim.Error(err.Error())
	}
	// === Save random number to state ===
	err = stub.PutState("TX0RandNum", rJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("6)\tGenerating random numbers for next 5 TXs...")
	for i := 1; i < 6; i++ {
		rNext := pb.GetR(banknum) //generate r for each tx
		//var randomNumsNext randomNums
		randomNumsNext.ObjectType = "RandNum"
		randomNumsNext.Object = make([]randomNum, 0)
		for j := range rNext {
			rbigint := *rNext[j]
			randomNum := randomNum{bankname[j], rbigint.String()}//save *big.Int as string
			randomNumsNext.Object = append(randomNumsNext.Object, randomNum)
		}
		//generating alpha and rho for RangeProof
		alpha, err := rand.Int(rand.Reader, EC.N)
		if err != nil {
			return shim.Error(err.Error())
		}
		randomNumsNext.Ralpha = alpha.String()
		//randomNumAlpha := randomNum{"RPalpha", alpha.String()}
		//randomNumsNext.Object = append(randomNumsNext.Object, randomNumAlpha)
		rho, err := rand.Int(rand.Reader, EC.N)
		if err != nil {
			return shim.Error(err.Error())
		}
		randomNumsNext.Rrho = rho.String()
		// randomNumRho := randomNum{"RPrho", rho.String()}
		// randomNumsNext.Object = append(randomNumsNext.Object, randomNumRho)

		tau1, err := rand.Int(rand.Reader, EC.N)
		if err != nil {
			return shim.Error(err.Error())
		}
		randomNumsNext.Rtau1 = tau1.String()
		tau2, err := rand.Int(rand.Reader, EC.N)
		if err != nil {
			return shim.Error(err.Error())
		}
		randomNumsNext.Rtau2 = tau2.String()
		//////////////////////
		//generate two random number Vectors
		rVec1 := pb.RandVector(EC.V)
		randomNumsNext.ObjectV1 = make([]string, EC.V)
		for j := range rVec1 {
			randomNumsNext.ObjectV1[j] = rVec1[i].String()
		}
		rVec2 := pb.RandVector(EC.V)
		randomNumsNext.ObjectV2 = make([]string, EC.V)
		for j := range rVec2 {
			randomNumsNext.ObjectV2[j] = rVec2[i].String()
		}
		/////////////////////
		rJSONasBytesNext, err := json.Marshal(randomNumsNext)
		if err != nil {
			return shim.Error(err.Error())
		}
		// === Save random number to state ===
		randName := "TX"+strconv.Itoa(i)+"RandNum"
		err = stub.PutState(randName, rJSONasBytesNext)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	fmt.Println("6)\tComputing Pedersen commiments and Token...")
	value := []*big.Int{big.NewInt(int64(Aval)),big.NewInt(int64(Bval)),big.NewInt(int64(Cval)),big.NewInt(int64(Dval)),}
	var zkrow zkrow
  zkrow.ObjectType = "ZKrow"
	zkrow.Object = make([]zkelement, 0)
	for i := 0; i < banknum; i++ {
		Pcommitment, Token := pb.PCommitToken(value[i],r[i],PK[i])
		fmt.Println("Pcommit, Token are ", Pcommitment, Token)
		//rpresult := RangeProofStr{}
		//ZKElement := pb.ZKElement{Pcommitment, Token, rpresult}
		ZKElement := ZKElementStr{
			ECPointStr{Pcommitment.X.String(), Pcommitment.Y.String()},
			ECPointStr{Token.X.String(), Token.Y.String()},
			RangeProofStr{},
			ECPointStr{},
		}
		zkelement := zkelement{bankname[i], ZKElement}
		zkrow.Object = append(zkrow.Object, zkelement)
	}

	// Write the state to the ledger
	fmt.Println("7)\tWriting Pedersen commiments and Token into the ledger...")
	zkJSONasBytes, err := json.Marshal(zkrow)
	if err != nil {
		return shim.Error(err.Error())
	}
	// === Save pubkey to state ===
	err = stub.PutState("TX0", zkJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//fmt.Println("Writing TX ID into the ledger...")
	err = stub.PutState("TXID", []byte(strconv.Itoa(0)))
	if err != nil {
		return shim.Error(err.Error())
	}

	// err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }
	//
	// err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }
	//
	// err = stub.PutState(C, []byte(strconv.Itoa(Cval)))
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }
	//
	// err = stub.PutState(D, []byte(strconv.Itoa(Dval)))
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }

	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("========chaincode_example Invoke=========")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// Make payment of X units from A to B
		return t.invoke(stub, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemtned in invoke
		return t.query(stub, args)
	} else if function == "generater" {
		return t.generateR(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}

func (t *SimpleChaincode) generateR(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//fmt.Println("Reading TX ID from the ledger...")
	banknum := 4
	var A, B, C, D string    // Entities
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}
	// Initialize the chaincode
	A = args[0]
	B = args[1]
	C = args[2]
	D = args[3]
	bankname := []string{A, B, C, D,}

	TXbytes, err := stub.GetState("TXID")
	if err != nil || TXbytes == nil {
		return shim.Error("Failed to get TX ID.")
	}
	TXval, _ := strconv.Atoi(string(TXbytes))

	fmt.Println("1)\tGenerating random numbers for the next TX...")
	//r := make([]*big.Int, banknum)
	// ==== Check if random numbers already exists ====
	randName := "TX"+strconv.Itoa(TXval+1)+"RandNum"
	fmt.Println("randName = ", randName)
	randNameAsBytes, err := stub.GetState(randName)
	if err != nil {
		return shim.Error("Failed to get randomNum: " + err.Error())
	} else if randNameAsBytes != nil {
		return shim.Error("The randoms are already assigned.")
	}
	rNext := pb.GetR(banknum) //generate r for each tx
	var randomNumsNext randomNums
	randomNumsNext.ObjectType = "RandNum"
	randomNumsNext.Object = make([]randomNum, 0)
	for j := range rNext {
		rbigint := *rNext[j]
		randomNum := randomNum{bankname[j], rbigint.String()}//save *big.Int as string
		randomNumsNext.Object = append(randomNumsNext.Object, randomNum)
	}
	rJSONasBytesNext, err := json.Marshal(randomNumsNext)
	if err != nil {
		return shim.Error(err.Error())
	}
	// === Save random number to state ===
	err = stub.PutState(randName, rJSONasBytesNext)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	beginTime := time.Now()
	var A, B string    // Entities
	var Aval, Bval, Avaleft int          // Transaction value
	var err error
	banknum := 4
	dimension := 64
	var Aindex int
	//indexArry := make(int, banknum)

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	A = args[0]
	B = args[3]
	// Perform the execution
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Avaleft, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	if(Avaleft < 0){
		Avaleft = 0 //The spending bank want to generate the RangeProof for Avaleft only when Avaleft is in range [0, EC.N)
	              //When the spending bank is malicous, it randomly choose a non-negative value, e.g., 0, and provide the RangeProof(0)
	}
	Bval, err = strconv.Atoi(args[4])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}


	//fmt.Println("1)\tInitialize Elliptic Curve parameters...")
	//dimension := 64//just need to guarantee that they all use 64
	//EC := pb.NewECPrimeGroupKey(dimension)

	fmt.Println("1)\tReading Public keys from the ledger...")
	pkJSONasBytes, err := stub.GetState("PubKey")
	if err != nil {
		return shim.Error("Failed to get marble:" + err.Error())
	} else if pkJSONasBytes == nil {
		return shim.Error("PK does not exist!")
	}

	var pubkeys pubkeys
	// marbleToTransfer := marble{}
	//fmt.Println("pkJSONasBytes = ", pkJSONasBytes)
	err = json.Unmarshal(pkJSONasBytes, &pubkeys) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("PK: ", pubkeys.Object)

	//PK := make([]pb.ECPoint, banknum)
	PK := make([]pb.ECPoint, banknum)
	bankname := make([]string, banknum)
	value := make([]*big.Int, banknum)
	// valueLeft := make([]*big.Int, banknum)//it does not store the real remaining assets, only the spending entry has the real value, others are all 0

	for i := 0; i < len(pubkeys.Object); i++ {
		//convert the ECPointStr to ECPoint
		PKX := new(big.Int)
    PKX, _ = PKX.SetString(pubkeys.Object[i].PubKey.X, 10)//10 is base
		PKY := new(big.Int)
    PKY, _ = PKY.SetString(pubkeys.Object[i].PubKey.Y, 10)//10 is base
		PK[i] = pb.ECPoint{PKX, PKY}

		bankname[i] = pubkeys.Object[i].Name
		if(pubkeys.Object[i].Name == A) {
			Aindex = i //spending bank's index in the ledger
			value[i] = big.NewInt(int64(Aval))//value and
			// valueLeft[i] = big.NewInt(int64(Avaleft))//only the spending entry has the real value
		} else if (pubkeys.Object[i].Name == B) {
			value[i] = big.NewInt(int64(Bval))
			// valueLeft[i] = big.NewInt(int64(0))
		} else {
			value[i] = big.NewInt(0)
			// valueLeft[i] = big.NewInt(int64(0))
		}
	}

	//fmt.Println("Reading TX ID from the ledger...")
	TXbytes, err := stub.GetState("TXID")
	if err != nil || TXbytes == nil {
		return shim.Error("Failed to get state")
	}
	TXval, _ := strconv.Atoi(string(TXbytes))
	TXval++

	fmt.Println("2)\tReading assigned random numbers...")
	r := make([]*big.Int, banknum)
	alpha := new(big.Int)
	rho := new(big.Int)
	tau1 := new(big.Int)
	tau2 := new(big.Int)
	sL := make([]*big.Int, dimension)
	sR := make([]*big.Int, dimension)
	// ==== Check if random numbers already exists ====
	randName := "TX"+strconv.Itoa(TXval)+"RandNum"
	//fmt.Println("randName = ", randName)
	randNameAsBytes, err := stub.GetState(randName)
	// //retry 10 times unitl gets the random numbers
	// for j := 0; j < 10; j++ {
	// 	if randNameAsBytes != nil{
	// 		break
	// 	}
	// 	time.Sleep(2 * time.Second)
	// 	randNameAsBytes, err = stub.GetState(randName)
	// }
	//check last time after the loop
	if err != nil {
		return shim.Error("Failed to get randomNum: " + err.Error())
	} else if randNameAsBytes == nil {
		return shim.Error("Failed to get randomNums, not assigned yet.")
	} else {
		var randomNums randomNums
		err = json.Unmarshal(randNameAsBytes, &randomNums) //unmarshal it aka JSON.parse()
		if err != nil {
			return shim.Error(err.Error())
		}
		for i := 0; i < len(randomNums.Object); i++ {
			//convert the string to big.int
			rbigint := new(big.Int)
			rbigint, _ = rbigint.SetString(randomNums.Object[i].RandNum, 10)//10 is base
			r[i] = rbigint
		}
		//alpha
		rabigint := new(big.Int)
		rabigint, _ = rabigint.SetString(randomNums.Ralpha, 10)//10 is base
		alpha = rabigint
		//rho
		rrbigint := new(big.Int)
		rrbigint, _ = rrbigint.SetString(randomNums.Rrho, 10)//10 is base
		rho = rrbigint
		//tau1
		rt1bigint := new(big.Int)
		rt1bigint, _ = rt1bigint.SetString(randomNums.Rtau1, 10)//10 is base
		tau1 = rt1bigint
		//tau2
		rt2bigint := new(big.Int)
		rt2bigint, _ = rt2bigint.SetString(randomNums.Rtau2, 10)//10 is base
		tau2 = rt2bigint
		for i := 0; i < len(randomNums.ObjectV1); i++ {
			//convert the string to big.int
			rbigint := new(big.Int)
			rbigint, _ = rbigint.SetString(randomNums.ObjectV1[i], 10)//10 is base
			sL[i] = rbigint
		}
		for i := 0; i < len(randomNums.ObjectV2); i++ {
			//convert the string to big.int
			rbigint := new(big.Int)
			rbigint, _ = rbigint.SetString(randomNums.ObjectV2[i], 10)//10 is base
			sR[i] = rbigint
		}
	}

	fmt.Println("3)\tReading all previsously assigned random numbers for the spending entity...")
  //rP := make([]*big.Int, TXval)
	rA := big.NewInt(0)
  for j := 0; j < TXval; j++ {
  	// ==== Check if random numbers already exists ====
  	randNameP := "TX"+strconv.Itoa(j)+"RandNum"
  	//fmt.Println("randName = ", randName)
  	randNameAsBytesP, err := stub.GetState(randNameP)
  	if err != nil {
  		return shim.Error("Failed to get randomNum: " + err.Error())
  	} else if randNameAsBytes == nil {
  		return shim.Error("Failed to get randomNums, not assigned yet.")
  	} else {
  		var randomNumsP randomNums
  		err = json.Unmarshal(randNameAsBytesP, &randomNumsP) //unmarshal it aka JSON.parse()
  		if err != nil {
  			return shim.Error(err.Error())
  		}
  		for i := 0; i < len(randomNumsP.Object) ; i++ {
  			if(randomNumsP.Object[i].Name == A){//A is the spending entity
  				//convert the string to big.int
  				rbigint := new(big.Int)
  				rbigint, _ = rbigint.SetString(randomNumsP.Object[i].RandNum, 10)//10 is base
  				//rP[j] = rbigint
					rA = rA.Add(rA, rbigint)
  				break
  			}
  		}
  	}
  }


	fmt.Println("4)\tReading all previsouly Pedersen Commitments and computing the product of them for the spending entity...")
	PcommitA := pb.ECPoint{big.NewInt(0), big.NewInt(0)}
	for j := 0; j < TXval; j++ {
		txNameP := "TX"+strconv.Itoa(j)
		txAsBytesP, err := stub.GetState(txNameP)
  	if err != nil {
  		return shim.Error("Failed to get tx: " + err.Error())
  	} else if txAsBytesP == nil {
  		return shim.Error("Failed to get tx, not written into ledger yet.")
  	} else {
			var zkrowP zkrow
  		err = json.Unmarshal(txAsBytesP, &zkrowP) //unmarshal it aka JSON.parse()
  		if err != nil {
  			return shim.Error(err.Error())
  		}
			pcomStrP := zkrowP.Object[Aindex].ZKElement.Commitment
			PCX := new(big.Int)
			PCX, _ = PCX.SetString(pcomStrP.X, 10)//10 is base
			PCY := new(big.Int)
			PCY, _ = PCY.SetString(pcomStrP.Y, 10)//10 is base
			pcomP := pb.ECPoint{PCX, PCY}
			PcommitA = PcommitA.Add(pcomP)
		}
	}

	// randNameAsBytes, err := stub.GetState(randName)
	// if err != nil {
	// 	return shim.Error("Failed to get randomNum: " + err.Error())
	// } else if randNameAsBytes != nil {
	// 	//fmt.Println("The randoms are already assigned.")
	// 	var randomNums randomNums
	// 	err = json.Unmarshal(randNameAsBytes, &randomNums) //unmarshal it aka JSON.parse()
	// 	if err != nil {
	// 		return shim.Error(err.Error())
	// 	}
	// 	for i := 0; i < len(randomNums.Object); i++ {
	// 		//convert the string to big.int
	// 		rbigint := new(big.Int)
	// 		rbigint, _ = rbigint.SetString(randomNums.Object[i].RandNum, 10)//10 is base
	// 		r[i] = rbigint
	// 	}
	// } else {//random numbers does not assigned yet
	// 	return shim.Error("Failed to get randomNums, not assigned yet.")
	// }

	// fmt.Println("3)\tGenerating random numbers for next tx...")
	// randNameNext := "TX"+strconv.Itoa(TXval+1)+"RandNum"
	// rNext := pb.GetR(banknum) //generate r for each tx
	// var randomNumsNext randomNums
	// randomNumsNext.ObjectType = "RandNum"
	// randomNumsNext.Object = make([]randomNum, 0)
	// for j := range rNext {
	// 	rbigint := *rNext[j]
	// 	randomNum := randomNum{bankname[j], rbigint.String()}//save *big.Int as string
	// 	randomNumsNext.Object = append(randomNumsNext.Object, randomNum)
	// }
	// rJSONasBytesNext, err := json.Marshal(randomNumsNext)
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }
	// // === Save random number to state ===
	// err = stub.PutState(randNameNext, rJSONasBytesNext)
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }

	fmt.Println("5)\tComputing Pedersen commiments, Token and Rangeproof...")
	fmt.Println("value = ", value)
	fmt.Println("r = ", r)
	fmt.Println("PK = ", PK)
	var zkrow zkrow
	zkrow.ObjectType = "ZKrow"
	zkrow.Object = make([]zkelement, 0)
	for i := 0; i < banknum; i++ {
		Pcommitment, Token := pb.PCommitToken(value[i],r[i],PK[i])
		fmt.Println("Pcommit, Token are ", Pcommitment, Token)
		rpresult := pb.RangeProof{}
		if(bankname[i] == A){//spending entity
			rA = rA.Add(rA, r[i]) //add the spending bank's current random number
			rpresult = pb.RPProveNew(big.NewInt(int64(Avaleft)),rA,alpha,rho,sL,sR,tau1,tau2)
		} else {
			rpresult = pb.RPProveNew(value[i],r[i],alpha,rho,sL,sR,tau1,tau2)//this rangeproof actually doesn't matter, they only used to hide the real RP in the spending entity
		}
		//fmt.Println("RangeProof = ", rpresult)
		//ZKElement := pb.ZKElement{Pcommitment, Token, rpresult}
		//converting the RangeProof into RangeProofStr
		Lstr := make([]ECPointStr, len(rpresult.IPP.L))
		for j := range Lstr {
			Lstr[j] = ECPointStr{rpresult.IPP.L[j].X.String(), rpresult.IPP.L[j].Y.String()}
		}
		Rstr := make([]ECPointStr, len(rpresult.IPP.R))
		for j := range Rstr {
			Rstr[j] = ECPointStr{rpresult.IPP.R[j].X.String(), rpresult.IPP.R[j].Y.String()}
		}
		challstr := make([]string, len(rpresult.IPP.Challenges))
		for j := range challstr {
			challstr[j] = rpresult.IPP.Challenges[j].String()
		}
		ippStr := InnerProdArgStr{
			Lstr,
			Rstr,
			rpresult.IPP.A.String(),
			rpresult.IPP.B.String(),
			challstr,
		}
		rpresultStr := RangeProofStr{
			ECPointStr{rpresult.Comm.X.String(), rpresult.Comm.Y.String()}, //Comm
			ECPointStr{rpresult.A.X.String(), rpresult.A.Y.String()}, //A
			ECPointStr{rpresult.S.X.String(), rpresult.S.Y.String()}, //S
			ECPointStr{rpresult.T1.X.String(), rpresult.T1.Y.String()}, //T1
			ECPointStr{rpresult.T2.X.String(), rpresult.T2.Y.String()}, //T2
			rpresult.Tau.String(),//Tau
			rpresult.Th.String(),//Th
			rpresult.Mu.String(),//Mu
			ippStr,
			rpresult.Cy.String(),//Cy
			rpresult.Cz.String(),//Cz
			rpresult.Cx.String(),//Cx
		}
		// testC := convertBIntStrArry(rpresultStr.IPP.Challenges)
		// for index := 0; index < len(testC); index++{
		// 	fmt.Println(testC[index])
		// }
		if(i == Aindex){
			PcommitA = PcommitA.Add(Pcommitment) //spending bank needs to multiple the current PC into the PCAll
		}
		ZKElement := ZKElementStr{
			ECPointStr{Pcommitment.X.String(), Pcommitment.Y.String()},
			ECPointStr{Token.X.String(), Token.Y.String()},
			rpresultStr,
			ECPointStr{PcommitA.X.String(), PcommitA.Y.String()},
		}
		zkelement := zkelement{bankname[i], ZKElement}
		zkrow.Object = append(zkrow.Object, zkelement)
	}

	// Write the state to the ledger
	fmt.Println("6)\tWriting Pedersen commiments, Token and Rangeproof into the ledger...")
	zkJSONasBytes, err := json.Marshal(zkrow)
	if err != nil {
		return shim.Error(err.Error())
	}
	// === Save pubkey to state ===
	txName := "TX"+strconv.Itoa(TXval)
	err = stub.PutState(txName, zkJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//fmt.Println("Writing TX ID into the ledger...")
	err = stub.PutState("TXID", []byte(strconv.Itoa(TXval)))
	if err != nil {
		return shim.Error(err.Error())
	}
	endTime := time.Now()
	fmt.Println("Chaincode StartTime : ", beginTime)
	fmt.Println("Chaincode EndTime : ", endTime)

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

// query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

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
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
