package internal

import (
	"testing"
)

func TestIsValidCommitHash(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		expected bool
	}{
		{
			name:     "valid commit hash",
			hash:     "a1b2c3d4e5f678901234567890123456",
			expected: true,
		},
		{
			name:     "valid commit hash with numbers only",
			hash:     "12345678901234567890123456789012",
			expected: true,
		},
		{
			name:     "valid commit hash with letters only",
			hash:     "abcdefghijklmnopqrstuvwxyzabcdef",
			expected: true,
		},
		{
			name:     "too short",
			hash:     "a1b2c3d4e5f6789012345678901234",
			expected: false,
		},
		{
			name:     "too long",
			hash:     "a1b2c3d4e5f6789012345678901234567890abcd",
			expected: false,
		},
		{
			name:     "contains uppercase letters",
			hash:     "A1b2c3d4e5f6789012345678901234567890abcd",
			expected: false,
		},
		{
			name:     "contains special characters",
			hash:     "a1b2c3d4e5f6789012345678901234567890ab!d",
			expected: false,
		},
		{
			name:     "contains spaces",
			hash:     "a1b2c3d4e5f6789012345678901234567890ab d",
			expected: false,
		},
		{
			name:     "empty string",
			hash:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidCommitHash(tt.hash)
			if result != tt.expected {
				t.Errorf("IsValidCommitHash(%q) = %v, want %v", tt.hash, result, tt.expected)
			}
		})
	}
}
