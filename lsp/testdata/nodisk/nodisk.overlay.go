package nodisk

import (
	"github.com/kevinswiber/languageserver-go/lsp/foo"
)

func _() {
	foo.Foo() //@complete("F", Foo, IntFoo, StructFoo)
}
