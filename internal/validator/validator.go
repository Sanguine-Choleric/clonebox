package validator

import (
	"mime"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Double check this
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Validator type contains a map of validation errors for form fields.
type Validator struct {
	FieldErrors    map[string]string
	NonFieldErrors []string
}

// Valid returns true if the FieldErrors and NonFieldErrors map don't contain any entries.
func (v *Validator) Valid() bool {
	return (len(v.FieldErrors) == 0) && (len(v.NonFieldErrors) == 0)
}

// AddFieldError adds an error message to the FieldErrors map (so long as no entry already exists for the given key).
func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}
	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// NotBlank returns true if a value is not an empty string.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxChars returns true if a value contains no more than n characters.
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// // PermittedInt returns true if a value is in a list of permitted integers.
// func PermittedInt(value int, permittedValues ...int) bool {
// 	for i := range permittedValues {
// 		if value == permittedValues[i] {
// 			return true
// 		}
// 	}
// 	return false
// }

// PermittedValue is a generic PermittedValue() function that returns true if the value of type T equals one of the variadic
// permittedValues parameters
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}
	return false
}

func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func IsURL(value string) bool {
	u, err := url.ParseRequestURI(value)
	if err != nil || u.Scheme == "" || u.Hostname() == "" {
		return false
	}

	return true
}

// PermittedMimeType is used for validating uploaded file types based on provided bytes
func PermittedMimeType(s string, ss []string) bool {
	//var acceptedMimes = [2]string{"image/*", "application/pdf"}
	s = strings.TrimSpace(s)
	s, _, err := mime.ParseMediaType(s)
	if err != nil {
		return false
	}

	for _, v := range ss {
		if strings.HasSuffix(v, "/*") {
			// Wildcard match: "image/*" matches "image/jpeg", "image/png"
			prefix := strings.TrimSuffix(v, "*")
			if strings.HasPrefix(s, prefix) {
				return true
			}
		} else if v == s {
			// Exact match: "application/pdf" == "application/pdf"
			return true
		}
	}

	return false
}
