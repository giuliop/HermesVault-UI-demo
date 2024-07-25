package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"webapp/handlers"
	"webapp/templates"
)

func main() {
	templates.InitTemplates()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := templates.Main.Execute(w, nil); err != nil {
			log.Printf("Error executing main template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/deposit", handlers.DepositHandler)
	http.HandleFunc("/withdraw", handlers.WithdrawHandler)
	http.HandleFunc("/confirm-deposit", handlers.ConfirmDepositHandler)
	http.HandleFunc("/confirm-withdraw", handlers.ConfirmWithdrawHandler)

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
