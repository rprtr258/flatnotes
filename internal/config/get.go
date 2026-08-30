package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
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
