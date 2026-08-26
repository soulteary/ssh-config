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

import "fmt"

const Usage = `Usage:
  ssh-config -to-yaml
  ssh-config -to-ssh
  ssh-config -to-json
  ssh-config -to-yaml -src <SSH config file>
  ssh-config -to-ssh -src <v3 YAML or JSON file>
  ssh-config -legacy -src <source file or directory> -dest <destination file>
  ssh-config -help
  ssh-config -version
`

func ShowHelp() {
	fmt.Print(Usage)
}
