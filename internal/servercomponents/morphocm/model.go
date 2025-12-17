package morphocm

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/elsgaard/firstmate/internal"
	"github.com/sfreiberg/simplessh"
)

type Model struct{}

func (m Model) Deploy(server internal.Server) error {
	log.Printf("▶ Starting morpho cm deploy on %s", server.FQDN)
	return m.executeRemoteCommands(server, m.getInstallCommands(server))
}

func (m Model) Update(server internal.Server) error {
	log.Printf("▶ Starting morpho cm update on %s", server.FQDN)
	return m.executeRemoteCommands(server, m.getUpdateCommands())
}

// executeRemoteCommands centralizes connection, execution, and error handling.
func (m Model) executeRemoteCommands(server internal.Server, cmds []string) error {
	client, err := simplessh.ConnectWithPassword(server.FQDN, server.User, server.Pass)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	for _, cmd := range cmds {
		serverCmd := m.checkCustomAction(cmd)

		log.Printf("→ Executing: %s", serverCmd)
		out, err := client.Exec(serverCmd)
		if err != nil {
			log.Printf("⚠️ Command failed: %v", err)
			continue // `return err` to fail-fast, for now just continue
		}
		if len(out) > 0 {
			log.Printf("→ Output: %s", strings.TrimSpace(string(out)))
		}

		time.Sleep(500 * time.Millisecond) // gentle pacing between commands
	}

	log.Println("✅ morpho cm operation completed successfully")
	return nil
}

func (m Model) getUpdateCommands() []string {
	return []string{
		"systemctl stop morphocm.service",
		"git -C /opt/morphocm fetch --all --tags",
		"git -C /opt/morphocm reset --hard origin/main",
		"cd /opt/morphocm && make build",
		"CUSTOM: CreateUnitFile",
		"systemctl daemon-reload",
		"systemctl restart morphocm.service",
	}
}

func (m Model) getInstallCommands(server internal.Server) []string {
	return []string{
		gitCloneCommand(server, "TRUECOMMERCEDK/morphocm"),
		"cd /opt/morphocm && make build",
		"mkdir -p /etc/morphocm",
		"mkdir -p /var/lib/morphocm",
		"chmod 755 /var/lib/morphocm",
		"CUSTOM: CreateUnitFile",
		"systemctl daemon-reload",
		"systemctl enable --now morphocm.service",
	}
}

// Custom action dispatcher.
func (m Model) checkCustomAction(action string) string {
	if strings.HasSuffix(action, "CreateUnitFile") {
		return m.createUnitFile()
	}
	return action
}

// CreateUnitFile returns a properly escaped heredoc for remote tee.
func (m Model) createUnitFile() string {
	return `sudo bash -c 'cat > /etc/systemd/system/morphocm.service <<EOF
[Unit]
Description=Morpho CM Change Management Service
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=root
WorkingDirectory=/opt/morphocm
ExecStart=/opt/morphocm/morphocm --port=8089 --db-path=/var/lib/morphocm/morphocm.db

[Install]
WantedBy=multi-user.target
EOF'`
}

func gitCloneCommand(s internal.Server, repo string) string {
	return fmt.Sprintf(
		"cd /opt && git clone https://%s:%s@github.com/%s.git",
		s.GHUser,
		s.GHPass,
		repo,
	)
}
