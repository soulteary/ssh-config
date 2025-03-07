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

package define

type HostExtraConfig struct {
	Prefix string
}

// ssh config
type HostConfig struct {
	Name   string `yaml:"Name,omitempty"`
	Notes  string `yaml:"Notes,omitempty"`
	Config map[string]string
	Extra  HostExtraConfig `yaml:"Extra,omitempty"`
}

// json
type HostConfigDataForJSON map[string]string

type HostConfigForJSON struct {
	Name  string                `json:"Name,omitempty"`
	Notes string                `json:"Notes,omitempty"`
	Data  HostConfigDataForJSON `json:"Data,omitempty"`
}

// yaml
type GroupConfig struct {
	Prefix string                `yaml:"Prefix,omitempty"`
	Common map[string]string     `yaml:"Common,omitempty"`
	Hosts  map[string]HostConfig `yaml:"Hosts,omitempty"`
}
type YAMLOutput struct {
	Global  map[string]string      `yaml:"global,omitempty"`
	Default map[string]string      `yaml:"default,omitempty"`
	Groups  map[string]GroupConfig `yaml:",inline"`
}

var ExcludePatterns = []string{
	"known_hosts",
	"authorized_keys",
	"*.pub",
	"id_*",
	"*.key",
	"*.pem",
	"*.ppk",
}
