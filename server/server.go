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
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/longsleep/go-metrics/loggedwriter"
	"github.com/longsleep/go-metrics/timing"
	"github.com/sirupsen/logrus"

	"stash.kopano.io/kc/kopano-api/proxy"
)

// Server represents the base for a HTTP server.
type Server struct {
	mutex sync.RWMutex

	listenAddr string
	socketPath string
	logger     logrus.FieldLogger

	proxy proxy.Proxy
}

// NewServer creates a new Server with the provided parameters.
func NewServer(listenAddr string, socketPath string, logger logrus.FieldLogger) *Server {
	s := &Server{
		listenAddr: listenAddr,
		socketPath: socketPath,
		logger:     logger,
	}

	return s
}

// ServerHTTP implements the http.HandlerFunc interface.
func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	switch path := req.URL.Path; {
	case path == "/health-check":
		s.HealthCheckHandler(rw, req)
	case strings.HasPrefix(path, "/api/gc/v0/"):
		s.ProxyHandler(rw, req)
	default:
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

		// Run the request.
		next.ServeHTTP(loggedWriter, req.WithContext(ctx))
		// Cancel per request context when done.
		cancel()
	})
}

// Serve is the accociated Server's main blocking runner.
func (s *Server) Serve(ctx context.Context) error {
	serveCtx, serveCtxCancel := context.WithCancel(ctx)
	defer serveCtxCancel()

	logger := s.logger

	errCh := make(chan error, 2)
	exitCh := make(chan bool, 1)
	signalCh := make(chan os.Signal)

	go func() {
		var err error
		for {
			for {
				socketPaths, globErr := filepath.Glob(fmt.Sprintf("%s/*.sock", s.socketPath))
				if globErr != nil {
					err = globErr
					break
				}
				if len(socketPaths) == 0 {
					err = fmt.Errorf("no .sock files found in socket-path")
					break
				}

				p, proxyErr := proxy.New(socketPaths)
				if err != nil {
					errCh <- proxyErr
					return
				}

				s.mutex.Lock()
				s.proxy = p
				logger.Debugf("using %d upstream workers", len(socketPaths))
				s.mutex.Unlock()
				return
			}

			if err != nil {
				logger.WithError(err).Warnln("proxy start delayed")
			}

			select {
			case <-exitCh:
				return
			case <-time.After(1 * time.Second):
				// retry.
			}
		}
	}()

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

	if s.proxy != nil {
		// TODO(longsleep): close upstreams cleanly.
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
