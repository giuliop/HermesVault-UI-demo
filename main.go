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
	"syscall"
	"time"
)

var (
	tmplMain              *template.Template
	tmplDeposit           *template.Template
	tmplWithdraw          *template.Template
	tmplConfirmWithdrawal *template.Template
)

const (
	// TODO: Set a cache control header where useful
	cacheControl = "public, max-age=600" // 600 sec = 10 min
)

type Input string

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
	}
	tmpl := template.Must(template.New("main").Funcs(funcMap).ParseFiles(
		"main.html", "confirm.html",
	))
	tmplMain = tmpl.Lookup("main")
	tmplDeposit = tmpl.Lookup("depositForm")
	tmplWithdraw = tmpl.Lookup("withdrawForm")
	tmplConfirmWithdrawal = tmpl.Lookup("confirmWithdrawal")
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
			amount, errAmount := Input(r.FormValue("amount")).toMicroAlgo()
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
			fmt.Fprintf(w, "You want to deposit %d microalgos from %s\n",
				amount, address)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/withdraw", func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Cache-Control", cacheControl)
		if true {
			data := map[string]interface{}{
				"Amount":           "100",
				"Fee":              "1",
				"AddressFirstFive": "LXUTB",
				"AddressMiddle":    "24U5OES3D4ZOOQKYY6DISYD7TIYCT5XXJKC36HUT4XLLMGVC",
				"AddressLastFive":  "QJQ3A",
				"Note":             "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				"NewSecretNote":    "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
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
			amount, errAmount := Input(r.FormValue("amount")).toMicroAlgo()
			address, errAddress := Input(r.FormValue("address")).toAddress()
			noteK, noteR, errNote := Input(r.FormValue("note")).toSecretNote()
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
			// if errorMsg != "" {
			// 	http.Error(w, errorMsg, http.StatusUnprocessableEntity)
			// 	return
			// }
			isValid := verifyWithdrawal(amount, address, noteK, noteR)
			if isValid {
				data := map[string]interface{}{
					"Amount":  Input(r.FormValue("amount")),
					"Address": Input(r.FormValue("address")),
					"Note":    Input(r.FormValue("note")),
					// TODO: Replace with actual logic for new note
					"NewSecretNote": "generatedNewSecretNote",
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
