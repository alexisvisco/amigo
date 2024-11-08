package amigo

import (
	"github.com/alexisvisco/amigo/pkg/amigoctx"
	"github.com/alexisvisco/amigo/pkg/types"
)

type Amigo struct {
	ctx    *amigoctx.Context
	Driver types.Driver
}

// NewAmigo create a new amigo instance
func NewAmigo(ctx *amigoctx.Context) Amigo {
	return Amigo{
		ctx:    ctx,
		Driver: types.GetDriver(ctx.DSN),
	}
}
