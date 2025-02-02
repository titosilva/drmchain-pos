package identityhandler

import "github.com/titosilva/drmchain-pos/internal/di"

type IdentityHandler struct {
}

func DIFactory(diCtx *di.DIContext) *IdentityHandler {
	return NewIdentityHandler()
}

func NewIdentityHandler() *IdentityHandler {
	return &IdentityHandler{}
}
