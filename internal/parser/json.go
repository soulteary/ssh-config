package parser

import (
	Define "github.com/soulteary/ssh-yaml/internal/define"
	Fn "github.com/soulteary/ssh-yaml/internal/fn"
)

func ConvertToJSON(input []Define.HostConfig) []byte {
	hostConfigs := make([]Define.HostConfigForJSON, 0)
	for _, hostConfig := range input {
		config := Define.HostConfigForJSON{}

		orderMaps := Fn.GetOrderMaps(hostConfig.Config)
		for _, field := range orderMaps.Keys {
			config[field] = orderMaps.Data[field]
		}
		hostConfigs = append(hostConfigs, config)
	}
	return Fn.GetJSONBytes(hostConfigs)
}
