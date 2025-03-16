package database

import "time"

type OtpRecord struct {
	Id        int       `json:"id"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func CreateOtpRecord(email, otp string) error {
	query := "INSERT INTO otps(email, code, expires_at) VALUES(?, ?, ?)"
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err := db.Exec(query, email, otp, expiresAt)
	return err
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

	return &otpRecord, nil
}
