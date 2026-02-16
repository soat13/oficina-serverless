package apigw

type TokenRequest struct {
	CPF      string `json:"cpf"`
	Password string `json:"password"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type VerifyRequest struct {
	Token string `json:"token"`
}

type VerifyResponse struct {
	Subject string `json:"subject"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
