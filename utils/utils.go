package utils

import (
	"time"
	"webapp/db"
	"webapp/models"
)

type VerificationError int

const (
	NoteDoesNotExist VerificationError = iota
	NoteAmountTooSmall
)

// Error returns the error message for the VerificationError.
func (e VerificationError) Error() string {
	switch e {
	case NoteDoesNotExist:
		return "The secret note you provided does not exist"
	case NoteAmountTooSmall:
		return "The withdrawal amount is greater than the available funds"
	default:
		return "Internal error"
	}
}

// VerifyWithdrawal verifies a withdrawal is valid
func VerifyWithdrawal(w *models.WithdrawData) (bool, error) {
	// TODO: implement this function
	time.Sleep(2 * time.Second)
	existNote, err := db.ExistNote(w.OldNote)
	if !existNote {
		return false, NoteDoesNotExist
	}
	if err != nil {
		return false, err
	}
	if w.Amount.Microalgos > w.OldNote.Amount {
		return false, NoteAmountTooSmall
	}
	db.DeleteNote(w.OldNote)
	db.SaveWithdrawal(w)
	return true, nil
}
