package danglingstmt

import "github.com/kevinswiber/languageserver-go/lsp/foo"

func _() {
	foo. //@rank(" //", Foo)
	var _ = []string{foo.} //@rank("}", Foo)
}
