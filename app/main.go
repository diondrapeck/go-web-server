package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
)

func newRouter() *httprouter.Router {
	mux := httprouter.New()
	mux.GET("/data", getResponse())
	return mux
}

func getResponse() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Write([]byte("response!"))
	}
}

func main() {
	port := "10101"

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: newRouter(),
	}

	idleConnectionClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)
		<-sigint

		log.Println("Service interrupt received")

		cntxt, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		if err := server.Shutdown(cntxt); err != nil {
			log.Printf("http server shutdown error: %v", err)
		}

		log.Println("Shutdown complete")

		close(idleConnectionClosed)
	}()

	log.Printf("Starting server on port %v", port)
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Fatal: http server failed to start: %v", err)
		}
	}

	<-idleConnectionClosed
	log.Println("Service stopped")
}
