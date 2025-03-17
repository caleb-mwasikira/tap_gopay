package database

import (
	"fmt"
	"strings"

	v "github.com/caleb-mwasikira/tap_gopay/validators"
)

type CreditCardDetails struct {
	v.CreditCardDto
	Username       string  `json:"username,omitempty"`
	Email          string  `json:"email,omitempty"`
	CurrentBalance float64 `json:"current_balance"`
}

func CreateCreditCard(newCreditCard v.CreditCardDto) error {
	query := `
		INSERT INTO credit_cards(user_id, card_no, cvv, initial_deposit)
		VALUES(?, ?, ?, ?)
	`

	_, err := db.Exec(
		query,
		newCreditCard.UserId,
		newCreditCard.CardNo,
		newCreditCard.Cvv,
		newCreditCard.InitialDeposit,
	)
	if err != nil {
		return err
	}

	return nil
}

func GetCreditCardsAssocWith(phoneNos []string) ([]v.CreditCardDto, error) {
	if len(phoneNos) == 0 {
		return nil, fmt.Errorf("empty search parameter phone numbers")
	}

	// generate placeholders (?, ?, ?)
	placeholders := ""
	args := make([]interface{}, len(phoneNos))

	for i, phoneNo := range phoneNos {
		placeholders += "?,"
		args[i] = phoneNo
	}
	placeholders = strings.TrimSuffix(placeholders, ",")

	query := fmt.Sprintf(
		`
			SELECT cc.id, cc.user_id, cc.card_no
			FROM credit_cards cc
			INNER JOIN users u ON cc.user_id = u.id
			WHERE u.phone_no IN (%s) AND cc.is_active = TRUE
		`, placeholders,
	)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	creditCards := []v.CreditCardDto{}
	creditCard := v.CreditCardDto{}

	for rows.Next() {
		err = rows.Scan(
			&creditCard.Id,
			&creditCard.UserId,
			&creditCard.CardNo,
		)
		if err != nil {
			return nil, err
		}

		creditCards = append(creditCards, creditCard)
	}

	return creditCards, nil
}

func GetCreditCardsFor(username string) ([]CreditCardDetails, error) {
	query := `
		SELECT cc.id, cc.card_no, cc.is_active, cc.created_at,
		u.username, u.email,
		b.balance
		FROM credit_cards cc
		INNER JOIN users u ON u.id = cc.user_id
		INNER JOIN balances b ON b.card_no = cc.card_no
		WHERE username = ?
	`

	rows, err := db.Query(query, username)
	if err != nil {
		return nil, err
	}

	creditCards := []CreditCardDetails{}
	creditCard := CreditCardDetails{}

	for rows.Next() {
		err = rows.Scan(
			&creditCard.Id,
			&creditCard.CardNo,
			&creditCard.IsActive,
			&creditCard.CreatedAt,
			&creditCard.Username,
			&creditCard.Email,
			&creditCard.CurrentBalance,
		)
		if err != nil {
			return nil, err
		}

		creditCards = append(creditCards, creditCard)
	}

	return creditCards, nil
}

func GetCreditCardWhere(username, cardNo string, isActive bool) (*v.CreditCardDto, error) {
	if strings.TrimSpace(username) == "" || strings.TrimSpace(cardNo) == "" {
		return nil, fmt.Errorf("empty search parameter username or card_no")
	}

	query := `
		SELECT cc.id, cc.user_id, cc.card_no
		FROM credit_cards cc
		INNER JOIN users u ON cc.user_id = u.id
		WHERE u.username = ? 
		AND cc.card_no = ?
	`

	if isActive {
		query += "AND cc.is_active = TRUE"
	}

	row := db.QueryRow(query, username, cardNo)
	creditCard := v.CreditCardDto{}

	err := row.Scan(
		&creditCard.Id,
		&creditCard.UserId,
		&creditCard.CardNo,
	)
	if err != nil {
		return nil, err
	}

	return &creditCard, nil
}

func DeactivateCard(cardNo string) error {
	query := "UPDATE credit_cards SET is_active = FALSE WHERE card_no = ?"
	_, err := db.Exec(query, cardNo)
	return err
}

func CreateTransaction(transaction v.SendMoneyDto) error {
	query := "INSERT INTO transactions(senders_card, receivers_card, amount) VALUES(?, ?, ?)"

	_, err := db.Exec(
		query,
		transaction.SendersCard,
		transaction.ReceiversCard,
		transaction.Amount,
	)
	return err
}
