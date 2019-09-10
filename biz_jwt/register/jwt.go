package main

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

var (
	secretKey = []byte("adcd1234!@#$")
)

type BizCustomClaim struct {
	UserID string `json:"userID"`
	Name   string `json:"name"`
	jwt.StandardClaims
}

func JwtKeyFunc(token *jwt.Token) (interface{}, error) {
	return secretKey, nil
}

func Sign(name, uid string) (string, error) {
	expAt := time.Now().Add(time.Duration(10) * time.Minute).Unix()

	claims := BizCustomClaim{
		UserID: uid,
		Name:   name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expAt,
			Issuer:    "system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secretKey)
}
