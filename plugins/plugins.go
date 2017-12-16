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

package plugins

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"

	"stash.kopano.io/kc/kopano-api/proxy"
)

// RegisterPluginV1 is the register function plugins needs to expose as Register
// to be recognized to implement PluginV1.
type RegisterPluginV1 func() PluginV1

// PluginV1 is the interface a plugin needs to implement to be registered as
// a plugin.
type PluginV1 interface {
	Initialize(ctx context.Context, errCh chan<- error, srv ServerV1) error
	Close() error
	ServeHTTP(rw http.ResponseWriter, req *http.Request) (bool, error)
}

// ServerV1 is the interface how a plugin can integrate calls provided by
// Kopano API server.
type ServerV1 interface {
	Logger() logrus.FieldLogger

	AccessTokenRequired(next http.Handler) http.Handler
	HandleWithProxy(proxy proxy.Proxy, next http.Handler) http.Handler
}
