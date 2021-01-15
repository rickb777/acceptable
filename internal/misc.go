package internal

import "strings"

func Split(value string, b byte) (string, string) {
	i := strings.IndexByte(value, b)
	if i < 0 {
		return value, ""
	}
	return value[:i], value[i+1:]
}

func CallDataSuppliers(data interface{}, template, language string) (result interface{}, err error) {
	result = data
loop:
	for {
		switch fn := result.(type) {
		case func(string, string) (interface{}, error):
			result, err = fn(template, language)
		case func() (interface{}, error):
			result, err = fn()
		default:
			break loop
		}
		if err != nil {
			return nil, err
		}
	}
	return result, err
}
