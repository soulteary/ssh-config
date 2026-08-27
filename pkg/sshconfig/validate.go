package sshconfig

import "fmt"

// Severity is the importance of a validation issue.
type Severity uint8

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
)

// UnknownDirectivePolicy controls diagnostics for directives absent from the
// OpenSSH keyword registry.
type UnknownDirectivePolicy uint8

const (
	UnknownDirectiveIgnore UnknownDirectivePolicy = iota
	UnknownDirectiveWarn
	UnknownDirectiveError
)

// ValidateOptions controls syntax and keyword validation.
type ValidateOptions struct {
	UnknownDirective UnknownDirectivePolicy
}

// ValidationIssue is a structured, source-positioned diagnostic.
type ValidationIssue struct {
	NodeID   NodeID
	Position Position
	Severity Severity
	Code     string
	Message  string
}

// Validate checks syntax and OpenSSH keyword compatibility without changing
// or evaluating the document.
func (d *Document) Validate(options ValidateOptions) []ValidationIssue {
	if d == nil {
		return []ValidationIssue{{
			Severity: SeverityError,
			Code:     "nil-document",
			Message:  "document is nil",
		}}
	}
	issues := make([]ValidationIssue, 0, len(d.diagnostics))
	for _, diagnostic := range d.diagnostics {
		nodeID := d.nodeIDAtOffset(diagnostic.Position.Offset)
		if d.removed[nodeID] {
			continue
		}
		if _, replaced := d.replacement[nodeID]; replaced {
			continue
		}
		issues = append(issues, ValidationIssue{
			NodeID:   nodeID,
			Position: diagnostic.Position,
			Severity: SeverityError,
			Code:     "invalid-syntax",
			Message:  diagnostic.Message,
		})
	}
	for _, node := range d.nodes {
		if d.removed[node.ID] || node.Directive == nil {
			continue
		}
		directive := node.Directive
		position := directive.Keyword.Position
		if position.Line == 0 {
			position = Position{Line: 1, Column: 1}
		}
		if len(directive.Arguments) == 0 {
			issues = append(issues, ValidationIssue{
				NodeID:   node.ID,
				Position: position,
				Severity: SeverityError,
				Code:     "missing-argument",
				Message:  fmt.Sprintf("directive %q requires an argument", directive.KeywordValue),
			})
		}

		info, ok := LookupKeyword(directive.KeywordValue)
		if !ok {
			severity, report := unknownSeverity(options.UnknownDirective)
			if report {
				issues = append(issues, ValidationIssue{
					NodeID:   node.ID,
					Position: position,
					Severity: severity,
					Code:     "unknown-directive",
					Message:  fmt.Sprintf("unknown OpenSSH directive %q", directive.KeywordValue),
				})
			}
			continue
		}
		switch info.Status {
		case KeywordIgnored:
			issues = append(issues, keywordIssue(node.ID, position, SeverityInfo, "ignored-directive", info.Name, "is ignored by OpenSSH"))
		case KeywordDeprecated:
			issues = append(issues, keywordIssue(node.ID, position, SeverityWarning, "deprecated-directive", info.Name, "is deprecated by OpenSSH"))
		case KeywordUnsupported:
			issues = append(issues, keywordIssue(node.ID, position, SeverityWarning, "unsupported-directive", info.Name, "is unsupported by OpenSSH"))
		case KeywordPlatformDependent:
			issues = append(issues, keywordIssue(node.ID, position, SeverityInfo, "platform-dependent-directive", info.Name, "depends on OpenSSH build features"))
		}
	}
	return issues
}

func unknownSeverity(policy UnknownDirectivePolicy) (Severity, bool) {
	switch policy {
	case UnknownDirectiveWarn:
		return SeverityWarning, true
	case UnknownDirectiveError:
		return SeverityError, true
	default:
		return SeverityInfo, false
	}
}

func keywordIssue(id NodeID, position Position, severity Severity, code, name, suffix string) ValidationIssue {
	return ValidationIssue{
		NodeID:   id,
		Position: position,
		Severity: severity,
		Code:     code,
		Message:  fmt.Sprintf("directive %q %s", name, suffix),
	}
}

func (d *Document) nodeIDAtOffset(offset int) NodeID {
	for _, node := range d.nodes {
		if offset >= node.Span.Start && offset < node.Span.End {
			return node.ID
		}
	}
	return 0
}
