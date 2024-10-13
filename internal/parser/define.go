package parser

// docs: https://man.openbsd.org/ssh_config
type HostConfig struct {
	Host  string `yaml:"Host,omitempty"`
	Match string `yaml:"Match,omitempty"`

	AddKeysToAgent string `yaml:"AddKeysToAgent,omitempty"`
	AddressFamily  string `yaml:"AddressFamily,omitempty"`
	BatchMode      string `yaml:"BatchMode,omitempty"`
	BindAddress    string `yaml:"BindAddress,omitempty"`
	BindInterface  string `yaml:"BindInterface,omitempty"`

	CanonicalDomains            string `yaml:"CanonicalDomains,omitempty"`
	CanonicalizeFallbackLocal   string `yaml:"CanonicalizeFallbackLocal,omitempty"`
	CanonicalizeHostname        string `yaml:"CanonicalizeHostname,omitempty"`
	CanonicalizeMaxDots         string `yaml:"CanonicalizeMaxDots,omitempty"`
	CanonicalizePermittedCNAMEs string `yaml:"CanonicalizePermittedCNAMEs,omitempty"`
	CASignatureAlgorithms       string `yaml:"CASignatureAlgorithms,omitempty"`
	CertificateFile             string `yaml:"CertificateFile,omitempty"`
	ChannelTimeout              string `yaml:"ChannelTimeout,omitempty"`
	CheckHostIP                 string `yaml:"CheckHostIP,omitempty"`
	Ciphers                     string `yaml:"Ciphers,omitempty"`
	ClearAllForwardings         string `yaml:"ClearAllForwardings,omitempty"`
	Compression                 string `yaml:"Compression,omitempty"`
	ConnectionAttempts          string `yaml:"ConnectionAttempts,omitempty"`
	ConnectTimeout              string `yaml:"ConnectTimeout,omitempty"`
	ControlMaster               string `yaml:"ControlMaster,omitempty"`
	ControlPath                 string `yaml:"ControlPath,omitempty"`
	ControlPersist              string `yaml:"ControlPersist,omitempty"`

	DynamicForward string `yaml:"DynamicForward,omitempty"`

	EnableEscapeCommandline string `yaml:"EnableEscapeCommandline,omitempty"`
	EnableSSHKeysign        string `yaml:"EnableSSHKeysign,omitempty"`
	EscapeChar              string `yaml:"EscapeChar,omitempty"`
	ExitOnForwardFailure    string `yaml:"ExitOnForwardFailure,omitempty"`

	FingerprintHash         string `yaml:"FingerprintHash,omitempty"`
	ForkAfterAuthentication string `yaml:"ForkAfterAuthentication,omitempty"`
	ForwardAgent            string `yaml:"ForwardAgent,omitempty"`
	ForwardX11              string `yaml:"ForwardX11,omitempty"`
	ForwardX11Timeout       string `yaml:"ForwardX11Timeout,omitempty"`
	ForwardX11Trusted       string `yaml:"ForwardX11Trusted,omitempty"`

	GatewayPorts              string `yaml:"GatewayPorts,omitempty"`
	GlobalKnownHostsFile      string `yaml:"GlobalKnownHostsFile,omitempty"`
	GSSAPIAuthentication      string `yaml:"GSSAPIAuthentication,omitempty"`
	GSSAPIDelegateCredentials string `yaml:"GSSAPIDelegateCredentials,omitempty"`

	HashKnownHosts              string `yaml:"HashKnownHosts,omitempty"`
	HostbasedAcceptedAlgorithms string `yaml:"HostbasedAcceptedAlgorithms,omitempty"`
	HostbasedAuthentication     string `yaml:"HostbasedAuthentication,omitempty"`
	HostKeyAlgorithms           string `yaml:"HostKeyAlgorithms,omitempty"`
	HostKeyAlias                string `yaml:"HostKeyAlias,omitempty"`
	HostName                    string `yaml:"HostName,omitempty"`

	IdentitiesOnly string `yaml:"IdentitiesOnly,omitempty"`
	IdentityFile   string `yaml:"IdentityFile,omitempty"`
	IgnoreUnknown  string `yaml:"IgnoreUnknown,omitempty"`
	Include        string `yaml:"Include,omitempty"`
	IPQoS          string `yaml:"IPQoS,omitempty"`

	KbdInteractiveAuthentication string `yaml:"KbdInteractiveAuthentication,omitempty"`
	KbdInteractiveDevices        string `yaml:"KbdInteractiveDevices,omitempty"`
	KexAlgorithms                string `yaml:"KexAlgorithms,omitempty"`
	KnownHostsCommand            string `yaml:"KnownHostsCommand,omitempty"`

	LocalCommand    string `yaml:"LocalCommand,omitempty"`
	LocalForward    string `yaml:"LocalForward,omitempty"`
	LogLevel        string `yaml:"LogLevel,omitempty"`
	LogLevelVerbose string `yaml:"LogLevelVerbose,omitempty"`

	MACs string `yaml:"MACs,omitempty"`

	NoHostAuthenticationForLocalhost string `yaml:"NoHostAuthenticationForLocalhost,omitempty"`
	NumberOfPasswordPrompts          string `yaml:"NumberOfPasswordPrompts,omitempty"`

	ObscureKeystrokeTiming string `yaml:"ObscureKeystrokeTiming,omitempty"`

	PasswordAuthentication   string `yaml:"PasswordAuthentication,omitempty"`
	PermitLocalCommand       string `yaml:"PermitLocalCommand,omitempty"`
	PermitRemoteOpen         string `yaml:"PermitRemoteOpen,omitempty"`
	PKCS11Provider           string `yaml:"PKCS11Provider,omitempty"`
	Port                     string `yaml:"Port,omitempty"`
	PreferredAuthentications string `yaml:"PreferredAuthentications,omitempty"`
	ProxyCommand             string `yaml:"ProxyCommand,omitempty"`
	ProxyJump                string `yaml:"ProxyJump,omitempty"`
	ProxyUseFdpass           string `yaml:"ProxyUseFdpass,omitempty"`
	PubkeyAcceptedAlgorithms string `yaml:"PubkeyAcceptedAlgorithms,omitempty"`
	PubkeyAuthentication     string `yaml:"PubkeyAuthentication,omitempty"`
	RekeyLimit               string `yaml:"RekeyLimit,omitempty"`
	RemoteCommand            string `yaml:"RemoteCommand,omitempty"`
	RemoteForward            string `yaml:"RemoteForward,omitempty"`
	RequestTTY               string `yaml:"RequestTTY,omitempty"`
	RequireRSASize           string `yaml:"RequireRSASize,omitempty"`
	RevokedHostKeys          string `yaml:"RevokedHostKeys,omitempty"`

	SecurityKeyProvider   string `yaml:"SecurityKeyProvider,omitempty"`
	SendEnv               string `yaml:"SendEnv,omitempty"`
	ServerAliveCountMax   string `yaml:"ServerAliveCountMax,omitempty"`
	ServerAliveInterval   string `yaml:"ServerAliveInterval,omitempty"`
	SessionType           string `yaml:"SessionType,omitempty"`
	SetEnv                string `yaml:"SetEnv,omitempty"`
	StreamLocalBindMask   string `yaml:"StreamLocalBindMask,omitempty"`
	StreamLocalBindUnlink string `yaml:"StreamLocalBindUnlink,omitempty"`
	StrictHostKeyChecking string `yaml:"StrictHostKeyChecking,omitempty"`
	SyslogFacility        string `yaml:"SyslogFacility,omitempty"`

	TCPKeepAlive string `yaml:"TCPKeepAlive,omitempty"`
	Tag          string `yaml:"Tag,omitempty"`
	Tunnel       string `yaml:"Tunnel,omitempty"`
	TunnelDevice string `yaml:"TunnelDevice,omitempty"`

	UpdateHostKeys     string `yaml:"UpdateHostKeys,omitempty"`
	User               string `yaml:"User,omitempty"`
	UserKnownHostsFile string `yaml:"UserKnownHostsFile,omitempty"`
	VerifyHostKeyDNS   string `yaml:"VerifyHostKeyDNS,omitempty"`
	VisualHostKey      string `yaml:"VisualHostKey,omitempty"`

	XAuthLocation string `yaml:"XAuthLocation,omitempty"`

	YamlUserNotes string `yaml:"YamlUserNotes,omitempty"`
	YamlUserHost  string `yaml:"YamlUserHost,omitempty"`
}
