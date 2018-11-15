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
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/mattn/go-sqlite3"    // SQLite3 driver

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/mysql"   // MySQL migrate support
	"github.com/golang-migrate/migrate/database/sqlite3" // SQLite3 migrate support
	_ "github.com/golang-migrate/migrate/source/file"    // File source support

	"github.com/sirupsen/logrus"
)

const (
	driverNameSQLite3 = "sqlite3"
	driverNameMySQL   = "mysql"
)

const (
	waitForInitializeTimeout = 5 * time.Second
	dbConnectionTimeout      = 10 * time.Second
	dbMigrateLockTimeout     = 15 * time.Second
)

// KV implements a key value storage using a backend sql storage.
type KV struct {
	sync.Mutex

	db                 *sql.DB
	dbDriverName       string
	dbDataSourceName   string
	dbMigrationsSource string

	logger logrus.FieldLogger

	quit         chan struct{}
	initializing chan struct{}
	cancel       context.CancelFunc
	migrate      *migrate.Migrate
	stmts        map[stmtID]*sql.Stmt
}

// New creates a new KV using the provided options.
func New(dbDriverName string, dbDataSourceName string, dbMigrationsBasePath string, logger logrus.FieldLogger) (*KV, error) {
	if dbDataSourceName == "" {
		return nil, fmt.Errorf("datasource is empty")
	}
	if dbDriverName == "" {
		dbDriverName = driverNameSQLite3
	}
	if dbMigrationsBasePath == "" {
		dbMigrationsBasePath = "./plugins/kvs/kv/migrations"
	}

	kv := &KV{
		dbDataSourceName:   dbDataSourceName,
		dbDriverName:       dbDriverName,
		dbMigrationsSource: "file://" + dbMigrationsBasePath + "/" + dbDriverName,

		quit:         make(chan struct{}),
		initializing: make(chan struct{}),
		logger:       logger,
	}

	return kv, nil
}

// Initialize connects to the associated KV store and runs migrations
// as required.
func (kv *KV) Initialize(parentCtx context.Context) error {
	logger := kv.logger
	dbDriverName := kv.dbDriverName
	dbDataSourceName := kv.dbDataSourceName
	dbMigrationsSource := kv.dbMigrationsSource

	kv.Lock()
	select {
	case <-kv.quit:
		// Do nothing.
		kv.Unlock()
		return nil
	default:
		// Continue normal.
	}
	ctx, cancel := context.WithCancel(parentCtx)
	kv.cancel = cancel
	kv.Unlock()

	var db *sql.DB
	var err error
	switch dbDriverName {
	case driverNameSQLite3:
		fallthrough
	case driverNameMySQL:
		db, err = sql.Open(dbDriverName, dbDataSourceName)
	default:
		return fmt.Errorf("unsupported database: %v", dbDriverName)
	}
	if err != nil {
		return err
	}

	testCtx, timeout := context.WithTimeout(ctx, dbConnectionTimeout)
	defer timeout()
	err = db.PingContext(testCtx)
	if err != nil {
		return fmt.Errorf("database not available: %v", err)
	}

	var driver database.Driver
	switch dbDriverName {
	case driverNameSQLite3:
		driver, err = sqlite3.WithInstance(db, &sqlite3.Config{})
	case driverNameMySQL:
		driver, err = mysql.WithInstance(db, &mysql.Config{})
	}
	if err != nil {
		return fmt.Errorf("failed to open database migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		dbMigrationsSource,
		dbDriverName,
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to open database migrations: %v", err)
	}
	m.Log = &migrationLogger{logger, true}
	m.LockTimeout = dbMigrateLockTimeout
	kv.Lock()
	select {
	case <-kv.quit:
		// Do nothing.
		kv.Unlock()
		return nil
	default:
		// Continue normal.
	}
	kv.migrate = m
	kv.Unlock()

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		logger.WithError(err).Errorf("kv: failed to fetch database version info: %v", err)
		return err
	}
	logger.Debugf("kv: database version: %d dirty: %v", version, dirty)

	err = m.Up()
	switch err {
	case migrate.ErrNoChange:
	case nil:
	default:
		logger.WithError(err).Warnf("kv: database migration failed: %v", err)
	}

	// Continue locked to allow clean shut down.
	kv.Lock()
	defer kv.Unlock()

	kv.migrate = nil
	select {
	case <-kv.quit:
		// Do nothing.
		return nil
	default:
		// Continue normal.
	}

	stmts, err := prepareStmts(db, dbDriverName)
	if err != nil {
		return err
	}

	kv.db = db
	kv.stmts = stmts
	close(kv.initializing)

	return nil
}

// Close closes the accociated KV including everything in it.
func (kv *KV) Close() error {
	kv.Lock()
	defer kv.Unlock()

	var err error
	if kv.migrate != nil {
		kv.migrate.GracefulStop <- true
	}
	if kv.cancel != nil {
		kv.cancel()
	}
	close(kv.quit)

	for _, stmt := range kv.stmts {
		err = stmt.Close()
		if err != nil {
			kv.logger.Warnf("kv: failed to close statement: %v", err)
		}
	}

	if kv.db != nil {
		err = kv.db.Close()
	}
	kv.db = nil
	kv.stmts = nil

	return err
}

// Get Implements data retrieval from the accociated store.
func (kv *KV) Get(ctx context.Context, realm string, record *Record) ([]*Record, error) {
	var stmt *sql.Stmt
	var err error
	var rows *sql.Rows

	if record.Collection == nil {
		stmt, err = kv.Stmt(ctx, stmtIDGet)
		if err == nil {
			rows, err = stmt.QueryContext(ctx, record.Key, record.OwnerID, record.ClientID, realm)
		}
	} else {
		stmt, err = kv.Stmt(ctx, stmtIDGetCollection)
		if err == nil {
			rows, err = stmt.QueryContext(ctx, record.Collection, record.OwnerID, record.ClientID, realm)
		}
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*Record, 0)
	for rows.Next() {
		r := &Record{}
		err = rows.Scan(&r.Key, &r.Value, &r.ContentType, &r.RequiredScopes)
		if err != nil {
			kv.logger.Warnf("kv: failed to process database result: %v", err)
		}
		result = append(result, r)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// CreateOrUpdate implements data storage.
func (kv *KV) CreateOrUpdate(ctx context.Context, realm string, record *Record) error {
	stmt, err := kv.Stmt(ctx, stmtIDCreateOrUpdate)
	if err != nil {
		return err
	}

	res, err := stmt.ExecContext(ctx, record.Collection, record.Key, record.Value, record.ContentType, record.OwnerID, record.ClientID, realm, record.RequiredScopes)
	if err != nil {
		return err
	}

	lastInsertID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	kv.logger.Debugf("kv: create or update ID = %d, affected = %d\n", lastInsertID, rowsAffected)

	return nil
}
