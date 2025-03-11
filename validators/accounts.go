package validators

type CreateAccountForm struct {
	AccountNo      string  `json:"-"`
	UserId         int     `json:"-"`
	AccountType    string  `json:"account_type" validate:"account_type"`
	InitialDeposit float64 `json:"initial_deposit" validate:"min=100"`
}
