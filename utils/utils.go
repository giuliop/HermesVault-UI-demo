package utils

import (
	"time"
	"webapp/models"
)

// TODO: implement this function
func VerifyWithdrawal(amount *models.Amount, address *models.Address, note *models.Note,
) bool {
	// sleep for 3 sec
	time.Sleep(3 * time.Second)
	// return true is amount is 100 algos, false otherwise
	return amount.Microalgos != 1_000_000
}
