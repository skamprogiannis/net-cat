package main

import "testing"

func TestResolvePort(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantPort string
		wantOK   bool
	}{
		{"no argument uses the default port", []string{"./TCPChat"}, "8989", true},
		{"one argument uses that port", []string{"./TCPChat", "2525"}, "2525", true},
		{"too many arguments is a usage error", []string{"./TCPChat", "2525", "localhost"}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPort, gotOK := resolvePort(tt.args)
			if gotPort != tt.wantPort || gotOK != tt.wantOK {
				t.Fatalf("resolvePort(%q) = (%q, %v), want (%q, %v)",
					tt.args, gotPort, gotOK, tt.wantPort, tt.wantOK)
			}
		})
	}
}
