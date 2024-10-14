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
		fmt.Println("JSON")
		fmt.Println("Implemented later")
		os.Exit(0)
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

	userInput := ""
	if Cmd.CheckUseStdin(os.Stdin.Stat) {
		userInput = Fn.GetUserInputFromStdin()
	} else {
		isValid, notValidReason := Cmd.CheckIOArgvValid(args)
		if !isValid {
			fmt.Println(notValidReason)
			os.Exit(1)
		}
		fmt.Println("Use Files")
	}

	fileType := Fn.DetectStringType(userInput)
	result := Process(fileType, userInput, args)
	fmt.Println(string(result))
	os.Exit(0)
}
