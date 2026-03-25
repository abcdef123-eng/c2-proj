package server

import (
	"time"

	"github.com/execute-assembly/c2-proj/newserver/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

func CreateJWT(Guid string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"guid": Guid,
			"exp":  time.Now().Add(30 * 24 * time.Hour).Unix(),
		})

	tokenStr, err := token.SignedString([]byte(config.Cfg.JwtSecret))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func VerifyToken(tokenStr string) error {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(config.Cfg.JwtSecret), nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrTokenInvalidClaims
	}
	return nil
}
