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
