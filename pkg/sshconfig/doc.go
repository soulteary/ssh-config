// Package sshconfig provides a lossless syntax model for OpenSSH client
// configuration files.
//
// Parsing is deliberately separate from evaluating a configuration for a
// particular host. The syntax model retains comments, whitespace, quoting,
// line endings, repeated directives, and unknown directives.
package sshconfig
