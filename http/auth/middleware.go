package auth

import (
	"common"
	"context"
	"strings"

	"github.com/labstack/echo/v5"
)

const bearerPrefix = "Bearer "

func Middleware(verifier TokenVerifier, skip func(c *echo.Context) bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if skip != nil && skip(c) {
				return next(c)
			}

			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)

			if !strings.HasPrefix(authHeader, bearerPrefix) {
				return common.NewUnauthorizedError(
					"missing_bearer_token",
					"missing or malformed authorization header",
				)
			}

			accessToken := strings.TrimPrefix(authHeader, bearerPrefix)
			if accessToken == "" {
				return common.NewUnauthorizedError(
					"missing_bearer_token",
					"missing or malformed authorization header",
				)
			}

			req := c.Request()

			session, err := verifier.Verify(req.Context(), accessToken)
			if err != nil {
				return common.NewUnauthorizedError(
					"invalid_access_token",
					"invalid or expired access token",
				).WithInternalError(err)
			}

			c.SetRequest(
				req.WithContext(
					WithSession(req.Context(), session),
				),
			)

			return next(c)
		}
	}
}

type Session struct {
	UserID string
	Roles  []string
	Extra  map[string]any
}

type sessionKey struct{}

func WithSession(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionKey{}, s)
}

func SessionFromContext(ctx context.Context) (*Session, bool) {
	s, ok := ctx.Value(sessionKey{}).(*Session)
	return s, ok
}

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (*Session, error)
}
