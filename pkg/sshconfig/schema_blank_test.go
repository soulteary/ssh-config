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

func TestSchemaAppendedBlankRemainsAPhysicalLine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input []byte
		want  []byte
	}{
		{input: []byte("Host example"), want: []byte("Host example\n\n")},
		{input: []byte("Host example\n"), want: []byte("Host example\n\n")},
		{input: []byte("Host example\r\n"), want: []byte("Host example\r\n\r\n")},
		{input: []byte("Host example\r"), want: []byte("Host example\r\r")},
	}
	for _, test := range tests {
		doc, err := Parse(test.input)
		if err != nil {
			t.Fatal(err)
		}
		schemaDocument, err := doc.ToSchema("")
		if err != nil {
			t.Fatal(err)
		}
		schemaDocument.Nodes = append(schemaDocument.Nodes, SchemaNode{Type: "blank"})
		reconstructed, err := NewSchema(schemaDocument).Document("")
		if err != nil {
			t.Fatal(err)
		}
		got, err := reconstructed.MarshalPreserve()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, test.want) {
			t.Fatalf("appended blank = %q, want %q", got, test.want)
		}
	}
}
