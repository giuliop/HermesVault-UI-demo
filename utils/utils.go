package utils

import (
	"log"
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
	if w.Amount.Microalgos+w.Fee.Microalgos > w.OldNote.Amount {
		return false, NoteAmountTooSmall
	}
	return true, nil
}

// CommitWithdrawal saves a withdrawal to the database, deleting the old note
func CommitWithdrawal(w *models.WithdrawData) error {
	err := db.SaveWithdrawal(w)
	if err != nil {
		return err
	}

	// try 10 times to delete the old note with a delay between attempts
	for i := 0; i < 10; i++ {
		err = db.DeleteNote(w.OldNote)
		if err == nil {
			return nil
		}
		log.Printf("Attempt %d to delete old note failed: %v", i+1, err)
		time.Sleep(100 * time.Millisecond) // add a delay between retries
	}
	return err
}
