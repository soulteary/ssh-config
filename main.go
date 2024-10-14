package main

import (
	"fmt"
	"os"
	"strings"

	Cmd "github.com/soulteary/ssh-config/cmd"
	"github.com/soulteary/ssh-config/internal/define"
	Fn "github.com/soulteary/ssh-config/internal/fn"
	Parser "github.com/soulteary/ssh-config/internal/parser"
)

func Process(fileType string, userInput string, args Cmd.Args) []byte {
	var hostConfigs []define.HostConfig

	switch strings.ToUpper(fileType) {
	case "YAML":
		hostConfigs = Parser.GroupYAMLConfig(userInput)
	case "JSON":
		hostConfigs = Parser.GroupJSONConfig(userInput)
	case "TEXT":
		hostConfigs = Parser.GroupSSHConfig(userInput)
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
	return nil
}

func main() {
	args := Cmd.ParseArgs()
	isValid, notValidReason := Cmd.CheckConvertArgvValid(args)
	if !isValid {
		fmt.Println(notValidReason)
		os.Exit(1)
	}

	pipeMode := Cmd.CheckUseStdin(os.Stdin.Stat)
	userInput := ""
	if pipeMode {
		userInput = Fn.GetUserInputFromStdin()
	} else {
		isValid, notValidReason := Cmd.CheckIOArgvValid(args)
		if !isValid {
			fmt.Println(notValidReason)
			os.Exit(1)
		}

		content, err := Fn.GetPathContent(args.Src)
		if err != nil {
			fmt.Println("Error reading file:", err)
			os.Exit(1)
		}
		userInput = string(content)
	}

	fileType := Fn.DetectStringType(userInput)
	result := Process(fileType, userInput, args)

	if pipeMode {
		fmt.Println(string(result))
	} else {
		err := Fn.Save(args.Dest, result)
		if err != nil {
			fmt.Println("Error saving file:", err)
			os.Exit(1)
		}
		fmt.Println("File has been saved successfully")
		fmt.Println("File path:", args.Dest)
	}
}
