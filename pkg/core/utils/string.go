package utils

import "regexp"

func RemoveOuterQuotes(s string) string {
	re := regexp.MustCompile(`^"(.*)"$`)
	return re.ReplaceAllString(s, `$1`)
}
