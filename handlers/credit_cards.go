package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	db "github.com/caleb-mwasikira/tap_gopay/database"
	"github.com/caleb-mwasikira/tap_gopay/handlers/api"
	"github.com/caleb-mwasikira/tap_gopay/utils"
	v "github.com/caleb-mwasikira/tap_gopay/validators"
)

const (
	CREDIT_CARD_NO_LEN int = 14
	CVV_LEN            int = 4
)

func NewCreditCard(w http.ResponseWriter, r *http.Request) {
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

	newCreditCard, ok := v.GetValidJsonInput[v.CreditCardDto](w, r.Body)
	if !ok {
		return
	}

	newCardNo := utils.RandNumbers(CREDIT_CARD_NO_LEN)
	newCvv := utils.RandNumbers(CVV_LEN)
	if newCardNo == "" || newCvv == "" {
		api.Error(
			w,
			"Unexpected error generating new credit card",
			fmt.Errorf("error generating card number or card cvv value"),
			http.StatusInternalServerError,
		)
		return
	}

	newCreditCard.UserId = user.Id
	newCreditCard.CardNo = newCardNo
	newCreditCard.Cvv = newCvv

	err := db.CreateCreditCard(newCreditCard)
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

func SearchCreditCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	contacts, ok := v.GetValidJsonInput[[]v.ContactDto](w, r.Body)
	if !ok {
		return
	}

	phoneNos := []string{}
	for _, contact := range contacts {
		phoneNos = append(phoneNos, contact.PhoneNo)
	}

	creditCards, err := db.GetCreditCardsAssocWith(phoneNos)
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

func MyCreditCards(w http.ResponseWriter, r *http.Request) {
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
			"Invalid JSON data provided as input",
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

func SendMoney(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	user := getLoggedInUser(r.Context())
	if user == nil {
		api.Error(
			w,
			"Unauthorized action detected. Please login and try again",
			nil,
			http.StatusUnauthorized,
		)
		return
	}

	request, ok := v.GetValidJsonInput[v.SendMoneyDto](w, r.Body)
	if !ok {
		return
	}

	// check if senders_card number belongs to the logged in user
	_, err := db.GetCreditCardWhere(user.Username, request.SendersCard, true)
	if err != nil {
		if err == sql.ErrNoRows {
			api.SendResponse(
				w,
				"Invalid or deactivated senders credit card",
				nil, nil,
				http.StatusBadRequest,
			)
			return
		}

		api.Error(
			w,
			"Unexpected error sending money across credit cards",
			err,
			http.StatusInternalServerError,
		)
		return
	}

	err = db.CreateTransaction(request)
	if err != nil {
		api.Error(
			w,
			fmt.Sprintf("Unexpected error sending money to %v", request.ReceiversCard),
			err,
			http.StatusInternalServerError,
		)
		return
	}

	api.SendResponse(
		w,
		fmt.Sprintf("KSH %.2f sent successfully to %v", request.Amount, request.ReceiversCard),
		nil, nil,
		http.StatusOK,
	)
}
