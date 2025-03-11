package api

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func SendJSONResponse[T any](w http.ResponseWriter, resp Response[T]) {
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		Error(
			w,
			"Error marshalling JSON data",
			err,
			http.StatusBadRequest,
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
