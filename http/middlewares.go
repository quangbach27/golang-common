package http

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
	"unicode/utf8"
	"uuid"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"common/log"
)

const (
	TestNameHeader          = "TestName"
	CorrelationIDHttpHeader = "Correlation-ID"
)

func useMiddlewares(e *echo.Echo, allowedOrigins []string) {
	e.Use(
		corsMiddleware(allowedOrigins),
		middleware.ContextTimeout(10*time.Second),
		middleware.Recover(),
		// Correlation-ID runs first: available in context for the request log middleware.
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				req := c.Request()
				ctx := req.Context()

				reqCorrelationID := req.Header.Get(CorrelationIDHttpHeader)
				if reqCorrelationID == "" {
					reqCorrelationID = uuid.NewV4().String()
				}

				logger := slog.With("correlation_id", reqCorrelationID)

				if testName := c.Request().Header.Get(TestNameHeader); testName != "" {
					logger = logger.With("test_name", testName)
				}

				ctx = log.ToContext(ctx, logger)
				ctx = ContextWithCorrelationID(ctx, reqCorrelationID)
				c.SetRequest(req.WithContext(ctx))
				c.Response().Header().Set(CorrelationIDHttpHeader, reqCorrelationID)

				return next(c)
			}
		},
		requestLogMiddleware,
	)
}

func corsMiddleware(allowedOrigins []string) echo.MiddlewareFunc {
	corsConfig := middleware.CORSConfig{
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			CorrelationIDHttpHeader,
			TestNameHeader,
		},
		ExposeHeaders:    []string{CorrelationIDHttpHeader},
		AllowCredentials: true,
		MaxAge:           300,
	}

	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		// Echo v5 refuses AllowOrigins=["*"] combined with AllowCredentials=true (insecure combination
		// per the CORS spec). UnsafeAllowOriginFunc reflects the request's own Origin back, which is
		// wire-compatible with the old AllowOrigins=["*"] behavior while satisfying that guard.
		slog.Warn("UNSAFE CORS config: all origins allowed with credentials, request Origin is reflected back")
		corsConfig.UnsafeAllowOriginFunc = func(_ *echo.Context, origin string) (string, bool, error) {
			return origin, true, nil
		}
	} else {
		corsConfig.AllowOrigins = allowedOrigins
	}

	return middleware.CORSWithConfig(corsConfig)
}

type bodyCapturingWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *bodyCapturingWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyCapturingWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyCapturingWriter) Flush() {
	err := http.NewResponseController(w.ResponseWriter).Flush()
	if err != nil && !errors.Is(err, http.ErrNotSupported) {
		slog.Warn("response writer flush failed", "error", err)
	}
}

func (w *bodyCapturingWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(w.ResponseWriter).Hijack()
}

func (w *bodyCapturingWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func requestLogMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// Read request body and restore it for the handler.
		var reqBody []byte
		if c.Request().Body != nil {
			reqBody, _ = io.ReadAll(c.Request().Body)
		}
		c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody))

		// Capture response body via MultiWriter.
		resBody := new(bytes.Buffer)
		respWriter := c.Response()
		mw := io.MultiWriter(respWriter, resBody)
		c.SetResponse(&bodyCapturingWriter{Writer: mw, ResponseWriter: respWriter})

		start := time.Now()
		err := next(c)
		duration := time.Since(start)

		ctx := c.Request().Context()

		status := http.StatusOK
		if echoResp, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil {
			status = echoResp.Status
		}

		logger := log.FromContext(ctx).With(
			"URI", c.Request().RequestURI,
			"status", status,
			"method", c.Request().Method,
			"duration", duration.String(),
		)
		if err != nil {
			logger = logger.With("error", err)
		}
		logger = logger.With("request_body", truncateBodyForLog(string(reqBody)))

		body := resBody.String()
		if utf8.ValidString(body) {
			if isDebug := log.FromContext(ctx).Enabled(ctx, slog.LevelDebug); !isDebug {
				body = truncateBodyForLog(body)
			}
			logger = logger.With("response_body", body)
		} else {
			logger = logger.With("response_body", "<binary data>")
		}

		logger.Info("Request done")
		return err
	}
}
