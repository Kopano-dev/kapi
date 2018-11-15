CREATE TABLE kv (
	id INTEGER PRIMARY KEY ASC,
	collection TEXT,
	ekey TEXT NOT NULL,
	value BLOB NOT NULL,
	content_type TEXT,
	owner_id TEXT,
	client_id TEXT,
	realm TEXT NOT NULL,
	required_scopes TEXT
);

CREATE INDEX idx_get ON kv (ekey, owner_id, client_id, realm);

CREATE UNIQUE INDEX idx_no_dup ON kv (ekey, owner_id, realm);

CREATE INDEX idx_collection ON kv (collection, owner_id, realm);
