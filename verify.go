package main

import "regexp"

func Verify(password string) bool {
	verified := true
	regexTests := []string{".{7,}", "[a-z]", "[A-Z]", "[0-9]", "[^\\d\\w]"}
	for _, test := range regexTests {
		t, _ := regexp.MatchString(test, password)
		if !t {
			verified = false
			break
		}
	}
	return verified
}
