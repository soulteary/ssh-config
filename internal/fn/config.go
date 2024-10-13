package fn

import (
	Define "github.com/soulteary/ssh-config/internal/define"
)

func FindGlobalConfig(configs []Define.HostConfig) (result []Define.HostConfig) {
	for _, config := range configs {
		if config.Name == "*" {
			result = append(result, config)
		}
	}
	return result
}

func FindNormalConfig(configs []Define.HostConfig) (result []Define.HostConfig) {
	for _, config := range configs {
		if config.Name != "*" {
			result = append(result, config)
		}
	}
	return result
}
