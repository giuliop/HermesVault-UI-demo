package handlers

import (
	"log"
	"net/http"
	"webapp/config"
	"webapp/models"
	"webapp/templates"
	"webapp/utils"
)

func WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", config.CacheControl)
	switch r.Method {
	case http.MethodGet:
		if err := templates.Withdraw.Execute(w, nil); err != nil {
			log.Printf("Error executing withdraw template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
		amount, errAmount := models.Input(r.FormValue("amount")).ToAmount()
		address, errAddress := models.Input(r.FormValue("address")).ToAddress()
		note, errNote := models.Input(r.FormValue("note")).ToNote()
		errorMsg := ""
		if errAmount != nil {
			log.Printf("Error parsing withdrawal amount: %v", errAmount)
			errorMsg += "Invalid algo amount<br>"
		}
		if errAddress != nil {
			log.Printf("Error parsing withdrawal address: %v", errAddress)
			errorMsg += "Invalid Algorand address<br>"
		}
		if errNote != nil {
			log.Printf("Error parsing withdrawal note: %v", errNote)
			errorMsg += "Invalid secret note"
		}
		if errorMsg != "" {
			http.Error(w, errorMsg, http.StatusUnprocessableEntity)
			return
		}
		withdrawData := &models.WithdrawData{
			Amount:  amount,
			Fee:     amount.Fee(),
			Address: address,
			OldNote: note,
			NewNote: nil,
		}
		isValid, err := utils.VerifyWithdrawal(withdrawData)
		if isValid {
			newNote, err := models.GenerateChangeNote(amount, note)
			if err != nil {
				log.Printf("Error generating new note: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			withdrawData.NewNote = newNote
			if err := templates.ConfirmWithdrawal.Execute(w, withdrawData); err != nil {
				log.Printf("Error executing success template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		} else {
			if err == utils.NoteAmountTooSmall || err == utils.NoteDoesNotExist {
				errorMsg = "<b>Proof verification failed.</b><br>" + err.Error()
				http.Error(w, errorMsg, http.StatusUnprocessableEntity)
			} else {
				errorMsg = "<b>Something went wrong.</b><br>Please try again."
				http.Error(w, errorMsg, http.StatusInternalServerError)
			}
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
