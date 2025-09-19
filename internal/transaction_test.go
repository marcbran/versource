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

func TestIsValidBranch(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		expected bool
	}{
		{
			name:     "valid branch name",
			branch:   "feature-branch",
			expected: true,
		},
		{
			name:     "valid branch with numbers",
			branch:   "branch123",
			expected: true,
		},
		{
			name:     "valid branch with underscores",
			branch:   "feature_branch",
			expected: true,
		},
		{
			name:     "valid branch with hyphens",
			branch:   "my-feature-branch",
			expected: true,
		},
		{
			name:     "valid single character",
			branch:   "a",
			expected: true,
		},
		{
			name:     "empty string",
			branch:   "",
			expected: false,
		},
		{
			name:     "starts with period",
			branch:   ".hidden",
			expected: false,
		},
		{
			name:     "contains two periods",
			branch:   "branch..name",
			expected: false,
		},
		{
			name:     "contains @{",
			branch:   "branch@{ref}",
			expected: false,
		},
		{
			name:     "contains colon",
			branch:   "branch:name",
			expected: false,
		},
		{
			name:     "contains question mark",
			branch:   "branch?name",
			expected: false,
		},
		{
			name:     "contains square bracket",
			branch:   "branch[name]",
			expected: false,
		},
		{
			name:     "contains backslash",
			branch:   "branch\\name",
			expected: false,
		},
		{
			name:     "contains caret",
			branch:   "branch^name",
			expected: false,
		},
		{
			name:     "contains tilde",
			branch:   "branch~name",
			expected: false,
		},
		{
			name:     "contains asterisk",
			branch:   "branch*name",
			expected: false,
		},
		{
			name:     "contains space",
			branch:   "branch name",
			expected: false,
		},
		{
			name:     "contains tab",
			branch:   "branch\tname",
			expected: false,
		},
		{
			name:     "contains newline",
			branch:   "branch\nname",
			expected: false,
		},
		{
			name:     "contains carriage return",
			branch:   "branch\rname",
			expected: false,
		},
		{
			name:     "ends with slash",
			branch:   "branch/",
			expected: false,
		},
		{
			name:     "ends with .lock",
			branch:   "branch.lock",
			expected: false,
		},
		{
			name:     "ends with .lock (longer)",
			branch:   "my-branch.lock",
			expected: false,
		},
		{
			name:     "HEAD (uppercase)",
			branch:   "HEAD",
			expected: false,
		},
		{
			name:     "head (lowercase)",
			branch:   "head",
			expected: false,
		},
		{
			name:     "Head (mixed case)",
			branch:   "Head",
			expected: false,
		},
		{
			name:     "hEaD (mixed case)",
			branch:   "hEaD",
			expected: false,
		},
		{
			name:     "valid commit hash format (32 chars)",
			branch:   "a1b2c3d4e5f678901234567890123456",
			expected: false,
		},
		{
			name:     "valid commit hash format with numbers only",
			branch:   "12345678901234567890123456789012",
			expected: false,
		},
		{
			name:     "valid commit hash format with letters only",
			branch:   "abcdefghijklmnopqrstuvwxyzabcdef",
			expected: false,
		},
		{
			name:     "not a commit hash (31 chars)",
			branch:   "a1b2c3d4e5f67890123456789012345",
			expected: true,
		},
		{
			name:     "not a commit hash (33 chars)",
			branch:   "a1b2c3d4e5f6789012345678901234567",
			expected: true,
		},
		{
			name:     "not a commit hash (contains uppercase)",
			branch:   "A1b2c3d4e5f678901234567890123456",
			expected: true,
		},
		{
			name:     "contains non-ASCII character",
			branch:   "branchñame",
			expected: false,
		},
		{
			name:     "contains control character",
			branch:   "branch\x00name",
			expected: false,
		},
		{
			name:     "contains DEL character",
			branch:   "branch\x7fname",
			expected: false,
		},
		{
			name:     "valid branch with dots (not consecutive)",
			branch:   "branch.name.here",
			expected: true,
		},
		{
			name:     "valid branch with @ (not followed by {)",
			branch:   "branch@name",
			expected: true,
		},
		{
			name:     "valid branch with { (not preceded by @)",
			branch:   "branch{name}",
			expected: true,
		},
		{
			name:     "SQL injection attempt - DROP TABLE",
			branch:   "branch'; DROP TABLE users; --",
			expected: false,
		},
		{
			name:     "SQL injection attempt - UNION SELECT",
			branch:   "branch' UNION SELECT * FROM users --",
			expected: false,
		},
		{
			name:     "SQL injection attempt - OR 1=1",
			branch:   "branch' OR 1=1 --",
			expected: false,
		},
		{
			name:     "SQL injection attempt - semicolon",
			branch:   "branch; DROP TABLE users",
			expected: false,
		},
		{
			name:     "SQL injection attempt - single quote",
			branch:   "branch'",
			expected: false,
		},
		{
			name:     "SQL injection attempt - double quote",
			branch:   "branch\"",
			expected: false,
		},
		{
			name:     "SQL injection attempt - backtick",
			branch:   "branch`",
			expected: false,
		},
		{
			name:     "SQL injection attempt - comment",
			branch:   "branch--",
			expected: false,
		},
		{
			name:     "SQL injection attempt - slash comment",
			branch:   "branch/*",
			expected: false,
		},
		{
			name:     "SQL injection attempt - slash star comment",
			branch:   "branch*/",
			expected: false,
		},
		{
			name:     "SQL injection attempt - parentheses",
			branch:   "branch()",
			expected: false,
		},
		{
			name:     "SQL injection attempt - equals",
			branch:   "branch=",
			expected: false,
		},
		{
			name:     "SQL injection attempt - less than",
			branch:   "branch<",
			expected: false,
		},
		{
			name:     "SQL injection attempt - greater than",
			branch:   "branch>",
			expected: false,
		},
		{
			name:     "SQL injection attempt - pipe",
			branch:   "branch|",
			expected: false,
		},
		{
			name:     "SQL injection attempt - ampersand",
			branch:   "branch&",
			expected: false,
		},
		{
			name:     "SQL injection attempt - dollar sign",
			branch:   "branch$",
			expected: false,
		},
		{
			name:     "SQL injection attempt - percent",
			branch:   "branch%",
			expected: false,
		},
		{
			name:     "SQL injection attempt - plus",
			branch:   "branch+",
			expected: false,
		},
		{
			name:     "SQL injection attempt - exclamation",
			branch:   "branch!",
			expected: false,
		},
		{
			name:     "SQL injection attempt - at sign",
			branch:   "branch@",
			expected: true,
		},
		{
			name:     "SQL injection attempt - hash",
			branch:   "branch#",
			expected: false,
		},
		{
			name:     "SQL injection attempt - parentheses with content",
			branch:   "branch(select)",
			expected: false,
		},
		{
			name:     "SQL injection attempt - semicolon with space",
			branch:   "branch ;",
			expected: false,
		},
		{
			name:     "SQL injection attempt - comment with space",
			branch:   "branch --",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidBranch(tt.branch)
			if result != tt.expected {
				t.Errorf("IsValidBranch(%q) = %v, want %v", tt.branch, result, tt.expected)
			}
		})
	}
}
