package database

import (
	"database/sql"

	v "github.com/caleb-mwasikira/tap_gopay/validators"
)

type User struct {
	Id          int            `json:"id"`
	Username    string         `json:"username"`
	Email       string         `json:"email"`
	Password    string         `json:"password"`
	IsActive    bool           `json:"is_active"`
	PhoneNumber sql.NullString `json:"phone_no"`
}

func GetUser(email string) (*User, error) {
	row := db.QueryRow("SELECT * FROM users WHERE email = ?", email)

	dbUser := User{}
	err := row.Scan(
		&dbUser.Id,
		&dbUser.Username,
		&dbUser.Email,
		&dbUser.Password,
		&dbUser.IsActive,
		&dbUser.PhoneNumber,
	)
	if err != nil {
		return nil, err
	}

	return &dbUser, nil
}

func CreateUser(user v.RegisterDto) error {
	_, err := db.Exec(
		"INSERT INTO users(username, email, password, phone_no) VALUES(?, ?, ?, ?)",
		user.Username,
		user.Email,
		user.Password,
		user.PhoneNumber,
	)

	return err
}

func ActiveUserAccountWhere(email string) error {
	query := "UPDATE users SET is_active = TRUE WHERE email = ?"
	_, err := db.Exec(query, email)
	return err
}
