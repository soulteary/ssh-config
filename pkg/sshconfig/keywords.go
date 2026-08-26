package sshconfig

import "strings"

// KeywordStatus describes how the OpenSSH client treats a directive name.
type KeywordStatus uint8

const (
	KeywordSupported KeywordStatus = iota
	KeywordIgnored
	KeywordDeprecated
	KeywordUnsupported
	KeywordPlatformDependent
)

// KeywordInfo describes an OpenSSH client configuration keyword.
type KeywordInfo struct {
	Name   string
	Status KeywordStatus
}

// The registry follows openssh-portable readconf.c as of OpenBSD revision
// 1.415 (2026-07-21). It is data derived from the public keyword table, not a
// translation of OpenSSH's evaluator.
var keywordRegistry = buildKeywordRegistry()

// LookupKeyword performs a case-insensitive OpenSSH keyword lookup.
func LookupKeyword(name string) (KeywordInfo, bool) {
	info, ok := keywordRegistry[strings.ToLower(name)]
	return info, ok
}

func buildKeywordRegistry() map[string]KeywordInfo {
	registry := make(map[string]KeywordInfo)
	add := func(status KeywordStatus, names string) {
		for _, name := range strings.Fields(names) {
			registry[name] = KeywordInfo{Name: name, Status: status}
		}
	}

	add(KeywordIgnored, `
		protocol
	`)
	add(KeywordDeprecated, `
		cipher fallbacktorsh globalknownhostsfile2 rhostsauthentication
		userknownhostsfile2 useroaming usersh useprivilegedport
	`)
	add(KeywordUnsupported, `
		afstokenpassing kerberosauthentication kerberostgtpassing
		rsaauthentication rhostsrsaauthentication compressionlevel
	`)
	add(KeywordPlatformDependent, `
		gssapiauthentication gssapidelegatecredentials pkcs11provider
		smartcarddevice
	`)
	add(KeywordSupported, `
		addkeystoagent addressfamily batchmode bindaddress bindinterface
		canonicaldomains canonicalizefallbacklocal canonicalizehostname
		canonicalizemaxdots canonicalizepermittedcnames casignaturealgorithms
		certificatefile challengeresponseauthentication channeltimeout checkhostip
		ciphers clearallforwardings compression connectionattempts connecttimeout
		controlmaster controlpath controlpersist dsaauthentication dynamicforward
		enableescapecommandline enablesshkeysign escapechar exitonforwardfailure
		fingerprinthash forkafterauthentication forwardagent forwardx11
		forwardx11timeout forwardx11trusted gatewayports globalknownhostsfile
		hashknownhosts host hostbasedacceptedalgorithms hostbasedauthentication
		hostbasedkeytypes hostkeyalgorithms hostkeyalias hostname identityagent
		identityfile identityfile2 identitiesonly ignoreunknown include ipqos
		kbdinteractiveauthentication kbdinteractivedevices keepalive
		knownhostscommand kexalgorithms localcommand localforward loglevel
		logverbose match macs nohostauthenticationforlocalhost
		numberofpasswordprompts obscurekeystroketiming passwordauthentication
		permitlocalcommand permitremoteopen port preferredauthentications
		proxycommand proxyjump proxyusefdpass pubkeyacceptedalgorithms
		pubkeyacceptedkeytypes pubkeyauthentication refuseconnection rekeylimit
		remotecommand remoteforward requesttty requiredrsasize revokedhostkeys
		securitykeyprovider sendenv serveralivecountmax serveraliveinterval
		sessiontype setenv skeyauthentication stdinnull streamlocalbindmask
		streamlocalbindunlink stricthostkeychecking syslogfacility tag
		tcpkeepalive tisauthentication tunnel tunneldevice updatehostkeys user
		userknownhostsfile verifyhostkeydns versionaddendum visualhostkey
		warnweakcrypto xauthlocation
	`)
	return registry
}
