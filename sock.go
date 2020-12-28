package zkm

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type Sock struct {
	c net.Conn
}

func NewSock(conn net.Conn) *Sock {
	return &Sock{c: conn}
}

func (s *Sock) Close() error {
	return s.c.Close()
}

func (s *Sock) Write(pdu *Pdu) error {
	_, err := s.c.Write(pdu.Serialize())
	return err
}

func (s *Sock) Read() (*Pdu, error) {
	rawL := make([]byte, pduHeaderPartSize)
	_, err := io.ReadFull(s.c, rawL)

	if err != nil {
		return nil, err
	}

	l := binary.BigEndian.Uint32(rawL)
	if l < 4*pduHeaderPartSize {
		return nil, fmt.Errorf("PDU too small: %d < %d", l, 4*pduHeaderPartSize)
	}

	b := make([]byte, l-pduHeaderPartSize)
	_, err = io.ReadFull(s.c, b)

	if err != nil {
		return nil, err
	}

	pdu := NewEmptyPdu()
	err = pdu.Deserialize(append(rawL, b...))

	if err != nil {
		return nil, err
	}

	return pdu, nil
}
