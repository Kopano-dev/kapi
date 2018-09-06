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
	"context"

	"github.com/dgrijalva/jwt-go"
	kcoidc "stash.kopano.io/kc/libkcoidc"
)

type contextKey string

const (
	authRecordcontextKey contextKey = "authRecord"
)

// Record is the auth record holding the current authenticated details.
type Record struct {
	AuthenticatedUserID string

	StandardClaims *jwt.StandardClaims
	ExtraClaims    *kcoidc.ExtraClaimsWithType
}

// AuthenticatedUserIDFromContext returns the provided requests authentication
// ID if present.
func AuthenticatedUserIDFromContext(ctx context.Context) (string, bool) {
	if v := ctx.Value(authRecordcontextKey); v != nil {
		record, _ := v.(*Record)
		return record.AuthenticatedUserID, true
	}

	return "", false
}

// RecordFromContext returns the provided requests authentication
// ID if present.
func RecordFromContext(ctx context.Context) (*Record, bool) {
	if v := ctx.Value(authRecordcontextKey); v != nil {
		return v.(*Record), true
	}

	return nil, false
}

// ContextWithRecord adds the provided auth record to the provided parent
// context and returns a context holding the value.
func ContextWithRecord(parent context.Context, record *Record) context.Context {
	return context.WithValue(parent, authRecordcontextKey, record)
}
