package internal

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func strip_ext(filename string) string {
	_, fname := filepath.Split(filename)
	fnamefext := strings.Split(fname, ".")
	return fnamefext[0]
}

// Return the declared snake_case string in camelCase.
func camel_case(snake_case_str string) string {
	res := ""
	for _, part := range strings.Split(snake_case_str, "_") {
		if part == "" {
			continue
		}
		if res == "" {
			res += part
		} else {
			res += strings.ToTitle(part)
		}
	}
	return res
}

func empty_dir(path string) {
	os.RemoveAll(path)
}

// Similar to re.sub but returns a tuple of:
//
// - `string` with matches removed
// - list of matches
func re_extract(re *regexp.Regexp, string string) (string, []string) {
	text := re.ReplaceAllLiteralString(string, "")
	matches := re.FindStringSubmatch(string)
	return text, matches
}
