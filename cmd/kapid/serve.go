/*
 * Copyright 2018 Kopano and its licensors
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

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"

	"stash.kopano.io/kc/kapi/server"
)

// Defaults.
const (
	defaultListenAddr = "127.0.0.1:8039"
)

func commandServe() *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve [...args]",
		Short: "Start server and listen for requests",
		Run: func(cmd *cobra.Command, args []string) {
			if err := serve(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
	serveCmd.Flags().String("listen", defaultListenAddr, "TCP listen address")
	serveCmd.Flags().String("plugins-path", "", "Historic unused parameter")
	serveCmd.Flags().String("plugins", "", "Enabled plugin IDs. When empty, all found plugins are enabled. Separate multiple IDs with comma.")
	serveCmd.Flags().String("iss", "", "OIDC issuer URL")
	serveCmd.Flags().Bool("insecure", false, "Disable TLS certificate and hostname validation")
	serveCmd.Flags().Bool("log-timestamp", true, "Prefix each log line with timestamp")
	serveCmd.Flags().String("log-level", "info", "Log level (one of panic, fatal, error, warn, info or debug)")
	serveCmd.Flags().Bool("with-pprof", false, "With pprof enabled")
	serveCmd.Flags().String("pprof-listen", "127.0.0.1:6060", "TCP listen address for pprof")
	serveCmd.Flags().Bool("with-metrics", false, "Enable metrics")
	serveCmd.Flags().String("metrics-listen", "127.0.0.1:6039", "TCP listen address for metrics")

	return serveCmd
}

func serve(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	logTimestamp, _ := cmd.Flags().GetBool("log-timestamp")
	logLevel, _ := cmd.Flags().GetString("log-level")

	logger, err := newLogger(!logTimestamp, logLevel)
	if err != nil {
		return fmt.Errorf("failed to create logger: %v", err)
	}
	logger.Infoln("serve start")

	listenAddr, _ := cmd.Flags().GetString("listen")

	enabledPlugins := make([]string, 0)
	if pluginsString, strErr := cmd.Flags().GetString("plugins"); strErr == nil && pluginsString != "" {
		for _, id := range strings.Split(pluginsString, ",") {
			if id == "none" {
				enabledPlugins = nil
				break
			}
			enabledPlugins = append(enabledPlugins, strings.TrimSpace(id))
		}
	}
	if len(enabledPlugins) > 0 || enabledPlugins == nil {
		logger.Debugf("enabled plugins: %v", enabledPlugins)
	} else {
		logger.Debug("all plugins enabled")
	}

	var iss *url.URL
	if issString, parseErr := cmd.Flags().GetString("iss"); parseErr == nil && issString != "" {
		iss, parseErr = url.Parse(issString)
		if parseErr != nil {
			return fmt.Errorf("invalid iss url: %v", parseErr)
		}
	}
	if iss == nil {
		return fmt.Errorf("missing --iss parameter")
	}

	var tlsClientConfig *tls.Config
	tlsInsecureSkipVerify, _ := cmd.Flags().GetBool("insecure")
	if tlsInsecureSkipVerify {
		// NOTE(longsleep): This disable http2 client support. See https://github.com/golang/go/issues/14275 for reasons.
		tlsClientConfig = &tls.Config{
			InsecureSkipVerify: tlsInsecureSkipVerify,
		}
		logger.Warnln("insecure mode, TLS client connections are susceptible to man-in-the-middle attacks")
		logger.Debugln("http2 client support is disabled (insecure mode)")
	}
	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       tlsClientConfig,
		},
	}

	// Metrics support.
	withMetrics, _ := cmd.Flags().GetBool("with-metrics")
	metricsListenAddr, _ := cmd.Flags().GetString("metrics-listen")
	if withMetrics && metricsListenAddr != "" {
		go func() {
			metricsListen := metricsListenAddr
			handler := http.NewServeMux()
			logger.WithField("listenAddr", metricsListen).Infoln("metrics enabled, starting listener")
			handler.Handle("/metrics", promhttp.Handler())
			listenErr := http.ListenAndServe(metricsListen, handler)
			if listenErr != nil {
				logger.WithError(listenErr).Errorln("unable to start metrics listener")
			}
		}()
	}

	srv, err := server.NewServer(listenAddr, "", iss, enabledPlugins, logger, client)
	if err != nil {
		return err
	}

	// Profiling support.
	withPprof, _ := cmd.Flags().GetBool("with-pprof")
	pprofListenAddr, _ := cmd.Flags().GetString("pprof-listen")
	if withPprof && pprofListenAddr != "" {
		runtime.SetMutexProfileFraction(5)
		go func() {
			pprofListen := pprofListenAddr
			logger.WithField("listenAddr", pprofListen).Infoln("pprof enabled, starting listener")
			err := http.ListenAndServe(pprofListen, nil)
			if err != nil {
				logger.WithError(err).Errorln("unable to start pprof listener")
			}
		}()
	}

	logger.Infof("serve started")
	return srv.Serve(ctx)
}
