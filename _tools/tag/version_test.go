package main

import "testing"

func TestBumpVersion(t *testing.T) {
	tests := []struct {
		version  string
		level    string
		expected string
	}{
		{"v0.7.10", "major", "v0.8.0"},
		{"v0.7.10", "minor", "v0.7.11"},
		{"v0.7.10", "patch", "v0.7.10-patch.1"},
		{"v0.7.10-patch.1", "patch", "v0.7.10-patch.2"},
		{"v0.7.10-patch.2", "patch", "v0.7.10-patch.3"},
		{"v0.7.10-patch.3", "minor", "v0.7.11"},
		{"v0.7.10-patch.3", "major", "v0.8.0"},
		{"0.7.10", "major", "v0.8.0"},
		{"0.7.10", "minor", "v0.7.11"},
		{"0.7.10", "patch", "v0.7.10-patch.1"},
		{"v1.0.0", "major", "v0.1.0"},
		{"v1.0.0", "minor", "v0.0.1"},
		{"v1.0.0", "patch", "v0.0.0-patch.1"},
		{"v0.1", "minor", "v0.1.1"},
	}

	for _, tt := range tests {
		got := bumpVersion(tt.version, tt.level)
		if got != tt.expected {
			t.Errorf("bumpVersion(%q, %q) = %q; want %q", tt.version, tt.level, got, tt.expected)
		}
	}
}
