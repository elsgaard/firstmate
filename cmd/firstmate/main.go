package main

import (
	models "github.com/elsgaard/firstmate/internal"
	"github.com/elsgaard/firstmate/internal/job"
	"github.com/elsgaard/firstmate/internal/servercomponents/glec"
)

func main() {

	job.Reset()
	job.Start()

	server := models.Server{}
	server.ID = 1
	server.FQDN = "localhost"
	server.User = "thel"
	server.Pass = "Filofaxx"

	component := glec.GlecModel{}
	//component.Update(server)
	component.Deploy(server)

}
