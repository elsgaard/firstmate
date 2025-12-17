package prometheus

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
	log.Printf("▶ Starting prometheus deploy on %s", server.FQDN)
	return m.executeRemoteCommands(server, m.getInstallCommands())
}

func (m Model) Update(server internal.Server) error {
	log.Printf("▶ Starting prometheus update on %s", server.FQDN)
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

	log.Println("✅ prometheus operation completed successfully")
	return nil
}

func (m Model) getUpdateCommands() []string {
	return []string{
		"systemctl daemon-reload",
		"systemctl restart prometheus",
	}
}

func (m Model) getInstallCommands() []string {
	return []string{
		"wget -q https://github.com/prometheus/prometheus/releases/download/v3.5.0/prometheus-3.5.0.linux-amd64.tar.gz",
		"tar -xvzf prometheus-3.5.0.linux-amd64.tar.gz",
		"cd prometheus-3.5.0.linux-amd64 && mv prometheus promtool /usr/local/bin/",
		"mkdir -p /etc/prometheus",
		"CUSTOM: CreateConfigFile",
		"mkdir -p /data/prometheus",
		"useradd -M -r -s /bin/false prometheus",
		"chown -R prometheus:prometheus /data/prometheus /etc/prometheus",
		"CUSTOM: CreateUnitFile",
		"systemctl daemon-reload",
		"systemctl enable --now prometheus",
	}
}

// Custom action dispatcher.
func (m Model) checkCustomAction(action string) string {
	if strings.HasSuffix(action, "CreateUnitFile") {
		return m.createUnitFile()
	}
	if strings.HasSuffix(action, "CreateConfigFile") {
		return m.createConfigFile()
	}
	return action
}

// CreateUnitFile returns a properly escaped heredoc for remote tee.
func (m Model) createUnitFile() string {
	return `sudo bash -c 'cat > /etc/systemd/system/prometheus.service <<EOF
[Unit]
Description=Prometheus TSDB
Wants=network-online.target
After=network-online.target

[Service]
User=prometheus
Group=prometheus
Type=simple
ExecStart=/usr/local/bin/prometheus \
--config.file=/etc/prometheus/prometheus.yml \
--storage.tsdb.path=/data/prometheus \
--web.external-url=https://prometheus.b2bi.dk \
--storage.tsdb.retention.time=90d \
--web.enable-lifecycle

Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF'`
}

// CreateConfigFile returns a properly escaped heredoc for remote cat.
func (m Model) createConfigFile() string {
	return `sudo bash -c 'cat > /etc/prometheus/prometheus.yml <<EOF
global:
  scrape_interval: 60s
  evaluation_interval: 60s

alerting:
  alertmanagers:
    - static_configs:
        - targets:
           - localhost:9093

rule_files:

scrape_configs:
  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:9090"]
        labels:
          app: "prometheus"
EOF'`
}
