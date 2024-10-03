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

func Parse(fileData []byte) (map[string]*Leaf, error) {
	items, err := parseGlobal(fileData)
	if err != nil {
		return nil, err
	}
	g := errgroup.Group{}
	g.SetLimit(1)

	var (
		mu sync.Mutex
	)

	leafs := make(map[string]*Leaf)

	for name, item := range items {
		name, item := name, item
		g.Go(func() error {
			l, err := parseLeaf(name, item)
			if err != nil {
				return err
			}
			mu.Lock()
			leafs[l.Name] = l
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return leafs, nil
}

func Match(content []byte, opener, closer byte) int {
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

func parseGlobal(fileData []byte) (map[string][]byte, error) {
	prevNl := -1
	name := ""
	items := make(map[string][]byte)

	for i := 0; i < len(fileData); i++ {
		char := fileData[i]
		switch char {
		case CharParanthesesOpen:
			name = strings.TrimSpace(string(fileData[prevNl+1 : i]))
			if name == globalName {
				return nil, fmt.Errorf("using reserved leaf name: %s", name)
			}
			closePos := Match(fileData[i+1:], CharParanthesesOpen, CharParanthesesClosed)
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
	cellImport
	cellMandatory
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
	{
		t:         cellImport,
		rgxIdent:  regexp.MustCompile(`(?m)import?{`),
		parseFunc: parseImport,
	},
	{
		t:         cellMandatory,
		rgxIdent:  regexp.MustCompile(`(?m)mandatory?{`),
		parseFunc: parseMandatory,
	},
}

func parseLeaf(name string, leafData []byte) (*Leaf, error) {
	l := &Leaf{
		Name: name,
	}

	fmt.Println("aaaaaa", name, string(leafData))

	for _, parser := range cellParsers {
		loc := parser.rgxIdent.FindIndex(leafData)
		if loc == nil {
			continue
		}
		if leafData[loc[1]] == charCurlyBraceClosed {
			if err := parser.parseFunc(nil, l); err != nil {
				return nil, err
			}
			return l, nil
		}
		endPos := Match(leafData[loc[1]+1:], charCurlyBraceOpen, charCurlyBraceClosed)
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

func parseRequest(request []byte, l *Leaf) error {
	requestLineParts := bytes.Split(bytes.TrimSpace(request), []byte(" "))
	if len(requestLineParts) != 2 {
		return errors.New("invalid request line format")
	}
	l.Request.Method = string(requestLineParts[0])
	l.Request.URL = string(requestLineParts[1])
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
	l.Request.Headers = headerMap
	return nil
}

func parseBody(body []byte, l *Leaf) error {
	l.Request.Body = body
	return nil
}

func parsePreCode(preCode []byte, l *Leaf) error {
	l.PreCode = string(bytes.TrimSpace(preCode))
	return nil
}

func parsePostCode(postCode []byte, l *Leaf) error {
	l.PostCode = string(bytes.TrimSpace(postCode))
	return nil
}

func parseDeps(deps []byte, l *Leaf) error {
	deps = bytes.TrimSpace(deps)
	depsParts := bytes.Split(deps, []byte(","))
	var dependsOn []string
	for _, depsPart := range depsParts {
		dependsOn = append(dependsOn, string(bytes.TrimSpace(depsPart)))
	}
	l.DependsOn = dependsOn
	return nil
}

func parseImport(imports []byte, l *Leaf) error {
	var importItems []string
	importLines := bytes.Split(imports, []byte("\n"))
	for _, importLine := range importLines {
		importItem := bytes.TrimSpace(importLine)
		importItems = append(importItems, string(importItem))
	}

	l.Imports = importItems
	return nil
}

func parseMandatory(_ []byte, l *Leaf) error {
	l.Mandatory = true
	return nil
}
