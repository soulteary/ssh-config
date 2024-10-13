package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	Define "github.com/soulteary/ssh-config/internal/define"
	Fn "github.com/soulteary/ssh-config/internal/fn"
	Parser "github.com/soulteary/ssh-config/internal/parser"
)

func CheckUseStdin() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("Error getting stdin stat:", err)
		return false
	}
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return true
	}
	return false
}

func main() {
	toYAML := flag.Bool("to-yaml", false, "Convert SSH config(Text/JSON) to YAML")
	toSSH := flag.Bool("to-ssh", false, "Convert SSH config(YAML/JSON) to YAML")
	toJSON := flag.Bool("to-json", false, "Convert SSH config(YAML/Text) to JSON")

	flag.Parse()

	if (*toYAML == *toSSH && *toYAML == *toJSON) || (*toSSH == *toJSON && *toSSH == *toYAML) || (*toJSON == *toYAML && *toJSON == *toSSH) {
		fmt.Println("Please specify either -to-yaml or -to-ssh or -to-json")
		os.Exit(1)
	}

	if CheckUseStdin() {
		input := Fn.GetUserInputFromStdin()
		fileType := Fn.DetectStringType(input)

		switch strings.ToUpper(fileType) {
		case "YAML":
			hostConfigs := Parser.GroupYAMLConfig(input)

			if *toYAML {
				fmt.Println(string(Parser.ConvertToYAML(hostConfigs)))
			}

			if *toSSH {
				fmt.Println(string(Parser.ConvertToSSH(hostConfigs)))
			}

			if *toJSON {
				fmt.Println(string(Parser.ConvertToJSON(hostConfigs)))
			}
		case "JSON":
			fmt.Println("JSON")
		case "TEXT":
			configs := Parser.GroupSSHConfig(input)
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

			if *toYAML {
				fmt.Println(string(Parser.ConvertToYAML(hostConfigs)))
			}

			if *toSSH {
				fmt.Println(string(Parser.ConvertToSSH(hostConfigs)))
			}

			if *toJSON {
				fmt.Println(string(Parser.ConvertToJSON(hostConfigs)))
			}
		}
	} else {
		fmt.Println("Use Files")
	}
	os.Exit(0)
}
