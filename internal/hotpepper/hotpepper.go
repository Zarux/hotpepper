package hotpepper

import (
	"fmt"
	"os"

	"github.com/zarux/hotpepper/pkg/parse"
	"github.com/zarux/hotpepper/pkg/runner"
)

func Run() {
	data, err := os.ReadFile("./test.pepper")
	if err != nil {
		panic(err)
	}

	leafs, err := parse.Parse(data)
	if err != nil {
		panic(err)
	}

	for _, leaf := range leafs {
		fmt.Printf("%+v\n", leaf)
	}

	runner.Run(*leafs[0])
}
