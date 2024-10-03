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
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
)

func Run() {
	data, err := os.ReadFile("./t.pepper")
	if err != nil {
		panic(err)
	}

	leafs, err := parse.Parse(data)
	if err != nil {
		panic(err)
	}

	fmt.Println(leafs["doStuffEvenLater"])

	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}

	globals := make(map[string]any)
	globalMutex := sync.Mutex{}

	ctx := context.Background()

	/* order := []string{
		"global",
		"doStuff",
		"doStuffLater",
		"doStuffEvenLater",
	} */
	order := []string{
		"do1",
		"do2",
	}

	for _, leafName := range order {
		leaf, ok := leafs[leafName]
		if !ok {
			continue
		}
		err = runner.Run(ctx, runner.Options{
			HttpClient: &httpClient,
			Globals:    globals,
			Gmu:        &globalMutex,
			Leaf:       leaf,
			Silent:     leaf.Name == "global",
		})
		if err != nil {
			fmt.Printf("%s with error: %s\n", red("FAILED"), err.Error())
			if leaf.Mandatory {
				fmt.Printf("%s must pass. %s\n", yellow(leafName), red("EXITING"))
				break
			}
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
