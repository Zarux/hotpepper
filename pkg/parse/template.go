package parse

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type Template struct {
	values map[string]any
}

var errMissingTemplateValue = errors.New("missing template value")

func NewTemplate(values map[string]any) *Template {
	return &Template{
		values: values,
	}
}

func (t *Template) MustExpand(s string) string {
	s, err := t.Expand(s)
	if err != nil {
		panic(err)
	}
	return s
}

var rgxTemplateString = regexp.MustCompile(`(?m)(\$\{(.+?)\})`)

func (t *Template) Expand(s string) (string, error) {
	match := rgxTemplateString.FindStringSubmatch(s)
	if s == "" || len(match) == 0 {
		return s, nil
	}
	for len(match) > 0 {
		match := rgxTemplateString.FindStringSubmatch(s)
		if len(match) == 0 {
			break
		}
		tString := match[1]
		valKey := match[2]
		val := t.values[valKey]
		if val == nil {
			return "", fmt.Errorf("%w for %s", errMissingTemplateValue, valKey)
		}
		valStr, ok := val.(string)
		if !ok {
			return "", fmt.Errorf("invalid type for var %s got %s", valKey, reflect.TypeOf(val))
		}
		s = strings.ReplaceAll(s, tString, valStr)
	}
	return s, nil
}
