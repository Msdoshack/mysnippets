package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

type Validator struct {
	FieldErrors map[string]string
	NonFieldErrors []string
}

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zAZ0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = map[string]string{}
	}

	if _, exist := v.FieldErrors[key]; !exist {
		v.FieldErrors[key] = message
	}
}

func (v *Validator) AddNonFieldError (message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

// CheckField() adds an error message to the FieldErrors map only if a
// validation check is not 'ok'.
func (v *Validator) CheckField(ok bool, key, message string) {
	
	if !ok {
		v.AddFieldError(key, message)
	}
}


// NotBlank() returns true if a value is not an empty string.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}


// MaxChars() returns true if a value contains no more than n characters.
func MaxChars (value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

func MinChars (value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

func PermittedInt(value int, permittedValues ...int) bool {
	for i := range permittedValues{
		if value == permittedValues[i]{
			return true
		}
	}
	return false
}

func PermittedVisibility(value string, permittedValues ...string) bool {
	for i := range permittedValues{
		if value == permittedValues[i]{
			return true
		}
	}
	return false
}


func IsPasswordMatch (password1, password2 string) bool {
	return password1 == password2
}


func Matches (value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

