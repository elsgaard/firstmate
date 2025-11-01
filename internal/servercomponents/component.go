// internal/servercomponents/component.go
package servercomponents

import (
	"github.com/elsgaard/firstmate/internal"
	_ "github.com/elsgaard/firstmate/internal"
)

type Component interface {
	Deploy(server internal.Server) error
	Update(server internal.Server) error
}
