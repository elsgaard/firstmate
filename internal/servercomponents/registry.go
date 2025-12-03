// internal/servercomponents/registry.go
package servercomponents

import (
	"github.com/elsgaard/firstmate/internal/servercomponents/F5Exporter"
	"github.com/elsgaard/firstmate/internal/servercomponents/alerthistory"
	"github.com/elsgaard/firstmate/internal/servercomponents/alertmanager"
	"github.com/elsgaard/firstmate/internal/servercomponents/edicheck"
	"github.com/elsgaard/firstmate/internal/servercomponents/morphocm"
	"github.com/elsgaard/firstmate/internal/servercomponents/prometheus"
	"github.com/elsgaard/firstmate/internal/servercomponents/sftrip"
)

var Registry = map[string]func() Component{
	"f5exporter":   func() Component { return F5Exporter.Model{} },
	"alerthistory": func() Component { return alerthistory.Model{} },
	"morphocm":     func() Component { return morphocm.Model{} },
	"alertmanager": func() Component { return alertmanager.Model{} },
	"prometheus":   func() Component { return prometheus.Model{} },
	"sftrip":       func() Component { return sftrip.Model{} },
	"edicheck":     func() Component { return edicheck.Model{} },
}
