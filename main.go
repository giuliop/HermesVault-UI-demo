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
	tmplMain     *template.Template
	tmplDeposit  *template.Template
	tmplWithdraw *template.Template
)

const (
	// TODO: Set a cache control header where useful
	cacheControl = "public, max-age=600" // 600 sec = 10 min
)

type Input string

func init() {
	tmpl := template.Must(template.New("main").ParseFiles("main.html"))

	tmplMain = tmpl.Lookup("main")
	tmplDeposit = tmpl.Lookup("deposit")
	tmplWithdraw = tmpl.Lookup("withdraw")
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
				errorMsg += "Please submit a valid Algorand address<br>"
			}
			if errNote != nil {
				log.Printf("Error parsing withdrawal note: %v", errNote)
				errorMsg += "Please submit a valid secret note"
			}
			if errorMsg != "" {
				http.Error(w, errorMsg, http.StatusUnprocessableEntity)
				return
			}
			log.Printf("Withdrawal request: %d microalgos to %s with note (%x, %x)\n", amount, address, noteK, noteR)
			fmt.Fprint(w, "Success !")
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
