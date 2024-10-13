package parser

import (
	"bufio"
	"fmt"
	"strings"

	Fn "github.com/soulteary/ssh-yaml/internal/fn"
)

type SSHHostConfigGroup struct {
	Comments []string
	Config   map[string]string
}

func GroupSSHConfig(input string) map[string]SSHHostConfigGroup {
	hostConfigs := make(map[string]SSHHostConfigGroup)
	var currentHost string
	var currentComments []string

	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			currentComments = append(currentComments, line)
		} else if strings.HasPrefix(line, "Host ") {
			currentHost = strings.TrimSpace(strings.TrimPrefix(line, "Host "))
			hostConfigs[currentHost] = SSHHostConfigGroup{
				Comments: currentComments,
				Config:   make(map[string]string),
			}
			currentComments = nil
		} else {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				hostConfigs[currentHost].Config[key] = value
			}
		}
	}

	return hostConfigs
}

type SSHHostConfigGrouped struct {
	Comments []string
	Config   string
}

func GetSSHConfigContent(host string, input SSHHostConfigGroup) (config SSHHostConfigGrouped) {
	config.Comments = input.Comments
	var lines []string

	configs := Fn.GetOrderMaps(input.Config)

	for key, value := range configs {
		lines = append(lines, fmt.Sprintf("    %s %s", key, value))
	}
	config.Config = strings.Join(lines, "\n")
	config.Config = "Host " + host + "\n" + config.Config
	return config
}

func ParseSSHConfig(input string, notes []string) (config HostConfig) {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			key := strings.ToLower(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "hostname":
				config.HostName = value
			case "user":
				config.User = value
			case "identityfile":
				config.IdentityFile = value
			case "port":
				config.Port = value
			case "controlpath":
				config.ControlPath = value
			case "controlpersist":
				config.ControlPersist = value
			case "tcpkeepalive":
				config.TCPKeepAlive = value
			case "compression":
				config.Compression = value
			case "forwardagent":
				config.ForwardAgent = value
			case "ciphers":
				config.Ciphers = value
			case "hostkeyalgorithms":
				config.HostKeyAlgorithms = value
			case "kexalgorithms":
				config.KexAlgorithms = value
			case "pubkeyauthentication":
				config.PubkeyAuthentication = value
			case "proxycommand":
				config.ProxyCommand = value
			case "host":
				config.YamlUserHost = value
			default:
				fmt.Println("Unknown key", key)
			}
		}
	}

	if len(notes) > 0 {
		config.YamlUserNotes = strings.Join(notes, "\n")
	}
	return config
}
