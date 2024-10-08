// Licensed to The Moov Authors under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. The Moov Authors licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/moov-io/ach"
)

func main() {
	// Open a file for reading, any io.Reader can be used
	f, err := os.Open("enr-read.ach")
	if err != nil {
		log.Fatalln(err)
	}
	r := ach.NewReader(f)
	achFile, err := r.Read()
	if err != nil {
		log.Fatalf("reading file: %v\n", err)
	}
	// If you trust the file but its formatting is off, building will probably resolve the malformed file
	if err := achFile.Create(); err != nil {
		log.Fatalf("creating file: %v\n", err)
	}
	// Validate the ACH file
	if err := achFile.Validate(); err != nil {
		log.Fatalf("validating file: %v\n", err)
	}

	fmt.Printf("Total Amount: %d\n", achFile.Batches[0].GetEntries()[0].Amount)
	fmt.Printf("SEC Code: %s\n", achFile.Batches[0].GetHeader().StandardEntryClassCode)

	batch, ok := achFile.Batches[0].(*ach.BatchENR)
	if !ok {
		log.Fatalf("Batch not ENR, got %T %#v\n", achFile.Batches[0], achFile.Batches[0])
	}
	add := batch.GetEntries()[0].Addenda05[0]

	fmt.Printf("Payment Related Information: %s\n", add.PaymentRelatedInformation)
	info, err := ach.ParseENRPaymentInformation(add)
	if err != nil {
		log.Fatalf("Problem Parsing ENR Addenda05 PaymentRelatedInformation: %v\n", err)
	}
	fmt.Println(info.String())
}
