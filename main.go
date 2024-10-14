package main

import (
	"fmt"
	"os"
	"strings"

	Cmd "github.com/soulteary/ssh-config/cmd"
	Define "github.com/soulteary/ssh-config/internal/define"
	Fn "github.com/soulteary/ssh-config/internal/fn"
	Parser "github.com/soulteary/ssh-config/internal/parser"
)

func Process(fileType string, userInput string, args Cmd.Args) []byte {
	switch strings.ToUpper(fileType) {
	case "YAML":
		hostConfigs := Parser.GroupYAMLConfig(userInput)

		if args.ToYAML {
			return Parser.ConvertToYAML(hostConfigs)
		}

		if args.ToSSH {
			return Parser.ConvertToSSH(hostConfigs)
		}

		if args.ToJSON {
			return Parser.ConvertToJSON(hostConfigs)
		}
	case "JSON":
		fmt.Println("JSON")
	case "TEXT":
		configs := Parser.GroupSSHConfig(userInput)
		hostConfigs := make([]Define.HostConfig, 0)
		for host, hostConfig := range configs {
			rawInfo := Parser.GetSSHConfigContent(host, hostConfig)
			hostInfo := Parser.ParseSSHConfig(rawInfo.Config, rawInfo.Comments)
			config, name, notes := Parser.GetSingleHostData(hostInfo)

			hostConfigs = append(hostConfigs, Define.HostConfig{
				Name:   name,
				Notes:  notes,
				Config: config,
			})
		}

		if args.ToYAML {
			return Parser.ConvertToYAML(hostConfigs)
		}

		if args.ToSSH {
			return Parser.ConvertToSSH(hostConfigs)
		}

		if args.ToJSON {
			return Parser.ConvertToJSON(hostConfigs)
		}
	}
	return nil
}

func main() {
	args := Cmd.ParseArgs()
	isValid, validReason := Cmd.CheckConvertArgvValid(args)
	if !isValid {
		fmt.Println(validReason)
		os.Exit(1)
	}

	userInput := ""
	if Cmd.CheckUseStdin(os.Stdin.Stat) {
		userInput = Fn.GetUserInputFromStdin()
	} else {
		isValid, validReason := Cmd.CheckIOArgvValid(args)
		if !isValid {
			fmt.Println(validReason)
			os.Exit(1)
		}
		fmt.Println("Use Files")
	}

	fileType := Fn.DetectStringType(userInput)
	result := Process(fileType, userInput, args)
	fmt.Println(string(result))
	os.Exit(0)
}
