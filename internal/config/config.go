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
func get_env[T interface {
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

func get_auth_type() AuthType {
	const key = "FLATNOTES_AUTH_TYPE"
	rawAuthType := get_env[string](key, false, string(AuthTypePassword))
	switch auth_type := AuthType(strings.ToLower(rawAuthType)); auth_type {
	case AuthTypeNone, AuthTypeReadOnly, AuthTypePassword, AuthTypeTOTP:
		return auth_type
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

func get_totp_key(auth_type AuthType) string {
	totp_key := get_env[string]("FLATNOTES_TOTP_KEY", auth_type == AuthTypeTOTP, "")
	// if totp_key!=nil {
	// 	return b32encode(totp_key.encode("utf-8"))
	// }
	return totp_key
}

func New() (Config, error) {
	auth_type := get_auth_type()
	auth_needed := auth_type != AuthTypeNone && auth_type != AuthTypeReadOnly
	return Config{
		DataPath:      get_env[string]("FLATNOTES_PATH", false, "/data"),
		AuthType:      auth_type,
		Username:      get_env[string]("FLATNOTES_USERNAME", auth_needed, ""),
		Password:      get_env[string]("FLATNOTES_PASSWORD", auth_needed, ""),
		SessionKey:    get_env[string]("FLATNOTES_SECRET_KEY", auth_needed, ""),
		SessionExpiry: time.Duration(get_env[int]("FLATNOTES_SESSION_EXPIRY_DAYS", false, 30)) * 24 * time.Hour,
		TotpKey:       get_totp_key(auth_type),
	}, nil
}
