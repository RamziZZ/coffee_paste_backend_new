package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Role string `json:"role"`
	Ver  int    `json:"ver"` // versi sesi (users.token_version)
	jwt.RegisteredClaims
}

func NewToken(secret []byte, ttl time.Duration, userID int64, role string, ver int) (string, error) {
	now := time.Now()
	claims := Claims{
		Role: role,
		Ver:  ver,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

func ParseToken(secret []byte, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	t, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, errors.New("token tidak valid")
	}
	return claims, nil
}
