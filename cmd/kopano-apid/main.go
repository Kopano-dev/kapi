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

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"stash.kopano.io/kc/kopano-api/cmd"
	"stash.kopano.io/kc/kopano-api/server"
)

func main() {
	cmd.RootCmd.AddCommand(commandServe())

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

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
	serveCmd.Flags().String("gc-socket-path", "", "Parent directory for Kopano Groupware Core unix sockets")

	return serveCmd
}

func serve(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger := &logrus.Logger{
		Out:       os.Stderr,
		Formatter: &logrus.TextFormatter{},
		Level:     logrus.DebugLevel,
	}

	logger.Infoln("serve start")

	var socketPath string

	listenAddr, _ := cmd.Flags().GetString("listen")
	if socketPathString, err := cmd.Flags().GetString("gc-socket-path"); err == nil && socketPathString != "" {
		// TODO(longsleep): validate path
		socketPath = socketPathString
	}
	if socketPath == "" {
		return fmt.Errorf("gc-socket-path is required")
	}

	srv := server.NewServer(listenAddr, socketPath, logger)

	logger.Infof("serve started")
	return srv.Serve(ctx)
}
