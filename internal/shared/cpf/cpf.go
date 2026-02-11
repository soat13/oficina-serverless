package cpf

import (
	"errors"
	"strings"
	"unicode"
)

const (
	size = 11
)

var (
	ErrInvalidCPF = errors.New("invalid cpf")
)

type CPF struct {
	value string
}

func Parse(v string) (CPF, error) {
	n := onlyNumbers(strings.TrimSpace(v))

	if !isValid(n) {
		return CPF{}, ErrInvalidCPF
	}

	return CPF{value: n}, nil
}

func (c CPF) String() string {
	return c.value
}

func isValid(cpf string) bool {
	if len(cpf) != size || hasAllEqualDigits(cpf) {
		return false
	}

	firstCheckDigit := int(cpf[9] - '0')
	if firstCheckDigit != calculateCheckDigit(cpf, 9, 10) {
		return false
	}

	secondCheckDigit := int(cpf[10] - '0')
	if secondCheckDigit != calculateCheckDigit(cpf, 10, 11) {
		return false
	}

	return true
}

func hasAllEqualDigits(s string) bool {
	for i := 1; i < size; i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}

func calculateCheckDigit(cpf string, length int, weightStart int) int {
	sum := 0
	for i := 0; i < length; i++ {
		sum += int(cpf[i]-'0') * (weightStart - i)
	}

	d := (sum * 10) % 11
	if d == 10 {
		return 0
	}
	return d
}

func onlyNumbers(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
