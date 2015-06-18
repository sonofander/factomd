// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
    "time"
    "bytes"
    "github.com/hoisie/web"
    fct "github.com/FactomProject/factoid"
    "github.com/FactomProject/factoid/state/stateinit"
    "github.com/btcsuitereleases/btcutil/base58"
)

var _ = fct.Address{}

const (
    httpOK  = 200
    httpBad = 400
)

var (
	cfg = fct.ReadConfig().Wallet
    ipAddress        = cfg.Address
    portNumber       = cfg.Port
    applicationName  = "Factom/fctwallet"
    dataStorePath    = cfg.DataFile
    refreshInSeconds = cfg.RefreshInSeconds
    
    ipaddressFD      = "localhost:"
    portNumberFD     = "8088"
)

var factoidState = stateinit.NewFactoidState("/tmp/factoid_wallet_bolt.db")

var server = web.NewServer()

 
func Start() {
    // Balance
    // localhost:8089/v1/factoid-balance/<name or address>
    // Returns the balance of factoids at that address, or the address tied to
    // the given name.
    server.Get("/v1/factoid-balance/([^/]+)", handleFactoidBalance)
    
    // Generate Address
    // localhost:8089/v1/factoid-generate-address/<name>
    // Generate an address, and tie it to the given name within the wallet. You
    // can use the name for the address in this API
    server.Get("/v1/factoid-generate-address/([^/]+)", handleFactoidGenerateAddress)
    
    // New Transaction
    // localhost:8089/v1/factoid-new-transaction/<key>
    // Use the key in subsequent calls to add inputs, outputs, ecoutputs, and to
    // sign and submit the transaction. Returns Success == true if all is well.
    // Multiple transactions can be in process.  Only one transaction per key.
    // Once the transaction has been submitted or deleted, the key can be
    // reused.
    server.Post("/v1/factoid-new-transaction/([^/]+)", handleFactoidNewTransaction)

    // Add Input
    // localhost:8089/v1/factoid-add-input/?key=<key>&name=<name or address>&amount=<amount>
    // Add an input to a transaction in process.  Start with new-transaction.
    server.Post("/v1/factoid-add-input/(.*)", handleFactoidAddInput)
    
    // Add Output
    // localhost:8089/v1/factoid-add-output/?key=<key>&name=<name or address>&amount=<amount>
    // Add an output to a transaction in process.  Start with new-transaction.
    server.Post("/v1/factoid-add-output/(.*)", handleFactoidAddOutput)
    
    // Add Entry Credit Output
    // localhost:8089/v1/factoid-add-ecoutput/?key=<key>&name=<name or address>&amount=<amount>
    // Add an ecoutput to a transaction in process.  Start with new-transaction.
    server.Post("/v1/factoid-add-ecoutput/(.*)", handleFactoidAddECOutput)
    
    // Sign Transaction
    // localhost:8089/v1/factoid-sign-transaction/<key>
    // If the transaction validates structure wise and all signatures can be
    // applied, then all inputs are signed, and returns success = true
    // Otherwise returns false. Note that this doesn't check that the inputs
    // can cover the transaction.  Use validate to do that.
    server.Post("/v1/factoid-sign-transaction/(.*)", handleFactoidSignTransaction)
    
    // Submit
    // localhost:8089/v1/factoid-submit/
    // Put the key for the transaction in {Transaction string}
    server.Post("/v1/factoid-submit/", handleFactoidSubmit)
    
    // Validate
    // localhost:8089/v1/factoid-validate/<key>
    // Validates amounts and that all required signatures are applied, returns success = true
    // Otherwise returns false.
    server.Get("/v1/factoid-validate/(.*)", handleFactoidValidate)
    
    // Get Fee
    // localhost:8089/v1/factoid-get-fee/
    // Get the Transaction fee
    server.Get("/v1/factoid-get-fee/", handleGetFee)
    
    go server.Run(fmt.Sprintf("%s:%d", ipAddress, portNumber))
}   
 
func main() {
    Start()
    for { 
        time.Sleep(time.Second)
    }    
}

/****************************
 * Helper Functions
 ****************************/


// Factoid Address
// 
// 
// Add a prefix of 0x5fb1 at the start, and the first 4 bytes of a SHA256d to
// the end.  Using zeros for the address, this might look like:
// 
//     5fb10000000000000000000000000000000000000000000000000000000000000000d48a8e32
// 
// A typical Factoid Address:
//
//     FA1y5ZGuHSLmf2TqNf6hVMkPiNGyQpQDTFJvDLRkKQaoPo4bmbgu
// 
// Entry credits only differ by the prefix of 0x592a and typically look like:
//
//     EC3htx3MxKqKTrTMYj4ApWD8T3nYBCQw99veRvH1FLFdjgN6GuNK
//
// More words on this can be found here:
//
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#human-readable-addresses
//

var FactoidPrefix = []byte{ 0x5f, 0xb1 }
var EntryCreditPrefix = []byte{ 0x59, 0x2a }

//  Convert Factoid and Entry Credit addresses to their more user
//  friendly and human readable formats.
//
//  Creates the binary form.  Just needs the conversion to base58
//  for display.
func ConvertAddressToUser(prefix []byte, addr fct.IAddress) []byte {
    sha256d := fct.Sha(fct.Sha(addr.Bytes()).Bytes()).Bytes()
    userd := make([]byte,0,32)
    userd = append(userd, prefix...)
    userd = append(userd, addr.Bytes()...)
    userd = append(userd, sha256d[:4]...)
    return userd
}

// Convert Factoid Addresses
func ConvertFAddressToUserStr(addr fct.IAddress) string {
    userd := ConvertAddressToUser(FactoidPrefix, addr)
    return base58.Encode(userd)
}

// Convert Entry Credits
func ConvertECAddressToUserStr(addr fct.IAddress) string {
    userd := ConvertAddressToUser(FactoidPrefix, addr)
    return base58.Encode(userd)
}


//
// Validates a User representation of a Factom and 
// Entry Credit addresses.
//
// Returns false if the length is wrong.
// Returns false if the prefix is wrong.  
// Returns false if the checksum is wrong.
//
func validateUserStr(prefix []byte, userFAddr string) bool {
    if len(userFAddr) != 52 {  
        return false 
        
    }
    v := base58.Decode(userFAddr)
    if bytes.Compare(prefix, v[:2]) != 0 { 
        return false 
        
    }
    sha256d := fct.Sha(fct.Sha(v[2:34]).Bytes()).Bytes()
    if bytes.Compare (sha256d[:4],v[34:]) != 0 {
        return false 
    }
    return true
}

// Validate Factoids
func ValidateFUserStr(userFAddr string) bool {
    return validateUserStr(FactoidPrefix, userFAddr)
}

// Validate Entry Credits
func ValidateECUserStr(userFAddr string) bool {
    return validateUserStr(EntryCreditPrefix, userFAddr)
}

// Convert a User facing Factoid or Entry Credit address
// to the regular form.  Note validation must be done
// separately!
func ConvertUserStrToAddress(userFAddr string) []byte {
    v := base58.Decode(userFAddr)
    return v[2:34]
}


