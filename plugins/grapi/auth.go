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

package plugin

import (
	"errors"
	"net/http"

	"stash.kopano.io/kc/kapi/auth"
)

func (p *KopanoGroupwareCorePlugin) injectAuthIntoRequestHeaders(req *http.Request) error {
	var err error

	authRecord, _ := auth.RecordFromContext(req.Context())
	if authRecord != nil {
		authenticatedUserID := authRecord.AuthenticatedUserID
		if authRecord.ExtraClaims != nil {
			kcIDUserID, kcIDUsername := auth.KCIDFromClaims(authRecord.ExtraClaims)
			if kcIDUserID != "" {
				authenticatedUserID = kcIDUserID
			}
			if kcIDUsername != "" {
				req.Header.Set(usernameRequestHeaderName, kcIDUsername)
			} else {
				err = errors.New("missing kc.identity with username")
			}
		} else {
			req.Header.Del(usernameRequestHeaderName)
		}
		req.Header.Set(entryIDRequestHeaderName, authenticatedUserID)
	} else {
		err = errors.New("no auth record to inject")
	}

	return err
}
