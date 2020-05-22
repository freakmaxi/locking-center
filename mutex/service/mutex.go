package service

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/freakmaxi/locking-center/mutex/common"
)

type Mutex interface {
	Listen(wg *sync.WaitGroup) error
}

type mutex struct {
	address  *net.TCPAddr
	mutex    *common.Mutex
	socketIO *SocketIO

	listener *net.TCPListener
	quiting  bool
}

func NewMutex(address string, m *common.Mutex) (Mutex, error) {
	if len(address) == 0 {
		return nil, fmt.Errorf("address should be defined")
	}
	addr, _ := net.ResolveTCPAddr("tcp4", address)

	return &mutex{
		address:  addr,
		mutex:    m,
		socketIO: NewSocketIO(),
	}, nil
}

func (m *mutex) Listen(wg *sync.WaitGroup) error {
	var err error
	m.listener, err = net.ListenTCP("tcp", m.address)
	if err != nil {
		return err
	}

	fmt.Printf("INFO: Mutex Service has started listening at %s\n", m.address.String())

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

func (m *mutex) handler(conn net.Conn) {
	defer conn.Close()

	if err := m.process(conn); err != nil {
		if err != io.EOF {
			fmt.Printf("ERROR: Service process is failed: address: %s,%s\n", conn.RemoteAddr(), err)
		}
		if err := m.socketIO.WriteWithTimeout(conn, []byte{'-'}); err != nil {
			fmt.Printf("ERROR: Service failed on unsuccess message: address: %s,%s\n", conn.RemoteAddr(), err)
		}
		return
	}
	if err := m.socketIO.WriteWithTimeout(conn, []byte{'+'}); err != nil {
		fmt.Printf("ERROR: Service failed on success message: address: %s,%s\n", conn.RemoteAddr(), err)
	}
}

func (m *mutex) process(conn net.Conn) error {
	var keySize int8
	if err := m.socketIO.ReadBinaryWithTimeout(conn, &keySize); err != nil {
		return err
	}

	keyBytes := make([]byte, keySize)
	if err := m.socketIO.ReadWithTimeout(conn, keyBytes, len(keyBytes)); err != nil {
		return err
	}
	key := string(keyBytes)

	var action byte
	if err := m.socketIO.ReadBinaryWithTimeout(conn, &action); err != nil {
		return err
	}

	switch action {
	case 1:
		for !m.mutex.Lock(key, conn.RemoteAddr()) {
		}
		return nil
	case 2:
		m.mutex.Unlock(key)
		return nil
	case 3:
		m.mutex.Reset(key)
		return nil
	}

	return fmt.Errorf("undefined action")
}
