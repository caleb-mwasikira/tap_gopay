package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	db "github.com/caleb-mwasikira/banking/database"
	"github.com/caleb-mwasikira/banking/handlers/api"
	"github.com/caleb-mwasikira/banking/validators"
)

func GetAllBankAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	dbAccounts, err := db.GetAllAccounts()
	if err != nil {
		if err == sql.ErrNoRows {
			resp := api.Response[any]{
				Message: "Zero bank accounts stored in database",
				Data:    nil,
			}
			api.SendJSONResponse(w, resp)
			return
		}

		api.Error(
			w,
			"Unexpected error fetching bank accounts",
			err,
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusOK)

	resp := api.Response[[]db.AccountDetails]{
		Message: "Success fetching bank accounts",
		Data:    dbAccounts,
	}
	api.SendJSONResponse(w, resp)
}

func GetBankAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	accountNo := strings.TrimSpace(r.PathValue("acc_no"))
	if accountNo == "" {
		api.Error(
			w,
			"Missing account number path parameter in url",
			nil,
			http.StatusBadRequest,
		)
		return
	}

	accountDetails, err := db.GetAccountByAccNo(accountNo)
	if err != nil {
		if err == sql.ErrNoRows {
			resp := api.Response[any]{
				Message: fmt.Sprintf("Bank account with account number %v not found", accountNo),
				Data:    nil,
			}
			api.SendJSONResponse(w, resp)
			return
		}

		api.Error(
			w,
			"Unexpected error fetching bank account",
			err,
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusOK)

	resp := api.Response[*db.AccountDetails]{
		Message: "Successfully fetched bank account",
		Data:    accountDetails,
	}
	api.SendJSONResponse(w, resp)
}

func generateAccountNo() (string, error) {
	buff := make([]byte, 32)
	_, err := rand.Read(buff)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(buff), nil
}

func CreateBankAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	acc := validators.CreateAccountForm{}
	err := json.NewDecoder(r.Body).Decode(&acc)
	if err != nil {
		api.Error(
			w,
			"Invalid JSON data provided as input",
			err,
			http.StatusBadRequest,
		)
		return
	}

	errs := validators.ValidateStruct(acc)
	if len(errs) != 0 {
		w.WriteHeader(http.StatusBadRequest)

		resp := api.Response[map[string]string]{
			Message: "Validation errors",
			Data:    errs,
		}
		api.SendJSONResponse(w, resp)
		return
	}

	user := getLoggedInUser(r.Context())
	if user == nil {
		api.Error(
			w,
			"Unauthorized action detected",
			fmt.Errorf("error acquiring logged in user from context"),
			http.StatusInternalServerError,
		)
		return
	}

	// check if user already has account
	dbAccount, err := db.GetAccountByUserId(user.Id)
	if err != nil && err != sql.ErrNoRows {
		api.Error(
			w,
			"Unexpected error creating bank account",
			err,
			http.StatusInternalServerError,
		)
		return
	}

	if dbAccount != nil {
		api.Error(
			w,
			"User already has an account under their name",
			nil,
			http.StatusConflict,
		)
		return
	}

	accountNo, err := generateAccountNo()
	if err != nil {
		api.Error(
			w,
			"Unexpected error creating bank account",
			err,
			http.StatusInternalServerError,
		)
		return
	}

	acc.AccountNo = accountNo
	acc.UserId = user.Id

	err = db.CreateAccount(acc)
	if err != nil {
		api.Error(
			w,
			"Unexpected error creating bank account",
			err,
			http.StatusInternalServerError,
		)
		return
	}

	resp := api.Response[validators.CreateAccountForm]{
		Message: "Bank account created successfully",
		Data:    acc,
	}
	api.SendJSONResponse(w, resp)
}

func DeleteBankAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	accountNo := strings.TrimSpace(r.PathValue("acc_no"))
	if accountNo == "" {
		api.Error(
			w,
			"Missing account number path parameter in url",
			nil,
			http.StatusBadRequest,
		)
		return
	}

	user := getLoggedInUser(r.Context())
	if user == nil {
		api.Error(
			w,
			"Unauthorized action detected",
			fmt.Errorf("error extracting logged in user from request context"),
			http.StatusUnauthorized,
		)
		return
	}

	if user.Role == "user" { // ensure user is deleting their own bank account
		_, err := db.GetAccount(user.Id, accountNo, true)
		if err != nil {
			var (
				errMsg     string = "Unexpected error deleting bank account"
				statusCode int    = http.StatusInternalServerError
			)

			if err == sql.ErrNoRows {
				errMsg = fmt.Sprintf("Bank account %v does not belong to you", accountNo)
				statusCode = http.StatusUnauthorized
			}

			api.Error(w, errMsg, err, statusCode)
			return
		}
	}

	err := db.DeleteAccount(accountNo)
	if err != nil {
		api.Error(
			w,
			"Unexpected error deleting bank account",
			err,
			http.StatusInternalServerError,
		)
		return
	}

	resp := api.Response[any]{
		Message: fmt.Sprintf("Bank account %v deleted successfully", err),
		Data:    nil,
	}
	api.SendJSONResponse(w, resp)
}
