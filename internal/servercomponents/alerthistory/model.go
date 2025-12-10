package alerthistory

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
	log.Printf("▶ Starting Alerthistory deploy on %s", server.FQDN)
	return m.executeRemoteCommands(server, m.getInstallCommands())
}

func (m Model) Update(server internal.Server) error {
	log.Printf("▶ Starting Alerthistory update on %s", server.FQDN)
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
			log.Printf("Output: %s", strings.TrimSpace(string(out)))
		}

		time.Sleep(500 * time.Millisecond) // gentle pacing between commands
	}

	log.Println("✅ Alerthistory operation completed successfully")
	return nil
}

func (m Model) getUpdateCommands() []string {
	return []string{
		"sudo systemctl stop alerthistory.service",
		"git -C /opt/alerthistory fetch origin main",
		"git -C /opt/alerthistory reset --hard origin/main",
		"cd /opt/alerthistory && sudo make build",
		"CUSTOM: CreateUnitFile",
		"sudo systemctl daemon-reload",
		"sudo systemctl restart alerthistory.service",
	}
}

func (m Model) getInstallCommands() []string {
	return []string{
		"cd /opt && sudo git clone https://github.com/TRUECOMMERCEDK/alerthistory.git",
		"cd /opt/alerthistory && sudo make build",
		"mkdir -p /etc/alerthistory",
		"mkdir -p /var/lib/alerthistory",
		"chmod 700 /var/lib/alerthistory",
		"CUSTOM: CreateUnitFile",
		"sudo systemctl daemon-reload",
		"sudo systemctl enable --now alerthistory.service",
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
	return `sudo bash -c 'cat > /etc/systemd/system/alerthistory.service <<EOF
[Unit]
Description=Alerthistory Service
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=root
WorkingDirectory=/opt/alerthistory
ExecStart=/opt/alerthistory/alerthistoryserver --port=8082 --db-path=/var/lib/alerthistory/alerthistory.db

[Install]
WantedBy=multi-user.target
EOF'`
}
