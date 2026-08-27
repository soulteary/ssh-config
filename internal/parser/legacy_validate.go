package parser

import (
	"fmt"
	"strings"

	Define "github.com/soulteary/ssh-config/v3/internal/define"
	"github.com/soulteary/ssh-config/v3/pkg/sshconfig"
)

func validateLegacyHostConfigs(configs []Define.HostConfig) error {
	for index, host := range configs {
		name := host.Extra.Prefix + host.Name
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("legacy host %d: host name is empty", index)
		}
		if err := sshconfig.ValidateDirectiveInput("Host", []string{name}, ""); err != nil {
			return fmt.Errorf("legacy host %d: %w", index, err)
		}
		for _, note := range strings.Split(host.Notes, "\n") {
			if err := sshconfig.ValidateDirectiveInput("Host", nil, note); err != nil {
				return fmt.Errorf("legacy host %d note: %w", index, err)
			}
		}
		for key, value := range host.Config {
			if err := sshconfig.ValidateDirectiveInput(key, []string{value}, ""); err != nil {
				return fmt.Errorf("legacy host %d directive %q: %w", index, key, err)
			}
		}
	}
	return nil
}
