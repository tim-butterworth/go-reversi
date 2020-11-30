package tcpimpl

import (
	"github.com/google/uuid"
	"net"
)

type PlayerConnection struct {
	Connection net.Conn
	Id         uuid.UUID
}
