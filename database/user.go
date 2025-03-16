package database

import (
	"database/sql"
	"fmt"
	"strings"

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

func UpdateUser(email string, updateValues map[string]interface{}) error {
	placeholders := []string{}
	values := []any{}

	for key, value := range updateValues {
		placeholders = append(placeholders, fmt.Sprintf("%s = ?", key))
		values = append(values, value)
	}

	values = append(values, email)

	query := fmt.Sprintf("UPDATE users SET %s WHERE email = ?", strings.Join(placeholders, ", "))
	_, err := db.Exec(query, values...)
	return err
}
