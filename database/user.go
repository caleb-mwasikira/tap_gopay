package database

import (
	"database/sql"

	"github.com/caleb-mwasikira/banking/validators"
)

type User struct {
	Id             int            `json:"id"`
	FirstName      string         `json:"firstname"`
	LastName       string         `json:"lastname"`
	Email          string         `json:"email"`
	Password       string         `json:"password"`
	Role           string         `json:"role"`
	PhoneNumber    sql.NullString `json:"phone_no"`
	ProfilePicture sql.NullString `json:"profile_pic"`
	IsActive       bool           `json:"is_active"`
}

func GetUser(email string) (*User, error) {
	row := db.QueryRow("SELECT * FROM users WHERE email = ?", email)

	dbUser := User{}
	err := row.Scan(
		&dbUser.Id,
		&dbUser.FirstName,
		&dbUser.LastName,
		&dbUser.Email,
		&dbUser.Password,
		&dbUser.Role,
		&dbUser.PhoneNumber,
		&dbUser.ProfilePicture,
		&dbUser.IsActive,
	)
	if err != nil {
		return nil, err
	}

	return &dbUser, nil
}

func CreateUser(user validators.RegisterForm) error {
	_, err := db.Exec(
		"INSERT INTO users(firstname, lastname, email, password) VALUES(?, ?, ?, ?)",
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password,
	)

	return err
}
