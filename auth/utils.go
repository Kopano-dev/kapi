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

package auth

import (
	kcoidc "stash.kopano.io/kc/libkcoidc"
)

// Token claims used by Kopano Konnect.
const (
	IdentityClaim           = "kc.identity"
	IdentifiedUserIDClaim   = "kc.i.id"
	IdentifiedUsernameClaim = "kc.i.un"
	AuthorizedScopesClaim   = "kc.authorizedScopes"
)

// KCIDFromClaims extracts extra Kopano Connect identified claims from the
// provided extra claims.
func KCIDFromClaims(claims *kcoidc.ExtraClaimsWithType) (kcIDUserID, kcIDUsername string) {
	if identityClaims, _ := (*claims)[IdentityClaim].(map[string]interface{}); identityClaims != nil {
		kcIDUserID, _ = identityClaims[IdentifiedUserIDClaim].(string)
		kcIDUsername, _ = identityClaims[IdentifiedUsernameClaim].(string)
	}

	return kcIDUserID, kcIDUsername
}

// KCAuthorizedScopesFromClaims authorize scopes as bool map from the provided
// extra claims.
func KCAuthorizedScopesFromClaims(claims *kcoidc.ExtraClaimsWithType) map[string]bool {
	if authorizedScopes, _ := (*claims)[AuthorizedScopesClaim].([]interface{}); authorizedScopes != nil {
		authorizedScopesMap := make(map[string]bool)
		for _, scope := range authorizedScopes {
			authorizedScopesMap[scope.(string)] = true
		}

		return authorizedScopesMap
	}

	return nil
}
