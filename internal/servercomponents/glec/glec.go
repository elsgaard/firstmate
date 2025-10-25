package glec

import (
	"fmt"
	"strings"

	models "github.com/elsgaard/firstmate/internal"
	"github.com/elsgaard/firstmate/internal/job"
	"github.com/sfreiberg/simplessh"
)

type GlecModel struct{}

func (m GlecModel) Deploy(server models.Server) {

	fmt.Println("Component GLEC is executing")
	job.SetStatusText("Firstmate job is started")

	var client *simplessh.Client
	var err error

	if client, err = simplessh.ConnectWithPassword(server.FQDN, server.User, server.Pass); err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}

	defer client.Close()

	for _, s := range m.getInstallCommands() {
		serverCmd := m.checkCustomAction(s)
		out, err := client.Exec(serverCmd)
		if err != nil {
			fmt.Println("Error executing command:", err)
		}
		job.SetStatusText(s + "\n" + string(out))
	}

	fmt.Println("Component GLEC has completed")
	job.Stop()

}

func (m GlecModel) Update(server models.Server) {

	fmt.Println("Component GLEC is executing")
	job.SetStatusText("Firstmate job is started")

	var client *simplessh.Client
	var err error

	if client, err = simplessh.ConnectWithPassword(server.FQDN, server.User, server.Pass); err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}

	defer client.Close()

	for _, s := range m.getUpdateCommands() {
		serverCmd := m.checkCustomAction(s)
		out, err := client.Exec(serverCmd)
		if err != nil {
			fmt.Println("Error executing command:", err)
		}
		job.SetStatusText(s + "\n" + string(out))
	}

	fmt.Println("Component GLEC has completed")
	job.Stop()

}

func (m GlecModel) getUpdateCommands() []string {

	var serverCommand []string
	serverCommand = append(serverCommand, "git -C /opt/glec/ fetch origin main")
	serverCommand = append(serverCommand, "git -C /opt/glec/ reset --hard origin/main")
	serverCommand = append(serverCommand, "cd /opt/glec && make build")

	//serverCommand = append(serverCommand, "CUSTOM: CreateEnvFile")
	//serverCommand = append(serverCommand, "docker compose -f /var/msdp.b2bi.dk/docker-compose.yml pull --quiet")
	//serverCommand = append(serverCommand, "docker compose -f /var/msdp.b2bi.dk/docker-compose.yml down")
	//serverCommand = append(serverCommand, "docker compose -f /var/msdp.b2bi.dk/docker-compose.yml up -d --remove-orphans")
	return serverCommand

}

func (m GlecModel) getInstallCommands() []string {

	var serverCommand []string
	serverCommand = append(serverCommand, "cd /opt && sudo git clone https://github.com/elsgaard/glec.git")
	serverCommand = append(serverCommand, "cd /opt/glec && sudo make build")
	serverCommand = append(serverCommand, "CUSTOM: CreateUnitFile")
	serverCommand = append(serverCommand, "sudo systemctl start glec.service")
	return serverCommand

}

func (n GlecModel) checkCustomAction(action string) string {

	if strings.HasSuffix(action, "CreateUnitFile") {
		return n.createUnitFile()
	}
	return action

}

func (nse GlecModel) createUnitFile() string {
	return "sudo tee /etc/systemd/system/glec.service <<EOF\n" +
		"[Unit]\n" +
		"Description=Go Leader Election Component Service\n" +
		"After=network.target\n" +
		"StartLimitIntervalSec=0\n" +
		"[Service]\n" +
		"Type=simple\n" +
		"Restart=always\n" +
		"RestartSec=1\n" +
		"User=root\n" +
		"ExecStart=/opt/glec/glec\n" +
		"\n" +
		"[Install]\n" +
		"WantedBy=multi-user.target\n" +
		"EOF\n"
}
