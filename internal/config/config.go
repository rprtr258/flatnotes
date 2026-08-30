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
	authNeeded := authType.Map(func(authType AuthType) bool {
		return !fun.Contains(authType, AuthTypeNone, AuthTypeReadOnly)
	})
	dataPath := getEnvStr("FLATNOTES_PATH", false, "/data")
	username := authNeeded.FlatMap(func(authNeeded bool) fun.Result[string] {
		return getEnvStr("FLATNOTES_USERNAME", authNeeded, "")
	})
	password := authNeeded.FlatMap(func(authNeeded bool) fun.Result[string] {
		return getEnvStr("FLATNOTES_PASSWORD", authNeeded, "")
	})
	sessionKey := authNeeded.FlatMap(func(authNeeded bool) fun.Result[string] {
		return getEnvStr("FLATNOTES_SECRET_KEY", authNeeded, "")
	})
	sessionExpiryDays := getEnvInt("FLATNOTES_SESSION_EXPIRY_DAYS", 30)
	totpKey := authType.FlatMap(getTotpKey)
	return dataPath.FlatMap(func(dataPath string) fun.Result[Config] {
		return authType.FlatMap(func(authType AuthType) fun.Result[Config] {
			return username.FlatMap(func(username string) fun.Result[Config] {
				return password.FlatMap(func(password string) fun.Result[Config] {
					return sessionKey.FlatMap(func(sessionKey string) fun.Result[Config] {
						return sessionExpiryDays.FlatMap(func(sessionExpiryDays int) fun.Result[Config] {
							return totpKey.FlatMap(func(totpKey string) fun.Result[Config] {
								return fun.Ok(Config{
									DataPath:      dataPath,
									AuthType:      authType,
									Username:      username,
									Password:      password,
									SessionKey:    sessionKey,
									SessionExpiry: time.Duration(sessionExpiryDays) * 24 * time.Hour,
									TotpKey:       totpKey,
								})
							})
						})
					})
				})
			})
		})
	}).Unpack()
}
