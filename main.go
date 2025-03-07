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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	Cmd "github.com/soulteary/ssh-config/cmd"
	Fn "github.com/soulteary/ssh-config/internal/fn"
	Parser "github.com/soulteary/ssh-config/internal/parser"
)

type Dependencies struct {
	StdinStat             func() (os.FileInfo, error)
	Exit                  func(int)
	Println               func(...interface{}) (n int, err error)
	GetContent            func(string) ([]byte, error)
	SaveFile              func(string, []byte) error
	GetUserInputFromStdin func() string
	Process               func(string, string, Cmd.Args) []byte
	CheckUseStdin         func() bool
	UserHomeDir           func() (string, error)
}

func Run(args Cmd.Args, deps Dependencies) error {
	isValid, notValidReason := Cmd.CheckConvertArgvValid(args)
	if !isValid {
		deps.Println(notValidReason)
		return fmt.Errorf(notValidReason)
	}

	pipeMode := deps.CheckUseStdin()
	var userInput string
	if pipeMode {
		userInput = deps.GetUserInputFromStdin()
	} else {
		isValid, notValidReason := Cmd.CheckIOArgvValid(args)
		if !isValid {
			deps.Println(notValidReason)
			return fmt.Errorf(notValidReason)
		}

		content, err := deps.GetContent(args.Src)
		if err != nil {
			deps.Println("Error reading file:", err)
			return err
		}
		userInput = string(content)
	}

	fileType := Fn.DetectStringType(userInput)
	result := deps.Process(fileType, userInput, args)

	if pipeMode {
		deps.Println(string(result))
	} else {
		if args.Dest == "" {
			deps.Println(string(result))
			return nil
		}

		err := deps.SaveFile(args.Dest, result)
		if err != nil {
			deps.Println("Error saving file:", err)
			return err
		}
		deps.Println("File has been saved successfully")
		deps.Println("File path:", args.Dest)
	}

	return nil
}

func MainWithDependencies(exit func(int), userHomeDir func() (string, error)) {
	deps := Dependencies{
		StdinStat:             os.Stdin.Stat,
		Exit:                  os.Exit,
		Println:               fmt.Println,
		GetContent:            Fn.GetPathContent,
		SaveFile:              Fn.Save,
		GetUserInputFromStdin: Fn.GetUserInputFromStdin,
		Process:               Parser.Process,
		CheckUseStdin:         func() bool { return Cmd.CheckUseStdin(os.Stdin.Stat) },
	}
	args := Cmd.ParseArgs()

	// default src to ~/.ssh
	if args.Src == "" {
		homeDir, err := userHomeDir()
		if err != nil {
			fmt.Println("Error: getting user home directory:", err)
			exit(1)
		}
		args.Src = filepath.Join(homeDir, ".ssh")
	}

	// default to YAML
	if !(args.ToYAML && args.ToJSON && args.ToSSH) {
		args.ToYAML = true
	}

	if err := Run(args, deps); err != nil {
		exit(1)
	}
}

func main() {
	MainWithDependencies(os.Exit, os.UserHomeDir)
}
