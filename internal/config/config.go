package config

import (
	"time"

	"github.com/rprtr258/fun"
)

type AuthType string

const (
	AuthTypeNone     AuthType = "none"
	AuthTypeReadOnly AuthType = "read_only"
	AuthTypePassword AuthType = "password"
	AuthTypeTOTP     AuthType = "totp"
)

type Config struct {
	DataPath      string
	AuthType      AuthType
	Username      string
	Password      string
	SessionKey    string
	SessionExpiry time.Duration
	TotpKey       string
}

func New() (Config, error) {
	authType := getAuthType()
	authNeeded := !fun.Contains(authType, AuthTypeNone, AuthTypeReadOnly)
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
