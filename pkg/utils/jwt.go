package utils

import (
	"errors"
	"github.com/Nubes3/nubes3-user-service/pkg/config"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type UserClaims struct {
	Id string
	jwt.StandardClaims
}

func CreateToken(oid string) (string, error) {
	userClaims := &UserClaims{
		Id: oid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 3600).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	signed, err := token.SignedString([]byte(config.Conf.Secret))
	if err != nil {
		return "", err
	}

	return signed, nil
}

func ParseToken(authToken string, claims *UserClaims) (*jwt.Token, error) {
	return jwt.ParseWithClaims(authToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("wrong method")
		}
		return []byte(config.Conf.Secret), nil
	})
}
