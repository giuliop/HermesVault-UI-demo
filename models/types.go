// package contains the data types used in the application
package models

import (
	"crypto/rand"
	"fmt"
	"strings"
	"webapp/config"
)

// Address represents a valid Algorand address
type Address string

type Amount struct {
	Algostring string
	Microalgos uint64
}

type Note struct {
	Amount uint64
	K      [31]byte
	R      [31]byte
	Text   string
}

type WithdrawData struct {
	Amount  *Amount
	Fee     *Amount
	Address Address
	OldNote *Note
	NewNote *Note
}

type DepositData struct {
	Amount  *Amount
	Address Address
	Note    *Note
}

// Fee calculates the fee for a withdrawal amount
func (withdrawalAmount *Amount) Fee() *Amount {
	fee := CalculateFee(withdrawalAmount.Microalgos)
	return &Amount{
		Algostring: MicroAlgosToAlgoString(fee),
		Microalgos: fee,
	}
}

// GenerateNote generates a new note for a given amount
func GenerateNote(amount uint64) (*Note, error) {
	k, errK := GenerateRandom31()
	r, errR := GenerateRandom31()
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
func GenerateChangeNote(withdrawalAmount *Amount, fromNote *Note) (*Note, error) {
	change := (fromNote.Amount - withdrawalAmount.Microalgos -
		CalculateFee(withdrawalAmount.Microalgos))
	note, err := GenerateNote(change)
	if err != nil {
		return nil, fmt.Errorf("error generating note: %v", err)
	}
	return note, nil
}

// GenerateRandom31 generates a cryptographically secure random 31-byte array
func GenerateRandom31() ([31]byte, error) {
	var arr [31]byte
	_, err := rand.Read(arr[:])
	if err != nil {
		return [31]byte{}, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return arr, nil
}

// CalculateFee calculates the fee for a given amount; the fee is 0.1% of the amount with a minimum of 1000 microalgos
func CalculateFee(amount uint64) uint64 {
	fee := amount / config.BaseFeeDivisor
	if fee < config.MinimumFee {
		fee = config.MinimumFee
	}
	return fee
}

// MicroAlgosToAlgoString converts microalgos (uint64) to a string representing algos.
func MicroAlgosToAlgoString(microalgos uint64) string {
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

func (a Address) Start() string {
	return SplitEnds(string(a), config.NumCharsToHighlight, Start)
}
func (a Address) Middle() string {
	return SplitEnds(string(a), config.NumCharsToHighlight, Middle)
}
func (a Address) End() string {
	return SplitEnds(string(a), config.NumCharsToHighlight, End)
}

type part int

const (
	Start = iota
	Middle
	End
)

// SplitEnds splits a string into three parts: the first n characters, the middle
// part, the last n characters
func SplitEnds(s string, n int, part part) string {
	var start, middle, end string
	if len(s) <= 2*n {
		start, middle, end = s, "", ""
	} else {
		start, middle, end = s[:n], s[n:len(s)-n], s[len(s)-n:]
	}
	if part == Start {
		return start
	} else if part == Middle {
		return middle
	} else {
		return end
	}
}
