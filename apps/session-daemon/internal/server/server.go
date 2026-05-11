// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/daytonaio/session-daemon/internal/config"
	"github.com/daytonaio/session-daemon/internal/interpreter"
	"github.com/daytonaio/session-daemon/internal/loadstat"
)

// Server is the HTTP+WS surface of the session-daemon. It listens on
// 127.0.0.1:<port> by default and relies on the proxy chain for auth.
type Server struct {
	cfg       *config.Config
	logger    *slog.Logger
	manager   *interpreter.Manager
	httpSrv   *http.Server
	loadStats *loadstat.Collector
}

// NewServer wires up the manager + worker factories and prepares an HTTP/Gin
// router. It does NOT start the TCP listener — that's [Server.Run]'s job.
func NewServer(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	pyFactory, err := interpreter.NewPythonFactory(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("python factory: %w", err)
	}
	tsFactory, err := interpreter.NewTSFactory(cfg, logger)
	if err != nil {
		// TS support is optional at v1: log and continue with python-only mode if
		// node isn't on PATH. The plan calls this out — daemons where the runtime
		// image lacks node should still serve python contexts.
		logger.Warn("typescript factory unavailable; daemon will serve python only", slog.String("error", err.Error()))
		tsFactory = nil
	}
	bashFactory, err := interpreter.NewBashFactory(cfg, logger)
	if err != nil {
		// Bash support is best-effort like TS: if node / just-bash is missing,
		// log and continue without the bash engine (and the Python bash() bridge).
		logger.Warn("bash factory unavailable; daemon will serve without bash", slog.String("error", err.Error()))
		bashFactory = nil
	}

	// Wire the Python bash() bridge: Python user code shells out to just-bash via
	// the shared bash host. No-op when bash is unavailable (bash() then errors
	// with a clear "bash runtime unavailable" message in the worker).
	if bashFactory != nil {
		pyFactory.SetBashInvoker(bashFactory)
	}

	mgr := interpreter.NewManager(cfg, logger, pyFactory, asWorkerFactory(tsFactory), asBashWorkerFactory(bashFactory))
	return &Server{
		cfg:       cfg,
		logger:    logger,
		manager:   mgr,
		loadStats: loadstat.NewCollector("", cfg.WorkspaceRoot),
	}, nil
}

// asWorkerFactory adapts *TSFactory to the WorkerFactory interface, returning a
// nil interface when the input is nil so the Manager can detect the unavailable case.
func asWorkerFactory(f *interpreter.TSFactory) interpreter.WorkerFactory {
	if f == nil {
		return nil
	}
	return f
}

// asBashWorkerFactory mirrors asWorkerFactory for *BashFactory so a nil concrete
// pointer becomes a nil interface (avoids the typed-nil-in-interface trap).
func asBashWorkerFactory(f *interpreter.BashFactory) interpreter.WorkerFactory {
	if f == nil {
		return nil
	}
	return f
}

// Run starts the HTTP server and blocks until ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggingMiddleware(s.logger))

	r.GET("/healthz", s.handleHealthz)
	r.POST("/sessions", s.handleCreateSession)
	r.GET("/sessions", s.handleListSessions)
	r.DELETE("/sessions/:id", s.handleDeleteSession)
	r.GET("/sessions/:id/execute", s.handleExecute)
	r.GET("/packages", s.handleListPackages)
	r.GET("/load", s.handleLoad)

	addr := net.JoinHostPort(s.cfg.BindAddr, strconv.Itoa(s.cfg.Port))
	s.httpSrv = &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	s.logger.Info("session-daemon listening", slog.String("addr", addr))
	errCh := make(chan error, 1)
	go func() {
		err := s.httpSrv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("session-daemon listen: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.httpSrv.Shutdown(shutdownCtx)
		s.manager.Close()
		return nil
	}
}

func (s *Server) handleHealthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": s.manager.Healthz()})
}

func (s *Server) handleCreateSession(c *gin.Context) {
	var req interpreter.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	ctx, err := s.manager.CreateSession(req)
	if err != nil {
		switch {
		case errors.Is(err, interpreter.ErrContextExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, interpreter.ErrCapacity):
			c.Header("Retry-After", "30")
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		case errors.Is(err, interpreter.ErrUnsupportedLang):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, ctx.SnapshotInfo())
}

func (s *Server) handleListSessions(c *gin.Context) {
	c.JSON(http.StatusOK, s.manager.ListSessions())
}

// handleLoad reports the daemon's self-observed load for the API's scheduler: logical
// concurrency (active/busy contexts vs caps) plus cgroup-aware resource pressure. Resource
// sub-blocks are omitted when cgroup files aren't readable, so the API falls back to
// concurrency-only saturation.
func (s *Server) handleLoad(c *gin.Context) {
	active, busy, pyMax, tsMax, bashMax := s.manager.LoadCounts()
	resp := gin.H{
		"activeContexts": active,
		"busyContexts":   busy,
		"pyMax":          pyMax,
		"tsMax":          tsMax,
		"bashMax":        bashMax,
	}
	sample := s.loadStats.Sample(time.Now())
	if sample.CPU != nil {
		resp["cpu"] = sample.CPU
	}
	if sample.Memory != nil {
		resp["memory"] = sample.Memory
	}
	if sample.IO != nil {
		resp["io"] = sample.IO
	}
	if sample.Disk != nil {
		resp["disk"] = sample.Disk
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleDeleteSession(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}
	if err := s.manager.DeleteSession(id); err != nil {
		if errors.Is(err, interpreter.ErrContextNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) handleListPackages(c *gin.Context) {
	lang := c.Query("language")
	if lang == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "language is required"})
		return
	}
	pkgs, err := s.manager.ListPackages(lang)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pkgs)
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 16384,
	// Loopback-only by design; cross-origin checks aren't useful here. The proxy
	// chain handles tenant/auth, and the daemon binds 127.0.0.1.
	CheckOrigin: func(*http.Request) bool { return true },
}

// handleExecute is the WebSocket entrypoint. The first frame is an ExecuteRequest;
// subsequent frames from the client are ignored (drained for ping/pong).
func (s *Server) handleExecute(c *gin.Context) {
	id := c.Param("id")
	ctx, err := s.manager.GetSession(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "context not found"})
		return
	}

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Warn("ws upgrade failed", slog.String("err", err.Error()))
		return
	}

	cl := ctx.AttachWebSocket(conn, slogTarget{logger: s.logger})

	// Read the first frame as ExecuteRequest, then enqueue.
	conn.SetReadLimit(1 << 20)
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	var req interpreter.ExecuteRequest
	if err := conn.ReadJSON(&req); err != nil {
		s.logger.Warn("ws first-frame read failed", slog.String("err", err.Error()))
		cl.RequestClose(websocket.CloseProtocolError, "expected exec frame")
		return
	}
	_ = conn.SetReadDeadline(time.Time{})

	cl.StartReader()

	timeout := time.Duration(0)
	if req.Timeout != nil && *req.Timeout > 0 {
		timeout = time.Duration(*req.Timeout) * time.Second
	}

	doneCh := ctx.Enqueue(req.Code, req.Envs, timeout, req.Reset)
	go func() {
		result := <-doneCh
		if result.Err != nil {
			// Log the full internal error server-side, but send a generic reason to
			// the client: the raw error string can leak internal detail and WS close
			// reasons are capped at 123 bytes anyway. The CloseInternalServerErr code
			// is preserved so the client can still tell the run failed.
			s.logger.Warn("exec failed",
				slog.String("session", id),
				slog.String("err", result.Err.Error()),
			)
			cl.RequestClose(websocket.CloseInternalServerErr, "internal error")
		} else {
			cl.RequestClose(websocket.CloseNormalClosure, "completed")
		}
	}()

	cl.AwaitDone()
}

func loggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Debug("http",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("dur", time.Since(start)),
		)
	}
}

type slogTarget struct{ logger *slog.Logger }

func (l slogTarget) Debug(msg string, args ...any) { l.logger.Debug(msg, args...) }
func (l slogTarget) Warn(msg string, args ...any)  { l.logger.Warn(msg, args...) }
