package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"webapp/config"
	"webapp/db"
	"webapp/frontend/templates"
	"webapp/models"
)

func WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Cache-Control", config.CacheControl)
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
			errorMsg += "The note you provided is not valid"
		}
		if errorMsg != "" {
			http.Error(w, errorMsg, http.StatusUnprocessableEntity)
			return
		}
		withdrawData := &models.WithdrawalData{
			Amount:     amount,
			Fee:        amount.Fee(),
			Address:    address,
			FromNote:   note,
			ChangeNote: nil,
		}
		var err error
		withdrawData.FromNote.LeafIndex, err = db.GetLeafIndexByCommitment(
			withdrawData.FromNote.Commitment())
		switch err {
		case nil:
			changeNote, err := models.GenerateChangeNote(amount, note)
			if err != nil {
				log.Printf("Error generating new note: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			withdrawData.ChangeNote = changeNote
			if err := templates.ConfirmWithdrawal.Execute(w, &withdrawData); err != nil {
				log.Printf("Error executing success template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		case sql.ErrNoRows:
			errorMsg = "The note you provided is not valid<br>"
			http.Error(w, errorMsg, http.StatusUnprocessableEntity)
		default:
			errorMsg = "<b>Something went wrong.</b><br>Please try again."
			http.Error(w, errorMsg, http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
