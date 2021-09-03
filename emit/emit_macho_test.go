// +build darwin

package emit

import (
	"bytes"
	"debug/macho"
	"testing"
)

func extractText(t *testing.T, b []byte) []byte {
	f, err := macho.NewFile(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("macho failed: %v", err)
	}

	s := f.Section("__text")
	if s == nil {
		t.Fatal("macho missing __text")
	}

	d, err := s.Data()
	if err != nil {
		t.Fatalf("macho data failed: %v", err)
	}

	return d
}
