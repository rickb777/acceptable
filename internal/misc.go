package internal

import (
	"strings"
)

func Split1(value string, b byte) (string, string) {
	i := strings.IndexByte(value, b)
	if i < 0 {
		return value, ""
	}
	return value[:i], value[i+1:]
}

//-------------------------------------------------------------------------------------------------

type Strings []string

func Split(value, cut string) Strings {
	return strings.Split(value, cut)
}

func (ss Strings) TrimSpace() Strings {
	for i := 0; i < len(ss); i++ {
		ss[i] = strings.TrimSpace(ss[i])
	}
	return ss
}

func (ss Strings) RemoveQuotes() Strings {
	for i := 0; i < len(ss); i++ {
		ss[i] = strings.Trim(ss[i], `"`)
	}
	return ss
}

func (ss Strings) Contains(v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}
