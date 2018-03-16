/*
 * Copyright 2017 Kopano and its licensors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/longsleep/go-metrics/loggedwriter"
	"github.com/longsleep/go-metrics/timing"
	"github.com/sirupsen/logrus"

	"stash.kopano.io/kc/kapi/plugins"
)

// Server represents the base for a HTTP server.
type Server struct {
	listenAddr  string
	pluginsPath string
	logger      logrus.FieldLogger

	plugins        []plugins.PluginV1
	enabledPlugins map[string]bool

	requestLog           bool
	allowAuthPassthrough bool
}

// NewServer creates a new Server with the provided parameters.
func NewServer(listenAddr string, pluginsPath string, enabledPlugins map[string]bool, logger logrus.FieldLogger) *Server {
	s := &Server{
		listenAddr:     listenAddr,
		pluginsPath:    pluginsPath,
		enabledPlugins: enabledPlugins,
		logger:         logger,

		requestLog:           os.Getenv("KOPANO_DEBUG_SERVER_REQUEST_LOG") == "1",
		allowAuthPassthrough: os.Getenv("KOPANO_ALLOW_AUTH_PASSTHROUGH") == "1",
	}

	s.loadPlugins()

	return s
}

// ServerHTTP implements the http.HandlerFunc interface.
func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	switch path := req.URL.Path; {
	case path == "/health-check":
		s.HealthCheckHandler(rw, req)

	default:
		// Try all registered plugins.
		for _, p := range s.plugins {
			handled, err := p.ServeHTTP(rw, req)
			if err != nil {
				s.logger.WithError(err).Errorf("error in plugin http handler: %#v", p)
				http.Error(rw, "", http.StatusInternalServerError)
				return
			}
			if handled {
				// Done.
				return
			}
		}

		// If nothing felt responsible, 404.
		http.NotFound(rw, req)
	}
}

// AddContext adds the accociated server context with cancel to the the provided
// httprouter.Handle. When the handler is done, the per Request context is
// cancelled.
func (s *Server) AddContext(parent context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Create per request context.
		ctx, cancel := context.WithCancel(parent)

		if s.requestLog {
			loggedWriter := metrics.NewLoggedResponseWriter(rw)
			// Create per request context.
			ctx = timing.NewContext(ctx, func(duration time.Duration) {
				// This is the stop callback, called when complete with duration.
				durationMs := float64(duration) / float64(time.Millisecond)
				// Log request.
				s.logger.WithFields(logrus.Fields{
					"status":     loggedWriter.Status(),
					"method":     req.Method,
					"path":       req.URL.Path,
					"remote":     req.RemoteAddr,
					"duration":   durationMs,
					"referer":    req.Referer(),
					"user-agent": req.UserAgent(),
					"origin":     req.Header.Get("Origin"),
				}).Debug("HTTP request complete")
			})
			rw = loggedWriter
		}

		// Run the request.
		next.ServeHTTP(rw, req.WithContext(ctx))

		// Cancel per request context when done.
		cancel()
	})
}

// Logger returns the accociated logger.
func (s *Server) Logger() logrus.FieldLogger {
	return s.logger
}

// Serve is the accociated Server's main blocking runner.
func (s *Server) Serve(ctx context.Context) error {
	serveCtx, serveCtxCancel := context.WithCancel(ctx)
	defer serveCtxCancel()

	logger := s.logger

	errCh := make(chan error, 2)
	exitCh := make(chan bool, 1)
	signalCh := make(chan os.Signal)

	// Plugins.
	for _, p := range s.plugins {
		if pluginErr := p.Initialize(serveCtx, errCh, s); pluginErr != nil {
			return fmt.Errorf("failed to initialize plugin %T: %v", p, pluginErr)
		}
	}

	// HTTP listener.
	srv := &http.Server{
		Handler: s.AddContext(serveCtx, s),
	}

	logger.WithField("listenAddr", s.listenAddr).Infoln("starting http listener")
	listener, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}

	logger.Infoln("ready to handle requests")

	go func() {
		serveErr := srv.Serve(listener)
		if serveErr != nil {
			errCh <- serveErr
		}

		logger.Debugln("http listener stopped")
		close(exitCh)
	}()

	// Wait for exit or error.
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err = <-errCh:
		// breaks
	case reason := <-signalCh:
		logger.WithField("signal", reason).Warnln("received signal")
		// breaks
	}

	// Shutdown, server will stop to accept new connections, requires Go 1.8+.
	logger.Infoln("clean server shutdown start")
	shutDownCtx, shutDownCtxCancel := context.WithTimeout(ctx, 10*time.Second)
	if shutdownErr := srv.Shutdown(shutDownCtx); shutdownErr != nil {
		logger.WithError(shutdownErr).Warn("clean server shutdown failed")
	}

	// Close plugins.
	for _, p := range s.plugins {
		if closeErr := p.Close(); closeErr != nil {
			logger.WithError(err).Debugln("failed to close plugin %T: %v", p, closeErr)
		}
	}

	// Cancel our own context, wait on managers.
	serveCtxCancel()
	func() {
		for {
			select {
			case <-exitCh:
				return
			default:
				// HTTP listener has not quit yet.
				logger.Info("waiting for http listener to exit")
			}
			select {
			case reason := <-signalCh:
				logger.WithField("signal", reason).Warn("received signal")
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}()
	shutDownCtxCancel() // prevent leak.

	return err
}
