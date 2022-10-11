package utils

import (
	"regexp"
)

var re = regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)

func ExtractRemoteAddress(value string) string {
	for _, element := range re.FindAllString(value, 1) {
		return element
	}
	return ""
}
