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

package cmd

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

var (
	once sync.Once
	args Args
)

type Args struct {
	ToYAML   bool
	ToSSH    bool
	ToJSON   bool
	Src      string
	Dest     string
	ShowHelp bool
}

const (
	DEFAULT_TO_YAML = false
	DEFAULT_TO_SSH  = false
	DEFAULT_TO_JSON = false
	DEFAULT_SRC     = ""
	DEFAULT_DEST    = ""
	DEFAULT_HELP    = false
)

func initFlags() {
	flag.BoolVar(&args.ToYAML, "to-yaml", DEFAULT_TO_YAML, "Convert SSH config(Text/JSON) to YAML")
	flag.BoolVar(&args.ToSSH, "to-ssh", DEFAULT_TO_SSH, "Convert SSH config(YAML/JSON) to YAML")
	flag.BoolVar(&args.ToJSON, "to-json", DEFAULT_TO_JSON, "Convert SSH config(YAML/Text) to JSON")
	flag.StringVar(&args.Src, "src", DEFAULT_SRC, "Source file or directories path, valid when using non-pipeline mode")
	flag.StringVar(&args.Dest, "dest", DEFAULT_DEST, "Destination file path, valid when using non-pipeline mode")
	flag.BoolVar(&args.ShowHelp, "help", DEFAULT_HELP, "Show help")
}

func ParseArgs() Args {
	once.Do(func() {
		initFlags()
		flag.Parse()
	})
	return args
}

func ResetFlags() {
	flag.CommandLine = flag.NewFlagSet(flag.CommandLine.Name(), flag.ExitOnError)
	args = Args{
		ToYAML:   DEFAULT_TO_YAML,
		ToSSH:    DEFAULT_TO_SSH,
		ToJSON:   DEFAULT_TO_JSON,
		Src:      DEFAULT_SRC,
		Dest:     DEFAULT_DEST,
		ShowHelp: DEFAULT_HELP,
	} // Reset the args
	once = sync.Once{} // Reset the once
}

func CheckUseStdin(osStdinStat func() (fs.FileInfo, error)) bool {
	fi, err := osStdinStat()
	if err != nil {
		fmt.Println("Error getting stdin stat:", err)
		return false
	}
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return true
	}
	return false
}

func CheckConvertArgvValid(args Args) (result bool, desc string) {
	trueCount := 0
	if args.ToJSON {
		trueCount++
	}
	if args.ToSSH {
		trueCount++
	}
	if args.ToYAML {
		trueCount++
	}

	if trueCount != 1 {
		return false, "Please specify either -to-yaml or -to-ssh or -to-json"
	}

	return true, ""
}

func CheckIOArgvValid(args Args) (result bool, desc string) {
	if args.Src == "" {
		return false, "Please specify source and destination file path"
	}

	// Check if src exists
	_, err := os.Stat(args.Src)
	if os.IsNotExist(err) {
		return false, fmt.Sprintf("Error: Source path '%s' does not exist", args.Src)
	}

	// allow empty dest
	if args.Dest == "" {
		return true, ""
	}

	// Check if dist exists
	_, err = os.Stat(args.Dest)
	if os.IsNotExist(err) {
		// If dist doesn't exist, check if its parent directory exists
		parentDir := filepath.Dir(args.Dest)
		_, err := os.Stat(parentDir)
		if os.IsNotExist(err) {
			return false, fmt.Sprintf("Error: Parent directory of destination '%s' does not exist", args.Dest)
		}
	}

	return true, ""
}
