package internal

import (
	"fmt"
	"strings"
	"time"

	// from fastapi import Depends, HTTPException
	// from fastapi.security import OAuth2PasswordBearer
	// from jose import JWTError, jwt

	"github.com/golang-jwt/jwt/v5"
)

const JWT_ALGORITHM = "HS256"

// var oauth2_scheme = OAuth2PasswordBearer(tokenUrl="api/token")

type claims struct {
	jwt.RegisteredClaims
}

func Create_access_token(config Config, username string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(time.Now().AddDate(0, 0, config.SessionExpiryDays)),
		},
	}).SignedString([]byte(config.SessionKey)) //(to_encode, config.session_key, JWT_ALGORITHM)
}

func Validate_token(config Config, token string /*= Depends(oauth2_scheme)*/) error {
	// try:
	var claims claims
	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("aboba")
		}

		return []byte(config.SessionKey), nil
	})
	if err != nil {
		return fmt.Errorf("parse token: %w", err)
	}

	username, err := claims.GetSubject()
	if err != nil || !strings.EqualFold(username, config.Username) {
		return fmt.Errorf("ValueError")
	}

	return nil
	// except (JWTError, ValueError):
	//     raise HTTPException(
	//         status_code=401,
	//         detail="Invalid authentication credentials",
	//         headers={"WWW-Authenticate": "Bearer"},
	//     )
}
