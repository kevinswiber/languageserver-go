package other

import "github.com/kevinswiber/languageserver-go/lsp/rename/crosspkg"

func Other() {
	crosspkg.Bar
	crosspkg.Foo() //@rename("Foo", "Flamingo")
}
