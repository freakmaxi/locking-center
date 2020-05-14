package service

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/freakmaxi/locking-center/mutex/common"
)

const commandBuffer = 4             // 4b
const defaultTransferSpeed = 625000 // bytes/s

type Manager interface {
	Listen(wg *sync.WaitGroup) error
}

type manager struct {
	address *net.TCPAddr
	mutex   *common.Mutex

	listener *net.TCPListener
	quiting  bool
}

func NewManager(address string, mutex *common.Mutex) (Manager, error) {
	if len(address) == 0 {
		return nil, fmt.Errorf("address should be defined")
	}
	addr, _ := net.ResolveTCPAddr("tcp", address)

	return &manager{
		address: addr,
		mutex:   mutex,
	}, nil
}

func (m *manager) Listen(wg *sync.WaitGroup) error {
	var err error
	m.listener, err = net.ListenTCP("tcp", m.address)
	if err != nil {
		return err
	}

	fmt.Printf("INFO: Manager Service has started listening at %s\n", m.address.String())

	go func() {
		defer wg.Done()

		for !m.quiting {
			c, err := m.listener.Accept()
			if err != nil {
				fmt.Printf("ERROR: Unable to accept connection: %s\n", err.Error())
				continue
			}
			go m.handler(c)
		}
	}()

	return nil
}

func (m *manager) setDeadline(conn net.Conn, expectedTransferSize int) error {
	seconds := expectedTransferSize / defaultTransferSpeed
	if seconds < 0 {
		seconds = 0
	}
	seconds += 30

	return conn.SetDeadline(time.Now().Add(time.Second * time.Duration(seconds)))
}

func (m *manager) readWithTimeout(conn net.Conn, buffer []byte, size int) error {
	if err := m.setDeadline(conn, size); err != nil {
		return err
	}
	_, err := io.ReadAtLeast(conn, buffer, size)
	return err
}

func (m *manager) readBinaryWithTimeout(conn net.Conn, data interface{}) error {
	if err := m.setDeadline(conn, 0); err != nil {
		return err
	}
	return binary.Read(conn, binary.LittleEndian, data)
}

func (m *manager) writeWithTimeout(conn net.Conn, b []byte) error {
	if err := m.setDeadline(conn, len(b)); err != nil {
		return err
	}
	_, err := conn.Write(b)
	return err
}

func (m *manager) writeBinaryWithTimeout(conn net.Conn, data interface{}) error {
	if err := m.setDeadline(conn, 0); err != nil {
		return err
	}
	return binary.Write(conn, binary.LittleEndian, data)
}

func (m *manager) handler(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, commandBuffer)

	if err := m.readWithTimeout(conn, buffer, len(buffer)); err != nil {
		fmt.Printf("ERROR: Stream unable to read: Connection: %s, %s\n", conn.RemoteAddr().String(), err.Error())
		return
	}

	if err := m.process(string(buffer), conn); err != nil {
		if err != io.EOF {
			fmt.Printf("ERROR: Service process is failed: address: %s,%s\n", conn.RemoteAddr(), err)
		}
		if _, err := conn.Write([]byte{'-'}); err != nil {
			fmt.Printf("ERROR: Service failed on unsuccess message: address: %s,%s\n", conn.RemoteAddr(), err)
		}
		return
	}
	if _, err := conn.Write([]byte{'+'}); err != nil {
		fmt.Printf("ERROR: Service failed on success message: address: %s,%s\n", conn.RemoteAddr(), err)
	}
}

func (m *manager) process(command string, conn net.Conn) error {
	switch command {
	case "KEYS":
		return m.keys(conn)
	case "RSET":
		return m.reset(conn)
	default:
		return fmt.Errorf("not a meaningful command")
	}
}

func (m *manager) keys(conn net.Conn) error {
	keys := m.mutex.Keys()

	if err := m.writeBinaryWithTimeout(conn, uint32(len(keys))); err != nil {
		return err
	}

	for _, key := range keys {
		keySize := uint8(len(key))
		if err := m.writeBinaryWithTimeout(conn, keySize); err != nil {
			return err
		}

		keyBytes := []byte(key)
		if err := m.writeWithTimeout(conn, keyBytes); err != nil {
			return err
		}
	}

	return nil
}

func (m *manager) reset(conn net.Conn) error {
	var resetKeysCount uint32
	if err := m.readBinaryWithTimeout(conn, &resetKeysCount); err != nil {
		return err
	}

	for ; resetKeysCount > 0; resetKeysCount-- {
		var keySize uint8
		if err := m.readBinaryWithTimeout(conn, &keySize); err != nil {
			return err
		}

		keyBytes := make([]byte, keySize)
		if err := m.readWithTimeout(conn, keyBytes, len(keyBytes)); err != nil {
			return err
		}

		m.mutex.Reset(string(keyBytes))

		if err := m.writeWithTimeout(conn, []byte("+")); err != nil {
			return err
		}
	}

	return nil
}
