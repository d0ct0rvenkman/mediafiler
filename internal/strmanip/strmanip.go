package strmanip

import "strings"

/*
strtr — Translate characters or replace substrings

Poached from https://www.php2golang.com/method/function.strtr.html
*/
func Strtr(str string, replace map[string]string) string {
	if len(replace) == 0 || len(str) == 0 {
		return str
	}
	for old, new := range replace {
		str = strings.ReplaceAll(str, old, new)
	}
	return str
}
