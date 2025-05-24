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

package ach

import (
	"strconv"
)

// BatchCTX holds the BatchHeader and BatchControl and all EntryDetail for CTX Entries.
//
// The Corporate Trade Exchange (CTX) application provides the ability to collect and disburse
// funds and information between companies. Generally it is used by businesses paying one another
// for goods or services. These payments replace checks with an electronic process of debiting and
// crediting invoices between the financial institutions of participating companies.
type BatchCTX struct {
	Batch
}

// NewBatchCTX returns a *BatchCTX
func NewBatchCTX(bh *BatchHeader) *BatchCTX {
	batch := new(BatchCTX)
	batch.SetControl(NewBatchControl())
	batch.SetHeader(bh)
	batch.SetID(bh.ID)
	return batch
}

// Validate checks properties of the ACH batch to ensure they match NACHA guidelines.
// This includes computing checksums, totals, and sequence orderings.
//
// Validate will never modify the batch.
func (batch *BatchCTX) Validate() error {
	if batch.validateOpts != nil && batch.validateOpts.SkipAll {
		return nil
	}

	// basic verification of the batch before we validate specific rules.
	if err := batch.verify(); err != nil {
		return err
	}

	// Add configuration and type specific validation for this type.
	if batch.Header.StandardEntryClassCode != CTX {
		return batch.Error("StandardEntryClassCode", ErrBatchSECType, CTX)
	}

	invalidEntries := batch.InvalidEntries()
	if len(invalidEntries) > 0 {
		return invalidEntries[0].Error // return the first invalid entry's error
	}

	return nil
}

// InvalidEntries returns entries with validation errors in the batch
func (batch *BatchCTX) InvalidEntries() []InvalidEntry {
	var out []InvalidEntry

	for _, entry := range batch.Entries {
		addendaCount := len(entry.Addenda05)

		// Trapping this error, as entry.CTXAddendaRecordsField() can not be greater than 9999
		if addendaCount > 9999 {
			out = append(out, InvalidEntry{
				Entry: entry,
				Error: batch.Error("AddendaCount", NewErrBatchAddendaCount(len(entry.Addenda05), 9999)),
			})
		}

		// Add to addendaCount so Corrections and Returns compare AddendaRecordIndicator correctly
		if entry.Addenda98 != nil {
			addendaCount += 1
		}
		if entry.Addenda99 != nil {
			addendaCount += 1
		}

		// validate CTXAddendaRecord Field is equal to the actual number of Addenda records
		// use 0 value if there is no Addenda records
		indicator, _ := strconv.Atoi(entry.CATXAddendaRecordsField())
		if addendaCount != indicator {
			if batch.validateOpts == nil || !batch.validateOpts.UnequalAddendaCounts {
				out = append(out, InvalidEntry{
					Entry: entry,
					Error: batch.Error("AddendaCount", NewErrBatchExpectedAddendaCount(addendaCount, indicator)),
				})
			}
		}
		// Verify TransactionCode is valid for CTX
		switch entry.TransactionCode {
		case CheckingPrenoteCredit, CheckingPrenoteDebit,
			SavingsPrenoteCredit, SavingsPrenoteDebit,
			GLPrenoteCredit, GLPrenoteDebit, LoanPrenoteCredit:
			if entry.Amount != 0 {
				out = append(out, InvalidEntry{
					Entry: entry,
					Error: batch.Error("TransactionCode", ErrBatchTransactionCode, entry.TransactionCode),
				})
			}
		}
		// Verify the Amount is valid for SEC code and TransactionCode
		if err := batch.ValidAmountForCodes(entry); err != nil {
			out = append(out, InvalidEntry{
				Entry: entry,
				Error: err,
			})
		}
		// Verify the TransactionCode is valid for a ServiceClassCode
		if err := batch.ValidTranCodeForServiceClassCode(entry); err != nil {
			out = append(out, InvalidEntry{
				Entry: entry,
				Error: err,
			})
		}
		// Verify Addenda* FieldInclusion based on entry.Category and batchHeader.StandardEntryClassCode
		if err := batch.addendaFieldInclusion(entry); err != nil {
			out = append(out, InvalidEntry{
				Entry: entry,
				Error: err,
			})
		}
	}

	return out
}

// Create will tabulate and assemble an ACH batch into a valid state. This includes
// setting any posting dates, sequence numbers, counts, and sums.
//
// Create implementations are free to modify computable fields in a file and should
// call the Batch's Validate function at the end of their execution.
func (batch *BatchCTX) Create() error {
	// generates sequence numbers and batch control
	if err := batch.build(); err != nil {
		return err
	}
	// Additional steps specific to batch type
	// ...
	return batch.Validate()
}
