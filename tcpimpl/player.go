package tcpimpl

import (
	"github.com/google/uuid"
)

type PlayerConnection struct {
	Connection Connection
	Id         uuid.UUID
}
