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
			errorMsg += "Please submit a valid algo amount<br>"
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
		isValid := utils.VerifyWithdrawal(amount, address, note)
		if isValid {
			newNote, err := models.GenerateChangeNote(amount, note)
			if err != nil {
				log.Printf("Error generating new note: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			data := models.WithdrawData{
				Amount:  amount,
				Fee:     amount.Fee(),
				Address: address,
				Note:    note,
				NewNote: newNote,
			}
			if err := templates.ConfirmWithdrawal.Execute(w, data); err != nil {
				log.Printf("Error executing success template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		} else {
			errorMsg = `<b>Proof verification failed.</b><br>
						The secret note you provided does not exist
						or does not contain enough funds.`
			http.Error(w, errorMsg, http.StatusUnprocessableEntity)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
