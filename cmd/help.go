package cmd

import "fmt"

const Usage = `Usage:
  ssh-config -to-yaml
  ssh-config -to-ssh
  ssh-config -to-json
  ssh-config -src <source file or directories path> -dest <destination file path>
  ssh-config -help
`

func ShowHelp() {
	fmt.Print(Usage)
}
