# Kopano API kvs plugin

The kvs plugin provides a key/value store system HTTP API to store data
persistently.

In its current state, the kv store can only store documents with size up to
16 KiB and will not return more than 500 documents for recursive get requests.
This is by design - if that does not fit your storage requirements, either
reorganize the data set or kvs is not for you.

## Configuration

`KOPANO_KVS_DB_DRIVER` is an environment variable which defines what backend
database driver to use for persistent storage. Currently supported drivers are
`sqlite3` and `mysql.`

`KOPANO_KVS_DB_DATASOURCE` is an environment variable which describes the data
source name (DNS) of the backend database. The exact value depends on the used
database driver. For `sqlite3` the value is the full path to a database file and
for `mysql` it isa MYSQL DSN (See https://github.com/go-sql-driver/mysql#dsn-data-source-name)
for details.

`KOPANO_KVS_DB_MIGRATIONS` is an environment variable pointing to the folder
where to find database migrations. It must point to a folder which contains a
sub folder which name matches the configured database driver.

`KOPANO_KVS_ALLOW_CORS` is an environment variable which if set to `1`
enables CORS (Cross Origin Resource Sharing) HTTP requests and headers so that
the REST endpoints provided by this plugin can be used from a browser cross
origin.

## HTTP API v1

The base URL to this API is `/api/kvs/v1`. All example URLs are sub paths of
this base URL.

### Authentication

All endpoints of the kvs API require [OAuth2 Bearer authentication](https://tools.ietf.org/html/rfc6750#section-2.1) with an access
token (if not stated otherwise).

### Realms

The kvs data set is grouped into realms and keys. Currently the only
supported realm is `user` which limits data access to particular users. The key
is the URL to a specific document. It can contain slashes to group individual
keys loosely together. The first segment of such a key path becomes a collection
and all keys with the same collection can be fetched efficiently using a
recursive query.

### Create or Update

Creates or replaces a document identified by key with the given value.

```
PUT /kv/${realm}/${collection/key/path}
200
```

```
PUT /kv/${realm}/${collection/prefix}?batch=1
200
```

The `realm` is the realm which defines kv storage behavior. Rest of the URL is
the key where the first part becomes the `collection`.

If the batch parameter is given, the kvs accepts a JSON array as input data
similar to what is returned by Get in recurse=1 mode. All keys of the entries
will have the collection and prefix used from the URL. Batch mode is
transactional and will only complete if all entries have been written using
ReadCommited transaction isolation. Entry values must be Base64 encoded if the
entry `content_type` attribute is not `application/json`.

### Get

Retrieves the item identified by realm and key path.

```
GET /kv/${realm}/${collection/key}
200
```
```
{
	"key": "${collection/key}"
	"value": "whatever was stored, Base64 encoded if not JSON",
	"content_type": "application/json"

}
```

```
GET /kv/${realm}/${collection}?recurse=1
200
```
```
[
	{
		"key": "${collection}/first-doc",
		"value": {},
		"content_type": "application/json"
	},
	{
		"key": "${collection}/other-doc",
		"value": {},
		"content_type": "application/json"
	}
]
```

```
GET /kv/${realm}/${collection/key}?raw=1
200
```
```
{}
```

```
GET /kv/${realm}/${collection/non/existing/key}
404
```

#### Supported parameters for Get requests

`?raw=1`     returns the raw document as stored instead of the JSON envelope.
`?recurse=1` returns a JSON array returning all keys which have the same prefix
             as the given key. For efficient use, limit this to collections. You
             can use more specific prefixes for nested keys but keep in mind
             that the result set is filtered after it was received from storage.

### Delete

Removes the item from storage so it can no longer be retrieved.

```
DELETE /kv/${realm}/${collection/key}
200
```

```
DELETE /kv/${realm}/${collection/key}
404
```

### Usage examples

This assumes you have [curl](https://curl.haxx.se/) in your path.

```
$ export KVS_HOST=https://localhost:8428
$ export TOKEN_VALUE=<access_token>
$ curl -s -k -XPUT -H "Authorization: Bearer $TOKEN_VALUE" "$KVS_HOST/api/kvs/v1/kv/user/test1/doc1" -H "Content-Type: application/json" --data '{}'
$ curl -s -k -H "Authorization: Bearer $TOKEN_VALUE" "$KVS_HOST/api/kvs/v1/kv/user/test1/doc1"
{
  "key": "test1/doc1",
  "value": {}
}
$ curl -s -k -XDELETE -H "Authorization: Bearer $TOKEN_VALUE" "$KVS_HOST/api/kvs/v1/kv/user/test1/doc1"
```
