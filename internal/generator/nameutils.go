package generator

import (
	"errors"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

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

func GoIdentLowercase(name string) string {
	match := 1
	for i := range commonInitialisms {
		if strings.HasPrefix(name, i) {
			match = max(match, len(i))
		}
	}
	lowerCaser := cases.Lower(language.Und)
	if len(name) < match {
		return lowerCaser.String(name)
	}

	return lowerCaser.String(name[0:match]) + name[match:]
}

func FormatGoLikeIdentifier(name string) string {
	name = strings.ReplaceAll(name, "{", "")
	name = strings.ReplaceAll(name, "}", "")

	items := strings.Split(name, "/")
	items2 := make([]string, 0, len(items))
	for _, item := range items {
		items2 = append(items2, strings.Split(item, "-")...)
	}
	items3 := make([]string, 0, len(items2))
	for _, item := range items2 {
		items3 = append(items3, strings.Split(item, "_")...)
	}

	titleCaser := cases.Title(language.Und)
	upperCaser := cases.Upper(language.Und)

	result := make([]string, 0, len(items3))
	for _, item := range items3 {
		if commonInitialisms[upperCaser.String(item)] {
			result = append(result, upperCaser.String(item))
			continue
		}

		result = append(result, titleCaser.String(item))
	}

	return strings.Join(result, "")
}

func ParseRefTypeName(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1]
}

var ErrUnsupportedContentType = errors.New("unsupported content type")

func NameSuffixFromContentType(contentType string) (string, error) {
	switch contentType {
	case applicationJSONCT:
		return "Json", nil
	default:
		return "", ErrUnsupportedContentType
	}
}
