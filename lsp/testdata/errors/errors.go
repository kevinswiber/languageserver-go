package errors

import (
	"github.com/kevinswiber/languageserver-go/lsp/types"
)

func _() {
	bob.Bob() //@complete(".")
	types.b //@complete(" //", Bob_interface)
}
