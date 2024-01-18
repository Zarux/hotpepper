package parse

import (
	"errors"
)

type Leaf struct {
	Name    string
	Request struct {
		Method  string
		URL     string
		Headers map[string]string
		Body    []byte
	}
	PreCode   string
	PostCode  string
	DependsOn []string
	Imports   []string
}

func (l *Leaf) ExpandTemplate(t *Template, strict bool) error {
	method, err := t.Expand(l.Request.Method)
	if err != nil && (!errors.Is(err, errMissingTemplateValue) || strict) {
		return err
	}
	if !errors.Is(err, errMissingTemplateValue) {
		l.Request.Method = method
	}

	url, err := t.Expand(l.Request.URL)
	if err != nil && (!errors.Is(err, errMissingTemplateValue) || strict) {
		return err
	}
	if !errors.Is(err, errMissingTemplateValue) {
		l.Request.URL = url
	}

	for key, value := range l.Request.Headers {
		val, err := t.Expand(value)
		if err != nil && (!errors.Is(err, errMissingTemplateValue) || strict) {
			return err
		}
		if !errors.Is(err, errMissingTemplateValue) {
			l.Request.Headers[key] = val
		}
	}
	return nil
}

func (l *Leaf) MustExpandTemplate(t *Template) {
	l.Request.Method = t.MustExpand(l.Request.Method)
	l.Request.URL = t.MustExpand(l.Request.URL)
	for key, value := range l.Request.Headers {
		l.Request.Headers[key] = t.MustExpand(value)
	}
}
