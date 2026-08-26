# Security policy

## Supported versions

| Version | Supported |
| --- | --- |
| Current `2.x` release and `main` | Yes |
| Releases before `2.0.0` | No |

Schema version `3` is a data-format version, not an application major version.

## Reporting a vulnerability

Please report suspected vulnerabilities through a
[private GitHub security advisory](https://github.com/soulteary/ssh-config/security/advisories/new).
Do not open a public issue for an unpatched vulnerability.

Include enough information to reproduce and assess the problem:

- affected release or commit;
- operating system and architecture;
- minimal configuration input and command line;
- observed and expected behavior;
- security impact and any known preconditions;
- proof-of-concept code or logs with credentials, hostnames, keys, and tokens
  removed.

Relevant reports include unintended configuration loss or mutation, unsafe
destination-file replacement, symbolic-link or path handling flaws, unexpected
file access through `Include`, command execution, and parser resource-exhaustion
issues.

The maintainers will coordinate disclosure and remediation in the private
advisory. Publish details only after a fixed release is available or the
maintainers agree on a disclosure date.

## Safe testing

Use synthetic configuration files and accounts you control. Never include
private keys, authentication agents, production hostnames, tokens, or other
secrets in a report. Do not test against infrastructure without authorization.
