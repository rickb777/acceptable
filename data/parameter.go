package data

import (
	"fmt"
	"strings"
)

type Parameter interface {
	fmt.Stringer
}

type TemplateName string

func (t TemplateName) String() string {
	return string(t)
}

type Language string

func (t Language) String() string {
	return string(t)
}

type Languages []Language

func SplitLangs(s, sep string) Languages {
	ss := strings.Split(s, sep)
	ls := make([]Language, len(ss))
	for i, s := range ss {
		ls[i] = Language(s)
	}
	return ls
}
