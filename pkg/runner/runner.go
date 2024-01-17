package runner

import (
	"context"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/zarux/hotpepper/pkg/parse"
)

func Run(l parse.Leaf) {
	ctx := context.Background()
	i := interp.New(interp.Options{})

	i.Use(stdlib.Symbols)

	_, err := i.EvalWithContext(ctx, `import "fmt"`)
	if err != nil {
		panic(err)
	}

	_, err = i.EvalWithContext(ctx, `fmt.Println("Hello Yaegi")`)
	if err != nil {
		panic(err)
	}
}
