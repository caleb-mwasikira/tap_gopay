package database

import (
	"fmt"
	"log"
	"time"

	"github.com/caleb-mwasikira/tap_gopay/utils"
)

type OtpRecord struct {
	Id        int       `json:"id"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

const (
	OTP_DIGIT_LEN int = 4
)

func GenerateAndSaveOtp(email string) (string, error) {
	otp := utils.RandNumbers(OTP_DIGIT_LEN)
	if otp == "" {
		return "", fmt.Errorf("error generating OTP code")
	}

	query := "INSERT INTO otps(email, code, expires_at) VALUES(?, ?, ?)"
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err := db.Exec(query, email, otp, expiresAt)
	return otp, err
}

func GetOtpRecord(email, otp string) (*OtpRecord, error) {
	query := `
		SELECT id, email, code, created_at, expires_at FROM otps
		WHERE email = ? 
		AND code = ?
		AND expires_at > NOW()
		LIMIT 1
	`

	otpRecord := OtpRecord{}
	row := db.QueryRow(query, email, otp)

	err := row.Scan(
		&otpRecord.Id,
		&otpRecord.Email,
		&otpRecord.Code,
		&otpRecord.CreatedAt,
		&otpRecord.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	// one-time-passwords are one time use only
	go func(id int) {
		query := "DELETE FROM otps WHERE id = ?"

		_, err := db.Exec(query, id)
		if err != nil {
			log.Printf("error deleting used OTP code; %v\n", err)
		}

	}(otpRecord.Id)

	return &otpRecord, nil
}
