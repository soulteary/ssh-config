# Lossless schema v3

The v3 YAML/JSON schema is a transport and editing format for OpenSSH client
configuration files. Its version number is independent of the application's
semantic version and the Go module's `/v3` path.

## Design guarantees

- Parsing and exporting an unchanged document preserves every source byte.
- Comments, blank lines, directive order, repeated directives, unknown
  directives, line endings, quoting, spacing, and non-UTF-8 bytes remain in
  `rawBase64`.
- A malformed line is exported as a raw-only `invalid` node. Diagnostics do
  not make the original bytes unavailable.
- Editing one directive rewrites only that physical line. Other nodes are
  copied from their original bytes.
- An edited final line keeps its missing final newline. A brand-new directive
  without raw bytes defaults to LF.

The parser builds a concrete syntax tree; it does not evaluate host matching,
expand tokens, follow `Include`, or execute `Match exec` while importing a
document.

## Envelope

```yaml
schemaVersion: 3
documents:
  - path: ~/.ssh/config
    nodes:
      - type: directive
        rawBase64: SG9zdCBleGFtcGxlCg==
        directive:
          keyword: Host
          arguments:
            - example
```

The root fields are:

- `schemaVersion`: must be the integer `3`.
- `documents`: a non-empty array of physical source documents.

Each document has a `path` and a `nodes` array. The path may be empty only when
the envelope contains one document. Every document in a multi-document
envelope needs a non-empty, unique path. An empty selector can reconstruct a
document only when the envelope contains exactly one document.

JSON uses the same field names and shapes. Both decoders reject unknown fields,
unsupported versions, duplicate paths, invalid Base64, trailing JSON values,
and invalid node shapes.

## Nodes

Every node represents one physical source line and has one of these types:

| Type | Editable directive | Meaning |
| --- | --- | --- |
| `blank` | no | Empty or whitespace-only line |
| `comment` | no | Full-line comment |
| `directive` | optional | Parsed OpenSSH directive |
| `invalid` | no | Malformed line retained verbatim |

`rawBase64` contains the complete physical line, including indentation,
spacing, comments, and its line ending. Base64 is used so arbitrary source
bytes survive YAML and JSON encoding.

A directive view contains:

- `keyword`: original keyword spelling;
- `arguments`: decoded OpenSSH argument values;
- `comment`: optional inline comment, normally including its leading `#`.

Exported directive nodes contain both raw bytes and the editable view. If the
view still describes the raw line exactly, reconstruction copies the raw bytes.
If any semantic field changes, reconstruction renders that line canonically
while retaining its original line-ending presence and style. A new directive
may omit `rawBase64`; it is rendered canonically with LF.

Blank nodes may represent an empty line without raw bytes. Every other
non-directive node needs raw bytes. `comment` and `invalid` nodes never accept a
directive view.

## CLI workflow

```bash
ssh-config -to-yaml -src ~/.ssh/config -dest config.v3.yaml
# Edit only directive.keyword, directive.arguments, or directive.comment.
ssh-config -to-ssh -src config.v3.yaml -dest ~/.ssh/config
```

Use `-to-json` instead of `-to-yaml` for a JSON envelope. Destination files are
replaced atomically; existing regular-file permissions are preserved, and
symbolic-link destinations are rejected.

## Go API

```go
doc, err := sshconfig.Parse(source)
if err != nil {
    return err
}
schemaDoc, err := doc.ToSchema("~/.ssh/config")
if err != nil {
    return err
}
encoded, err := sshconfig.MarshalSchemaYAML(sshconfig.NewSchema(schemaDoc))
```

Use `UnmarshalSchemaYAML` or `UnmarshalSchemaJSON` for strict decoding, then
call `Schema.Document(path)` and `Document.MarshalPreserve()` to reconstruct
SSH source.

## Legacy migration

The default converter accepts the previous map-based YAML and host-array JSON
formats and migrates them deterministically into v3. YAML group and host source
order, including order inherited through merge aliases, is retained because
SSH host precedence is order-sensitive. Directive map keys are sorted, and
legacy `Default` and `Common` values use the existing precedence rules.

Migration cannot recover information that the legacy representation never
stored, including repeated directive values, physical comments and
spacing, `Match` structure, `Include` boundaries, or unknown directives that
were discarded earlier. Keep an original SSH source file when exact recovery
matters.

## Versioning

Readers must reject schema versions they do not support. Additive or breaking
schema changes require a new `schemaVersion`; they must not be inferred from
the application release number. Producers should preserve `rawBase64` and
avoid rebuilding unchanged nodes from their semantic views.
