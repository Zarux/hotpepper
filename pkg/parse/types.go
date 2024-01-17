package parse

type Leaf struct {
	name    string
	request struct {
		method  string
		url     string
		headers map[string]string
		body    []byte
	}
	preCode   []byte
	postCode  []byte
	dependsOn []string
}

const globalName = "global"

const (
	charNewLine           = 0xa  // \n
	charParanthesesOpen   = 0x28 // (
	charParanthesesClosed = 0x29 // )
	charCurlyBraceOpen    = 0x7b // {
	charCurlyBraceClosed  = 0x7d // }
)
