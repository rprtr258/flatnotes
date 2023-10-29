package internal

import (
	"log"
	"os"
	"strconv"
	"strings"
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
func get_env(key string, mandatory bool, defaultT any, cast_int bool) any { // string | int
	value, ok := os.LookupEnv(key)
	if !ok {
		if mandatory {
			log.Fatalf("Environment variable %s must be set.", key)
		}
		return defaultT
	}

	if cast_int {
		res, err := strconv.Atoi(value)
		if err != nil {
			log.Fatalf("Invalid value %q for %s.", value, key)
		}
		return res
	}
	return value
}

type Config struct {
	DataPath          string
	AuthType          AuthType
	Username          string
	Password          string
	SessionKey        string
	SessionExpiryDays int // TODO: time.Duration
	TotpKey           string
}

func get_auth_type() AuthType {
	const key = "FLATNOTES_AUTH_TYPE"
	rawAuthType := get_env(key, false, string(AuthTypePassword), false).(string)
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
		log.Fatalf("Invalid value %s for %s. Must be one of{ "+variants+".", rawAuthType, key)
	}
	panic("unreachable")
}

func get_totp_key(auth_type AuthType) string {
	totp_key := get_env("FLATNOTES_TOTP_KEY", auth_type == AuthTypeTOTP, "", false).(string)
	// if totp_key!=nil {
	// 	return b32encode(totp_key.encode("utf-8"))
	// }
	return totp_key
}

func NewConfig() Config {
	auth_type := get_auth_type()
	auth_needed := auth_type != AuthTypeNone && auth_type != AuthTypeReadOnly
	return Config{
		DataPath:          get_env("FLATNOTES_PATH", false, "/data", false).(string),
		AuthType:          auth_type,
		Username:          get_env("FLATNOTES_USERNAME", auth_needed, Optional[string]{}, false).(string),
		Password:          get_env("FLATNOTES_PASSWORD", auth_needed, Optional[string]{}, false).(string),
		SessionKey:        get_env("FLATNOTES_SECRET_KEY", auth_needed, Optional[string]{}, false).(string),
		SessionExpiryDays: get_env("FLATNOTES_SESSION_EXPIRY_DAYS", false, 30, true).(int),
		TotpKey:           get_totp_key(auth_type),
	}
}
