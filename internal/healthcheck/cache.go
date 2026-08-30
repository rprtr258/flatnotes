package healthcheck

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func cachePath(url string) string {
	sum := sha256.Sum256([]byte(url))
	return filepath.Join(_cacheDir, hex.EncodeToString(sum[:]))
}

// readCache returns a cached status code if a fresh file exists for url.
func readCache(url string) (int, bool) {
	p := cachePath(url)
	info, err := os.Stat(p)
	if err != nil || time.Since(info.ModTime()) >= _cacheTTL {
		return 0, false
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return 0, false
	}
	code, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		return 0, false
	}
	return code, true
}

// writeCache stores code for url, creating the cache dir if needed.
func writeCache(url string, code int) {
	_ = os.MkdirAll(_cacheDir, 0o755)
	_ = os.WriteFile(cachePath(url), []byte(strconv.Itoa(code)), 0o644)
}

func getOr(url string, get func() (int, error)) (int, bool) {
	if res, ok := readCache(url); ok {
		return res, true
	}
	res, err := get()
	if err != nil {
		return 0, false
	}
	writeCache(url, res)
	return res, true
}
