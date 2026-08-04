package main

//go:generate templ generate

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/tfharrelson/scicomp-bench/app/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	mux := app.CreateApp()

	port := ":8080"
	go func() {
		log.Printf("Server starting on %s", port)
		log.Fatal(http.ListenAndServe(port, mux))
	}()

	<-ctx.Done()
	log.Println("Shutting down server")
	//	serverShutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 30*time.Second)
	//	defer shutdownRelease()

	//	if err := mux.Shutdown(serverShutdownCtx); err != nil {
	//		log.Fatalf("Error shutting down server: %v", err)
	//	}

	log.Println("Server shutdown complete")
}
