package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func ConfirmDepositHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Bad Request. Your deposit was not processed.",
			http.StatusBadRequest)
		return
	}
	// TODO: send the withdrawal transaction to the Algorand network
	time.Sleep(2 * time.Second)
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
	// TODO: send the withdrawal transaction to the Algorand network
	time.Sleep(2 * time.Second)
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
