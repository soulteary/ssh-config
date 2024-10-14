package parser

import (
	Define "github.com/soulteary/ssh-config/internal/define"
	Fn "github.com/soulteary/ssh-config/internal/fn"
)

func ConvertToJSON(input []Define.HostConfig) []byte {
	hostConfigs := make([]Define.HostConfigForJSON, 0)
	for _, hostConfig := range input {
		var config Define.HostConfigForJSON
		config.Name = hostConfig.Name
		config.Notes = hostConfig.Notes
		config.Data = make(Define.HostConfigDataForJSON)

		orderMaps := Fn.GetOrderMaps(hostConfig.Config)
		for _, field := range orderMaps.Keys {
			config.Data[field] = orderMaps.Data[field]
		}
		hostConfigs = append(hostConfigs, config)
	}
	return Fn.GetJSONBytes(hostConfigs)
}

func GroupJSONConfig(input string) []Define.HostConfig {
	jsonConfig := Fn.GetJSONData(input)

	var hostConfigs []Define.HostConfig

	for _, hostConfig := range jsonConfig {
		var config Define.HostConfig
		config.Name = hostConfig.Name
		config.Notes = hostConfig.Notes
		config.Config = make(map[string]string)

		for key, value := range hostConfig.Data {
			config.Config[key] = value
		}
		hostConfigs = append(hostConfigs, config)
	}

	return hostConfigs
}
