package http

import (
	"strings"
	"testing"
	"time"
)

func validConfig() *EchoServerConfig {
	c := &EchoServerConfig{}
	c.setDefault()
	return c
}

func TestEchoServerConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(c *EchoServerConfig)
		wantErr   bool
		wantParts []string
	}{
		{
			name:    "defaults are valid",
			mutate:  func(c *EchoServerConfig) {},
			wantErr: false,
		},
		{
			name: "port too low",
			mutate: func(c *EchoServerConfig) {
				c.Port = 0
			},
			wantErr:   true,
			wantParts: []string{"port must be between 1 and 65535"},
		},
		{
			name: "port too high",
			mutate: func(c *EchoServerConfig) {
				c.Port = 70000
			},
			wantErr:   true,
			wantParts: []string{"port must be between 1 and 65535"},
		},
		{
			name: "negative graceful timeout",
			mutate: func(c *EchoServerConfig) {
				c.GracefulTimeout = -time.Second
			},
			wantErr:   true,
			wantParts: []string{"graceful timeout must be positive"},
		},
		{
			name: "zero idle timeout",
			mutate: func(c *EchoServerConfig) {
				c.IdleTimeout = 0
			},
			wantErr:   true,
			wantParts: []string{"idle timeout must be positive"},
		},
		{
			name: "zero read header timeout",
			mutate: func(c *EchoServerConfig) {
				c.ReadHeaderTimeout = 0
			},
			wantErr:   true,
			wantParts: []string{"read header timeout must be positive"},
		},
		{
			name: "empty allowed origins",
			mutate: func(c *EchoServerConfig) {
				c.AllowedOrigins = nil
			},
			wantErr:   true,
			wantParts: []string{"allowed origins must not be empty"},
		},
		{
			name: "blank allowed origin entry",
			mutate: func(c *EchoServerConfig) {
				c.AllowedOrigins = []string{"http://localhost:3000", "  "}
			},
			wantErr:   true,
			wantParts: []string{"allowed origins must not contain blank entries"},
		},
		{
			name: "wildcard origin is valid",
			mutate: func(c *EchoServerConfig) {
				c.AllowedOrigins = []string{"*"}
			},
			wantErr: false,
		},
		{
			name: "multiple violations are all reported",
			mutate: func(c *EchoServerConfig) {
				c.Port = -1
				c.GracefulTimeout = 0
				c.AllowedOrigins = nil
			},
			wantErr: true,
			wantParts: []string{
				"port must be between 1 and 65535",
				"graceful timeout must be positive",
				"allowed origins must not be empty",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig()
			tt.mutate(c)

			err := c.validate()

			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			for _, part := range tt.wantParts {
				if !strings.Contains(err.Error(), part) {
					t.Errorf("expected error to contain %q, got: %v", part, err)
				}
			}
		})
	}
}

func TestNewEchoServer_InvalidConfig(t *testing.T) {
	_, err := NewEchoServer(func(c *EchoServerConfig) {
		c.Port = -1
	})
	if err == nil {
		t.Fatal("expected error for invalid port, got nil")
	}
	if !strings.Contains(err.Error(), "invalid config") {
		t.Errorf("expected error to be wrapped with 'invalid config', got: %v", err)
	}
}

func TestNewEchoServer_ValidConfig(t *testing.T) {
	server, err := NewEchoServer()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if server == nil {
		t.Fatal("expected non-nil server")
	}
}
