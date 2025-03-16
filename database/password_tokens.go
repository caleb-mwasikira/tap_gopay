package database

import (
	"fmt"
	"log"
	"time"

	"github.com/caleb-mwasikira/tap_gopay/utils"
)

type PasswordResetToken struct {
	Id        int       `json:"id"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

const (
	PASSWORD_TOKEN_LEN int = 6
)

func GenerateAndSavePasswordToken(email string) (string, error) {
	token := utils.RandNumbers(PASSWORD_TOKEN_LEN)
	if token == "" {
		return "", fmt.Errorf("error generating password-reset token")
	}

	query := "INSERT INTO password_reset_tokens(email, token, expires_at) VALUES(?, ?, ?)"
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err := db.Exec(query, email, token, expiresAt)
	return token, err
}

func GetPasswordResetToken(email, token string) (*PasswordResetToken, error) {
	query := `
		SELECT id, email, token, created_at, expires_at FROM password_reset_tokens
		WHERE email = ? 
		AND token = ?
		AND expires_at > NOW()
		LIMIT 1
	`

	passwordResetToken := PasswordResetToken{}
	row := db.QueryRow(query, email, token)

	err := row.Scan(
		&passwordResetToken.Id,
		&passwordResetToken.Email,
		&passwordResetToken.Token,
		&passwordResetToken.CreatedAt,
		&passwordResetToken.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	// password-reset-tokens are one time use only
	go func(id int) {
		query := "DELETE FROM password_reset_tokens WHERE id = ?"

		_, err := db.Exec(query, id)
		if err != nil {
			log.Printf("error deleting used password reset token; %v\n", err)
		}
	}(passwordResetToken.Id)

	return &passwordResetToken, nil
}
