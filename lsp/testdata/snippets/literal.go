package snippets

import (
	"github.com/kevinswiber/languageserver-go/lsp/signature"
	t "github.com/kevinswiber/languageserver-go/lsp/types"
)

type structy struct {
	x signature.MyType
}

func X(_ map[signature.Alias]t.CoolAlias) (map[signature.Alias]t.CoolAlias) {
	return nil
}

func _() {
	X() //@signature(")", "X(_ map[signature.Alias]t.CoolAlias) map[signature.Alias]t.CoolAlias", 0)
	_ = signature.MyType{} //@item(literalMyType, "signature.MyType{}", "", "var")
	s := structy{
		x: //@snippet(" //", literalMyType, "signature.MyType{\\}", "signature.MyType{\\}")
	}
}