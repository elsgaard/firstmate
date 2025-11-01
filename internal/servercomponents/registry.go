// internal/servercomponents/registry.go
package servercomponents

import (
	"github.com/elsgaard/firstmate/internal/servercomponents/F5Exporter"
	//"github.com/elsgaard/firstmate/internal/servercomponents/Netapp"
)

var Registry = map[string]func() Component{
	"f5exporter": func() Component { return F5Exporter.Model{} },
	//etapp":     func() Component { return Netapp.Model{} },
}
