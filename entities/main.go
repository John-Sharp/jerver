package entities

import (
	"github.com/satori/go.uuid"
)

type Generic interface{}

type Message struct {
	Id       uuid.UUID
	ThreadId uuid.UUID
	AuthorId uuid.UUID
	Content  string
}

type MessageEdit struct {
	ThreadId *uuid.UUID
	AuthorId *uuid.UUID
	Content  *string
}
