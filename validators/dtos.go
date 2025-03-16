package validators

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"time"

	"github.com/caleb-mwasikira/tap_gopay/handlers/api"
)

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

type EmailDto struct {
	Email string `json:"email" validate:"email"`
}

type VerifyEmailDto struct {
	Email string `json:"email" validate:"email"`
	Otp   string `json:"otp" validate:"required,min=4"`
}

type ResetPasswordDto struct {
	PasswordResetToken string `json:"password_reset_token" validate:"min=6"`
	Email              string `json:"email" validate:"email"`
	NewPassword        string `json:"new_password" validate:"password"`
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

// Gets valid JSON input from request body.
// Supports both JSON objects and arrays as input.
// Validates each object in case of an array.
// All errors that occur parsing JSON input are written to the response body.
func GetValidJsonInput[T any](w http.ResponseWriter, body io.ReadCloser) (T, bool) {
	var raw json.RawMessage
	var errValue T

	// read the raw JSON input first
	err := json.NewDecoder(body).Decode(&raw)
	if err != nil {
		api.Error(
			w,
			"Invalid JSON data provided as input",
			err,
			http.StatusBadRequest,
		)
		return errValue, false
	}

	// try to unmarshal into the expected type
	var obj T
	err = json.Unmarshal(raw, &obj)
	if err != nil {
		api.Error(
			w,
			"Invalid JSON structure",
			err,
			http.StatusBadRequest,
		)
		return errValue, false
	}

	// validate JSON input
	validationErrors := map[string]string{}

	// use reflection to check if T is an array/slice
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Interface()
			errs := validateStruct(item)
			if len(errs) != 0 {
				validationErrors = errs
				break
			}
		}

	} else {
		errs := validateStruct(obj)
		if len(errs) != 0 {
			validationErrors = errs
		}
	}

	if len(validationErrors) > 0 {
		api.SendResponse(
			w,
			"Validation errors",
			nil,
			validationErrors,
			http.StatusBadRequest,
		)
		return errValue, false
	}

	return obj, true
}
