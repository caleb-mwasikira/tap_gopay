package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"

	db "github.com/caleb-mwasikira/tap_gopay/database"
	"github.com/caleb-mwasikira/tap_gopay/handlers/api"
	v "github.com/caleb-mwasikira/tap_gopay/validators"
)

const (
	CREDIT_CARD_NO_LEN int = 14
	CVV_LEN            int = 4

	ErrInvalidJsonInput string = "Invalid JSON data provided as input"
)

func generateCardNo(len int) string {
	if len <= 0 {
		log.Println("invalid card number length")
		return ""
	}

	nums := []string{}

	for i := 0; i < len; i++ {
		bigNum, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			log.Printf("error generating random number; %v\n", err)
			return ""
		}
		nums = append(nums, fmt.Sprintf("%v", bigNum.Int64()))
	}

	return strings.Join(nums, "")
}

func NewAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	user := getLoggedInUser(r.Context())
	if user == nil {
		api.Error(
			w,
			"Unauthorized action detected",
			nil,
			http.StatusUnauthorized,
		)
		return
	}

	newCreditCard := v.CreditCardDto{}
	err := json.NewDecoder(r.Body).Decode(&newCreditCard)
	if err != nil {
		api.Error(
			w,
			ErrInvalidJsonInput,
			err,
			http.StatusBadRequest,
		)
		return
	}

	errs := v.ValidateStruct(newCreditCard)
	if len(errs) != 0 {
		api.SendResponse(
			w,
			"Validation errors",
			nil,
			errs,
			http.StatusBadRequest,
		)
		return
	}

	newCardNo := generateCardNo(CREDIT_CARD_NO_LEN)
	newCvv := generateCardNo(CVV_LEN)
	if newCardNo == "" || newCvv == "" {
		api.Error(
			w,
			"Unexpected error generating new credit card",
			err,
			http.StatusInternalServerError,
		)
		return
	}

	newCreditCard.UserId = user.Id
	newCreditCard.CardNo = newCardNo
	newCreditCard.Cvv = newCvv

	err = db.CreateCreditCard(newCreditCard)
	if err != nil {
		api.Error(
			w,
			"Unexpected error generating new credit card",
			err,
			http.StatusInternalServerError,
		)
		return
	}

	api.SendResponse(
		w,
		"Credit card created successfully",
		newCreditCard,
		nil,
		http.StatusCreated,
	)
}

func SearchAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	contacts := []v.ContactDto{}
	err := json.NewDecoder(r.Body).Decode(&contacts)
	if err != nil {
		api.Error(
			w,
			ErrInvalidJsonInput,
			err,
			http.StatusBadRequest,
		)
		return
	}

	phoneNos := make([]string, len(contacts))

	for index, userContact := range contacts {
		errs := v.ValidateStruct(userContact)
		if len(errs) != 0 {
			api.SendResponse(
				w,
				fmt.Sprintf("Validation errors at index %v", index),
				nil,
				errs,
				http.StatusBadRequest,
			)
			return
		}

		phoneNos[index] = userContact.PhoneNo
	}

	var creditCards []v.CreditCardDto = []v.CreditCardDto{}

	creditCards, err = db.GetCreditCardsAssocWith(phoneNos)
	if err != nil {
		if err == sql.ErrNoRows {
			api.SendResponse(
				w,
				"No credit cards found",
				nil,
				nil,
				http.StatusNoContent,
			)
			return
		}

		api.Error(
			w,
			"Unexpected error searching for credit cards",
			err,
			http.StatusInternalServerError,
		)
		return
	}

	api.SendResponse(
		w,
		"Credit cards found",
		creditCards,
		nil,
		http.StatusOK,
	)
}

func MyAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	user := getLoggedInUser(r.Context())
	if user == nil {
		api.Error(
			w,
			"Unauthorized action detected",
			nil,
			http.StatusUnauthorized,
		)
		return
	}

	dbCreditCards, err := db.GetCreditCardsFor(user.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			api.SendResponse(
				w,
				"No credit cards found under your name",
				nil, nil,
				http.StatusNoContent,
			)
			return
		}

		api.Error(
			w,
			"Unexpected error fetching your credit card accounts",
			err,
			http.StatusInternalServerError,
		)
		return
	}

	log.Println(dbCreditCards)

	api.SendResponse(
		w,
		"Success fetching your accounts",
		dbCreditCards, nil,
		http.StatusOK,
	)
}

func DeactivateCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	user := getLoggedInUser(r.Context())
	if user == nil {
		api.Error(
			w,
			"Unauthorized action detected",
			nil,
			http.StatusUnauthorized,
		)
		return
	}

	card := v.CardNoDto{}
	err := json.NewDecoder(r.Body).Decode(&card)
	if err != nil {
		api.Error(
			w,
			ErrInvalidJsonInput,
			err,
			http.StatusBadRequest,
		)
		return
	}

	_, err = db.GetCreditCardWhere(user.Username, card.CardNo, true)
	if err != nil {
		if err == sql.ErrNoRows {
			api.SendResponse(
				w,
				fmt.Sprintf("No credit card with account number %v found under your name", card.CardNo),
				nil, nil,
				http.StatusNoContent,
			)
			return
		}

		api.Error(
			w,
			"Unexpected error freezing credit card",
			err,
			http.StatusInternalServerError,
		)
		return
	}

	err = db.DeactivateCard(card.CardNo)
	if err != nil {
		api.Error(
			w,
			"Unexpected error freezing card",
			err,
			http.StatusInternalServerError,
		)
		return
	}

	api.SendResponse(
		w,
		fmt.Sprintf("Success freezing card %v", card.CardNo),
		nil,
		nil,
		http.StatusOK,
	)
}
