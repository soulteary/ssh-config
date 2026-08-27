package fn

import "github.com/soulteary/ssh-config/v3/pkg/sshconfig"

func validateLegacyYAML(data []byte) error {
	return sshconfig.ValidateLegacyYAML(data)
}
