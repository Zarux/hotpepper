package runner

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/fatih/color"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/zarux/hotpepper/pkg/parse"
)

type Options struct {
	HttpClient *http.Client
	Leaf       *parse.Leaf
	Globals    map[string]any
	Gmu        *sync.Mutex
	Silent     bool
}

var (
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
)

func Run(ctx context.Context, opts Options) error {
	locals := make(map[string]any)
	var (
		req  *http.Request
		resp *http.Response
		err  error
	)

	l := opts.Leaf

	if !opts.Silent {
		fmt.Printf("-------\nRunning: %s\n", yellow(l.Name))
	}

	if l.Request.Method != "" {
		req, err = http.NewRequestWithContext(ctx, l.Request.Method, l.Request.URL, bytes.NewReader(l.Request.Body))
		if err != nil {
			return err
		}
	}

	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)

	code, ok := getSideCode(l.PreCode, l.PostCode, l.Imports)
	if ok {
		if _, err := i.EvalWithContext(ctx, code); err != nil {
			return err
		}
	}

	if l.PreCode != "" {
		v, err := i.Eval("sidecode." + preCodeFunc)
		if err != nil {
			return err
		}
		exec := v.Interface().(func(request *http.Request, globals, locals map[string]any, gmu *sync.Mutex) error)
		if err := exec(req, opts.Globals, locals, opts.Gmu); err != nil {
			return err
		}
	}

	localTemplate := parse.NewTemplate(locals)
	globalTemplate := parse.NewTemplate(opts.Globals)

	if err = l.ExpandTemplate(localTemplate, false); err != nil {
		return err
	}

	if err = l.ExpandTemplate(globalTemplate, true); err != nil {
		return err
	}

	if req != nil {
		req.Method = l.Request.Method
		req.URL, err = url.Parse(l.Request.URL)
		if err != nil {
			return err
		}
		fmt.Printf("%s %s\n", blue(req.Method), blue(req.URL.String()))
		if !opts.Silent {
			fmt.Print(blue("headers{\n"))
		}
		for key, val := range l.Request.Headers {
			req.Header.Add(key, val)
			if !opts.Silent {
				fmt.Printf("  %s: %s\n", blue(key), blue(val))
			}
		}
		if !opts.Silent {
			fmt.Print(blue("}\n"))
		}
		resp, err = opts.HttpClient.Do(req)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", green(resp.Status))
	}

	if l.PostCode != "" {
		v, err := i.Eval("sidecode." + postCodeFunc)
		if err != nil {
			return err
		}

		exec := v.Interface().(func(response *http.Response, globals, locals map[string]any, gmu *sync.Mutex) error)
		if err := exec(resp, opts.Globals, locals, opts.Gmu); err != nil {
			if !opts.Silent {
				fmt.Println(red("Post Check failed X"))
			}
			return err
		}
		if !opts.Silent {
			fmt.Println(green("Post Check passed ✓"))
		}
	}

	return nil
}
