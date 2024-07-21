package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/algorand/go-algorand-sdk/types"
)

// stringToMicroAlgo converts an input representing an algo amount to microalgos
func (input Input) toMicroAlgo() (uint64, error) {
	intStr, decStr, hasDecimal := strings.Cut(string(input), ".")
	integer, err := strconv.ParseUint(intStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid integer part: %w", err)
	}
	var decimal uint64
	if hasDecimal {
		if len(decStr) > 6 {
			return 0, fmt.Errorf("too many decimal places")
		}
		if len(decStr) < 6 {
			decStr += strings.Repeat("0", 6-len(decStr))
		}
		decimal, err = strconv.ParseUint(decStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid decimal part: %w", err)
		}
	}
	return integer*1_000_000 + decimal, nil
}

// toAddress converts an input to an Algorand address
func (input Input) toAddress() (types.Address, error) {
	address := string(input)
	return types.DecodeAddress(address)
}

// toSecretNote converts an input to the k, r values of a secret note.
// Input is expected to be a hex-encoded string of 62 bytes (124 hex characters).
func (input Input) toSecretNote() ([]byte, []byte, error) {
	if len(input) != 124 {
		return nil, nil, errors.New("invalid secret note length")
	}
	decoded, err := hex.DecodeString(string(input))
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding hex string: %v", err)
	}
	k := decoded[:31]
	r := decoded[31:]
	return k, r, nil
}

func verifyWithdrawal(amount uint64, address types.Address, noteK, noteR []byte,
) bool {
	// sleep for 3 sec
	time.Sleep(3 * time.Second)
	// return true is amount is 100 algos, false otherwise
	return amount == 100_000_000
}
