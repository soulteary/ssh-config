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
	"strings"

	Cmd "github.com/soulteary/ssh-config/cmd"
	Define "github.com/soulteary/ssh-config/internal/define"
	Fn "github.com/soulteary/ssh-config/internal/fn"
)

func Process(fileType string, userInput string, args Cmd.Args) []byte {
	var hostConfigs []Define.HostConfig

	switch strings.ToUpper(fileType) {
	case "YAML":
		hostConfigs = GroupYAMLConfig(userInput)
	case "JSON":
		hostConfigs = GroupJSONConfig(userInput)
	case "TEXT":
		hostConfigs = GroupSSHConfig(userInput)
	}

	if args.ToYAML {
		return Fn.TidyLastEmptyLines(ConvertToYAML(hostConfigs))
	}

	if args.ToSSH {
		return Fn.TidyLastEmptyLines(ConvertToSSH(hostConfigs))
	}

	if args.ToJSON {
		return Fn.TidyLastEmptyLines(ConvertToJSON(hostConfigs))
	}
	return nil
}
