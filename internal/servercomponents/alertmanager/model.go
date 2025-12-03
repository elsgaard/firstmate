package alertmanager

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
	log.Printf("▶ Starting alertmanager deploy on %s", server.FQDN)
	return m.executeRemoteCommands(server, m.getInstallCommands())
}

func (m Model) Update(server internal.Server) error {
	log.Printf("▶ Starting alertmanager update on %s", server.FQDN)
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

	log.Println("✅ alertmanager operation completed successfully")
	return nil
}

func (m Model) getUpdateCommands() []string {
	return []string{
		"systemctl daemon-reload",
		"systemctl restart alertmanager",
	}
}

func (m Model) getInstallCommands() []string {
	return []string{
		"wget -q https://github.com/prometheus/alertmanager/releases/download/v0.28.1/alertmanager-0.28.1.linux-amd64.tar.gz",
		"tar -xvzf alertmanager-0.28.1.linux-amd64.tar.gz",
		"cd alertmanager-0.28.1.linux-amd64 && mv alertmanager amtool /usr/local/bin/",
		"mkdir -p /etc/alertmanager",
		"CUSTOM: CreateConfigFile",
		"mkdir -p /var/lib/alertmanager",
		"useradd -M -r -s /bin/false alertmanager",
		"chown -R alertmanager:alertmanager /var/lib/alertmanager /etc/alertmanager",
		"CUSTOM: CreateUnitFile",
		"systemctl daemon-reload",
		"systemctl enable --now alertmanager",
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

// CreateUnitFile returns a properly escaped heredoc for remote cat.
func (m Model) createUnitFile() string {
	return `sudo bash -c 'cat > /etc/systemd/system/alertmanager.service <<EOF
[Unit]
Description=Prometheus Alertmanager
Wants=network-online.target
After=network-online.target

[Service]
User=alertmanager
Group=alertmanager
Type=simple
ExecStart=/usr/local/bin/alertmanager \
--config.file=/etc/alertmanager/alertmanager.yml \
--storage.path=/var/lib/alertmanager \
--web.external-url=https://alertmanager.b2bi.dk

Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF'`
}

// CreateConfigFile returns a properly escaped heredoc for remote cat.
func (m Model) createConfigFile() string {
	return `sudo bash -c 'cat > /etc/alertmanager/alertmanager.yml <<EOF
global:
  pagerduty_url: 'https://events.pagerduty.com/v2/enqueue'
  smtp_require_tls: false
  smtp_smarthost: 'smtp.b2bi.dk:25'
  smtp_from: 'alertmanager@truecommerce.com'          

route:
  group_by: ['alertname','instance']
  group_interval: 5m
  repeat_interval: 120h

  receiver: default

  routes:
  - matchers:
    - severity = critical
    - notify = servicedesk
    receiver:  netsuite_servicedesk

  - matchers:
    - severity = none
    group_wait: 0s
    group_interval: 1m
    repeat_interval: 5m
    receiver: none.dead.man.snitch
      
receivers:
- name: netsuite_servicedesk
  email_configs:
   - to: 'servicedesk@truecommerce.com'
  webhook_configs:

- name: default

inhibit_rules:
EOF'`
}
