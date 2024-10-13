package parser

import (
	"fmt"
	"reflect"
	"strings"

	Define "github.com/soulteary/ssh-yaml/internal/define"
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

	lines := strings.Split(input, "\n")
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			currentComments = append(currentComments, strings.TrimSpace(line[1:]))
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
	Comments string
	Config   string
}

func GetSSHConfigContent(host string, input SSHHostConfigGroup) (config SSHHostConfigGrouped) {
	config.Comments = strings.Join(input.Comments, "\n")

	var lines []string

	configs := Fn.GetOrderMaps(input.Config)
	for _, field := range configs.Keys {
		lines = append(lines, fmt.Sprintf("    %s %s", field, configs.Data[field]))
	}
	config.Config = strings.Join(lines, "\n")
	config.Config = "Host " + host + "\n" + config.Config
	return config
}

func ParseSSHConfig(input string, notes string) (config HostConfig) {
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
			case "include":
				config.Include = value
			case "pubkeyacceptedalgorithms":
				config.PubkeyAcceptedAlgorithms = value
			default:
				fmt.Println("Unknown key", key)
			}
		}
	}

	if notes != "" {
		config.YamlUserNotes = notes
	}
	return config
}

func GetSingleHostData(input HostConfig) (result map[string]string, name string, notes string) {
	v := reflect.ValueOf(input)
	t := v.Type()

	config := make(map[string]string)
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		tag := field.Tag.Get("yaml")
		if tag == "" {
			tag = field.Name
		} else {
			tag = strings.Split(tag, ",")[0]
		}

		if !value.IsZero() {
			val := value.Interface().(string)
			if val != "" {
				config[tag] = val
			}
		}
	}

	name = config["YamlUserHost"]
	notes = config["YamlUserNotes"]
	delete(config, "YamlUserHost")
	delete(config, "YamlUserNotes")

	return config, name, notes
}

func ConvertToSSH(hostConfigs []Define.HostConfig) []byte {
	lines := make([]string, 0)

	globalConfigs := Fn.FindGlobalConfig(hostConfigs)
	if len(globalConfigs) > 0 {
		for _, config := range globalConfigs {
			if config.Notes != "" {
				lines = append(lines, fmt.Sprintf("# %s", config.Notes))
			}
			lines = append(lines, fmt.Sprintf("Host %s", config.Name))

			orderMaps := Fn.GetOrderMaps(config.Config)
			for _, field := range orderMaps.Keys {
				lines = append(lines, fmt.Sprintf("    %s %s", field, orderMaps.Data[field]))
			}
		}
		lines = append(lines, "")
	}

	normalConfigs := Fn.FindNormalConfig(hostConfigs)
	if len(normalConfigs) > 0 {
		for _, config := range normalConfigs {
			if config.Notes != "" {
				lines = append(lines, fmt.Sprintf("# %s", config.Notes))
			}
			lines = append(lines, fmt.Sprintf("Host %s", config.Name))
			orderMaps := Fn.GetOrderMaps(config.Config)
			for _, field := range orderMaps.Keys {
				lines = append(lines, fmt.Sprintf("    %s %s", field, orderMaps.Data[field]))
			}
			lines = append(lines, "")
		}
		lines = append(lines, "")
	}
	return []byte(strings.Join(lines, "\n"))
}
