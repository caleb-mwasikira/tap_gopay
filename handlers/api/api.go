package api

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response[T any] struct {
	Message string            `json:"message"`
	Data    T                 `json:"data"`
	Errs    map[string]string `json:"errors,omitempty"`
}

func SendResponse(w http.ResponseWriter, message string, data any, errs map[string]string, code int) {
	w.WriteHeader(code)

	resp := Response[any]{
		Message: message,
		Data:    data,
		Errs:    errs,
	}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		Error(
			w,
			"Error marshalling JSON data",
			err,
			http.StatusInternalServerError,
		)
		return
	}
}

func Error(w http.ResponseWriter, err_msg string, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	log.Printf("[%v] %v; %v\n", code, err_msg, err)

	resp := Response[any]{
		Message: err_msg,
		Data:    nil,
	}
	json.NewEncoder(w).Encode(resp)
}
