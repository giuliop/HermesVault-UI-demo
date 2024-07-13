package main

import (
	"fmt"
	"strconv"
	"strings"
)

// stringToMicroAlgo converts a string representing an algo amount to microalgos.
func stringToMicroAlgo(amountStr string) (uint64, error) {
	intStr, decStr, hasDecimal := strings.Cut(amountStr, ".")
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
