package main

import (
	"log"
	"net/http"
	"time"

	h "github.com/caleb-mwasikira/banking/handlers"
	"github.com/caleb-mwasikira/banking/utils"
)

// responseWriter is a wrapper to capture the response status
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func init() {
	utils.LoadEnvVariables()
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		log.Printf("[%d] %s %s - %v\n", rw.statusCode, r.Method, r.URL.Path, duration)
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /signup", h.HandleSignUp)
	mux.HandleFunc("POST /login", h.HandleLogin)
	mux.Handle("GET /accounts", h.AuthMiddleware(
		http.HandlerFunc(h.GetAllBankAccounts),
	))
	mux.Handle("GET /accounts/{acc_no}", h.AuthMiddleware(
		http.HandlerFunc(h.GetBankAccount),
	))
	mux.Handle("POST /accounts", h.AuthMiddleware(
		http.HandlerFunc(h.CreateBankAccount),
	))
	mux.Handle("GET /accounts{acc_no}", h.AuthMiddleware(
		http.HandlerFunc(h.DeleteBankAccount),
	))

	loggedMux := LoggingMiddleware(mux)

	address := "localhost:8080"
	log.Printf("starting HTTP server on %v\n", address)

	err := http.ListenAndServe(address, loggedMux)
	if err != nil {
		log.Fatalf("error starting HTTP server; %v", err)
	}
}
