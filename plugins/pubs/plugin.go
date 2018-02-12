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

	"encoding/hex"
	"github.com/cskr/pubsub"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/websocket"
	"stash.kopano.io/kgol/rndm"

	"stash.kopano.io/kc/kopano-api/plugins"
	"stash.kopano.io/kc/kopano-api/version"
)

// Buffer sizes for websocket payload data.
const (
	websocketReadBufferSize  = 1024 * 10
	websocketWriteBufferSize = 1024 * 10
)

var pluginInfo = &plugins.InfoV1{
	ID:        "pubs",
	Version:   version.Version,
	BuildDate: version.BuildDate,
}

// PubsPlugin implements a flexible Webhook system providing a RESTful API
// to register hooks and a Websocket API for efficient receival.
type PubsPlugin struct {
	ctx context.Context
	srv plugins.ServerV1

	handler   http.Handler
	upgrader  *websocket.Upgrader
	cookie    *securecookie.SecureCookie
	pubsub    *pubsub.PubSub
	broadcast string

	count uint64
}

// Info returns the accociated plugins plugin.Info.
func (p *PubsPlugin) Info() *plugins.InfoV1 {
	return pluginInfo
}

// Initialize initizalizes the accociated plugin.
func (p *PubsPlugin) Initialize(ctx context.Context, errCh chan<- error, srv plugins.ServerV1) error {
	var err error

	p.ctx = ctx
	p.srv = srv

	p.handler = p.addRoutes(ctx, mux.NewRouter())
	p.upgrader = &websocket.Upgrader{
		ReadBufferSize:  websocketReadBufferSize,
		WriteBufferSize: websocketWriteBufferSize,
		CheckOrigin: func(req *http.Request) bool {
			// TODO(longsleep): Check if its a good idea to allow all origins.
			return true
		},
	}

	var hashKey []byte
	if hashKeyString := os.Getenv("KOPANO_PUBS_SECRET_KEY"); hashKeyString != "" {
		hashKey, err = hex.DecodeString(hashKeyString)
		if err != nil {
			return fmt.Errorf("pubs: failed to hex decode secret key: %v", err)
		}
	} else {
		hashKey = rndm.GenerateRandomBytes(64)
		p.srv.Logger().Warnln("pubs: using random secret key")
	}
	p.cookie = securecookie.New(hashKey, nil)
	p.cookie.MaxAge(0)

	p.pubsub = pubsub.New(256) //TODO(longsleep): Add capacity to configuration.
	p.broadcast = rndm.GenerateRandomString(32)

	srv.Logger().WithField("broadcast", p.broadcast).Debugf("pubs: initialize with %d bit key", len(hashKey)*8)

	return nil
}

// Close closes the accociated plugin.
func (p *PubsPlugin) Close() error {
	p.srv.Logger().Debugln("pubs: close")

	p.pubsub.Shutdown()

	return nil
}

// Register is the exported registration entry point as loaded by Kopano API to
// register plugins.
var Register plugins.RegisterPluginV1 = func() plugins.PluginV1 {
	return &PubsPlugin{}
}

// NOTE(longsleep): Keep main() to make the linter happy.
func main() {}
