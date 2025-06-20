package service

import (
	"testing"
)

func TestIsValidUserName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty", "", false},
		{"one character", "a", false},
		{"two characters", "ab", true},
		{"numeric", "1234", true},
		{"alphanumeric", "user123", true},
		{"underscore", "user_name", true},
		{"hyphen", "user-name", true},
		{"mixed valid", "user_name-123", true},
		{"space", "user name", false},
		{"special char", "user!name", false},
		{"dot", "user.name", false},
		{"leading hyphen", "-user", true},
		{"trailing underscore", "user_", true},
		{"leading space", " user", false},
		{"trailing space", " user", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidUserName(tt.input); got != tt.want {
				t.Errorf("IsValidUserName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
