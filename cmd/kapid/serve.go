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
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"stash.kopano.io/kc/kapi/server"
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
	serveCmd.Flags().String("listen", "127.0.0.1:8039", "TCP listen address")
	serveCmd.Flags().String("plugins-path", "./plugins", "Directory where to find plugin .so files")
	serveCmd.Flags().String("plugins", "", "Enabled plugin IDs. When empty, all found plugins are enabled. Seperate multiple IDs with comma.")
	serveCmd.Flags().String("iss", "", "OIDC issuer URL")
	serveCmd.Flags().Bool("insecure", false, "Disable TLS certificate and hostname validation")
	serveCmd.Flags().Bool("log-timestamp", true, "Prefix each log line with timestamp")
	serveCmd.Flags().String("log-level", "info", "Log level (one of panic, fatal, error, warn, info or debug)")

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

	var pluginsPath string

	listenAddr, _ := cmd.Flags().GetString("listen")

	if pluginsPathString, err := cmd.Flags().GetString("plugins-path"); err == nil && pluginsPathString != "" {
		pluginsPath, err = filepath.Abs(pluginsPathString)
		if err != nil {
			return fmt.Errorf("invalid plugins-path: %v", err)
		}
	}
	logger.Infof("loading plugins from %s", pluginsPath)

	var enabledPlugins map[string]bool
	if pluginsString, err := cmd.Flags().GetString("plugins"); err == nil && pluginsString != "" {
		enabledPlugins = make(map[string]bool)
		for _, id := range strings.Split(pluginsString, ",") {
			enabledPlugins[strings.TrimSpace(id)] = true
		}
	}
	logger.Debugf("enabled plugins: %#v", enabledPlugins)

	var iss *url.URL
	if issString, err := cmd.Flags().GetString("iss"); err == nil && issString != "" {
		iss, err = url.Parse(issString)
		if err != nil {
			return fmt.Errorf("invalid iss url: %v", err)
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

	srv, err := server.NewServer(listenAddr, pluginsPath, iss, enabledPlugins, logger, client)
	if err != nil {
		return err
	}

	logger.Infof("serve started")
	return srv.Serve(ctx)
}
