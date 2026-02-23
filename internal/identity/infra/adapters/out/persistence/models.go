package persistence

import "github.com/uptrace/bun"

type credentialRow struct {
	bun.BaseModel `bun:"table:public.users"`
	ID            string `bun:"id,pk"`
	CPF           string `bun:"document"`
	Password      string `bun:"password"`
}
