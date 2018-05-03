package db

import (
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const dbFilename = "data/data.db"

const schema = `
CREATE TABLE users(
   id              INTEGER PRIMARY KEY   NOT NULL,
   first_name      TEXT                  NOT NULL   DEFAULT "",
   last_name       TEXT                  NOT NULL   DEFAULT "",
   username        TEXT                  NOT NULL   DEFAULT "",
   language_code   TEXT                  NOT NULL   DEFAULT "",
   chat_state      TEXT                  NOT NULL   DEFAULT "",
   first_message   INTEGER               NOT NULL   DEFAULT 0,
   last_message    INTEGER               NOT NULL   DEFAULT 0
);

CREATE TABLE sockets(
   id                  INTEGER PRIMARY KEY   NOT NULL,
   lat                 REAL                  NOT NULL,
   lng                 REAL                  NOT NULL,
   name                TEXT                  NOT NULL  DEFAULT "",
   description         TEXT                  NOT NULL  DEFAULT "",
   added_by            INTEGER               NOT NULL,
   last_confirmation   INTEGER               NOT NULL  DEFAULT 0
);

CREATE TABLE photos(
	id          INTEGER PRIMARY KEY   NOT NULL,
	socket      INTEGER               NOT NULL,
	user        INTEGER               NOT NULL,
	added       INTEGER               NOT NULL   DEFAULT 0,
	url         TEXT                  NOT NULL   DEFAULT "",
	media_id    TEXT                  NOT NULL   DEFAULT "",
	downloaded  INTEGER               NOT NULL   DEFAULT 0
);
`

var db *sqlx.DB

func init() {
	needCreation := false
	if _, err := os.Stat(dbFilename); os.IsNotExist(err) {
		needCreation = true
	}

	db = sqlx.MustConnect("sqlite3", dbFilename)
	if needCreation {
		db.MustExec(schema)
	}
}

func Close() {
	db.Close()
}
