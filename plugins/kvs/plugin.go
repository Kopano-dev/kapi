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
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"stash.kopano.io/kc/kapi/plugins"
	"stash.kopano.io/kc/kapi/plugins/kvs/kv"
	"stash.kopano.io/kc/kapi/version"
)

const (
	storeInitializeRetryInterval = 10 * time.Second
)

var pluginInfo = &plugins.InfoV1{
	ID:        "kvs",
	Version:   version.Version,
	BuildDate: version.BuildDate,
}

// KVSPlugin implements a key value store for Kopano API.
type KVSPlugin struct {
	ctx context.Context
	srv plugins.ServerV1

	cors *cors.Cors

	quit    chan struct{}
	handler http.Handler
	store   *kv.KV
}

// Info returns the accociated plugins plugin.Info.
func (p *KVSPlugin) Info() *plugins.InfoV1 {
	return pluginInfo
}

// Initialize initizalizes the accociated plugin.
func (p *KVSPlugin) Initialize(ctx context.Context, errCh chan<- error, srv plugins.ServerV1) error {
	p.ctx = ctx
	p.srv = srv

	dbDataSourceName := os.Getenv("KOPANO_KVS_DB_DATASOURCE")
	dbMigrationsPath := os.Getenv("KOPANO_KVS_DB_MIGRATIONS")
	dbDriverName := os.Getenv("KOPANO_KVS_DB_DRIVER")

	p.handler = p.addRoutes(ctx, mux.NewRouter())

	store, err := kv.New(dbDriverName, dbDataSourceName, dbMigrationsPath, srv.Logger())
	if err != nil {
		return fmt.Errorf("failed to create store: %v", err)
	}
	p.store = store
	go func() {
		for {
			initializeErr := store.Initialize(ctx)
			if initializeErr != nil {
				p.srv.Logger().Errorf("kvs: store initialize failed: %v", initializeErr)
			} else {
				p.srv.Logger().Debugln("kvs: store initialize complete")
				return
			}
			select {
			case <-p.quit:
				return
			case <-ctx.Done():
				return
			case <-time.After(storeInitializeRetryInterval):
				// Retry after short timeout.
			}
		}
	}()

	if os.Getenv("KOPANO_KVS_ALLOW_CORS") == "1" {
		p.srv.Logger().Warnln("kvs: CORS support enabled")
		p.cors = cors.AllowAll()
	}

	srv.Logger().Debugln("kvs: initialize")
	return nil
}

// Close closes the accociated plugin.
func (p *KVSPlugin) Close() error {
	p.srv.Logger().Debugln("kvs: close")
	close(p.quit)

	err := p.store.Close()
	if err != nil {
		p.srv.Logger().WithError(err).Warnln("kvs: failed to close database")
	}

	return nil
}

// Register is the exported registration entry point as loaded by Kopano API to
// register plugins.
var Register plugins.RegisterPluginV1 = func() plugins.PluginV1 {
	return &KVSPlugin{
		quit: make(chan struct{}),
	}
}

// NOTE(longsleep): Keep main() to make the linter happy.
func main() {}
