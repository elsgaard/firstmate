package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	models "github.com/elsgaard/firstmate/internal"
	"github.com/elsgaard/firstmate/internal/servercomponents"
)

func main() {
	// Load optional .env (ignored if missing)
	_ = loadDotEnv(".env")

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		run(os.Args[2:], "install")
	case "update":
		run(os.Args[2:], "update")
	default:
		fmt.Println("Unknown command:", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func run(args []string, mode string) {
	fs := flag.NewFlagSet(mode, flag.ExitOnError)

	app := fs.String("app", "", "Application name")
	host := fs.String("host", "", "Target host (e.g. server.example.com)")
	user := fs.String("user", os.Getenv("SSH_USER"), "SSH username (or SSH_USER)")
	pass := fs.String("pass", os.Getenv("SSH_PASS"), "SSH password (or SSH_PASS)")
	gh_user := fs.String("gh_user", os.Getenv("GITHUB_USER"), "github username (or GITHUB_USER)")
	gh_pass := fs.String("gh_pass", os.Getenv("GITHUB_PASS"), "github password (or GITHUB_PASS)")

	fs.Parse(args)

	require(fs,
		"--app", *app,
		"--host", *host,
		"--user", *user,
		"--pass", *pass,
	)

	server := models.Server{
		ID:     1,
		FQDN:   *host,
		User:   *user,
		Pass:   *pass,
		GHUser: *gh_user,
		GHPass: *gh_pass,
	}

	factory, ok := servercomponents.Registry[*app]
	if !ok {
		fmt.Println("Unknown app:", *app)
		os.Exit(3)
	}

	component := factory()

	var err error
	switch mode {
	case "install":
		err = component.Deploy(server)
	case "update":
		err = component.Update(server)
	}

	if err != nil {
		fmt.Printf("%s failed: %v\n", strings.Title(mode), err)
		os.Exit(4)
	}
}

func require(fs *flag.FlagSet, pairs ...string) {
	var missing []string

	for i := 0; i < len(pairs); i += 2 {
		name, value := pairs[i], pairs[i+1]
		if value == "" {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		fmt.Println("Error: missing required flags:", strings.Join(missing, ", "))
		fs.Usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println(`Usage:
  firstmate install [flags]
  firstmate update  [flags]

Flags:
  --app   Application name
  --host  Target host
  --user  SSH user (or SSH_USER)
  --pass  SSH password (or SSH_PASS)
`)
}

func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		val = stripQuotes(strings.TrimSpace(val))

		// Do not override existing environment variables
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, val); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}

func stripQuotes(v string) string {
	if len(v) >= 2 {
		if (v[0] == '"' && v[len(v)-1] == '"') ||
			(v[0] == '\'' && v[len(v)-1] == '\'') {
			return v[1 : len(v)-1]
		}
	}
	return v
}
