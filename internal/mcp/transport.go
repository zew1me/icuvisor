package mcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// StreamableHTTPPath is the local HTTP path serving the MCP Streamable HTTP endpoint.
const StreamableHTTPPath = "/mcp"

const streamableHTTPSessionTimeout = 30 * time.Minute
const streamableHTTPShutdownTimeout = 5 * time.Second

// Run serves one MCP session until the client disconnects or ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	if s == nil || s.server == nil || s.transport == nil {
		return errors.New("mcp server is not initialized")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	logger := s.logger
	if logger == nil {
		logger = slog.Default()
	}
	version := s.version
	if version == "" {
		version = "dev"
	}
	serverSession, err := s.server.Connect(ctx, s.transport, nil)
	if err != nil {
		logger.Error("server startup failed", "version", version, "transport", transportName(s.transport), "error", err)
		return err
	}
	logger.Info("server started listening", "version", version, "transport", transportName(s.transport))

	closed := make(chan error, 1)
	go func() {
		defer close(closed)
		closed <- serverSession.Wait()
	}()

	err, cancelled := waitForSessionClose(ctx, closed, func() { serverSession.Close() })
	if cancelled {
		logger.Error("server run cancelled", "version", version, "transport", transportName(s.transport), "error", err)
		return err
	}
	if err != nil {
		logger.Error("server session ended with error", "version", version, "transport", transportName(s.transport), "error", err)
	} else {
		logger.Info("server session ended", "version", version, "transport", transportName(s.transport))
	}
	return err
}

func waitForSessionClose(ctx context.Context, closed <-chan error, closeSession func()) (error, bool) {
	select {
	case <-ctx.Done():
		closeSession()
		<-closed
		return ctx.Err(), true
	case err := <-closed:
		return err, false
	}
}

// RunStreamableHTTP serves the shared MCP server over Streamable HTTP.
func (s *Server) RunStreamableHTTP(ctx context.Context, address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listening for streamable HTTP on %s: %w", address, err)
	}
	return s.ServeStreamableHTTP(ctx, listener)
}

// ServeStreamableHTTP serves Streamable HTTP on listener until ctx is cancelled.
func (s *Server) ServeStreamableHTTP(ctx context.Context, listener net.Listener) error {
	if s == nil || s.server == nil {
		return errors.New("mcp server is not initialized")
	}
	if listener == nil {
		return errors.New("streamable HTTP listener is nil")
	}
	if err := ctx.Err(); err != nil {
		_ = listener.Close()
		return err
	}
	logger := s.logger
	if logger == nil {
		logger = slog.Default()
	}
	version := s.version
	if version == "" {
		version = "dev"
	}

	streamableHandler := sdkmcp.NewStreamableHTTPHandler(func(*http.Request) *sdkmcp.Server {
		return s.server
	}, &sdkmcp.StreamableHTTPOptions{
		Stateless:                  false,
		JSONResponse:               false,
		Logger:                     logger,
		SessionTimeout:             streamableHTTPSessionTimeout,
		DisableLocalhostProtection: false,
		CrossOriginProtection:      nil,
	})
	mux := http.NewServeMux()
	mux.Handle(StreamableHTTPPath, streamableHandler)
	httpServer := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	serveDone := make(chan error, 1)
	go func() {
		serveDone <- httpServer.Serve(listener)
	}()
	logger.Info("server started listening", "version", version, "transport", "streamable_http", "address", listener.Addr().String(), "path", StreamableHTTPPath)

	select {
	case err := <-serveDone:
		return normalizeHTTPServerError(err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), streamableHTTPShutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("streamable HTTP shutdown failed; closing listener", "version", version, "error", err)
			_ = httpServer.Close()
		}
		if err := normalizeHTTPServerError(<-serveDone); err != nil {
			return err
		}
		logger.Info("server session ended", "version", version, "transport", "streamable_http")
		return ctx.Err()
	}
}

func normalizeHTTPServerError(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func transportName(transport sdkmcp.Transport) string {
	switch transport.(type) {
	case *sdkmcp.StdioTransport:
		return "stdio"
	case *sdkmcp.IOTransport:
		return "io"
	case *sdkmcp.InMemoryTransport:
		return "in_memory"
	case *sdkmcp.StreamableServerTransport:
		return "streamable_http"
	default:
		return fmt.Sprintf("%T", transport)
	}
}
