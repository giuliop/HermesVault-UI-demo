package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/algorand/go-algorand-sdk/types"
)

var (
	tmplMain              *template.Template
	tmplDeposit           *template.Template
	tmplWithdraw          *template.Template
	tmplConfirmDeposit    *template.Template
	tmplConfirmWithdrawal *template.Template
)

const (
	// TODO: Set a cache control header where useful
	cacheControl = "public, max-age=600" // 600 sec = 10 min
)

type Input string
type Address struct {
	Native types.Address
	Start  string
	Middle string
	End    string
}
type Amount struct {
	Algostring string
	Microalgos uint64
}
type Note struct {
	Amount uint64
	K      [31]byte
	R      [31]byte
	Text   string
}
type WithdrawData struct {
	Amount  *Amount
	Fee     *Amount
	Address *Address
	Note    *Note
	NewNote *Note
}
type DepositData struct {
	Amount  *Amount
	Address *Address
	NewNote *Note
}

func init() {
	// Helper function to create a map for passing multiple values to templates
	funcMap := template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"safeHTMLAttr": func(s string) template.HTMLAttr {
			return template.HTMLAttr(s)
		},
	}
	tmpl := template.Must(template.New("main").Funcs(funcMap).ParseFiles(
		"main.html", "confirm.html",
	))
	tmplMain = tmpl.Lookup("main")
	tmplDeposit = tmpl.Lookup("depositForm")
	tmplWithdraw = tmpl.Lookup("withdrawForm")
	tmplConfirmWithdrawal = tmpl.Lookup("confirmWithdrawal")
	tmplConfirmDeposit = tmpl.Lookup("confirmDeposit")
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := tmplMain.Execute(w, nil); err != nil {
			log.Printf("Error executing main template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/deposit", func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Cache-Control", cacheControl)
		switch r.Method {
		case http.MethodGet:
			if err := tmplDeposit.Execute(w, nil); err != nil {
				log.Printf("Error executing deposit template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				log.Printf("Error parsing form: %v", err)
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			amount, errAmount := Input(r.FormValue("amount")).toAmount()
			address, errAddress := Input(r.FormValue("address")).toAddress()
			errorMsg := ""
			if errAmount != nil {
				log.Printf("Error parsing withdrawal amount: %v", errAmount)
				errorMsg += "Please submit a valid algo amount<br>"
			}
			if errAddress != nil {
				log.Printf("Error parsing withdrawal address: %v", errAddress)
				errorMsg += "Please submit a valid Algorand address<br>"
			}
			if errorMsg != "" {
				http.Error(w, errorMsg, http.StatusUnprocessableEntity)
				return
			}
			newNote, err := generateNote(amount.Microalgos)
			if err != nil {
				log.Printf("Error generating new note: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			data := DepositData{
				Amount:  amount,
				Address: address,
				NewNote: newNote,
			}

			if err := tmplConfirmDeposit.Execute(w, data); err != nil {
				log.Printf("Error executing success template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/withdraw", func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Cache-Control", cacheControl)
		if false {
			amount, _ := Input("100").toAmount()
			address, _ := Input("LXUTB24U5OES3D4ZOOQKYY6DISYD7TIYCT5XXJKC36HUT4XLLMGVCORKNM").toAddress()
			note, _ := Input(strings.Repeat("aa", 70)).toNote()
			newNote, _ := generateChangeNote(amount, note)

			data := WithdrawData{
				Amount:  amount,
				Fee:     amount.Fee(),
				Address: address,
				Note:    note,
				NewNote: newNote,
			}
			if err := tmplConfirmWithdrawal.Execute(w, data); err != nil {
				log.Printf("Error executing success template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		switch r.Method {
		case http.MethodGet:
			if err := tmplWithdraw.Execute(w, nil); err != nil {
				log.Printf("Error executing withdraw template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				log.Printf("Error parsing form: %v", err)
				http.Error(w, "Bad Request", http.StatusBadRequest)
			}
			amount, errAmount := Input(r.FormValue("amount")).toAmount()
			address, errAddress := Input(r.FormValue("address")).toAddress()
			note, errNote := Input(r.FormValue("note")).toNote()
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
			isValid := verifyWithdrawal(amount, address, note)
			if isValid {
				newNote, err := generateChangeNote(amount, note)
				if err != nil {
					log.Printf("Error generating new note: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				data := WithdrawData{
					Amount:  amount,
					Fee:     amount.Fee(),
					Address: address,
					Note:    note,
					NewNote: newNote,
				}
				if err := tmplConfirmWithdrawal.Execute(w, data); err != nil {
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
	})

	http.HandleFunc("/confirm-deposit", func(w http.ResponseWriter, r *http.Request) {
		// read the form data and log it
		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Bad Request. Your deposit was not processed.",
				http.StatusBadRequest)
			return
		}
		// TODO: send the deposit transaction to the Algorand network
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
	})

	http.HandleFunc("/confirm-withdraw", func(w http.ResponseWriter, r *http.Request) {
		// read the form data and log it
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
	})

	// Serve static files from the "static" directory
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Load the dev ssl certificates
	cert, err := tls.LoadX509KeyPair("dev-ssl-certificates/localhost+4.pem",
		"dev-ssl-certificates/localhost+4-key.pem")
	if err != nil {
		log.Fatalf("Error loading certificates: %v", err)
	}

	// Create a custom HTTPS server
	server := &http.Server{
		Addr: ":3000",
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		fmt.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v\n", err)
		}
	}()

	fmt.Println("Server is running on https://localhost:3000 and https://maya.local:3000")
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatalf("Error starting HTTPS server: %v", err)
	}
}
