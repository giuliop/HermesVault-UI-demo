package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"webapp/db"
	"webapp/models"
)

func ConfirmDepositHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Bad Request. Your deposit was not processed.",
			http.StatusBadRequest)
		return
	}
	amount, errAmount := models.Input(r.FormValue("amount")).ToAmount()
	address, errAddress := models.Input(r.FormValue("address")).ToAddress()
	note, errNote := models.Input(r.FormValue("note")).ToNote()
	errorMsg := ""
	if errAmount != nil {
		log.Printf("Error parsing withdrawal amount: %v", errAmount)
		errorMsg += "Invalid deposit amount<br>"
	}
	if errAddress != nil {
		log.Printf("Error parsing withdrawal address: %v", errAddress)
		errorMsg += "Invalid Algorand address<br>"
	}
	if errNote != nil {
		log.Printf("Error parsing withdrawal address: %v", errAddress)
		errorMsg += "Invalid note<br>"
	}
	if errorMsg != "" {
		http.Error(w, errorMsg+"Your deposit was not processed.",
			http.StatusUnprocessableEntity)
		return
	}
	// TODO: send the deposit transaction to the Algorand network
	time.Sleep(2 * time.Second)

	depositData := &models.DepositData{
		Amount:  amount,
		Address: address,
		Note:    note,
	}
	err := db.SaveDeposit(depositData)
	if err != nil {
		log.Printf("Error saving deposit: %v", err)
		http.Error(w, "Something went wrong. Your deposit was not processed.",
			http.StatusInternalServerError)
		return
	}
	html := `<dialog class="modal">
			   <h1>Deposit successful</h1>
				 <p>
				   You can use your new secret note to withdraw your funds in the future.
				</p>
				<button hx-get="/withdraw"
						onclick="this.parentElement.close()">
				  Close
				</button>
			 </dialog>
			 <script>
			   document.querySelectorAll('dialog')[0].showModal()
			 </script>
			`
	fmt.Fprint(w, html)
}

func ConfirmWithdrawHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Bad Request. Your withdrawal was not processed.",
			http.StatusBadRequest)
		return
	}
	amount, errAmount := models.Input(r.FormValue("amount")).ToAmount()
	address, errAddress := models.Input(r.FormValue("address")).ToAddress()
	oldNote, errOldNote := models.Input(r.FormValue("oldNote")).ToNote()
	newNote, errNewNote := models.Input(r.FormValue("newNote")).ToNote()

	errorMsg := ""
	if errAmount != nil {
		log.Printf("Error parsing withdrawal amount: %v", errAmount)
		errorMsg += "Invalid withdrawal amount<br>"
	}
	if errAddress != nil {
		log.Printf("Error parsing withdrawal address: %v", errAddress)
		errorMsg += "Invalid withdrawal address<br>"
	}
	if errOldNote != nil {
		log.Printf("Error parsing withdrawal address: %v", errAddress)
		errorMsg += "Invalid secret note<br>"
	}
	if errNewNote != nil {
		log.Printf("Error parsing withdrawal address: %v", errAddress)
		errorMsg += "Invalid new secret note<br>"
	}
	if errorMsg != "" {
		http.Error(w, errorMsg+"Your withdrawal was not processed.",
			http.StatusUnprocessableEntity)
		return
	}
	// TODO: send the withdrawal transaction to the Algorand network
	time.Sleep(2 * time.Second)

	withdrawData := &models.WithdrawData{
		Amount:  amount,
		Fee:     amount.Fee(),
		Address: address,
		OldNote: oldNote,
		NewNote: newNote,
	}

	err := db.SaveWithdrawal(withdrawData)
	if err != nil {
		log.Printf("Error saving withdrawal: %v", err)
		http.Error(w, "Something went wrong. Your withdrawal was not processed.",
			http.StatusInternalServerError)
		return
	}

	html := `<dialog class="modal">
			   <h1>Withdrawal successful</h1>
				 <p>
				   You can use your new secret note to withdraw any remaining balance in the future.
				</p>
				<button hx-get="/withdraw"
						onclick="this.parentElement.close()">
				  Close
				</button>
			 </dialog>
			 <script>
			   document.querySelectorAll('dialog')[0].showModal()
			 </script>
			`
	fmt.Fprint(w, html)
}
