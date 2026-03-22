package content

import (
	"testing"
)

func TestIsText(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"empty", []byte{}, true},
		{"ascii text", []byte("hello world\n"), true},
		{"utf8 text", []byte("héllo wörld\n"), true},
		{"binary with null", []byte{0x89, 0x50, 0x4E, 0x47, 0x00}, false},
		{"null at start", []byte{0x00, 0x41, 0x42}, false},
		{"long text no null", []byte("abcdefghijklmnopqrstuvwxyz"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsText(tt.data)
			if got != tt.want {
				t.Errorf("IsText(%q) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestIsDiffable(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"small text", []byte("hello"), true},
		{"binary", []byte{0x00, 0x01}, false},
		{"too large", make([]byte, maxTextFileSize+1), false},
		{"exactly max size", make([]byte, maxTextFileSize), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill large slices with 'a' to avoid null bytes
			if len(tt.data) > 10 && tt.data[0] == 0 && tt.name != "binary" {
				for i := range tt.data {
					tt.data[i] = 'a'
				}
			}
			got := IsDiffable(tt.data)
			if got != tt.want {
				t.Errorf("IsDiffable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"etc/nginx.conf", "etc/nginx.conf"},
		{"./etc/nginx.conf", "etc/nginx.conf"},
		{"/etc/nginx.conf", "etc/nginx.conf"},
		{"usr/", "usr"},
		{"./usr/bin/", "usr/bin"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizePath(tt.input)
			if got != tt.want {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
