package server

import "testing"

func TestParseFileMode(t *testing.T) {
	mode := ParseFileMode("-rw-------")
	if mode != 0600 {
		t.Errorf("Expected 0600, got %x", mode)
	}
	mode = ParseFileMode("-rw-r-x---")
	if mode != 0650 {
		t.Errorf("Expected 0650, got %x", mode)
	}
}
