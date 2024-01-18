package hotpepper

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/zarux/hotpepper/pkg/deptree"
	"github.com/zarux/hotpepper/pkg/parse"
	"github.com/zarux/hotpepper/pkg/runner"
)

var (
	red = color.New(color.FgRed).SprintFunc()
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

	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}

	globals := make(map[string]any)
	globalMutex := sync.Mutex{}

	ctx := context.Background()

	order := []string{
		"global",
		"doStuff",
		"doStuffLater",
	}

	for _, leaf := range order {
		err = runner.Run(ctx, runner.Options{
			HttpClient: &httpClient,
			Globals:    globals,
			Gmu:        &globalMutex,
			Leaf:       leafs[leaf],
			Silent:     leafs[leaf].Name == "global",
		})
		if err != nil {
			fmt.Printf("%s with error: %s\n", red("FAILED"), err.Error())
		}
	}

}

func leavesToNodes(leafs []*parse.Leaf) []*deptree.Node {
	var nodes []*deptree.Node
	for _, leaf := range leafs {
		node := deptree.NewNode(leaf.Name, nil)
		nodes = append(nodes, node)
	}
	return nodes
}
