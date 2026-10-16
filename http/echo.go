package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/quangbach27/golang-common/http/auth"
)

type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.RouteInfo
}

type EchoServer struct {
	router *echo.Echo

	config *EchoServerConfig
}

func NewEchoServer(updateConfigFns ...func(config *EchoServerConfig)) (*EchoServer, error) {
	router := echo.New()

	config, err := initConfig(updateConfigFns...)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	useMiddlewares(router, config.AllowedOrigins)

	router.HTTPErrorHandler = EchoErrorHandler
	router.Logger = slog.Default()

	return &EchoServer{
		router: router,
		config: config,
	}, nil
}

func NewEchoServerWithTokenVerifier(
	tokenVerifier auth.TokenVerifier,
	updateConfigFns ...func(config *EchoServerConfig),
) (*EchoServer, error) {
	server, err := NewEchoServer(updateConfigFns...)
	if err != nil {
		return nil, err
	}

	server.router.Use(auth.Middleware(tokenVerifier, server.config.AuthConfig.skipper()))

	return server, nil
}

func (e *EchoServer) Router() EchoRouter {
	return e.router
}

func (e *EchoServer) Start(ctx context.Context) error {
	startConfig := echo.StartConfig{
		Address:         fmt.Sprintf(":%d", e.config.Port),
		HideBanner:      true,
		GracefulTimeout: e.config.GracefulTimeout,
		BeforeServeFunc: func(s *http.Server) error {
			s.IdleTimeout = e.config.IdleTimeout
			s.ReadHeaderTimeout = e.config.ReadHeaderTimeout
			return nil
		},
	}

	if err := startConfig.Start(ctx, e.router); err != nil {
		return fmt.Errorf("starting http server failed: %w", err)
	}

	return nil
}

type EchoServerConfig struct {
	Port              int
	GracefulTimeout   time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	AllowedOrigins    []string
	AuthConfig        AuthConfig
}

func initConfig(updateConfigFns ...func(config *EchoServerConfig)) (*EchoServerConfig, error) {
	config := &EchoServerConfig{}
	config.setDefault()

	for _, updateFn := range updateConfigFns {
		updateFn(config)
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *EchoServerConfig) setDefault() {
	c.Port = 4000
	c.GracefulTimeout = 60 * time.Second
	c.IdleTimeout = 60 * time.Second
	c.ReadHeaderTimeout = 45 * time.Second
	c.AllowedOrigins = []string{"http://localhost:3000"}
	c.AuthConfig = AuthConfig{
		PublicPaths: []string{"/healthz"},
	}
}

func (c *EchoServerConfig) validate() error {
	var errs []error

	if c.Port < 1 || c.Port > 65535 {
		errs = append(errs, fmt.Errorf("port must be between 1 and 65535, got %d", c.Port))
	}
	if c.GracefulTimeout <= 0 {
		errs = append(errs, fmt.Errorf("graceful timeout must be positive, got %s", c.GracefulTimeout))
	}
	if c.IdleTimeout <= 0 {
		errs = append(errs, fmt.Errorf("idle timeout must be positive, got %s", c.IdleTimeout))
	}
	if c.ReadHeaderTimeout <= 0 {
		errs = append(errs, fmt.Errorf("read header timeout must be positive, got %s", c.ReadHeaderTimeout))
	}
	if len(c.AllowedOrigins) == 0 {
		errs = append(errs, errors.New("allowed origins must not be empty"))
	}
	for _, origin := range c.AllowedOrigins {
		if strings.TrimSpace(origin) == "" {
			errs = append(errs, errors.New("allowed origins must not contain blank entries"))
			break
		}
	}
	if err := c.AuthConfig.validate(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	return nil
}

// AuthConfig lists the routes that bypass the auth middleware. Entries are
// matched against the registered route pattern (c.Path()).
type AuthConfig struct {
	// PublicPaths are routes that skip auth when the route pattern equals an
	// entry exactly. E.g. "/login" skips "/login" but not "/login/reset".
	PublicPaths []string
	// PublicPrefixes are routes that skip auth when the route pattern starts
	// with an entry. E.g. "/public/" skips "/public/docs" and "/public/a/b".
	PublicPrefixes []string
}

func (a AuthConfig) skipper() func(c *echo.Context) bool {
	paths := make(map[string]struct{}, len(a.PublicPaths))
	for _, p := range a.PublicPaths {
		paths[p] = struct{}{}
	}

	return func(c *echo.Context) bool {
		path := c.Path()
		if _, ok := paths[path]; ok {
			return true
		}
		for _, prefix := range a.PublicPrefixes {
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
		return false
	}
}

func (a AuthConfig) validate() error {
	var errs []error

	check := func(name string, entries []string) {
		for _, e := range entries {
			if !strings.HasPrefix(e, "/") {
				errs = append(errs, fmt.Errorf("auth %s entry must start with '/', got %q", name, e))
			}
		}
	}

	check("public paths", a.PublicPaths)
	check("public prefixes", a.PublicPrefixes)

	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	return nil
}
