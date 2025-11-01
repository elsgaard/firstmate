package main

import (
	"flag"
	"fmt"
	"os"

	models "github.com/elsgaard/firstmate/internal"
	"github.com/elsgaard/firstmate/internal/servercomponents"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: deploy <install|update> [flags]")
		os.Exit(1)
	}

	cmd := os.Args[1]

	app := flag.String("app", "", "Name of the application to install or update")
	host := flag.String("host", "", "Target host (e.g. server.example.com)")
	userFlag := flag.String("user", "", "SSH username")
	passFlag := flag.String("pass", "", "SSH password")

	flag.CommandLine.Parse(os.Args[2:])

	if *app == "" || *host == "" || *userFlag == "" || *passFlag == "" {
		fmt.Println("Error: missing required flags: --app, --host, --user, --pass")
		os.Exit(2)
	}

	server := models.Server{}
	server.ID = 1
	server.FQDN = *host
	server.User = *userFlag
	server.Pass = *passFlag

	factory, ok := servercomponents.Registry[*app]
	if !ok {
		fmt.Printf("Unknown app: %s\n", *app)
		os.Exit(3)
	}

	component := factory()

	switch cmd {
	case "install":
		if err := component.Deploy(server); err != nil {
			fmt.Println("Deploy failed:", err)
			os.Exit(4)
		}
	case "update":
		if err := component.Update(server); err != nil {
			fmt.Println("Update failed:", err)
			os.Exit(5)
		}
	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(6)
	}

}
