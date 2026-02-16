package persistence

type credentialRow struct {
	ID           string `bun:"id,pk"`
	CPF          string `bun:"cpf"`
	PasswordHash string `bun:"password_hash"`
}
