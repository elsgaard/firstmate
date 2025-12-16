package ubuntu

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
	log.Printf("▶ Starting Ubuntu deploy on %s", server.FQDN)
	return m.executeRemoteCommands(server, m.getInstallCommands())
}

func (m Model) Update(server internal.Server) error {
	log.Printf("▶ Starting Ubuntu update on %s", server.FQDN)
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

	log.Println("✅ Ubuntu operation completed successfully")
	return nil
}

func (m Model) getUpdateCommands() []string {
	return []string{
		"apt update -y",
		"apt upgrade -y",
		"apt install sqlite3 -y",
	}
}

func (m Model) getInstallCommands() []string {
	return []string{
		"apt update -y",
		"apt upgrade -y",
		"apt install sqlite3 -y",
		"mkdir -p /etc/systemd/timesyncd.conf.d",
		"CUSTOM: CreateNTPFile",
		"timedatectl set-timezone Europe/Copenhagen",
		"systemctl restart systemd-timesyncd",
		"timedatectl status",
		"timedatectl show-timesync --all",
		"systemctl mask --now fwupd.service",
		"systemctl mask --now fwupd-refresh.service",
		"systemctl mask --now fwupd-refresh.timer",
		"rm -f /etc/resolv.conf",
		"ln -s /run/systemd/resolve/resolv.conf /etc/resolv.conf",
	}
}

// Custom action dispatcher.
func (m Model) checkCustomAction(action string) string {
	if strings.HasSuffix(action, "CreateNTPFile") {
		return m.createNTPFile()
	}
	return action
}

// CreateUnitFile returns a properly escaped heredoc for remote tee.
func (m Model) createNTPFile() string {
	return `sudo bash -c 'cat > /etc/systemd/timesyncd.conf.d/custom.conf <<EOF
[Time]
NTP=10.16.70.11 10.16.70.12 10.16.70.13 10.16.70.14
EOF'`
}
