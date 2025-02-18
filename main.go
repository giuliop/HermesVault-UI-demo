package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"webapp/config"
	"webapp/db"
	"webapp/frontend/templates"
	"webapp/handlers"
)

func main() {
	// Parse the -dev flag
	dev := flag.Bool("dev", false, "run in development mode")
	flag.Parse()

	defer db.Close()

	// Start periodic cleanup of internal database
	cleanupCancel := db.StartCleanupRoutine(context.Background(), config.CleanupInterval)
	defer cleanupCancel()

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
	http.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("./frontend/static/"))))

	var server *http.Server

	// Determine the mode and configure the server accordingly
	if *dev {
		// Development mode, we use a self-signed certificate to serve HTTPS
		cert, err := tls.LoadX509KeyPair("dev-ssl-certificates/localhost+4.pem",
			"dev-ssl-certificates/localhost+4-key.pem")
		if err != nil {
			log.Fatalf("Error loading certificates: %v", err)
		}

		// Create a custom HTTPS server
		server = &http.Server{
			Addr: ":" + config.DevelopmentPort,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		}

		fmt.Printf("Server running in development mode on port %s\n",
			config.DevelopmentPort)
		go func() {
			err = server.ListenAndServeTLS("", "")
			if err != nil && err != http.ErrServerClosed {
				log.Fatalf("Error starting HTTPS server: %v", err)
			}
		}()
	} else {
		// Production mode, we serve HTTP to a reverse proxy
		server = &http.Server{
			Addr: ":" + config.ProductionPort,
		}

		fmt.Printf("Server running in production mode on port %s\n",
			config.ProductionPort)
		go func() {
			err := server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Fatalf("Error starting HTTP server: %v", err)
			}
		}()
	}

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// Block the main goroutine until the server is shut down
	<-quit
	fmt.Print("\nShutting down server...\n\n")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}
}
