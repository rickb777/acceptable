package header

import "strings"

// Strings is simply a slice of strings.
type Strings []string

// Split is a convenience wrapper for strings.Split.
func Split(value, cut string) Strings {
	return strings.Split(value, cut)
}

// TrimSpace trims all the strings in the slice.
func (ss Strings) TrimSpace() Strings {
	for i := 0; i < len(ss); i++ {
		ss[i] = strings.TrimSpace(ss[i])
	}
	return ss
}

// RemoveQuotes removes quotes from all the strings in the slice.
func (ss Strings) RemoveQuotes() Strings {
	for i := 0; i < len(ss); i++ {
		ss[i] = strings.Trim(ss[i], `"`)
	}
	return ss
}

// Contains looks for a string in the slice.
func (ss Strings) Contains(v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}
