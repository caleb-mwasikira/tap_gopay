package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/go-sql-driver/mysql"
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

func Error(w http.ResponseWriter, errMsg string, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	log.Printf("[%v] %v; %v\n", code, errMsg, err)

	// check if the error is a MySQL error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {

		// handle user-defined error (1644, which comes from SIGNAL SQLSTATE '45000')
		// https://fromdual.com/mysql-error-codes-and-messages-1600-1649
		if mysqlErr.Number == 1644 {
			fields := strings.SplitAfter(mysqlErr.Message, ":")

			if len(fields) > 0 {
				errMsg = strings.TrimSpace(fields[len(fields)-1])
			}
		}
	}

	resp := Response[any]{
		Message: errMsg,
		Data:    nil,
	}
	json.NewEncoder(w).Encode(resp)
}
