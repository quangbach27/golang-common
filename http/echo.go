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

	"common/http/auth"
)

// publicPathSegment marks routes that need no authentication: every route
// registered under <APIPrefix>/public/.
const publicPathSegment = "/public/"

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
	globalRouter    *echo.Echo
	protectedRouter *echo.Group

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

	router.GET("/healthz", func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	return &EchoServer{
		globalRouter:    router,
		protectedRouter: router.Group(""),
		config:          config,
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

	server.protectedRouter.Use(
		auth.Middleware(
			tokenVerifier,
			func(c *echo.Context) bool {
				return strings.HasPrefix(c.Path(), publicPathSegment)
			},
		),
	)

	return server, nil
}

func (e *EchoServer) GlobalRouter() EchoRouter {
	return e.globalRouter
}

func (e *EchoServer) ProtectedRouter() EchoRouter {
	return e.protectedRouter
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

	if err := startConfig.Start(ctx, e.globalRouter); err != nil {
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

	return errors.Join(errs...)
}
