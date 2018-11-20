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
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

const dbPath = "../../../test/tests.sqlite3"
const migrationsPaths = "./migrations"

var localRealm = "testing"
var localTestDB *sql.DB
var localTestKV *KV

func setup() {
	ctx := context.Background()

	db, err := sql.Open(driverNameSQLite3, dbPath)
	if err != nil {
		fmt.Printf("failed to setup local test database: %v\n", err)
		os.Exit(1)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	localTestDB = db
	localTestKV, err = New(driverNameSQLite3, dbPath, migrationsPaths, logger)
	if err != nil {
		fmt.Printf("failed to create KV: %v\n", err)
		os.Exit(1)
	}

	err = localTestKV.initialize(ctx, db)
	if err != nil {
		fmt.Printf("failed to initialize KV: %v\n", err)
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func teardown() {
	err := localTestKV.Close()
	if err != nil {
		fmt.Printf("failed to close local KV: %v\n", err)
	}
	err = localTestDB.Close()
	if err != nil {
		fmt.Printf("failed to close local test database: %v\n", err)
	}
	err = os.Remove(dbPath)
	if err != nil {
		fmt.Printf("failed to remove local test database: %v\n", err)
	}
}

func TestPingContext(t *testing.T) {
	db := localTestDB

	db.PingContext(context.Background())
}

func TestDBOpenSQLite3(t *testing.T) {
	dsn := dbPath + ".tmp"

	db, err := sql.Open(driverNameSQLite3, dsn)
	if err != nil {
		t.Fatal(err)
	}
	db.Ping()

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = os.Remove(dsn)
	if err != nil {
		t.Fatal(err)
	}
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

func kvGet(ctx context.Context, t *testing.T, kv *KV, record *Record, realm string) []*Record {
	records, err := kv.Get(ctx, realm, record)
	if err != nil {
		t.Fatal(err)
	}

	return records
}

func kvDelete(ctx context.Context, t *testing.T, kv *KV, record *Record, realm string) bool {
	ok, err := kv.Delete(ctx, realm, record)
	if err != nil {
		t.Fatal(err)
	}

	return ok
}

func compareRecords(t *testing.T, recordA *Record, recordB *Record) bool {
	a, err := recordA.EncodeToJSON()
	if err != nil {
		t.Fatal(err)
	}

	b, err := recordB.EncodeToJSON()
	if err != nil {
		t.Fatal(err)
	}

	return bytes.Compare(a, b) == 0
}

func TestKVCreateAndDelete(t *testing.T) {
	kv := localTestKV
	ctx := context.Background()

	collection := "createAndDelete"
	ownerID := "ownerA"
	clientID := "clientA"

	recordCreate := &Record{
		Collection:  &collection,
		Key:         "test1/doc1",
		Value:       []byte("aGVsbG8K"),
		ContentType: "text/plain",
		OwnerID:     ownerID,
		ClientID:    clientID,
	}
	err := kv.CreateOrUpdate(ctx, localRealm, recordCreate)
	if err != nil {
		t.Fatal(err)
	}

	recordDelete := &Record{
		Collection: recordCreate.Collection,
		Key:        recordCreate.Key,
	}
	ok := kvDelete(ctx, t, kv, recordDelete, localRealm)
	if ok {
		t.Fatalf("delete of doc without OwnerID and ClientID returned ok but must be not ok")
	}
	recordDelete.OwnerID = recordCreate.OwnerID
	ok = kvDelete(ctx, t, kv, recordDelete, localRealm)
	if ok {
		t.Fatalf("delete of doc without ClientID returned ok but must be not ok")
	}
	recordDelete.ClientID = recordCreate.ClientID
	ok = kvDelete(ctx, t, kv, recordDelete, "")
	if ok {
		t.Fatalf("delete of doc with wrong realm returned ok but must be not ok")
	}
	ok = kvDelete(ctx, t, kv, recordDelete, localRealm)
	if !ok {
		t.Fatalf("delete of doc retruned not ok but must be ok")
	}
}

func TestKVGetAndUpdateGet(t *testing.T) {
	kv := localTestKV
	ctx := context.Background()

	collection := "getAndUpdateGet"
	ownerID := "ownerB"
	clientID := "clientB"

	recordCreate := &Record{
		Collection:  &collection,
		Key:         "test1/doc2",
		Value:       []byte("aGVsbG8K"),
		ContentType: "text/plain",
		OwnerID:     ownerID,
		ClientID:    clientID,
	}
	err := kv.CreateOrUpdate(ctx, localRealm, recordCreate)
	if err != nil {
		t.Fatal(err)
	}

	recordGet := &Record{
		Collection: recordCreate.Collection,
		Key:        recordCreate.Key,
		OwnerID:    recordCreate.OwnerID,
		ClientID:   recordCreate.ClientID,
	}
	records := kvGet(ctx, t, kv, recordGet, localRealm)
	if len(records) != 1 {
		t.Fatalf("get yielded wrong number of records: %v\n", len(records))
	}
	if !compareRecords(t, recordCreate, records[0]) {
		t.Fatalf("get yielded different records\n")
	}

	recordCreate.Value = []byte("updated")
	err = kv.CreateOrUpdate(ctx, localRealm, recordCreate)
	if err != nil {
		t.Fatal(err)
	}
	records = kvGet(ctx, t, kv, recordGet, localRealm)
	if len(records) != 1 {
		t.Fatalf("get after update yielded wrong number of records: %v\n", len(records))
	}
	if !compareRecords(t, recordCreate, records[0]) {
		t.Fatalf("get after update yielded different records\n")
	}
}

func TestKVGetRecurse(t *testing.T) {
	kv := localTestKV
	ctx := context.Background()

	collection := "getRecurse"
	ownerID := "ownerC"
	clientID := "clientC"

	recordCreate1 := &Record{
		Collection:  &collection,
		Key:         "/doc1",
		Value:       []byte("aGVsbG8K"),
		ContentType: "text/plain",
		OwnerID:     ownerID,
		ClientID:    clientID,
	}
	err := kv.CreateOrUpdate(ctx, localRealm, recordCreate1)
	if err != nil {
		t.Fatal(err)
	}
	recordCreate2 := &Record{
		Collection:  recordCreate1.Collection,
		Key:         "/doc2",
		Value:       []byte("aGVsbG8K"),
		ContentType: "text/plain",
		OwnerID:     recordCreate1.OwnerID,
		ClientID:    recordCreate1.ClientID,
	}
	err = kv.CreateOrUpdate(ctx, localRealm, recordCreate2)
	if err != nil {
		t.Fatal(err)
	}

	recordGet := &Record{
		Collection: &collection,
		OwnerID:    recordCreate1.OwnerID,
		ClientID:   recordCreate1.ClientID,
	}
	records := kvGet(ctx, t, kv, recordGet, localRealm)
	if len(records) != 2 {
		t.Fatalf("get collection yielded wrong number of records: %v\n", len(records))
	}
}

func TestKVBatchCreateOrUpdate(t *testing.T) {
	kv := localTestKV
	ctx := context.Background()

	collection := "batchCreateOrUpdate"
	ownerID := "ownerD"
	clientID := "clientD"
	number := 100

	records := make([]*Record, number, number)
	for i := 0; i < number; i++ {
		records[i] = &Record{
			Collection:  &collection,
			Key:         "num" + string(i),
			Value:       []byte("aGVsbG8K"),
			ContentType: "text/plain",
			OwnerID:     ownerID,
			ClientID:    clientID,
		}
	}

	err := kv.BatchCreateOrUpdate(ctx, localRealm, records)
	if err != nil {
		t.Fatal(err)
	}

	recordGet := &Record{
		Collection: &collection,
		OwnerID:    ownerID,
		ClientID:   clientID,
	}
	records = kvGet(ctx, t, kv, recordGet, localRealm)
	if len(records) != number {
		t.Fatalf("get collection after batch yielded wrong number of records: %v\n", len(records))
	}
}
