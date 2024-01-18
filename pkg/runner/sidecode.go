package runner

import (
	"fmt"
	"strings"
)

const (
	preCodeFunc  = "ExecPreCode"
	postCodeFunc = "ExecPostCode"
)

var preDefinedImports = []string{
	"\"net/http\"",
	"\"sync\"",
	"\"fmt\"",
}

const sideCodeTemplate = `
			package sidecode

			import (
				%s
			)

			func %s(request *http.Request, globals, locals map[string]any, gmu *sync.Mutex) error {
				%s
				return nil
			}
			
			func %s(response *http.Response, globals, locals map[string]any, gmu *sync.Mutex) error {
				%s
				return nil
			}
		`

func getSideCode(preCode, postCode string, imports []string) (string, bool) {
	return fmt.Sprintf(
		sideCodeTemplate,
		getImportString(imports),
		preCodeFunc, preCode,
		postCodeFunc, postCode,
	), preCode != "" || postCode != ""
}

func getImportString(imports []string) string {
	imports = append(imports, preDefinedImports...)
	importSet := make(map[string]struct{})
	var finalImports []string
	for _, item := range imports {
		if _, value := importSet[item]; !value {
			importSet[item] = struct{}{}
			finalImports = append(finalImports, item)
		}
	}

	return strings.Join(finalImports, "\n")
}
