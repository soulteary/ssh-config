package parser

import (
	"bufio"
	"strings"
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
