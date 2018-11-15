CREATE TABLE kv (
	id INTEGER PRIMARY KEY AUTO_INCREMENT,
	collection VARCHAR(128),
	ekey VARCHAR(255) NOT NULL,
	value BLOB NOT NULL,
	content_type TEXT,
	owner_id VARCHAR(128),
	client_id VARCHAR(128),
	realm VARCHAR(64) NOT NULL,
	required_scopes TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
