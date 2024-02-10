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

func getEnvStr(key string, mandatory bool, defaultT string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		if mandatory {
			return "", fmt.Errorf("environment variable %s must be set", key)
		}

		return defaultT, nil
	}

	return value, nil
}

func getEnvInt(key string, defaultT int) (int, error) {
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

func getAuthType() (AuthType, error) {
	const _key = "FLATNOTES_AUTH_TYPE"
	rawAuthType, err := getEnvStr(_key, false, string(AuthTypePassword))
	if err != nil {
		return AuthTypeNone, err
	}

	authType := AuthType(strings.ToLower(rawAuthType))
	if fun.Contains(authType, AuthTypeNone, AuthTypeReadOnly, AuthTypePassword, AuthTypeTOTP) {
		return authType, nil
	}

	variants := strings.Join([]string{
		string(AuthTypeNone),
		string(AuthTypeReadOnly),
		string(AuthTypePassword),
		string(AuthTypeTOTP),
	}, ", ")
	return "", fmt.Errorf("Invalid value %s for %s. Must be one of{ "+variants+".", rawAuthType, _key)
}

func getTotpKey(authType AuthType) (string, error) {
	totpKey, err := getEnvStr("FLATNOTES_TOTP_KEY", authType == AuthTypeTOTP, "")
	// if totpKey!=nil {
	// 	return b32encode(totpKey.encode("utf-8"))
	// }
	return totpKey, err
}

func Read() (Config, error) {
	authType, err := getAuthType()
	if err != nil {
		return Config{}, err
	}

	authNeeded := !fun.Contains(authType, AuthTypeNone, AuthTypeReadOnly)

	dataPath, err := getEnvStr("FLATNOTES_PATH", false, "/data")
	if err != nil {
		return Config{}, err
	}

	username, err := getEnvStr("FLATNOTES_USERNAME", authNeeded, "")
	if err != nil {
		return Config{}, err
	}

	password, err := getEnvStr("FLATNOTES_PASSWORD", authNeeded, "")
	if err != nil {
		return Config{}, err
	}

	sessionKey, err := getEnvStr("FLATNOTES_SECRET_KEY", authNeeded, "")
	if err != nil {
		return Config{}, err
	}

	sessionExpiryDays, err := getEnvInt("FLATNOTES_SESSION_EXPIRY_DAYS", 30)
	if err != nil {
		return Config{}, err
	}

	totpKey, err := getTotpKey(authType)
	if err != nil {
		return Config{}, err
	}

	return Config{
		DataPath:          dataPath,
		AuthType:          authType,
		Username:          username,
		Password:          password,
		SessionKey:        sessionKey,
		SessionExpiryDays: sessionExpiryDays,
		TotpKey:           totpKey,
	}, nil
}
