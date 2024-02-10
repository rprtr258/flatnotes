package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rprtr258/fun"
)

type AuthType string

const (
	AuthTypeNone     AuthType = "none"
	AuthTypeReadOnly AuthType = "read_only"
	AuthTypePassword AuthType = "password"
	AuthTypeTOTP     AuthType = "totp"
)

func get_env_str(key string, mandatory bool, defaultT string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		if mandatory {
			return "", fmt.Errorf("environment variable %s must be set", key)
		}

		return defaultT, nil
	}

	return value, nil
}

func get_env_int(key string, defaultT int) (int, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultT, nil
	}

	res, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid value %q for %s", value, key)
	}

	return res, nil
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

func get_auth_type() (AuthType, error) {
	const _key = "FLATNOTES_AUTH_TYPE"
	rawAuthType, err := get_env_str(_key, false, string(AuthTypePassword))
	if err != nil {
		return AuthTypeNone, err
	}

	auth_type := AuthType(strings.ToLower(rawAuthType))
	if fun.Contains(auth_type, AuthTypeNone, AuthTypeReadOnly, AuthTypePassword, AuthTypeTOTP) {
		return auth_type, nil
	}

	variants := strings.Join([]string{
		string(AuthTypeNone),
		string(AuthTypeReadOnly),
		string(AuthTypePassword),
		string(AuthTypeTOTP),
	}, ", ")
	return "", fmt.Errorf("Invalid value %s for %s. Must be one of{ "+variants+".", rawAuthType, _key)
}

func get_totp_key(auth_type AuthType) (string, error) {
	totp_key, err := get_env_str("FLATNOTES_TOTP_KEY", auth_type == AuthTypeTOTP, "")
	// if totp_key!=nil {
	// 	return b32encode(totp_key.encode("utf-8"))
	// }
	return totp_key, err
}

func Read() (Config, error) {
	auth_type, err := get_auth_type()
	if err != nil {
		return Config{}, err
	}

	auth_needed := !fun.Contains(auth_type, AuthTypeNone, AuthTypeReadOnly)

	dataPath, err := get_env_str("FLATNOTES_PATH", false, "/data")
	if err != nil {
		return Config{}, err
	}

	username, err := get_env_str("FLATNOTES_USERNAME", auth_needed, "")
	if err != nil {
		return Config{}, err
	}

	password, err := get_env_str("FLATNOTES_PASSWORD", auth_needed, "")
	if err != nil {
		return Config{}, err
	}

	sessionKey, err := get_env_str("FLATNOTES_SECRET_KEY", auth_needed, "")
	if err != nil {
		return Config{}, err
	}

	sessionExpiryDays, err := get_env_int("FLATNOTES_SESSION_EXPIRY_DAYS", 30)
	if err != nil {
		return Config{}, err
	}

	totpKey, err := get_totp_key(auth_type)
	if err != nil {
		return Config{}, err
	}

	return Config{
		DataPath:          dataPath,
		AuthType:          auth_type,
		Username:          username,
		Password:          password,
		SessionKey:        sessionKey,
		SessionExpiryDays: sessionExpiryDays,
		TotpKey:           totpKey,
	}, nil
}
