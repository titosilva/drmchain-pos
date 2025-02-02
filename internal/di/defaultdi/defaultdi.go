package defaultdi

import (
	"github.com/titosilva/drmchain-pos/internal/di"
	identityprovider "github.com/titosilva/drmchain-pos/internal/shared/identity_provider"
	"github.com/titosilva/drmchain-pos/storage/localstorage"
)

func ConfigureDefaultDI() *di.DIContext {
	diCtx := di.NewContext()

	di.AddInterfaceFactory(diCtx, localstorage.Factory)
	di.AddFactory(diCtx, identityprovider.Factory)

	return diCtx
}
