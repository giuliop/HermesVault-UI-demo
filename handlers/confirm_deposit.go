package handlers

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"webapp/avm"
	"webapp/db"
	"webapp/memstore"
	"webapp/models"

	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/types"
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
		log.Printf("Error parsing deposit amount: %v", errAmount)
		errorMsg += "Invalid deposit amount<br>"
	}
	if errAddress != nil {
		log.Printf("Error parsing deposit address: %v", errAddress)
		errorMsg += "Invalid Algorand address<br>"
	}
	if errNote != nil {
		log.Printf("Error parsing deposit note: %v", errNote)
		errorMsg += "Invalid note<br>"
	}
	if errorMsg != "" {
		http.Error(w, errorMsg+"Your deposit was not processed.",
			http.StatusUnprocessableEntity)
		return
	}

	signedTxnBase64 := r.FormValue("signedTxn")
	signedTxnBytes, err := base64.StdEncoding.DecodeString(signedTxnBase64)
	if err != nil {
		log.Printf("Error decoding signed transaction: %v", err)
		http.Error(w, "Bad Request. Your deposit was not processed.", http.StatusBadRequest)
		return
	}

	var signedTxn types.SignedTxn
	err = msgpack.Decode(signedTxnBytes, &signedTxn)
	if err != nil {
		log.Printf("Error decoding signed transaction: %v", err)
		http.Error(w, "Bad Request. Your deposit was not processed.", http.StatusBadRequest)
		return
	}

	groupId := signedTxn.Txn.Group
	ms := memstore.UserSessions
	depositData, err := ms.RetrieveDeposit(groupId)
	if err != nil {
		log.Printf("Error retrieving deposit data: %v", err)
		http.Error(w, "Something went wrong. Your deposit was not processed."+
			"<br>Please try again.",
			http.StatusInternalServerError)
		return
	}
	ms.DeleteDeposit(groupId)

	if amount.Microalgos != depositData.Amount.Microalgos || address != depositData.Address ||
		note.Text() != depositData.Note.Text() {
		log.Printf("deposit data does not match. Form submitted:\nAmount: %v\nAddress: "+
			"%v\nNote: %v\n, while memory store had Amount: %v\nAddress: %v\nNote: %v\n",
			amount, address, note, depositData.Amount, depositData.Address, depositData.Note)
		http.Error(w, "Bad Request. Your deposit was not processed.", http.StatusBadRequest)
		return
	}

	noteId, err := db.RegisterUnconfirmedNote(depositData.Note)
	if err != nil {
		log.Printf("Error saving unconfirmed deposit: %v", err)
		http.Error(w, "Something went wrong. Your deposit was not processed."+
			"<br>Please try again.",
			http.StatusInternalServerError)
		return
	}

	var leafIndex uint64
	var txnId string
	var confirmationError *avm.TxnConfirmationError
	var saveNoteToDbError error

	// We can delete the unconfirmed note if one of these is true:
	// * txn confirmed by the blockchain and the note saved to the database,
	// * txn rejected by the blockchain
	// * internal error sending the txn
	// If we timeout waiting for confirmation or get confirmation but fail to save to the db,
	// we keep the unconfirmed note, the cleanup process will eventually handle it
	defer func() {
		if (confirmationError == nil && saveNoteToDbError == nil) ||
			confirmationError.Type != avm.ErrWaitTimeout {
			db.DeleteUnconfirmedNote(noteId)
		}
	}()

	leafIndex, txnId, confirmationError = avm.SendDepositToNetwork(depositData.Txns, signedTxnBytes)
	if confirmationError != nil {
		switch confirmationError.Type {
		case avm.ErrRejected:
			log.Printf("Deposit transaction rejected: %v", confirmationError.Error())
			http.Error(w, "Your deposit was rejected. Please try again.",
				http.StatusUnprocessableEntity)
			return
		case avm.ErrOverSpend:
			log.Printf("Deposit transaction overspent: %v", confirmationError.Error())
			http.Error(w, "You do not have enough funds in your wallet for this deposit",
				http.StatusUnprocessableEntity)
			return
		case avm.ErrWaitTimeout:
			log.Printf("Deposit transaction timed out: %v", confirmationError.Error())
			http.Error(w, "Your deposit has not been confirmed by the blockchain yet.<br>"+
				"Please wait a few minutes and check your wallet to see if the deposit was sent."+
				"<br>If not, please try again.",
				http.StatusRequestTimeout)
			return
		case avm.ErrInternal:
			log.Printf("Internal error sending deposit transaction: %v",
				confirmationError.Error())
			http.Error(w, "Something went wrong. Your deposit was not processed."+
				"<br>Please try again.",
				http.StatusInternalServerError)
			return
		}
	}

	depositData.Note.LeafIndex = int(leafIndex)
	if txnId != depositData.Note.TxnID {
		log.Printf("Deposit txnId mismatch. %v != %v", txnId, depositData.Note.TxnID)
	}

	successHtml := `<dialog class="modal">
			   <h1>Deposit successful</h1>
				 <p>
				   You can use your new secret note to withdraw your funds in the future.
				</p>
				<button hx-get="withdraw"
						onclick="this.parentElement.close()">
				  Close
				</button>
			 </dialog>
			 <script>
			   document.querySelectorAll('dialog')[0].showModal()
			 </script>
			`
	fmt.Fprint(w, successHtml)

	saveNoteToDbError = db.SaveNote(depositData.Note)
	if saveNoteToDbError != nil {
		log.Printf("Error saving deposit to db: %v", saveNoteToDbError)
	}
}
