package parser

import (
	Define "github.com/soulteary/ssh-yaml/internal/define"
	Fn "github.com/soulteary/ssh-yaml/internal/fn"
)

func ConvertToJSON(hostConfigs []Define.HostConfig) []byte {
	return Fn.GetJSONBytes(hostConfigs)
}
