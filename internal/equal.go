package internal

import (
	"strings"
)

func EqualOrPrefix(acceptedLang, offeredLang string) bool {
	return acceptedLang == "*" ||
		offeredLang == "*" ||
		acceptedLang == offeredLang ||
		strings.HasPrefix(acceptedLang, offeredLang+"-")
}

func EqualOrWildcard(accepted, offered string) bool {
	return offered == "*" ||
		accepted == "*" ||
		accepted == offered
}
