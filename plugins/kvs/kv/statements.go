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

package kv

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type stmtID uint64

const (
	stmtIDGet stmtID = iota
	stmtIDGetCollection
	stmtIDCreateOrUpdate
	stmtIDDelete
)

// NOTE(longsleep): Those statements must work with both MySQL and SQLite3.
var preparedStmts = map[stmtID]string{
	stmtIDGet: `
		SELECT
			ekey,
			value,
			content_type,
			required_scopes
		FROM kv WHERE
			ekey = ? AND
			owner_id = ? AND
			client_id = ? AND
			realm = ?
		LIMIT 500`,

	stmtIDGetCollection: `
		SELECT
			ekey,
			value,
			content_type,
			required_scopes
		FROM kv WHERE
			collection = ? AND
			owner_id = ? AND
			client_id = ? AND
			realm = ?
		LIMIT 500`,

	stmtIDCreateOrUpdate: `
		REPLACE INTO kv(
			collection,
			ekey,
			value,
			content_type,
			owner_id,
			client_id,
			realm,
			required_scopes
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,

	stmtIDDelete: `
		DELETE FROM kv WHERE
			ekey = ? AND
			owner_id = ? AND
			client_id = ? AND
			realm = ?`,
}

func prepareStmts(db *sql.DB, dbDriverName string) (map[stmtID]*sql.Stmt, error) {
	stmts := make(map[stmtID]*sql.Stmt)
	for id, stmtString := range preparedStmts {
		stmt, err := db.Prepare(stmtString)
		if err != nil {
			return nil, err
		}
		stmts[id] = stmt
	}

	return stmts, nil
}

// Stmt gets a statemen from the accociated kv with the provided ID.
func (kv *KV) Stmt(ctx context.Context, id stmtID) (*sql.Stmt, error) {
	select {
	case <-kv.initializing:
		// All good, initialized.
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(waitForInitializeTimeout):
		return nil, errors.New("timeout")
	}

	stmt, ok := kv.stmts[id]
	if !ok {
		return nil, errors.New("no such statement")
	}
	return stmt, nil
}
