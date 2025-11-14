package email

type AddEmailInput struct {
	Email     string
	ListID    string
	Variables map[string]string
}

type Provider interface {
	AddEmailToList(input AddEmailInput) error
}
