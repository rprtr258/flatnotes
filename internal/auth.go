package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const JWT_ALGORITHM = "HS256"

// var oauth2_scheme = OAuth2PasswordBearer(tokenUrl="api/token")

type claims struct {
	jwt.RegisteredClaims
}

func CreateAccessToken(config Config, username string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(time.Now().AddDate(0, 0, config.SessionExpiryDays)),
		},
	}).SignedString([]byte(config.SessionKey)) //(to_encode, config.session_key, JWT_ALGORITHM)
}

func ValidateToken(config Config, token string /*= Depends(oauth2_scheme)*/) error {
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

func Authenticate(config Config, data LoginModel, last_used_totp *string) (TokenModel, error) {
	expected_password := config.Password
	var current_totp string
	if config.AuthType == AuthTypeTOTP {
		current_totp = "" // totp.now()
		// expected_password += current_totp
	}

	if config.Username != data.Username || expected_password != data.Password ||
		// Prevent TOTP from being reused
		config.AuthType == AuthTypeTOTP && *last_used_totp != "" && current_totp == *last_used_totp {
		return TokenModel{}, fmt.Errorf("Incorrect login credentials.")
	}

	access_token, err := CreateAccessToken(config, config.Username)
	if err != nil {
		return TokenModel{}, fmt.Errorf("create access token: %s", err.Error())
	}

	if config.AuthType == AuthTypeTOTP {
		*last_used_totp = current_totp
	}
	return TokenModel{
		AccessToken: access_token,
		TokenType:   "bearer",
	}, nil
}
