package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/algorand/go-algorand-sdk/types"
)

const BaseFeeDivisor = 1000 // 0.1% base fee
const MinimumFee = 100_000  // microalgos

// stringToMicroAlgo converts an input representing an algo amount to an Amount
func (input Input) toAmount() (*Amount, error) {
	intStr, decStr, hasDecimal := strings.Cut(string(input), ".")
	integer, err := strconv.ParseUint(intStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid integer part: %w", err)
	}
	var decimal uint64
	if hasDecimal {
		if len(decStr) > 6 {
			return nil, fmt.Errorf("too many decimal places")
		}
		if len(decStr) < 6 {
			decStr += strings.Repeat("0", 6-len(decStr))
		}
		decimal, err = strconv.ParseUint(decStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid decimal part: %w", err)
		}
	}
	microalgos := integer*1_000_000 + decimal
	return &Amount{
		Algostring: microAlgosToAlgoString(microalgos),
		Microalgos: microalgos,
	}, nil
}

// toAddress converts an input to an Address
func (input Input) toAddress() (*Address, error) {
	a := Address{}
	native, err := types.DecodeAddress(string(input))
	if err != nil {
		return nil, fmt.Errorf("error decoding address: %v", err)
	}
	a.Native = native
	a.Start, a.Middle, a.End = splitEnds(native.String(), 5)
	return &a, nil
}

// toNote converts an input to a Note
// Input is expected to be a hex-encoded string of 70 bytes (140 hex characters),
// 8 bytes for the amount, 31 bytes for K, 31 bytes for R
func (input Input) toNote() (*Note, error) {
	if len(input) != 140 {
		return nil, errors.New("invalid secret note length")
	}
	decoded, err := hex.DecodeString(string(input))
	if err != nil {
		return nil, fmt.Errorf("error decoding hex string: %v", err)
	}
	amount := decoded[:8]
	var k, r [31]byte
	copy(k[:], decoded[8:39])
	copy(r[:], decoded[39:])
	return &Note{
		Amount: binary.BigEndian.Uint64(amount),
		K:      k,
		R:      r,
		Text:   string(input),
	}, nil
}

// splitEnds splits a string into three parts: the first n characters, the middle
// part, the last n characters
func splitEnds(s string, n int) (string, string, string) {
	if len(s) <= 2*n {
		return s, "", ""
	}
	return s[:n], s[n : len(s)-n], s[len(s)-n:]
}

// TODO: implement this function
func verifyWithdrawal(amount *Amount, address *Address, note *Note,
) bool {
	// sleep for 3 sec
	time.Sleep(3 * time.Second)
	// return true is amount is 100 algos, false otherwise
	return amount.Microalgos != 1_000_000
}

// generateNote generates a new note for a given amount
func generateNote(amount uint64) (*Note, error) {
	k, errK := generateRandom31()
	r, errR := generateRandom31()
	if errK != nil || errR != nil {
		return nil, fmt.Errorf("error generating random bytes for k / r: %v / %v",
			errK, errR)
	}
	return &Note{
		Amount: amount,
		K:      k,
		R:      r,
		// fmt.Sprintf in hex format results in big-endian representation
		Text: fmt.Sprintf("%016x%x%x", amount, k, r),
	}, nil
}

// generateDepositNote generates a new deposit note for the change amount
// after a withdrawal
func generateChangeNote(withdrawalAmount *Amount, fromNote *Note) (*Note, error) {
	change := (fromNote.Amount - withdrawalAmount.Microalgos -
		calculateFee(withdrawalAmount.Microalgos))
	note, err := generateNote(change)
	if err != nil {
		return nil, fmt.Errorf("error generating note: %v", err)
	}
	return note, nil
}

// generateRandom31 generates a cryptographically secure random 31-byte array
func generateRandom31() ([31]byte, error) {
	var arr [31]byte
	_, err := rand.Read(arr[:])
	if err != nil {
		return [31]byte{}, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return arr, nil
}

// calculateFee calculates the fee for a given amount; the fee is 0.1% of the amount with a minimum of 1000 microalgos
func calculateFee(amount uint64) uint64 {
	fee := amount / BaseFeeDivisor
	if fee < MinimumFee {
		fee = MinimumFee
	}
	return fee
}

// Fee calculates the fee for a withdrawal amount
func (withdrawalAmount *Amount) Fee() *Amount {
	fee := calculateFee(withdrawalAmount.Microalgos)
	return &Amount{
		Algostring: microAlgosToAlgoString(fee),
		Microalgos: fee,
	}
}

// microAlgosToAlgoString converts microalgos (uint64) to a string representing algos.
func microAlgosToAlgoString(microalgos uint64) string {
	const microAlgosPerAlgo = 1_000_000
	wholeAlgos := microalgos / microAlgosPerAlgo
	remainingMicroAlgos := microalgos % microAlgosPerAlgo
	s := fmt.Sprintf("%d.%06d", wholeAlgos, remainingMicroAlgos)
	s = strings.TrimRight(s, "0")
	if s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}
	return s
}
