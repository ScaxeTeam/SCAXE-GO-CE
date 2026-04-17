package net

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
)

type Conn struct {
	Raw	net.Conn
	Reader	*bufio.Reader
}

func Dial(addr string) (*Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Conn{
		Raw:	conn,
		Reader:	bufio.NewReader(conn),
	}, nil
}

func (c *Conn) Close() error {
	return c.Raw.Close()
}

func (c *Conn) ReadPacket() (packetID int32, payload []byte, err error) {
	length, err := ReadVarInt(c.Reader)
	if err != nil {
		return 0, nil, err
	}

	if length <= 0 || length > 4194304 {
		return 0, nil, fmt.Errorf("invalid packet length: %d", length)
	}

	fullPayload := make([]byte, length)
	_, err = io.ReadFull(c.Reader, fullPayload)
	if err != nil {
		return 0, nil, err
	}

	bufReader := bytes.NewReader(fullPayload)
	packetID, err = ReadVarInt(bufReader)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read packet ID: %v", err)
	}

	payload = fullPayload[len(fullPayload)-bufReader.Len():]
	return packetID, payload, nil
}

func (c *Conn) WritePacket(packetID int32, payload []byte) error {
	buf := new(bytes.Buffer)

	_, err := WriteVarInt(buf, packetID)
	if err != nil {
		return err
	}

	buf.Write(payload)
	packetData := buf.Bytes()

	outBuf := new(bytes.Buffer)
	_, err = WriteVarInt(outBuf, int32(len(packetData)))
	if err != nil {
		return err
	}
	outBuf.Write(packetData)

	_, err = c.Raw.Write(outBuf.Bytes())
	return err
}
