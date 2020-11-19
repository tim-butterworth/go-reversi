package tcpimpl

import (
	"net"
	"github.com/google/uuid"
)

type PlayerConnection struct {
	Connection net.Conn
	Id uuid.UUID
}