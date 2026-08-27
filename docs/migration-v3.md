# Migrating from v2 to v3

Version 3 makes the lossless schema the default CLI representation. This is a
deliberate breaking change: YAML and JSON output now use the v3 envelope rather
than the previous map and host-array shapes.

## Command-line changes

The default command reads one physical SSH configuration file and writes v3
YAML:

```bash
ssh-config -lossless -to-yaml -src ~/.ssh/config -dest config.v3.yaml
ssh-config -lossless -to-ssh -src config.v3.yaml -dest ~/.ssh/config
```

Without `-src`, the default converter reads `~/.ssh/config`. It does not scan a
directory because joining physical files would discard their boundaries.

Use `-legacy` when an existing consumer still needs the v2 representation or
directory scan:

```bash
ssh-config -legacy -to-yaml -src ~/.ssh -dest config.legacy.yaml
```

The old `-lossless` option is retained as a deprecated compatibility alias, so
commands remain safe while v2 and v3 installations coexist. It is a no-op in
v3 and can be removed after every installation has upgraded. Legacy YAML and
JSON remain valid inputs in the default mode and are migrated to v3; `-legacy`
controls the output pipeline, not input compatibility.

## Go module path

The public module path changes with the application major version:

```go
import "github.com/soulteary/ssh-config/v3/pkg/sshconfig"
```

Update dependencies with:

```bash
go get github.com/soulteary/ssh-config/v3@latest
go mod tidy
```

The schema's `schemaVersion: 3` is independent of the Go module and application
version. The v3 application continues to read and write schema version 3.

Version 3 requires Go 1.27 or newer. Applications and tools that build the
module from source must upgrade their Go toolchain before moving from v2.

## Release archives

The Linux PowerPC archive changes from big-endian `ppc64` to little-endian
`ppc64le`. Automated downloads should update the architecture name. Users who
still require a big-endian `ppc64` binary must build v3 from source or remain on
v2; v3 does not publish that prebuilt archive.

## Information that cannot be recovered

Migrating an existing legacy document cannot recreate repeated directives,
comments, whitespace, quoting, unknown directives, or ordering that the old
map representation never stored. Convert from the original SSH source when an
exact round trip matters.
