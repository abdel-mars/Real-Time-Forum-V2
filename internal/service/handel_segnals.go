package service

import (
	db "forum/internal/db"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func HandleSignals(server *http.Server) {
	
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-signalChannel
	fmt.Printf("Received signal: %s. Shutting down gracefully...\n", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	db.CloseDB()
	fmt.Println("Server exited gracefully.")
} 