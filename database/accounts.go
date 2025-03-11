package database

import (
	"fmt"
	"log"

	"github.com/caleb-mwasikira/banking/validators"
)

type Account struct {
	Id             int     `json:"-"`
	UserId         int     `json:"-"`
	AccountNo      string  `json:"account_no"`
	AccountType    string  `json:"account_type"`
	InitialDeposit float64 `json:"initial_deposit"`
}

type AccountDetails struct {
	Account
	FirstName      string  `json:"firstname"`
	LastName       string  `json:"lastname"`
	Email          string  `json:"email"`
	CurrentBalance float64 `json:"current_balance"`
}

func GetAllAccounts() ([]AccountDetails, error) {
	query := `
		SELECT a.id, a.account_no, a.account_type, a.initial_deposit,
		u.firstname, u.lastname, u.email
		FROM accounts a
		INNER JOIN account_balances ab ON a.user_id = ab.user_id
		INNER JOIN users u ON a.user_id = u.id
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	accounts := []AccountDetails{}
	acc_details := AccountDetails{}

	for rows.Next() {
		err := rows.Scan(
			&acc_details.Id,
			&acc_details.AccountNo,
			&acc_details.AccountType,
			&acc_details.InitialDeposit,
			&acc_details.FirstName,
			&acc_details.LastName,
			&acc_details.Email,
		)
		if err != nil {
			log.Printf("error scanning row; %v\n", err)
			continue
		}

		accounts = append(accounts, acc_details)
	}

	return accounts, nil
}

func GetAccountByAccNo(acc_no string) (*AccountDetails, error) {
	query := `
		SELECT a.id, a.account_no, a.account_type, a.initial_deposit,
		u.firstname, u.lastname, u.email,
		ab.current_balance
		FROM accounts a
		INNER JOIN account_balances ab ON a.user_id = ab.user_id
		INNER JOIN users u ON a.user_id = u.id
		WHERE a.account_no = ?
	`

	acc_details := AccountDetails{}

	row := db.QueryRow(query, acc_no)
	err := row.Scan(
		&acc_details.Id,
		&acc_details.AccountNo,
		&acc_details.AccountType,
		&acc_details.InitialDeposit,
		&acc_details.FirstName,
		&acc_details.LastName,
		&acc_details.Email,
		&acc_details.CurrentBalance,
	)
	if err != nil {
		return nil, err
	}

	return &acc_details, nil
}

func CreateAccount(acc validators.CreateAccountForm) error {
	query := `
		INSERT INTO accounts (account_no,user_id,account_type,initial_deposit)
		VALUES(?, ?, ?, ?)
	`

	_, err := db.Exec(query, acc.AccountNo, acc.UserId, acc.AccountType, acc.InitialDeposit)
	if err != nil {
		return err
	}

	return nil
}

func GetAccount(id int, accountNo string, isActive bool) (*AccountDetails, error) {
	query := `
		SELECT a.id, a.account_no, a.account_type, a.initial_deposit,
		u.firstname, u.lastname, u.email,
		ab.current_balance
		FROM accounts a
		INNER JOIN account_balances ab ON a.user_id = ab.user_id
		INNER JOIN users u ON a.user_id = u.id
		WHERE u.id = ? AND a.account_no = ? AND a.is_active = ?
	`
	acc_details := AccountDetails{}

	row := db.QueryRow(query, id, accountNo, isActive)
	err := row.Scan(
		&acc_details.Id,
		&acc_details.AccountNo,
		&acc_details.AccountType,
		&acc_details.InitialDeposit,
		&acc_details.FirstName,
		&acc_details.LastName,
		&acc_details.Email,
		&acc_details.CurrentBalance,
	)
	if err != nil {
		return nil, err
	}

	return &acc_details, nil
}

func GetAccountByUserId(id int) (*AccountDetails, error) {
	query := `
		SELECT a.id, a.account_no, a.account_type, a.initial_deposit,
		u.firstname, u.lastname, u.email,
		ab.current_balance
		FROM accounts a
		INNER JOIN account_balances ab ON a.user_id = ab.user_id
		INNER JOIN users u ON a.user_id = u.id
		WHERE u.id = ?
	`
	acc_details := AccountDetails{}

	row := db.QueryRow(query, id)
	err := row.Scan(
		&acc_details.Id,
		&acc_details.AccountNo,
		&acc_details.AccountType,
		&acc_details.InitialDeposit,
		&acc_details.FirstName,
		&acc_details.LastName,
		&acc_details.Email,
		&acc_details.CurrentBalance,
	)
	if err != nil {
		return nil, err
	}

	return &acc_details, nil
}

func DeleteAccount(accountNo string) error {
	query := "UPDATE accounts SET is_active=false WHERE account_no=?"

	result, err := db.Exec(query, accountNo)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account with account_no=%v NOT found", acccountNo)
	}

	return nil
}
