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
	"strings"
	"time"

	"encoding/hex"
	"github.com/cskr/pubsub"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/websocket"
	"github.com/orcaman/concurrent-map"
	"stash.kopano.io/kgol/rndm"

	"stash.kopano.io/kc/kapi/plugins"
	"stash.kopano.io/kc/kapi/version"
)

// Buffer sizes for websocket payload data.
const (
	websocketReadBufferSize  = 1024 * 10
	websocketWriteBufferSize = 1024 * 10

	connectExpiration      = time.Duration(30) * time.Second
	connectCleanupInterval = time.Duration(1) * time.Minute
	connectKeySize         = 24
)

var pluginInfo = &plugins.InfoV1{
	ID:        "pubs",
	Version:   version.Version,
	BuildDate: version.BuildDate,
}

var scopesRequired = []string{"kopano/gc"}

// PubsPlugin implements a flexible Webhook system providing a RESTful API
// to register hooks and a Websocket API for efficient receival.
type PubsPlugin struct {
	ctx context.Context
	srv plugins.ServerV1

	handler   http.Handler
	keys      cmap.ConcurrentMap
	upgrader  *websocket.Upgrader
	cookie    *securecookie.SecureCookie
	pubsub    *pubsub.PubSub
	broadcast string

	count       uint64
	connections cmap.ConcurrentMap
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
	p.keys = cmap.New()
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

	if len(hashKey) < 32 {
		return fmt.Errorf("pubs: secret key too small, at least 32 bytes are required")
	}

	scopesRequiredString := os.Getenv("KOPANO_PUBS_REQUIRED_SCOPES")
	if scopesRequiredString != "" {
		scopesRequired = strings.Split(scopesRequiredString, " ")
	}
	p.srv.Logger().WithField("required_scopes", scopesRequired).Infoln("pubs: access requirements set up")

	p.pubsub = pubsub.New(256) //TODO(longsleep): Add capacity to configuration.
	p.broadcast = rndm.GenerateRandomString(32)

	p.connections = cmap.New()

	// Cleanup function.
	go func() {
		ticker := time.NewTicker(connectCleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.purgeExpiredKeys()
			case <-ctx.Done():
				return
			}
		}
	}()

	srv.Logger().WithField("broadcast", p.broadcast).Debugf("pubs: initialize with %d bits HMAC-SHA256 key", len(hashKey)*8)

	return nil
}

// Close closes the accociated plugin.
func (p *PubsPlugin) Close() error {
	p.srv.Logger().Debugln("pubs: close")

	p.pubsub.Shutdown()

	return nil
}

// NumActive returns the number of the currently active connections.
func (p *PubsPlugin) NumActive() uint64 {
	n := p.connections.Count()
	p.srv.Logger().Debugf("active connections: %d", n)
	return uint64(n)
}

type keyRecord struct {
	when time.Time
	user *userRecord
}

func (p *PubsPlugin) purgeExpiredKeys() {
	expired := make([]string, 0)
	deadline := time.Now().Add(-connectExpiration)
	var record *keyRecord
	for entry := range p.keys.IterBuffered() {
		record = entry.Val.(*keyRecord)
		if record.when.Before(deadline) {
			expired = append(expired, entry.Key)
		}
	}
	for _, key := range expired {
		p.keys.Remove(key)
	}
}

type userRecord struct {
	id string
}

// Register is the exported registration entry point as loaded by Kopano API to
// register plugins.
var Register plugins.RegisterPluginV1 = func() plugins.PluginV1 {
	return &PubsPlugin{}
}

// NOTE(longsleep): Keep main() to make the linter happy.
func main() {}
