package asciiapi

import (
	"testing"
)

func TestNormalizePhoneNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"081234567890", "6281234567890"},
		{"81234567890", "6281234567890"},
		{"+62812-3456-7890", "6281234567890"},
		{"6281234567890@s.whatsapp.net", "6281234567890"},
		{"0812 3456 7890", "6281234567890"},
		{"", ""},
	}

	for _, tt := range tests {
		got := NormalizePhoneNumber(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizePhoneNumber(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}
