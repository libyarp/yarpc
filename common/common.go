package common

import (
	"fmt"
	"github.com/libyarp/idl"
	"github.com/urfave/cli"
	"regexp"
	"strings"
	"unicode"
)

func ProcessInputFiles(c *cli.Context) (*idl.FileSet, error) {
	if c.NArg() == 0 {
		return nil, cli.NewExitError("No input files", 0)
	}
	fs := idl.NewFileSet()

	ok := true
	for _, v := range c.Args() {
		if err := fs.Load(v); err != nil {
			ok = false
			fmt.Printf("ERROR: %s\n", err)
		}
	}

	if !ok {
		return nil, cli.NewExitError("Errors found processing input files", 2)
	}

	return fs, nil
}

var commonInitialisms = map[string]bool{
	"ACL":   true,
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SQL":   true,
	"SSH":   true,
	"TCP":   true,
	"TLS":   true,
	"TTL":   true,
	"UDP":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"XMPP":  true,
	"XSRF":  true,
	"XSS":   true,
}

var (
	firstCap = regexp.MustCompile(`(.)([A-Z][a-z]+)`)
	allCap   = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

func CamelToSnake(input string) string {
	tmpl := "${1}_${2}"
	snake := firstCap.ReplaceAllString(input, tmpl)
	snake = allCap.ReplaceAllString(snake, tmpl)
	return strings.ToLower(snake)
}

func SnakeToCamel(input string) string {
	components := strings.Split(strings.ToLower(input), "_")
	result := make([]string, 0, len(components))
	for _, c := range components {
		converted := strings.ToUpper(c)
		if _, ok := commonInitialisms[converted]; ok {
			result = append(result, converted)
		} else {
			arr := []rune(c)
			arr[0] = unicode.ToUpper(arr[0])
			result = append(result, string(arr))
		}
	}
	return strings.Join(result, "")
}

func Titleize(input string) string {
	v := []rune(input)
	v[0] = unicode.ToUpper(v[0])
	return string(v)
}
