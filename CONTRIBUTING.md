# Contributing

Thank you for helping improve `ssh-config`.

## Before opening a change

- Use an issue for behavior changes that need design discussion.
- Use synthetic SSH configuration in reports and tests. Never commit keys,
  tokens, production hostnames, or other credentials.
- Keep pull requests focused on one concern and include regression tests for
  bug fixes.

## Development

Install the Go version declared in `go.mod`, then run:

```sh
go mod tidy -diff
go test ./...
go vet ./...
```

Parser and editor changes must preserve unmodified bytes. Add a round-trip test
when changing parsing, formatting, validation, or mutation behavior. File-system
changes should cover symbolic links and non-regular files where applicable.

Before requesting review, complete the pull request checklist and confirm that
the relevant GitHub Actions checks pass.

## Security reports

Do not open a public issue for a suspected vulnerability. Follow
[`SECURITY.md`](SECURITY.md) and use a private security advisory.
