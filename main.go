package main

import (
	"log"
	"net/http"
	"time"

	h "github.com/caleb-mwasikira/tap_gopay/handlers"
	"github.com/caleb-mwasikira/tap_gopay/utils"
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
	mux.HandleFunc("POST /send-verification-email", h.SendVerificationEmail)
	mux.HandleFunc("POST /verify-email", h.VerifyEmail)
	mux.HandleFunc("POST /request-password-reset", h.RequestPasswordReset)
	mux.HandleFunc("POST /reset-password", h.ResetPassword)

	mux.Handle("POST /new-account", h.AuthMiddleware(
		http.HandlerFunc(h.NewAccount),
	))
	mux.Handle("GET /my-accounts", h.AuthMiddleware(
		http.HandlerFunc(h.MyAccounts),
	))
	mux.Handle("POST /search-accounts", h.AuthMiddleware(
		http.HandlerFunc(h.SearchAccounts),
	))
	mux.Handle("POST /deactivate-account", h.AuthMiddleware(
		http.HandlerFunc(h.DeactivateCard),
	))

	loggedMux := LoggingMiddleware(mux)

	address := "localhost:8080"
	log.Printf("starting HTTP server on %v\n", address)

	err := http.ListenAndServe(address, loggedMux)
	if err != nil {
		log.Fatalf("error starting HTTP server; %v", err)
	}
}
