package c

import "github.com/kevinswiber/languageserver-go/lsp/rename/b"

func _() {
	b.Hello() //@rename("Hello", "Goodbye")
}
