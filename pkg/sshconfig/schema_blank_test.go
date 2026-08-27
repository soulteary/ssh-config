package sshconfig

import (
	"bytes"
	"testing"
)

func TestSchemaNewBlankDefaultsToLF(t *testing.T) {
	t.Parallel()
	schema := NewSchema(SchemaDocument{Nodes: []SchemaNode{
		{Type: "blank"},
		{Type: "blank"},
	}})
	doc, err := schema.Document("")
	if err != nil {
		t.Fatal(err)
	}
	got, err := doc.MarshalPreserve()
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte("\n\n"); !bytes.Equal(got, want) {
		t.Fatalf("new blank nodes = %q, want %q", got, want)
	}
}
