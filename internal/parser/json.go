/**
 * Copyright 2024-2025 Su Yang (soulteary)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
