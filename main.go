package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	tmplMain     *template.Template
	tmplDeposit  *template.Template
	tmplWithdraw *template.Template
)

const (
	cacheControl = "public, max-age=600"
)

func init() {
	tmpl := template.Must(template.New("main").ParseFiles("main.html"))

	tmplMain = tmpl.Lookup("main")
	tmplDeposit = tmpl.Lookup("depositForm")
	tmplWithdraw = tmpl.Lookup("withdrawForm")
}
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmplMain.Execute(w, nil)
	})

	http.HandleFunc("/deposit", func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Cache-Control", cacheControl)
		switch r.Method {
		case http.MethodGet:
			tmplDeposit.Execute(w, nil)
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			amountStr := r.FormValue("amount")
			amount, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				fmt.Fprintf(w, "Please submit a valid amount.")
			} else {
				fmt.Fprintf(w, "You want to deposit %f", amount)
			}
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/withdraw", func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Cache-Control", cacheControl)
		switch r.Method {
		case http.MethodGet:
			tmplWithdraw.Execute(w, nil)
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			amountStr := r.FormValue("amount")
			amount, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				fmt.Fprintf(w, "Please submit a valid amount.")
			} else {
				fmt.Fprintf(w, "You want to withdraw %f", amount)
			}
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Serve static files from the "static" directory
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Load the dev ssl certificates
	cert, err := tls.LoadX509KeyPair("dev-ssl-certificates/localhost+4.pem",
		"dev-ssl-certificates/localhost+4-key.pem")
	if err != nil {
		fmt.Println("Error loading ssl certificates:", err)
		return
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
			fmt.Printf("Server forced to shutdown: %v\n", err)
		}
	}()

	fmt.Println("Server is running on https://localhost:3000 and https://maya.local:3000")
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		fmt.Println("Error starting HTTPS server:", err)
	}
}
