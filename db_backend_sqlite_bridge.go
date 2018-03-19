// +build SQLITE_BACKEND

package main

import (
	"database/sql"
	"github.com/john-sharp/jerver/entities"
	"github.com/john-sharp/jerver/sqlite-dbbackend"
	"github.com/satori/go.uuid"
)

type messageCollection struct {
	getFromUuidStmt  *sql.Stmt
	createEntityStmt *sql.Stmt
	deleteEntityStmt *sql.Stmt
}

func (mc *messageCollection) getFromUuid(targetUuid uuid.UUID) (*entities.Message, error) {
	return dbbackend.GetMessageFromUuid(targetUuid)
}
