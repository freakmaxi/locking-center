package service

import (
	"encoding/binary"
	"io"
	"net"
	"time"
)

type SocketIO struct {
}

func NewSocketIO() *SocketIO {
	return &SocketIO{}
}

func (s *SocketIO) setDeadline(conn net.Conn, expectedTransferSize int) error {
	seconds := expectedTransferSize / defaultTransferSpeed
	if seconds < 0 {
		seconds = 0
	}
	seconds += 30

	return conn.SetDeadline(time.Now().Add(time.Second * time.Duration(seconds)))
}

func (s *SocketIO) ReadWithTimeout(conn net.Conn, buffer []byte, size int) error {
	if err := s.setDeadline(conn, size); err != nil {
		return err
	}
	_, err := io.ReadAtLeast(conn, buffer, size)
	return err
}

func (s *SocketIO) ReadBinaryWithTimeout(conn net.Conn, data interface{}) error {
	if err := s.setDeadline(conn, 0); err != nil {
		return err
	}
	return binary.Read(conn, binary.LittleEndian, data)
}

func (s *SocketIO) WriteWithTimeout(conn net.Conn, b []byte) error {
	if err := s.setDeadline(conn, len(b)); err != nil {
		return err
	}
	_, err := conn.Write(b)
	return err
}

func (s *SocketIO) WriteBinaryWithTimeout(conn net.Conn, data interface{}) error {
	if err := s.setDeadline(conn, 0); err != nil {
		return err
	}
	return binary.Write(conn, binary.LittleEndian, data)
}

func (s *SocketIO) Idle(conn net.Conn) {
	_ = conn.SetDeadline(time.Time{})
}
