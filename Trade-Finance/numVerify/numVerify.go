package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	
	"errors"	
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// ManagePO example simple Chaincode implementation
type ManagePO struct {
}

var phoneDetailsIndexStr = "_PhoneDetailsindex"			//name for the key/value that will store a list of all known PhoneDetails

type Numverify struct {
	Valid               bool   `json:"valid"`
	Number              string `json:"number"`
	LocalFormat         string `json:"local_format"`
	InternationalFormat string `json:"international_format"`
	CountryPrefix       string `json:"country_prefix"`
	CountryCode         string `json:"country_code"`
	CountryName         string `json:"country_name"`
	Location            string `json:"location"`
	Carrier             string `json:"carrier"`
	LineType            string `json:"line_type"`
}

// ============================================================================================================================
// Main - start the chaincode for Verify Number
// ============================================================================================================================
func main() {			
	err := shim.Start(new(ManagePO))
	if err != nil {
		fmt.Printf("Error starting Verify Number chaincode: %s", err)
	}
}
// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *ManagePO) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var msg string
	var err error
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	// Initialize the chaincode
	msg = args[0]
	fmt.Println("NumVerify chaincode is deployed successfully.");
	
	// Write the state to the ledger
	err = stub.PutState("chalpat", []byte(msg)) //making a test var "chalpat", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}
	
	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)				//marshal an emtpy array of strings to clear the index
	err = stub.PutState(phoneDetailsIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}
// ============================================================================================================================
// Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *ManagePO) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}
// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *ManagePO) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {			// initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "verify_number" {	// verify Number
		return t.verifyNumber(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)		// error
	return nil, errors.New("Received unknown function invocation")
}
// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *ManagePO) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	//fmt.Println("query did not find func: " + function)						//error
	return nil, errors.New("Received unknown function query")
}
// ============================================================================================================================
// verify number - verify number, store into chaincode state
// ============================================================================================================================
func (t *ManagePO) verifyNumber(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("start verifyNumber")

	phone := "14158586273"
	// QueryEscape escapes the phone string so
	// it can be safely placed inside a URL query
	safePhone := url.QueryEscape(phone)

	url := fmt.Sprintf("http://apilayer.net/api/validate?access_key=917d879a204238e61549b2f04d4e795b&number=%s", safePhone)

	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return nil, err
	}

	// For control over HTTP client headers,
	// redirect policy, and other settings,
	// create a Client
	// A Client is an HTTP client
	client := &http.Client{}

	// Send the request via a client
	// Do sends an HTTP request and
	// returns an HTTP response
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return nil, err
	}

	fmt.Println("The response is::"+strconv.Itoa(resp.StatusCode))

	// Callers should close resp.Body
	// when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()

	// Fill the record with the data from the JSON
	var record Numverify

	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		log.Println(err)
	}

	// build the phone_details json string manually
	phone_details := 	`{`+
		`"phone": "` + phone + `" , `+
		`"country": "` + record.CountryName + `" , `+
		`"location": "` + record.Location + `" , `+
		`"carrier": "` + record.Carrier + `" , `+ 
		`"linetype": "` + record.LineType + `" , `+ 
	`}`
	fmt.Println("phone_details: " + phone_details)
	fmt.Print("order in bytes array: ")
	fmt.Println([]byte(phone_details))
	
	err = stub.PutState(phone, []byte(phone_details))		//store Phone Details with phone number as key
	if err != nil {
		return nil, err
	}

	fmt.Println("Phone No. = ", record.InternationalFormat)
	fmt.Println("Country   = ", record.CountryName)
	fmt.Println("Location  = ", record.Location)
	fmt.Println("Carrier   = ", record.Carrier)
	fmt.Println("LineType  = ", record.LineType)
	
	//get the PhoneDetails index
	phoneDetailsIndexAsBytes, err := stub.GetState(phoneDetailsIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get phone details index")
	}
	var phoneDetailsIndex []string
	//fmt.Println("phone details index as bytes::"+phoneDetailsIndexAsBytes)
	
	json.Unmarshal(phoneDetailsIndexAsBytes, &phoneDetailsIndex)				//un stringify it aka JSON.parse()
	fmt.Print("phoneDetailsIndex after unmarshal..before append: ")
	fmt.Println(phoneDetailsIndex)
	//append
	phoneDetailsIndex = append(phoneDetailsIndex, phone)					//add phone details phone to index list
	fmt.Println("! Phone index after appending phone number: ", phoneDetailsIndex)
	jsonAsBytes, _ := json.Marshal(phoneDetailsIndex)
	fmt.Print("jsonAsBytes: ")
	fmt.Println(jsonAsBytes)
	err = stub.PutState(phoneDetailsIndexStr, jsonAsBytes)					//store name of phone
	if err != nil {
		return nil, err
	}

	fmt.Println("end verifyNumber")
	return nil, nil
}
