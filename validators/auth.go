package validators

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	MIN_NAME_LEN     int = 3
	MIN_PASSWORD_LEN int = 8
)

type LoginForm struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"password"`
}

type RegisterForm struct {
	FirstName   string `json:"firstname" validate:"required,min=3,max=255"`
	LastName    string `json:"lastname" validate:"required,min=3,max=255"`
	Email       string `json:"email" validate:"email"`
	Password    string `json:"password" validate:"password"`
	PhoneNumber string `json:"phone_no,omitempty"`
}

func validateEmail(email string) error {
	if strings.Trim(email, " ") == "" {
		return fmt.Errorf("email field is required")
	}

	regex, err := regexp.Compile("^[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+[.][A-Za-z.]{2,}$")
	if err != nil {
		return err
	}

	matched := regex.Match([]byte(email))
	if !matched {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func validatePassword(password string, verifyPasswordStrength bool) error {
	if strings.Trim(password, " ") == "" {
		return fmt.Errorf("password field is required")
	}

	if len(password) < MIN_PASSWORD_LEN {
		return fmt.Errorf("password cannot be less than %v characters long", MIN_PASSWORD_LEN)
	}

	if verifyPasswordStrength {
		if !containsSpecialChar(password) {
			return fmt.Errorf("password must contain at least one special character")
		}

		if !containsUpperAndLower(password) {
			return fmt.Errorf("password must contains at least one uppercase and lowercase letter")
		}
	}

	return nil
}

func containsSpecialChar(str string) bool {
	re := regexp.MustCompile(`[^a-zA-Z0-9]`) // matches any character that is not a lowercase or uppercase letter and is not a number
	return re.MatchString(str)
}

func containsUpperAndLower(str string) bool {
	reUpper := regexp.MustCompile(`[A-Z]`) // matcher any uppercase letter
	reLower := regexp.MustCompile(`[a-z]`) // matcher any lowercase letter

	return reUpper.MatchString(str) && reLower.MatchString(str)
}
