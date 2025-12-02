package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"127.0.0.1", "127.0.0.1", true},
		{"127.0.0.1 with port", "127.0.0.1:8080", true},
		{"localhost", "localhost", true},
		{"localhost with port", "localhost:8080", true},
		{"Remote IP", "192.168.1.1", false},
		{"Remote IP with port", "192.168.1.1:8080", false},
		{"IPv4 different", "10.0.0.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLocalhost(tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}
