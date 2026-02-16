package tokenadapter

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/soat13/oficina-serverless/internal/identity/app/ports/out"
	"github.com/soat13/oficina-serverless/internal/shared/token"
)

type JWTService struct {
	secret []byte
	issuer string
	method jwt.SigningMethod
}

func New(secret, issuer string) *JWTService {
	return NewWithMethod(secret, issuer, jwt.SigningMethodHS256)
}

func NewWithMethod(secret, issuer string, method jwt.SigningMethod) *JWTService {
	if method == nil {
		method = jwt.SigningMethodHS256
	}

	return &JWTService{
		secret: []byte(secret),
		issuer: issuer,
		method: method,
	}
}

func (s *JWTService) Sign(_ context.Context, sub token.Subject, opts token.SignOptions) (token.Token, error) {
	if len(s.secret) == 0 || s.issuer == "" || sub == "" || opts.TTLSeconds <= 0 {
		return "", out.ErrInvalidToken
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub": string(sub),
		"iss": s.issuer,
		"iat": now.Unix(),
		"exp": now.Add(time.Duration(opts.TTLSeconds) * time.Second).Unix(),
	}

	signed, err := jwt.NewWithClaims(s.method, claims).SignedString(s.secret)
	if err != nil {
		return "", out.ErrInvalidToken
	}

	return token.Token(signed), nil
}

func (s *JWTService) Verify(_ context.Context, tk token.Token) (token.Subject, error) {
	if len(s.secret) == 0 || s.issuer == "" || tk == "" {
		return "", out.ErrInvalidToken
	}

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(
		string(tk),
		claims,
		func(_ *jwt.Token) (interface{}, error) { return s.secret, nil },
		jwt.WithValidMethods([]string{s.method.Alg()}),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", out.ErrTokenExpired
		}
		return "", out.ErrInvalidToken
	}

	if iss, _ := claims["iss"].(string); iss != s.issuer {
		return "", out.ErrInvalidToken
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", out.ErrInvalidToken
	}

	return token.Subject(sub), nil
}
