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
	"testing"
)

func TestDBOpenSQLite3(t *testing.T) {
	db, err := sql.Open(driverNameSQLite3, "../../../test/tests.sqlite3")
	if err != nil {
		t.Fatal(err)
	}

	db.PingContext(context.Background())
	db.Close()
}

func TestDBOpenMySQL(t *testing.T) {
	db, err := sql.Open(driverNameMySQL, "user:password@/dbname")
	if err != nil {
		t.Fatal(err)
	}

	// MySQL Open does nothing with the server. But since we do not have a
	// MySQL server available while testing, we cannot do more.
	db.Close()
}
