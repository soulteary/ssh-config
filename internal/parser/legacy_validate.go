package parser

import (
	"fmt"
	"sort"
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

// validateLegacyGlobalsRepresentable rejects input whose global hosts cannot be
// expressed by the legacy YAML view. That view has a single "global" mapping, so
// two "*" entries setting the same keyword to different values would lose one of
// them no matter which is kept, and for an accumulating directive such as
// IdentityFile OpenSSH would have used both. The text path already refuses the
// equivalent input as a duplicate Host block; this closes the same gap for JSON
// and direct callers. Only the YAML rendering is constrained: -to-ssh and
// -to-json can represent the entries separately and stay available.
func validateLegacyGlobalsRepresentable(configs []Define.HostConfig) error {
	seen := make(map[string]string)
	for _, host := range configs {
		if host.Extra.Prefix+host.Name != "*" {
			continue
		}
		keys := make([]string, 0, len(host.Config))
		for key := range host.Config {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			value := host.Config[key]
			previous, exists := seen[key]
			if exists && previous != value {
				return fmt.Errorf(
					"legacy conversion would lose data: global directive %q has conflicting values %q and %q; retry without -legacy",
					key, previous, value)
			}
			seen[key] = value
		}
	}
	return nil
}
