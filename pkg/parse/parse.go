package parse

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

func Parse(fileData []byte) ([]*Leaf, error) {
	items, err := parseGlobal(fileData)
	if err != nil {
		return nil, err
	}
	g := errgroup.Group{}
	g.SetLimit(10)

	var (
		mu    sync.Mutex
		leafs []*Leaf
	)

	for name, item := range items {
		name, item := name, item
		g.Go(func() error {
			l, err := parseLeaf(name, item)
			if err != nil {
				return err
			}
			mu.Lock()
			leafs = append(leafs, l)
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return leafs, nil
}

func parseGlobal(fileData []byte) (map[string][]byte, error) {
	prevNl := -1
	name := ""
	items := make(map[string][]byte)

	for i := 0; i < len(fileData); i++ {
		char := fileData[i]
		switch char {
		case charParanthesesOpen:
			name = strings.TrimSpace(string(fileData[prevNl+1 : i]))
			if name == globalName {
				return nil, fmt.Errorf("using reserved leaf name: %s", name)
			}
			closePos := match(fileData[i+1:], charParanthesesOpen, charParanthesesClosed)
			if name == "" {
				name = globalName
			}
			_, exists := items[name]
			if exists {
				return nil, fmt.Errorf("duplicate name: %s", name)
			}
			items[name] = fileData[i+1 : i+closePos]
			name = ""
			i = i + closePos + 1
		case charNewLine:
			prevNl = i
		}
	}
	return items, nil
}

type cellType int

const (
	cellRequest cellType = iota
	cellHeaders
	cellBody
	cellPreCode
	cellPostCode
	cellDeps
)

type cellParser struct {
	t         cellType
	rgxIdent  *regexp.Regexp
	parseFunc func([]byte, *Leaf) error
}

var cellParsers = []cellParser{
	{
		t:         cellRequest,
		rgxIdent:  regexp.MustCompile(`(?m)do\s?{`),
		parseFunc: parseRequest,
	},
	{
		t:         cellHeaders,
		rgxIdent:  regexp.MustCompile(`(?m)headers\s?{`),
		parseFunc: parseHeaders,
	},
	{
		t:         cellBody,
		rgxIdent:  regexp.MustCompile(`(?m)body?{`),
		parseFunc: parseBody,
	},
	{
		t:         cellPreCode,
		rgxIdent:  regexp.MustCompile(`(?m)before?{`),
		parseFunc: parsePreCode,
	},
	{
		t:         cellPostCode,
		rgxIdent:  regexp.MustCompile(`(?m)after?{`),
		parseFunc: parsePostCode,
	},
	{
		t:         cellDeps,
		rgxIdent:  regexp.MustCompile(`(?m)depends_on?{`),
		parseFunc: parseDeps,
	},
}

func parseLeaf(name string, leafData []byte) (*Leaf, error) {
	l := &Leaf{
		name: name,
	}

	for _, parser := range cellParsers {
		loc := parser.rgxIdent.FindIndex(leafData)
		if loc == nil {
			continue
		}
		endPos := match(leafData[loc[1]+1:], charCurlyBraceOpen, charCurlyBraceClosed)
		result := leafData[loc[1] : loc[1]+endPos]
		if parser.parseFunc == nil {
			continue
		}
		if err := parser.parseFunc(result, l); err != nil {
			return nil, err
		}
	}
	return l, nil
}

func match(content []byte, opener, closer byte) int {
	pos := 0
	c := 1
	for c > 0 {
		char := content[pos]
		switch char {
		case opener:
			c++
		case closer:
			c--
		}
		pos++
	}
	return pos
}

func parseRequest(request []byte, l *Leaf) error {
	requestLineParts := bytes.Split(bytes.TrimSpace(request), []byte(" "))
	if len(requestLineParts) != 2 {
		return errors.New("invalid request line format")
	}
	l.request.method = string(requestLineParts[0])
	l.request.url = string(requestLineParts[1])
	return nil
}

func parseHeaders(headers []byte, l *Leaf) error {
	headers = bytes.TrimSpace(headers)
	headerLines := bytes.Split(headers, []byte("\n"))
	headerMap := make(map[string]string)
	for _, headerLine := range headerLines {
		headerLineParts := bytes.Split(headerLine, []byte(":"))
		if len(headerLineParts) != 2 {
			return fmt.Errorf("invalid head line format: %s", string(headerLine))
		}
		headerMap[string(bytes.TrimSpace(headerLineParts[0]))] = string(bytes.TrimSpace(headerLineParts[1]))
	}
	l.request.headers = headerMap
	return nil
}

func parseBody(body []byte, l *Leaf) error {
	l.request.body = body
	return nil
}

func parsePreCode(preCode []byte, l *Leaf) error {
	l.preCode = bytes.TrimSpace(preCode)
	return nil
}

func parsePostCode(postCode []byte, l *Leaf) error {
	l.postCode = bytes.TrimSpace(postCode)
	return nil
}

func parseDeps(deps []byte, l *Leaf) error {
	deps = bytes.TrimSpace(deps)
	depsParts := bytes.Split(deps, []byte(","))
	var dependsOn []string
	for _, depsPart := range depsParts {
		dependsOn = append(dependsOn, string(bytes.TrimSpace(depsPart)))
	}
	l.dependsOn = dependsOn
	return nil
}
