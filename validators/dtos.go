package validators

import "time"

type LoginDto struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"password"`
}

type RegisterDto struct {
	Username    string `json:"username" validate:"min=3,max=50"`
	Email       string `json:"email" validate:"email"`
	Password    string `json:"password" validate:"password"`
	PhoneNumber string `json:"phone_no" validate:"required,min=9"`
}

type CreditCardDto struct {
	Id             int       `json:"id"`
	UserId         int       `json:"user_id"`
	CardNo         string    `json:"card_no"`
	Cvv            string    `json:"-"`
	InitialDeposit float64   `json:"initial_deposit,omitempty" validate:"min=100"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

type ContactDto struct {
	Username string `json:"username"`
	PhoneNo  string `json:"phone_no" validate:"min=9"`
}

type CardNoDto struct {
	CardNo string `json:"card_no" validate:"min=10"`
}
