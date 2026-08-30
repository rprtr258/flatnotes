package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rprtr258/fun"
)

func getEnvStr(key string, mandatory bool, defaultT string) fun.Result[string] {
	value, ok := os.LookupEnv(key)
	if ok {
		return fun.Ok(value)
	}

	if mandatory {
		return fun.Err[string](fmt.Errorf("environment variable %s must be set", key))
	}

	return fun.Ok(defaultT)
}

func getEnvInt(key string, defaultT int) fun.Result[int] {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fun.Ok(defaultT)
	}

	res, err := strconv.Atoi(value)
	if err != nil {
		return fun.Err[int](fmt.Errorf("invalid value %q for %s", value, key))
	}

	return fun.Ok(res)
}

func getAuthType() fun.Result[AuthType] {
	const _key = "FLATNOTES_AUTH_TYPE"
	rawAuthType := getEnvStr(_key, false, string(AuthTypePassword))
	if rawAuthType.Err != nil {
		return fun.Err[AuthType](rawAuthType.Err)
	}

	authType := AuthType(strings.ToLower(rawAuthType.Value))
	if fun.Contains(authType, AuthTypeNone, AuthTypeReadOnly, AuthTypePassword, AuthTypeTOTP) {
		return fun.Ok(authType)
	}

	variants := strings.Join([]string{
		string(AuthTypeNone),
		string(AuthTypeReadOnly),
		string(AuthTypePassword),
		string(AuthTypeTOTP),
	}, ", ")
	return fun.Err[AuthType](fmt.Errorf("Invalid value %s for %s. Must be one of{ "+variants+".", rawAuthType.Value, _key))
}

func getTotpKey(authType AuthType) fun.Result[string] {
	totpKey := getEnvStr("FLATNOTES_TOTP_KEY", authType == AuthTypeTOTP, "")
	// if totpKey!=nil {
	// 	return b32encode(totpKey.encode("utf-8"))
	// }
	return totpKey
}
