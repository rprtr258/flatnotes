package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// from base64 import b32encode

type AuthType string

const (
	AuthTypeNone     AuthType = "none"
	AuthTypeReadOnly AuthType = "read_only"
	AuthTypePassword AuthType = "password"
	AuthTypeTOTP     AuthType = "totp"
)

// Get an environment variable.
func getEnv[T interface {
	string | int
}](key string, mandatory bool, defaultT T) T {
	value, ok := os.LookupEnv(key)
	if !ok {
		if mandatory {
			log.Fatal().Str("env", key).Msg("environment variable must be set")
		}
		return defaultT
	}

	if _, ok := any(*new(T)).(int); ok {
		res, err := strconv.Atoi(value)
		if err != nil {
			log.Fatal().Str("env", key).Str("val", value).Msg("invalid value")
		}
		return any(res).(T)
	}
	return any(value).(T)
}

type Config struct {
	DataPath      string
	AuthType      AuthType
	Username      string
	Password      string
	SessionKey    string
	SessionExpiry time.Duration
	TotpKey       string
}

func getAuthType() AuthType {
	const key = "FLATNOTES_AUTH_TYPE"
	rawAuthType := getEnv[string](key, false, string(AuthTypePassword))
	switch authType := AuthType(strings.ToLower(rawAuthType)); authType {
	case AuthTypeNone, AuthTypeReadOnly, AuthTypePassword, AuthTypeTOTP:
		return authType
	default:
		variants := strings.Join([]string{
			string(AuthTypeNone),
			string(AuthTypeReadOnly),
			string(AuthTypePassword),
			string(AuthTypeTOTP),
		}, ", ")
		log.Fatal().Str("env", key).Str("val", rawAuthType).Msg("invalid value, must be one of: " + variants)
	}
	panic("unreachable")
}

func getTOTPKey(authType AuthType) string {
	totpKey := getEnv[string]("FLATNOTES_TOTP_KEY", authType == AuthTypeTOTP, "")
	// if totpKey!=nil {
	// 	return b32encode(totpKey.encode("utf-8"))
	// }
	return totpKey
}

func New() (Config, error) {
	authType := getAuthType()
	authNeeded := authType != AuthTypeNone && authType != AuthTypeReadOnly
	return Config{
		DataPath:      getEnv[string]("FLATNOTES_PATH", false, "/data"),
		AuthType:      authType,
		Username:      getEnv[string]("FLATNOTES_USERNAME", authNeeded, ""),
		Password:      getEnv[string]("FLATNOTES_PASSWORD", authNeeded, ""),
		SessionKey:    getEnv[string]("FLATNOTES_SECRET_KEY", authNeeded, ""),
		SessionExpiry: time.Duration(getEnv[int]("FLATNOTES_SESSION_EXPIRY_DAYS", false, 30)) * 24 * time.Hour,
		TotpKey:       getTOTPKey(authType),
	}, nil
}
