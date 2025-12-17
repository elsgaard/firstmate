package nodeexp

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
	log.Printf("▶ Starting Node Exporter deploy on %s", server.FQDN)
	return m.executeRemoteCommands(server, m.getInstallCommands())
}

func (m Model) Update(server internal.Server) error {
	log.Printf("▶ Starting Node Exporter update on %s", server.FQDN)
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

	log.Println("✅ Node Exporter operation completed successfully")
	return nil
}

func (m Model) getUpdateCommands() []string {
	return []string{
		"CUSTOM: CreateUnitFile",
		"systemctl daemon-reload",
		"systemctl restart prometheus",
	}
}

func (m Model) getInstallCommands() []string {
	return []string{
		"wget -q https://github.com/prometheus/node_exporter/releases/download/v1.10.2/node_exporter-1.10.2.linux-amd64.tar.gz",
		"tar -xvf node_exporter-1.10.2.linux-amd64.tar.gz",
		"cd node_exporter-1.10.2.linux-amd64 && mv node_exporter /usr/local/bin/",
		"CUSTOM: CreateUnitFile",
		"systemctl daemon-reload",
		"systemctl enable --now node_exporter.service",
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
	return `sudo bash -c 'cat > /etc/systemd/system/node_exporter.service <<EOF
[Unit]
Description=Node Exporter
Wants=network-online.target
After=network-online.target
StartLimitIntervalSec=500
StartLimitBurst=5
[Service]
Type=simple
Restart=on-failure
RestartSec=5s
ExecStart=/usr/local/bin/node_exporter --collector.logind --collector.systemd --web.listen-address=:9182
[Install]
WantedBy=multi-user.target
EOF'`
}
